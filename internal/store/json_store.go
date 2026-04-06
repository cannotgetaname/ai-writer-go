package store

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"

	"ai-writer/internal/model"
)

// validName 验证名称是否合法（防止路径注入）
// 只允许字母、数字、中文、下划线、横杠，禁止路径遍历字符
func validName(name string) bool {
	if name == "" || len(name) > 100 {
		return false
	}
	// 禁止路径遍历和分隔符
	if strings.Contains(name, "..") || strings.ContainsAny(name, "/\\") {
		return false
	}
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			return false
		}
	}
	return true
}

// JSONStore JSON 文件存储
type JSONStore struct {
	basePath string
	mu       sync.RWMutex
}

// NewJSONStore 创建存储实例
func NewJSONStore(basePath string) *JSONStore {
	return &JSONStore{
		basePath: basePath,
	}
}

// 确保目录存在
func (s *JSONStore) ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// getBookPath 获取书籍路径（带验证）
func (s *JSONStore) getBookPath(bookName string) (string, error) {
	if !validName(bookName) {
		return "", fmt.Errorf("书名不合法: %s", bookName)
	}
	return filepath.Join(s.basePath, "projects", bookName), nil
}

// 项目目录结构:
// data/projects/{book_name}/
// ├── metadata.json      # 书籍元数据
// ├── volumes.json       # 分卷
// ├── structure.json     # 章节结构
// ├── characters.json    # 人物
// ├── items.json         # 物品
// ├── locations.json     # 地点
// ├── worldview.json     # 世界观
// ├── foreshadows.json   # 伏笔
// ├── causal_chains.json # 因果链
// ├── threads.json       # 叙事线程
// └── chapters/
//     ├── 1.txt          # 章节内容
//     └── 1_paragraphs.json

// ==================== 书籍管理 ====================

// ListBooks 列出所有书籍
func (s *JSONStore) ListBooks() ([]*model.BookMeta, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	projectsPath := filepath.Join(s.basePath, "projects")
	entries, err := os.ReadDir(projectsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*model.BookMeta{}, nil
		}
		return nil, err
	}

	var books []*model.BookMeta
	for _, entry := range entries {
		if entry.IsDir() {
			meta, err := s.loadBookMeta(entry.Name())
			if err != nil {
				continue // 跳过无效项目
			}
			books = append(books, meta)
		}
	}

	return books, nil
}

