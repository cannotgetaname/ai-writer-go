package service

import (
	"context"
	"fmt"
	"strings"

	"ai-writer/internal/llm"
)

// ToolboxService 智能工具箱服务
type ToolboxService struct {
	llmClient llm.Client
}

// NewToolboxService 创建工具箱服务
func NewToolboxService(llmClient llm.Client) *ToolboxService {
	return &ToolboxService{
		llmClient: llmClient,
	}
}

// NamingRequest 命名请求
type NamingRequest struct {
	Type   string // 人名/功法/法宝/宗门/地点
	Genre  string // 玄幻/都市/仙侠/科幻
	Count  int    // 生成数量
	Gender string // 性别（人名用）
	Style  string // 风格
}

// NamingResult 命名结果
type NamingResult struct {
	Names []NameItem `json:"names"`
}

// NameItem 名称项
type NameItem struct {
	Name        string `json:"name"`
	Meaning     string `json:"meaning"`
	Pronunciation string `json:"pronunciation,omitempty"`
}

// GenerateNames 生成名称
func (s *ToolboxService) GenerateNames(ctx context.Context, req *NamingRequest) (*NamingResult, error) {
	if req.Count == 0 {
		req.Count = 5
	}

	prompt := fmt.Sprintf(`请生成%d个%s名称，题材是%s。

要求：
1. 名字要有特色，符合题材风格
2. 每个名字附带简短寓意说明

请用JSON格式输出：
{
  "names": [
    {"name": "名字", "meaning": "寓意"}
  ]
}`, req.Count, req.Type, req.Genre)

	if req.Type == "人名" && req.Gender != "" {
		prompt = fmt.Sprintf(`请生成%d个%s%s名字，题材是%s。

要求：
1. 名字要有特色，符合题材风格
2. 每个名字附带简短寓意说明

请用JSON格式输出：
{
  "names": [
    {"name": "名字", "meaning": "寓意"}
  ]
}`, req.Count, req.Gender, req.Type, req.Genre)
	}

	result, err := s.llmClient.Call(ctx, prompt, "writer")
	if err != nil {
		return nil, err
	}

	// 解析结果
	names := parseNames(result)
	return &NamingResult{Names: names}, nil
}

// CharacterRequest 角色生成请求
type CharacterRequest struct {
	Type   string // 主角/配角/反派
	Gender string // 男/女
	Genre  string // 题材
	Theme  string // 主题/特点
}

// CharacterResult 角色生成结果
type CharacterResult struct {
	Name        string            `json:"name"`
	Gender      string            `json:"gender"`
	Role        string            `json:"role"`
	Bio         string            `json:"bio"`
	Personality string            `json:"personality"`
	Goal        string            `json:"goal"`
	Background  string            `json:"background"`
	Abilities   []string          `json:"abilities"`
	Traits      map[string]string `json:"traits"`
}

// GenerateCharacter 生成角色
func (s *ToolboxService) GenerateCharacter(ctx context.Context, req *CharacterRequest) (*CharacterResult, error) {
	prompt := fmt.Sprintf(`请创建一个%s角色，性别%s，题材是%s。

主题/特点: %s

请生成完整的角色设定，包括：
1. 姓名
2. 性格特点
3. 外貌特征
4. 背景故事
5. 核心目标
6. 特殊能力
7. 优点和缺点

请用JSON格式输出。`, req.Type, req.Gender, req.Genre, req.Theme)

	result, err := s.llmClient.Call(ctx, prompt, "writer")
	if err != nil {
		return nil, err
	}

	char := &CharacterResult{
		Gender: req.Gender,
		Role:   req.Type,
	}

	// 简单解析
	char.Name = extractJSONValue(result, "name")
	char.Bio = extractJSONValue(result, "bio")
	char.Personality = extractJSONValue(result, "personality")
	char.Goal = extractJSONValue(result, "goal")
	char.Background = extractJSONValue(result, "background")

	return char, nil
}

// ConflictRequest 冲突生成请求
type ConflictRequest struct {
	Type    string // 人物/利益/情感/理念
	Genre   string // 题材
	Context string // 背景上下文
}

// ConflictResult 冲突生成结果
type ConflictResult struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Parties     []string `json:"parties"`
	Stakes      string   `json:"stakes"`
	Resolution  string   `json:"possible_resolution"`
}

