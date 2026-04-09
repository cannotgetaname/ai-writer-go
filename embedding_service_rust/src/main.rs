//! Lightweight embedding server using Candle (Rust ML framework)
//!
//! Features:
//! - Sentence embeddings using sentence-transformers models
//! - HTTP API compatible with Go backend
//! - HF mirror support for China users
//! - Minimal binary size (~15-30MB)

use axum::{
    extract::Json,
    http::StatusCode,
    response::{IntoResponse, Response},
    routing::{get, post},
    Router,
};
use candle_core::{Device, Tensor, safetensors::load};
use candle_nn::VarBuilder;
use candle_transformers::models::bert::{BertModel, Config};
use anyhow::{Context, Result};
use serde::{Deserialize, Serialize};
use std::env;
use std::fs;
use std::net::SocketAddr;
use std::path::PathBuf;
use std::sync::Arc;
use tokenizers::Tokenizer;
use tracing::info;

/// Embedding request from client (supports both single and batch)
#[derive(Debug, Deserialize)]
struct EmbedRequest {
    /// Single text (for convenience)
    text: Option<String>,
    /// Multiple texts (batch request, matches Python API)
    texts: Option<Vec<String>>,
}

/// Embedding response to client (matches Python API format)
#[derive(Debug, Serialize)]
struct EmbedResponse {
    /// Batch embeddings
    embeddings: Vec<Vec<f32>>,
    /// Model name
    model: String,
    /// Embedding dimension
    dimension: usize,
}

/// Health check response
#[derive(Debug, Serialize)]
struct HealthResponse {
    status: String,
    model_loaded: bool,
    model: String,
}

/// Global embedding model state
struct EmbeddingModel {
    model: BertModel,
    tokenizer: Tokenizer,
    model_name: String,
    dimensions: usize,
}

/// Mean pooling for sentence embeddings
fn mean_pooling(last_hidden_state: &Tensor, attention_mask: &Tensor) -> Result<Tensor> {
    // Convert attention mask to F32 for multiplication
    let attention_mask_f32 = attention_mask.to_dtype(candle_core::DType::F32)?;

    // Get dimensions
    let (batch_size, seq_len) = (attention_mask_f32.dim(0)?, attention_mask_f32.dim(1)?);
    let hidden_size = last_hidden_state.dim(2)?;

    // Expand attention mask: [batch, seq_len] -> [batch, seq_len, hidden_size]
    let mask_expanded = attention_mask_f32
        .unsqueeze(2)?
        .expand((batch_size, seq_len, hidden_size))?;

    // Apply mask and sum: [batch, seq_len, hidden_size] * [batch, seq_len, hidden_size] -> [batch, hidden_size]
    let masked = last_hidden_state.mul(&mask_expanded)?;
    let sum = masked.sum(1)?;

    // Get mask sum for normalization: [batch, seq_len] -> [batch]
    let mask_sum = attention_mask_f32.sum(1)?;
    // [batch] -> [batch, hidden_size] for broadcasting division
    let mask_sum_expanded = mask_sum.unsqueeze(1)?.expand((batch_size, hidden_size))?;

    // Normalize (avoid division by zero)
    let pooled = sum.div(&mask_sum_expanded)?;
    Ok(pooled)
}

/// L2 normalize a tensor along axis 1
fn l2_normalize(tensor: &Tensor) -> Result<Tensor> {
    // Get dimensions
    let batch_size = tensor.dim(0)?;
    let hidden_size = tensor.dim(1)?;

    // Compute L2 norm for each row: [batch, hidden_size] -> [batch]
    let norm = tensor.sqr()?.sum(1)?;
    let sqrt_norm = norm.sqrt()?;
    // [batch] -> [batch, hidden_size] for broadcasting division
    let sqrt_norm_expanded = sqrt_norm.unsqueeze(1)?.expand((batch_size, hidden_size))?;

    // Divide by norm
    let normalized = tensor.div(&sqrt_norm_expanded)?;
    Ok(normalized)
}

