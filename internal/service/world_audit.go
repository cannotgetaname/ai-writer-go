package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ai-writer/internal/config"
	"ai-writer/internal/llm"
	"ai-writer/internal/model"
	"ai-writer/internal/store"
)

// WorldAuditService 世界状态审计服务
type WorldAuditService struct {
	llmClient llm.Client
	store     *store.JSONStore
	prompts   *config.PromptsConfig
}

// NewWorldAuditService 创建审计服务
func NewWorldAuditService(llmClient llm.Client, jsonStore *store.JSONStore, prompts *config.PromptsConfig) *WorldAuditService {
	return &WorldAuditService{
		llmClient: llmClient,
		store:     jsonStore,
		prompts:   prompts,
	}
}

// ExtractAll 提取所有图谱数据
func (s *WorldAuditService) ExtractAll(ctx context.Context, bookName string, chapterID int) (*model.PendingGraphSync, error) {
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
	foreshadows, _ := s.store.LoadForeshadows(bookName)
	threads, _ := s.store.LoadThreads(bookName)
	timeline, _ := s.store.LoadTimeline(bookName)

	// 构建提示词
	prompt := s.buildExtractPrompt(content, chapterID, characters, items, locations, foreshadows, threads, timeline)

	// 调用 LLM
	systemPrompt := s.prompts.AuditorSystem
	result, err := s.llmClient.CallWithSystem(ctx, systemPrompt, prompt, "auditor")
	if err != nil {
		return nil, err
	}

	// 解析结果
	pending := s.parseExtractResult(result, bookName, chapterID)

	// 保存待审核数据
	if err := s.store.SavePendingGraphSync(bookName, pending); err != nil {
		return nil, err
	}

	return pending, nil
}

