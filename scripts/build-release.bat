@echo off
REM 打包 AI Writer 为可发布的一键启动包

setlocal enabledelayedexpansion

set VERSION=%1
if "%VERSION%"=="" set VERSION=1.0.0

set DIST_DIR=dist\ai-writer-%VERSION%
set SCRIPT_DIR=%~dp0
set PROJECT_ROOT=%SCRIPT_DIR%..

echo === Building AI Writer Release v%VERSION% ===

REM 1. 创建发布目录
echo Creating release directory...
if exist "%PROJECT_ROOT%\%DIST_DIR%" rmdir /s /q "%PROJECT_ROOT%\%DIST_DIR%"
mkdir "%PROJECT_ROOT%\%DIST_DIR%"
mkdir "%PROJECT_ROOT%\%DIST_DIR%\data"

REM 2. 编译 Go 后端
echo Building Go backend...
cd /d "%PROJECT_ROOT%"
set CGO_ENABLED=0
go build -ldflags="-s -w" -o "%DIST_DIR%\ai-writer.exe" .

REM 3. 编译 Rust embedding 服务
echo Building Rust embedding service...
cd /d "%PROJECT_ROOT%\embedding_service_rust"
if not exist "target\release\embedding_server.exe" (
    echo Embedding server not built, please run: cargo build --release
    echo Make sure Rust is installed: https://rustup.rs
    exit /b 1
)
copy "target\release\embedding_server.exe" "%PROJECT_ROOT%\%DIST_DIR%\" >nul

REM 4. 复制启动脚本
echo Copying startup scripts...
copy "%PROJECT_ROOT%\scripts\start_with_embedding.bat" "%PROJECT_ROOT%\%DIST_DIR%\" >nul
copy "%PROJECT_ROOT%\scripts\start_with_embedding.sh" "%PROJECT_ROOT%\%DIST_DIR%\" >nul

REM 5. 复制前端静态文件
if exist "%PROJECT_ROOT%\web\dist" (
    echo Copying web assets...
    mkdir "%PROJECT_ROOT%\%DIST_DIR%\web\dist"
    xcopy "%PROJECT_ROOT%\web\dist\*" "%PROJECT_ROOT%\%DIST_DIR%\web\dist\" /s /e /q >nul
)

REM 6. 创建默认配置
echo Creating default config...
(
echo server:
echo   port: "8081"
echo   data_dir: "./data"
echo.
echo llm:
echo   provider: "deepseek"
echo   api_key: ""
echo   base_url: ""
echo   models:
echo     writer: "deepseek-chat"
echo     architect: "deepseek-chat"
echo     reviewer: "deepseek-chat"
echo   temperatures:
echo     writer: 1.5
echo     architect: 1.0
echo     reviewer: 0.5
echo.
echo embedding:
echo   provider: "python"
echo   timeout: 30
echo.
echo vector_store:
echo   chunk_size: 500
echo   overlap: 100
) > "%PROJECT_ROOT%\%DIST_DIR%\config.yaml"

REM 7. 创建 README
echo Creating README...
(
echo AI Writer - 网文创作助手
echo ========================
echo.
echo 启动方式:
echo   双击 start_with_embedding.bat
echo.
echo 首次启动:
echo   - 自动下载 embedding 模型（约 22MB）
echo   - 中国用户会自动检测并使用镜像加速下载
echo.
echo 配置文件:
echo   - 编辑 config.yaml 配置 LLM API
echo   - 支持 DeepSeek、OpenAI、Ollama 等提供商
echo.
echo 数据目录:
echo   - .\data\ 存储所有书籍数据
) > "%PROJECT_ROOT%\%DIST_DIR%\README.txt"

REM 8. 打包
echo Creating archive...
cd /d "%PROJECT_ROOT%\dist"
tar -a -c -f "ai-writer-%VERSION%-windows-x64.zip" "ai-writer-%VERSION%"

echo.
echo === Build Complete ===
echo Release directory: %DIST_DIR%
echo Archive: dist\ai-writer-%VERSION%-windows-x64.zip
echo.
echo Contents:
dir /b "%PROJECT_ROOT%\%DIST_DIR%\"

endlocal