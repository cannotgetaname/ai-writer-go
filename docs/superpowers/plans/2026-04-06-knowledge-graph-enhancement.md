# 知识图谱增强实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现多图谱网格布局系统，支持 6 种图谱类型的切换和展示。

**Architecture:** 后端扩展 GetEChartsData handler 支持 type 参数，新增 5 个图谱构建方法；前端实现网格容器组件和 6 个图谱子组件。

**Tech Stack:** Go Gin, Vue 3, Element Plus, ECharts, vue-draggable-plus

---

## 文件结构

**后端修改:**
- `internal/service/graph.go` - 新增 EmotionGraphData 结构体和 5 个构建方法
- `internal/api/handler/handler.go:2064-2075` - 修改 GetEChartsData 函数

**前端创建:**
- `web/src/components/GraphGrid.vue` - 网格容器组件
- `web/src/components/GraphCard.vue` - 图谱卡片组件
- `web/src/components/graphs/RelationshipGraph.vue` - 基础关系图
- `web/src/components/graphs/CausalGraph.vue` - 剧情因果图
- `web/src/components/graphs/ForeshadowGraph.vue` - 伏笔追踪图
- `web/src/components/graphs/ThreadGraph.vue` - 叙事线程图
- `web/src/components/graphs/EmotionGraph.vue` - 情感弧线图
- `web/src/components/graphs/TimelineGraph.vue` - 时间线图

**前端修改:**
- `web/src/api/index.js:168-171` - 扩展 graphApi
- `web/src/views/GraphView.vue` - 重构为网格布局

---

## Phase 1: 后端 API 扩展

### Task 1: 扩展 GetEChartsData Handler

**Files:**
- Modify: `internal/api/handler/handler.go:2064-2075`

- [ ] **Step 1: 修改 GetEChartsData 函数，支持 type 参数**

将现有函数替换为：

```go
// GetEChartsData 获取 ECharts 图谱数据
func GetEChartsData(c *gin.Context) {
	bookID := c.Param("id")
	graphType := c.DefaultQuery("type", "relationship")

	// 验证图谱类型
	validTypes := map[string]bool{
		"relationship": true,
		"causal":       true,
		"foreshadow":   true,
		"thread":       true,
		"emotion":      true,
		"timeline":     true,
	}
	if !validTypes[graphType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的图谱类型: " + graphType})
		return
	}

	graphService := service.NewGraphService(jsonStore)

	var data interface{}
	var err error

	switch graphType {
	case "relationship":
		data, err = graphService.BuildGraph(bookID)
	case "causal":
		data, err = graphService.BuildCausalGraph(bookID)
	case "foreshadow":
		data, err = graphService.BuildForeshadowGraph(bookID)
	case "thread":
		data, err = graphService.BuildThreadGraph(bookID)
	case "emotion":
		data, err = graphService.BuildEmotionGraph(bookID)
	case "timeline":
		data, err = graphService.BuildTimelineGraph(bookID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}
```

- [ ] **Step 2: 构建验证**

Run: `cd /home/zcz/program/ai-writer-go && go build -o ai-writer .`
Expected: 编译成功（此时会报错缺少方法，下一步添加）

---

## Phase 2: 后端图谱构建方法

### Task 2: 添加 EmotionGraphData 结构体

**Files:**
- Modify: `internal/service/graph.go:18-56`

- [ ] **Step 1: 在 GraphData 结构体后添加 EmotionGraphData 结构体**

在 `internal/service/graph.go` 第 56 行（Category 结构体后）添加：

```go
// EmotionGraphData 情感图谱数据（折线图/热力图）
type EmotionGraphData struct {
	Type       string            `json:"type"`
	ChartType  string            `json:"chart_type"` // line / heatmap
	Data       []EmotionArcData  `json:"data"`
	Metadata   EmotionMetadata   `json:"metadata"`
}

// EmotionArcData 单个角色的情感弧线
type EmotionArcData struct {
	Character string          `json:"character"`
	Points    []EmotionPoint  `json:"points"`
}

// EmotionPoint 情感点
type EmotionPoint struct {
	Chapter   int    `json:"chapter"`
	Emotion   string `json:"emotion"`
	Intensity int    `json:"intensity"`
	Trigger   string `json:"trigger"`
}

// EmotionMetadata 情感图谱元数据
type EmotionMetadata struct {
	Characters    []string `json:"characters"`
	ChapterRange  []int    `json:"chapter_range"`
	EmotionTypes  []string `json:"emotion_types"`
}
```

- [ ] **Step 2: 构建验证**

Run: `cd /home/zcz/program/ai-writer-go && go build -o ai-writer .`
Expected: 编译成功

---

### Task 3: 实现 BuildCausalGraph 方法

**Files:**
- Modify: `internal/service/graph.go` - 在文件末尾添加方法

- [ ] **Step 1: 添加 BuildCausalGraph 方法**

在 `internal/service/graph.go` 文件末尾（FindPath 函数后）添加：

