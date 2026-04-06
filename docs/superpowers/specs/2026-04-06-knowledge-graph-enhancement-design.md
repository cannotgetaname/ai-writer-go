# 知识图谱增强设计

## 背景

当前图谱只展示从属关系（人物→势力、物品→持有者、地点→势力），不够全面。用户希望面向不同使用场景（创作辅助、复盘分析、世界观展示）提供多种图谱视图模式。

## 设计目标

实现类似手机桌面应用的图谱网格布局：
- 网格布局 + 可拖拽调整位置 + 自动填补空位 + 响应式
- 每个图谱是一个独立格子，可选择打开或关闭
- 只有被打开的类型才会显示对应格子

## 支持的图谱类型

| 类型 | 用途 | 展示形式 |
|------|------|----------|
| relationship | 基础关系图 | 力导向图（已有） |
| causal | 剧情因果图 | 力导向图，因→事→果→决流程 |
| foreshadow | 伏笔追踪图 | 力导向图，埋设→回收连线 |
| thread | 叙事线程图 | 力导向图，主线/支线分布 |
| emotion | 情感弧线图 | 折线图/热力图 |
| timeline | 时间线图 | 时间轴 + 事件节点 |

---

## 后端 API 扩展

### API 设计

```
GET /api/books/:id/graph?type=relationship   # 基础关系图（默认）
GET /api/books/:id/graph?type=causal         # 剧情因果图
GET /api/books/:id/graph?type=foreshadow     # 伏笔追踪图
GET /api/books/:id/graph?type=thread         # 叙事线程图
GET /api/books/:id/graph?type=emotion        # 情感弧线图
GET /api/books/:id/graph?type=timeline       # 时间线图
GET /api/books/:id/graph?type=all            # 返回所有图谱数据（可选，减少请求次数）
```

注：保持现有路由 `/api/books/:id/graph/echarts`，通过 query 参数 `type` 区分图谱类型。若无 type 参数则默认返回 relationship 类型。

### 统一响应格式

```json
{
  "type": "causal",
  "nodes": [...],
  "links": [...],
  "categories": [...],
  "metadata": {
    // 图谱特有的附加信息
  }
}
```

### 改动文件

- `internal/api/handler/handler.go`：扩展 `GetEChartsData` handler，增加 type 参数处理
- `internal/service/graph.go`：新增以下方法：
  - `BuildCausalGraph(bookName string) (*GraphData, error)`
  - `BuildForeshadowGraph(bookName string) (*GraphData, error)`
  - `BuildThreadGraph(bookName string) (*GraphData, error)`
  - `BuildEmotionGraph(bookName string) (*EmotionGraphData, error)`
  - `BuildTimelineGraph(bookName string) (*GraphData, error)`
- `internal/api/router.go`：保持现有路由，通过 query 参数区分

---

## 前端组件结构

```
web/src/views/GraphView.vue        # 主页面（重构）
web/src/components/
  GraphGrid.vue                    # 网格容器组件（拖拽布局管理）
  GraphCard.vue                    # 图谱卡片组件（单个格子）
  graphs/
    RelationshipGraph.vue          # 基础关系图（复用现有逻辑）
    CausalGraph.vue                # 剧情因果图
    ForeshadowGraph.vue            # 伏笔追踪图
    ThreadGraph.vue                # 叙事线程图
    EmotionGraph.vue               # 情感弧线图
    TimelineGraph.vue              # 时间线图
```

### GraphGrid 网格容器

- CSS Grid 布局，响应式
  - 宽屏（>1200px）：3 列
  - 中屏（800-1200px）：2 列
  - 窄屏（<800px）：1 列
- 拖拽功能：使用 Vue Draggable Plus 或类似库
- 关闭卡片后自动填补空位
- 位置和开关状态保存到 localStorage

### GraphCard 卡片组件

- 标题栏：图谱名称 + 控制按钮（最小化、最大化、关闭）
- 内容区：嵌入具体的图谱组件
- 拖拽 handle（标题栏区域）
- 最小化时只显示标题和统计摘要（如：12 节点，5 条关系）
- 最大化行为：
  - 卡片宽度撑满整行（跨所有列）
  - 高度增加（如 600px → 800px）
  - 其他卡片保持原位置，宽度不变，高度压缩为最小化状态

### 图谱开关面板

- 右侧或顶部浮动面板
- 勾选开启/关闭
- 显示每种图谱的数据统计

---

## 各图谱数据结构

### 剧情因果图 (causal)

