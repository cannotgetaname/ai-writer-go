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
			loc := char.Faction
			if char.Sect != "" {
				loc = char.Sect
			}
			prompt.WriteString(fmt.Sprintf("- %s: 状态=%s, 势力=%s\n", char.Name, char.Status, loc))
		}
		prompt.WriteString("\n")
	}

	// 现有物品
	if len(items) > 0 {
		prompt.WriteString("【现有物品】\n")
		for _, item := range items {
			prompt.WriteString(fmt.Sprintf("- %s: 持有者=%s, 势力=%s\n", item.Name, item.Owner, item.Faction))
		}
		prompt.WriteString("\n")
	}

	// 现有地点
	if len(locations) > 0 {
		prompt.WriteString("【现有地点】\n")
		for _, loc := range locations {
			prompt.WriteString(fmt.Sprintf("- %s: 势力=%s\n", loc.Name, loc.Faction))
		}
		prompt.WriteString("\n")
	}

	// 使用与 AuditorSystem 提示词一致的 JSON 结构
	prompt.WriteString(`请严格按以下 JSON 结构输出（不要使用 Markdown 代码块）：
{
  "char_updates": [{"name": "名字", "field": "属性名", "new_value": "新值"}],
  "item_updates": [{"name": "物品名", "field": "属性名", "new_value": "新值"}],
  "new_chars": [{"name": "名字", "gender": "性别", "role": "角色类型", "status": "状态", "bio": "简介"}],
  "new_items": [{"name": "物品名", "type": "类型", "owner": "持有者", "desc": "描述"}],
  "new_locs": [{"name": "地名", "faction": "所属势力", "desc": "描述"}],
  "relation_updates": [{"source": "主角", "target": "配角", "type": "关系类型"}]
}

只提取明确发生的变化，不要推测。如果没有变化，返回空数组。`)

	return prompt.String()
}