// CreateBook 创建新书
func (s *JSONStore) CreateBook(name string) (*model.BookMeta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 验证书名防止路径注入
	if !validName(name) {
		return nil, fmt.Errorf("书名不合法: %s (只允许字母、数字、中文、下划线、横杠，长度1-100)", name)
	}

	bookPath := filepath.Join(s.basePath, "projects", name)

	// 检查是否存在
	if _, err := os.Stat(bookPath); err == nil {
		return nil, fmt.Errorf("书籍已存在: %s", name)
	}

	// 创建目录结构
	if err := s.ensureDir(bookPath); err != nil {
		return nil, err
	}
	if err := s.ensureDir(filepath.Join(bookPath, "chapters")); err != nil {
		return nil, err
	}

	// 创建元数据
	now := time.Now()
	meta := &model.BookMeta{
		ID:        name,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 初始化默认数据
	if err := s.initBookData(bookPath, name); err != nil {
		return nil, err
	}

	return meta, s.saveJSON(filepath.Join(bookPath, "metadata.json"), meta)
}

// initBookData 初始化书籍默认数据
func (s *JSONStore) initBookData(bookPath, name string) error {
	now := time.Now()

	// 默认分卷
	volumes := []*model.Volume{
		{ID: "vol_1", BookID: name, Title: "正文卷", Order: 1},
	}
	if err := s.saveJSON(filepath.Join(bookPath, "volumes.json"), volumes); err != nil {
		return err
	}

	// 默认章节结构
	chapters := []*model.Chapter{
		{
			ID:        1,
			BookID:    name,
			VolumeID:  "vol_1",
			Title:     "第一章",
			Outline:   "故事开始...",
			TimeInfo:  model.TimeInfo{Label: "故事开始", Duration: "0", Events: []string{}},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	if err := s.saveJSON(filepath.Join(bookPath, "structure.json"), chapters); err != nil {
		return err
	}

	// 空数据文件
	empty := []interface{}{}
	if err := s.saveJSON(filepath.Join(bookPath, "characters.json"), empty); err != nil {
		return err
	}
	if err := s.saveJSON(filepath.Join(bookPath, "items.json"), empty); err != nil {
		return err
	}
	if err := s.saveJSON(filepath.Join(bookPath, "locations.json"), empty); err != nil {
		return err
	}

	// 默认世界观
	worldview := &model.WorldView{
		BookID: name,
		BasicInfo: model.WorldViewBasic{
			Genre:     "",
			Era:       "",
			TechLevel: "",
		},
		CoreSettings: model.WorldViewCore{
			PowerSystem:     "",
			SocialStructure: "",
			SpecialRules:    "",
		},
		KeyElements: model.WorldViewElements{
			ImportantItems: "",
			Organizations:  "",
			Locations:      "",
		},
		Background: model.WorldViewBackground{
			History:      "",
			MainConflict: "",
			Development:  "",
		},
	}
	if err := s.saveJSON(filepath.Join(bookPath, "worldview.json"), worldview); err != nil {
		return err
	}

	// 伏笔
	if err := s.saveJSON(filepath.Join(bookPath, "foreshadows.json"), empty); err != nil {
		return err
	}

	// 因果链
	if err := s.saveJSON(filepath.Join(bookPath, "causal_chains.json"), empty); err != nil {
		return err
	}

	// 叙事线程
	if err := s.saveJSON(filepath.Join(bookPath, "threads.json"), empty); err != nil {
		return err
	}

	return nil
}

// DeleteBook 删除书籍
func (s *JSONStore) DeleteBook(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 验证书名防止路径注入
	if !validName(name) {
		return fmt.Errorf("书名不合法: %s", name)
	}

	bookPath := filepath.Join(s.basePath, "projects", name)
	return os.RemoveAll(bookPath)
}

// RenameBook 重命名书籍
func (s *JSONStore) RenameBook(oldName, newName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 验证旧名称
	if !validName(oldName) {
		return fmt.Errorf("原书名不合法: %s", oldName)
	}

	// 验证新名称
	if !validName(newName) {
		return fmt.Errorf("新书名不合法: %s (只允许字母、数字、中文、下划线、横杠，长度1-100)", newName)
	}

	oldPath := filepath.Join(s.basePath, "projects", oldName)
	newPath := filepath.Join(s.basePath, "projects", newName)

	// 检查原书籍是否存在
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return fmt.Errorf("书籍不存在: %s", oldName)
	}

	// 检查新名称是否已存在
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("书籍已存在: %s", newName)
	}

	// 重命名目录
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("重命名失败: %v", err)
	}

	// 更新元数据文件中的名称
	metaPath := filepath.Join(newPath, "metadata.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil // 元数据文件不存在不影响重命名
	}

	var meta model.BookMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil // 解析失败不影响重命名
	}

	meta.ID = newName
	meta.Name = newName
	meta.UpdatedAt = time.Now()

	return s.saveJSON(metaPath, &meta)
}

