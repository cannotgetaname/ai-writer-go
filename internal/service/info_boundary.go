package service

import (
	"context"

	"ai-writer/internal/engine"
	"ai-writer/internal/llm"
	"ai-writer/internal/model"
	"ai-writer/internal/store"
)

// InfoBoundaryService 信息边界服务
type InfoBoundaryService struct {
	manager *engine.InfoBoundaryManager
}

// NewInfoBoundaryService 创建信息边界服务
func NewInfoBoundaryService(llmClient llm.Client, jsonStore *store.JSONStore) *InfoBoundaryService {
	return &InfoBoundaryService{
		manager: engine.NewInfoBoundaryManager(llmClient, jsonStore),
	}
}

// CheckLeak 检测信息越界
func (s *InfoBoundaryService) CheckLeak(ctx context.Context, bookName string, chapterID int) (map[string]interface{}, error) {
	leaks, err := s.manager.CheckInfoLeak(ctx, bookName, chapterID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"book_name":  bookName,
		"chapter_id": chapterID,
		"leaks":      leaks,
		"count":      len(leaks),
	}, nil
}

// ExtractInfo 提取角色信息
func (s *InfoBoundaryService) ExtractInfo(ctx context.Context, bookName string, chapterID int) (map[string][]model.KnownInfo, error) {
	return s.manager.ExtractAndSave(ctx, bookName, chapterID)
}

// GetCharacterPOV 获取角色视角
func (s *InfoBoundaryService) GetCharacterPOV(bookName string, charName string, chapterID int) (string, error) {
	characters, err := s.manager.Store.LoadCharacters(bookName)
	if err != nil {
		return "", err
	}

	for _, char := range characters {
		if char.Name == charName {
			return s.manager.GetCharacterPOV(char, chapterID), nil
		}
	}

	return "", nil
}