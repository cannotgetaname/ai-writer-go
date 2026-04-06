# 世界状态审计系统设计

## 背景

知识图谱功能已支持 6 种图谱类型（关系图、因果链、伏笔、叙事线程、情感弧线、时间线），但图谱数据目前需要手动创建和维护。用户希望通过 AI 自动提取这些数据，并能集中审核确认。

## 设计目标

1. **一键提取** — 在写作页面一键触发 AI 分析当前章节，自动提取所有图谱数据
2. **分类审核** — 按类型分标签页审核，用户逐项确认/修改/拒绝
3. **分散编辑** — 审核确认后的数据在各自设定页面管理
4. **分析增强** — 审稿时自动运行分析 + 手动触发深度分析
5. **报告存储** — 分析结果保存为历史报告，可回顾

---

## 一、核心流程

```
写作页面 → 点击"审计世界状态" → AI分析当前章节 → 弹出分类审核面板 → 用户逐类确认 → 应用变更
```

**审核后编辑入口：**
- 状态变更 → 设定-人物/物品/地点管理
- 因果链 → 设定-因果链管理
- 伏笔 → 设定-伏笔管理
- 叙事线程 → 设定-叙事线程管理
- 情感弧线 → 设定-人物管理（情感标签页）
- 时间线 → 时间线页面

---

## 二、审核面板设计

**6 个分类标签页：**

| 标签 | 提取内容 | 存储位置 | 审核后编辑入口 |
|------|----------|----------|----------------|
| 状态变更 | 人物状态/物品持有者/关系变化 | characters.json, items.json | 设定-人物/物品管理 |
| 因果链 | 因-事-果-决 结构事件 | causal_chains.json | 设定-因果链管理 |
| 伏笔 | 埋设/回收的伏笔 | foreshadows.json | 设定-伏笔管理 |
| 叙事线程 | 线程涉及章节、POV角色 | threads.json | 设定-叙事线程管理 |
| 情感弧线 | 角色情感变化点 | characters.json[].emotional_arc | 设定-人物管理(情感标签页) |
| 时间线 | 事件时间标签、持续时间 | timeline.json | 时间线页面 |

**每个标签页布局：**
```
┌─────────────────────────────────────────────────┐
│ [状态变更] [因果链] [伏笔] [叙事线程] [情感弧线] [时间线] │
├─────────────────────────────────────────────────┤
│ 待确认项列表                                      │
│ ┌─────────────────────────────────────────────┐ │
│ │ ☑ 王刚 状态: 存活 → 重伤                     │ │
│ │   原因: 第3章与魔族战斗受创                  │ │
│ │   [编辑] [接受] [拒绝]                       │ │
│ └─────────────────────────────────────────────┘ │
│ ┌─────────────────────────────────────────────┐ │
│ │ ☐ 新因果链: 偷取秘籍 → 被追杀 → 逃亡        │ │
│ │   涉及角色: 主角、长老                       │ │
│ │   [编辑] [接受] [拒绝]                       │ │
│ └─────────────────────────────────────────────┘ │
│                                                 │
│ [全选] [全不选] [批量接受] [批量拒绝]  [确认应用]  │
└─────────────────────────────────────────────────┘
```

---

## 三、模型扩展

### 3.1 CausalEvent 扩展

```go
type CausalEvent struct {
    ID        string       `json:"id"`
    BookID    string       `json:"book_id"`
    ChapterID int          `json:"chapter_id"`

    // 核心因果结构
    Cause      string   `json:"cause"`
    Event      string   `json:"event"`
    Effect     string   `json:"effect"`
    Decision   string   `json:"decision"`

    // 涉及角色
    Characters []string `json:"characters"`

    // 关联伏笔
    ForeshadowIDs []string `json:"foreshadow_ids,omitempty"`

    // 状态
    Status CausalStatus `json:"status"`

    // 新增字段
    AutoDetected  bool   `json:"auto_detected"`   // 是否AI检测
    SourceContext string `json:"source_context"`  // 原文片段

    CreatedAt time.Time `json:"created_at"`
}
```

