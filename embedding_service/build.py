#!/usr/bin/env python3
"""
PyInstaller 打包脚本
打包 embedding_server 为独立可执行文件
"""

import PyInstaller.__main__
import sys
from pathlib import Path

# 确保在正确目录
script_dir = Path(__file__).parent.absolute()
main_script = script_dir / "embedding_server.py"

print(f"Building embedding_server from {main_script}")

PyInstaller.__main__.run([
    str(main_script),
    '--name=embedding_server',
    '--onefile',
    '--clean',
    '--noconfirm',
    # 收集 sentence_transformers 数据文件
    '--collect-data=sentence_transformers',
    # 隐藏导入
    '--hidden-import=sentence_transformers',
    '--hidden-import=sentence_transformers.models',
    '--hidden-import=fastapi',
    '--hidden-import=uvicorn',
    '--hidden-import=uvicorn.logging',
    '--hidden-import=uvicorn.loops',
    '--hidden-import=uvicorn.loops.auto',
    '--hidden-import=uvicorn.protocols',
    '--hidden-import=uvicorn.protocols.http',
    '--hidden-import=uvicorn.protocols.http.auto',
    '--hidden-import=uvicorn.protocols.websockets',
    '--hidden-import=uvicorn.protocols.websockets.auto',
    '--hidden-import=uvicorn.protocols.websockets.websockets_impl',
    '--hidden-import=uvicorn.lifespan',
    '--hidden-import=uvicorn.lifespan.on',
    '--hidden-import=pydantic',
    '--hidden-import=huggingface_hub',
    '--hidden-import=transformers',
    # 排除不必要的模块以减小体积
    '--exclude-module=tkinter',
    '--exclude-module=PIL',
    '--exclude-module=matplotlib',
    '--exclude-module=numpy.f2py',
    # 工作目录
    '--distpath', str(script_dir / 'dist'),
    '--workpath', str(script_dir / 'build'),
    '--specpath', str(script_dir),
])

print("\nBuild complete!")
print(f"Output: {script_dir / 'dist' / 'embedding_server'}")

# Windows 下提示
if sys.platform == 'win32':
    print("\nNote: On Windows, the output file will be embedding_server.exe")