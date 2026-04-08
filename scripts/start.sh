#!/bin/bash

# start.sh - 启动脚本

# 获取脚本所在目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# 默认配置
TEI_PORT=${TEI_PORT:-8081}
MODEL_ID=${MODEL_ID:-"BAAI/bge-base-zh-v1.5"}

echo "=== AI Writer 启动 ==="

# 检测模型是否已下载
MODEL_DIR="$HOME/.cache/huggingface/hub/models--BAAI--bge-base-zh-v1.5"
if [ ! -d "$MODEL_DIR" ]; then
    echo "首次运行，正在下载模型 (约 400MB)..."
    echo "模型将保存到: $MODEL_DIR"
fi

# 启动 TEI 服务
echo "启动 Embedding 服务 (端口: $TEI_PORT)..."
./text-embeddings-router \
    --model-id "$MODEL_ID" \
    --port "$TEI_PORT" \
    --dtype float16 &

TEI_PID=$!
echo "TEI PID: $TEI_PID"

# 保存 PID
echo $TEI_PID > .tei.pid

# 等待 TEI 就绪
echo "等待 Embedding 服务就绪..."
for i in {1..60}; do
    if curl -s "http://127.0.0.1:$TEI_PORT/health" > /dev/null 2>&1; then
        echo "Embedding 服务就绪"
        break
    fi
    if [ $i -eq 60 ]; then
        echo "Embedding 服务启动超时"
        kill $TEI_PID 2>/dev/null
        exit 1
    fi
    sleep 1
done

# 启动主服务
echo "启动 AI Writer..."
./ai-writer server

# 清理
kill $TEI_PID 2>/dev/null
rm -f .tei.pid