// loadBookMeta 加载书籍元数据
func (s *JSONStore) loadBookMeta(name string) (*model.BookMeta, error) {
	// 验证书名防止路径注入
	if !validName(name) {
		return nil, fmt.Errorf("书名不合法: %s", name)
	}

	path := filepath.Join(s.basePath, "projects", name, "metadata.json")

	data, err := os.ReadFile(path)
	if err != nil {
		// 如果 metadata.json 不存在，从目录名创建
		return &model.BookMeta{
			ID:   name,
			Name: name,
		}, nil
	}

	var meta model.BookMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

// ==================== 数据加载 ====================

// LoadBook 加载书籍完整信息
func (s *JSONStore) LoadBook(bookName string) (*model.Book, error) {
	// 验证书名防止路径注入
	if !validName(bookName) {
		return nil, fmt.Errorf("书名不合法: %s", bookName)
	}

	// 加载元数据
	meta, err := s.loadBookMeta(bookName)
	if err != nil {
		return nil, err
	}

	book := &model.Book{
		ID:        meta.ID,
		Name:      meta.Name,
		CreatedAt: meta.CreatedAt,
		UpdatedAt: meta.UpdatedAt,
	}

	// 加载关联数据（这些方法内部有自己的锁）
	book.Volumes, _ = s.LoadVolumes(bookName)
	book.Chapters, _ = s.LoadChapters(bookName)
	book.Characters, _ = s.LoadCharacters(bookName)
	book.Items, _ = s.LoadItems(bookName)
	book.Locations, _ = s.LoadLocations(bookName)

	return book, nil
}

// LoadVolumes 加载分卷
func (s *JSONStore) LoadVolumes(bookName string) ([]*model.Volume, error) {
	if !validName(bookName) {
		return nil, fmt.Errorf("书名不合法: %s", bookName)
	}
	path := filepath.Join(s.basePath, "projects", bookName, "volumes.json")
	var volumes []*model.Volume
	if err := s.loadJSON(path, &volumes); err != nil {
		return nil, err
	}
	return volumes, nil
}

// LoadChapters 加载章节结构
func (s *JSONStore) LoadChapters(bookName string) ([]*model.Chapter, error) {
	if !validName(bookName) {
		return nil, fmt.Errorf("书名不合法: %s", bookName)
	}
	path := filepath.Join(s.basePath, "projects", bookName, "structure.json")
	var chapters []*model.Chapter
	if err := s.loadJSON(path, &chapters); err != nil {
		return nil, err
	}
	return chapters, nil
}

// LoadCharacters 加载人物
func (s *JSONStore) LoadCharacters(bookName string) ([]*model.Character, error) {
	if !validName(bookName) {
		return nil, fmt.Errorf("书名不合法: %s", bookName)
	}
	path := filepath.Join(s.basePath, "projects", bookName, "characters.json")
	var characters []*model.Character
	if err := s.loadJSON(path, &characters); err != nil {
		return nil, err
	}
	return characters, nil
}

// LoadItems 加载物品
func (s *JSONStore) LoadItems(bookName string) ([]*model.Item, error) {
	if !validName(bookName) {
		return nil, fmt.Errorf("书名不合法: %s", bookName)
	}
	path := filepath.Join(s.basePath, "projects", bookName, "items.json")
	var items []*model.Item
	if err := s.loadJSON(path, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// LoadLocations 加载地点
func (s *JSONStore) LoadLocations(bookName string) ([]*model.Location, error) {
	if !validName(bookName) {
		return nil, fmt.Errorf("书名不合法: %s", bookName)
	}
	path := filepath.Join(s.basePath, "projects", bookName, "locations.json")
	var locations []*model.Location
	if err := s.loadJSON(path, &locations); err != nil {
		return nil, err
	}
	return locations, nil
}

// LoadWorldView 加载世界观
func (s *JSONStore) LoadWorldView(bookName string) (*model.WorldView, error) {
	if !validName(bookName) {
		return nil, fmt.Errorf("书名不合法: %s", bookName)
	}
	path := filepath.Join(s.basePath, "projects", bookName, "worldview.json")
	var worldview model.WorldView
	if err := s.loadJSON(path, &worldview); err != nil {
		return nil, err
	}
	return &worldview, nil
}

// LoadForeshadows 加载伏笔
func (s *JSONStore) LoadForeshadows(bookName string) ([]*model.Foreshadow, error) {
	if !validName(bookName) {
		return nil, fmt.Errorf("书名不合法: %s", bookName)
	}
	path := filepath.Join(s.basePath, "projects", bookName, "foreshadows.json")
	var foreshadows []*model.Foreshadow
	if err := s.loadJSON(path, &foreshadows); err != nil {
		return nil, err
	}
	return foreshadows, nil
}

// LoadCausalChains 加载因果链
func (s *JSONStore) LoadCausalChains(bookName string) ([]*model.CausalEvent, error) {
	if !validName(bookName) {
		return nil, fmt.Errorf("书名不合法: %s", bookName)
	}
	path := filepath.Join(s.basePath, "projects", bookName, "causal_chains.json")
	var events []*model.CausalEvent
	if err := s.loadJSON(path, &events); err != nil {
		return nil, err
	}
	return events, nil
}

// LoadThreads 加载叙事线程
func (s *JSONStore) LoadThreads(bookName string) ([]*model.NarrativeThread, error) {
	if !validName(bookName) {
		return nil, fmt.Errorf("书名不合法: %s", bookName)
	}
	path := filepath.Join(s.basePath, "projects", bookName, "threads.json")
	var threads []*model.NarrativeThread
	if err := s.loadJSON(path, &threads); err != nil {
		return nil, err
	}
	return threads, nil
}

// LoadTimeline 加载时间线事件
func (s *JSONStore) LoadTimeline(bookName string) ([]model.TimelineEvent, error) {
	if !validName(bookName) {
		return nil, fmt.Errorf("书名不合法: %s", bookName)
	}
	path := filepath.Join(s.basePath, "projects", bookName, "timeline.json")
	var timeline []model.TimelineEvent
	if err := s.loadJSON(path, &timeline); err != nil {
		return nil, err
	}
	return timeline, nil
}

// ==================== 章节内容（段落存储） ====================

// LoadChapterParagraphs 加载章节段落
func (s *JSONStore) LoadChapterParagraphs(bookName string, chapterID int) (*model.ChapterParagraphs, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 验证书名防止路径注入
	if !validName(bookName) {
		return nil, fmt.Errorf("书名不合法: %s", bookName)
	}

	path := filepath.Join(s.basePath, "projects", bookName, "chapters", fmt.Sprintf("%d.json", chapterID))

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// 返回空的段落结构
			return &model.ChapterParagraphs{
				ChapterID:  chapterID,
				Paragraphs: []model.Paragraph{},
				Metadata: model.ParagraphMetadata{
					ParagraphCount: 0,
					TotalWords:     0,
					UpdatedAt:      time.Now(),
				},
			}, nil
		}
		return nil, err
	}

	var paragraphs model.ChapterParagraphs
	if err := json.Unmarshal(data, &paragraphs); err != nil {
		return nil, err
	}

	return &paragraphs, nil
}

// SaveChapterParagraphs 保存章节段落
func (s *JSONStore) SaveChapterParagraphs(bookName string, paragraphs *model.ChapterParagraphs) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 验证书名防止路径注入
	if !validName(bookName) {
		return fmt.Errorf("书名不合法: %s", bookName)
	}

	chaptersPath := filepath.Join(s.basePath, "projects", bookName, "chapters")
	if err := s.ensureDir(chaptersPath); err != nil {
		return err
	}

	// 更新元数据
	paragraphs.Metadata.UpdatedAt = time.Now()
	paragraphs.Metadata.ParagraphCount = len(paragraphs.Paragraphs)

	totalWords := 0
	for _, p := range paragraphs.Paragraphs {
		totalWords += p.WordCount
	}
	paragraphs.Metadata.TotalWords = totalWords

	path := filepath.Join(chaptersPath, fmt.Sprintf("%d.json", paragraphs.ChapterID))
	return s.saveJSON(path, paragraphs)
}

