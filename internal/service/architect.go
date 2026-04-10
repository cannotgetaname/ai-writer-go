package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"ai-writer/internal/llm"
	"ai-writer/internal/model"
	"ai-writer/internal/store"
)

// ArchitectService 架构师服务
type ArchitectService struct {
	llmClient llm.Client
	store     *store.JSONStore
}

// NewArchitectService 创建架构师服务
func NewArchitectService(llmClient llm.Client, store *store.JSONStore) *ArchitectService {
	return &ArchitectService{
		llmClient: llmClient,
		store:     store,
	}
}

// ==================== 分形写作流程 ====================

// FractalLevel 分形层级
type FractalLevel string

const (
	LevelSynopsis FractalLevel = "synopsis"   // 全书总纲
	LevelWorld    FractalLevel = "world"      // 世界观
	LevelVolume   FractalLevel = "volume"     // 分卷大纲
	LevelChapter  FractalLevel = "chapter"    // 章节大纲
	LevelDetail   FractalLevel = "detail"     // 章节细纲
)

// SynopsisResult 全书总纲结果
type SynopsisResult struct {
	Title       string   `json:"title"`        // 书名
	Genre       string   `json:"genre"`        // 题材
	Theme       string   `json:"theme"`        // 主题
	WordCount   int      `json:"word_count"`   // 预估字数
	Synopsis    string   `json:"synopsis"`     // 故事梗概（500字以内）
	MainPlot    string   `json:"main_plot"`    // 主线剧情
	SubPlots    []string `json:"sub_plots"`    // 支线剧情
	MainChars   []string `json:"main_chars"`   // 主要人物
	EndingType  string   `json:"ending_type"`  // 结局类型
	VolumeCount int      `json:"volume_count"` // 分卷数量
}

// WorldViewGenerateRequest 世界观生成请求
type WorldViewGenerateRequest struct {
	BookName string `json:"book_name"`
	Genre    string `json:"genre"`
	Synopsis string `json:"synopsis"` // 故事梗概
	Theme    string `json:"theme"`
}

// WorldViewGenerateResult 世界观生成结果
type WorldViewGenerateResult struct {
	Genre          string `json:"genre"`
	Era            string `json:"era"`
	TechLevel      string `json:"tech_level"`
	PowerSystem    string `json:"power_system"`
	SocialStructure string `json:"social_structure"`
	SpecialRules   string `json:"special_rules"`
	ImportantItems string `json:"important_items"`
	Organizations  string `json:"organizations"`
	Locations      string `json:"locations"`
	History        string `json:"history"`
	MainConflict   string `json:"main_conflict"`
	Development    string `json:"development"`
}

// VolumeOutline 分卷大纲
type VolumeOutline struct {
	ID          string          `json:"id"`
	Index       int             `json:"index"`        // 第几卷
	Title       string          `json:"title"`        // 卷名
	Synopsis    string          `json:"synopsis"`     // 本卷梗概
	MainEvent   string          `json:"main_event"`   // 核心事件
	EmotionArc  string          `json:"emotion_arc"`  // 情感弧线
	ChapterCount int            `json:"chapter_count"` // 章节数量
	Chapters    []ChapterOutline `json:"chapters"`    // 章节列表
}

// ChapterOutline 章节大纲
type ChapterOutline struct {
	ID          string `json:"id"`
	Index       int    `json:"index"`        // 第几章
	VolumeIndex int    `json:"volume_index"` // 所属卷
	Title       string `json:"title"`        // 章节名
	Synopsis    string `json:"synopsis"`     // 章节梗概
	MainEvent   string `json:"main_event"`   // 核心事件
	Characters  string `json:"characters"`   // 出场人物
	Location    string `json:"location"`     // 场景地点
	Foreshadow  string `json:"foreshadow"`   // 伏笔设置
}

// ChapterDetail 章节细纲
type ChapterDetail struct {
	ChapterID    string        `json:"chapter_id"`
	Scenes       []SceneDetail `json:"scenes"`       // 场景列表
	Dialogues    []string      `json:"dialogues"`    // 关键对话
	Actions      []string      `json:"actions"`      // 动作设计
	Emotions     []string      `json:"emotions"`     // 情感变化
	Foreshadows  []string      `json:"foreshadows"`  // 伏笔
	WordTarget   int           `json:"word_target"`  // 目标字数
}

// SceneDetail 场景细纲
type SceneDetail struct {
	Index      int    `json:"index"`
	Location   string `json:"location"`
	Characters string `json:"characters"`
	Event      string `json:"event"`
	Mood       string `json:"mood"`
}

// ==================== 分形生成方法 ====================

