# Python Embedding 服务设计文档

> **日期**: 2026-04-08
> **状态**: 待用户审核

---

## 背景

当前 TEI (Text Embeddings Inference) embedding 方案存在以下问题：
- 体积过大（2GB+）
- 必须依赖 Docker
- 对轻量级部署不友好

需要一个轻量级、无 Docker 依赖的 embedding 服务方案，支持几百万字长篇网文创作。

---

## 目标

- 轻量级 embedding 服务（打包体积约 100-150MB）
- 无 Docker 依赖，纯 Python 可执行文件
- 自动模型管理（下载、换源、缓存）
- 支持 Go 后端一体化启动体验
- 支持百万字级别索引（上万个文本块）

---

## 架构概览

```
┌─────────────────────────────────────────────────────────────┐
│                    启动脚本 start.sh                          │
│  1. 启动 Python embedding 服务 (动态端口)                      │
│  2. 等待端口文件生成                                          │
│  3. 启动 Go 后端                                              │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────┐    ┌──────────────────────┐
│   Go Backend         │    │  Python Embedding    │
│   (端口 8081)        │◄──►│  Service (动态端口)   │
│                      │    │                      │
│  - 检查服务就绪      │    │  - 自动下载模型       │
│  - HTTP 调用 embed   │    │  - 写端口到临时文件   │
│  - 向量索引/搜索     │    │  - HTTP embed API    │
└──────────────────────┘    └──────────────────────┘
                              │
                              ▼
                    ┌──────────────────────┐
                    │  模型缓存目录         │
                    │  ./models/           │
                    │                      │
                    │  自动换源下载          │
                    │  (HF → hf-mirror)    │
                    └──────────────────────┘
```

---

## 模型选择

**模型**: `sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2`

**特点**:
- 多语言支持（中文效果不错）
- 模型大小约 118MB
- 384 维向量，计算速度快
- 适合百万字级别索引

---

## Python Embedding 服务设计

### 文件结构

```
embedding_service/
├── embedding_server.py      # 主服务代码
├── build.py                 # PyInstaller 打包脚本
├── requirements.txt         # Python 依赖
└── models/                  # 模型缓存目录（运行时填充）
```

### API 接口

#### POST /embed

请求：
```json
{
  "texts": ["文本1", "文本2", ...]
}
```

响应：
```json
{
  "embeddings": [[0.1, 0.2, ...], [0.3, 0.4, ...], ...],
  "model": "paraphrase-multilingual-MiniLM-L12-v2",
  "dimension": 384
}
```

#### GET /health

响应：`{"status": "ok", "model_loaded": true}`

#### GET /model/info

响应：`{"model_name": "...", "model_size_mb": 118, "dimension": 384}`

### 模型下载逻辑

```python
MODEL_NAME = "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2"
MODEL_DIR = "./models"  # 程序所在目录

def download_model():
    # 1. 检查本地是否已有模型
    local_path = os.path.join(MODEL_DIR, MODEL_NAME.replace("/", "--"))
    if os.path.exists(local_path):
        return local_path

    # 2. 尝试 HuggingFace 官方源
    try:
        return SentenceTransformer(MODEL_NAME, cache_folder=MODEL_DIR)
    except:
        pass

    # 3. 切换到 hf-mirror.com
    os.environ['HF_ENDPOINT'] = 'https://hf-mirror.com'
    return SentenceTransformer(MODEL_NAME, cache_folder=MODEL_DIR)
```

### 端口协调

- Python 服务启动时选择可用端口（从 8082 开始尝试）
- 将端口写入 `embedding_port.txt` 文件
- Go 后端读取该文件获取端口地址

---

## Go 后端集成设计

### 新增文件

```
internal/llm/embedding_python.go    # Python 服务客户端实现
scripts/start_with_embedding.sh     # 启动脚本（Linux）
scripts/start_with_embedding.bat    # 启动脚本（Windows）
```

### Python Embedding 客户端实现

```go
// PythonEmbeddingClient 调用独立 Python embedding 服务
type PythonEmbeddingClient struct {
    baseURL    string  // 动态端口，如 http://127.0.0.1:xxxxx
    httpClient *http.Client
}

// 等待服务就绪，最多 30 秒
func (c *PythonEmbeddingClient) WaitForReady(timeout time.Duration) error {
    // 轮询 /health 接口
}

// GetEmbedding 获取单个文本向量
func (c *PythonEmbeddingClient) GetEmbedding(ctx context.Context, text string) ([]float64, error) {
    // POST /embed {"texts": [text]}
}
```

### 启动时等待逻辑

在 `router.go` 中：