```go
// BuildCausalGraph 构建因果链图谱
func (s *GraphService) BuildCausalGraph(bookName string) (*GraphData, error) {
	data := &GraphData{
		Nodes: []GraphNode{},
		Links: []GraphLink{},
		Categories: []Category{
			{Name: "event"},
			{Name: "chapter"},
		},
	}

	colorMap := map[string]string{
		"event":   "#5470c6", // 蓝色
		"chapter": "#91cc75", // 绿色
	}

	nodeNames := make(map[string]bool)

	addNode := func(name, category string, symbolSize int, value string, chapterID int) {
		if nodeNames[name] {
			return
		}
		nodeNames[name] = true
		node := GraphNode{
			Name:       name,
			Category:   category,
			SymbolSize: symbolSize,
			Value:      truncateStr(value, 30),
		}
		node.ItemStyle.Color = colorMap[category]
		node.Label.Show = true
		node.Label.Position = "right"
		data.Nodes = append(data.Nodes, node)
	}

	addLink := func(source, target, linkType string) {
		if source == "" || target == "" || source == target {
			return
		}

		link := GraphLink{
			Source: source,
			Target: target,
			Value:  linkType,
			Symbol: []string{"none", "arrow"},
		}

		// 根据类型设置线条样式
		switch linkType {
		case "leads_to":
			link.LineStyle.Type = "solid"
			link.LineStyle.Color = "#5470c6"
		case "enables":
			link.LineStyle.Type = "dashed"
			link.LineStyle.Color = "#91cc75"
		case "blocks":
			link.LineStyle.Type = "dashed"
			link.LineStyle.Color = "#ee6666"
		default:
			link.LineStyle.Type = "solid"
			link.LineStyle.Color = "#5470c6"
		}
		link.LineStyle.Curveness = 0.2

		data.Links = append(data.Links, link)
	}

	// 加载因果链数据
	causalChain, err := s.store.LoadCausalChain(bookName)
	if err != nil {
		// 返回空图谱
		return data, nil
	}

	// 构建事件节点和因果链接
	for _, event := range causalChain.Events {
		// 事件节点
		eventLabel := fmt.Sprintf("第%d章: %s", event.ChapterID, truncateStr(event.Event, 20))
		addNode(eventLabel, "event", 35, event.Event, event.ChapterID)

		// 添加章节节点（用于定位）
		chapterLabel := fmt.Sprintf("第%d章", event.ChapterID)
		addNode(chapterLabel, "chapter", 25, "", event.ChapterID)

		// 事件 → 章节
		addLink(eventLabel, chapterLabel, "归属")
	}

	// 构建因果链接（事件之间）
	for _, link := range causalChain.Links {
		// 查找源事件和目标事件的名称
		var sourceName, targetName string
		for _, event := range causalChain.Events {
			if event.ID == link.FromEventID {
				sourceName = fmt.Sprintf("第%d章: %s", event.ChapterID, truncateStr(event.Event, 20))
			}
			if event.ID == link.ToEventID {
				targetName = fmt.Sprintf("第%d章: %s", event.ChapterID, truncateStr(event.Event, 20))
			}
		}
		if sourceName != "" && targetName != "" {
			addLink(sourceName, targetName, link.LinkType)
		}
	}

	return data, nil
}
```

- [ ] **Step 2: 添加 fmt import**

检查文件开头 import 部分，确保包含 `"fmt"`。若无则添加。

- [ ] **Step 3: 构建验证**

Run: `cd /home/zcz/program/ai-writer-go && go build -o ai-writer .`
Expected: 编译成功

- [ ] **Step 4: Commit**

```bash
git add internal/service/graph.go internal/api/handler/handler.go
git commit -m "$(cat <<'EOF'
feat: add BuildCausalGraph method and extend GetEChartsData handler

- Add EmotionGraphData struct for emotion arc visualization
- Extend GetEChartsData to support type parameter
- Implement BuildCausalGraph for causal chain visualization

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

---

### Task 4: 实现 BuildForeshadowGraph 方法

**Files:**
- Modify: `internal/service/graph.go` - 在 BuildCausalGraph 后添加

- [ ] **Step 1: 添加 BuildForeshadowGraph 方法**

在 BuildCausalGraph 方法后添加：

```go
// BuildForeshadowGraph 构建伏笔追踪图谱
func (s *GraphService) BuildForeshadowGraph(bookName string) (*GraphData, error) {
	data := &GraphData{
		Nodes: []GraphNode{},
		Links: []GraphLink{},
		Categories: []Category{
			{Name: "foreshadow"},
			{Name: "chapter"},
		},
	}

	// 状态颜色映射
	statusColorMap := map[string]string{
		"active":    "#5470c6", // 蓝色 - 埋设中
		"resolved":  "#91cc75", // 绿色 - 已回收
		"expired":   "#fac858", // 橙色 - 过期预警
		"abandoned": "#aaaaaa", // 灰色 - 已放弃
	}

	// 重要程度大小映射
	importanceSizeMap := map[string]int{
		"high":   40,
		"medium": 30,
		"low":    25,
	}

	nodeNames := make(map[string]bool)

	addNode := func(name, category string, symbolSize int, value string, status string) {
		if nodeNames[name] {
			return
		}
		nodeNames[name] = true
		node := GraphNode{
			Name:       name,
			Category:   category,
			SymbolSize: symbolSize,
			Value:      truncateStr(value, 50),
		}
		if category == "foreshadow" {
			node.ItemStyle.Color = statusColorMap[status]
		} else {
			node.ItemStyle.Color = "#91cc75"
		}
		node.Label.Show = true
		node.Label.Position = "right"
		data.Nodes = append(data.Nodes, node)
	}

	addLink := func(source, target, relation string, isDashed bool, color string) {
		if source == "" || target == "" || source == target {
			return
		}
		link := GraphLink{
			Source: source,
			Target: target,
			Value:  relation,
		}
		if isDashed {
			link.LineStyle.Type = "dashed"
		} else {
			link.LineStyle.Type = "solid"
		}
		link.LineStyle.Color = color
		link.LineStyle.Curveness = 0.2
		data.Links = append(data.Links, link)
	}

	// 加载伏笔数据
	foreshadows, err := s.store.LoadForeshadows(bookName)
	if err != nil {
		return data, nil
	}

	for _, fs := range foreshadows {
		// 伏笔节点
		fsLabel := truncateStr(fs.Content, 25)
		size := importanceSizeMap[string(fs.Importance)]
		if size == 0 {
			size = 30
		}
		addNode(fsLabel, "foreshadow", size, fs.Content, string(fs.Status))

		// 埋设章节节点
		sourceChLabel := fmt.Sprintf("第%d章", fs.SourceChapter)
		addNode(sourceChLabel, "chapter", 25, "", "")
		addLink(fsLabel, sourceChLabel, "埋设", false, "#5470c6")

		// 回收章节（已回收时）
		if fs.Status == model.ForeshadowResolved && fs.ResolvedChapter > 0 {
			resolvedChLabel := fmt.Sprintf("第%d章", fs.ResolvedChapter)
			addNode(resolvedChLabel, "chapter", 25, "", "")
			addLink(fsLabel, resolvedChLabel, "回收", false, "#91cc75")
		}

		// 预期回收章节（埋设中时，虚线显示）
		if fs.Status == model.ForeshadowActive && fs.TargetChapter > 0 {
			targetChLabel := fmt.Sprintf("第%d章(预期)", fs.TargetChapter)
			addNode(targetChLabel, "chapter", 20, "", "")
			addLink(fsLabel, targetChLabel, "预期", true, "#fac858")
		}
	}

	return data, nil
}
```

- [ ] **Step 2: 确保 model import 包含 ForeshadowStatus 类型**

检查 import 是否包含 `"ai-writer/internal/model"`，代码中使用了 `model.ForeshadowResolved` 和 `model.ForeshadowActive`。

- [ ] **Step 3: 构建验证**

Run: `cd /home/zcz/program/ai-writer-go && go build -o ai-writer .`
Expected: 编译成功

- [ ] **Step 4: Commit**

```bash
git add internal/service/graph.go
git commit -m "$(cat <<'EOF'
feat: add BuildForeshadowGraph method

