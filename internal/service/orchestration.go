package service

import (
	"context"
	"fmt"
	"strings"

	"ai-writer/internal/engine"
	"ai-writer/internal/llm"
	"ai-writer/internal/model"
	"ai-writer/internal/store"
)

// OrchestrationService 编排服务
// 负责协调各服务之间的自动化流程
type OrchestrationService struct {
	llmClient      llm.Client
	store          *store.JSONStore
	syncService    *SyncService
	causalEngine   *engine.CausalChainEngine
	emotionTracker *engine.EmotionalArcTracker
	infoManager    *engine.InfoBoundaryManager
}

// NewOrchestrationService 创建编排服务
func NewOrchestrationService(llmClient llm.Client, jsonStore *store.JSONStore) *OrchestrationService {
	return &OrchestrationService{
		llmClient:      llmClient,
		store:          jsonStore,
		syncService:    NewSyncService(llmClient, jsonStore),
		causalEngine:   engine.NewCausalChainEngine(llmClient, jsonStore),
		emotionTracker: engine.NewEmotionalArcTracker(llmClient, jsonStore),
		infoManager:    engine.NewInfoBoundaryManager(llmClient, jsonStore),
	}
}

// AfterChapterGenerated 章节生成后的自动编排
// 在后台异步执行，不阻塞主流程
func (s *OrchestrationService) AfterChapterGenerated(ctx context.Context, bookName string, chapterID int) error {
	var errors []string

	// 1. 状态同步 - 提取状态变更
	if err := s.extractAndApplyStateChanges(ctx, bookName, chapterID); err != nil {
		errors = append(errors, fmt.Sprintf("状态同步失败: %v", err))
	}

	// 2. 因果链提取
	if err := s.extractAndSaveCausalChain(ctx, bookName, chapterID); err != nil {
		errors = append(errors, fmt.Sprintf("因果链提取失败: %v", err))
	}

	// 3. 情感追踪并持久化
	if err := s.trackAndSaveEmotions(ctx, bookName, chapterID); err != nil {
		errors = append(errors, fmt.Sprintf("情感追踪失败: %v", err))
	}

	// 4. 信息边界提取并持久化
	if err := s.extractAndSaveKnownInfos(ctx, bookName, chapterID); err != nil {
		errors = append(errors, fmt.Sprintf("信息提取失败: %v", err))
	}

	// 5. 更新章节实体索引
	if err := s.updateChapterEntityIndex(bookName, chapterID); err != nil {
		errors = append(errors, fmt.Sprintf("实体索引更新失败: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("编排过程部分失败: %s", strings.Join(errors, "; "))
	}
	return nil
}

// extractAndApplyStateChanges 提取并应用状态变更
func (s *OrchestrationService) extractAndApplyStateChanges(ctx context.Context, bookName string, chapterID int) error {
	pending, err := s.syncService.ExtractStateChanges(ctx, bookName, chapterID)
	if err != nil {
		return err
	}

	// 自动应用所有变更
	for _, change := range pending.Changes {
		if err := s.syncService.ApplyChange(bookName, &change); err != nil {
			// 记录错误但继续处理其他变更
			continue
		}
	}

	return nil
}

// extractAndSaveCausalChain 提取并保存因果链
func (s *OrchestrationService) extractAndSaveCausalChain(ctx context.Context, bookName string, chapterID int) error {
	event, err := s.causalEngine.ExtractFromChapter(ctx, bookName, chapterID)
	if err != nil {
		return err
	}

	if event == nil || event.Event == "" {
		return nil // 没有提取到因果事件
	}

	// 保存因果事件
	events, _ := s.store.LoadCausalChains(bookName)
	events = append(events, event)
	return s.store.SaveCausalChains(bookName, events)
}

// trackAndSaveEmotions 追踪并保存情感
func (s *OrchestrationService) trackAndSaveEmotions(ctx context.Context, bookName string, chapterID int) error {
	emotions, err := s.emotionTracker.TrackEmotion(ctx, bookName, chapterID)
	if err != nil {
		return err
	}

	if len(emotions) == 0 {
		return nil
	}

	// 持久化到角色数据
	characters, err := s.store.LoadCharacters(bookName)
	if err != nil {
		return err
	}

	for _, char := range characters {
		if point, ok := emotions[char.Name]; ok {
			char.EmotionalArc = append(char.EmotionalArc, point)
		}
	}

	return s.store.SaveCharacters(bookName, characters)
}

// extractAndSaveKnownInfos 提取并保存已知信息
func (s *OrchestrationService) extractAndSaveKnownInfos(ctx context.Context, bookName string, chapterID int) error {
	infoMap, err := s.infoManager.ExtractInfoFromChapter(ctx, bookName, chapterID)
	if err != nil {
		return err
	}

	if len(infoMap) == 0 {
		return nil
	}

	// 持久化到角色数据
	characters, err := s.store.LoadCharacters(bookName)
	if err != nil {
		return err
	}

	for _, char := range characters {
		if infos, ok := infoMap[char.Name]; ok {
			char.KnownInfos = append(char.KnownInfos, infos...)
		}
	}

	return s.store.SaveCharacters(bookName, characters)
}

// updateChapterEntityIndex 更新章节实体索引
func (s *OrchestrationService) updateChapterEntityIndex(bookName string, chapterID int) error {
	// 加载章节内容
	content, err := s.store.LoadChapterContent(bookName, chapterID)
	if err != nil || content == "" {
		return err
	}

	// 加载实体数据用于匹配
	characters, _ := s.store.LoadCharacters(bookName)
	items, _ := s.store.LoadItems(bookName)
	locations, _ := s.store.LoadLocations(bookName)

	// 简单的名称匹配（可以后续用 LLM 增强）
	var charNames, itemNames, locNames []string

	for _, char := range characters {
		if strings.Contains(content, char.Name) {
			charNames = append(charNames, char.Name)
			// 更新角色出场章节
			s.addCharacterAppearChapter(bookName, char.Name, chapterID)
		}
	}

	for _, item := range items {
		if strings.Contains(content, item.Name) {
			itemNames = append(itemNames, item.Name)
			// 更新物品出场章节
			s.addItemAppearChapter(bookName, item.Name, chapterID)
		}
	}

	for _, loc := range locations {
		if strings.Contains(content, loc.Name) {
			locNames = append(locNames, loc.Name)
		}
	}

	// 更新章节结构
	chapters, _ := s.store.LoadChapters(bookName)
	for _, ch := range chapters {
		if ch.ID == chapterID {
			ch.Characters = charNames
			ch.Items = itemNames
			ch.Locations = locNames
			break
		}
	}

	return s.store.SaveChapters(bookName, chapters)
}

// addCharacterAppearChapter 添加角色出场章节
func (s *OrchestrationService) addCharacterAppearChapter(bookName, charName string, chapterID int) {
	characters, _ := s.store.LoadCharacters(bookName)
	for _, char := range characters {
		if char.Name == charName {
			// 检查是否已存在
			for _, chID := range char.AppearChapters {
				if chID == chapterID {
					return
				}
			}
			char.AppearChapters = append(char.AppearChapters, chapterID)
			break
		}
	}
	s.store.SaveCharacters(bookName, characters)
}

// addItemAppearChapter 添加物品出场章节
func (s *OrchestrationService) addItemAppearChapter(bookName, itemName string, chapterID int) {
	items, _ := s.store.LoadItems(bookName)
	for _, item := range items {
		if item.Name == itemName {
			// 检查是否已存在
			for _, chID := range item.AppearChapters {
				if chID == chapterID {
					return
				}
			}
			item.AppearChapters = append(item.AppearChapters, chapterID)
			break
		}
	}
	s.store.SaveItems(bookName, items)
}

// AfterChapterSaved 章节保存后的编排（用户手动编辑后）
func (s *OrchestrationService) AfterChapterSaved(ctx context.Context, bookName string, chapterID int) error {
	// 只更新实体索引，不做完整的编排
	return s.updateChapterEntityIndex(bookName, chapterID)
}

// RunConsistencyCheck 运行一致性检查
func (s *OrchestrationService) RunConsistencyCheck(ctx context.Context, bookName string, chapterID int) (*engine.ConsistencyReport, error) {
	consEngine := engine.NewConsistencyEngine(s.llmClient, s.store)
	return consEngine.CheckChapter(ctx, bookName, chapterID)
}

// CheckThreadWarnings 检查线程掉线预警
func (s *OrchestrationService) CheckThreadWarnings(bookName string, currentChapter int) []string {
	threadMgr := engine.NewNarrativeThreadManager(s.store)
	return threadMgr.CheckThreadWarnings(bookName, currentChapter)
}

// GetForeshadowWarnings 获取伏笔预警
func (s *OrchestrationService) GetForeshadowWarnings(bookName string, currentChapter int) []*model.ForeshadowWarning {
	foreshadows, _ := s.store.LoadForeshadows(bookName)
	var warnings []*model.ForeshadowWarning

	const warningThreshold = 10 // 超过10章未回收预警

	for _, fs := range foreshadows {
		if fs.Status != model.ForeshadowActive {
			continue
		}

		gap := currentChapter - fs.SourceChapter
		if gap > warningThreshold {
			warnings = append(warnings, &model.ForeshadowWarning{
				Foreshadow:     fs,
				WarningType:    "overdue",
				WarningMessage: fmt.Sprintf("伏笔已埋设 %d 章未回收", gap),
				ChaptersSince:  gap,
			})
		}
	}

	return warnings
}