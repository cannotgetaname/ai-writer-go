#!/usr/bin/env python3
"""
轻量级 Embedding 服务
- 自动下载模型（支持 HuggingFace 换源）
- 动态端口选择
- FastAPI HTTP API
"""

import os
import sys
import socket
import json
import logging
import threading
from pathlib import Path
from typing import List

from fastapi import FastAPI, HTTPException
from fastapi.responses import JSONResponse
from pydantic import BaseModel
import uvicorn

# 配置日志
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# 模型配置
MODEL_NAME = "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2"
MODEL_DIR = Path(__file__).parent.absolute() / "models"
PORT_FILE = Path(__file__).parent.absolute() / "embedding_port.txt"

# Input validation limits
MAX_TEXTS_PER_REQUEST = 1000
MAX_TOTAL_SIZE_BYTES = 100 * 1024  # 100KB

# 全局模型实例和线程锁
model = None
model_lock = threading.Lock()
model_loaded = False


class EmbedRequest(BaseModel):
    texts: List[str]


class EmbedResponse(BaseModel):
    embeddings: List[List[float]]
    model: str
    dimension: int


class HealthResponse(BaseModel):
    status: str
    model_loaded: bool


class ModelInfoResponse(BaseModel):
    model_name: str
    model_size_mb: float
    dimension: int


def find_available_port(start_port: int = 8082, max_attempts: int = 100) -> int:
    """查找可用端口"""
    for port in range(start_port, start_port + max_attempts):
        try:
            with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
                s.bind(('127.0.0.1', port))
                return port
        except OSError:
            continue
    raise RuntimeError("No available port found")


def write_port_file(port: int):
    """写入端口文件"""
    PORT_FILE.write_text(str(port))
    logger.info(f"Port file written: {PORT_FILE} = {port}")


def download_model():
    """下载模型，支持换源"""
    from sentence_transformers import SentenceTransformer

    # 确保模型目录存在
    MODEL_DIR.mkdir(parents=True, exist_ok=True)

    # 本地模型路径格式：sentence-transformers--paraphrase-multilingual-MiniLM-L12-v2
    local_model_name = MODEL_NAME.replace("/", "--")
    local_path = MODEL_DIR / local_model_name

    # 检查本地是否已有模型
    if local_path.exists() and any(local_path.iterdir()):
        logger.info(f"Loading model from local cache: {local_path}")
        try:
            return SentenceTransformer(str(local_path))
        except Exception as e:
            logger.warning(f"Failed to load local model: {e}")

    # 尝试 HuggingFace 官方源
    logger.info(f"Downloading model from HuggingFace: {MODEL_NAME}")
    try:
        return SentenceTransformer(
            MODEL_NAME,
            cache_folder=str(MODEL_DIR)
        )
    except Exception as e:
        logger.warning(f"HuggingFace download failed: {e}")

    # 切换到 hf-mirror.com
    logger.info("Switching to hf-mirror.com")
    os.environ['HF_ENDPOINT'] = 'https://hf-mirror.com'
    try:
        return SentenceTransformer(
            MODEL_NAME,
            cache_folder=str(MODEL_DIR)
        )
    except Exception as e:
        logger.error(f"hf-mirror download failed: {e}")
        raise RuntimeError(f"Failed to download model from both sources: {e}")


# FastAPI 应用
app = FastAPI(title="Embedding Service", version="1.0.0")


@app.on_event("startup")
async def startup_event():
    """启动时加载模型"""
    global model, model_loaded
    logger.info("Starting embedding service...")

    # 下载/加载模型
    try:
        with model_lock:
            model = download_model()
            model_loaded = True
        logger.info(f"Model loaded successfully, dimension: {model.get_sentence_embedding_dimension()}")
    except Exception as e:
        logger.error(f"Failed to load model: {e}")
        raise RuntimeError(f"Failed to load embedding model: {e}")


def validate_embed_request(request: EmbedRequest):
    """验证嵌入请求的输入"""
    if not request.texts:
        raise HTTPException(status_code=400, detail="texts array is empty")

    if len(request.texts) > MAX_TEXTS_PER_REQUEST:
        raise HTTPException(
            status_code=400,
            detail=f"Too many texts: {len(request.texts)} exceeds maximum of {MAX_TEXTS_PER_REQUEST}"
        )

    # Check total size
    total_size = sum(len(text.encode('utf-8')) for text in request.texts)
    if total_size > MAX_TOTAL_SIZE_BYTES:
        raise HTTPException(
            status_code=400,
            detail=f"Total text size {total_size} bytes exceeds maximum of {MAX_TOTAL_SIZE_BYTES} bytes"
        )


@app.post("/embed", response_model=EmbedResponse)
async def embed_texts(request: EmbedRequest):
    """获取文本向量"""
    global model_loaded

    if not model_loaded or model is None:
        raise HTTPException(status_code=503, detail="Model not loaded")

    # Validate input
    validate_embed_request(request)

    try:
        with model_lock:
            embeddings = model.encode(request.texts, convert_to_numpy=True)
        return EmbedResponse(
            embeddings=[emb.tolist() for emb in embeddings],
            model=MODEL_NAME.split("/")[-1],
            dimension=len(embeddings[0])
        )
    except Exception as e:
        logger.error(f"Embedding error: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/health", response_model=HealthResponse)
async def health_check():
    """健康检查"""
    return HealthResponse(
        status="ok",
        model_loaded=model_loaded
    )


@app.get("/model/info", response_model=ModelInfoResponse)
async def model_info():
    """模型信息"""
    if not model_loaded or model is None:
        raise HTTPException(status_code=503, detail="Model not loaded")

    # 计算模型目录大小
    model_size = 0
    local_model_name = MODEL_NAME.replace("/", "--")
    local_path = MODEL_DIR / local_model_name
    if local_path.exists():
        for f in local_path.rglob("*"):
            if f.is_file():
                model_size += f.stat().st_size

    return ModelInfoResponse(
        model_name=MODEL_NAME.split("/")[-1],
        model_size_mb=round(model_size / (1024 * 1024), 2),
        dimension=model.get_sentence_embedding_dimension()
    )


def main():
    """主入口"""
    # 查找可用端口（只调用一次）
    port = find_available_port()

    # 写入端口文件
    write_port_file(port)

    # 启动服务
    logger.info(f"Starting embedding service on port {port}")
    uvicorn.run(
        app,
        host="127.0.0.1",
        port=port,
        log_level="info"
    )


if __name__ == "__main__":
    main()