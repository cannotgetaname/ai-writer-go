#!/bin/bash

# start.sh - 启动脚本
# 用法: 在 release 目录中运行此脚本

# 默认配置
TEI_PORT=${TEI_PORT:-8081}
MODEL_ID=${MODEL_ID:-"BAAI/bge-base-zh-v1.5"}

echo "=== AI Writer 启动 ==="

# 检查必要文件
if [ ! -f "./ai-writer" ]; then
    echo "错误: 找不到 ai-writer"
    echo "请先运行 ./scripts/build.sh 构建项目，然后在 release 目录中运行此脚本"
    exit 1
fi

# 检查 Embedding 服务
check_embedding_service() {
    curl -s "http://127.0.0.1:$TEI_PORT/health" > /dev/null 2>&1
}

if check_embedding_service; then
    echo "Embedding 服务已运行 (端口: $TEI_PORT)"
else
    echo "Embedding 服务未运行"
    echo ""
    echo "请选择 Embedding 服务方式："
    echo "  1. Docker 运行 TEI (推荐)"
    echo "  2. 使用 Ollama"
    echo "  3. 使用云端 API (DeepSeek/OpenAI)"
    echo "  4. 跳过，稍后手动启动"
    echo ""
    read -p "请选择 [1-4]: " choice

    case $choice in
        1)
            if command -v docker &> /dev/null; then
                echo "启动 TEI Docker 容器..."
                docker run -d --name tei \
                    --gpus all \
                    -p $TEI_PORT:80 \
                    -v $HOME/.cache/huggingface:/data \
                    ghcr.io/huggingface/text-embeddings-inference:latest \
                    --model-id $MODEL_ID
                echo "等待 TEI 服务启动..."
                for i in {1..60}; do
                    if check_embedding_service; then
                        echo "TEI 服务就绪"
                        break
                    fi
                    sleep 1
                done
            else
                echo "错误: 未安装 Docker"
                echo "请先安装 Docker 或选择其他方式"
                exit 1
            fi
            ;;
        2)
            echo "请确保 Ollama 已运行并安装了 embedding 模型"
            echo "运行: ollama pull embeddinggemma"
            ;;
        3)
            echo "请在 config.yaml 中配置 API Key"
            ;;
        4)
            echo "已跳过。请手动启动 Embedding 服务后重试"
            ;;
        *)
            echo "无效选择"
            exit 1
            ;;
    esac
fi

# 启动主服务
echo "启动 AI Writer..."
./ai-writer server