### 3.2 Foreshadow 扩展

```go
type Foreshadow struct {
    ID     string          `json:"id"`
    BookID string          `json:"book_id"`
    Type   ForeshadowType  `json:"type"`
    Content string         `json:"content"`
    Importance Importance  `json:"importance"`

    // 埋设信息
    SourceChapter   int    `json:"source_chapter"`
    SourceParagraph string `json:"source_paragraph,omitempty"`

    // 预期回收
    TargetChapter int `json:"target_chapter,omitempty"`

    // 回收信息
    Status          ForeshadowStatus `json:"status"`
    ResolvedChapter int              `json:"resolved_chapter,omitempty"`
    ResolvedContent string           `json:"resolved_content,omitempty"`

    // 关联因果链
    CausalEventID string `json:"causal_event_id,omitempty"`

    // 新增字段
    AutoDetected  bool   `json:"auto_detected"`
    SourceContext string `json:"source_context"`

    Notes string `json:"notes,omitempty"`

    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### 3.3 EmotionPoint 扩展

```go
type EmotionPoint struct {
    ChapterID int    `json:"chapter_id"`
    Emotion   string `json:"emotion"`
    Intensity int    `json:"intensity"`
    Trigger   string `json:"trigger"`

    // 新增字段
    SourceContext string `json:"source_context"` // 触发原文片段
}
```

### 3.4 NarrativeThread 扩展

```go
type NarrativeThread struct {
    ID     string      `json:"id"`
    BookID string      `json:"book_id"`
    Name   string      `json:"name"`
    Type   ThreadType  `json:"type"`

    // 视角角色
    POVCharacters []string `json:"pov_characters"`

    // 目标弧线
    Goal         string `json:"goal"`
    StartChapter int    `json:"start_chapter"`
    EndChapter   int    `json:"end_chapter,omitempty"`

    // 篇幅权重
    Weight int `json:"weight"`

    // 状态
    Status            ThreadStatus `json:"status"`
    LastActiveChapter int          `json:"last_active_chapter"`

    // 关联章节
    Chapters []int `json:"chapters"`

    // 新增字段
    AutoDetected bool `json:"auto_detected"`

    CreatedAt time.Time `json:"created_at"`
}
```

### 3.5 TimelineEvent 扩展

```go
type TimelineEvent struct {
    ChapterID   int      `json:"chapter_id"`
    ThreadID    string   `json:"thread_id,omitempty"`
    TimeLabel   string   `json:"time_label"`
    Duration    string   `json:"duration"`
    Events      []string `json:"events"`
    Characters  []string `json:"characters,omitempty"`
    Location    string   `json:"location,omitempty"`

    // 新增字段
    AutoDetected bool `json:"auto_detected"`
}
```

### 3.6 新增：分析报告

```go
type AnalysisReport struct {
    ID            string    `json:"id"`
    BookID        string    `json:"book_id"`
    ChapterID     int       `json:"chapter_id"`
    Type          string    `json:"type"`  // review / manual
    CreatedAt     time.Time `json:"created_at"`

    // 分析结果
    ForeshadowAnalysis ForeshadowAnalysis `json:"foreshadow_analysis"`
    CausalAnalysis     CausalAnalysis     `json:"causal_analysis"`
    ThreadAnalysis     ThreadAnalysis     `json:"thread_analysis"`
    EmotionAnalysis    EmotionAnalysis    `json:"emotion_analysis"`
    TimelineAnalysis   TimelineAnalysis   `json:"timeline_analysis"`
}

type ForeshadowAnalysis struct {
    Warnings []ForeshadowWarning `json:"warnings"`
    Suggestions []string `json:"suggestions"` // 回收建议
    Score int `json:"score"` // 伏笔管理健康度 0-100
}

type CausalAnalysis struct {
    BrokenChains []BrokenChain `json:"broken_chains"` // 断裂的因果链
    OrphanEvents []OrphanEvent `json:"orphan_events"` // 孤立事件
    CircularDeps []CircularDep `json:"circular_deps"` // 循环依赖
    Score int `json:"score"`
}