// DeleteChapterParagraphs 删除章节段落文件
func (s *JSONStore) DeleteChapterParagraphs(bookName string, chapterID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.basePath, "projects", bookName, "chapters", fmt.Sprintf("%d.json", chapterID))

	// 如果文件不存在，直接返回成功
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	return os.Remove(path)
}

// AddParagraph 添加段落到章节
func (s *JSONStore) AddParagraph(bookName string, chapterID int, paragraph *model.Paragraph) error {
	paragraphs, err := s.LoadChapterParagraphs(bookName, chapterID)
	if err != nil {
		return err
	}

	paragraphs.Paragraphs = append(paragraphs.Paragraphs, *paragraph)

	return s.SaveChapterParagraphs(bookName, paragraphs)
}

// UpdateParagraph 更新段落
func (s *JSONStore) UpdateParagraph(bookName string, chapterID int, paragraphID string, text string) error {
	paragraphs, err := s.LoadChapterParagraphs(bookName, chapterID)
	if err != nil {
		return err
	}

	for i, p := range paragraphs.Paragraphs {
		if p.ID == paragraphID {
			paragraphs.Paragraphs[i].Text = text
			paragraphs.Paragraphs[i].WordCount = countChineseCharsInString(text)
			return s.SaveChapterParagraphs(bookName, paragraphs)
		}
	}

	return fmt.Errorf("段落 %s 不存在", paragraphID)
}

// DeleteParagraph 删除段落
func (s *JSONStore) DeleteParagraph(bookName string, chapterID int, paragraphID string) error {
	paragraphs, err := s.LoadChapterParagraphs(bookName, chapterID)
	if err != nil {
		return err
	}

	for i, p := range paragraphs.Paragraphs {
		if p.ID == paragraphID {
			paragraphs.Paragraphs = append(paragraphs.Paragraphs[:i], paragraphs.Paragraphs[i+1:]...)
			return s.SaveChapterParagraphs(bookName, paragraphs)
		}
	}

	return fmt.Errorf("段落 %s 不存在", paragraphID)
}

