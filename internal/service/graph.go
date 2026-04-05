package service

import (
	"ai-writer/internal/model"
	"ai-writer/internal/store"
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
			Value:      truncateStr(value, 30),
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
		}

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
		if faction.Leader != "" {
			addNode(faction.Leader, "character", 35, "")
			addOwnershipLink(faction.Leader, faction.Name, "统领")
		}

		// 势力成员关系：成员 → 势力
		for _, member := range faction.Members {
			addNode(member, "character", 35, "")
			addOwnershipLink(member, faction.Name, "成员")
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

// truncateStr 截断字符串
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
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