// GenerateSynopsis 生成全书总纲
func (s *ArchitectService) GenerateSynopsis(ctx context.Context, genre, theme, mainChar string, targetWords int) (*SynopsisResult, error) {
	prompt := fmt.Sprintf(`请为一部%s题材的网络小说设计全书总纲。

主题方向: %s
主角设定: %s
目标字数: %d字

请生成一份完整的创作蓝图，包括：
1. 书名建议
2. 故事梗概（300-500字）
3. 主线剧情概述
4. 主要支线（2-3条）
5. 主要人物列表
6. 结局类型
7. 建议分卷数量

请严格按JSON格式输出：
{
  "title": "书名",
  "genre": "题材",
  "theme": "主题",
  "word_count": 预估字数,
  "synopsis": "故事梗概",
  "main_plot": "主线剧情概述",
  "sub_plots": ["支线1", "支线2"],
  "main_chars": ["主角名(身份)", "配角名(身份)"],
  "ending_type": "结局类型",
  "volume_count": 分卷数
}`, genre, theme, mainChar, targetWords)

	result, err := s.llmClient.Call(ctx, prompt, "architect")
	if err != nil {
		return nil, err
	}

	var synopsis SynopsisResult
	if err := parseJSON(result, &synopsis); err != nil {
		// 尝试提取基本信息
		synopsis = SynopsisResult{
			Title:      extractJSONValue(result, "title"),
			Genre:      genre,
			Theme:      theme,
			WordCount:  targetWords,
			Synopsis:   extractJSONValue(result, "synopsis"),
			MainPlot:   extractJSONValue(result, "main_plot"),
			EndingType: extractJSONValue(result, "ending_type"),
		}
	}

	return &synopsis, nil
}

// GenerateWorldView 生成世界观
func (s *ArchitectService) GenerateWorldView(ctx context.Context, req *WorldViewGenerateRequest) (*WorldViewGenerateResult, error) {
	prompt := fmt.Sprintf(`请为以下小说设计完整的世界观设定。

题材: %s
主题: %s
故事梗概: %s

请设计一个丰富、自洽的世界观，包括：
1. 时代背景和科技水平
2. 力量体系（修炼等级、能力设定等）
3. 社会结构和势力分布
4. 特殊规则（这个世界独有的法则）
5. 重要物品和地点
6. 历史背景和主要矛盾

请严格按JSON格式输出：
{
  "genre": "题材",
  "era": "时代背景",
  "tech_level": "科技水平",
  "power_system": "力量体系详细描述",
  "social_structure": "社会结构和主要势力",
  "special_rules": "特殊规则",
  "important_items": "重要物品",
  "organizations": "主要势力组织",
  "locations": "主要地点",
  "history": "历史背景",
  "main_conflict": "主要矛盾",
  "development": "世界发展趋势"
}`, req.Genre, req.Theme, req.Synopsis)

	result, err := s.llmClient.Call(ctx, prompt, "architect")
	if err != nil {
		return nil, err
	}

	var worldView WorldViewGenerateResult
	if err := parseJSON(result, &worldView); err != nil {
		worldView.Genre = req.Genre
		worldView.PowerSystem = result
	}

	return &worldView, nil
}

// GenerateVolumeOutlines 生成分卷大纲
func (s *ArchitectService) GenerateVolumeOutlines(ctx context.Context, bookName string, synopsis *SynopsisResult, worldView *WorldViewGenerateResult) ([]VolumeOutline, error) {
	// 参数验证
	if synopsis == nil {
		return nil, fmt.Errorf("缺少总纲数据")
	}
	if worldView == nil {
		return nil, fmt.Errorf("缺少世界观数据")
	}

	// 简化prompt，只生成卷级大纲，不包含详细章节
	// 章节大纲可以通过"展开分卷"功能单独生成
	prompt := fmt.Sprintf(`请根据以下信息设计分卷大纲。

【全书总纲】
%s

【世界观要点】
力量体系: %s
社会结构: %s

【要求】
共%d卷，只设计卷级大纲，不需要详细章节。

请严格按JSON格式输出：
{
  "volumes": [
    {
      "id": "vol_1",
      "index": 1,
      "title": "第一卷名称",
      "synopsis": "本卷梗概(100字以内)",
      "main_event": "核心事件(50字以内)",
      "emotion_arc": "情感弧线",
      "chapter_count": 20
    }
  ]
}`, synopsis.Synopsis, worldView.PowerSystem, worldView.SocialStructure, synopsis.VolumeCount)

	result, err := s.llmClient.Call(ctx, prompt, "architect")
	if err != nil {
		log.Printf("LLM Call error: %v", err)
		return nil, err
	}

	log.Printf("LLM result length: %d", len(result))
	if len(result) > 500 {
		log.Printf("LLM result first 500 chars: %s...", result[:500])
	} else {
		log.Printf("LLM result: %s", result)
	}

	var data struct {
		Volumes []VolumeOutline `json:"volumes"`
	}
	if err := parseJSON(result, &data); err != nil {
		log.Printf("parseJSON error: %v, raw result length: %d", err, len(result))
		return nil, fmt.Errorf("JSON解析失败: %w", err)
	}

	log.Printf("Parsed %d volumes", len(data.Volumes))
	return data.Volumes, nil
}