// GenerateConflict 生成冲突
func (s *ToolboxService) GenerateConflict(ctx context.Context, req *ConflictRequest) (*ConflictResult, error) {
	prompt := fmt.Sprintf(`请设计一个%s类型的冲突，题材是%s。

背景: %s

请生成：
1. 冲突名称
2. 冲突描述
3. 涉及方
4. 利害关系
5. 可能的解决方式

请用JSON格式输出。`, req.Type, req.Genre, req.Context)

	result, err := s.llmClient.Call(ctx, prompt, "architect")
	if err != nil {
		return nil, err
	}

	conflict := &ConflictResult{
		Title:       extractJSONValue(result, "title"),
		Description: extractJSONValue(result, "description"),
		Stakes:      extractJSONValue(result, "stakes"),
		Resolution:  extractJSONValue(result, "resolution"),
	}

	return conflict, nil
}

// SceneRequest 场景生成请求
type SceneRequest struct {
	Type        string // 战斗/日常/对话/冒险
	Location    string // 地点
	Characters  string // 涉及角色
	Mood        string // 氛围
	Description string // 简要描述
}

// SceneResult 场景生成结果
type SceneResult struct {
	Title       string `json:"title"`
	Setting     string `json:"setting"`
	Atmosphere  string `json:"atmosphere"`
	Events      string `json:"events"`
	Description string `json:"full_description"`
}

// GenerateScene 生成场景
func (s *ToolboxService) GenerateScene(ctx context.Context, req *SceneRequest) (*SceneResult, error) {
	prompt := fmt.Sprintf(`请设计一个%s场景。

地点: %s
角色: %s
氛围: %s
简要描述: %s

请生成详细的场景描写，包括：
1. 环境设定
2. 氛围渲染
3. 主要事件
4. 完整描写（200-300字）

请用JSON格式输出。`, req.Type, req.Location, req.Characters, req.Mood, req.Description)

	result, err := s.llmClient.Call(ctx, prompt, "writer")
	if err != nil {
		return nil, err
	}

	scene := &SceneResult{
		Title:       extractJSONValue(result, "title"),
		Setting:     extractJSONValue(result, "setting"),
		Atmosphere:  extractJSONValue(result, "atmosphere"),
		Events:      extractJSONValue(result, "events"),
		Description: extractJSONValue(result, "full_description"),
	}

	return scene, nil
}

// GoldfingerRequest 金手指生成请求
type GoldfingerRequest struct {
	Type   string // 系统/天赋/宝物/传承
	Genre  string // 题材
	Theme  string // 主题
	Level  string // 强度等级
}

// GoldfingerResult 金手指生成结果
type GoldfingerResult struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Abilities   []string `json:"abilities"`
	Limitations string   `json:"limitations"`
	Origin      string   `json:"origin"`
}

// GenerateGoldfinger 生成金手指
func (s *ToolboxService) GenerateGoldfinger(ctx context.Context, req *GoldfingerRequest) (*GoldfingerResult, error) {
	prompt := fmt.Sprintf(`请设计一个%s类型的金手指，题材是%s。

主题: %s
强度: %s

请生成：
1. 名称
2. 类型说明
3. 详细描述
4. 主要能力（3-5个）
5. 限制/代价
6. 来源

请用JSON格式输出。`, req.Type, req.Genre, req.Theme, req.Level)

	result, err := s.llmClient.Call(ctx, prompt, "architect")
	if err != nil {
		return nil, err
	}

	goldfinger := &GoldfingerResult{
		Type:        req.Type,
		Name:        extractJSONValue(result, "name"),
		Description: extractJSONValue(result, "description"),
		Limitations: extractJSONValue(result, "limitations"),
		Origin:      extractJSONValue(result, "origin"),
	}

	return goldfinger, nil
}

// TitleRequest 书名生成请求
type TitleRequest struct {
	Genre   string `json:"genre"`   // 题材
	Theme   string `json:"theme"`   // 主题
	Count   int    `json:"count"`   // 数量
	Style   string `json:"style"`   // 风格 (霸气/文艺/悬疑等)
}

// TitleResult 书名生成结果
type TitleResult struct {
	Titles []TitleItem `json:"titles"`
}

// TitleItem 书名项
type TitleItem struct {
	Title       string `json:"title"`
	Meaning     string `json:"meaning"`
	Attraction  string `json:"attraction"` // 吸引力分析
}

