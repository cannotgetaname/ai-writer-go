@echo off
REM start.bat - Windows 启动脚本

set TEI_PORT=8081
set MODEL_ID=BAAI/bge-base-zh-v1.5
set MODEL_CACHE=%USERPROFILE%\.cache\huggingface\hub\models--BAAI--bge-base-zh-v1.5

echo === AI Writer 启动 ===

REM 检查模型是否已下载
if not exist "%MODEL_CACHE%" (
    echo 首次运行，正在下载模型，请耐心等待...
    echo 模型: %MODEL_ID%
)

REM 启动 TEI 服务
echo 启动 Embedding 服务 (端口: %TEI_PORT%)...
start /B text-embeddings-router.exe --model-id %MODEL_ID% --port %TEI_PORT% --dtype float16

REM 等待 TEI 就绪 (最多 60 秒)
echo 等待 Embedding 服务就绪...
set /a COUNT=0
:wait_loop
curl -s http://127.0.0.1:%TEI_PORT%/health > nul 2>&1
if errorlevel 1 (
    set /a COUNT+=1
    if %COUNT% GEQ 60 (
        echo 错误: Embedding 服务启动超时 (60秒)
        exit /b 1
    )
    timeout /t 1 > nul
    goto wait_loop
)
echo Embedding 服务就绪

REM 启动主服务
echo 启动 AI Writer...
ai-writer.exe server

REM 清理：终止 TEI 进程
echo 正在关闭 Embedding 服务...
taskkill /F /IM text-embeddings-router.exe > nul 2>&1