// buildExtractPrompt 构建提取提示词
func (s *WorldAuditService) buildExtractPrompt(content string, chapterID int,
	characters []*model.Character, items []*model.Item, locations []*model.Location,
	foreshadows []*model.Foreshadow, threads []*model.NarrativeThread, timeline []model.TimelineEvent) string {

	var prompt strings.Builder

	prompt.WriteString("请分析以下章节内容，提取世界状态变更信息。\n\n")

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
			prompt.WriteString(fmt.Sprintf("- %s: 角色=%s, 状态=%s, 势力=%s\n",
				char.Name, char.Role, char.Status, char.Faction))
		}
		prompt.WriteString("\n")
	}

	// 现有物品
	if len(items) > 0 {
		prompt.WriteString("【现有物品】\n")
		for _, item := range items {
			prompt.WriteString(fmt.Sprintf("- %s: 类型=%s, 持有者=%s\n",
				item.Name, item.Type, item.Owner))
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

	// 现有伏笔
	if len(foreshadows) > 0 {
		prompt.WriteString("【现有伏笔】\n")
		for _, fs := range foreshadows {
			if fs.Status == model.ForeshadowActive {
				prompt.WriteString(fmt.Sprintf("- [%s] %s (埋设于第%d章)\n",
					fs.Type, fs.Content, fs.SourceChapter))
			}
		}
		prompt.WriteString("\n")
	}

	// 现有线程
	if len(threads) > 0 {
		prompt.WriteString("【现有叙事线程】\n")
		for _, th := range threads {
			if th.Status == model.ThreadActive {
				prompt.WriteString(fmt.Sprintf("- [%s] %s (最后活跃: 第%d章)\n",
					th.Type, th.Name, th.LastActiveChapter))
			}
		}
		prompt.WriteString("\n")
	}

	// 时间线
	if len(timeline) > 0 {
		prompt.WriteString("【时间线】\n")
		for _, te := range timeline {
			prompt.WriteString(fmt.Sprintf("- 第%d章: %s\n", te.ChapterID, te.TimeLabel))
		}
		prompt.WriteString("\n")
	}

	// JSON 输出格式
	prompt.WriteString(`请严格按以下 JSON 格式输出（不要使用 Markdown 代码块）：
{
  "state_changes": [
    {"id": "sc_1", "type": "character_status", "entity": "角色名", "field": "status", "old_value": "旧值", "new_value": "新值", "reason": "原因"}
  ],
  "causal_events": [
    {"cause": "原因", "event": "事件", "effect": "后果", "decision": "决定", "characters": ["角色1", "角色2"]}
  ],
  "foreshadows": [
    {"type": "item", "content": "伏笔内容", "importance": "high", "target_chapter": 20}
  ],
  "thread_updates": [
    {"thread_name": "线程名", "update_type": "chapter_add", "chapters": [5], "pov_characters": ["角色名"]}
  ],
  "emotion_points": [
    {"character_name": "角色名", "emotion": "愤怒", "intensity": 8, "trigger": "触发事件"}
  ],
  "timeline_events": [
    {"time_label": "时间描述", "duration": "持续时间", "events": ["事件1", "事件2"], "characters": ["角色名"], "location": "地点"}
  ]
}

只提取明确发生的变化，不要推测。如果没有变化，返回空数组。`)

	return prompt.String()
}

// parseExtractResult 解析提取结果
func (s *WorldAuditService) parseExtractResult(result string, bookName string, chapterID int) *model.PendingGraphSync {
	pending := &model.PendingGraphSync{
		BookID:      bookName,
		ChapterID:   chapterID,
		ExtractedAt: time.Now(),
		StateChanges:   []model.StateChangeItem{},
		CausalEvents:   []model.CausalEventItem{},
		Foreshadows:    []model.ForeshadowItem{},
		ThreadUpdates:  []model.ThreadUpdateItem{},
		EmotionPoints:  []model.EmotionPointItem{},
		TimelineEvents: []model.TimelineEventItem{},
	}

	// 定义解析结构
	var data struct {
		StateChanges []struct {
			Type     string `json:"type"`
			Entity   string `json:"entity"`
			Field    string `json:"field"`
			OldValue string `json:"old_value"`
			NewValue string `json:"new_value"`
			Reason   string `json:"reason"`
		} `json:"state_changes"`
		CausalEvents []struct {
			Cause     string   `json:"cause"`
			Event     string   `json:"event"`
			Effect    string   `json:"effect"`
			Decision  string   `json:"decision"`
			Characters []string `json:"characters"`
		} `json:"causal_events"`
		Foreshadows []struct {
			Type         string `json:"type"`
			Content      string `json:"content"`
			Importance   string `json:"importance"`
			TargetChapter int   `json:"target_chapter"`
		} `json:"foreshadows"`
		ThreadUpdates []struct {
			ThreadName    string   `json:"thread_name"`
			ThreadID      string   `json:"thread_id"`
			UpdateType    string   `json:"update_type"`
			Chapters      []int    `json:"chapters"`
			POVCharacters []string `json:"pov_characters"`
		} `json:"thread_updates"`
		EmotionPoints []struct {
			CharacterName string `json:"character_name"`
			Emotion       string `json:"emotion"`
			Intensity     int    `json:"intensity"`
			Trigger       string `json:"trigger"`
		} `json:"emotion_points"`
		TimelineEvents []struct {
			TimeLabel  string   `json:"time_label"`
			Duration   string   `json:"duration"`
			Events     []string `json:"events"`
			Characters []string `json:"characters"`
			Location   string   `json:"location"`
		} `json:"timeline_events"`
	}

	if err := parseJSON(result, &data); err != nil {
		return pending
	}

	// 转换状态变更
	for _, sc := range data.StateChanges {
		pending.StateChanges = append(pending.StateChanges, model.StateChangeItem{
			ID:       generateID(),
			Type:     sc.Type,
			Entity:   sc.Entity,
			Field:    sc.Field,
			OldValue: sc.OldValue,
			NewValue: sc.NewValue,
			Reason:   sc.Reason,
			Status:   "pending",
		})
	}

	// 转换因果事件
	for _, ce := range data.CausalEvents {
		pending.CausalEvents = append(pending.CausalEvents, model.CausalEventItem{
			CausalEvent: model.CausalEvent{
				ID:        generateID(),
				BookID:    bookName,
				ChapterID: chapterID,
				Cause:     ce.Cause,
				Event:     ce.Event,
				Effect:    ce.Effect,
				Decision:  ce.Decision,
				Characters: ce.Characters,
				Status:    model.CausalActive,
				AutoDetected: true,
				CreatedAt: time.Now(),
			},
			Status: "pending",
		})
	}

	// 转换伏笔
	for _, fs := range data.Foreshadows {
		pending.Foreshadows = append(pending.Foreshadows, model.ForeshadowItem{
			Foreshadow: model.Foreshadow{
				ID:             generateID(),
				BookID:         bookName,
				Type:           model.ForeshadowType(fs.Type),
				Content:        fs.Content,
				Importance:     model.Importance(fs.Importance),
				SourceChapter:  chapterID,
				TargetChapter:  fs.TargetChapter,
				Status:         model.ForeshadowActive,
				AutoDetected:   true,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			Status: "pending",
		})
	}

	// 转换线程更新
	for _, tu := range data.ThreadUpdates {
		pending.ThreadUpdates = append(pending.ThreadUpdates, model.ThreadUpdateItem{
			ID:            generateID(),
			ThreadName:    tu.ThreadName,
			ThreadID:      tu.ThreadID,
			UpdateType:    tu.UpdateType,
			Chapters:      tu.Chapters,
			POVCharacters: tu.POVCharacters,
			Status:        "pending",
		})
	}

	// 转换情感点
	for _, ep := range data.EmotionPoints {
		pending.EmotionPoints = append(pending.EmotionPoints, model.EmotionPointItem{
			ID:            generateID(),
			CharacterName: ep.CharacterName,
			EmotionPoint: model.EmotionPoint{
				ChapterID: chapterID,
				Emotion:   ep.Emotion,
				Intensity: ep.Intensity,
				Trigger:   ep.Trigger,
			},
			Status: "pending",
		})
	}

	// 转换时间线事件
	for _, te := range data.TimelineEvents {
		pending.TimelineEvents = append(pending.TimelineEvents, model.TimelineEventItem{
			ID: generateID(),
			TimelineEvent: model.TimelineEvent{
				ChapterID:    chapterID,
				TimeLabel:    te.TimeLabel,
				Duration:     te.Duration,
				Events:       te.Events,
				Characters:   te.Characters,
				Location:     te.Location,
				AutoDetected: true,
			},
			Status: "pending",
		})
	}

	return pending
}

// ApplyChanges 应用审核后的变更
func (s *WorldAuditService) ApplyChanges(ctx context.Context, bookName string, acceptedIDs []string) error {
	pending, err := s.store.LoadPendingGraphSync(bookName)
	if err != nil || pending == nil {
		return fmt.Errorf("无待审核变更")
	}

	// 创建已接受 ID 的集合
	acceptedSet := make(map[string]bool)
	for _, id := range acceptedIDs {
		acceptedSet[id] = true
	}

	// 应用状态变更
	if err := s.applyStateChanges(bookName, pending.StateChanges, acceptedSet, pending.ChapterID); err != nil {
		return err
	}

	// 应用因果链
	if err := s.applyCausalEvents(bookName, pending.CausalEvents, acceptedSet); err != nil {
		return err
	}

	// 应用伏笔
	if err := s.applyForeshadows(bookName, pending.Foreshadows, acceptedSet); err != nil {
		return err
	}

	// 应用线程更新
	if err := s.applyThreadUpdates(bookName, pending.ThreadUpdates, acceptedSet, pending.ChapterID); err != nil {
		return err
	}

	// 应用情感点
	if err := s.applyEmotionPoints(bookName, pending.EmotionPoints, acceptedSet); err != nil {
		return err
	}

	// 应用时间线
	if err := s.applyTimelineEvents(bookName, pending.TimelineEvents, acceptedSet); err != nil {
		return err
	}

	// 清除已处理的待审核数据
	return s.store.ClearPendingGraphSync(bookName)
}

// applyStateChanges 应用状态变更
func (s *WorldAuditService) applyStateChanges(bookName string, changes []model.StateChangeItem, acceptedSet map[string]bool, chapterID int) error {
	for _, change := range changes {
		if !acceptedSet[change.ID] {
			continue
		}

		switch change.Type {
		case "character_status":
			if err := s.applyCharacterStatusChange(bookName, &change, chapterID); err != nil {
				return err
			}
		case "item_owner":
			if err := s.applyItemOwnerChange(bookName, &change, chapterID); err != nil {
				return err
			}
		case "relation":
			if err := s.applyRelationChange(bookName, &change, chapterID); err != nil {
				return err
			}
		}
	}
	return nil
}

// applyCharacterStatusChange 应用人物状态变更
func (s *WorldAuditService) applyCharacterStatusChange(bookName string, change *model.StateChangeItem, chapterID int) error {
	characters, err := s.store.LoadCharacters(bookName)
	if err != nil {
		return err
	}

	for _, char := range characters {
		if char.Name == change.Entity {
			switch change.Field {
			case "status":
				char.StatusHistory = append(char.StatusHistory, model.StatusChange{
					ChapterID: chapterID,
					Field:     "status",
					OldValue:  char.Status,
					NewValue:  change.NewValue,
					Reason:    change.Reason,
					ChangedAt: time.Now(),
				})
				char.Status = change.NewValue
			case "faction":
				char.FactionHistory = append(char.FactionHistory, model.FactionChange{
					ChapterID:  chapterID,
					OldFaction: char.Faction,
					NewFaction: change.NewValue,
					Reason:     change.Reason,
				})
				char.Faction = change.NewValue
			case "cultivation":
				char.Cultivation = change.NewValue
			case "position":
				char.Position = change.NewValue
			}
			break
		}
	}

	return s.store.SaveCharacters(bookName, characters)
}

// applyItemOwnerChange 应用物品持有者变更
func (s *WorldAuditService) applyItemOwnerChange(bookName string, change *model.StateChangeItem, chapterID int) error {
	items, err := s.store.LoadItems(bookName)
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.Name == change.Entity {
			item.OwnerHistory = append(item.OwnerHistory, model.ItemOwnerChange{
				ChapterID: chapterID,
				OldOwner:  change.OldValue,
				NewOwner:  change.NewValue,
				Action:    "变更",
				Reason:    change.Reason,
			})
			item.Owner = change.NewValue
			break
		}
	}

	return s.store.SaveItems(bookName, items)
}

// applyRelationChange 应用关系变更
func (s *WorldAuditService) applyRelationChange(bookName string, change *model.StateChangeItem, chapterID int) error {
	characters, err := s.store.LoadCharacters(bookName)
	if err != nil {
		return err
	}

	// 解析关系变更信息
	// 格式: target 在 OldValue 中，关系类型在 NewValue 中
	targetName := change.OldValue
	relType := change.NewValue

	for _, char := range characters {
		if char.Name == change.Entity {
			// 检查是否已有此关系
			existing := false
			for i, rel := range char.Relations {
				if rel.TargetName == targetName {
					char.Relations[i].Type = relType
					char.Relations[i].History = append(char.Relations[i].History, model.RelationChange{
						ChapterID: chapterID,
						Change:    0,
						Reason:    change.Reason,
					})
					existing = true
					break
				}
			}

			if !existing {
				char.Relations = append(char.Relations, model.Relation{
					TargetID:   "",
					TargetName: targetName,
					Type:       relType,
					Value:      0,
					History: []model.RelationChange{
						{
							ChapterID: chapterID,
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

// applyCausalEvents 应用因果链事件
func (s *WorldAuditService) applyCausalEvents(bookName string, events []model.CausalEventItem, acceptedSet map[string]bool) error {
	existingEvents, err := s.store.LoadCausalChains(bookName)
	if err != nil {
		existingEvents = []*model.CausalEvent{}
	}

	for _, event := range events {
		if !acceptedSet[event.CausalEvent.ID] {
			continue
		}

		// 添加到现有因果链
		existingEvents = append(existingEvents, &event.CausalEvent)
	}

	return s.store.SaveCausalChains(bookName, existingEvents)
}

// applyForeshadows 应用伏笔
func (s *WorldAuditService) applyForeshadows(bookName string, foreshadows []model.ForeshadowItem, acceptedSet map[string]bool) error {
	existingForeshadows, err := s.store.LoadForeshadows(bookName)
	if err != nil {
		existingForeshadows = []*model.Foreshadow{}
	}

	for _, fs := range foreshadows {
		if !acceptedSet[fs.Foreshadow.ID] {
			continue
		}

		// 添加到现有伏笔列表
		existingForeshadows = append(existingForeshadows, &fs.Foreshadow)
	}

	return s.store.SaveForeshadows(bookName, existingForeshadows)
}

// applyThreadUpdates 应用线程更新
func (s *WorldAuditService) applyThreadUpdates(bookName string, updates []model.ThreadUpdateItem, acceptedSet map[string]bool, chapterID int) error {
	threads, err := s.store.LoadThreads(bookName)
	if err != nil {
		threads = []*model.NarrativeThread{}
	}

	for _, update := range updates {
		if !acceptedSet[update.ID] {
			continue
		}

		switch update.UpdateType {
		case "new":
			// 创建新线程
			newThread := &model.NarrativeThread{
				ID:             generateID(),
				BookID:         bookName,
				Name:           update.ThreadName,
				Type:           model.ThreadSub, // 默认为支线
				POVCharacters:  update.POVCharacters,
				StartChapter:   chapterID,
				Chapters:       update.Chapters,
				Status:         model.ThreadActive,
				LastActiveChapter: chapterID,
				AutoDetected:   true,
				CreatedAt:      time.Now(),
			}
			threads = append(threads, newThread)

		case "chapter_add":
			// 向现有线程添加章节
			for _, th := range threads {
				if th.Name == update.ThreadName || th.ID == update.ThreadID {
					th.Chapters = append(th.Chapters, update.Chapters...)
					th.LastActiveChapter = chapterID
					break
				}
			}

		case "pov_change":
			// 更新 POV 角色
			for _, th := range threads {
				if th.Name == update.ThreadName || th.ID == update.ThreadID {
					th.POVCharacters = update.POVCharacters
					th.LastActiveChapter = chapterID
					break
				}
			}
		}
	}

	return s.store.SaveThreads(bookName, threads)
}

// applyEmotionPoints 应用情感点
func (s *WorldAuditService) applyEmotionPoints(bookName string, points []model.EmotionPointItem, acceptedSet map[string]bool) error {
	// 使用 AppendCharacterEmotion 方法逐个添加
	for _, point := range points {
		if !acceptedSet[point.ID] {
			continue
		}

		err := s.store.AppendCharacterEmotion(bookName, point.CharacterName, point.EmotionPoint)
		if err != nil {
			return err
		}
	}

	return nil
}

// applyTimelineEvents 应用时间线事件
func (s *WorldAuditService) applyTimelineEvents(bookName string, events []model.TimelineEventItem, acceptedSet map[string]bool) error {
	// 加载现有时间线
	timeline, err := s.store.LoadTimeline(bookName)
	if err != nil {
		timeline = []model.TimelineEvent{}
	}

	for _, event := range events {
		if !acceptedSet[event.ID] {
			continue
		}

		timeline = append(timeline, event.TimelineEvent)
	}

	// 保存时间线
	return s.saveTimeline(bookName, timeline)
}

// saveTimeline 保存时间线（JSONStore 没有直接的 SaveTimeline 方法，需要手动实现）
func (s *WorldAuditService) saveTimeline(bookName string, timeline []model.TimelineEvent) error {
	// 使用 JSONStore 的内部方法保存
	// 由于没有直接的 SaveTimeline 方法，我们需要绕过保存
	// 这里可以添加一个通用的保存方法或者直接操作文件
	// 暂时通过 SaveChapters 更新章节的时间信息

	// 更新章节的 TimeInfo
	chapters, err := s.store.LoadChapters(bookName)
	if err != nil {
		return err
	}

	for _, event := range timeline {
		for _, ch := range chapters {
			if ch.ID == event.ChapterID {
				ch.TimeInfo.Label = event.TimeLabel
				ch.TimeInfo.Duration = event.Duration
				ch.TimeInfo.Events = event.Events
				break
			}
		}
	}

	return s.store.SaveChapters(bookName, chapters)
}

// RejectChanges 拒绝变更
func (s *WorldAuditService) RejectChanges(bookName string, rejectedIDs []string) error {
	// 清除待审核数据即可，因为未被接受的变更不会被应用
	// 如果需要记录拒绝历史，可以扩展此方法
	return s.store.ClearPendingGraphSync(bookName)
}

// GetPending 获取待审核变更
func (s *WorldAuditService) GetPending(bookName string) (*model.PendingGraphSync, error) {
	return s.store.LoadPendingGraphSync(bookName)
}

// GetPendingStats 获取待审核变更统计
func (s *WorldAuditService) GetPendingStats(bookName string) (map[string]int, error) {
	pending, err := s.store.LoadPendingGraphSync(bookName)
	if err != nil || pending == nil {
		return map[string]int{}, nil
	}

	return map[string]int{
		"state_changes":    len(pending.StateChanges),
		"causal_events":    len(pending.CausalEvents),
		"foreshadows":      len(pending.Foreshadows),
		"thread_updates":   len(pending.ThreadUpdates),
		"emotion_points":   len(pending.EmotionPoints),
		"timeline_events":  len(pending.TimelineEvents),
		"total":            len(pending.StateChanges) + len(pending.CausalEvents) +
		                    len(pending.Foreshadows) + len(pending.ThreadUpdates) +
		                    len(pending.EmotionPoints) + len(pending.TimelineEvents),
	}, nil
}