// ExpandChapterDetail 展开章节细纲
func (s *ArchitectService) ExpandChapterDetail(ctx context.Context, bookName string, chapter *ChapterOutline, worldView *WorldViewGenerateResult) (*ChapterDetail, error) {
	// 参数验证
	if chapter == nil {
		return nil, fmt.Errorf("缺少章节数据")
	}
	if worldView == nil {
		return nil, fmt.Errorf("缺少世界观数据")
	}

	prompt := fmt.Sprintf(`请为以下章节设计详细细纲。

【章节信息】
章节: %s
梗概: %s
核心事件: %s
出场人物: %s
场景: %s

【世界观参考】
力量体系: %s

请设计：
1. 场景划分（2-4个场景）
2. 关键对话设计
3. 动作戏设计
4. 情感变化曲线
5. 伏笔设置
6. 目标字数

请严格按JSON格式输出：
{
  "chapter_id": "%s",
  "scenes": [
    {
      "index": 1,
      "location": "场景地点",
      "characters": "出场人物",
      "event": "场景事件",
      "mood": "氛围基调"
    }
  ],
  "dialogues": ["关键对话1", "关键对话2"],
  "actions": ["动作设计1", "动作设计2"],
  "emotions": ["开篇情绪", "中段转折", "结尾情绪"],
  "foreshadows": ["伏笔1", "伏笔2"],
  "word_target": 3000
}`, chapter.Title, chapter.Synopsis, chapter.MainEvent, chapter.Characters, chapter.Location,
		worldView.PowerSystem, chapter.ID)

	result, err := s.llmClient.Call(ctx, prompt, "architect")
	if err != nil {
		return nil, err
	}

	var detail ChapterDetail
	if err := parseJSON(result, &detail); err != nil {
		detail.ChapterID = chapter.ID
		detail.WordTarget = 3000
	}

	return &detail, nil
}

// ExpandVolume 展开单个分卷（生成章节）
func (s *ArchitectService) ExpandVolume(ctx context.Context, volume *VolumeOutline, synopsis *SynopsisResult, worldView *WorldViewGenerateResult) (*VolumeOutline, error) {
	// 参数验证
	if volume == nil {
		return nil, fmt.Errorf("缺少分卷数据")
	}
	if synopsis == nil {
		return nil, fmt.Errorf("缺少总纲数据")
	}

	prompt := fmt.Sprintf(`请为以下分卷设计详细章节大纲。

【分卷信息】
卷名: %s
梗概: %s
核心事件: %s
预计章节数: %d

【全书背景】
%s

请设计每一章的大纲，严格按JSON格式输出：
{
  "chapters": [
    {
      "id": "chap_%d_1",
      "index": 1,
      "volume_index": %d,
      "title": "章节名",
      "synopsis": "章节梗概(50字以内)",
      "main_event": "核心事件",
      "characters": "出场人物",
      "location": "场景",
      "foreshadow": "伏笔"
    }
  ]
}`, volume.Title, volume.Synopsis, volume.MainEvent, volume.ChapterCount,
		synopsis.Synopsis, volume.Index, volume.Index)

	result, err := s.llmClient.Call(ctx, prompt, "architect")
	if err != nil {
		return nil, err
	}

	var data struct {
		Chapters []ChapterOutline `json:"chapters"`
	}
	if err := parseJSON(result, &data); err != nil {
		return nil, err
	}

	volume.Chapters = data.Chapters
	return volume, nil
}

