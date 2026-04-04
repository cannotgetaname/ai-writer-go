package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ai-writer/internal/llm"
	"ai-writer/internal/model"
	"ai-writer/internal/store"
)

// SyncService 状态同步服务
type SyncService struct {
	llmClient llm.Client
	store     *store.JSONStore
}

// NewSyncService 创建状态同步服务
func NewSyncService(llmClient llm.Client, jsonStore *store.JSONStore) *SyncService {
	return &SyncService{
		llmClient: llmClient,
		store:     jsonStore,
	}
}

// StateChange 状态变更
type StateChange struct {
	ID        string `json:"id"`
	Type      string `json:"type"`       // character_status/item_owner/new_character/new_item/new_location/time_progression
	Entity    string `json:"entity"`     // 实体名称
	Field     string `json:"field"`      // 变更字段
	OldValue  string `json:"old_value"`  // 旧值
	NewValue  string `json:"new_value"`  // 新值
	Reason    string `json:"reason"`     // 变更原因
	Approved  bool   `json:"approved"`   // 是否已审核
	ChapterID int    `json:"chapter_id"` // 来源章节
}

// PendingChanges 待审核变更
type PendingChanges struct {
	BookID      string        `json:"book_id"`
	ChapterID   int           `json:"chapter_id"`
	ExtractedAt time.Time     `json:"extracted_at"`
	Changes     []StateChange `json:"changes"`
}

// ExtractStateChanges 从章节提取状态变更
func (s *SyncService) ExtractStateChanges(ctx context.Context, bookName string, chapterID int) (*PendingChanges, error) {
	// 加载章节内容
	content, err := s.store.LoadChapterContent(bookName, chapterID)
	if err != nil {
		return nil, err
	}

	if content == "" {
		return nil, fmt.Errorf("章节 %d 没有内容", chapterID)
	}

	// 加载现有数据作为上下文
	characters, _ := s.store.LoadCharacters(bookName)
	items, _ := s.store.LoadItems(bookName)
	locations, _ := s.store.LoadLocations(bookName)

	// 构建提取提示词
	prompt := s.buildExtractionPrompt(content, characters, items, locations)

	// 调用 LLM
	result, err := s.llmClient.Call(ctx, prompt, "auditor")
	if err != nil {
		return nil, err
	}

	// 解析结果
	changes := s.parseExtractionResult(result, chapterID)

	pending := &PendingChanges{
		BookID:      bookName,
		ChapterID:   chapterID,
		ExtractedAt: time.Now(),
		Changes:     changes,
	}

	return pending, nil
}

