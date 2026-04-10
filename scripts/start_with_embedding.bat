@echo off
REM Start Rust embedding service and Go backend

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set SCRIPT_DIR=%SCRIPT_DIR:~0,-1%

REM Change to script directory so relative paths work
cd /d "%SCRIPT_DIR%"

echo Starting AI Writer with Rust embedding service...

REM 1. Clean old port file
if exist "%SCRIPT_DIR%\embedding_port.txt" del "%SCRIPT_DIR%\embedding_port.txt" 2>nul

REM 2. Check embedding_server exists
set EMBEDDING_SERVER=%SCRIPT_DIR%\embedding_server.exe

if not exist "%EMBEDDING_SERVER%" (
    echo Error: embedding_server.exe not found in %SCRIPT_DIR%
    echo Please build it: cd embedding_service_rust && cargo build --release
    exit /b 1
)

REM 3. Start Rust embedding service (background)
echo Starting embedding service...
start /B "" "%EMBEDDING_SERVER%" >nul 2>&1

REM 4. Wait for port file (max 60 seconds)
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

REM Validate port is numeric
echo !PORT!| findstr /r "^[0-9][0-9]*$" >nul
if errorlevel 1 (
    echo Error: Invalid port number in file: !PORT!
    taskkill /F /IM embedding_server.exe 2>nul
    exit /b 1
)

echo Embedding service started on port !PORT!

REM 5. Start Go backend
set AI_WRITER=%SCRIPT_DIR%\ai-writer.exe
if not exist "%AI_WRITER%" set AI_WRITER=%SCRIPT_DIR%\..\ai-writer.exe

if not exist "%AI_WRITER%" (
    echo Error: ai-writer.exe not found
    taskkill /F /IM embedding_server.exe 2>nul
    exit /b 1
)

echo Starting AI Writer backend...
"%AI_WRITER%" server

REM 6. Cleanup on exit
echo Shutting down...
taskkill /F /IM embedding_server.exe 2>nul
if exist "%SCRIPT_DIR%\embedding_port.txt" del "%SCRIPT_DIR%\embedding_port.txt" 2>nul

endlocal