package service

import (
	"fmt"
	"time"

	"ai-writer/internal/model"
	"ai-writer/internal/store"
)

// BookAnalysisService 书籍图谱分析服务
// 用于分析书籍内部的伏笔、因果链、叙事线程、情感弧线、时间线等图谱数据
type BookAnalysisService struct {
	store *store.JSONStore
}

// NewBookAnalysisService 创建书籍图谱分析服务
func NewBookAnalysisService(jsonStore *store.JSONStore) *BookAnalysisService {
	return &BookAnalysisService{store: jsonStore}
}

// RunAnalysis 运行完整分析
func (s *BookAnalysisService) RunAnalysis(bookName string, chapterID int, analysisType string) (*model.AnalysisReport, error) {
	report := &model.AnalysisReport{
		ID:        generateID(),
		BookID:    bookName,
		ChapterID: chapterID,
		Type:      analysisType,
		CreatedAt: time.Now(),
	}

	// 运行各类分析
	report.ForeshadowAnalysis = s.analyzeForeshadows(bookName, chapterID)
	report.CausalAnalysis = s.analyzeCausalChains(bookName)
	report.ThreadAnalysis = s.analyzeThreads(bookName, chapterID)
	report.EmotionAnalysis = s.analyzeEmotions(bookName)
	report.TimelineAnalysis = s.analyzeTimeline(bookName)

	// 保存报告
	if err := s.store.AppendAnalysisReport(bookName, report); err != nil {
		return report, err
	}

	return report, nil
}

// analyzeForeshadows 分析伏笔
func (s *BookAnalysisService) analyzeForeshadows(bookName string, currentChapter int) model.ForeshadowAnalysis {
	analysis := model.ForeshadowAnalysis{Score: 100}

	foreshadows, err := s.store.LoadForeshadows(bookName)
	if err != nil {
		return analysis
	}

	for _, fs := range foreshadows {
		// 检查超时（超过5章未回收）
		if fs.Status == model.ForeshadowActive {
			gap := currentChapter - fs.SourceChapter
			if gap > 5 {
				analysis.Warnings = append(analysis.Warnings, model.ForeshadowWarning{
					Foreshadow:     fs,
					WarningType:    "timeout",
					WarningMessage: fmt.Sprintf("伏笔已过 %d 章未回收", gap),
					ChaptersSince:  gap,
				})
				analysis.Score -= 10
			}

			// 检查目标章节错过
			if fs.TargetChapter > 0 && currentChapter > fs.TargetChapter {
				analysis.Warnings = append(analysis.Warnings, model.ForeshadowWarning{
					Foreshadow:     fs,
					WarningType:    "target_missed",
					WarningMessage: fmt.Sprintf("预期在第%d章回收，已错过", fs.TargetChapter),
					ChaptersSince:  currentChapter - fs.TargetChapter,
				})
				analysis.Score -= 15
			}
		}

		// 检查过期预警状态
		if fs.Status == model.ForeshadowExpired {
			analysis.Suggestions = append(analysis.Suggestions,
				fmt.Sprintf("伏笔[%s]已过期，建议尽快回收或标记放弃", fs.Content))
		}
	}

	// 确保分数不低于0
	if analysis.Score < 0 {
		analysis.Score = 0
	}

	return analysis
}

// analyzeCausalChains 分析因果链
func (s *BookAnalysisService) analyzeCausalChains(bookName string) model.CausalAnalysis {
	analysis := model.CausalAnalysis{Score: 100}

	events, err := s.store.LoadCausalChains(bookName)
	if err != nil {
		return analysis
	}

	// 检查断裂的因果链（有因无果、有果无因）
	for _, event := range events {
		// 有因无果：原因明确但后果未定义
		if event.Cause != "" && event.Effect == "" && event.Status == model.CausalActive {
			analysis.BrokenChains = append(analysis.BrokenChains, model.BrokenChain{
				EventID:   event.ID,
				EventName: event.Event,
				Issue:     "有因无果",
			})
			analysis.Score -= 5
		}

		// 有果无因：后果明确但原因未定义
		if event.Cause == "" && event.Effect != "" && event.Status == model.CausalActive {
			analysis.BrokenChains = append(analysis.BrokenChains, model.BrokenChain{
				EventID:   event.ID,
				EventName: event.Event,
				Issue:     "有果无因",
			})
			analysis.Score -= 5
		}

		// 检查孤立事件（无角色关联）
		if len(event.Characters) == 0 && event.Status == model.CausalActive {
			analysis.OrphanEvents = append(analysis.OrphanEvents, model.OrphanEvent{
				EventID:   event.ID,
				EventName: event.Event,
			})
			analysis.Score -= 3
		}
	}

	// 检查循环依赖
	circularDeps := s.detectCircularDependencies(events)
	for _, dep := range circularDeps {
		analysis.CircularDeps = append(analysis.CircularDeps, dep)
		analysis.Score -= 10
	}

	// 确保分数不低于0
	if analysis.Score < 0 {
		analysis.Score = 0
	}

	return analysis
}