Implement foreshadow tracking graph with:
- Status-based node coloring (active/resolved/expired/abandoned)
- Importance-based node sizing
- Links to source/resolved/target chapters

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

---

### Task 5: 实现 BuildThreadGraph 方法

**Files:**
- Modify: `internal/service/graph.go` - 在 BuildForeshadowGraph 后添加

- [ ] **Step 1: 添加 BuildThreadGraph 方法**

在 BuildForeshadowGraph 方法后添加：

```go
// BuildThreadGraph 构建叙事线程图谱
func (s *GraphService) BuildThreadGraph(bookName string) (*GraphData, error) {
	data := &GraphData{
		Nodes: []GraphNode{},
		Links: []GraphLink{},
		Categories: []Category{
			{Name: "thread"},
			{Name: "chapter"},
			{Name: "character"},
		},
	}

	// 线程类型颜色映射
	threadTypeColorMap := map[string]string{
		"main":      "#ee6666", // 红色 - 主线
		"sub":       "#5470c6", // 蓝色 - 支线
		"parallel":  "#91cc75", // 绿色 - 并行线
		"flashback": "#9a60b4", // 紫色 - 闪回线
	}

	// 线程类型大小映射
	threadTypeSizeMap := map[string]int{
		"main":      45,
		"sub":       35,
		"parallel":  30,
		"flashback": 30,
	}

	nodeNames := make(map[string]bool)

	addNode := func(name, category string, symbolSize int, value string, threadType string) {
		if nodeNames[name] {
			return
		}
		nodeNames[name] = true
		node := GraphNode{
			Name:       name,
			Category:   category,
			SymbolSize: symbolSize,
			Value:      truncateStr(value, 30),
		}
		if category == "thread" {
			node.ItemStyle.Color = threadTypeColorMap[threadType]
		} else if category == "chapter" {
			node.ItemStyle.Color = "#91cc75"
		} else {
			node.ItemStyle.Color = "#5470c6"
		}
		node.Label.Show = true
		node.Label.Position = "right"
		data.Nodes = append(data.Nodes, node)
	}

	addLink := func(source, target, relation string) {
		if source == "" || target == "" || source == target {
			return
		}
		link := GraphLink{
			Source: source,
			Target: target,
			Value:  relation,
			Symbol: []string{"none", "arrow"},
		}
		link.LineStyle.Type = "solid"
		link.LineStyle.Color = "source"
		link.LineStyle.Curveness = 0.2
		data.Links = append(data.Links, link)
	}

	// 加载线程数据
	threads, err := s.store.LoadThreads(bookName)
	if err != nil {
		return data, nil
	}

	// 加载角色数据（用于 POV 连接）
	characters, _ := s.store.LoadCharacters(bookName)
	characterMap := make(map[string]bool)
	for _, char := range characters {
		characterMap[char.Name] = true
	}

	for _, thread := range threads {
		// 线程节点
		size := threadTypeSizeMap[string(thread.Type)]
		if size == 0 {
			size = 30
		}
		if thread.Weight > 0 {
			size += thread.Weight * 5
		}
		addNode(thread.Name, "thread", size, thread.Goal, string(thread.Type))

		// 涉及章节
		for _, chID := range thread.Chapters {
			chLabel := fmt.Sprintf("第%d章", chID)
			addNode(chLabel, "chapter", 25, "", "")
			addLink(thread.Name, chLabel, "涉及")
		}

		// POV 角色
		for _, povChar := range thread.POVCharacters {
			if characterMap[povChar] {
				addNode(povChar, "character", 35, "", "")
				addLink(thread.Name, povChar, "视角")
			}
		}
	}

	return data, nil
}
```

- [ ] **Step 2: 构建验证**

Run: `cd /home/zcz/program/ai-writer-go && go build -o ai-writer .`
Expected: 编译成功

- [ ] **Step 3: Commit**

