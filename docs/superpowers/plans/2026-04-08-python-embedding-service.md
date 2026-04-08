# Python Embedding Service Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 创建轻量级 Python embedding 服务，替代 TEI，支持自动模型下载和换源，与 Go 后端一体化启动。

**Architecture:** Python FastAPI 服务（PyInstaller 打包）+ Go 后端 HTTP 客户端 + 启动脚本协调端口。模型运行时下载，支持 HuggingFace 换源。

**Tech Stack:** FastAPI, sentence-transformers, uvicorn, PyInstaller (Python); Go HTTP client (Go)

---

## File Structure

| 文件 | 说明 |
|------|------|
| `embedding_service/embedding_server.py` | Python embedding 服务主程序 |
| `embedding_service/requirements.txt` | Python 依赖 |
| `embedding_service/build.py` | PyInstaller 打包脚本 |
| `internal/llm/embedding_python.go` | Go Python embedding 客户端 |
| `internal/api/router.go` | 修改：添加 python provider 支持和等待逻辑 |
| `internal/config/config.go` | 修改：添加 EmbeddingPythonConfig |
| `scripts/start_with_embedding.sh` | Linux 启动脚本 |
| `scripts/start_with_embedding.bat` | Windows 启动脚本 |

---

### Task 1: Python Embedding 服务主程序

**Files:**
- Create: `embedding_service/embedding_server.py`

- [ ] **Step 1: 创建 embedding_service 目录**

```bash
mkdir -p embedding_service
```

- [ ] **Step 2: 编写 Python embedding 服务主程序**

创建文件 `embedding_service/embedding_server.py`:

```python
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

# 全局模型实例
model = None


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
    global model
    logger.info("Starting embedding service...")

    # 查找可用端口并写入文件
    port = find_available_port()
    write_port_file(port)

    # 下载/加载模型
    try:
        model = download_model()
        logger.info(f"Model loaded successfully, dimension: {model.get_sentence_embedding_dimension()}")
    except Exception as e:
        logger.error(f"Failed to load model: {e}")
        sys.exit(1)


@app.post("/embed", response_model=EmbedResponse)
async def embed_texts(request: EmbedRequest):
    """获取文本向量"""
    if model is None:
        raise HTTPException(status_code=503, detail="Model not loaded")

    if not request.texts:
        raise HTTPException(status_code=400, detail="texts array is empty")

    try:
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
        model_loaded=model is not None
    )


@app.get("/model/info", response_model=ModelInfoResponse)
async def model_info():
    """模型信息"""
    if model is None:
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
    # 查找可用端口
    port = find_available_port()

    # 写入端口文件（startup_event 也会写，但这里先写确保文件存在）
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
```

- [ ] **Step 3: 测试 Python 服务启动**

```bash
cd embedding_service
python embedding_server.py
```

Expected: 服务启动，打印端口信息，开始下载模型

- [ ] **Step 4: 测试 API 接口**

```bash
# 健康检查
curl http://127.0.0.1:8082/health

# 向量生成
curl -X POST http://127.0.0.1:8082/embed \
  -H "Content-Type: application/json" \
  -d '{"texts": ["测试文本", "另一段文本"]}'
```

Expected: 返回 embeddings 数组和维度信息

- [ ] **Step 5: Commit**

```bash
git add embedding_service/embedding_server.py
git commit -m "feat: add Python embedding service with auto model download"
```

---

### Task 2: Python 依赖和打包脚本

**Files:**
- Create: `embedding_service/requirements.txt`
- Create: `embedding_service/build.py`

- [ ] **Step 1: 创建 requirements.txt**

创建文件 `embedding_service/requirements.txt`:

```
sentence-transformers==2.7.0
fastapi==0.110.0
uvicorn==0.27.1
pydantic==2.6.1
torch==2.2.0
pyinstaller==6.4.0
```

- [ ] **Step 2: 创建 build.py 打包脚本**

创建文件 `embedding_service/build.py`:

