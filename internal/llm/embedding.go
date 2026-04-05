package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// EmbeddingClient 向量嵌入客户端接口
type EmbeddingClient interface {
	// GetEmbedding 获取文本的向量表示
	GetEmbedding(ctx context.Context, text string) ([]float64, error)
}

// OllamaEmbeddingRequest Ollama embedding 请求
type OllamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// OllamaEmbeddingResponse Ollama embedding 响应
type OllamaEmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

// GetEmbedding 获取文本的向量表示（Ollama 实现）
func (c *OllamaClient) GetEmbedding(ctx context.Context, text string) ([]float64, error) {
	url := c.config.BaseURL + "/api/embeddings"

	model := c.config.Models["embedding"]
	if model == "" {
		model = "embeddinggemma:latest"
	}

	req := OllamaEmbeddingRequest{
		Model:  model,
		Prompt: text,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("Ollama embedding error: %s - %s", httpResp.Status, string(respBody))
	}

	var resp OllamaEmbeddingResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, err
	}

	if len(resp.Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	return resp.Embedding, nil
}

// OpenAIEmbeddingRequest OpenAI embedding 请求
type OpenAIEmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// OpenAIEmbeddingResponse OpenAI embedding 响应
type OpenAIEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Index     int       `json:"index"`
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// GetEmbedding 获取文本的向量表示（OpenAI 实现）
func (c *OpenAIClient) GetEmbedding(ctx context.Context, text string) ([]float64, error) {
	baseURL := c.config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	url := baseURL + "/embeddings"

	model := c.config.Models["embedding"]
	if model == "" {
		model = "text-embedding-3-small"
	}

	req := OpenAIEmbeddingRequest{
		Model: model,
		Input: text,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("OpenAI embedding error: %s - %s", httpResp.Status, string(respBody))
	}

	var resp OpenAIEmbeddingResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 || len(resp.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	return resp.Data[0].Embedding, nil
}