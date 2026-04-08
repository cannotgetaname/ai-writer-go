@echo off
REM 同时启动 Python embedding 服务和 Go 后端

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set SCRIPT_DIR=%SCRIPT_DIR:~0,-1%

echo Starting AI Writer with Python embedding service...

REM 1. 清理旧的端口文件
if exist "%SCRIPT_DIR%\embedding_port.txt" del "%SCRIPT_DIR%\embedding_port.txt" 2>nul

REM 2. 检查 embedding_server 是否存在
set EMBEDDING_SERVER=%SCRIPT_DIR%\embedding_server.exe
if not exist "%EMBEDDING_SERVER%" set EMBEDDING_SERVER=%SCRIPT_DIR%\dist\embedding_server.exe

if not exist "%EMBEDDING_SERVER%" (
    echo Error: embedding_server.exe not found in %SCRIPT_DIR%
    echo Please run: cd embedding_service && python build.py
    exit /b 1
)

REM 3. 启动 Python embedding 服务（后台）
echo Starting embedding service...
start /B "" "%EMBEDDING_SERVER%" >nul 2>&1

REM 4. 等待端口文件生成（最多 60 秒）
echo Waiting for embedding service to start...
set COUNT=0
set MAX_WAIT=60

:wait_loop
if exist "%SCRIPT_DIR%\embedding_port.txt" goto got_port
set /a COUNT+=1
if !COUNT! geq %MAX_WAIT% goto timeout_error
timeout /t 1 /nobreak >nul
goto wait_loop

:timeout_error
echo Error: Embedding service failed to start (timeout after %MAX_WAIT%s)
taskkill /F /IM embedding_server.exe 2>nul
exit /b 1

:got_port
set /p PORT=<"%SCRIPT_DIR%\embedding_port.txt"
echo Embedding service started on port !PORT!

REM 5. 启动 Go 后端
set AI_WRITER=%SCRIPT_DIR%\ai-writer.exe
if not exist "%AI_WRITER%" set AI_WRITER=%SCRIPT_DIR%\..\ai-writer.exe

if not exist "%AI_WRITER%" (
    echo Error: ai-writer.exe not found
    taskkill /F /IM embedding_server.exe 2>nul
    exit /b 1
)

echo Starting AI Writer backend...
"%AI_WRITER%" server

REM 6. Go 后端退出时，清理 Python 进程
echo Shutting down...
taskkill /F /IM embedding_server.exe 2>nul
if exist "%SCRIPT_DIR%\embedding_port.txt" del "%SCRIPT_DIR%\embedding_port.txt" 2>nul

endlocal