type ThreadAnalysis struct {
    ForgottenThreads []ForgottenThread `json:"forgotten_threads"` // 遗忘的线程
    PacingIssues []PacingIssue `json:"pacing_issues"` // 节奏问题
    Conflicts []ThreadConflict `json:"conflicts"` // 线程冲突
    Score int `json:"score"`
}

type EmotionAnalysis struct {
    Inconsistencies []EmotionInconsistency `json:"inconsistencies"` // 情感不一致
    PacingIssues []EmotionPacingIssue `json:"pacing_issues"` // 情感节奏问题
    WeavingScore int `json:"weaving_score"` // 多角色情感交织评分
    Score int `json:"score"`
}

type TimelineAnalysis struct {
    TimeJumps []TimeJump `json:"time_jumps"` // 时间跳跃
    Overlaps []TimeOverlap `json:"overlaps"` // 重叠事件
    Inconsistencies []TimelineInconsistency `json:"inconsistencies"` // 时序矛盾
    Score int `json:"score"`
}
```

### 3.7 新增：待审核图谱变更

```go
type PendingGraphSync struct {
    BookID      string        `json:"book_id"`
    ChapterID   int           `json:"chapter_id"`
    ExtractedAt time.Time     `json:"extracted_at"`

    // 各类待审核数据
    StateChanges   []StateChangeItem   `json:"state_changes"`
    CausalEvents   []CausalEventItem   `json:"causal_events"`
    Foreshadows    []ForeshadowItem    `json:"foreshadows"`
    ThreadUpdates  []ThreadUpdateItem  `json:"thread_updates"`
    EmotionPoints  []EmotionPointItem  `json:"emotion_points"`
    TimelineEvents []TimelineEventItem `json:"timeline_events"`
}

type StateChangeItem struct {
    ID        string `json:"id"`
    Type      string `json:"type"`      // character_status / item_owner / relation
    Entity    string `json:"entity"`
    Field     string `json:"field"`
    OldValue  string `json:"old_value"`
    NewValue  string `json:"new_value"`
    Reason    string `json:"reason"`
    Status    string `json:"status"`    // pending / accepted / rejected
}

type CausalEventItem struct {
    CausalEvent
    Status string `json:"status"` // pending / accepted / rejected
}

type ForeshadowItem struct {
    Foreshadow
    Status string `json:"status"`
}

type ThreadUpdateItem struct {
    ThreadName   string `json:"thread_name"`
    UpdateType   string `json:"update_type"` // new / chapter_add / pov_change
    Chapters     []int  `json:"chapters"`
    POVCharacters []string `json:"pov_characters"`
    Status       string `json:"status"`
}

type EmotionPointItem struct {
    CharacterName string `json:"character_name"`
    EmotionPoint
    Status string `json:"status"`
}

