package engine

import (
	"fmt"
	"strings"

	"ai-writer/internal/model"
	"ai-writer/internal/store"
)

// NarrativeThreadManager 多线叙事管理器
type NarrativeThreadManager struct {
	store *store.JSONStore
}

// NewNarrativeThreadManager 创建多线叙事管理器
func NewNarrativeThreadManager(store *store.JSONStore) *NarrativeThreadManager {
	return &NarrativeThreadManager{
		store: store,
	}
}

// CreateThread 创建叙事线程
func (m *NarrativeThreadManager) CreateThread(bookName string, name string, threadType model.ThreadType) (*model.NarrativeThread, error) {
	threads, err := m.store.LoadThreads(bookName)
	if err != nil {
		threads = []*model.NarrativeThread{}
	}

	thread := &model.NarrativeThread{
		ID:       fmt.Sprintf("thread_%d", len(threads)+1),
		BookID:   bookName,
		Name:     name,
		Type:     threadType,
		Status:   model.ThreadActive,
		Chapters: []int{},
	}

	threads = append(threads, thread)
	if err := m.store.SaveThreads(bookName, threads); err != nil {
		return nil, err
	}

	return thread, nil
}

// AssignChapterToThread 将章节分配到线程
func (m *NarrativeThreadManager) AssignChapterToThread(bookName string, threadID string, chapterID int) error {
	threads, err := m.store.LoadThreads(bookName)
	if err != nil {
		return err
	}

	for _, thread := range threads {
		if thread.ID == threadID {
			thread.Chapters = append(thread.Chapters, chapterID)
			thread.LastActiveChapter = chapterID
			return m.store.SaveThreads(bookName, threads)
		}
	}

	return fmt.Errorf("线程 %s 不存在", threadID)
}

// GetThreadForChapter 获取章节所属线程
func (m *NarrativeThreadManager) GetThreadForChapter(bookName string, chapterID int) (*model.NarrativeThread, error) {
	threads, err := m.store.LoadThreads(bookName)
	if err != nil {
		return nil, err
	}

	for _, thread := range threads {
		for _, chID := range thread.Chapters {
			if chID == chapterID {
				return thread, nil
			}
		}
	}

	return nil, nil
}

// CheckThreadWarnings 检查线程掉线预警
func (m *NarrativeThreadManager) CheckThreadWarnings(bookName string, currentChapter int) []string {
	threads, err := m.store.LoadThreads(bookName)
	if err != nil {
		return []string{"无法加载线程数据"}
	}

	var warnings []string
	const warningThreshold = 5 // 超过5章未活跃则预警

	for _, thread := range threads {
		if thread.Status != model.ThreadActive {
			continue
		}

		gap := currentChapter - thread.LastActiveChapter
		if gap > warningThreshold {
			warnings = append(warnings, fmt.Sprintf(
				"⚠️ 线程【%s】已掉线 %d 章（最后活跃: 第%d章）",
				thread.Name, gap, thread.LastActiveChapter,
			))
		}
	}

	return warnings
}

// GetThreadStats 获取线程统计
func (m *NarrativeThreadManager) GetThreadStats(bookName string) map[string]interface{} {
	threads, err := m.store.LoadThreads(bookName)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	stats := map[string]interface{}{
		"total":    len(threads),
		"active":   0,
		"paused":   0,
		"complete": 0,
		"by_type":  map[string]int{},
	}

	for _, thread := range threads {
		switch thread.Status {
		case model.ThreadActive:
			stats["active"] = stats["active"].(int) + 1
		case model.ThreadPaused:
			stats["paused"] = stats["paused"].(int) + 1
		case model.ThreadComplete:
			stats["complete"] = stats["complete"].(int) + 1
		}

		byType := stats["by_type"].(map[string]int)
		byType[string(thread.Type)]++
	}

	return stats
}

// GetThreadTimeline 获取线程时间线视图
func (m *NarrativeThreadManager) GetThreadTimeline(bookName string) string {
	threads, err := m.store.LoadThreads(bookName)
	if err != nil {
		return "无法加载线程数据"
	}

	if len(threads) == 0 {
		return "暂无叙事线程"
	}

	chapters, _ := m.store.LoadChapters(bookName)
	maxChapter := 0
	for _, ch := range chapters {
		if ch.ID > maxChapter {
			maxChapter = ch.ID
		}
	}

	var timeline strings.Builder
	timeline.WriteString("叙事线程时间线:\n")
	timeline.WriteString("═══════════════════════════════════════\n")

	for _, thread := range threads {
		timeline.WriteString(fmt.Sprintf("\n【%s】(%s) - %s\n", thread.Name, thread.Type, thread.Status))

		// 绘制简单时间线
		line := make([]rune, maxChapter+1)
		for i := range line {
			line[i] = '─'
		}
		for _, chID := range thread.Chapters {
			if chID <= maxChapter {
				line[chID] = '●'
			}
		}

		timeline.WriteString(fmt.Sprintf("  %s\n", string(line[1:])))
		timeline.WriteString(fmt.Sprintf("  最后活跃: 第%d章\n", thread.LastActiveChapter))
	}

	return timeline.String()
}