// detectCircularDependencies 检测循环依赖
func (s *BookAnalysisService) detectCircularDependencies(events []*model.CausalEvent) []model.CircularDep {
	var circularDeps []model.CircularDep

	// 构建依赖图
	// 通过事件间的因果关系检测循环
	// 这里简化处理：检查是否有事件的 Effect 指向另一个事件的 Cause
	eventMap := make(map[string]*model.CausalEvent)
	for _, event := range events {
		eventMap[event.ID] = event
	}

	// 检查简单的循环（A->B->A）
	// 由于没有显式的链接数据，这里只能通过内容匹配进行简单检测
	// TODO: 当 CausalLink 数据可用时，实现更精确的循环检测

	return circularDeps
}

// analyzeThreads 分析叙事线程
func (s *BookAnalysisService) analyzeThreads(bookName string, currentChapter int) model.ThreadAnalysis {
	analysis := model.ThreadAnalysis{Score: 100}

	threads, err := s.store.LoadThreads(bookName)
	if err != nil {
		return analysis
	}

	for _, thread := range threads {
		// 检查遗忘的线程（超过3章未活跃）
		if thread.Status == model.ThreadActive {
			gap := currentChapter - thread.LastActiveChapter
			if gap > 3 {
				analysis.ForgottenThreads = append(analysis.ForgottenThreads, model.ForgottenThread{
					ThreadID:        thread.ID,
					ThreadName:      thread.Name,
					LastActive:      thread.LastActiveChapter,
					CurrentChapter:  currentChapter,
					ChaptersSkipped: gap,
				})
				analysis.Score -= 8
			}
		}

		// 检查主线的节奏问题
		if thread.Type == model.ThreadMain {
			// 主线应该保持较高活跃度
			if thread.Status == model.ThreadPaused {
				analysis.PacingIssues = append(analysis.PacingIssues, model.PacingIssue{
					ThreadID:   thread.ID,
					ThreadName: thread.Name,
					Issue:      "主线处于暂停状态，影响叙事连贯性",
				})
				analysis.Score -= 10
			}
		}

		// 检查预期的结束章节
		if thread.EndChapter > 0 && currentChapter > thread.EndChapter && thread.Status != model.ThreadComplete {
			analysis.PacingIssues = append(analysis.PacingIssues, model.PacingIssue{
				ThreadID:   thread.ID,
				ThreadName: thread.Name,
				Issue:      fmt.Sprintf("预期在第%d章完成，但尚未结束", thread.EndChapter),
			})
			analysis.Score -= 5
		}
	}

	// 确保分数不低于0
	if analysis.Score < 0 {
		analysis.Score = 0
	}

	return analysis
}