// GenerateTitle 生成书名
func (s *ToolboxService) GenerateTitle(ctx context.Context, req *TitleRequest) (*TitleResult, error) {
	if req.Count == 0 {
		req.Count = 5
	}

	prompt := fmt.Sprintf(`请为一部%s题材的网络小说生成%d个吸引眼球的书名。

主题: %s
风格: %s

要求：
1. 书名要有吸引力，能让读者产生阅读欲望
2. 每个书名附带简短的寓意说明和吸引力分析

请用JSON格式输出：
{
  "titles": [
    {"title": "书名", "meaning": "寓意", "attraction": "吸引力分析"}
  ]
}`, req.Genre, req.Count, req.Theme, req.Style)

	result, err := s.llmClient.Call(ctx, prompt, "writer")
	if err != nil {
		return nil, err
	}

	titles := parseTitles(result)
	return &TitleResult{Titles: titles}, nil
}

// SynopsisRequest 简介生成请求
type SynopsisRequest struct {
	Genre      string `json:"genre"`      // 题材
	MainChar   string `json:"main_char"`  // 主角设定
	WorldView  string `json:"world_view"` // 世界观
	Type       string `json:"type"`       // 类型 (short/long)
}

// SynopsisResult 简介生成结果
type SynopsisResult struct {
	Synopsis    string `json:"synopsis"`
	Highlights  []string `json:"highlights"` // 卖点
	Hook        string `json:"hook"`        // 开篇钩子
}

// GenerateSynopsis 生成简介
func (s *ToolboxService) GenerateSynopsis(ctx context.Context, req *SynopsisRequest) (*SynopsisResult, error) {
	length := "200字以内"
	if req.Type == "long" {
		length = "500字左右"
	}

	prompt := fmt.Sprintf(`请为一部%s题材的网络小说生成%s的简介。

主角设定: %s
世界观: %s

要求：
1. 简介要有吸引力，能吸引读者点击阅读
2. 突出故事的核心冲突和主角特色
3. 提取3-5个卖点
4. 设计一个吸引人的开篇钩子

请用JSON格式输出：
{
  "synopsis": "简介内容",
  "highlights": ["卖点1", "卖点2"],
  "hook": "开篇钩子"
}`, req.Genre, length, req.MainChar, req.WorldView)

	result, err := s.llmClient.Call(ctx, prompt, "writer")
	if err != nil {
		return nil, err
	}

	synopsis := &SynopsisResult{
		Synopsis:   extractJSONValue(result, "synopsis"),
		Hook:       extractJSONValue(result, "hook"),
		Highlights: parseHighlights(extractJSONArray(result, "highlights")),
	}

	return synopsis, nil
}

// TwistRequest 剧情转折生成请求
type TwistRequest struct {
	Type        string `json:"type"`        // 类型 (unexpected/reversal)
	Genre       string `json:"genre"`       // 题材
	Context     string `json:"context"`     // 当前剧情上下文
	Characters  string `json:"characters"`  // 涉及角色
}

// TwistResult 剧情转折生成结果
type TwistResult struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Impact      string   `json:"impact"`      // 影响分析
	Setup       string   `json:"setup"`       // 铺垫建议
	Clues       []string `json:"clues"`       // 伏笔线索
}

// GenerateTwist 生成剧情转折
func (s *ToolboxService) GenerateTwist(ctx context.Context, req *TwistRequest) (*TwistResult, error) {
	twistType := "意外转折"
	if req.Type == "reversal" {
		twistType = "剧情反转"
	}

	prompt := fmt.Sprintf(`请设计一个%s，题材是%s。

当前剧情: %s
涉及角色: %s

请生成：
1. 转折标题
2. 转折描述
3. 影响分析（对剧情和人物的影响）
4. 铺垫建议（如何提前埋下伏笔）
5. 相关线索（读者可以回溯发现的暗示）

请用JSON格式输出：
{
  "title": "转折标题",
  "description": "转折描述",
  "impact": "影响分析",
  "setup": "铺垫建议",
  "clues": ["线索1", "线索2"]
}`, twistType, req.Genre, req.Context, req.Characters)

	result, err := s.llmClient.Call(ctx, prompt, "architect")
	if err != nil {
		return nil, err
	}

	twist := &TwistResult{
		Title:       extractJSONValue(result, "title"),
		Description: extractJSONValue(result, "description"),
		Impact:      extractJSONValue(result, "impact"),
		Setup:       extractJSONValue(result, "setup"),
		Clues:       parseClues(extractJSONArray(result, "clues")),
	}

	return twist, nil
}