// MoveParagraph 移动段落位置
func (s *JSONStore) MoveParagraph(bookName string, chapterID int, paragraphID string, newPosition int) error {
	paragraphs, err := s.LoadChapterParagraphs(bookName, chapterID)
	if err != nil {
		return err
	}

	// 找到段落当前位置
	oldIndex := -1
	for i, p := range paragraphs.Paragraphs {
		if p.ID == paragraphID {
			oldIndex = i
			break
		}
	}

	if oldIndex == -1 {
		return fmt.Errorf("段落 %s 不存在", paragraphID)
	}

	// 边界检查
	if newPosition < 0 {
		newPosition = 0
	} else if newPosition >= len(paragraphs.Paragraphs) {
		newPosition = len(paragraphs.Paragraphs) - 1
	}

	// 移动段落
	paragraph := paragraphs.Paragraphs[oldIndex]
	paragraphs.Paragraphs = append(paragraphs.Paragraphs[:oldIndex], paragraphs.Paragraphs[oldIndex+1:]...)
	paragraphs.Paragraphs = append(paragraphs.Paragraphs[:newPosition], append([]model.Paragraph{paragraph}, paragraphs.Paragraphs[newPosition:]...)...)

	return s.SaveChapterParagraphs(bookName, paragraphs)
}

// LoadChapterContent 加载章节内容（兼容方法，从段落拼接）
func (s *JSONStore) LoadChapterContent(bookName string, chapterID int) (string, error) {
	paragraphs, err := s.LoadChapterParagraphs(bookName, chapterID)
	if err != nil {
		return "", err
	}

	var texts []string
	for _, p := range paragraphs.Paragraphs {
		texts = append(texts, p.Text)
	}

	return strings.Join(texts, "\n\n"), nil
}

// SaveChapterContent 保存章节内容（兼容方法，自动分段）
func (s *JSONStore) SaveChapterContent(bookName string, chapterID int, content string) error {
	// 按空行分段
	paragraphTexts := strings.Split(content, "\n\n")

	paragraphs := &model.ChapterParagraphs{
		ChapterID:  chapterID,
		Paragraphs: []model.Paragraph{},
	}

	for _, text := range paragraphTexts {
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		paragraphs.Paragraphs = append(paragraphs.Paragraphs, model.Paragraph{
			ID:        generateID(),
			Text:      text,
			WordCount: countChineseCharsInString(text),
		})
	}

	return s.SaveChapterParagraphs(bookName, paragraphs)
}

// countChineseCharsInString 统计中文字符数
func countChineseCharsInString(s string) int {
	count := 0
	for _, r := range s {
		if r >= 0x4e00 && r <= 0x9fff {
			count++
		}
	}
	return count
}

// generateID 生成唯一 ID
func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// ==================== 数据保存 ====================

// SaveVolumes 保存分卷
func (s *JSONStore) SaveVolumes(bookName string, volumes []*model.Volume) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !validName(bookName) {
		return fmt.Errorf("书名不合法: %s", bookName)
	}
	path := filepath.Join(s.basePath, "projects", bookName, "volumes.json")
	return s.saveJSON(path, volumes)
}

// SaveChapters 保存章节结构
func (s *JSONStore) SaveChapters(bookName string, chapters []*model.Chapter) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !validName(bookName) {
		return fmt.Errorf("书名不合法: %s", bookName)
	}
	path := filepath.Join(s.basePath, "projects", bookName, "structure.json")
	return s.saveJSON(path, chapters)
}

// SaveCharacters 保存人物
func (s *JSONStore) SaveCharacters(bookName string, characters []*model.Character) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !validName(bookName) {
		return fmt.Errorf("书名不合法: %s", bookName)
	}
	path := filepath.Join(s.basePath, "projects", bookName, "characters.json")
	return s.saveJSON(path, characters)
}

// SaveItems 保存物品
func (s *JSONStore) SaveItems(bookName string, items []*model.Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !validName(bookName) {
		return fmt.Errorf("书名不合法: %s", bookName)
	}
	path := filepath.Join(s.basePath, "projects", bookName, "items.json")
	return s.saveJSON(path, items)
}

// SaveLocations 保存地点
func (s *JSONStore) SaveLocations(bookName string, locations []*model.Location) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !validName(bookName) {
		return fmt.Errorf("书名不合法: %s", bookName)
	}
	path := filepath.Join(s.basePath, "projects", bookName, "locations.json")
	return s.saveJSON(path, locations)
}

