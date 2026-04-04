package service

import (
	"context"
	"fmt"
	"time"

	"ai-writer/internal/llm"
	"ai-writer/internal/model"
	"ai-writer/internal/store"
)

// InitService 初始化服务
type InitService struct {
	llmClient llm.Client
	store     *store.JSONStore
}

// NewInitService 创建初始化服务
func NewInitService(llmClient llm.Client, jsonStore *store.JSONStore) *InitService {
	return &InitService{
		llmClient: llmClient,
		store:     jsonStore,
	}
}

// InitRequest 初始化请求
type InitRequest struct {
	BookName string `json:"book_name"`
	Idea     string `json:"idea"`
	Genre    string `json:"genre"`
	TargetWords int  `json:"target_words"`
	VolumeCount int  `json:"volume_count"`
}

// InitResult 初始化结果
type InitResult struct {
	BookName    string         `json:"book_name"`
	WorldView   *model.WorldView `json:"worldview"`
	MainCharacter *model.Character `json:"main_character"`
	Volumes     []*VolumeInfo  `json:"volumes"`
	Synopsis    string         `json:"synopsis"`
}

// VolumeInfo 分卷信息
type VolumeInfo struct {
	ID       string        `json:"id"`
	Title    string        `json:"title"`
	Outline  string        `json:"outline"`
	Chapters []*ChapterInfo `json:"chapters"`
}

// ChapterInfo 章节信息
type ChapterInfo struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Outline string `json:"outline"`
}

// Initialize 从 idea 初始化整个项目
func (s *InitService) Initialize(ctx context.Context, req *InitRequest) (*InitResult, error) {
	result := &InitResult{
		BookName: req.BookName,
	}

	// Step 1: 生成世界观
	worldview, err := s.generateWorldView(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("生成世界观失败: %w", err)
	}
	result.WorldView = worldview

	// Step 2: 生成主角
	mainChar, err := s.generateMainCharacter(ctx, req, worldview)
	if err != nil {
		return nil, fmt.Errorf("生成主角失败: %w", err)
	}
	result.MainCharacter = mainChar

	// Step 3: 生成大纲
	volumes, synopsis, err := s.generateOutline(ctx, req, worldview, mainChar)
	if err != nil {
		return nil, fmt.Errorf("生成大纲失败: %w", err)
	}
	result.Volumes = volumes
	result.Synopsis = synopsis

	return result, nil
}

// generateWorldView 生成世界观
func (s *InitService) generateWorldView(ctx context.Context, req *InitRequest) (*model.WorldView, error) {
	prompt := fmt.Sprintf(`请为一部%s题材的网络小说设计世界观。

故事核心创意: %s

请生成详细的世界观设定，用JSON格式输出：
{
  "genre": "题材类型",
  "era": "时代背景",
  "tech_level": "科技/文明水平",
  "power_system": "力量体系描述（如修仙境界、武学等级等）",
  "social_structure": "社会结构（势力分布、阶层等）",
  "special_rules": "世界特殊规则",
  "important_items": "重要物品/宝物",
  "organizations": "主要势力组织",
  "locations": "主要地点",
  "history": "历史背景",
  "main_conflict": "世界主要矛盾",
  "development": "世界发展趋势"
}`, req.Genre, req.Idea)

	result, err := s.llmClient.Call(ctx, prompt, "architect")
	if err != nil {
		return nil, err
	}

	worldview := &model.WorldView{
		BookID: req.BookName,
	}

	// 解析 JSON
	var data map[string]string
	if err := parseJSON(result, &data); err != nil {
		// 尝试逐字段提取
		worldview.BasicInfo.Genre = extractJSONValue(result, "genre")
		worldview.BasicInfo.Era = extractJSONValue(result, "era")
		worldview.BasicInfo.TechLevel = extractJSONValue(result, "tech_level")
		worldview.CoreSettings.PowerSystem = extractJSONValue(result, "power_system")
		worldview.CoreSettings.SocialStructure = extractJSONValue(result, "social_structure")
		worldview.CoreSettings.SpecialRules = extractJSONValue(result, "special_rules")
		worldview.KeyElements.ImportantItems = extractJSONValue(result, "important_items")
		worldview.KeyElements.Organizations = extractJSONValue(result, "organizations")
		worldview.KeyElements.Locations = extractJSONValue(result, "locations")
		worldview.Background.History = extractJSONValue(result, "history")
		worldview.Background.MainConflict = extractJSONValue(result, "main_conflict")
		worldview.Background.Development = extractJSONValue(result, "development")
	} else {
		worldview.BasicInfo.Genre = data["genre"]
		worldview.BasicInfo.Era = data["era"]
		worldview.BasicInfo.TechLevel = data["tech_level"]
		worldview.CoreSettings.PowerSystem = data["power_system"]
		worldview.CoreSettings.SocialStructure = data["social_structure"]
		worldview.CoreSettings.SpecialRules = data["special_rules"]
		worldview.KeyElements.ImportantItems = data["important_items"]
		worldview.KeyElements.Organizations = data["organizations"]
		worldview.KeyElements.Locations = data["locations"]
		worldview.Background.History = data["history"]
		worldview.Background.MainConflict = data["main_conflict"]
		worldview.Background.Development = data["development"]
	}

	return worldview, nil
}