/// Check if we can connect to HuggingFace (detect China network)
async fn can_connect_huggingface() -> bool {
    // Try to connect to huggingface.co with a timeout
    use std::time::Duration;

    let client = reqwest::Client::builder()
        .timeout(Duration::from_secs(5))
        .build()
        .unwrap_or_default();

    // Try a lightweight request
    match client.get("https://huggingface.co").send().await {
        Ok(resp) => resp.status().is_success(),
        Err(_) => false,
    }
}

/// Load embedding model from HuggingFace Hub with mirror fallback
async fn load_model(model_name: &str) -> Result<Arc<EmbeddingModel>> {
    info!("Loading model: {}", model_name);

    // Determine which endpoint to use
    let use_mirror = if env::var("HF_ENDPOINT").is_ok() {
        // User explicitly set endpoint, respect it
        false
    } else if env::var("HF_MIRROR").is_ok() || env::var("CHINA_MIRROR").is_ok() {
        // User explicitly requested mirror
        true
    } else {
        // Auto-detect: try huggingface.co, fallback to mirror if unreachable
        info!("Detecting network connectivity...");
        let can_connect = can_connect_huggingface().await;
        if !can_connect {
            info!("Cannot reach huggingface.co, will use HF mirror");
        }
        !can_connect
    };

    if use_mirror && env::var("HF_ENDPOINT").is_err() {
        env::set_var("HF_ENDPOINT", "https://hf-mirror.com");
        info!("Using HF mirror: https://hf-mirror.com (for China users)");
    }

    // Download model files from HF Hub
    let api = hf_hub::api::tokio::Api::new()?;
    let repo = api.model(model_name.to_string());

    // Get model files
    info!("Downloading config.json...");
    let config_path = repo.get("config.json").await?;

    info!("Downloading model.safetensors...");
    let weights_path = match repo.get("model.safetensors").await {
        Ok(path) => path,
        Err(_) => {
            info!("model.safetensors not found, trying pytorch_model.bin...");
            repo.get("pytorch_model.bin").await?
        }
    };

    info!("Downloading tokenizer.json...");
    let tokenizer_path = repo.get("tokenizer.json").await?;

    // Load config
    let config_content = fs::read_to_string(&config_path)?;
    let config: Config = serde_json::from_str(&config_content)
        .context("Failed to parse config.json")?;

    // Determine dimensions from config
    let dimensions = config.hidden_size;

    // Load tokenizer
    let tokenizer = Tokenizer::from_file(&tokenizer_path)
        .map_err(|e| anyhow::anyhow!("Tokenizer error: {}", e))?;

    // Load model weights from safetensors
    let device = Device::Cpu;
    let tensors = load(&weights_path, &device)?;
    let vb = VarBuilder::from_tensors(tensors, candle_core::DType::F32, &device);
    let model = BertModel::load(vb, &config)?;

    info!(
        "Model loaded successfully: {} ({} dimensions)",
        model_name, dimensions
    );

    Ok(Arc::new(EmbeddingModel {
        model,
        tokenizer,
        model_name: model_name.to_string(),
        dimensions,
    }))
}

/// Generate embedding for text
fn embed_text(model: &EmbeddingModel, text: &str) -> Result<Vec<f32>> {
    // Tokenize input
    let encoding = model
        .tokenizer
        .encode(text, true)
        .map_err(|e| anyhow::anyhow!("Tokenization error: {}", e))?;

    let tokens = encoding.get_ids();
    let attention_mask = encoding.get_attention_mask();

    // Create tensors and add batch dimension (1, seq_len)
    let tokens_tensor = Tensor::new(tokens, &Device::Cpu)?.unsqueeze(0)?;
    let attention_tensor = Tensor::new(attention_mask, &Device::Cpu)?.unsqueeze(0)?;

    // Create token type IDs (all zeros for single sentence)
    let token_type_ids = Tensor::zeros((1, tokens.len()), candle_core::DType::I64, &Device::Cpu)?;

    // Run model forward pass (need attention mask as third argument)
    let hidden_state = model.model.forward(
        &tokens_tensor,
        &token_type_ids,
        Some(&attention_tensor),
    )?;

    // Mean pooling
    let pooled = mean_pooling(&hidden_state, &attention_tensor)?;

    // L2 normalize embedding
    let normalized = l2_normalize(&pooled)?;

    // Convert to vec (remove batch dimension)
    let embedding = normalized.get(0)?.to_vec1::<f32>()?;

    Ok(embedding)
}