// SaveWorldView 保存世界观
func (s *JSONStore) SaveWorldView(bookName string, worldview *model.WorldView) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !validName(bookName) {
		return fmt.Errorf("书名不合法: %s", bookName)
	}
	path := filepath.Join(s.basePath, "projects", bookName, "worldview.json")
	return s.saveJSON(path, worldview)
}

// SaveForeshadows 保存伏笔
func (s *JSONStore) SaveForeshadows(bookName string, foreshadows []*model.Foreshadow) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !validName(bookName) {
		return fmt.Errorf("书名不合法: %s", bookName)
	}
	path := filepath.Join(s.basePath, "projects", bookName, "foreshadows.json")
	return s.saveJSON(path, foreshadows)
}

// SaveCausalChains 保存因果链
func (s *JSONStore) SaveCausalChains(bookName string, events []*model.CausalEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !validName(bookName) {
		return fmt.Errorf("书名不合法: %s", bookName)
	}
	path := filepath.Join(s.basePath, "projects", bookName, "causal_chains.json")
	return s.saveJSON(path, events)
}

// SaveThreads 保存叙事线程
func (s *JSONStore) SaveThreads(bookName string, threads []*model.NarrativeThread) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !validName(bookName) {
		return fmt.Errorf("书名不合法: %s", bookName)
	}
	path := filepath.Join(s.basePath, "projects", bookName, "threads.json")
	return s.saveJSON(path, threads)
}

// ==================== 审稿结果存储 ====================

// ParagraphReviewIssue 段落审稿问题
type ParagraphReviewIssue struct {
	ParagraphID    string `json:"paragraph_id"`
	ParagraphIndex int    `json:"paragraph_index"`
	Type           string `json:"type"`
	Severity       string `json:"severity"`
	Description    string `json:"description"`
	Suggestion     string `json:"suggestion"`
	OriginalText   string `json:"original_text"`
}

// ChapterReviewData 章节审稿数据
type ChapterReviewData struct {
	ChapterID       int                    `json:"chapter_id"`
	OverallScore    int                    `json:"overall_score"`
	ParagraphIssues []ParagraphReviewIssue `json:"paragraph_issues"`
	Suggestions     []string               `json:"suggestions"`
	ReviewedAt      time.Time              `json:"reviewed_at"`
}

// BookReviews 书籍所有审稿结果
type BookReviews struct {
	BookName string                     `json:"book_name"`
	Reviews  map[int]*ChapterReviewData `json:"reviews"` // key: chapter_id
}

// LoadReviews 加载审稿结果
func (s *JSONStore) LoadReviews(bookName string) (*BookReviews, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !validName(bookName) {
		return nil, fmt.Errorf("书名不合法: %s", bookName)
	}

	path := filepath.Join(s.basePath, "projects", bookName, "reviews.json")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &BookReviews{
				BookName: bookName,
				Reviews:  make(map[int]*ChapterReviewData),
			}, nil
		}
		return nil, err
	}

	var reviews BookReviews
	if err := json.Unmarshal(data, &reviews); err != nil {
		return nil, err
	}

	if reviews.Reviews == nil {
		reviews.Reviews = make(map[int]*ChapterReviewData)
	}

	return &reviews, nil
}

// SaveReviews 保存审稿结果
func (s *JSONStore) SaveReviews(bookName string, reviews *BookReviews) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !validName(bookName) {
		return fmt.Errorf("书名不合法: %s", bookName)
	}

	bookPath := filepath.Join(s.basePath, "projects", bookName)
	if err := os.MkdirAll(bookPath, 0755); err != nil {
		return err
	}

	path := filepath.Join(bookPath, "reviews.json")
	return s.saveJSON(path, reviews)
}

// SaveChapterReview 保存单个章节的审稿结果
func (s *JSONStore) SaveChapterReview(bookName string, chapterID int, data *ChapterReviewData) error {
	reviews, err := s.LoadReviews(bookName)
	if err != nil {
		return err
	}

	reviews.Reviews[chapterID] = data
	return s.SaveReviews(bookName, reviews)
}

// GetChapterReview 获取单个章节的审稿结果
func (s *JSONStore) GetChapterReview(bookName string, chapterID int) (*ChapterReviewData, error) {
	reviews, err := s.LoadReviews(bookName)
	if err != nil {
		return nil, err
	}

	return reviews.Reviews[chapterID], nil
}

// ==================== 工具方法 ====================