```bash
git add internal/service/graph.go
git commit -m "$(cat <<'EOF'
feat: add BuildThreadGraph method

Implement narrative thread graph with:
- Thread type-based coloring (main/sub/parallel/flashback)
- Weight and type-based node sizing
- Links to involved chapters and POV characters

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

---

### Task 6: 实现 BuildEmotionGraph 方法

**Files:**
- Modify: `internal/service/graph.go` - 在 BuildThreadGraph 后添加

- [ ] **Step 1: 添加 BuildEmotionGraph 方法**

在 BuildThreadGraph 方法后添加：

```go
// BuildEmotionGraph 构建情感弧线图谱
func (s *GraphService) BuildEmotionGraph(bookName string) (*EmotionGraphData, error) {
	data := &EmotionGraphData{
		Type:      "emotion",
		ChartType: "line",
		Data:      []EmotionArcData{},
		Metadata: EmotionMetadata{
			Characters:   []string{},
			ChapterRange: []int{0, 0},
			EmotionTypes: []string{},
		},
	}

	// 加载角色数据
	characters, err := s.store.LoadCharacters(bookName)
	if err != nil {
		return data, nil
	}

	emotionSet := make(map[string]bool)
	minChapter := 0
	maxChapter := 0

	for _, char := range characters {
		// 只处理有情感弧线数据的角色
		if len(char.EmotionalArc) == 0 {
			continue
		}

		arcData := EmotionArcData{
			Character: char.Name,
			Points:    []EmotionPoint{},
		}

		// 转换情感点数据
		for _, ep := range char.EmotionalArc {
			arcData.Points = append(arcData.Points, EmotionPoint{
				Chapter:   ep.ChapterID,
				Emotion:   ep.Emotion,
				Intensity: ep.Intensity,
				Trigger:   ep.Trigger,
			})

			// 记录情感类型
			emotionSet[ep.Emotion] = true

			// 更新章节范围
			if minChapter == 0 || ep.ChapterID < minChapter {
				minChapter = ep.ChapterID
			}
			if ep.ChapterID > maxChapter {
				maxChapter = ep.ChapterID
			}
		}

		if len(arcData.Points) > 0 {
			data.Data = append(data.Data, arcData)
			data.Metadata.Characters = append(data.Metadata.Characters, char.Name)
		}
	}

	// 设置元数据
	data.Metadata.ChapterRange = []int{minChapter, maxChapter}
	for emotion := range emotionSet {
		data.Metadata.EmotionTypes = append(data.Metadata.EmotionTypes, emotion)
	}

	return data, nil
}
```

- [ ] **Step 2: 构建验证**

Run: `cd /home/zcz/program/ai-writer-go && go build -o ai-writer .`
Expected: 编译成功

- [ ] **Step 3: Commit**

```bash
git add internal/service/graph.go
git commit -m "$(cat <<'EOF'
feat: add BuildEmotionGraph method

Implement emotion arc graph with:
- Line chart format for emotion intensity over chapters
- Metadata with character list, chapter range, and emotion types
- Data extraction from character emotional_arc field

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

---

### Task 7: 实现 BuildTimelineGraph 方法

**Files:**
- Modify: `internal/service/graph.go` - 在 BuildEmotionGraph 后添加

- [ ] **Step 1: 添加 BuildTimelineGraph 方法**

在 BuildEmotionGraph 方法后添加：

```go
// BuildTimelineGraph 构建时间线图谱
func (s *GraphService) BuildTimelineGraph(bookName string) (*GraphData, error) {
	data := &GraphData{
		Nodes: []GraphNode{},
		Links: []GraphLink{},
		Categories: []Category{
			{Name: "event"},
			{Name: "chapter"},
			{Name: "location"},
			{Name: "character"},
		},
	}

	colorMap := map[string]string{
		"event":     "#5470c6", // 蓝色
		"chapter":   "#91cc75", // 绿色
		"location":  "#fac858", // 黄色
		"character": "#ee6666", // 红色
	}

	nodeNames := make(map[string]bool)

	addNode := func(name, category string, symbolSize int, value string) {
		if nodeNames[name] {
			return
		}
		nodeNames[name] = true
		node := GraphNode{
			Name:       name,
			Category:   category,
			SymbolSize: symbolSize,
			Value:      truncateStr(value, 30),
		}
		node.ItemStyle.Color = colorMap[category]
		node.Label.Show = true
		node.Label.Position = "right"
		data.Nodes = append(data.Nodes, node)
	}

	addLink := func(source, target, relation string, isDashed bool) {
		if source == "" || target == "" || source == target {
			return
		}
		link := GraphLink{
			Source: source,
			Target: target,
			Value:  relation,
		}
		if isDashed {
			link.LineStyle.Type = "dashed"
		} else {
			link.LineStyle.Type = "solid"
		}
		link.LineStyle.Color = "source"
		link.LineStyle.Curveness = 0.2
		data.Links = append(data.Links, link)
	}

	// 加载时间线数据
	timeline, err := s.store.LoadTimeline(bookName)
	if err != nil {
		return data, nil
	}

	// 加载角色和地点用于关联
	characters, _ := s.store.LoadCharacters(bookName)
	characterMap := make(map[string]bool)
	for _, char := range characters {
		characterMap[char.Name] = true
	}

	locations, _ := s.store.LoadLocations(bookName)
	locationMap := make(map[string]bool)
	for _, loc := range locations {
		locationMap[loc.Name] = true
	}

	// 按章节排序的事件
	chapterEvents := make(map[int][]model.TimelineEvent)
	for _, event := range timeline {
		chapterEvents[event.ChapterID] = append(chapterEvents[event.ChapterID], event)
	}

	// 构建章节节点和时间顺序链接
	var prevChapterLabel string
	for chID := 1; chID <= len(chapterEvents); chID++ {
		events := chapterEvents[chID]
		if len(events) == 0 {
			continue
		}

		chLabel := fmt.Sprintf("第%d章", chID)
		// 添加时间标签作为值
		timeLabel := events[0].TimeLabel
		addNode(chLabel, "chapter", 30, timeLabel)

		// 章节间时间顺序链接
		if prevChapterLabel != "" {
			addLink(prevChapterLabel, chLabel, "时间顺序", true)
		}
		prevChapterLabel = chLabel

		// 添加章节内的事件
		for _, event := range events {
			for _, evText := range event.Events {
				eventLabel := fmt.Sprintf("第%d章: %s", chID, truncateStr(evText, 20))
				addNode(eventLabel, "event", 35, evText)

				// 事件 → 章节
				addLink(eventLabel, chLabel, "归属", false)

				// 事件 → 地点
				if event.Location != "" && locationMap[event.Location] {
					addNode(event.Location, "location", 25, "")
					addLink(eventLabel, event.Location, "发生地点", false)
				}

				// 事件 → 角色
				for _, charName := range event.Characters {
					if characterMap[charName] {
						addNode(charName, "character", 35, "")
						addLink(eventLabel, charName, "参与角色", false)
					}
				}
			}
		}
	}

	return data, nil
}
```

