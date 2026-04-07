# AI Writer - AI辅助小说创作工具

[![Release](https://img.shields.io/github/v/release/cannotgetaname/ai-writer-go)](https://github.com/cannotgetaname/ai-writer-go/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
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
wget https://github.com/cannotgetaname/ai-writer-go/releases/download/v1.2.0/ai-writer-linux-amd64.tar.gz
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
- **设定管理**: 世界观、人物、物品、地点、势力、伏笔
- **势力管理**: 创建势力、设置势力关系（联盟/敌对/附属/中立）、成员与领地管理
- **状态同步**: 从章节提取状态变更并应用
- **导出功能**: 支持 txt/markdown/json 格式
- **时间线**: 可视化故事时间线
- **知识图谱**: 可视化展示人物、物品、地点、势力之间的关系网络
- **世界状态审计**: AI 自动提取图谱数据，集中审核确认
- **系统配置**: 模型设置、提示词配置、费用统计

#### 知识图谱

支持六种图谱类型：

| 图谱类型 | 说明 |
|---------|------|
| 基础关系图 | 人物、物品、地点、势力之间的关系网络 |
| 剧情因果图 | 因-事-果-决 结构的因果链事件 |
| 伏笔追踪图 | 伏笔埋设与回收状态追踪 |
| 叙事线程图 | 主线/支线叙事线程分布 |
| 情感弧线图 | 角色情感变化折线图 |
| 时间线图 | 故事时间线事件 |

**节点类型与颜色**：

| 节点类型 | 颜色 | 说明 |
|---------|------|------|
| 人物 | 蓝色 | 角色及其关系 |
| 物品 | 黄色 | 物品及其归属 |
| 地点 | 绿色 | 地点及其层级 |
| 势力 | 红色 | 势力及其关系 |

**关系类型**：
- 归属关系（实线箭头）：人物→势力、物品→人物/势力/地点、地点→势力
- 从属关系（实线箭头）：子地点→父地点、子势力→父势力
- 势力关系（实线）：联盟(绿)、敌对(红)、中立(灰)
- 联通关系（虚线）：地点间的可达路径
- 人物关系（点线）：朋友、敌人等平级关系

**交互功能**：
- 鼠标悬停节点：级联高亮显示整条归属链
- 筛选节点类型：只显示特定类型的节点
- 点击节点：查看详情和关联节点

#### 世界状态审计

在写作页面点击"审计世界状态"按钮，AI 会自动分析当前章节并提取：

- **状态变更**：人物状态、物品持有者、关系变化
- **因果链**：因-事-果-决 结构事件
- **伏笔**：埋设/回收的伏笔
- **叙事线程**：线程涉及章节、POV角色
- **情感弧线**：角色情感变化点
- **时间线**：事件时间标签、持续时间

提取的数据显示在分类审核面板中，用户可以逐项确认或批量接受，应用后数据会保存到对应的设定文件。

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