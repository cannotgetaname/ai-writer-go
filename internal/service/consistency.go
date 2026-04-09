package service

import (
	"context"

	"ai-writer/internal/engine"
	"ai-writer/internal/llm"
	"ai-writer/internal/store"
)

// ConsistencyService 一致性检查服务
type ConsistencyService struct {
	engine *engine.ConsistencyEngine
}

// NewConsistencyService 创建一致性检查服务
func NewConsistencyService(llmClient llm.Client, jsonStore *store.JSONStore) *ConsistencyService {
	return &ConsistencyService{
		engine: engine.NewConsistencyEngine(llmClient, jsonStore),
	}
}

// CheckChapter 检查章节一致性
func (s *ConsistencyService) CheckChapter(ctx context.Context, bookName string, chapterID int) (*engine.ConsistencyReport, error) {
	return s.engine.CheckChapter(ctx, bookName, chapterID)
}

// CheckRange 检查章节范围的一致性
func (s *ConsistencyService) CheckRange(ctx context.Context, bookName string, fromChapter, toChapter int) (*engine.ConsistencyReport, error) {
	return s.engine.CheckRange(ctx, bookName, fromChapter, toChapter)
}