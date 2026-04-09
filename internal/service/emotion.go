package service

import (
	"context"

	"ai-writer/internal/engine"
	"ai-writer/internal/llm"
	"ai-writer/internal/model"
	"ai-writer/internal/store"
)

// EmotionService 情感弧线服务
type EmotionService struct {
	tracker *engine.EmotionalArcTracker
	store   *store.JSONStore
}

// NewEmotionService 创建情感弧线服务
func NewEmotionService(llmClient llm.Client, jsonStore *store.JSONStore) *EmotionService {
	return &EmotionService{
		tracker: engine.NewEmotionalArcTracker(llmClient, jsonStore),
		store:   jsonStore,
	}
}

// TrackChapter 追踪章节情感
func (s *EmotionService) TrackChapter(ctx context.Context, bookName string, chapterID int) (map[string]model.EmotionPoint, error) {
	return s.tracker.TrackAndSave(ctx, bookName, chapterID)
}

// GetCharacterArc 获取角色情感弧线
func (s *EmotionService) GetCharacterArc(bookName string, charName string) ([]model.EmotionPoint, error) {
	return s.tracker.GetArcData(bookName, charName)
}

// GetEmotionSummary 获取情感摘要
func (s *EmotionService) GetEmotionSummary(bookName string, charName string) string {
	return s.tracker.GetEmotionSummary(bookName, charName)
}

// DetectArcComplete 检测弧线是否完成
func (s *EmotionService) DetectArcComplete(bookName string, charName string) (bool, string) {
	return s.tracker.DetectArcComplete(bookName, charName)
}