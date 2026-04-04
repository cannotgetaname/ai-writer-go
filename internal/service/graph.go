package service

import (
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
	Source string `json:"source"`
	Target string `json:"target"`
	Value  string `json:"value"`
	Label  struct {
		Show     bool   `json:"show"`
		Formatter string `json:"formatter"`
	} `json:"label"`
	LineStyle struct {
		Curveness float64 `json:"curveness"`
		Color     string  `json:"color"`
	} `json:"lineStyle"`
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
			{Name: "location"},
			{Name: "item"},
		},
	}

	// 颜色配置
	colorMap := map[string]string{
		"character": "#5470c6", // 蓝色
		"location":  "#91cc75", // 绿色
		"item":      "#fac858", // 黄色
	}

	// 加载人物
	characters, _ := s.store.LoadCharacters(bookName)
	for _, char := range characters {
		node := GraphNode{
			Name:       char.Name,
			Category:   "character",
			SymbolSize: 40,
			Value:      truncateStr(char.Bio, 20),
		}
		node.ItemStyle.Color = colorMap["character"]
		node.Label.Show = true
		node.Label.Position = "right"
		data.Nodes = append(data.Nodes, node)

		// 添加人物关系
		for _, rel := range char.Relations {
			targetName := rel.TargetName
			if targetName == "" {
				targetName = rel.TargetID
			}
			if targetName != "" {
				link := GraphLink{
					Source: char.Name,
					Target: targetName,
					Value:  rel.Type,
				}
				link.Label.Show = true
				link.LineStyle.Curveness = 0.2
				link.LineStyle.Color = "source"
				data.Links = append(data.Links, link)
			}
		}
	}

	// 加载物品
	items, _ := s.store.LoadItems(bookName)
	for _, item := range items {
		node := GraphNode{
			Name:       item.Name,
			Category:   "item",
			SymbolSize: 25,
			Value:      truncateStr(item.Description, 20),
		}
		node.ItemStyle.Color = colorMap["item"]
		node.Label.Show = true
		node.Label.Position = "right"
		data.Nodes = append(data.Nodes, node)

		// 添加持有关系
		if item.Owner != "" {
			link := GraphLink{
				Source: item.Owner,
				Target: item.Name,
				Value:  "持有",
			}
			link.Label.Show = true
			link.LineStyle.Curveness = 0.2
			link.LineStyle.Color = "source"
			data.Links = append(data.Links, link)
		}
	}

	// 加载地点
	locations, _ := s.store.LoadLocations(bookName)
	for _, loc := range locations {
		node := GraphNode{
			Name:       loc.Name,
			Category:   "location",
			SymbolSize: 35,
			Value:      truncateStr(loc.Description, 20),
		}
		node.ItemStyle.Color = colorMap["location"]
		node.Label.Show = true
		node.Label.Position = "right"
		data.Nodes = append(data.Nodes, node)

		// 添加相邻关系
		for _, neighbor := range loc.Neighbors {
			link := GraphLink{
				Source: loc.Name,
				Target: neighbor,
				Value:  "连通",
			}
			link.Label.Show = true
			link.LineStyle.Curveness = 0.2
			link.LineStyle.Color = "source"
			data.Links = append(data.Links, link)
		}

		// 添加父子关系
		if loc.Parent != "" {
			link := GraphLink{
				Source: loc.Name,
				Target: loc.Parent,
				Value:  "属于",
			}
			link.Label.Show = true
			link.LineStyle.Curveness = 0.2
			link.LineStyle.Color = "source"
			data.Links = append(data.Links, link)
		}
	}

	return data, nil
}

// GetCharacterContext 获取人物相关的上下文（用于AI写作）
func (s *GraphService) GetCharacterContext(bookName string, characterName string, hops int) (string, error) {
	// 简化实现：直接获取人物信息
	characters, _ := s.store.LoadCharacters(bookName)
	items, _ := s.store.LoadItems(bookName)

	var context string
	for _, char := range characters {
		if char.Name == characterName {
			context += "【人物信息】\n"
			context += "姓名: " + char.Name + "\n"
			context += "角色: " + char.Role + "\n"
			context += "状态: " + char.Status + "\n"
			context += "简介: " + char.Bio + "\n"

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

	return context, nil
}

// truncateStr 截断字符串
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ensureNodeExists 确保节点存在
func ensureNodeExists(nodes []GraphNode, name string) bool {
	for _, n := range nodes {
		if n.Name == name {
			return true
		}
	}
	return false
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