```python
#!/usr/bin/env python3
"""
PyInstaller 打包脚本
打包 embedding_server 为独立可执行文件
"""

import PyInstaller.__main__
import sys
from pathlib import Path

# 确保在正确目录
script_dir = Path(__file__).parent.absolute()
main_script = script_dir / "embedding_server.py"

print(f"Building embedding_server from {main_script}")

PyInstaller.__main__.run([
    str(main_script),
    '--name=embedding_server',
    '--onefile',
    '--clean',
    '--noconfirm',
    # 收集 sentence_transformers 数据文件
    '--collect-data=sentence_transformers',
    # 隐藏导入
    '--hidden-import=sentence_transformers',
    '--hidden-import=sentence_transformers.models',
    '--hidden-import=fastapi',
    '--hidden-import=uvicorn',
    '--hidden-import=uvicorn.logging',
    '--hidden-import=uvicorn.loops',
    '--hidden-import=uvicorn.loops.auto',
    '--hidden-import=uvicorn.protocols',
    '--hidden-import=uvicorn.protocols.http',
    '--hidden-import=uvicorn.protocols.http.auto',
    '--hidden-import=uvicorn.protocols.websockets',
    '--hidden-import=uvicorn.protocols.websockets.auto',
    '--hidden-import=uvicorn.protocols.websockets.websockets_impl',
    '--hidden-import=uvicorn.lifespan',
    '--hidden-import=uvicorn.lifespan.on',
    '--hidden-import=pydantic',
    '--hidden-import=huggingface_hub',
    '--hidden-import=transformers',
    # 排除不必要的模块以减小体积
    '--exclude-module=tkinter',
    '--exclude-module=PIL',
    '--exclude-module=matplotlib',
    '--exclude-module=numpy.f2py',
    # 工作目录
    '--distpath', str(script_dir / 'dist'),
    '--workpath', str(script_dir / 'build'),
    '--specpath', str(script_dir),
])

print("\nBuild complete!")
print(f"Output: {script_dir / 'dist' / 'embedding_server'}")

# Windows 下提示
if sys.platform == 'win32':
    print("\nNote: On Windows, the output file will be embedding_server.exe")
```

- [ ] **Step 3: 安装依赖**

```bash
cd embedding_service
pip install -r requirements.txt
```

- [ ] **Step 4: 执行打包（可选，需要时间）**

```bash
cd embedding_service
python build.py
```

Expected: 生成 `dist/embedding_server` (Linux) 或 `dist/embedding_server.exe` (Windows)

- [ ] **Step 5: Commit**

```bash
git add embedding_service/requirements.txt embedding_service/build.py
git commit -m "feat: add PyInstaller build configuration for embedding service"
```

---

### Task 3: Go Python Embedding 客户端

**Files:**
- Create: `internal/llm/embedding_python.go`

- [ ] **Step 1: 创建 Python embedding 客户端**

创建文件 `internal/llm/embedding_python.go`:

```go
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// PythonEmbeddingClient 调用独立 Python embedding 服务
type PythonEmbeddingClient struct {
	baseURL    string
	httpClient *http.Client
}

// pythonEmbedRequest Python 服务请求
type pythonEmbedRequest struct {
	Texts []string `json:"texts"`
}

// pythonEmbedResponse Python 服务响应
type pythonEmbedResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
	Model      string      `json:"model"`
	Dimension  int         `json:"dimension"`
}

// pythonHealthResponse Python 健康检查响应
type pythonHealthResponse struct {
	Status       string `json:"status"`
	ModelLoaded  bool   `json:"model_loaded"`
}

// NewPythonEmbeddingClient 创建 Python embedding 客户端
// portFile: 端口文件路径，如 ./embedding_port.txt
func NewPythonEmbeddingClient(portFile string) *PythonEmbeddingClient {
	baseURL := readPortFromFile(portFile)
	return &PythonEmbeddingClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// NewPythonEmbeddingClientFromURL 直接从 URL 创建客户端
func NewPythonEmbeddingClientFromURL(baseURL string) *PythonEmbeddingClient {
	return &PythonEmbeddingClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// readPortFromFile 从端口文件读取端口并构建 URL
func readPortFromFile(portFile string) string {
	// 尝试多个路径
	paths := []string{
		portFile,
		"./embedding_port.txt",
		filepath.Join(filepath.Dir(os.Args[0]), "embedding_port.txt"),
	}

	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err == nil {
			port := strconv.TrimSpace(string(content))
			return fmt.Sprintf("http://127.0.0.1:%s", port)
		}
	}

	// 默认端口
	return "http://127.0.0.1:8082"
}

// WaitForReady 等待服务就绪
func (c *PythonEmbeddingClient) WaitForReady(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("embedding service not ready after %v", timeout)
		case <-ticker.C:
			if c.checkHealth(ctx) {
				return nil
			}
		}
	}
}

// checkHealth 检查健康状态
func (c *PythonEmbeddingClient) checkHealth(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
	if err != nil {
		return false
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	var health pythonHealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return false
	}

	return health.Status == "ok" && health.ModelLoaded
}

// GetEmbedding 获取单个文本的向量表示
func (c *PythonEmbeddingClient) GetEmbedding(ctx context.Context, text string) ([]float64, error) {
	req := pythonEmbedRequest{Texts: []string{text}}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/embed", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("Python embedding error: %s - %s", httpResp.Status, string(respBody))
	}

	var resp pythonEmbedResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(resp.Embeddings) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	return resp.Embeddings[0], nil
}

// GetEmbeddings 批量获取向量（可选优化）
func (c *PythonEmbeddingClient) GetEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	req := pythonEmbedRequest{Texts: texts}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/embed", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("Python embedding error: %s - %s", httpResp.Status, string(respBody))
	}

	var resp pythonEmbedResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return resp.Embeddings, nil
}
```

