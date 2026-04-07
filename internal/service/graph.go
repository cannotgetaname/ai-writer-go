package service

import (
	"ai-writer/internal/model"
	"ai-writer/internal/store"
	"fmt"
	"sort"
)

// GraphService 知识图谱服务
type GraphService struct {
	store *store.JSONStore
}

// NewGraphService 创建图谱服务
func NewGraphService(store *store.JSONStore) *GraphService {
	return &GraphService{store: store}
}

// GraphData 图谱数据
type GraphData struct {
	Nodes      []GraphNode `json:"nodes"`
	Links      []GraphLink `json:"links"`
	Categories []Category  `json:"categories"`
}

// GraphNode 图谱节点
type GraphNode struct {
	Name       string `json:"name"`
	Category   string `json:"category"`
	SymbolSize int    `json:"symbolSize"`
	Value      string `json:"value"`
	ItemStyle  struct {
		Color string `json:"color"`
	} `json:"itemStyle"`
	Label struct {
		Show     bool   `json:"show"`
		Position string `json:"position"`
	} `json:"label"`
}

// GraphLink 图谱边
type GraphLink struct {
	Source  string  `json:"source"`
	Target  string  `json:"target"`
	Value   string  `json:"value"`
	LineStyle struct {
		Type     string  `json:"type"`     // solid/dashed/dotted
		Color    string  `json:"color"`
		Curveness float64 `json:"curveness"`
	} `json:"lineStyle"`
	Symbol []string `json:"symbol,omitempty"` // 箭头类型: ['none', 'arrow'] 表示单向箭头
}

// Category 分类
type Category struct {
	Name string `json:"name"`
}

// EmotionGraphData 情感图谱数据（折线图）
// 返回 ECharts 期望的格式：series/xAxis
type EmotionGraphData struct {
	Series []EmotionSeries `json:"series"`
	XAxis  []string        `json:"xAxis"`
}

// EmotionSeries 情感系列数据
type EmotionSeries struct {
	Name string `json:"name"`
	Data []int  `json:"data"`
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
	Characters   []string `json:"characters"`
	ChapterRange []int    `json:"chapter_range"`
	EmotionTypes []string `json:"emotion_types"`
}