- [ ] **Step 2: 构建验证**

Run: `cd /home/zcz/program/ai-writer-go && go build -o ai-writer .`
Expected: 编译成功

- [ ] **Step 3: 启动服务器验证**

Run: `cd /home/zcz/program/ai-writer-go && ./ai-writer server &`
访问: `http://localhost:8081/api/books/<book_id>/graph/echarts?type=causal`
Expected: 返回 JSON 数据或空图谱

- [ ] **Step 4: Commit**

```bash
git add internal/service/graph.go
git commit -m "$(cat <<'EOF'
feat: add BuildTimelineGraph method

Implement timeline graph with:
- Chapter nodes with time labels
- Events linked to chapters, locations, and characters
- Sequential chapter links showing time flow

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

---

## Phase 3: 前端 API 扩展

### Task 8: 扩展前端 graphApi

**Files:**
- Modify: `web/src/api/index.js:168-171`

- [ ] **Step 1: 修改 graphApi 支持类型参数**

将 `web/src/api/index.js` 第 168-171 行替换为：

```javascript
export const graphApi = {
  get: (bookId) => api.get(`/books/${bookId}/graph`),
  getECharts: (bookId, type = 'relationship') => api.get(`/books/${bookId}/graph/echarts`, { params: { type } }),
  getRelationship: (bookId) => api.get(`/books/${bookId}/graph/echarts`, { params: { type: 'relationship' } }),
  getCausal: (bookId) => api.get(`/books/${bookId}/graph/echarts`, { params: { type: 'causal' } }),
  getForeshadow: (bookId) => api.get(`/books/${bookId}/graph/echarts`, { params: { type: 'foreshadow' } }),
  getThread: (bookId) => api.get(`/books/${bookId}/graph/echarts`, { params: { type: 'thread' } }),
  getEmotion: (bookId) => api.get(`/books/${bookId}/graph/echarts`, { params: { type: 'emotion' } }),
  getTimeline: (bookId) => api.get(`/books/${bookId}/graph/echarts`, { params: { type: 'timeline' } })
}
```

- [ ] **Step 2: Commit**

```bash
git add web/src/api/index.js
git commit -m "$(cat <<'EOF'
feat: extend graphApi with type parameter support

Add convenience methods for each graph type:
- getRelationship, getCausal, getForeshadow
- getThread, getEmotion, getTimeline

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

---

## Phase 4: 前端容器组件

### Task 9: 创建 GraphCard 组件

**Files:**
- Create: `web/src/components/GraphCard.vue`

- [ ] **Step 1: 创建 GraphCard.vue**

