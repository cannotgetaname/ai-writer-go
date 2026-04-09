#!/bin/bash
# 多平台构建脚本
# 用法: ./scripts/build-all-platforms.sh <version>
#
# 依赖:
# - Go 1.21+
# - Rust + cargo
# - Node.js 18+
# - 交叉编译工具链（可选）:
#   - Windows: sudo apt install mingw-w64
#   - ARM64: rustup target add aarch64-unknown-linux-gnu

set -e

VERSION=${1:-"1.3.0"}
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DIST_DIR="$PROJECT_ROOT/dist"

echo "=== Building AI Writer v${VERSION} ==="

# 清理旧的构建
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

# 构建前端
echo "Building frontend..."
cd "$PROJECT_ROOT/web"
npm install --silent 2>/dev/null
npm run build
cd "$PROJECT_ROOT"

# 加载 Rust 环境
source ~/.cargo/env 2>/dev/null || true

# ============================================
# 构建函数
# ============================================
build_package() {
    local OS=$1
    local ARCH=$2
    local GOOS=$3
    local GOARCH=$4
    local RUST_TARGET=$5
    local BINARY_EXT=${6:-""}
    local OUTPUT_NAME="ai-writer-${VERSION}-${OS}-${ARCH}"
    local OUTPUT_DIR="$DIST_DIR/$OUTPUT_NAME"

    echo ""
    echo "=== Building $OUTPUT_NAME ==="

    # 创建输出目录
    rm -rf "$OUTPUT_DIR"
    mkdir -p "$OUTPUT_DIR/data"
    mkdir -p "$OUTPUT_DIR/web/dist"

    # 编译 Go 后端
    echo "  Building Go backend..."
    cd "$PROJECT_ROOT"
    CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o "$OUTPUT_DIR/ai-writer${BINARY_EXT}" . 2>/dev/null

    # 编译 Rust embedding 服务
    echo "  Building Rust embedding service..."
    cd "$PROJECT_ROOT/embedding_service_rust"

    local RUST_BIN="target/release/embedding_server${BINARY_EXT}"
    if [ -n "$RUST_TARGET" ] && [ "$RUST_TARGET" != "native" ]; then
        rustup target add "$RUST_TARGET" 2>/dev/null || true
        cargo build --release --target "$RUST_TARGET" 2>/dev/null
        RUST_BIN="target/$RUST_TARGET/release/embedding_server${BINARY_EXT}"
    else
        cargo build --release 2>/dev/null
    fi

    if [ -f "$RUST_BIN" ]; then
        cp "$RUST_BIN" "$OUTPUT_DIR/"
    else
        echo "  ⚠️  Rust binary not found, skipping..."
        return 1
    fi

    # 复制启动脚本
    cp "$PROJECT_ROOT/scripts/start_with_embedding.sh" "$OUTPUT_DIR/" 2>/dev/null || true
    cp "$PROJECT_ROOT/scripts/start_with_embedding.bat" "$OUTPUT_DIR/" 2>/dev/null || true
    chmod +x "$OUTPUT_DIR/start_with_embedding.sh" 2>/dev/null || true

    # 复制前端
    cp -r "$PROJECT_ROOT/web/dist/"* "$OUTPUT_DIR/web/dist/"

    # 创建干净配置（无敏感信息）
    cat > "$OUTPUT_DIR/config.yaml" << 'EOF'
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

    # 创建 README
    cat > "$OUTPUT_DIR/README.txt" << EOF
AI Writer v${VERSION} - 网文创作助手
========================

启动方式:
  Linux/Mac:   ./start_with_embedding.sh
  Windows:     start_with_embedding.bat

首次启动:
  - 自动下载 embedding 模型（约 22MB）
  - 中国用户自动使用镜像加速

配置:
  - 编辑 config.yaml 配置 LLM API Key
  - 支持 DeepSeek、OpenAI、Ollama

项目地址: https://github.com/cannotgetaname/ai-writer-go
EOF

    # 打包
    cd "$DIST_DIR"
    if [ "$OS" = "windows" ]; then
        rm -f "${OUTPUT_NAME}.zip"
        zip -rq "${OUTPUT_NAME}.zip" "$OUTPUT_NAME"
        echo "  ✅ ${OUTPUT_NAME}.zip ($(du -h "${OUTPUT_NAME}.zip" | cut -f1))"
    else
        rm -f "${OUTPUT_NAME}.tar.gz"
        tar -czf "${OUTPUT_NAME}.tar.gz" "$OUTPUT_NAME"
        echo "  ✅ ${OUTPUT_NAME}.tar.gz ($(du -h "${OUTPUT_NAME}.tar.gz" | cut -f1))"
    fi
}

# ============================================
# 构建各平台
# ============================================

# Linux AMD64 (本地构建)
echo ""
echo "Building Linux AMD64..."
build_package "linux" "amd64" "linux" "amd64" "" ""

# Linux ARM64 (交叉编译)
echo ""
echo "Building Linux ARM64..."
if rustup target list --installed | grep -q "aarch64-unknown-linux-gnu"; then
    build_package "linux" "arm64" "linux" "arm64" "aarch64-unknown-linux-gnu" ""
else
    echo "  ⚠️  Skipping: rust target aarch64-unknown-linux-gnu not installed"
    echo "  Install with: rustup target add aarch64-unknown-linux-gnu"
fi

# Windows AMD64 (交叉编译)
echo ""
echo "Building Windows AMD64..."
if command -v x86_64-w64-mingw32-gcc &> /dev/null; then
    build_package "windows" "amd64" "windows" "amd64" "x86_64-pc-windows-gnu" ".exe"
else
    echo "  ⚠️  Skipping: mingw-w64 not installed"
    echo "  Install with: sudo apt install mingw-w64"
fi

# ============================================
# 输出结果
# ============================================
echo ""
echo "=== Build Complete ==="
echo ""
echo "Output files:"
ls -lh "$DIST_DIR"/*.tar.gz "$DIST_DIR"/*.zip 2>/dev/null || echo "No packages created"

echo ""
echo "To install cross-compile targets:"
echo "  rustup target add aarch64-unknown-linux-gnu"
echo "  rustup target add x86_64-pc-windows-gnu"
echo "  sudo apt install mingw-w64  # For Windows builds"