// BuildGraph 构建知识图谱
func (s *GraphService) BuildGraph(bookName string) (*GraphData, error) {
	data := &GraphData{
		Nodes: []GraphNode{},
		Links: []GraphLink{},
		Categories: []Category{
			{Name: "character"},
			{Name: "item"},
			{Name: "location"},
			{Name: "faction"},
		},
	}

	// 颜色配置
	colorMap := map[string]string{
		"character": "#5470c6", // 蓝色
		"item":      "#fac858", // 黄色
		"location":  "#91cc75", // 绿色
		"faction":   "#ee6666", // 红色
	}

	// 势力间关系颜色
	factionRelColorMap := map[string]string{
		"ally":       "#91cc75", // 绿色 - 联盟
		"enemy":      "#ee6666", // 红色 - 敌对
		"subordinate": "#9a60b4", // 紫色 - 附属
		"neutral":    "#aaaaaa", // 灰色 - 中立
	}

	// 节点名集合（避免重复）
	nodeNames := make(map[string]bool)

	// 添加节点的辅助函数
	addNode := func(name, category string, symbolSize int, value string) {
		if nodeNames[name] {
			return
		}
		nodeNames[name] = true
		node := GraphNode{
			Name:       name,
			Category:   category,
			SymbolSize: symbolSize,
			Value:      truncateStr(value, 100),
		}
		node.ItemStyle.Color = colorMap[category]
		node.Label.Show = true
		node.Label.Position = "right"
		data.Nodes = append(data.Nodes, node)
	}

	// 添加归属关系边的辅助函数（子指向父，实线，有箭头）
	addOwnershipLink := func(source, target, relation string) {
		if source == "" || target == "" || source == target {
			return
		}
		link := GraphLink{
			Source: source,
			Target: target,
			Value:  relation,
			Symbol: []string{"none", "arrow"}, // 单向箭头：source端无，target端有
		}
		link.LineStyle.Type = "solid"
		link.LineStyle.Color = "source" // 继承源节点颜色
		link.LineStyle.Curveness = 0.2
		data.Links = append(data.Links, link)
	}

	// 添加双向关系边的辅助函数（无箭头，用于联通、平级关系）
	addBidirectionalLink := func(source, target, relation string, lineType, color string) {
		if source == "" || target == "" || source == target {
			return
		}
		link := GraphLink{
			Source: source,
			Target: target,
			Value:  relation,
			Symbol: []string{"none", "none"}, // 双向无箭头
		}
		link.LineStyle.Type = lineType // dashed/dotted
		link.LineStyle.Color = color
		link.LineStyle.Curveness = 0.1
		data.Links = append(data.Links, link)
	}

	// 加载所有数据
	characters, _ := s.store.LoadCharacters(bookName)
	items, _ := s.store.LoadItems(bookName)
	locations, _ := s.store.LoadLocations(bookName)
	worldview, _ := s.store.LoadWorldView(bookName)

	// ========== 1. 加载势力节点（作为父节点）==========
	for _, faction := range worldview.KeyElements.Factions {
		addNode(faction.Name, "faction", 50, faction.Description)
	}

	// ========== 2. 加载地点节点并计算层级深度 ==========
	// 计算地点层级深度（用于确定节点大小）
	locationDepth := make(map[string]int) // 深度：根节点=0，子节点=1...
	for _, loc := range locations {
		if loc.Parent == "" {
			locationDepth[loc.Name] = 0 // 根节点
		}
	}
	// 递归计算子节点深度
	for _, loc := range locations {
		if loc.Parent != "" {
			depth := calcLocationDepth(loc.Name, locations, locationDepth)
			locationDepth[loc.Name] = depth
		}
	}

	for _, loc := range locations {
		// 节点大小：深度越小（越接近根），节点越大
		// depth 0 -> 40, depth 1 -> 35, depth 2 -> 30...
		size := 40 - locationDepth[loc.Name]*5
		if size < 25 {
			size = 25
		}
		addNode(loc.Name, "location", size, loc.Description)

		// 地点 → 势力（领地归属）- 多重归属都显示
		if loc.Faction != "" {
			addNode(loc.Faction, "faction", 50, "")
			addOwnershipLink(loc.Name, loc.Faction, "领地")
		}

		// 子地点 → 父地点（包含关系）
		if loc.Parent != "" {
			addNode(loc.Parent, "location", 40 - locationDepth[loc.Parent]*5, "")
			addOwnershipLink(loc.Name, loc.Parent, "属于")
		}

		// 地点 ↔ 地点（联通关系）- 虚线，无箭头
		for _, neighbor := range loc.Neighbors {
			addNode(neighbor, "location", 30, "")
			addBidirectionalLink(loc.Name, neighbor, "联通", "dashed", "#91cc75")
		}
	}

	// ========== 3. 加载人物节点 ==========
	for _, char := range characters {
		addNode(char.Name, "character", 35, char.Bio)

		// 人物 → 势力（所属）
		if char.Faction != "" {
			addNode(char.Faction, "faction", 50, "")
			addOwnershipLink(char.Name, char.Faction, "所属")
		}

		// 人物 → 宗门（宗门作为势力节点）
		if char.Sect != "" && char.Sect != char.Faction {
			addNode(char.Sect, "faction", 45, "")
			addOwnershipLink(char.Name, char.Sect, "宗门")
		}

		// 人物间关系
		for _, rel := range char.Relations {
			targetName := rel.TargetName
			if targetName == "" {
				targetName = rel.TargetID
			}
			addNode(targetName, "character", 35, "")

			// 判断是否为上级关系（师徒、父子、上级等）
			if isHierarchicalRelation(rel.Type) {
				// 上级关系：人物 → 上级人物（子指向父）
				addOwnershipLink(char.Name, targetName, rel.Type)
			} else {
				// 平级关系：双向，点线，无箭头
				addBidirectionalLink(char.Name, targetName, rel.Type, "dotted", "#5470c6")
			}
		}
	}

	// ========== 4. 加载物品节点 ==========
	for _, item := range items {
		addNode(item.Name, "item", 25, item.Description)

		// 物品 → 人物（持有者）
		if item.Owner != "" {
			addNode(item.Owner, "character", 35, "")
			addOwnershipLink(item.Name, item.Owner, "持有")
			// 有持有者时，不再添加物品->势力/宗门/地点的边
			// 因为通过持有者可以追溯到其所属势力/地点
		} else {
			// 无持有者时，才显示物品的直接归属关系
			// 物品 → 势力（所属）
			if item.Faction != "" {
				addNode(item.Faction, "faction", 50, "")
				addOwnershipLink(item.Name, item.Faction, "所属")
			}

			// 物品 → 宗门（所属）
			if item.Sect != "" && item.Sect != item.Faction {
				addNode(item.Sect, "faction", 45, "")
				addOwnershipLink(item.Name, item.Sect, "宗门")
			}

			// 物品 → 地点（所在）
			if item.Location != "" {
				addNode(item.Location, "location", 30, "")
				addOwnershipLink(item.Name, item.Location, "所在")
			}
		}
	}

	// ========== 5. 势力间关系 ==========
	for _, faction := range worldview.KeyElements.Factions {
		for _, rel := range faction.Relations {
			addNode(rel.Name, "faction", 50, "")

			// 根据关系类型确定颜色和是否单向
			color := factionRelColorMap[rel.Type]
			if color == "" {
				color = "#aaaaaa" // 默认灰色
			}

			if rel.Type == "subordinate" {
				// 附属关系：子势力 → 父势力（有箭头）
				// 当前势力faction是子势力，rel.Name是父势力
				addOwnershipLink(faction.Name, rel.Name, "附属")
			} else {
				// 其他关系（联盟/敌对/中立）：双向，无箭头
				addBidirectionalLink(faction.Name, rel.Name, rel.Type, "solid", color)
			}
		}

		// 势力首领关系：首领人物 → 势力（统领）
		// 只有当首领在人物列表中存在时才添加边
		if faction.Leader != "" {
			leaderExists := false
			for _, char := range characters {
				if char.Name == faction.Leader {
					leaderExists = true
					break
				}
			}
			if leaderExists {
				addOwnershipLink(faction.Leader, faction.Name, "统领")
			}
		}

		// 势力成员关系：成员 → 势力
		// 只有当成员在人物列表中存在时才添加边
		for _, member := range faction.Members {
			memberExists := false
			for _, char := range characters {
				if char.Name == member {
					memberExists = true
					break
				}
			}
			if memberExists {
				addOwnershipLink(member, faction.Name, "成员")
			}
		}

		// 势力领地关系：地点 → 势力（已在上面的地点部分处理，这里跳过重复）
	}
	// 注：领地关系已经在地点循环中处理了，这里不需要重复添加

	return data, nil
}