```go
// 初始化 embedding 客户端
embeddingClient := llm.NewPythonEmbeddingClient(portFromFile)

// 等待服务就绪（30秒）
if err := embeddingClient.WaitForReady(30 * time.Second); err != nil {
    log.Fatalf("Embedding service not ready: %v", err)
}
```

---

## 启动脚本设计

### Linux 启动脚本 (start.sh)

```bash
#!/bin/bash
# 同时启动 Python embedding 服务和 Go 后端

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# 1. 清理旧的端口文件
rm -f "$SCRIPT_DIR/embedding_port.txt"

# 2. 启动 Python embedding 服务（后台）
"$SCRIPT_DIR/embedding_server" &
PYTHON_PID=$!

# 3. 等待端口文件生成（最多 60 秒）
for i in {1..60}; do
    if [ -f "$SCRIPT_DIR/embedding_port.txt" ]; then
        break
    fi
    sleep 1
done

if [ ! -f "$SCRIPT_DIR/embedding_port.txt" ]; then
    echo "Error: Embedding service failed to start"
    kill $PYTHON_PID 2>/dev/null
    exit 1
fi

# 4. 启动 Go 后端
"$SCRIPT_DIR/ai-writer" server

# 5. Go 后端退出时，清理 Python 进程
kill $PYTHON_PID 2>/dev/null
```

### Windows 启动脚本 (start.bat)

```batch
@echo off
REM 同时启动 Python embedding 服务和 Go 后端

set SCRIPT_DIR=%~dp0

REM 1. 清理旧的端口文件
del "%SCRIPT_DIR%embedding_port.txt" 2>nul

REM 2. 启动 Python embedding 服务（后台）
start /B "" "%SCRIPT_DIR%embedding_server.exe"

REM 3. 等待端口文件生成（最多 60 秒）
set COUNT=0
:wait_loop
if exist "%SCRIPT_DIR%embedding_port.txt" goto start_go
set /a COUNT+=1
if %COUNT% geq 60 goto error
timeout /t 1 /nobreak >nul
goto wait_loop

:error
echo Error: Embedding service failed to start
taskkill /F /IM embedding_server.exe 2>nul
exit /b 1

:start_go
REM 4. 启动 Go 后端
"%SCRIPT_DIR%ai-writer.exe" server

REM 5. 清理 Python 进程
taskkill /F /IM embedding_server.exe 2>nul
```

---

## PyInstaller 打包设计

### 打包脚本 (build.py)

```python
import PyInstaller.__main__

PyInstaller.__main__.run([
    'embedding_server.py',
    '--name=embedding_server',
    '--onefile',              # 打包成单个可执行文件
    '--clean',
    '--noconfirm',
    '--collect-data=sentence_transformers',
    '--hidden-import=sentence_transformers',
    '--hidden-import=fastapi',
    '--hidden-import=uvicorn',
])
```

### 依赖文件 (requirements.txt)

```
sentence-transformers==2.2.2
fastapi==0.104.1
uvicorn==0.24.0
pyinstaller==6.3.0  # 仅打包时需要
```

### 预期打包体积

- 不含模型：约 100-150MB（主要是 sentence-transformers 和 torch 基础库）
- 首次运行时下载模型：约 118MB

### 跨平台打包注意事项

- Linux 打包需要在 Linux 系统执行
- Windows 打包需要在 Windows 系统执行
- 仅保留 torch CPU 版本，不需要 GPU 支持

---

## 打包后的最终目录结构

```
ai-writer-release/
├── start.sh / start.bat        # 启动脚本
├── ai-writer / ai-writer.exe   # Go 后端
├── embedding_server / embedding_server.exe  # Python embedding 服务
├── models/                     # 模型缓存（运行时填充）
└── data/                       # 数据目录
```

---

## 实施计划概要

| 步骤 | 内容 | 文件 |
|------|------|------|
| 1 | Python embedding 服务开发 | `embedding_service/embedding_server.py` |
| 2 | PyInstaller 打包配置 | `embedding_service/build.py`, `requirements.txt` |
| 3 | Go 后端 Python 客户端实现 | `internal/llm/embedding_python.go` |
| 4 | Go 后端启动等待逻辑 | `internal/api/router.go` |
| 5 | 启动脚本开发 | `scripts/start_with_embedding.sh`, `scripts/start_with_embedding.bat` |
| 6 | 集成测试 | 完整流程验证 |

---

## 验收标准

1. Python embedding 服务能独立运行，自动下载模型
2. Go 后端启动时等待 Python 服务就绪
3. 首次运行时模型能自动下载（支持 HF 换源）
4. 向量索引和搜索功能正常工作
5. 启动脚本能同时启动两个服务
6. 打包后的体积符合预期（约 100-150MB + 118MB 模型）