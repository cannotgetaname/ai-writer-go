@echo off
REM start.bat - Windows 启动脚本

set TEI_PORT=8081
set MODEL_ID=BAAI/bge-base-zh-v1.5

echo === AI Writer 启动 ===

REM 启动 TEI 服务
echo 启动 Embedding 服务 (端口: %TEI_PORT%)...
start /B text-embeddings-router.exe --model-id %MODEL_ID% --port %TEI_PORT% --dtype float16

REM 等待 TEI 就绪
echo 等待 Embedding 服务就绪...
:wait_loop
curl -s http://127.0.0.1:%TEI_PORT%/health > nul 2>&1
if errorlevel 1 (
    timeout /t 1 > nul
    goto wait_loop
)
echo Embedding 服务就绪

REM 启动主服务
echo 启动 AI Writer...
ai-writer.exe server