// analyzeEmotions 分析情感弧线
func (s *BookAnalysisService) analyzeEmotions(bookName string) model.EmotionAnalysis {
	analysis := model.EmotionAnalysis{
		Score:        100,
		WeavingScore: 100,
	}

	characters, err := s.store.LoadCharacters(bookName)
	if err != nil {
		return analysis
	}

	for _, char := range characters {
		if len(char.EmotionalArc) < 2 {
			continue
		}

		// 检查情感跳跃
		for i := 1; i < len(char.EmotionalArc); i++ {
			prev := char.EmotionalArc[i-1]
			curr := char.EmotionalArc[i]

			// 情感强度跳跃超过3可能存在问题
			intensityJump := curr.Intensity - prev.Intensity
			if intensityJump > 3 || intensityJump < -3 {
				analysis.Inconsistencies = append(analysis.Inconsistencies, model.EmotionInconsistency{
					Character:     char.Name,
					ChapterID:     curr.ChapterID,
					FromEmotion:   prev.Emotion,
					ToEmotion:     curr.Emotion,
					IntensityJump: intensityJump,
				})
				analysis.Score -= 5
			}
		}

		// 检查情感节奏问题
		// 同一角色连续多个章节情感相同可能导致单调
		if len(char.EmotionalArc) >= 3 {
			sameEmotionCount := 0
			lastEmotion := ""
			for _, ep := range char.EmotionalArc {
				if ep.Emotion == lastEmotion {
					sameEmotionCount++
				} else {
					sameEmotionCount = 0
					lastEmotion = ep.Emotion
				}

				if sameEmotionCount >= 3 {
					analysis.PacingIssues = append(analysis.PacingIssues, model.EmotionPacingIssue{
						Character: char.Name,
						Issue:     fmt.Sprintf("情感连续%d章保持'%s'，可能过于单调", sameEmotionCount+1, ep.Emotion),
					})
					analysis.WeavingScore -= 3
				}
			}
		}
	}

	// 确保分数不低于0
	if analysis.Score < 0 {
		analysis.Score = 0
	}
	if analysis.WeavingScore < 0 {
		analysis.WeavingScore = 0
	}

	return analysis
}

// analyzeTimeline 分析时间线
func (s *BookAnalysisService) analyzeTimeline(bookName string) model.TimelineAnalysis {
	analysis := model.TimelineAnalysis{Score: 100}

	timeline, err := s.store.LoadTimeline(bookName)
	if err != nil {
		return analysis
	}

	// 检查时间跳跃
	for i := 1; i < len(timeline); i++ {
		prev := timeline[i-1]
		curr := timeline[i]

		// 检查章节顺序是否合理
		if curr.ChapterID < prev.ChapterID {
			analysis.Inconsistencies = append(analysis.Inconsistencies, model.TimelineInconsistency{
				ChapterID:  curr.ChapterID,
				EventOrder: fmt.Sprintf("第%d章出现在第%d章之后", curr.ChapterID, prev.ChapterID),
				Issue:      "时间线顺序异常",
			})
			analysis.Score -= 10
		}

		// 记录时间跳跃（相邻章节时间跨度较大）
		// 这里通过章节间隔来推断时间跳跃
		chapterGap := curr.ChapterID - prev.ChapterID
		if chapterGap > 1 {
			analysis.TimeJumps = append(analysis.TimeJumps, model.TimeJump{
				FromChapter: prev.ChapterID,
				ToChapter:   curr.ChapterID,
				FromTime:    prev.TimeLabel,
				ToTime:      curr.TimeLabel,
				Duration:    fmt.Sprintf("%d章间隔", chapterGap),
			})
		}
	}

	// 检查同一章节的事件重叠
	chapterEvents := make(map[int][]model.TimelineEvent)
	for _, te := range timeline {
		chapterEvents[te.ChapterID] = append(chapterEvents[te.ChapterID], te)
	}

	for chapterID, events := range chapterEvents {
		if len(events) > 1 {
			// 同一章节有多个时间线事件，检查是否可能重叠
			var eventNames []string
			for _, e := range events {
				eventNames = append(eventNames, e.TimeLabel)
			}

			// 如果同一章节的时间标签不同，可能存在重叠
			uniqueLabels := make(map[string]bool)
			for _, e := range events {
				uniqueLabels[e.TimeLabel] = true
			}

			if len(uniqueLabels) > 1 {
				analysis.Overlaps = append(analysis.Overlaps, model.TimeOverlap{
					ChapterID: chapterID,
					TimeLabel: fmt.Sprintf("多个时间标签: %v", eventNames),
					Events:    eventNames,
				})
				analysis.Score -= 3
			}
		}
	}

	// 确保分数不低于0
	if analysis.Score < 0 {
		analysis.Score = 0
	}

	return analysis
}

// GetAnalysisHistory 获取分析历史
func (s *BookAnalysisService) GetAnalysisHistory(bookName string) ([]*model.AnalysisReport, error) {
	return s.store.LoadAnalysisReports(bookName)
}

// GetLatestReport 获取最新分析报告
func (s *BookAnalysisService) GetLatestReport(bookName string) (*model.AnalysisReport, error) {
	reports, err := s.store.LoadAnalysisReports(bookName)
	if err != nil {
		return nil, err
	}

	if len(reports) == 0 {
		return nil, nil
	}

	// 返回最新的报告
	return reports[len(reports)-1], nil
}