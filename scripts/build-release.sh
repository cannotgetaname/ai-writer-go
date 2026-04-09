#!/bin/bash
# 打包 AI Writer 为可发布的一键启动包

set -e

VERSION=${1:-"1.0.0"}
DIST_DIR="dist/ai-writer-${VERSION}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "=== Building AI Writer Release v${VERSION} ==="

# 1. 创建发布目录
echo "Creating release directory..."
rm -rf "$PROJECT_ROOT/$DIST_DIR"
mkdir -p "$PROJECT_ROOT/$DIST_DIR"
mkdir -p "$PROJECT_ROOT/$DIST_DIR/data"

# 2. 编译 Go 后端
echo "Building Go backend..."
cd "$PROJECT_ROOT"
CGO_ENABLED=0 go build -ldflags="-s -w" -o "$DIST_DIR/ai-writer" .

# 3. 编译 Rust embedding 服务
echo "Building Rust embedding service..."
cd "$PROJECT_ROOT/embedding_service_rust"
if [ ! -f "target/release/embedding_server" ]; then
    echo "Embedding server not built, building now..."
    source ~/.cargo/env
    cargo build --release
fi
cp target/release/embedding_server "$PROJECT_ROOT/$DIST_DIR/"

# 4. 复制启动脚本
echo "Copying startup scripts..."
cp "$PROJECT_ROOT/scripts/start_with_embedding.sh" "$PROJECT_ROOT/$DIST_DIR/"
chmod +x "$PROJECT_ROOT/$DIST_DIR/start_with_embedding.sh"

# Windows 脚本也复制
if [ -f "$PROJECT_ROOT/scripts/start_with_embedding.bat" ]; then
    cp "$PROJECT_ROOT/scripts/start_with_embedding.bat" "$PROJECT_ROOT/$DIST_DIR/"
fi

# 5. 复制前端静态文件（如果有）
if [ -d "$PROJECT_ROOT/web/dist" ]; then
    echo "Copying web assets..."
    mkdir -p "$PROJECT_ROOT/$DIST_DIR/web/dist"
    cp -r "$PROJECT_ROOT/web/dist/"* "$PROJECT_ROOT/$DIST_DIR/web/dist/"
fi

# 6. 创建默认配置
echo "Creating default config..."
cat > "$PROJECT_ROOT/$DIST_DIR/config.yaml" << 'EOF'
server:
  port: "8081"
  data_dir: "./data"

llm:
  provider: "deepseek"
  api_key: ""
  base_url: ""
  models:
    writer: "deepseek-chat"
    architect: "deepseek-chat"
    reviewer: "deepseek-chat"
  temperatures:
    writer: 1.5
    architect: 1.0
    reviewer: 0.5

embedding:
  provider: "python"
  timeout: 30

vector_store:
  chunk_size: 500
  overlap: 100
EOF

# 7. 创建 README
echo "Creating README..."
cat > "$PROJECT_ROOT/$DIST_DIR/README.txt" << 'EOF'
AI Writer - 网文创作助手
========================

启动方式:
  Linux/Mac:   ./start_with_embedding.sh
  Windows:     start_with_embedding.bat

首次启动:
  - 自动下载 embedding 模型（约 22MB）
  - 中国用户会自动检测并使用镜像加速下载

配置文件:
  - 编辑 config.yaml 配置 LLM API
  - 支持 DeepSeek、OpenAI、Ollama 等提供商

数据目录:
  - ./data/ 存储所有书籍数据
EOF

# 8. 打包
echo "Creating archive..."
cd "$PROJECT_ROOT/dist"
tar -czvf "ai-writer-${VERSION}-linux-x64.tar.gz" "ai-writer-${VERSION}"

# 输出结果
echo ""
echo "=== Build Complete ==="
echo "Release directory: $DIST_DIR"
echo "Archive: dist/ai-writer-${VERSION}-linux-x64.tar.gz"
echo ""
echo "Contents:"
ls -la "$PROJECT_ROOT/$DIST_DIR/"