// DialogueRequest 对话生成请求
type DialogueRequest struct {
	Characters  string `json:"characters"`  // 角色列表
	Situation   string `json:"situation"`   // 场景情境
	Mood        string `json:"mood"`        // 氛围 (紧张/温馨/搞笑等)
	Genre       string `json:"genre"`       // 题材
}

// DialogueResult 对话生成结果
type DialogueResult struct {
	Content     string `json:"content"`     // 对话内容
	Annotations string `json:"annotations"` // 注释说明
}

// GenerateDialogue 生成对话
func (s *ToolboxService) GenerateDialogue(ctx context.Context, req *DialogueRequest) (*DialogueResult, error) {
	prompt := fmt.Sprintf(`请生成一段对话。

角色: %s
场景情境: %s
氛围: %s
题材: %s

要求：
1. 对话要符合角色性格
2. 展现角色关系和冲突
3. 推动剧情发展
4. 自然流畅，避免说教感

请输出对话内容，包含角色名和动作描写。`, req.Characters, req.Situation, req.Mood, req.Genre)

	result, err := s.llmClient.Call(ctx, prompt, "writer")
	if err != nil {
		return nil, err
	}

	return &DialogueResult{
		Content:     result,
		Annotations: "",
	}, nil
}

// parseTitles 解析书名列表
func parseTitles(result string) []TitleItem {
	var titles []TitleItem
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.Contains(line, `"title"`) {
			title := extractJSONValue(line, "title")
			meaning := extractJSONValue(line, "meaning")
			attraction := extractJSONValue(line, "attraction")
			if title != "" {
				titles = append(titles, TitleItem{
					Title:      title,
					Meaning:    meaning,
					Attraction: attraction,
				})
			}
		}
	}

	if len(titles) == 0 {
		titles = append(titles, TitleItem{
			Title:      "生成结果",
			Meaning:    result,
			Attraction: "",
		})
	}

	return titles
}

func parseHighlights(s string) []string {
	if s == "" {
		return []string{}
	}
	// 简单分割
	items := strings.Split(s, ",")
	var result []string
	for _, item := range items {
		item = strings.TrimSpace(item)
		item = strings.Trim(item, `"[]`)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func parseClues(s string) []string {
	return parseHighlights(s)
}

func extractJSONArray(jsonStr, key string) string {
	searchKey := `"` + key + `":`
	startIdx := strings.Index(jsonStr, searchKey)
	if startIdx == -1 {
		return ""
	}

	startIdx += len(searchKey)
	for startIdx < len(jsonStr) && (jsonStr[startIdx] == ' ' || jsonStr[startIdx] == '\n') {
		startIdx++
	}

	if startIdx >= len(jsonStr) || jsonStr[startIdx] != '[' {
		return ""
	}

	// 找到对应的 ]
	depth := 1
	endIdx := startIdx + 1
	for endIdx < len(jsonStr) && depth > 0 {
		if jsonStr[endIdx] == '[' {
			depth++
		} else if jsonStr[endIdx] == ']' {
			depth--
		}
		endIdx++
	}

	return jsonStr[startIdx:endIdx]
}

// parseNames 解析名称列表
func parseNames(result string) []NameItem {
	var names []NameItem

	// 简单解析
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.Contains(line, `"name"`) {
			name := extractJSONValue(line, "name")
			meaning := extractJSONValue(line, "meaning")
			if name != "" {
				names = append(names, NameItem{
					Name:    name,
					Meaning: meaning,
				})
			}
		}
	}

	if len(names) == 0 {
		// 尝试其他解析方式
		names = append(names, NameItem{
			Name:    "生成结果",
			Meaning: result,
		})
	}

	return names
}

// extractJSONValue 从 JSON 字符串中提取值
func extractJSONValue(jsonStr, key string) string {
	searchKey := `"` + key + `":`
	startIdx := strings.Index(jsonStr, searchKey)
	if startIdx == -1 {
		return ""
	}

	startIdx += len(searchKey)
	for startIdx < len(jsonStr) && (jsonStr[startIdx] == ' ' || jsonStr[startIdx] == '\n' || jsonStr[startIdx] == '\t') {
		startIdx++
	}

	if startIdx >= len(jsonStr) {
		return ""
	}

	if jsonStr[startIdx] == '"' {
		startIdx++
		endIdx := strings.Index(jsonStr[startIdx:], `"`)
		if endIdx == -1 {
			return ""
		}
		return jsonStr[startIdx : startIdx+endIdx]
	}

	return ""
}