func (s *JSONStore) loadJSON(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (s *JSONStore) saveJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// ==================== 编排服务辅助方法 ====================

// AppendCharacterEmotion 追加角色情感点
func (s *JSONStore) AppendCharacterEmotion(bookName string, charName string, point model.EmotionPoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	characters, err := s.loadCharactersLocked(bookName)
	if err != nil {
		return err
	}

	for _, char := range characters {
		if char.Name == charName {
			char.EmotionalArc = append(char.EmotionalArc, point)
			break
		}
	}

	return s.saveCharactersLocked(bookName, characters)
}

// AppendCharacterKnownInfo 追加角色已知信息
func (s *JSONStore) AppendCharacterKnownInfo(bookName string, charName string, info model.KnownInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	characters, err := s.loadCharactersLocked(bookName)
	if err != nil {
		return err
	}

	for _, char := range characters {
		if char.Name == charName {
			char.KnownInfos = append(char.KnownInfos, info)
			break
		}
	}

	return s.saveCharactersLocked(bookName, characters)
}

// AppendItemOwnerHistory 追加物品归属历史
func (s *JSONStore) AppendItemOwnerHistory(bookName string, itemName string, change model.ItemOwnerChange) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	items, err := s.loadItemsLocked(bookName)
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.Name == itemName {
			item.OwnerHistory = append(item.OwnerHistory, change)
			// 同时更新当前持有者
			if change.NewOwner != "" {
				item.Owner = change.NewOwner
			}
			break
		}
	}

	return s.saveItemsLocked(bookName, items)
}

// AddCharacterAppearChapter 添加角色出场章节
func (s *JSONStore) AddCharacterAppearChapter(bookName string, charName string, chapterID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	characters, err := s.loadCharactersLocked(bookName)
	if err != nil {
		return err
	}

	for _, char := range characters {
		if char.Name == charName {
			// 检查是否已存在
			for _, chID := range char.AppearChapters {
				if chID == chapterID {
					return nil
				}
			}
			char.AppearChapters = append(char.AppearChapters, chapterID)
			break
		}
	}

	return s.saveCharactersLocked(bookName, characters)
}

// AddItemAppearChapter 添加物品出场章节
func (s *JSONStore) AddItemAppearChapter(bookName string, itemName string, chapterID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	items, err := s.loadItemsLocked(bookName)
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.Name == itemName {
			// 检查是否已存在
			for _, chID := range item.AppearChapters {
				if chID == chapterID {
					return nil
				}
			}
			item.AppearChapters = append(item.AppearChapters, chapterID)
			break
		}
	}

	return s.saveItemsLocked(bookName, items)
}

// UpdateChapterEntityIndex 更新章节实体索引
func (s *JSONStore) UpdateChapterEntityIndex(bookName string, chapterID int, chars, items, locs []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	chapters, err := s.loadChaptersLocked(bookName)
	if err != nil {
		return err
	}

	for _, ch := range chapters {
		if ch.ID == chapterID {
			ch.Characters = chars
			ch.Items = items
			ch.Locations = locs
			break
		}
	}

	return s.saveChaptersLocked(bookName, chapters)
}

// 内部加锁版本的加载方法
func (s *JSONStore) loadCharactersLocked(bookName string) ([]*model.Character, error) {
	bookPath, err := s.getBookPath(bookName)
	if err != nil {
		return nil, err
	}

	var characters []*model.Character
	path := filepath.Join(bookPath, "characters.json")
	if err := s.loadJSON(path, &characters); err != nil {
		return []*model.Character{}, nil
	}
	return characters, nil
}

func (s *JSONStore) saveCharactersLocked(bookName string, characters []*model.Character) error {
	bookPath, err := s.getBookPath(bookName)
	if err != nil {
		return err
	}
	path := filepath.Join(bookPath, "characters.json")
	return s.saveJSON(path, characters)
}

func (s *JSONStore) loadItemsLocked(bookName string) ([]*model.Item, error) {
	bookPath, err := s.getBookPath(bookName)
	if err != nil {
		return nil, err
	}

	var items []*model.Item
	path := filepath.Join(bookPath, "items.json")
	if err := s.loadJSON(path, &items); err != nil {
		return []*model.Item{}, nil
	}
	return items, nil
}

func (s *JSONStore) saveItemsLocked(bookName string, items []*model.Item) error {
	bookPath, err := s.getBookPath(bookName)
	if err != nil {
		return err
	}
	path := filepath.Join(bookPath, "items.json")
	return s.saveJSON(path, items)
}