// buildExtractionPrompt 构建提取提示词
func (s *SyncService) buildExtractionPrompt(content string, characters []*model.Character, items []*model.Item, locations []*model.Location) string {
	var prompt strings.Builder

	prompt.WriteString("请分析以下章节内容，提取状态变更信息。\n\n")

	// 章节内容
	prompt.WriteString("【章节内容】\n")
	if len(content) > 3000 {
		prompt.WriteString(content[:3000] + "...")
	} else {
		prompt.WriteString(content)
	}
	prompt.WriteString("\n\n")

	// 现有人物
	if len(characters) > 0 {
		prompt.WriteString("【现有人物】\n")
		for _, char := range characters {
			prompt.WriteString(fmt.Sprintf("- %s: 状态=%s, 位置=%s\n", char.Name, char.Status, ""))
		}
		prompt.WriteString("\n")
	}

	// 现有物品
	if len(items) > 0 {
		prompt.WriteString("【现有物品】\n")
		for _, item := range items {
			prompt.WriteString(fmt.Sprintf("- %s: 持有者=%s\n", item.Name, item.Owner))
		}
		prompt.WriteString("\n")
	}

	// 现有地点
	if len(locations) > 0 {
		prompt.WriteString("【现有地点】\n")
		for _, loc := range locations {
			prompt.WriteString(fmt.Sprintf("- %s\n", loc.Name))
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString(`请提取以下类型的状态变更，用JSON格式输出：
{
  "character_changes": [
    {"name": "角色名", "field": "status/location/emotion", "old": "旧值", "new": "新值", "reason": "原因"}
  ],
  "item_changes": [
    {"name": "物品名", "old_owner": "旧持有者", "new_owner": "新持有者", "reason": "原因"}
  ],
  "new_characters": [
    {"name": "新角色名", "role": "角色类型", "bio": "简介"}
  ],
  "new_items": [
    {"name": "新物品名", "type": "类型", "owner": "持有者"}
  ],
  "new_locations": [
    {"name": "新地点名", "type": "类型"}
  ],
  "time_progression": "时间推进描述"
}

只提取明确发生的变化，不要推测。如果没有变化，返回空数组。`)

	return prompt.String()
}

// parseExtractionResult 解析提取结果
func (s *SyncService) parseExtractionResult(result string, chapterID int) []StateChange {
	var changes []StateChange

	// 解析 JSON
	var data struct {
		CharacterChanges []struct {
			Name   string `json:"name"`
			Field  string `json:"field"`
			Old    string `json:"old"`
			New    string `json:"new"`
			Reason string `json:"reason"`
		} `json:"character_changes"`
		ItemChanges []struct {
			Name      string `json:"name"`
			OldOwner  string `json:"old_owner"`
			NewOwner  string `json:"new_owner"`
			Reason    string `json:"reason"`
		} `json:"item_changes"`
		NewCharacters []struct {
			Name string `json:"name"`
			Role string `json:"role"`
			Bio  string `json:"bio"`
		} `json:"new_characters"`
		NewItems []struct {
			Name  string `json:"name"`
			Type  string `json:"type"`
			Owner string `json:"owner"`
		} `json:"new_items"`
		NewLocations []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"new_locations"`
		TimeProgression string `json:"time_progression"`
	}

	if err := parseJSON(result, &data); err != nil {
		return changes
	}

	// 转换人物状态变更
	for _, cc := range data.CharacterChanges {
		changes = append(changes, StateChange{
			ID:        generateID(),
			Type:      "character_status",
			Entity:    cc.Name,
			Field:     cc.Field,
			OldValue:  cc.Old,
			NewValue:  cc.New,
			Reason:    cc.Reason,
			ChapterID: chapterID,
		})
	}

	// 转换物品归属变更
	for _, ic := range data.ItemChanges {
		changes = append(changes, StateChange{
			ID:        generateID(),
			Type:      "item_owner",
			Entity:    ic.Name,
			Field:     "owner",
			OldValue:  ic.OldOwner,
			NewValue:  ic.NewOwner,
			Reason:    ic.Reason,
			ChapterID: chapterID,
		})
	}

	// 转换新人物
	for _, nc := range data.NewCharacters {
		changes = append(changes, StateChange{
			ID:        generateID(),
			Type:      "new_character",
			Entity:    nc.Name,
			Field:     "role",
			NewValue:  nc.Role,
			Reason:    nc.Bio,
			ChapterID: chapterID,
		})
	}

	// 转换新物品
	for _, ni := range data.NewItems {
		changes = append(changes, StateChange{
			ID:        generateID(),
			Type:      "new_item",
			Entity:    ni.Name,
			Field:     "type",
			NewValue:  ni.Type,
			Reason:    "持有者: " + ni.Owner,
			ChapterID: chapterID,
		})
	}

	// 转换新地点
	for _, nl := range data.NewLocations {
		changes = append(changes, StateChange{
			ID:        generateID(),
			Type:      "new_location",
			Entity:    nl.Name,
			Field:     "type",
			NewValue:  nl.Type,
			ChapterID: chapterID,
		})
	}

	// 时间推进
	if data.TimeProgression != "" {
		changes = append(changes, StateChange{
			ID:        generateID(),
			Type:      "time_progression",
			Field:     "time",
			NewValue:  data.TimeProgression,
			ChapterID: chapterID,
		})
	}

	return changes
}

// ApplyChange 应用状态变更
func (s *SyncService) ApplyChange(bookName string, change *StateChange) error {
	switch change.Type {
	case "character_status":
		return s.applyCharacterChange(bookName, change)
	case "item_owner":
		return s.applyItemChange(bookName, change)
	case "new_character":
		return s.applyNewCharacter(bookName, change)
	case "new_item":
		return s.applyNewItem(bookName, change)
	case "new_location":
		return s.applyNewLocation(bookName, change)
	case "time_progression":
		return s.applyTimeProgression(bookName, change)
	}
	return nil
}

func (s *SyncService) applyCharacterChange(bookName string, change *StateChange) error {
	characters, err := s.store.LoadCharacters(bookName)
	if err != nil {
		return err
	}

	for _, char := range characters {
		if char.Name == change.Entity {
			switch change.Field {
			case "status":
				char.Status = change.NewValue
			case "location":
				// 可以添加位置字段到 Character
			case "emotion":
				char.EmotionalArc = append(char.EmotionalArc, model.EmotionPoint{
					ChapterID: change.ChapterID,
					Emotion:   change.NewValue,
					Trigger:   change.Reason,
				})
			}
			break
		}
	}

	return s.store.SaveCharacters(bookName, characters)
}

func (s *SyncService) applyItemChange(bookName string, change *StateChange) error {
	items, err := s.store.LoadItems(bookName)
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.Name == change.Entity {
			item.Owner = change.NewValue
			break
		}
	}

	return s.store.SaveItems(bookName, items)
}

func (s *SyncService) applyNewCharacter(bookName string, change *StateChange) error {
	characters, err := s.store.LoadCharacters(bookName)
	if err != nil {
		characters = []*model.Character{}
	}

	// 检查是否已存在
	for _, char := range characters {
		if char.Name == change.Entity {
			return nil // 已存在，跳过
		}
	}

	newChar := &model.Character{
		ID:        generateID(),
		BookID:    bookName,
		Name:      change.Entity,
		Role:      change.NewValue,
		Bio:       change.Reason,
		Status:    "存活",
		Relations: []model.Relation{},
	}

	characters = append(characters, newChar)
	return s.store.SaveCharacters(bookName, characters)
}

func (s *SyncService) applyNewItem(bookName string, change *StateChange) error {
	items, err := s.store.LoadItems(bookName)
	if err != nil {
		items = []*model.Item{}
	}

	// 检查是否已存在
	for _, item := range items {
		if item.Name == change.Entity {
			return nil
		}
	}

	newItem := &model.Item{
		ID:          generateID(),
		BookID:      bookName,
		Name:        change.Entity,
		Type:        change.NewValue,
		Description: change.Reason,
	}

	items = append(items, newItem)
	return s.store.SaveItems(bookName, items)
}

func (s *SyncService) applyNewLocation(bookName string, change *StateChange) error {
	locations, err := s.store.LoadLocations(bookName)
	if err != nil {
		locations = []*model.Location{}
	}

	// 检查是否已存在
	for _, loc := range locations {
		if loc.Name == change.Entity {
			return nil
		}
	}

	newLoc := &model.Location{
		ID:          generateID(),
		BookID:      bookName,
		Name:        change.Entity,
		Description: change.NewValue,
	}

	locations = append(locations, newLoc)
	return s.store.SaveLocations(bookName, locations)
}

func (s *SyncService) applyTimeProgression(bookName string, change *StateChange) error {
	chapters, err := s.store.LoadChapters(bookName)
	if err != nil {
		return err
	}

	for _, ch := range chapters {
		if ch.ID == change.ChapterID {
			ch.TimeInfo.Label = change.NewValue
			break
		}
	}

	return s.store.SaveChapters(bookName, chapters)
}