**节点**：
```json
{
  "name": "事件简述",
  "category": "event",
  "chapter_id": 5,
  "characters": ["主角A", "反派B"],
  "symbolSize": 35,
  "itemStyle": { "color": "#5470c6" }
}
```

**边类型**：
- `leads_to`：直接导致（实线箭头）
- `enables`：促成条件（虚线箭头）
- `blocks`：阻碍关系（红色虚线）

### 伏笔追踪图 (foreshadow)

**节点**：
```json
{
  "name": "伏笔内容摘要",
  "category": "foreshadow",
  "status": "active/resolved/expired/abandoned",
  "importance": "high/medium/low",
  "source_chapter": 3,
  "target_chapter": 10,
  "resolved_chapter": 12,
  "symbolSize": 根据重要性,
  "itemStyle": { "color": 根据状态 }
}
```

**状态颜色**：
- active：蓝色
- resolved：绿色
- expired：橙色
- abandoned：灰色

**边类型**：
- 埋设：伏笔 → 埋设章节
- 回收：伏笔 → 回收章节
- 预期：伏笔 → 预期回收章节（虚线）
- 关联因果：伏笔 → 因果事件

### 叙事线程图 (thread)

**节点**：
```json
{
  "name": "线程名称",
  "category": "thread",
  "thread_type": "main/sub/parallel/flashback",
  "status": "active/paused/complete",
  "weight": 3,
  "pov_characters": ["主角A"],
  "chapters": [1,2,3,5,8],
  "symbolSize": 根据权重,
  "itemStyle": { "color": 根据类型 }
}
```

**线程类型颜色**：
- main：红色
- sub：蓝色
- parallel：绿色
- flashback：紫色

**边类型**：
- 涉及：线程 → 章节
- 视角：线程 → POV 角色

### 情感弧线图 (emotion)

特殊：折线图/热力图，非力导向图。

```json
{
  "type": "emotion",
  "chart_type": "line",
  "data": [
    {
      "character": "主角A",
      "points": [
        { "chapter": 1, "emotion": "喜悦", "intensity": 8, "trigger": "获得神器" },
        { "chapter": 3, "emotion": "愤怒", "intensity": 9, "trigger": "朋友被害" }
      ]
    }
  ],
  "metadata": {
    "characters": ["主角A", "配角B"],
    "chapter_range": [1, 10],
    "emotion_types": ["喜悦", "愤怒", "悲伤", "恐惧", "惊讶"]
  }
}
```

### 时间线图 (timeline)

**节点**：
```json
{
  "name": "事件描述",
  "category": "event",
  "chapter_id": 5,
  "time_label": "修仙历10年春",
  "duration": "3天",
  "location": "青云宗",
  "characters": ["主角A"],
  "symbolSize": 30
}
```

**边类型**：
- 时间顺序：章节 → 下一章节（虚线）
- 发生地点：事件 → 地点
- 参与角色：事件 → 角色

---

## 错误处理与边界情况

### 后端

- 数据不存在：返回空图谱结构 `{ nodes: [], links: [], metadata: { message: "暂无数据" } }`
- 无效图谱类型：返回 400 `{ error: "无效的图谱类型: xxx" }`
- 加载失败：返回 500，不阻断其他图谱

### 前端

- 空数据：卡片内显示"暂无数据"提示
- 单个图谱加载失败：卡片显示错误状态，其他图谱正常
- 拖拽冲突：自动交换位置
- 响应式切换：自动调整列数
- localStorage 损坏：使用默认布局

---

## 实施顺序

| 步骤 | 内容 | 文件 |
|------|------|------|
| 1 | 后端扩展 GetEChartsData handler，支持 type 参数 | handler.go |
| 2 | 后端实现 BuildCausalGraph | graph.go |
| 3 | 后端实现 BuildForeshadowGraph | graph.go |
| 4 | 后端实现 BuildThreadGraph | graph.go |
| 5 | 后端实现 BuildEmotionGraph | graph.go |
| 6 | 后端实现 BuildTimelineGraph | graph.go |
| 7 | 前端 API 层扩展，支持 type 参数 | api/index.js |
| 8 | 前端 GraphGrid 网格容器组件 | components/GraphGrid.vue |
| 9 | 前端 GraphCard 卡片组件 | components/GraphCard.vue |
| 10 | 前端重构 GraphView 主页面 | views/GraphView.vue |
| 11 | 前端各图谱子组件 | components/graphs/*.vue |
| 12 | 整合测试 | - |