// parseExtractionResult 解析提取结果（与 AuditorSystem 提示词格式一致）
func (s *SyncService) parseExtractionResult(result string, chapterID int) []StateChange {
	var changes []StateChange

	// 解析 JSON（使用提示词定义的键名）
	var data struct {
		CharUpdates []struct {
			Name     string `json:"name"`
			Field    string `json:"field"`
			NewValue string `json:"new_value"`
		} `json:"char_updates"`
		ItemUpdates []struct {
			Name     string `json:"name"`
			Field    string `json:"field"`
			NewValue string `json:"new_value"`
		} `json:"item_updates"`
		NewChars []struct {
			Name   string `json:"name"`
			Gender string `json:"gender"`
			Role   string `json:"role"`
			Status string `json:"status"`
			Bio    string `json:"bio"`
		} `json:"new_chars"`
		NewItems []struct {
			Name  string `json:"name"`
			Type  string `json:"type"`
			Owner string `json:"owner"`
			Desc  string `json:"desc"`
		} `json:"new_items"`
		NewLocs []struct {
			Name    string `json:"name"`
			Faction string `json:"faction"`
			Desc    string `json:"desc"`
		} `json:"new_locs"`
		RelationUpdates []struct {
			Source string `json:"source"`
			Target string `json:"target"`
			Type   string `json:"type"`
		} `json:"relation_updates"`
	}

	if err := parseJSON(result, &data); err != nil {
		return changes
	}

	// 转换人物状态变更
	for _, cu := range data.CharUpdates {
		changes = append(changes, StateChange{
			ID:        generateID(),
			Type:      "character_status",
			Entity:    cu.Name,
			Field:     cu.Field,
			NewValue:  cu.NewValue,
			Reason:    "从章节内容提取",
			ChapterID: chapterID,
		})
	}

	// 转换物品属性变更
	for _, iu := range data.ItemUpdates {
		changes = append(changes, StateChange{
			ID:        generateID(),
			Type:      "item_update",
			Entity:    iu.Name,
			Field:     iu.Field,
			NewValue:  iu.NewValue,
			Reason:    "从章节内容提取",
			ChapterID: chapterID,
		})
	}

	// 转换新人物
	for _, nc := range data.NewChars {
		changes = append(changes, StateChange{
			ID:        generateID(),
			Type:      "new_character",
			Entity:    nc.Name,
			Field:     "bio",
			NewValue:  nc.Bio,
			Reason:    fmt.Sprintf("性别: %s, 角色: %s, 状态: %s", nc.Gender, nc.Role, nc.Status),
			ChapterID: chapterID,
			// 附加信息存储在扩展字段中
		})
		// 额外记录 gender, role, status 信息
		if nc.Gender != "" {
			changes[len(changes)-1].OldValue = nc.Gender // 暂用 OldValue 存 gender
		}
	}

	// 转换新物品
	for _, ni := range data.NewItems {
		changes = append(changes, StateChange{
			ID:        generateID(),
			Type:      "new_item",
			Entity:    ni.Name,
			Field:     "desc",
			NewValue:  ni.Desc,
			Reason:    fmt.Sprintf("类型: %s, 持有者: %s", ni.Type, ni.Owner),
			ChapterID: chapterID,
		})
	}

	// 转换新地点
	for _, nl := range data.NewLocs {
		changes = append(changes, StateChange{
			ID:        generateID(),
			Type:      "new_location",
			Entity:    nl.Name,
			Field:     "desc",
			NewValue:  nl.Desc,
			Reason:    fmt.Sprintf("势力: %s", nl.Faction),
			ChapterID: chapterID,
		})
	}

	// 转换关系变更
	for _, ru := range data.RelationUpdates {
		changes = append(changes, StateChange{
			ID:        generateID(),
			Type:      "relation_update",
			Entity:    ru.Source,
			Field:     "relation",
			NewValue:  ru.Type,
			Reason:    ru.Target,
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
	case "item_update":
		return s.applyItemChange(bookName, change)
	case "new_character":
		return s.applyNewCharacter(bookName, change)
	case "new_item":
		return s.applyNewItem(bookName, change)
	case "new_location":
		return s.applyNewLocation(bookName, change)
	case "relation_update":
		return s.applyRelationChange(bookName, change)
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
				// 记录状态变更历史
				char.StatusHistory = append(char.StatusHistory, model.StatusChange{
					ChapterID: change.ChapterID,
					Field:     "status",
					OldValue:  char.Status,
					NewValue:  change.NewValue,
					Reason:    change.Reason,
				})
				char.Status = change.NewValue
			case "faction":
				// 记录势力变更历史
				char.FactionHistory = append(char.FactionHistory, model.FactionChange{
					ChapterID:  change.ChapterID,
					OldFaction: char.Faction,
					NewFaction: change.NewValue,
					Reason:     change.Reason,
				})
				char.Faction = change.NewValue
			case "cultivation":
				char.Cultivation = change.NewValue
			case "position":
				char.Position = change.NewValue
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
			switch change.Field {
			case "owner":
				// 记录归属变更历史
				item.OwnerHistory = append(item.OwnerHistory, model.ItemOwnerChange{
					ChapterID: change.ChapterID,
					OldOwner:  item.Owner,
					NewOwner:  change.NewValue,
					Action:    "变更",
					Reason:    change.Reason,
				})
				item.Owner = change.NewValue
			case "faction":
				item.Faction = change.NewValue
			case "location":
				item.Location = change.NewValue
			case "rank":
				item.Rank = change.NewValue
			case "abilities":
				item.Abilities = change.NewValue
			}
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

	// 解析附加信息（Reason 格式: "性别: x, 角色: y, 状态: z"）
	gender := ""
	role := "路人"
	status := "存活"
	bio := change.NewValue

	// OldValue 暂存 gender
	if change.OldValue != "" {
		gender = change.OldValue
	}

	// 从 Reason 解析 role 和 status
	if strings.Contains(change.Reason, "角色:") {
		parts := strings.Split(change.Reason, ", ")
		for _, part := range parts {
			if strings.HasPrefix(part, "角色:") {
				role = strings.TrimPrefix(part, "角色: ")
			}
			if strings.HasPrefix(part, "状态:") {
				status = strings.TrimPrefix(part, "状态: ")
			}
			if strings.HasPrefix(part, "性别:") {
				gender = strings.TrimPrefix(part, "性别: ")
			}
		}
	}

	// 记录出场章节
	newChar := &model.Character{
		ID:            generateID(),
		BookID:        bookName,
		Name:          change.Entity,
		Gender:        gender,
		Role:          role,
		Status:        status,
		Bio:           bio,
		Relations:     []model.Relation{},
		AppearChapters: []int{change.ChapterID},
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

	// 解析附加信息（Reason 格式: "类型: x, 持有者: y"）
	itemType := "未知"
	owner := ""
	desc := change.NewValue

	if strings.Contains(change.Reason, "类型:") {
		parts := strings.Split(change.Reason, ", ")
		for _, part := range parts {
			if strings.HasPrefix(part, "类型:") {
				itemType = strings.TrimPrefix(part, "类型: ")
			}
			if strings.HasPrefix(part, "持有者:") {
				owner = strings.TrimPrefix(part, "持有者: ")
			}
		}
	}

	newItem := &model.Item{
		ID:            generateID(),
		BookID:        bookName,
		Name:          change.Entity,
		Type:          itemType,
		Owner:         owner,
		Description:   desc,
		AppearChapters: []int{change.ChapterID},
	}

	if owner != "" {
		newItem.OwnerHistory = []model.ItemOwnerChange{
			{
				ChapterID: change.ChapterID,
				OldOwner:  "",
				NewOwner:  owner,
				Action:    "获得",
				Reason:    "新物品首次出现",
			},
		}
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

	// 解析附加信息（Reason 格式: "势力: x"）
	faction := ""
	desc := change.NewValue

	if strings.Contains(change.Reason, "势力:") {
		faction = strings.TrimPrefix(change.Reason, "势力: ")
	}

	newLoc := &model.Location{
		ID:          generateID(),
		BookID:      bookName,
		Name:        change.Entity,
		Description: desc,
		Faction:     faction,
	}

	locations = append(locations, newLoc)
	return s.store.SaveLocations(bookName, locations)
}

func (s *SyncService) applyRelationChange(bookName string, change *StateChange) error {
	characters, err := s.store.LoadCharacters(bookName)
	if err != nil {
		return err
	}

	// 找到源人物，添加关系
	for _, char := range characters {
		if char.Name == change.Entity {
			targetName := change.Reason // Reason 存储 target 名字
			relType := change.NewValue  // NewValue 存储 relation type

			// 检查是否已有此关系
			existing := false
			for _, rel := range char.Relations {
				if rel.TargetName == targetName {
					// 更新现有关系
					rel.History = append(rel.History, model.RelationChange{
						ChapterID: change.ChapterID,
						Change:    0,
						Reason:    "关系类型变更为: " + relType,
					})
					rel.Type = relType
					existing = true
					break
				}
			}

			if !existing {
				// 添加新关系
				char.Relations = append(char.Relations, model.Relation{
					TargetID:   "",
					TargetName: targetName,
					Type:       relType,
					Value:      0,
					History: []model.RelationChange{
						{
							ChapterID: change.ChapterID,
							Change:    0,
							Reason:    "新建立关系",
						},
					},
				})
			}
			break
		}
	}

	return s.store.SaveCharacters(bookName, characters)
}