func (s *JSONStore) loadChaptersLocked(bookName string) ([]*model.Chapter, error) {
	bookPath, err := s.getBookPath(bookName)
	if err != nil {
		return nil, err
	}

	var chapters []*model.Chapter
	path := filepath.Join(bookPath, "structure.json")
	if err := s.loadJSON(path, &chapters); err != nil {
		return []*model.Chapter{}, nil
	}
	return chapters, nil
}

func (s *JSONStore) saveChaptersLocked(bookName string, chapters []*model.Chapter) error {
	bookPath, err := s.getBookPath(bookName)
	if err != nil {
		return err
	}
	path := filepath.Join(bookPath, "structure.json")
	return s.saveJSON(path, chapters)
}

// ==================== 分析报告存储 ====================

// LoadAnalysisReports 加载分析报告列表
func (s *JSONStore) LoadAnalysisReports(bookName string) ([]*model.AnalysisReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !validName(bookName) {
		return nil, fmt.Errorf("书名不合法: %s", bookName)
	}

	path := filepath.Join(s.basePath, "projects", bookName, "analysis_reports.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []*model.AnalysisReport{}, nil
		}
		return nil, err
	}

	var reports []*model.AnalysisReport
	if err := json.Unmarshal(data, &reports); err != nil {
		return nil, err
	}
	return reports, nil
}

// SaveAnalysisReports 保存分析报告列表
func (s *JSONStore) SaveAnalysisReports(bookName string, reports []*model.AnalysisReport) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !validName(bookName) {
		return fmt.Errorf("书名不合法: %s", bookName)
	}

	path := filepath.Join(s.basePath, "projects", bookName, "analysis_reports.json")
	data, err := json.MarshalIndent(reports, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// loadAnalysisReportsLocked 加载分析报告列表（内部方法，调用者需持锁）
func (s *JSONStore) loadAnalysisReportsLocked(bookName string) ([]*model.AnalysisReport, error) {
	bookPath, err := s.getBookPath(bookName)
	if err != nil {
		return nil, err
	}

	path := filepath.Join(bookPath, "analysis_reports.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []*model.AnalysisReport{}, nil
		}
		return nil, err
	}

	var reports []*model.AnalysisReport
	if err := json.Unmarshal(data, &reports); err != nil {
		return nil, err
	}
	return reports, nil
}

// saveAnalysisReportsLocked 保存分析报告列表（内部方法，调用者需持锁）
func (s *JSONStore) saveAnalysisReportsLocked(bookName string, reports []*model.AnalysisReport) error {
	bookPath, err := s.getBookPath(bookName)
	if err != nil {
		return err
	}

	path := filepath.Join(bookPath, "analysis_reports.json")
	data, err := json.MarshalIndent(reports, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// AppendAnalysisReport 追加分析报告
func (s *JSONStore) AppendAnalysisReport(bookName string, report *model.AnalysisReport) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	reports, err := s.loadAnalysisReportsLocked(bookName)
	if err != nil {
		return err
	}
	reports = append(reports, report)
	return s.saveAnalysisReportsLocked(bookName, reports)
}

// LoadPendingGraphSync 加载待审核图谱变更
func (s *JSONStore) LoadPendingGraphSync(bookName string) (*model.PendingGraphSync, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !validName(bookName) {
		return nil, fmt.Errorf("书名不合法: %s", bookName)
	}

	path := filepath.Join(s.basePath, "projects", bookName, "pending_graph_sync.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var pending model.PendingGraphSync
	if err := json.Unmarshal(data, &pending); err != nil {
		return nil, err
	}
	return &pending, nil
}

// SavePendingGraphSync 保存待审核图谱变更
func (s *JSONStore) SavePendingGraphSync(bookName string, pending *model.PendingGraphSync) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !validName(bookName) {
		return fmt.Errorf("书名不合法: %s", bookName)
	}

	path := filepath.Join(s.basePath, "projects", bookName, "pending_graph_sync.json")
	data, err := json.MarshalIndent(pending, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// ClearPendingGraphSync 清除待审核图谱变更
func (s *JSONStore) ClearPendingGraphSync(bookName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !validName(bookName) {
		return fmt.Errorf("书名不合法: %s", bookName)
	}

	path := filepath.Join(s.basePath, "projects", bookName, "pending_graph_sync.json")
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil // 文件不存在视为成功
	}
	return err
}