#!/bin/bash
set -euo pipefail

# build.sh - 编译脚本
# 用法: ./scripts/build.sh [version]

VERSION=${1:-"dev"}
OUTPUT_DIR="release"

# Check for required tools
check_required_tools() {
    local missing=()

    if ! command -v go &> /dev/null; then
        missing+=("go")
    fi

    if ! command -v npm &> /dev/null; then
        missing+=("npm")
    fi

    if ! command -v wget &> /dev/null && ! command -v curl &> /dev/null; then
        missing+=("wget or curl")
    fi

    if [ ${#missing[@]} -ne 0 ]; then
        echo "Error: Missing required tools: ${missing[*]}"
        echo "Please install them and try again."
        exit 1
    fi
}

# Download file using wget or curl as fallback
download_file() {
    local url="$1"
    local output="$2"

    if command -v wget &> /dev/null; then
        if ! wget -q -O "$output" "$url"; then
            echo "Error: Failed to download $url with wget"
            return 1
        fi
    elif command -v curl &> /dev/null; then
        if ! curl -sL -o "$output" "$url"; then
            echo "Error: Failed to download $url with curl"
            return 1
        fi
    else
        echo "Error: Neither wget nor curl is available"
        return 1
    fi
}

echo "=== AI Writer Build Script ==="
echo "Version: $VERSION"
echo "Output: $OUTPUT_DIR"

# Check required tools
check_required_tools

# 清理输出目录
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

# 构建前端
echo "Building frontend..."
cd web
if ! npm install; then
    echo "Error: npm install failed"
    exit 1
fi
if ! npm run build; then
    echo "Error: npm run build failed"
    exit 1
fi
cd ..

# 编译后端 (Linux amd64) - 当前平台
echo "Building Linux amd64..."
if ! CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$VERSION" -o "$OUTPUT_DIR/ai-writer" .; then
    echo "Error: Failed to build Linux amd64"
    exit 1
fi

# 编译后端 (Linux arm64) - 可选，需要交叉编译器
echo "Building Linux arm64..."
if command -v aarch64-linux-gnu-gcc &> /dev/null; then
    CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc go build -ldflags="-s -w -X main.Version=$VERSION" -o "$OUTPUT_DIR/ai-writer-linux-arm64" . || \
        echo "Warning: Failed to build Linux arm64, skipping..."
else
    echo "Skipping Linux arm64 (aarch64-linux-gnu-gcc not found)"
fi

# 编译后端 (Windows amd64) - 可选，需要交叉编译器
echo "Building Windows amd64..."
if command -v x86_64-w64-mingw32-gcc &> /dev/null; then
    CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -ldflags="-s -w -X main.Version=$VERSION" -o "$OUTPUT_DIR/ai-writer-windows-amd64.exe" . || \
        echo "Warning: Failed to build Windows amd64, skipping..."
else
    echo "Skipping Windows amd64 (x86_64-w64-mingw32-gcc not found)"
fi

# 复制文件到 release 目录
echo "Copying files to release..."
cp -r web/dist "$OUTPUT_DIR/web"
cp scripts/start.sh "$OUTPUT_DIR/"
cp scripts/start.bat "$OUTPUT_DIR/"
cp configs/config.example.yaml "$OUTPUT_DIR/config.yaml" 2>/dev/null || true
chmod +x "$OUTPUT_DIR/start.sh"

echo ""
echo "Build complete!"
echo ""
echo "文件列表:"
ls -la "$OUTPUT_DIR/"
echo ""
echo "使用方法:"
echo "  cd release"
echo "  ./start.sh"