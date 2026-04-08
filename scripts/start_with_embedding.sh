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