// generateMainCharacter 生成主角
func (s *InitService) generateMainCharacter(ctx context.Context, req *InitRequest, worldview *model.WorldView) (*model.Character, error) {
	prompt := fmt.Sprintf(`请为一部%s题材的网络小说设计主角。

故事核心创意: %s
世界观: %s

请生成主角设定，用JSON格式输出：
{
  "name": "主角姓名",
  "gender": "性别",
  "bio": "人物简介（100字以内）",
  "personality": "性格特点",
  "goal": "外在目标",
  "desire": "内在渴望",
  "background": "背景故事",
  "ability": "特殊能力/金手指"
}`, req.Genre, req.Idea, worldview.CoreSettings.PowerSystem)

	result, err := s.llmClient.Call(ctx, prompt, "architect")
	if err != nil {
		return nil, err
	}

	character := &model.Character{
		ID:        generateID(),
		BookID:    req.BookName,
		Role:      "主角",
		Status:    "存活",
		Relations: []model.Relation{},
	}

	// 解析 JSON
	var data map[string]string
	if err := parseJSON(result, &data); err != nil {
		character.Name = extractJSONValue(result, "name")
		character.Gender = extractJSONValue(result, "gender")
		character.Bio = extractJSONValue(result, "bio")
		character.ExternalGoal = extractJSONValue(result, "goal")
		character.InternalDesire = extractJSONValue(result, "desire")
	} else {
		character.Name = data["name"]
		character.Gender = data["gender"]
		character.Bio = data["bio"]
		character.ExternalGoal = data["goal"]
		character.InternalDesire = data["desire"]
	}

	return character, nil
}

// generateOutline 生成大纲
func (s *InitService) generateOutline(ctx context.Context, req *InitRequest, worldview *model.WorldView, mainChar *model.Character) ([]*VolumeInfo, string, error) {
	targetWords := req.TargetWords
	if targetWords == 0 {
		targetWords = 1000000 // 默认100万字
	}

	volumeCount := req.VolumeCount
	if volumeCount == 0 {
		volumeCount = 5 // 默认5卷
	}

	prompt := fmt.Sprintf(`请为一部%s题材的网络小说生成大纲。

故事核心创意: %s
主角: %s - %s
力量体系: %s
目标字数: %d字
分卷数量: %d卷

请生成详细的大纲，包括分卷和章节，用JSON格式输出：
{
  "synopsis": "故事梗概（200字以内）",
  "volumes": [
    {
      "id": "vol_1",
      "title": "第一卷标题",
      "outline": "本卷主要内容概述",
      "chapters": [
        {"id": 1, "title": "章节标题", "outline": "章节大纲（50字以内）"},
        {"id": 2, "title": "章节标题", "outline": "章节大纲"}
      ]
    }
  ]
}

注意：
1. 每卷约10-20章
2. 每章约3000-5000字
3. 章节大纲要具体，能指导写作`,
		req.Genre, req.Idea, mainChar.Name, mainChar.Bio,
		worldview.CoreSettings.PowerSystem, targetWords, volumeCount)

	result, err := s.llmClient.Call(ctx, prompt, "architect")
	if err != nil {
		return nil, "", err
	}

	// 解析结果
	var outline struct {
		Synopsis string        `json:"synopsis"`
		Volumes  []*VolumeInfo `json:"volumes"`
	}

	if err := parseJSON(result, &outline); err != nil {
		return nil, "", err
	}

	return outline.Volumes, outline.Synopsis, nil
}

// SaveInitResult 保存初始化结果
func (s *InitService) SaveInitResult(result *InitResult) error {
	// 保存世界观
	if err := s.store.SaveWorldView(result.BookName, result.WorldView); err != nil {
		return err
	}

	// 保存主角
	if result.MainCharacter != nil {
		characters := []*model.Character{result.MainCharacter}
		if err := s.store.SaveCharacters(result.BookName, characters); err != nil {
			return err
		}
	}

	// 保存分卷和章节
	volumes := make([]*model.Volume, 0)
	chapters := make([]*model.Chapter, 0)
	now := time.Now()

	chapterID := 1
	for i, vol := range result.Volumes {
		volume := &model.Volume{
			ID:          vol.ID,
			BookID:      result.BookName,
			Title:       vol.Title,
			Order:       i + 1,
			Description: vol.Outline,
		}
		volumes = append(volumes, volume)

		for _, ch := range vol.Chapters {
			chapter := &model.Chapter{
				ID:        chapterID,
				BookID:    result.BookName,
				VolumeID:  vol.ID,
				Title:     ch.Title,
				Outline:   ch.Outline,
				TimeInfo:  model.TimeInfo{Label: "", Duration: "0", Events: []string{}},
				CreatedAt: now,
				UpdatedAt: now,
			}
			chapters = append(chapters, chapter)
			chapterID++
		}
	}

	if len(volumes) > 0 {
		if err := s.store.SaveVolumes(result.BookName, volumes); err != nil {
			return err
		}
	}

	if len(chapters) > 0 {
		if err := s.store.SaveChapters(result.BookName, chapters); err != nil {
			return err
		}
	}

	return nil
}

// generateID 生成唯一 ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}