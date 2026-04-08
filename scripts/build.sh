#!/bin/bash

# build.sh - 编译脚本
# 用法: ./scripts/build.sh [version]

VERSION=${1:-"dev"}
OUTPUT_DIR="release"

echo "=== AI Writer Build Script ==="
echo "Version: $VERSION"
echo "Output: $OUTPUT_DIR"

# 清理输出目录
rm -rf $OUTPUT_DIR
mkdir -p $OUTPUT_DIR

# 构建前端
echo "Building frontend..."
cd web
npm install
npm run build
cd ..

# 编译后端 (Linux amd64)
echo "Building Linux amd64..."
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$VERSION" -o $OUTPUT_DIR/ai-writer-linux-amd64 .

# 编译后端 (Linux arm64)
echo "Building Linux arm64..."
CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc go build -ldflags="-s -w -X main.Version=$VERSION" -o $OUTPUT_DIR/ai-writer-linux-arm64 .

# 编译后端 (Windows amd64)
echo "Building Windows amd64..."
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -ldflags="-s -w -X main.Version=$VERSION" -o $OUTPUT_DIR/ai-writer-windows-amd64.exe .

# 下载 TEI 二进制
echo "Downloading TEI binaries..."

# Linux amd64
wget -q -O $OUTPUT_DIR/text-embeddings-router-linux-amd64 \
  "https://github.com/huggingface/text-embeddings-inference/releases/latest/download/text-embeddings-router-linux-amd64"
chmod +x $OUTPUT_DIR/text-embeddings-router-linux-amd64

# Linux arm64
wget -q -O $OUTPUT_DIR/text-embeddings-router-linux-arm64 \
  "https://github.com/huggingface/text-embeddings-inference/releases/latest/download/text-embeddings-router-linux-arm64"
chmod +x $OUTPUT_DIR/text-embeddings-router-linux-arm64

# Windows amd64
wget -q -O $OUTPUT_DIR/text-embeddings-router-windows-amd64.exe \
  "https://github.com/huggingface/text-embeddings-inference/releases/latest/download/text-embeddings-router-windows-amd64.exe"

echo "Build complete!"
ls -la $OUTPUT_DIR/