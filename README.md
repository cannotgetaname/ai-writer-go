# AI Writer - AI辅助小说创作工具

[![Release](https://img.shields.io/github/v/release/cannotgetaname/ai-writer-go)](https://github.com/cannotgetaname/ai-writer-go/releases)
[![License](https://img.shields.io/github/license/cannotgetaname/ai-writer-go)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org)
[![Vue](https://img.shields.io/badge/Vue-3.4+-4FC08D?logo=vue.js)](https://vuejs.org)

一个基于 Go + Vue 3 的 AI 辅助小说创作工具，支持 CLI 和 Web UI 两种使用方式。

## 下载安装

### 预编译版本

从 [Releases](https://github.com/cannotgetaname/ai-writer-go/releases) 页面下载对应平台的版本：

| 平台 | 架构 | 文件 |
|------|------|------|
| Linux | x86_64 | ai-writer-linux-amd64.tar.gz |
| Linux | ARM64 | ai-writer-linux-arm64.tar.gz |
| Windows | x86_64 | ai-writer-windows-amd64.zip |

**Linux 安装：**
```bash
# 下载并解压
wget https://github.com/cannotgetaname/ai-writer-go/releases/download/v1.0.0/ai-writer-linux-amd64.tar.gz
tar -xzvf ai-writer-linux-amd64.tar.gz

# 配置
cp config.example.yaml config.yaml
vim config.yaml  # 填入你的 API Key

# 启动
./start-linux.sh
```

**Windows 安装：**
1. 下载并解压 `ai-writer-windows-amd64.zip`
2. 复制 `config.example.yaml` 为 `config.yaml`
3. 编辑 `config.yaml` 填入你的 API Key
4. 双击 `start-windows.bat` 启动

### 从源码编译

```bash
# 克隆仓库
git clone https://github.com/cannotgetaname/ai-writer-go.git
cd ai-writer-go

# 编译后端
go build -o ai-writer .

# 构建前端
cd web && npm install && npm run build
```

## 功能特性

### CLI 命令

```bash
# 书籍管理
ai-writer book list
ai-writer book create <name>
ai-writer book delete <name>
ai-writer book info
ai-writer book use <name>

# 章节管理
ai-writer chapter list
ai-writer chapter show <id>
ai-writer chapter add --title "第一章"
ai-writer chapter edit <id>
ai-writer chapter delete <id>

# AI 写作
ai-writer write <chapter_id> --stream
ai-writer continue --words 500

# 批量生成
ai-writer batch generate --from 1 --to 50
ai-writer batch continue
ai-writer batch status
ai-writer batch reset

# 审稿和审计
ai-writer review <chapter_id> --fix
ai-writer audit

# 状态同步
ai-writer sync extract <chapter_id>
ai-writer sync pending
ai-writer sync apply
ai-writer sync reject

# 导出
ai-writer export txt
ai-writer export markdown
ai-writer export json

# 智能工具箱
ai-writer tool naming --type person --genre 玄幻
ai-writer tool character --type 主角 --gender 男
ai-writer tool conflict --type 人物冲突
ai-writer tool scene --type 战斗
ai-writer tool goldfinger --type 系统
ai-writer tool title --genre 玄幻
```

### Web UI

访问 http://localhost:8081 即可使用 Web UI。

- **书籍管理**: 创建、编辑、删除书籍项目
- **章节编辑**: 编写章节内容，AI 生成、续写
- **批量生成**: 流水线式批量生成章节，支持断点续传
- **状态同步**: 从章节提取状态变更并应用
- **导出功能**: 支持 txt/markdown/json 格式
- **时间线**: 可视化故事时间线
- **知识图谱**: 展示人物、地点、物品关系
- **系统配置**: 模型设置、提示词配置、费用统计

## 技术栈

- **后端**: Go 1.21+, Gin, Cobra CLI
- **前端**: Vue 3, Element Plus, ECharts
- **存储**: JSON 文件存储
- **LLM**: 支持 DeepSeek / OpenAI / Ollama

## 项目结构

```
ai-writer-go/
├── cmd/                    # CLI 命令
├── internal/
│   ├── api/               # REST API
│   ├── config/            # 配置管理
│   ├── engine/            # 高级引擎（因果链、叙事线程等）
│   ├── llm/               # LLM 客户端
│   ├── model/             # 数据模型
│   ├── service/           # 业务逻辑
│   └── store/             # 存储层
├── web/                    # Vue 前端
│   └── src/
│       ├── api/           # API 客户端
│       ├── router/        # 路由配置
│       └── views/         # Vue 组件
├── configs/               # 示例配置
└── main.go
```

## 数据存储

所有数据存储在 `data/projects/{book_name}/` 目录下：

```
data/projects/{book_name}/
├── metadata.json        # 书籍元数据
├── structure.json       # 章节结构
├── characters.json      # 人物设定
├── items.json           # 物品设定
├── locations.json       # 地点设定
├── worldview.json       # 世界观
├── foreshadows.json     # 伏笔追踪
├── causal_chains.json   # 因果链
├── threads.json         # 叙事线程
└── chapters/
    └── 1.json           # 章节内容
```

## 配置说明

编辑 `config.yaml` 文件：

```yaml
llm:
  provider: deepseek           # deepseek / openai / ollama
  api_key: your-api-key        # API 密钥
  base_url: https://api.deepseek.com
  models:
    writer: deepseek-chat      # 写作模型
    architect: deepseek-reasoner  # 架构师模型
    reviewer: deepseek-chat    # 审稿模型
  temperatures:
    writer: 1.5                # 写作温度 (0-2)
    architect: 1.0
    reviewer: 0.5

server:
  port: "8081"
  data_dir: "data"
```

## License

本项目采用 [MIT License](LICENSE) 开源协议。

```
MIT License

Copyright (c) 2025 cannotgetaname

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.
```

简而言之，你可以：
- ✅ 商业使用
- ✅ 修改
- ✅ 分发
- ✅ 私人使用

唯一要求：在软件副本中包含版权声明和许可声明。