// SaveOutline 保存大纲到书籍
func (s *ArchitectService) SaveOutline(bookName string, volumes []VolumeOutline) error {
	// 保存章节结构
	var chapters []*model.Chapter
	chapterID := 1

	for _, vol := range volumes {
		for _, ch := range vol.Chapters {
			chapters = append(chapters, &model.Chapter{
				ID:        chapterID,
				BookID:    bookName,
				VolumeID:  vol.ID,
				Title:     ch.Title,
				Outline:   ch.Synopsis,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
			chapterID++
		}
	}

	if len(chapters) > 0 {
		if err := s.store.SaveChapters(bookName, chapters); err != nil {
			return err
		}
	}

	return nil
}

// SaveWorldView 保存世界观
func (s *ArchitectService) SaveWorldView(bookName string, wv *WorldViewGenerateResult) error {
	worldView := &model.WorldView{
		BookID: bookName,
		BasicInfo: model.WorldViewBasic{
			Genre:     wv.Genre,
			Era:       wv.Era,
			TechLevel: wv.TechLevel,
		},
		CoreSettings: model.WorldViewCore{
			PowerSystem:     wv.PowerSystem,
			SocialStructure: wv.SocialStructure,
			SpecialRules:    wv.SpecialRules,
		},
		KeyElements: model.WorldViewElements{
			ImportantItems: wv.ImportantItems,
			Organizations:  wv.Organizations,
			Locations:      wv.Locations,
		},
		Background: model.WorldViewBackground{
			History:      wv.History,
			MainConflict: wv.MainConflict,
			Development:  wv.Development,
		},
	}

	return s.store.SaveWorldView(bookName, worldView)
}

// ==================== 兼容旧接口 ====================

// NodeStatus 节点状态
type NodeStatus string

const (
	NodeStatusPlanned NodeStatus = "planned"
	NodeStatusWriting NodeStatus = "writing"
	NodeStatusDone    NodeStatus = "done"
	NodeStatusReview  NodeStatus = "review"
	NodeStatusHold    NodeStatus = "hold"
)

// TreeNode 树节点
type TreeNode struct {
	ID          string       `json:"id"`
	Label       string       `json:"label"`
	Type        string       `json:"type"`
	Status      NodeStatus   `json:"status"`
	Outline     string       `json:"outline"`
	Children    []*TreeNode  `json:"children,omitempty"`
	ParentID    string       `json:"parent_id,omitempty"`
}

// FissionStrategy 分形裂变策略
type FissionStrategy struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Prompt      string `json:"prompt"`
}

// FissionRequest 分形裂变请求
type FissionRequest struct {
	BookName       string `json:"book_name"`
	NodeID         string `json:"node_id"`
	NodeType       string `json:"node_type"`
	CurrentOutline string `json:"current_outline"`
	Strategy       string `json:"strategy"`
	Count          int    `json:"count"`
}

// FissionResult 分形裂变结果
type FissionResult struct {
	Nodes []*TreeNode `json:"nodes"`
}

// GenerateOutlineRequest 生成大纲请求
type GenerateOutlineRequest struct {
	BookName    string `json:"book_name"`
	Genre       string `json:"genre"`
	MainChar    string `json:"main_char"`
	Theme       string `json:"theme"`
	TargetWords int    `json:"target_words"`
	VolumeCount int    `json:"volume_count"`
}

// GenerateOutlineResult 生成大纲结果
type GenerateOutlineResult struct {
	Volumes []*TreeNode `json:"volumes"`
	Synopsis string      `json:"synopsis"`
}

// GenerateOutline 生成大纲（兼容旧接口）
func (s *ArchitectService) GenerateOutline(ctx context.Context, req *GenerateOutlineRequest) (*GenerateOutlineResult, error) {
	// 先生成总纲
	synopsis, err := s.GenerateSynopsis(ctx, req.Genre, req.Theme, req.MainChar, req.TargetWords)
	if err != nil {
		return nil, err
	}

	// 转换为旧格式
	result := &GenerateOutlineResult{
		Synopsis: synopsis.Synopsis,
		Volumes:  make([]*TreeNode, 0),
	}

	for i := 1; i <= synopsis.VolumeCount; i++ {
		volNode := &TreeNode{
			ID:      fmt.Sprintf("vol_%d", i),
			Label:   fmt.Sprintf("第%d卷", i),
			Type:    "volume",
			Status:  NodeStatusPlanned,
			Outline: "",
		}
		result.Volumes = append(result.Volumes, volNode)
	}

	return result, nil
}

// Fission 分形裂变（兼容旧接口）
func (s *ArchitectService) Fission(ctx context.Context, req *FissionRequest) (*FissionResult, error) {
	var prompt string

	switch req.Strategy {
	case "expand":
		prompt = fmt.Sprintf(`请将以下%s大纲展开为更详细的子节点。

当前大纲: %s
需要生成数量: %d

请生成详细的子节点，每个节点包含标题和概述。
用JSON格式输出：
{
  "nodes": [
    {"id": "node_1", "label": "标题", "outline": "概述", "status": "planned"}
  ]
}`, req.NodeType, req.CurrentOutline, req.Count)

	case "refine":
		prompt = fmt.Sprintf(`请优化以下%s大纲，使其更加完整和有吸引力。

当前大纲: %s

请优化并输出：
{
  "nodes": [
    {"id": "node_1", "label": "优化后标题", "outline": "优化后概述", "status": "planned"}
  ]
}`, req.NodeType, req.CurrentOutline)

	case "branch":
		prompt = fmt.Sprintf(`请基于以下%s大纲，生成多个可能的剧情分支。

当前大纲: %s
需要分支数量: %d

每个分支代表不同的剧情发展方向。
用JSON格式输出：
{
  "nodes": [
    {"id": "branch_1", "label": "分支标题", "outline": "分支概述", "status": "planned"}
  ]
}`, req.NodeType, req.CurrentOutline, req.Count)

	default:
		prompt = fmt.Sprintf(`请展开以下大纲: %s`, req.CurrentOutline)
	}

	result, err := s.llmClient.Call(ctx, prompt, "architect")
	if err != nil {
		return nil, err
	}

	var fissionResult FissionResult
	if err := parseJSON(result, &fissionResult); err != nil {
		fissionResult.Nodes = []*TreeNode{
			{
				ID:      fmt.Sprintf("node_%d", time.Now().Unix()),
				Label:   "生成结果",
				Outline: result,
				Status:  NodeStatusPlanned,
			},
		}
	}

	return &fissionResult, nil
}

// GetFissionStrategies 获取分形裂变策略
func GetFissionStrategies() map[string][]FissionStrategy {
	return map[string][]FissionStrategy{
		"expand": {
			{ID: "expand_detail", Name: "详细展开", Description: "将简单大纲展开为详细章节", Prompt: "展开为详细章节"},
			{ID: "expand_plot", Name: "剧情展开", Description: "展开剧情细节和转折", Prompt: "展开剧情细节"},
		},
		"refine": {
			{ID: "refine_logic", Name: "逻辑优化", Description: "优化剧情逻辑", Prompt: "优化逻辑"},
			{ID: "refine_pacing", Name: "节奏优化", Description: "优化叙事节奏", Prompt: "优化节奏"},
		},
		"branch": {
			{ID: "branch_plot", Name: "剧情分支", Description: "生成多条剧情线", Prompt: "生成剧情分支"},
			{ID: "branch_ending", Name: "结局分支", Description: "生成多种可能结局", Prompt: "生成结局分支"},
		},
	}
}

// AnalyzeStructure 分析结构
func (s *ArchitectService) AnalyzeStructure(ctx context.Context, bookName string, tree []*TreeNode) (map[string]interface{}, error) {
	stats := map[string]int{
		"planned": 0,
		"writing": 0,
		"done":    0,
		"review":  0,
		"hold":    0,
	}

	var countNodes func(nodes []*TreeNode)
	countNodes = func(nodes []*TreeNode) {
		for _, node := range nodes {
			stats[string(node.Status)]++
			if node.Children != nil {
				countNodes(node.Children)
			}
		}
	}
	countNodes(tree)

	total := 0
	for _, v := range stats {
		total += v
	}

	return map[string]interface{}{
		"total":   total,
		"stats":   stats,
		"progress": float64(stats["done"]+stats["review"]) / float64(total) * 100,
	}, nil
}

// ==================== 辅助函数 ====================

func parseJSON(s string, v interface{}) error {
	start := indexOf(s, "{")
	end := lastIndexOf(s, "}")
	if start == -1 || end == -1 {
		return fmt.Errorf("no JSON found")
	}
	jsonStr := s[start : end+1]
	return json.Unmarshal([]byte(jsonStr), v)
}

func indexOf(s string, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func lastIndexOf(s string, substr string) int {
	for i := len(s) - len(substr); i >= 0; i-- {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func extractJSONValue(s string, key string) string {
	// 简单提取
	searchKey := fmt.Sprintf(`"%s"`, key)
	idx := indexOf(s, searchKey)
	if idx == -1 {
		return ""
	}

	// 找到冒号后的值
	rest := s[idx+len(searchKey):]
	colonIdx := indexOf(rest, ":")
	if colonIdx == -1 {
		return ""
	}

	rest = rest[colonIdx+1:]
	// 跳过空白
	for len(rest) > 0 && (rest[0] == ' ' || rest[0] == '\n' || rest[0] == '\t') {
		rest = rest[1:]
	}

	// 如果是字符串
	if len(rest) > 0 && rest[0] == '"' {
		endIdx := indexOf(rest[1:], `"`)
		if endIdx != -1 {
			return rest[1 : endIdx+1]
		}
	}

	return ""
}