- [ ] **Step 2: 更新 NewEmbeddingClient 工厂函数**

修改 `internal/llm/embedding.go` 的 `NewEmbeddingClient` 函数，添加 `python` provider 支持：

```go
// NewEmbeddingClient 创建 Embedding 客户端（工厂函数）
func NewEmbeddingClient(provider, baseURL, apiKey, model string) EmbeddingClient {
	switch provider {
	case "python":
		// 从端口文件读取或直接使用 URL
		if baseURL != "" {
			return NewPythonEmbeddingClientFromURL(baseURL)
		}
		return NewPythonEmbeddingClient("./embedding_port.txt")
	case "tei":
		return NewTEIEmbeddingClient(baseURL)
	case "ollama":
		llmCfg := &Config{
			BaseURL: baseURL,
			APIKey:  apiKey,
			Models:  map[string]string{"embedding": model},
		}
		return NewOllamaClient(llmCfg)
	case "openai", "deepseek":
		llmCfg := &Config{
			BaseURL: baseURL,
			APIKey:  apiKey,
			Models:  map[string]string{"embedding": model},
		}
		return NewOpenAIClient(llmCfg)
	default:
		// 默认使用 Python 服务
		return NewPythonEmbeddingClient("./embedding_port.txt")
	}
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/llm/embedding_python.go internal/llm/embedding.go
git commit -m "feat: add Python embedding client for standalone service"
```

---

### Task 4: Go Router 集成等待逻辑

**Files:**
- Modify: `internal/api/router.go`

- [ ] **Step 1: 修改 router.go 添加 Python 服务等待逻辑**

在 `SetupRouter` 函数中，修改 embedding 初始化部分：

找到现有代码（约第 32-39 行）:
```go
// 初始化 embedding 客户端
embeddingClient := llm.NewEmbeddingClient(
    cfg.Embedding.Provider,
    cfg.Embedding.BaseURL,
    cfg.Embedding.APIKey,
    cfg.Embedding.Model,
)
handler.InitEmbeddingClient(embeddingClient)
```

替换为：
```go
// 初始化 embedding 客户端
embeddingClient := llm.NewEmbeddingClient(
    cfg.Embedding.Provider,
    cfg.Embedding.BaseURL,
    cfg.Embedding.APIKey,
    cfg.Embedding.Model,
)

// 如果是 Python provider，等待服务就绪
if cfg.Embedding.Provider == "python" || cfg.Embedding.Provider == "" {
    pythonClient, ok := embeddingClient.(*llm.PythonEmbeddingClient)
    if ok {
        log.Println("Waiting for Python embedding service...")
        if err := pythonClient.WaitForReady(30 * time.Second); err != nil {
            log.Fatalf("Embedding service not ready: %v", err)
        }
        log.Println("Python embedding service ready")
    }
}

handler.InitEmbeddingClient(embeddingClient)
```

- [ ] **Step 2: 更新 config.go 默认配置**

修改 `internal/config/config.go` 的 `defaultConfig`，将默认 provider 改为 `python`：