// calcLocationDepth 计算地点层级深度
func calcLocationDepth(locName string, locations []*model.Location, depthMap map[string]int) int {
	if d, ok := depthMap[locName]; ok {
		return d
	}
	for _, loc := range locations {
		if loc.Name == locName && loc.Parent != "" {
			parentDepth := calcLocationDepth(loc.Parent, locations, depthMap)
			depthMap[locName] = parentDepth + 1
			return parentDepth + 1
		}
	}
	return 0
}

// isHierarchicalRelation 判断是否为层级/上级关系
func isHierarchicalRelation(relType string) bool {
	hierarchicalTypes := []string{
		"师父", "师傅", "师尊", "师父", "师祖",
		"父亲", "母亲", "爷爷", "奶奶",
		"上级", "上司", "老板", "主上", "主子",
		"老师", "导师", "前辈",
		"首领", "领导", "老大",
		"养父", "养母", "义父", "义母",
	}
	for _, t := range hierarchicalTypes {
		if relType == t {
			return true
		}
	}
	return false
}

// GetCharacterContext 获取人物相关的上下文（用于AI写作）
func (s *GraphService) GetCharacterContext(bookName string, characterName string, hops int) (string, error) {
	// 简化实现：直接获取人物信息
	characters, _ := s.store.LoadCharacters(bookName)
	items, _ := s.store.LoadItems(bookName)
	locations, _ := s.store.LoadLocations(bookName)

	var context string
	for _, char := range characters {
		if char.Name == characterName {
			context += "【人物信息】\n"
			context += "姓名: " + char.Name + "\n"
			context += "角色: " + char.Role + "\n"
			context += "状态: " + char.Status + "\n"
			context += "简介: " + char.Bio + "\n"
			if char.Faction != "" {
				context += "势力: " + char.Faction + "\n"
			}
			if char.Sect != "" {
				context += "宗门: " + char.Sect + "\n"
			}
			if char.Cultivation != "" {
				context += "境界: " + char.Cultivation + "\n"
			}

			if len(char.Relations) > 0 {
				context += "\n【人物关系】\n"
				for _, rel := range char.Relations {
					targetName := rel.TargetName
					if targetName == "" {
						targetName = rel.TargetID
					}
					context += "- " + rel.Type + ": " + targetName + "\n"
				}
			}
			break
		}
	}

	// 查找持有的物品
	for _, item := range items {
		if item.Owner == characterName {
			context += "\n【持有物品】\n"
			context += "- " + item.Name + " (" + item.Type + "): " + item.Description + "\n"
		}
	}

	// 查找关联地点
	for _, loc := range locations {
		if loc.Faction != "" {
			// 检查人物是否属于同一势力
			for _, char := range characters {
				if char.Name == characterName && char.Faction == loc.Faction {
					context += "\n【势力领地】\n"
					context += "- " + loc.Name + ": " + loc.Description + "\n"
				}
			}
		}
	}

	return context, nil
}