```vue
<template>
  <div
    class="graph-card"
    :class="{ 'minimized': isMinimized, 'maximized': isMaximized }"
    ref="cardRef"
  >
    <div class="card-header" @mousedown="startDrag">
      <span class="card-title">{{ title }}</span>
      <div class="card-stats" v-if="isMinimized">
        <span>{{ stats.nodes }} 节点</span>
        <span>{{ stats.links }} 条关系</span>
      </div>
      <div class="card-controls">
        <el-button size="small" text @click.stop="toggleMinimize">
          <el-icon><component :is="isMinimized ? 'Expand' : 'Compress'" /></el-icon>
        </el-button>
        <el-button size="small" text @click.stop="toggleMaximize">
          <el-icon><component :is="isMaximized ? 'Shrink' : 'FullScreen'" /></el-icon>
        </el-button>
        <el-button size="small" text @click.stop="$emit('close')">
          <el-icon><Close /></el-icon>
        </el-button>
      </div>
    </div>
    <div class="card-content" v-show="!isMinimized">
      <div ref="chartContainer" class="chart-container"></div>
      <div v-if="loading" class="loading-overlay">
        <el-icon class="is-loading"><Loading /></el-icon>
      </div>
      <div v-if="error" class="error-overlay">
        <el-icon><Warning /></el-icon>
        <span>{{ error }}</span>
      </div>
      <div v-if="isEmpty" class="empty-overlay">
        <span>暂无数据</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { Close, Expand, Compress, FullScreen, Shrink, Loading, Warning } from '@element-plus/icons-vue'
import * as echarts from 'echarts'

const props = defineProps({
  title: String,
  graphType: String,
  bookId: String,
  graphData: Object,
  loading: Boolean,
  error: String
})

const emit = defineEmits(['close', 'dragStart', 'minimize', 'maximize'])

const cardRef = ref(null)
const chartContainer = ref(null)
const isMinimized = ref(false)
const isMaximized = ref(false)
let chartInstance = null

const stats = computed(() => ({
  nodes: props.graphData?.nodes?.length || 0,
  links: props.graphData?.links?.length || 0
}))

const isEmpty = computed(() => {
  return !props.loading && !props.error && stats.value.nodes === 0
})

const startDrag = (e) => {
  emit('dragStart', e)
}

const toggleMinimize = () => {
  isMinimized.value = !isMinimized.value
  emit('minimize', isMinimized.value)
}

const toggleMaximize = () => {
  isMaximized.value = !isMaximized.value
  emit('maximize', isMaximized.value)
}

const renderChart = async () => {
  if (!chartContainer.value || !props.graphData) return

  await nextTick()

  if (chartInstance) {
    chartInstance.dispose()
  }

  if (props.graphType === 'emotion') {
    renderEmotionChart()
  } else {
    renderForceGraph()
  }
}

const renderForceGraph = () => {
  chartInstance = echarts.init(chartContainer.value)

  const nodes = props.graphData.nodes || []
  const links = props.graphData.links || []

  if (nodes.length === 0) {
    chartInstance.setOption({
      title: { text: '暂无数据', left: 'center', top: 'center' }
    })
    return
  }

  const option = {
    tooltip: {
      trigger: 'item',
      formatter: (params) => {
        if (params.dataType === 'node') {
          return `${params.data.name}<br/>类型: ${params.data.category}<br/>${params.data.value || ''}`
        } else {
          return `${params.data.source} → ${params.data.target}<br/>关系: ${params.data.value}`
        }
      }
    },
    series: [{
      type: 'graph',
      layout: 'force',
      data: nodes.map(n => ({
        name: n.name,
        category: n.category,
        symbolSize: n.symbolSize || 30,
        value: n.value,
        itemStyle: n.itemStyle || { color: '#5470c6' }
      })),
      links: links.map(l => ({
        source: l.source,
        target: l.target,
        value: l.value,
        lineStyle: l.lineStyle || { type: 'solid', color: 'source' },
        symbol: l.symbol || 'none'
      })),
      roam: true,
      draggable: true,
      label: { show: true, position: 'right', fontSize: 12 },
      force: { repulsion: 200, edgeLength: 120 },
      lineStyle: { curveness: 0.2 }
    }]
  }

  chartInstance.setOption(option, true)
}

const renderEmotionChart = () => {
  chartInstance = echarts.init(chartContainer.value)

  const data = props.graphData.data || []

  if (data.length === 0) {
    chartInstance.setOption({
      title: { text: '暂无情感数据', left: 'center', top: 'center' }
    })
    return
  }

  const series = data.map(arc => ({
    name: arc.character,
    type: 'line',
    data: arc.points.map(p => [p.chapter, p.intensity]),
    smooth: true,
    markPoint: {
      data: arc.points.map(p => ({
        coord: [p.chapter, p.intensity],
        name: p.emotion,
        symbol: 'circle',
        symbolSize: 10,
        itemStyle: { color: getEmotionColor(p.emotion) }
      }))
    }
  }))

  const option = {
    tooltip: {
      trigger: 'item',
      formatter: (params) => {
        const char = params.seriesName
        const chapter = params.data[0]
        const intensity = params.data[1]
        const point = data.find(d => d.character === char)?.points.find(p => p.chapter === chapter)
        if (point) {
          return `${char}<br/>第${chapter}章<br/>情感: ${point.emotion}<br/>强度: ${intensity}<br/>触发: ${point.trigger}`
        }
        return `${char} 第${chapter}章 强度: ${intensity}`
      }
    },
    legend: { data: data.map(d => d.character), top: 10 },
    xAxis: { type: 'value', name: '章节' },
    yAxis: { type: 'value', name: '强度', min: 0, max: 10 },
    series: series
  }

  chartInstance.setOption(option, true)
}

const getEmotionColor = (emotion) => {
  const colors = {
    '喜悦': '#91cc75',
    '愤怒': '#ee6666',
    '悲伤': '#5470c6',
    '恐惧': '#9a60b4',
    '惊讶': '#fac858'
  }
  return colors[emotion] || '#5470c6'
}

const handleResize = () => {
  if (chartInstance) {
    chartInstance.resize()
  }
}

watch(() => props.graphData, renderChart, { deep: true })
watch(isMaximized, () => {
  nextTick(() => {
    handleResize()
  })
})

onMounted(() => {
  window.addEventListener('resize', handleResize)
  if (props.graphData) {
    renderChart()
  }
})

onUnmounted(() => {
  if (chartInstance) {
    chartInstance.dispose()
  }
  window.removeEventListener('resize', handleResize)
})
</script>

<style scoped>
.graph-card {
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  transition: all 0.3s ease;
}

.graph-card.minimized {
  height: 60px;
}

.graph-card.maximized {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 1000;
  border-radius: 0;
}

.card-header {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  background: #f5f7fa;
  cursor: move;
  user-select: none;
}

.card-title {
  font-weight: 600;
  flex: 1;
}

.card-stats {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: #909399;
}

.card-controls {
  display: flex;
  gap: 4px;
}

.card-content {
  height: 300px;
  position: relative;
}

.graph-card.maximized .card-content {
  height: calc(100vh - 60px);
}

.chart-container {
  width: 100%;
  height: 100%;
}

.loading-overlay,
.error-overlay,
.empty-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  background: rgba(255, 255, 255, 0.9);
}

.error-overlay {
  color: #f56c6c;
}

.empty-overlay {
  color: #909399;
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add web/src/components/GraphCard.vue
git commit -m "$(cat <<'EOF'
feat: create GraphCard component

Implement graph card with:
- Minimize/maximize/close controls
- Drag handle for reorder
- ECharts rendering for force and emotion graphs
- Loading/error/empty states

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

---

### Task 10: 创建 GraphGrid 组件

**Files:**
- Create: `web/src/components/GraphGrid.vue`

- [ ] **Step 1: 安装 vue-draggable-plus**

Run: `cd /home/zcz/program/ai-writer-go/web && npm install vue-draggable-plus`

- [ ] **Step 2: 创建 GraphGrid.vue**

```vue
<template>
  <div class="graph-grid" :class="gridClass">
    <VueDraggable
      v-model="cards"
      :animation="200"
      handle=".card-header"
      ghostClass="ghost-card"
      @end="onDragEnd"
      class="draggable-container"
    >
      <div
        v-for="card in cards"
        :key="card.type"
        class="grid-item"
        :class="{ 'maximized-slot': maximizedCard === card.type }"
      >
        <GraphCard
          :title="card.title"
          :graphType="card.type"
          :bookId="bookId"
          :graphData="graphDataMap[card.type]"
          :loading="loadingMap[card.type]"
          :error="errorMap[card.type]"
          :ref="el => setCardRef(card.type, el)"
          @close="closeCard(card.type)"
          @dragStart="onCardDragStart"
          @minimize="onCardMinimize(card.type, $event)"
          @maximize="onCardMaximize(card.type, $event)"
        />
      </div>
    </VueDraggable>

    <!-- 图谱开关面板 -->
    <div class="graph-panel">
      <div class="panel-header">
        <span>图谱类型</span>
      </div>
      <div class="panel-options">
        <el-checkbox
          v-for="option in allGraphTypes"
          :key="option.type"
          :model-value="isCardOpen(option.type)"
          @change="toggleCard(option.type, $event)"
          :label="option.title"
        >
          <span class="option-label">
            {{ option.title }}
            <span v-if="isCardOpen(option.type)" class="option-stats">
              ({{ graphDataMap[option.type]?.nodes?.length || 0 }}节点)
            </span>
          </span>
        </el-checkbox>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { VueDraggable } from 'vue-draggable-plus'