```go
Embedding: EmbeddingConfig{
    Provider: "python",
    Model:    "",  // Python 服务固定使用 paraphrase-multilingual-MiniLM-L12-v2
    BaseURL:  "",  // 从端口文件读取
},
```

- [ ] **Step 3: Commit**

```bash
git add internal/api/router.go internal/config/config.go
git commit -m "feat: add Python embedding service wait logic in router"
```

---

### Task 5: Linux 启动脚本

**Files:**
- Create: `scripts/start_with_embedding.sh`

- [ ] **Step 1: 创建 Linux 启动脚本**

创建文件 `scripts/start_with_embedding.sh`:

```bash
#!/bin/bash
# 同时启动 Python embedding 服务和 Go 后端

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "Starting AI Writer with Python embedding service..."

# 1. 清理旧的端口文件
rm -f "$SCRIPT_DIR/embedding_port.txt"

# 2. 检查 embedding_server 是否存在
EMBEDDING_SERVER="$SCRIPT_DIR/embedding_server"
if [ ! -f "$EMBEDDING_SERVER" ]; then
    EMBEDDING_SERVER="$SCRIPT_DIR/dist/embedding_server"
fi

if [ ! -f "$EMBEDDING_SERVER" ]; then
    echo "Error: embedding_server not found in $SCRIPT_DIR"
    echo "Please run: cd embedding_service && python build.py"
    exit 1
fi

# 3. 启动 Python embedding 服务（后台）
echo "Starting embedding service..."
"$EMBEDDING_SERVER" &
PYTHON_PID=$!

# 4. 等待端口文件生成（最多 60 秒）
echo "Waiting for embedding service to start..."
WAIT_COUNT=0
MAX_WAIT=60
while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
    if [ -f "$SCRIPT_DIR/embedding_port.txt" ]; then
        PORT=$(cat "$SCRIPT_DIR/embedding_port.txt")
        echo "Embedding service started on port $PORT"
        break
    fi
    sleep 1
    WAIT_COUNT=$((WAIT_COUNT + 1))
    echo -n "."
done
echo ""

if [ ! -f "$SCRIPT_DIR/embedding_port.txt" ]; then
    echo "Error: Embedding service failed to start (timeout after ${MAX_WAIT}s)"
    kill $PYTHON_PID 2>/dev/null || true
    exit 1
fi

# 5. 启动 Go 后端
AI_WRITER="$SCRIPT_DIR/ai-writer"
if [ ! -f "$AI_WRITER" ]; then
    # 尝试从上级目录查找
    AI_WRITER="$SCRIPT_DIR/../ai-writer"
fi

if [ ! -f "$AI_WRITER" ]; then
    echo "Error: ai-writer not found"
    kill $PYTHON_PID 2>/dev/null || true
    exit 1
fi

echo "Starting AI Writer backend..."
"$AI_WRITER" server

# 6. Go 后端退出时，清理 Python 进程
echo "Shutting down..."
kill $PYTHON_PID 2>/dev/null || true
rm -f "$SCRIPT_DIR/embedding_port.txt"
```

- [ ] **Step 2: 设置脚本权限**

```bash
chmod +x scripts/start_with_embedding.sh
```

- [ ] **Step 3: Commit**

```bash
git add scripts/start_with_embedding.sh
git commit -m "feat: add Linux startup script for embedding service"
```

---

### Task 6: Windows 启动脚本

**Files:**
- Create: `scripts/start_with_embedding.bat`

- [ ] **Step 1: 创建 Windows 启动脚本**

创建文件 `scripts/start_with_embedding.bat`:

```batch
@echo off
REM 同时启动 Python embedding 服务和 Go 后端

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set SCRIPT_DIR=%SCRIPT_DIR:~0,-1%

echo Starting AI Writer with Python embedding service...

REM 1. 清理旧的端口文件
if exist "%SCRIPT_DIR%\embedding_port.txt" del "%SCRIPT_DIR%\embedding_port.txt" 2>nul

REM 2. 检查 embedding_server 是否存在
set EMBEDDING_SERVER=%SCRIPT_DIR%\embedding_server.exe
if not exist "%EMBEDDING_SERVER%" set EMBEDDING_SERVER=%SCRIPT_DIR%\dist\embedding_server.exe

if not exist "%EMBEDDING_SERVER%" (
    echo Error: embedding_server.exe not found in %SCRIPT_DIR%
    echo Please run: cd embedding_service && python build.py
    exit /b 1
)

REM 3. 启动 Python embedding 服务（后台）
echo Starting embedding service...
start /B "" "%EMBEDDING_SERVER%" >nul 2>&1

REM 4. 等待端口文件生成（最多 60 秒）
echo Waiting for embedding service to start...
set COUNT=0
set MAX_WAIT=60

:wait_loop
if exist "%SCRIPT_DIR%\embedding_port.txt" goto got_port
set /a COUNT+=1
if !COUNT! geq %MAX_WAIT% goto timeout_error
timeout /t 1 /nobreak >nul
goto wait_loop

:timeout_error
echo Error: Embedding service failed to start (timeout after %MAX_WAIT%s)
taskkill /F /IM embedding_server.exe 2>nul
exit /b 1

:got_port
set /p PORT=<"%SCRIPT_DIR%\embedding_port.txt"
echo Embedding service started on port !PORT!

REM 5. 启动 Go 后端
set AI_WRITER=%SCRIPT_DIR%\ai-writer.exe
if not exist "%AI_WRITER%" set AI_WRITER=%SCRIPT_DIR%\..\ai-writer.exe

if not exist "%AI_WRITER%" (
    echo Error: ai-writer.exe not found
    taskkill /F /IM embedding_server.exe 2>nul
    exit /b 1
)

echo Starting AI Writer backend...
"%AI_WRITER%" server

REM 6. Go 后端退出时，清理 Python 进程
echo Shutting down...
taskkill /F /IM embedding_server.exe 2>nul
if exist "%SCRIPT_DIR%\embedding_port.txt" del "%SCRIPT_DIR%\embedding_port.txt" 2>nul

endlocal
```

- [ ] **Step 2: Commit**

```bash
git add scripts/start_with_embedding.bat
git commit -m "feat: add Windows startup script for embedding service"
```

---

### Task 7: 集成测试

**Files:**
- None (测试验证)

- [ ] **Step 1: 测试 Python 服务独立运行**

```bash
cd embedding_service
python embedding_server.py
```

验证：
- 服务启动成功
- 模型自动下载（首次运行）
- `/health` 返回 `{"status": "ok", "model_loaded": true}`
- `/embed` 返回正确向量

- [ ] **Step 2: 测试启动脚本（打包后）**

```bash
# 先打包
cd embedding_service && python build.py

# 使用启动脚本
./scripts/start_with_embedding.sh
```

验证：
- Python 服务启动
- 端口文件生成
- Go 后端启动并等待成功
- 服务正常运行

- [ ] **Step 3: 测试向量功能**

```bash
# 通过 API 测试向量索引
curl -X POST http://localhost:8081/api/vector/index \
  -H "Content-Type: application/json" \
  -d '{"book_name": "test_book"}'

# 测试向量搜索
curl -X POST http://localhost:8081/api/vector/search \
  -H "Content-Type: application/json" \
  -d '{"book_name": "test_book", "query": "测试查询", "top_k": 5}'
```

---

## Self-Review Checklist

**1. Spec Coverage:**
- [x] Python embedding 服务开发 → Task 1
- [x] PyInstaller 打包配置 → Task 2
- [x] Go 后端 Python 客户端 → Task 3
- [x] Go 后端启动等待逻辑 → Task 4
- [x] Linux 启动脚本 → Task 5
- [x] Windows 启动脚本 → Task 6
- [x] 集成测试 → Task 7

**2. Placeholder Scan:**
- 无 TBD/TODO
- 无 "Add appropriate error handling"
- 无 "Write tests for the above"
- 无 "Similar to Task N"
- 所有代码步骤都有完整代码

**3. Type Consistency:**
- `PythonEmbeddingClient` 实现了 `EmbeddingClient` 接口（GetEmbedding 方法）
- `pythonEmbedRequest.Texts` 与 API 请求格式匹配
- `pythonEmbedResponse.Embeddings` 类型为 `[][]float64`

---

## 验收标准

1. Python embedding 服务能独立运行，自动下载模型
2. Go 后端启动时等待 Python 服务就绪
3. 首次运行时模型能自动下载（支持 HF 换源）
4. 向量索引和搜索功能正常工作
5. 启动脚本能同时启动两个服务
6. 打包后的体积符合预期（约 100-150MB + 118MB 模型）