// truncateStr 截断字符串（按UTF-8字符截断，避免乱码）
func truncateStr(s string, maxLen int) string {
	// 将字符串转换为 rune 切片，按字符数截断
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// FindPath 查找两个节点之间的关系路径
func (s *GraphService) FindPath(bookName string, start string, end string) ([]string, error) {
	// 简化实现：通过关系查找路径
	graph, err := s.BuildGraph(bookName)
	if err != nil {
		return nil, err
	}

	// BFS 查找路径
	visited := make(map[string]bool)
	queue := [][]string{{start}}
	visited[start] = true

	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]
		current := path[len(path)-1]

		if current == end {
			return path, nil
		}

		// 查找相邻节点
		for _, link := range graph.Links {
			var next string
			if link.Source == current && !visited[link.Target] {
				next = link.Target
			} else if link.Target == current && !visited[link.Source] {
				next = link.Source
			}
			if next != "" {
				visited[next] = true
				newPath := append([]string{}, path...)
				newPath = append(newPath, next)
				queue = append(queue, newPath)
			}
		}
	}

	return nil, nil
}

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

	addNode := func(name, category string, symbolSize int, value string) {
		if nodeNames[name] {
			return
		}
		nodeNames[name] = true
		node := GraphNode{
			Name:       name,
			Category:   category,
			SymbolSize: symbolSize,
			Value:      truncateStr(value, 100),
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
	events, err := s.store.LoadCausalChains(bookName)
	if err != nil {
		// 返回空图谱
		return data, nil
	}

	// 构建事件节点和因果链接
	for _, event := range events {
		// 事件节点
		eventLabel := fmt.Sprintf("第%d章: %s", event.ChapterID, truncateStr(event.Event, 40))
		addNode(eventLabel, "event", 35, event.Event)

		// 添加章节节点（用于定位）
		chapterLabel := fmt.Sprintf("第%d章", event.ChapterID)
		addNode(chapterLabel, "chapter", 25, "")

		// 事件 → 章节
		addLink(eventLabel, chapterLabel, "归属")
	}

	// TODO: 构建因果链接（事件之间）- 需要store支持加载Links
	// 当前store只有LoadCausalChains返回events，links存储待实现

	return data, nil
}

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
			Value:      truncateStr(value, 50),
		}
		if category == "thread" {
			node.ItemStyle.Color = threadTypeColorMap[threadType]
		} else if category == "chapter" {
			node.ItemStyle.Color = "#91cc75" // 绿色
		} else if category == "character" {
			node.ItemStyle.Color = "#5470c6" // 蓝色
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

	// 加载线程数据
	threads, err := s.store.LoadThreads(bookName)
	if err != nil {
		return data, nil
	}

	// 加载角色数据用于 POV 连接
	characters, _ := s.store.LoadCharacters(bookName)
	characterMap := make(map[string]bool)
	for _, char := range characters {
		characterMap[char.Name] = true
	}

	for _, thread := range threads {
		// 线程节点
		threadLabel := truncateStr(thread.Name, 25)
		size := threadTypeSizeMap[string(thread.Type)]
		if size == 0 {
			size = 30
		}
		addNode(threadLabel, "thread", size, thread.Goal, string(thread.Type))

		// 连接涉及的章节
		for _, chapterID := range thread.Chapters {
			chapterLabel := fmt.Sprintf("第%d章", chapterID)
			addNode(chapterLabel, "chapter", 25, "", "")
			addLink(threadLabel, chapterLabel, "涉及", true, "#91cc75")
		}

		// 起止章节（实线连接）
		if thread.StartChapter > 0 {
			startChLabel := fmt.Sprintf("第%d章", thread.StartChapter)
			addNode(startChLabel, "chapter", 25, "", "")
			addLink(threadLabel, startChLabel, "起点", false, "#5470c6")
		}
		if thread.EndChapter > 0 {
			endChLabel := fmt.Sprintf("第%d章", thread.EndChapter)
			addNode(endChLabel, "chapter", 25, "", "")
			addLink(threadLabel, endChLabel, "终点", false, "#ee6666")
		}

		// 连接 POV 角色
		for _, povChar := range thread.POVCharacters {
			if characterMap[povChar] {
				addNode(povChar, "character", 30, "", "")
				addLink(threadLabel, povChar, "POV", false, "#9a60b4")
			}
		}
	}

	return data, nil
}