import GraphCard from './GraphCard.vue'
import { graphApi } from '@/api'

const props = defineProps({
  bookId: String
})

const allGraphTypes = [
  { type: 'relationship', title: '基础关系图' },
  { type: 'causal', title: '剧情因果图' },
  { type: 'foreshadow', title: '伏笔追踪图' },
  { type: 'thread', title: '叙事线程图' },
  { type: 'emotion', title: '情感弧线图' },
  { type: 'timeline', title: '时间线图' }
]

const cards = ref([])
const graphDataMap = ref({})
const loadingMap = ref({})
const errorMap = ref({})
const maximizedCard = ref(null)
const cardRefs = ref({})

const gridClass = computed(() => {
  const count = cards.value.length
  if (count <= 2) return 'grid-2'
  if (count <= 4) return 'grid-2'
  return 'grid-3'
})

const STORAGE_KEY = computed(() => `graph-layout-${props.bookId}`)

const isCardOpen = (type) => {
  return cards.value.some(c => c.type === type)
}

const toggleCard = (type, open) => {
  if (open) {
    openCard(type)
  } else {
    closeCard(type)
  }
}

const openCard = (type) => {
  if (isCardOpen(type)) return

  const option = allGraphTypes.find(o => o.type === type)
  cards.value.push({ type, title: option.title })
  loadGraphData(type)
  saveLayout()
}

const closeCard = (type) => {
  const index = cards.value.findIndex(c => c.type === type)
  if (index >= 0) {
    cards.value.splice(index, 1)
    delete graphDataMap.value[type]
    delete loadingMap.value[type]
    delete errorMap.value[type]
    saveLayout()
  }
}

const loadGraphData = async (type) => {
  loadingMap.value[type] = true
  errorMap.value[type] = null

  try {
    const res = await graphApi.getECharts(props.bookId, type)
    graphDataMap.value[type] = res.data
  } catch (e) {
    errorMap.value[type] = e.message || '加载失败'
  } finally {
    loadingMap.value[type] = false
  }
}

const loadAllGraphData = () => {
  cards.value.forEach(card => {
    loadGraphData(card.type)
  })
}

const onDragEnd = () => {
  saveLayout()
}

const onCardDragStart = (e) => {
  // 拖拽已在 VueDraggable 中处理
}

const onCardMinimize = (type, minimized) => {
  // 最小化状态由 GraphCard 内部管理
}

const onCardMaximize = (type, maximized) => {
  if (maximized) {
    maximizedCard.value = type
  } else {
    maximizedCard.value = null
  }
}

const setCardRef = (type, el) => {
  if (el) {
    cardRefs.value[type] = el
  }
}

const saveLayout = () => {
  const layout = cards.value.map(c => c.type)
  localStorage.setItem(STORAGE_KEY.value, JSON.stringify(layout))
}

const loadLayout = () => {
  try {
    const saved = localStorage.getItem(STORAGE_KEY.value)
    if (saved) {
      const types = JSON.parse(saved)
      cards.value = types.map(type => {
        const option = allGraphTypes.find(o => o.type === type)
        return { type, title: option?.title || type }
      })
    } else {
      // 默认打开基础关系图
      cards.value = [{ type: 'relationship', title: '基础关系图' }]
    }
  } catch (e) {
    // localStorage 损坏，使用默认布局
    cards.value = [{ type: 'relationship', title: '基础关系图' }]
  }
}

watch(() => props.bookId, () => {
  loadLayout()
  loadAllGraphData()
})

onMounted(() => {
  loadLayout()
  loadAllGraphData()
})
</script>

<style scoped>
.graph-grid {
  display: flex;
  gap: 20px;
}