/// HTTP handler for embedding (supports both single and batch)
async fn embed_handler(
    model_state: axum::extract::State<Arc<EmbeddingModel>>,
    Json(req): Json<EmbedRequest>,
) -> Result<Json<EmbedResponse>, AppError> {
    let model = model_state.0.as_ref();

    // Determine texts to process (support both formats)
    let texts = match (&req.text, &req.texts) {
        (Some(text), None) => vec![text.clone()],
        (None, Some(texts)) => texts.clone(),
        (Some(text), Some(_)) => {
            // If both provided, prefer texts array (Python API behavior)
            req.texts.clone().unwrap_or_else(|| vec![text.clone()])
        }
        (None, None) => {
            return Err(AppError(anyhow::anyhow!("Either 'text' or 'texts' field is required")));
        }
    };

    // Process each text
    let mut embeddings = Vec::with_capacity(texts.len());
    for text in texts {
        let embedding = embed_text(model, &text)?;
        embeddings.push(embedding);
    }

    Ok(Json(EmbedResponse {
        embeddings,
        model: model_state.0.model_name.clone(),
        dimension: model_state.0.dimensions,
    }))
}

/// Health check handler
async fn health_handler(
    model_state: axum::extract::State<Arc<EmbeddingModel>>,
) -> Json<HealthResponse> {
    Json(HealthResponse {
        status: "ok".to_string(),
        model_loaded: true,
        model: model_state.0.model_name.clone(),
    })
}

/// Custom error type for better error handling
struct AppError(anyhow::Error);

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        (
            StatusCode::INTERNAL_SERVER_ERROR,
            format!("Error: {}", self.0),
        )
            .into_response()
    }
}

impl<E> From<E> for AppError
where
    E: Into<anyhow::Error>,
{
    fn from(err: E) -> Self {
        Self(err.into())
    }
}

/// Write port file for Go backend coordination
fn write_port_file(port: u16) -> Result<()> {
    // Use relative path for portability (works with release package)
    let port_file = PathBuf::from("embedding_port.txt");
    fs::write(&port_file, port.to_string())?;
    info!("Port file written: {} -> {}", port_file.display(), port);
    Ok(())
}

#[tokio::main]
async fn main() -> Result<()> {
    // Initialize logging
    tracing_subscriber::fmt::init();

    // Get model name from environment or use default
    // all-MiniLM-L6-v2: 22MB model, 384 dimensions, fast
    // bge-small-en-v1.5: 33MB model, 384 dimensions, good quality
    // nomic-embed-text-v1: 274MB model, 768 dimensions, best quality
    let model_name = env::var("EMBEDDING_MODEL")
        .unwrap_or_else(|_| "sentence-transformers/all-MiniLM-L6-v2".to_string());

    info!("Starting embedding server with model: {}", model_name);

    // Load model
    let model_state = load_model(&model_name).await?;

    // Get port from environment or use default
    let port = env::var("EMBEDDING_PORT")
        .unwrap_or_else(|_| "8082".to_string())
        .parse::<u16>()
        .context("Invalid port number")?;

    // Write port file for Go backend
    write_port_file(port)?;

    // Build router
    let app = Router::new()
        .route("/embed", post(embed_handler))
        .route("/health", get(health_handler))
        .route("/", get(|| async { "Embedding Server (Candle/Rust)" }))
        .with_state(model_state);

    // Start server
    let addr = SocketAddr::from(([0, 0, 0, 0], port));
    info!("Server listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await?;
    axum::serve(listener, app).await?;

    Ok(())
}