// BuildEmotionGraph 构建情感图谱（折线图）
// 返回 ECharts 期望的格式：series 和 xAxis
func (s *GraphService) BuildEmotionGraph(bookName string) (*EmotionGraphData, error) {
	data := &EmotionGraphData{
		Series: []EmotionSeries{},
		XAxis:  []string{},
	}

	// 加载角色数据
	characters, err := s.store.LoadCharacters(bookName)
	if err != nil {
		return data, nil
	}

	// 找出所有章节范围
	chapterSet := make(map[int]bool)
	characterEmotions := make(map[string]map[int]int) // character -> chapter -> intensity

	// 遍历角色，提取有情感弧线数据的角色
	for _, char := range characters {
		if len(char.EmotionalArc) == 0 {
			continue
		}

		// 构建角色的情感数据
		characterEmotions[char.Name] = make(map[int]int)
		for _, ep := range char.EmotionalArc {
			characterEmotions[char.Name][ep.ChapterID] = ep.Intensity
			chapterSet[ep.ChapterID] = true
		}
	}

	// 如果没有数据，返回空
	if len(chapterSet) == 0 {
		return data, nil
	}

	// 构建 xAxis（按章节顺序排列）
	chapters := make([]int, 0, len(chapterSet))
	for ch := range chapterSet {
		chapters = append(chapters, ch)
	}
	sort.Ints(chapters)
	for _, ch := range chapters {
		data.XAxis = append(data.XAxis, fmt.Sprintf("第%d章", ch))
	}

	// 构建 series
	for charName, emotions := range characterEmotions {
		seriesData := []int{}
		for _, ch := range chapters {
			if intensity, ok := emotions[ch]; ok {
				seriesData = append(seriesData, intensity)
			} else {
				seriesData = append(seriesData, 0) // 没有数据的章节填0
			}
		}
		data.Series = append(data.Series, EmotionSeries{
			Name: charName,
			Data: seriesData,
		})
	}

	return data, nil
}

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
			Value:      truncateStr(value, 50),
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

	// 按章节分组事件
	chapterEvents := make(map[int][]model.TimelineEvent)
	for _, event := range timeline {
		chapterEvents[event.ChapterID] = append(chapterEvents[event.ChapterID], event)
	}

	// 找出最大章节ID用于遍历
	maxChapterID := 0
	for chID := range chapterEvents {
		if chID > maxChapterID {
			maxChapterID = chID
		}
	}

	// 构建章节节点和时间顺序链接
	var prevChapterLabel string
	for chID := 1; chID <= maxChapterID; chID++ {
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