.draggable-container {
  flex: 1;
  display: grid;
  gap: 20px;
}

.grid-3 .draggable-container {
  grid-template-columns: repeat(3, 1fr);
}

.grid-2 .draggable-container {
  grid-template-columns: repeat(2, 1fr);
}

@media (max-width: 1200px) {
  .grid-3 .draggable-container {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 800px) {
  .draggable-container {
    grid-template-columns: 1fr;
  }
}

.grid-item {
  min-height: 360px;
}

.maximized-slot {
  grid-column: 1 / -1;
  min-height: auto;
}

.ghost-card {
  opacity: 0.5;
  background: #f0f0f0;
}

.graph-panel {
  width: 200px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  padding: 16px;
}

.panel-header {
  font-weight: 600;
  margin-bottom: 12px;
}

.panel-options {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.option-label {
  display: flex;
  align-items: center;
  gap: 4px;
}

.option-stats {
  font-size: 12px;
  color: #909399;
}

@media (max-width: 1200px) {
  .graph-grid {
    flex-direction: column;
  }

  .graph-panel {
    width: 100%;
    order: -1;
  }

  .panel-options {
    flex-direction: row;
    flex-wrap: wrap;
  }
}
</style>
```

- [ ] **Step 3: Commit**

```bash
git add web/src/components/GraphGrid.vue web/package.json web/package-lock.json
git commit -m "$(cat <<'EOF'
feat: create GraphGrid component with drag-and-drop

Implement grid layout with:
- VueDraggable for card reordering
- Responsive grid (3/2/1 columns)
- Side panel for graph type toggling
- localStorage for layout persistence

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

---

## Phase 5: 重构 GraphView 主页面

### Task 11: 重构 GraphView.vue

**Files:**
- Modify: `web/src/views/GraphView.vue`

- [ ] **Step 1: 重构 GraphView.vue**

将整个文件替换为：

```vue
<template>
  <div class="graph-view">
    <div class="page-header">
      <el-button @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <h2>{{ bookId }} - 知识图谱</h2>
      <div class="header-actions">
        <el-button @click="refreshAll">
          <el-icon><Refresh /></el-icon>
          刷新全部
        </el-button>
      </div>
    </div>

    <GraphGrid :bookId="bookId" ref="gridRef" />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ArrowLeft, Refresh } from '@element-plus/icons-vue'
import GraphGrid from '@/components/GraphGrid.vue'

const router = useRouter()
const route = useRoute()
const bookId = computed(() => route.params.id)
const gridRef = ref(null)

const goBack = () => {
  router.push(`/books/${bookId.value}`)
}

const refreshAll = () => {
  gridRef.value?.loadAllGraphData()
}
</script>

<style scoped>
.graph-view {
  max-width: 1600px;
  margin: 0 auto;
  padding: 20px;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  flex: 1;
}

.header-actions {
  display: flex;
  gap: 10px;
}
</style>
```

- [ ] **Step 2: 构建前端**

Run: `cd /home/zcz/program/ai-writer-go/web && npm run build`
Expected: 构建成功

- [ ] **Step 3: Commit**

```bash
git add web/src/views/GraphView.vue
git commit -m "$(cat <<'EOF'
refactor: GraphView to use GraphGrid component

Simplify GraphView to:
- Header with back/refresh buttons
- GraphGrid for all graph rendering
- Remove old single-graph logic

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

---

## Phase 6: 整合测试

### Task 12: 整合测试

- [ ] **Step 1: 启动后端服务**

Run: `cd /home/zcz/program/ai-writer-go && ./ai-writer server`

- [ ] **Step 2: 访问图谱页面**

访问: `http://localhost:8081`，进入某书籍的图谱页面
Expected:
- 页面显示网格布局
- 右侧显示图谱类型开关面板
- 默认显示基础关系图

- [ ] **Step 3: 测试图谱切换**

操作: 勾选其他图谱类型（因果图、伏笔图等）
Expected:
- 新卡片出现
- 加载对应数据并渲染
- 空数据显示"暂无数据"

- [ ] **Step 4: 测试拖拽**

操作: 拖拽卡片标题栏调整位置
Expected:
- 卡片位置交换
- 刷新页面后位置保持

- [ ] **Step 5: 测试最小化/最大化**

操作: 点击卡片上的最小化/最大化按钮
Expected:
- 最小化后只显示标题和统计
- 最大化后卡片撑满屏幕

- [ ] **Step 6: 测试响应式**

操作: 调整浏览器窗口宽度
Expected:
- 宽屏 3 列，中屏 2 列，窄屏 1 列

- [ ] **Step 7: Final Commit**

```bash
git add -A
git commit -m "$(cat <<'EOF'
feat: complete multi-view knowledge graph system

Implement 6 graph types with grid layout:
- relationship: base entity relations
- causal: cause-effect chains
- foreshadow: foreshadowing tracking
- thread: narrative threads
- emotion: emotion arcs (line chart)
- timeline: time-based events

Features:
- Drag-and-drop card reordering
- Minimize/maximize cards
- Responsive grid layout
- localStorage persistence

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

---

## 自检清单

**Spec 覆盖检查:**
- [x] 6 种图谱类型全部实现
- [x] 后端 API type 参数支持
- [x] 前端网格布局 + 拖拽
- [x] 最小化/最大化功能
- [x] 响应式布局
- [x] localStorage 持久化
- [x] 空数据/错误处理

**Placeholder 检查:** 无 TBD、TODO、"add validation" 等占位符。

**类型一致性检查:**
- EmotionGraphData 结构体定义与 BuildEmotionGraph 返回类型匹配
- graphApi 方法名与 handler 中 type 参数值匹配
- GraphCard props 与 GraphGrid 传递的数据匹配