type TimelineEventItem struct {
    TimelineEvent
    Status string `json:"status"`
}
```

---

## 四、分析能力

| 分析类型 | 检测内容 | 触发时机 | 报告存储 |
|----------|----------|----------|----------|
| 伏笔分析 | 超时未回收、错过预期、回收质量、建议时机 | 审稿+手动 | analysis_reports.json |
| 因果链检查 | 断裂链、孤立事件、循环依赖 | 审稿+手动 | analysis_reports.json |
| 叙事线程追踪 | 支线遗忘、主线节奏、线程冲突 | 审稿+手动 | analysis_reports.json |
| 情感弧线分析 | 情感连贯性、高潮低谷节奏、多角色交织 | 审稿+手动 | analysis_reports.json |
| 时间线一致性 | 时间跳跃、重叠事件、时序矛盾 | 审稿+手动 | analysis_reports.json |

### 4.1 伏笔预警规则

- **超时预警**：active 状态伏笔超过 N 章未回收（默认 5 章）
- **错过预期**：当前章节已超过 target_chapter
- **回收质量**：回收方式是否与埋设时暗示一致
- **回收建议**：根据伏笔类型和重要性，建议最佳回收章节

### 4.2 因果链检查规则

- **断裂链**：有因无果（event 有 cause 但无后续 effect）
- **孤立事件**：未被任何因果链引用的事件
- **循环依赖**：A→B→C→A 形成闭环

### 4.3 叙事线程检查规则

- **支线遗忘**：thread 超过 N 章未涉及（last_active_chapter 落后太多）
- **主线节奏**：main thread 的推进频率是否合理
- **线程冲突**：同一章节 POV 角色重叠

### 4.4 情感弧线检查规则

- **情感不一致**：同一章节情感强度跳跃过大
- **节奏问题**：长期无情感变化 或 变化过于频繁
- **多角色交织**：分析多角色情感线是否形成有效的交织模式

### 4.5 时间线检查规则

- **时间跳跃**：章节间时间跨度异常（需手动标注或推断）
- **重叠事件**：同一时间点不同地点的矛盾
- **时序矛盾**：事件顺序与章节顺序不一致

---

## 五、前端改动

### 5.1 写作页面

**新增按钮：**
- 位置：工具栏，与"审稿"、"重写"等按钮并列
- 名称：审计世界状态
- 点击后：弹出分类审核对话框

**审核对话框：**
- 6 个分类标签页
- 每个标签页显示待确认项列表
- 支持单项和批量操作
- 确认应用后更新对应数据文件

### 5.2 审稿功能

**审稿时自动运行分析：**
- 审稿完成后自动触发图谱数据分析
- 审稿报告中显示分析摘要
- 提供"查看详细分析"链接，跳转到分析报告

### 5.3 分析页面

**保留并增强：**
- 显示历史分析报告列表
- 支持手动触发各类分析
- 显示各项分析的评分和问题详情
- 提供快速跳转到对应设定页面的链接

### 5.4 设定页面

**各设定页面编辑入口：**
- 人物管理：编辑人物状态、关系、情感弧线
- 物品管理：编辑物品持有者、势力归属
- 地点管理：编辑地点信息
- 因果链管理：编辑因果事件和链接
- 伏笔管理：编辑伏笔信息
- 叙事线程管理：编辑线程信息

---

## 六、后端 API 扩展

### 6.1 新增接口

```
POST /api/books/:id/sync/extract-all     # 一键提取所有图谱数据
GET  /api/books/:id/sync/pending-graphs  # 获取待审核图谱变更
POST /api/books/:id/sync/apply-graphs    # 应用审核后的图谱变更
POST /api/books/:id/analysis/run         # 手动运行分析
GET  /api/books/:id/analysis/reports     # 获取历史分析报告
GET  /api/books/:id/analysis/reports/:id # 获取单个分析报告
```

### 6.2 扩展现有接口

```
POST /api/ai/review  # 审稿时顺带返回分析摘要
```

---

## 七、存储结构

```
data/projects/{book_name}/
├── reviews.json            # 审稿结果（按章节）
├── analysis_reports.json   # 分析报告（新增）
├── pending_graph_sync.json # 待审核的图谱变更（新增）
├── pending_changes.json    # 待审核的状态变更（现有）
├── causal_chains.json      # 因果链（扩展字段）
├── foreshadows.json        # 伏笔（扩展字段）
├── threads.json            # 叙事线程（扩展字段）
├── timeline.json           # 时间线事件（扩展字段）
├── characters.json         # 人物（含情感弧线，扩展字段）
├── items.json              # 物品
├── locations.json          # 地点
└── worldview.json          # 世界观
```

---

## 八、实施优先级

### Phase 1：核心提取功能
1. 扩展模型结构（新增字段）
2. 实现 `extract-all` 接口
3. 实现前端审核对话框
4. 实现应用变更逻辑

### Phase 2：分析能力
1. 实现各类分析逻辑
2. 审稿时集成分析
3. 分析报告存储和展示

### Phase 3：管理页面增强
1. 因果链管理页面
2. 伏笔管理页面
3. 叙事线程管理页面
4. 分析中心页面