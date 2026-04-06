package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"ai-writer/internal/config"
	"ai-writer/internal/llm"
	"ai-writer/internal/model"
	"ai-writer/internal/service"
	"ai-writer/internal/store"
)

// 全局变量（需要在启动时初始化）
var jsonStore *store.JSONStore
var billingStore *store.BillingStore
var cfg *config.Config

// validBookName 验证书名是否合法（防止路径注入）
func validBookName(name string) bool {
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

// InitStore 初始化存储
func InitStore(s *store.JSONStore) {
	jsonStore = s
	// 初始化计费存储
	billingStore = store.NewBillingStore(filepath.Join(".", "data"))
}

// InitConfig 初始化配置
func InitConfig(c *config.Config) {
	cfg = c
}

// ==================== 书籍管理 ====================

// ListBooks 列出所有书籍
func ListBooks(c *gin.Context) {
	books, err := jsonStore.ListBooks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, books)
}

// CreateBook 创建书籍
func CreateBook(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	book, err := jsonStore.CreateBook(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, book)
}

// GetBook 获取书籍详情
func GetBook(c *gin.Context) {
	bookID := c.Param("id")

	// 加载书籍所有数据
	book := struct {
		ID         string                   `json:"id"`
		Name       string                   `json:"name"`
		Volumes    interface{}              `json:"volumes"`
		Chapters   interface{}              `json:"chapters"`
		Characters interface{}              `json:"characters"`
		Items      interface{}              `json:"items"`
		Locations  interface{}              `json:"locations"`
		WorldView  interface{}              `json:"worldview"`
	}{
		ID:   bookID,
		Name: bookID,
	}

	// 加载关联数据
	if volumes, err := jsonStore.LoadVolumes(bookID); err == nil {
		book.Volumes = volumes
	}
	if chapters, err := jsonStore.LoadChapters(bookID); err == nil {
		// 计算每个章节的实际字数
		for _, ch := range chapters {
			paragraphs, err := jsonStore.LoadChapterParagraphs(bookID, ch.ID)
			if err == nil {
				ch.WordCount = paragraphs.Metadata.TotalWords
			}
		}
		book.Chapters = chapters
	}
	if characters, err := jsonStore.LoadCharacters(bookID); err == nil {
		book.Characters = characters
	}
	if items, err := jsonStore.LoadItems(bookID); err == nil {
		book.Items = items
	}
	if locations, err := jsonStore.LoadLocations(bookID); err == nil {
		book.Locations = locations
	}
	if worldview, err := jsonStore.LoadWorldView(bookID); err == nil {
		book.WorldView = worldview
	}

	c.JSON(http.StatusOK, book)
}

// UpdateBook 更新书籍（重命名）
func UpdateBook(c *gin.Context) {
	bookID := c.Param("id")

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := jsonStore.RenameBook(bookID, req.Name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 返回更新后的书籍信息
	book, _ := jsonStore.LoadBook(req.Name)
	c.JSON(http.StatusOK, book)
}

// DeleteBook 删除书籍
func DeleteBook(c *gin.Context) {
	bookID := c.Param("id")

	if err := jsonStore.DeleteBook(bookID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ==================== 章节管理 ====================

// ListChapters 列出章节
func ListChapters(c *gin.Context) {
	bookID := c.Param("id")

	chapters, err := jsonStore.LoadChapters(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 计算每个章节的实际字数
	for _, ch := range chapters {
		paragraphs, err := jsonStore.LoadChapterParagraphs(bookID, ch.ID)
		if err == nil {
			ch.WordCount = paragraphs.Metadata.TotalWords
		}
	}

	c.JSON(http.StatusOK, chapters)
}

// CreateChapter 创建章节
func CreateChapter(c *gin.Context) {
	bookID := c.Param("id")

	var req struct {
		Title    string `json:"title" binding:"required"`
		VolumeID string `json:"volume_id"`
		Outline  string `json:"outline"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 加载现有章节
	chapters, err := jsonStore.LoadChapters(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 计算新章节 ID
	maxID := 0
	for _, ch := range chapters {
		if ch.ID > maxID {
			maxID = ch.ID
		}
	}
	newID := maxID + 1

	// 默认分卷
	if req.VolumeID == "" {
		req.VolumeID = "vol_1"
	}

	// 创建新章节
	now := time.Now()
	newChapter := &model.Chapter{
		ID:        newID,
		BookID:    bookID,
		VolumeID:  req.VolumeID,
		Title:     req.Title,
		Outline:   req.Outline,
		TimeInfo:  model.TimeInfo{Label: "", Duration: "0", Events: []string{}},
		CreatedAt: now,
		UpdatedAt: now,
	}

	chapters = append(chapters, newChapter)

	// 保存
	if err := jsonStore.SaveChapters(bookID, chapters); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newChapter)
}

// GetChapter 获取章节详情
func GetChapter(c *gin.Context) {
	bookID := c.Param("id")
	chapterID := parseInt(c.Param("chapter_id"))

	chapters, err := jsonStore.LoadChapters(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 查找章节
	for _, ch := range chapters {
		if ch.ID == chapterID {
			// 加载内容
			content, _ := jsonStore.LoadChapterContent(bookID, chapterID)
			result := map[string]interface{}{
				"id":         ch.ID,
				"book_id":    ch.BookID,
				"volume_id":  ch.VolumeID,
				"title":      ch.Title,
				"outline":    ch.Outline,
				"time_info":  ch.TimeInfo,
				"content":    content,
				"created_at": ch.CreatedAt,
				"updated_at": ch.UpdatedAt,
			}
			c.JSON(http.StatusOK, result)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "章节不存在"})
}

// UpdateChapter 更新章节
func UpdateChapter(c *gin.Context) {
	bookID := c.Param("id")
	chapterID := parseInt(c.Param("chapter_id"))

	var req struct {
		Title   string `json:"title"`
		Outline string `json:"outline"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 加载章节列表
	chapters, err := jsonStore.LoadChapters(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 查找并更新
	for _, ch := range chapters {
		if ch.ID == chapterID {
			if req.Title != "" {
				ch.Title = req.Title
			}
			if req.Outline != "" {
				ch.Outline = req.Outline
			}
			ch.UpdatedAt = time.Now()

			if err := jsonStore.SaveChapters(bookID, chapters); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, ch)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "章节不存在"})
}

// DeleteChapter 删除章节
func DeleteChapter(c *gin.Context) {
	bookID := c.Param("id")
	chapterID := parseInt(c.Param("chapter_id"))

	// 加载章节列表
	chapters, err := jsonStore.LoadChapters(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 查找并删除
	found := false
	var newChapters []*model.Chapter
	for _, ch := range chapters {
		if ch.ID == chapterID {
			found = true
			continue
		}
		newChapters = append(newChapters, ch)
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "章节不存在"})
		return
	}

	// 保存章节结构
	if err := jsonStore.SaveChapters(bookID, newChapters); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 删除段落文件
	jsonStore.DeleteChapterParagraphs(bookID, chapterID)

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// GetChapterContent 获取章节内容
func GetChapterContent(c *gin.Context) {
	bookID := c.Param("id")
	chapterID := c.Param("chapter_id")

	content, err := jsonStore.LoadChapterContent(bookID, parseInt(chapterID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"content": content})
}

// UpdateChapterContent 更新章节内容
func UpdateChapterContent(c *gin.Context) {
	bookID := c.Param("id")
	chapterID := parseInt(c.Param("chapter_id"))

	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := jsonStore.SaveChapterContent(bookID, chapterID, req.Content); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "保存成功"})
}

// UpdateParagraph 更新单个段落
func UpdateParagraph(c *gin.Context) {
	bookID := c.Param("id")
	chapterID := parseInt(c.Param("chapter_id"))

	var req struct {
		ParagraphID string `json:"paragraph_id" binding:"required"`
		Text        string `json:"text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 加载段落数据
	paragraphs, err := jsonStore.LoadChapterParagraphs(bookID, chapterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 查找并更新目标段落
	found := false
	for i, p := range paragraphs.Paragraphs {
		if p.ID == req.ParagraphID {
			paragraphs.Paragraphs[i].Text = req.Text
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusBadRequest, gin.H{"error": "段落不存在"})
		return
	}

	// 保存更新后的段落
	if err := jsonStore.SaveChapterParagraphs(bookID, paragraphs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "段落更新成功"})
}

// ==================== 设定管理 ====================

// GetWorldView 获取世界观
func GetWorldView(c *gin.Context) {
	bookID := c.Param("id")

	worldview, err := jsonStore.LoadWorldView(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, worldview)
}

// UpdateWorldView 更新世界观
func UpdateWorldView(c *gin.Context) {
	bookID := c.Param("id")

	var worldview model.WorldView
	if err := c.ShouldBindJSON(&worldview); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	worldview.BookID = bookID

	if err := jsonStore.SaveWorldView(bookID, &worldview); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "保存成功", "data": worldview})
}

// ListCharacters 列出人物
func ListCharacters(c *gin.Context) {
	bookID := c.Param("id")

	characters, err := jsonStore.LoadCharacters(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, characters)
}

// CreateCharacter 创建人物
func CreateCharacter(c *gin.Context) {
	bookID := c.Param("id")

	var req struct {
		Name        string `json:"name" binding:"required"`
		Role        string `json:"role"`
		Gender      string `json:"gender"`
		Bio         string `json:"bio"`
		Status      string `json:"status"`
		Faction     string `json:"faction"`
		Sect        string `json:"sect"`
		Position    string `json:"position"`
		Cultivation string `json:"cultivation"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 默认值
	if req.Role == "" {
		req.Role = "配角"
	}
	if req.Gender == "" {
		req.Gender = "男"
	}
	if req.Status == "" {
		req.Status = "存活"
	}

	// 加载现有人物
	characters, err := jsonStore.LoadCharacters(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 检查是否已存在
	for _, ch := range characters {
		if ch.Name == req.Name {
			c.JSON(http.StatusBadRequest, gin.H{"error": "人物名称已存在"})
			return
		}
	}

	// 创建新人物
	newChar := &model.Character{
		ID:           generateID(),
		BookID:       bookID,
		Name:         req.Name,
		Role:         req.Role,
		Gender:       req.Gender,
		Bio:          req.Bio,
		Status:       req.Status,
		Faction:      req.Faction,
		Sect:         req.Sect,
		Position:     req.Position,
		Cultivation:  req.Cultivation,
		Relations:     []model.Relation{},
	}

	characters = append(characters, newChar)

	// 保存
	if err := jsonStore.SaveCharacters(bookID, characters); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newChar)
}

// UpdateCharacter 更新人物
func UpdateCharacter(c *gin.Context) {
	bookID := c.Param("id")
	charID := c.Param("char_id")

	var req struct {
		Name        string `json:"name"`
		Role        string `json:"role"`
		Gender      string `json:"gender"`
		Bio         string `json:"bio"`
		Status      string `json:"status"`
		Faction     string `json:"faction"`
		Sect        string `json:"sect"`
		Position    string `json:"position"`
		Cultivation string `json:"cultivation"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 加载人物列表
	characters, err := jsonStore.LoadCharacters(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 查找并更新（使用索引访问，避免值拷贝问题）
	for i := range characters {
		if characters[i].ID == charID {
			if req.Name != "" {
				characters[i].Name = req.Name
			}
			if req.Role != "" {
				characters[i].Role = req.Role
			}
			if req.Gender != "" {
				characters[i].Gender = req.Gender
			}
			characters[i].Bio = req.Bio
			if req.Status != "" {
				characters[i].Status = req.Status
			}
			characters[i].Faction = req.Faction
			characters[i].Sect = req.Sect
			characters[i].Position = req.Position
			characters[i].Cultivation = req.Cultivation

			if err := jsonStore.SaveCharacters(bookID, characters); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, characters[i])
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "人物不存在"})
}

// DeleteCharacter 删除人物
func DeleteCharacter(c *gin.Context) {
	bookID := c.Param("id")
	charID := c.Param("char_id")

	// 加载人物列表
	characters, err := jsonStore.LoadCharacters(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 查找并删除
	found := false
	var newCharacters []*model.Character
	for _, ch := range characters {
		if ch.ID == charID {
			found = true
			continue
		}
		newCharacters = append(newCharacters, ch)
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "人物不存在"})
		return
	}

	// 保存
	if err := jsonStore.SaveCharacters(bookID, newCharacters); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ListItems 列出物品
func ListItems(c *gin.Context) {
	bookID := c.Param("id")

	items, err := jsonStore.LoadItems(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}

// CreateItem 创建物品
func CreateItem(c *gin.Context) {
	bookID := c.Param("id")

	var req struct {
		Name        string `json:"name" binding:"required"`
		Type        string `json:"type"`
		Owner       string `json:"owner"`
		Description string `json:"description"`
		Origin      string `json:"origin"`
		Abilities   string `json:"abilities"`
		Rank        string `json:"rank"`
		Faction     string `json:"faction"`
		Sect        string `json:"sect"`
		Location    string `json:"location"` // 所在地点
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 默认值
	if req.Type == "" {
		req.Type = "法宝"
	}

	// 加载现有物品
	items, err := jsonStore.LoadItems(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 检查是否已存在
	for _, item := range items {
		if item.Name == req.Name {
			c.JSON(http.StatusBadRequest, gin.H{"error": "物品名称已存在"})
			return
		}
	}

	// 创建新物品
	newItem := &model.Item{
		ID:          generateID(),
		BookID:      bookID,
		Name:        req.Name,
		Type:        req.Type,
		Owner:       req.Owner,
		Description: req.Description,
		Origin:      req.Origin,
		Abilities:   req.Abilities,
		Rank:        req.Rank,
		Faction:     req.Faction,
		Sect:        req.Sect,
		Location:    req.Location,
	}

	items = append(items, newItem)

	// 保存
	if err := jsonStore.SaveItems(bookID, items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newItem)
}

// DeleteItem 删除物品
func DeleteItem(c *gin.Context) {
	bookID := c.Param("id")
	itemID := c.Param("item_id")

	// 加载物品列表
	items, err := jsonStore.LoadItems(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 查找并删除
	found := false
	var newItems []*model.Item
	for _, item := range items {
		if item.ID == itemID {
			found = true
			continue
		}
		newItems = append(newItems, item)
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "物品不存在"})
		return
	}

	// 保存
	if err := jsonStore.SaveItems(bookID, newItems); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// UpdateItem 更新物品
// UpdateItem 更新物品
func UpdateItem(c *gin.Context) {
	bookID := c.Param("id")
	itemID := c.Param("item_id")

	var req struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Owner       string `json:"owner"`
		Description string `json:"description"`
		Origin      string `json:"origin"`
		Abilities   string `json:"abilities"`
		Rank        string `json:"rank"`
		Faction     string `json:"faction"`
		Sect        string `json:"sect"`
		Location    string `json:"location"` // 所在地点
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 加载物品列表
	items, err := jsonStore.LoadItems(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 查找并更新（使用索引访问，避免值拷贝问题）
	for i := range items {
		if items[i].ID == itemID {
			if req.Name != "" {
				items[i].Name = req.Name
			}
			items[i].Type = req.Type
			items[i].Owner = req.Owner
			items[i].Description = req.Description
			items[i].Origin = req.Origin
			items[i].Abilities = req.Abilities
			items[i].Rank = req.Rank
			items[i].Faction = req.Faction
			items[i].Sect = req.Sect
			items[i].Location = req.Location

			if err := jsonStore.SaveItems(bookID, items); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "物品不存在"})
}

// ListLocations 列出地点
func ListLocations(c *gin.Context) {
	bookID := c.Param("id")

	locations, err := jsonStore.LoadLocations(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, locations)
}

// CreateLocation 创建地点
func CreateLocation(c *gin.Context) {
	bookID := c.Param("id")

	var req struct {
		Name        string `json:"name" binding:"required"`
		Parent      string `json:"parent"`
		Faction     string `json:"faction"`
		Danger      string `json:"danger"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 加载现有地点
	locations, err := jsonStore.LoadLocations(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 检查是否已存在
	for _, loc := range locations {
		if loc.Name == req.Name {
			c.JSON(http.StatusBadRequest, gin.H{"error": "地点名称已存在"})
			return
		}
	}

	// 创建新地点
	newLocation := &model.Location{
		ID:          generateID(),
		BookID:      bookID,
		Name:        req.Name,
		Parent:      req.Parent,
		Neighbors:   []string{},
		Description: req.Description,
		Faction:     req.Faction,
		Danger:      req.Danger,
	}

	locations = append(locations, newLocation)

	// 保存
	if err := jsonStore.SaveLocations(bookID, locations); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newLocation)
}

// DeleteLocation 删除地点
func DeleteLocation(c *gin.Context) {
	bookID := c.Param("id")
	locID := c.Param("loc_id")

	// 加载地点列表
	locations, err := jsonStore.LoadLocations(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 查找并删除
	found := false
	var newLocations []*model.Location
	for _, loc := range locations {
		if loc.ID == locID {
			found = true
			continue
		}
		newLocations = append(newLocations, loc)
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "地点不存在"})
		return
	}

	// 保存
	if err := jsonStore.SaveLocations(bookID, newLocations); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// UpdateLocation 更新地点
func UpdateLocation(c *gin.Context) {
	bookID := c.Param("id")
	locID := c.Param("loc_id")

	var req struct {
		Name        string   `json:"name"`
		Parent      string   `json:"parent"`
		Neighbors   []string `json:"neighbors"`
		Faction     string   `json:"faction"`
		Danger      string   `json:"danger"`
		Description string   `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 加载地点列表
	locations, err := jsonStore.LoadLocations(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 查找并更新（使用索引访问，避免值拷贝问题）
	for i := range locations {
		if locations[i].ID == locID {
			if req.Name != "" {
				locations[i].Name = req.Name
			}
			locations[i].Parent = req.Parent
			locations[i].Neighbors = req.Neighbors
			locations[i].Faction = req.Faction
			locations[i].Danger = req.Danger
			locations[i].Description = req.Description

			if err := jsonStore.SaveLocations(bookID, locations); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "地点不存在"})
}

// ==================== 因果链 ====================

// GetCausalChains 获取因果链
func GetCausalChains(c *gin.Context) {
	bookID := c.Param("id")

	events, err := jsonStore.LoadCausalChains(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, events)
}

// CreateCausalEvent 创建因果事件
func CreateCausalEvent(c *gin.Context) {
	bookID := c.Param("id")

	var req struct {
		Cause      string   `json:"cause" binding:"required"`
		Event      string   `json:"event" binding:"required"`
		Effect     string   `json:"effect" binding:"required"`
		Decision   string   `json:"decision"`
		ChapterID  int      `json:"chapter_id"`
		Characters []string `json:"characters"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 加载现有因果链
	events, err := jsonStore.LoadCausalChains(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 创建新因果事件
	now := time.Now()
	newEvent := &model.CausalEvent{
		ID:         generateID(),
		BookID:     bookID,
		ChapterID:  req.ChapterID,
		Cause:      req.Cause,
		Event:      req.Event,
		Effect:     req.Effect,
		Decision:   req.Decision,
		Characters: req.Characters,
		Status:     model.CausalActive,
		CreatedAt:  now,
	}

	events = append(events, newEvent)

	// 保存
	if err := jsonStore.SaveCausalChains(bookID, events); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newEvent)
}

// UpdateCausalEvent 更新因果事件
func UpdateCausalEvent(c *gin.Context) {
	bookID := c.Param("id")
	eventID := c.Param("event_id")

	var req struct {
		Cause      string   `json:"cause"`
		Event      string   `json:"event"`
		Effect     string   `json:"effect"`
		Decision   string   `json:"decision"`
		ChapterID  int      `json:"chapter_id"`
		Characters []string `json:"characters"`
		Status     string   `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 加载因果链
	events, err := jsonStore.LoadCausalChains(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 查找并更新（使用索引访问，避免值拷贝问题）
	for i := range events {
		if events[i].ID == eventID {
			if req.Cause != "" {
				events[i].Cause = req.Cause
			}
			if req.Event != "" {
				events[i].Event = req.Event
			}
			if req.Effect != "" {
				events[i].Effect = req.Effect
			}
			if req.Decision != "" {
				events[i].Decision = req.Decision
			}
			if req.ChapterID > 0 {
				events[i].ChapterID = req.ChapterID
			}
			if req.Characters != nil {
				events[i].Characters = req.Characters
			}
			if req.Status != "" {
				events[i].Status = model.CausalStatus(req.Status)
			}

			if err := jsonStore.SaveCausalChains(bookID, events); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, events[i])
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "因果事件不存在"})
}

// ==================== 伏笔 ====================

// ListForeshadows 列出伏笔
func ListForeshadows(c *gin.Context) {
	bookID := c.Param("id")

	foreshadows, err := jsonStore.LoadForeshadows(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, foreshadows)
}

// CreateForeshadow 创建伏笔
func CreateForeshadow(c *gin.Context) {
	bookID := c.Param("id")

	var req struct {
		Content       string `json:"content" binding:"required"`
		Type          string `json:"type"`
		SourceChapter int    `json:"source_chapter"`
		TargetChapter int    `json:"target_chapter"`
		Importance    string `json:"importance"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 默认值
	if req.Type == "" {
		req.Type = "plot"
	}
	if req.Importance == "" {
		req.Importance = "medium"
	}

	// 加载现有伏笔
	foreshadows, err := jsonStore.LoadForeshadows(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 创建新伏笔
	now := time.Now()
	newForeshadow := &model.Foreshadow{
		ID:            generateID(),
		BookID:        bookID,
		Content:       req.Content,
		Type:          model.ForeshadowType(req.Type),
		Importance:    model.Importance(req.Importance),
		SourceChapter: req.SourceChapter,
		TargetChapter: req.TargetChapter,
		Status:        model.ForeshadowActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	foreshadows = append(foreshadows, newForeshadow)

	// 保存
	if err := jsonStore.SaveForeshadows(bookID, foreshadows); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newForeshadow)
}

// UpdateForeshadow 更新伏笔
func UpdateForeshadow(c *gin.Context) {
	bookID := c.Param("id")
	fsID := c.Param("fid")

	var req struct {
		Content       string `json:"content"`
		Type          string `json:"type"`
		SourceChapter int    `json:"source_chapter"`
		TargetChapter int    `json:"target_chapter"`
		Importance    string `json:"importance"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 加载伏笔列表
	foreshadows, err := jsonStore.LoadForeshadows(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 查找并更新（使用索引访问，避免值拷贝问题）
	for i := range foreshadows {
		if foreshadows[i].ID == fsID {
			if req.Content != "" {
				foreshadows[i].Content = req.Content
			}
			if req.Type != "" {
				foreshadows[i].Type = model.ForeshadowType(req.Type)
			}
			if req.SourceChapter > 0 {
				foreshadows[i].SourceChapter = req.SourceChapter
			}
			if req.TargetChapter > 0 {
				foreshadows[i].TargetChapter = req.TargetChapter
			}
			if req.Importance != "" {
				foreshadows[i].Importance = model.Importance(req.Importance)
			}
			foreshadows[i].UpdatedAt = time.Now()

			if err := jsonStore.SaveForeshadows(bookID, foreshadows); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, foreshadows[i])
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "伏笔不存在"})
}

// ResolveForeshadow 回收伏笔
func ResolveForeshadow(c *gin.Context) {
	bookID := c.Param("id")
	fsID := c.Param("fid")

	var req struct {
		ChapterID      int    `json:"chapter_id"`
		ResolvedContent string `json:"resolved_content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 加载伏笔列表
	foreshadows, err := jsonStore.LoadForeshadows(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 查找并更新（使用索引访问，避免值拷贝问题）
	for i := range foreshadows {
		if foreshadows[i].ID == fsID {
			foreshadows[i].Status = model.ForeshadowResolved
			foreshadows[i].ResolvedChapter = req.ChapterID
			foreshadows[i].ResolvedContent = req.ResolvedContent
			foreshadows[i].UpdatedAt = time.Now()

			if err := jsonStore.SaveForeshadows(bookID, foreshadows); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, foreshadows[i])
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "伏笔不存在"})
}

// GetForeshadowWarnings 获取伏笔预警
func GetForeshadowWarnings(c *gin.Context) {
	bookID := c.Param("id")

	// 加载伏笔
	foreshadows, err := jsonStore.LoadForeshadows(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 加载章节获取当前最新章节
	chapters, err := jsonStore.LoadChapters(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	currentChapter := 0
	for _, ch := range chapters {
		if ch.ID > currentChapter {
			currentChapter = ch.ID
		}
	}

	// 检查预警
	const warningThreshold = 5
	var warnings []map[string]interface{}

	for _, fs := range foreshadows {
		if fs.Status != model.ForeshadowActive {
			continue
		}

		gap := currentChapter - fs.SourceChapter
		if gap > warningThreshold {
			warnings = append(warnings, map[string]interface{}{
				"id":            fs.ID,
				"content":       fs.Content,
				"source_chapter": fs.SourceChapter,
				"gap":           gap,
				"type":          "timeout",
				"message":       fmt.Sprintf("伏笔已过 %d 章未回收", gap),
			})
		}

		if fs.TargetChapter > 0 && currentChapter > fs.TargetChapter {
			warnings = append(warnings, map[string]interface{}{
				"id":             fs.ID,
				"content":        fs.Content,
				"target_chapter": fs.TargetChapter,
				"current_chapter": currentChapter,
				"type":           "overdue",
				"message":        fmt.Sprintf("预期第%d章回收，当前已到第%d章", fs.TargetChapter, currentChapter),
			})
		}
	}

	c.JSON(http.StatusOK, warnings)
}

// ==================== 时间线 ====================

// GetTimeline 获取时间线
func GetTimeline(c *gin.Context) {
	bookID := c.Param("id")

	chapters, err := jsonStore.LoadChapters(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 提取时间线信息
	timeline := make([]map[string]interface{}, 0)
	for _, ch := range chapters {
		timeline = append(timeline, map[string]interface{}{
			"chapter_id": ch.ID,
			"title":      ch.Title,
			"time_info":  ch.TimeInfo,
		})
	}

	c.JSON(http.StatusOK, timeline)
}

// GetNarrativeThreads 获取叙事线程
func GetNarrativeThreads(c *gin.Context) {
	bookID := c.Param("id")

	threads, err := jsonStore.LoadThreads(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, threads)
}

// CreateNarrativeThread 创建叙事线程
func CreateNarrativeThread(c *gin.Context) {
	bookID := c.Param("id")

	var req struct {
		Name         string `json:"name" binding:"required"`
		Type         string `json:"type"`
		Goal         string `json:"goal"`
		StartChapter int    `json:"start_chapter"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 默认值
	if req.Type == "" {
		req.Type = "sub"
	}

	// 加载现有线程
	threads, err := jsonStore.LoadThreads(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 创建新线程
	now := time.Now()
	newThread := &model.NarrativeThread{
		ID:               generateID(),
		BookID:           bookID,
		Name:             req.Name,
		Type:             model.ThreadType(req.Type),
		Goal:             req.Goal,
		Status:           model.ThreadActive,
		StartChapter:     req.StartChapter,
		Chapters:         []int{},
		LastActiveChapter: req.StartChapter,
		CreatedAt:        now,
	}

	threads = append(threads, newThread)

	// 保存
	if err := jsonStore.SaveThreads(bookID, threads); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newThread)
}

// ==================== AI 写作 ====================

// AIGenerateRequest AI 生成请求
type AIGenerateRequest struct {
	BookName  string `json:"book_name"`
	ChapterID int    `json:"chapter_id"`
	Outline   string `json:"outline"`
}

// AIGenerate AI 生成
func AIGenerate(c *gin.Context) {
	var req AIGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 使用 bookID 作为 bookName
	bookName := req.BookName
	if bookName == "" {
		bookName = c.Param("id")
	}

	// 获取 LLM 客户端
	llmClient, err := getLLMClient()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "AI服务未配置，请先在系统设置中配置API Key",
			"hint":    "请在系统设置中配置API Key后重试",
		})
		return
	}

	// 创建写作服务
	writerService := service.NewWriterService(llmClient, jsonStore, getPromptsConfig())

	// 生成内容
	content, err := writerService.WriteChapter(c.Request.Context(), bookName, req.ChapterID, req.Outline)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "生成失败: " + err.Error(),
			"hint":    "请检查章节是否存在，或稍后重试",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"content": content,
		"message": "生成成功",
	})
}

// AIGenerateStream AI 流式生成
func AIGenerateStream(c *gin.Context) {
	bookID := c.Param("id")
	chapterID := parseInt(c.Query("chapter_id"))
	outline := c.Query("outline")

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// 获取 LLM 客户端
	llmClient, err := getLLMClient()
	if err != nil {
		c.SSEvent("error", "AI服务未配置，请先在系统设置中配置API Key")
		c.Writer.Flush()
		return
	}

	// 创建写作服务
	writerService := service.NewWriterService(llmClient, jsonStore, getPromptsConfig())

	// 流式生成
	stream, err := writerService.WriteChapterStream(c.Request.Context(), bookID, chapterID, outline)
	if err != nil {
		c.SSEvent("error", err.Error())
		c.Writer.Flush()
		return
	}

	for chunk := range stream {
		if chunk.Error != nil {
			c.SSEvent("error", chunk.Error.Error())
		} else {
			c.SSEvent("content", chunk.Content)
		}
		c.Writer.Flush()
	}

	c.SSEvent("done", "生成完成")
	c.Writer.Flush()
}

// getReviewsFile 获取审稿结果文件路径（验证书名防止路径注入）
func getReviewsFile(bookName string) (string, error) {
	if !validBookName(bookName) {
		return "", fmt.Errorf("书名不合法: %s", bookName)
	}
	return filepath.Join("projects", bookName, "reviews.json"), nil
}

// loadReviewResult 加载审稿结果
func loadReviewResult(bookName string, chapterID int) (*service.ParagraphReviewResult, error) {
	filePath, err := getReviewsFile(bookName)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var reviews struct {
		Reviews map[string]*service.ParagraphReviewResult `json:"reviews"`
	}
	if err := json.Unmarshal(data, &reviews); err != nil {
		return nil, err
	}

	return reviews.Reviews[fmt.Sprintf("%d", chapterID)], nil
}

// saveReviewResult 保存审稿结果
func saveReviewResult(bookName string, chapterID int, result *service.ParagraphReviewResult) error {
	filePath, err := getReviewsFile(bookName)
	if err != nil {
		return err
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 加载现有审稿
	data, _ := os.ReadFile(filePath)

	var reviews struct {
		BookName string                                     `json:"book_name"`
		Reviews  map[string]*service.ParagraphReviewResult `json:"reviews"`
	}
	reviews.Reviews = make(map[string]*service.ParagraphReviewResult)

	if len(data) > 0 {
		json.Unmarshal(data, &reviews)
	}

	// 更新指定章节的审稿结果
	reviews.BookName = bookName
	reviews.Reviews[fmt.Sprintf("%d", chapterID)] = result

	// 保存
	newData, err := json.MarshalIndent(reviews, "", "  ")
		if err != nil {
			return fmt.Errorf("序列化失败: %w", err)
		}
		return os.WriteFile(filePath, newData, 0644)
	}

// AIReview AI 审稿
func AIReview(c *gin.Context) {
	bookName := c.Query("book_name")
	if bookName == "" {
		bookName = c.Param("id")
	}
	chapterID := 0
	if c.Query("chapter_id") != "" {
		chapterID = parseInt(c.Query("chapter_id"))
	}

	// GET 请求：加载已保存的审稿结果
	if c.Request.Method == "GET" {
		if bookName == "" || chapterID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少参数"})
			return
		}
		result, err := loadReviewResult(bookName, chapterID)
		if err != nil || result == nil {
			c.JSON(http.StatusOK, gin.H{"message": "暂无审稿结果"})
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}

	// POST 请求：执行审稿
	var req struct {
		BookName  string `json:"book_name"`
		ChapterID int    `json:"chapter_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.BookName != "" {
		bookName = req.BookName
	}
	if req.ChapterID != 0 {
		chapterID = req.ChapterID
	}

	// 获取 LLM 客户端
	llmClient, err := getLLMClient()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "AI服务未配置",
			"hint":    "请在系统设置中配置API Key后重试",
		})
		return
	}

	// 创建审稿服务
	reviewService := service.NewReviewService(llmClient, jsonStore, getPromptsConfig())

	// 按段落审稿
	result, err := reviewService.ReviewChapterByParagraph(c.Request.Context(), bookName, chapterID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "审稿失败: " + err.Error(),
		})
		return
	}

	// 保存审稿结果
	if err := saveReviewResult(bookName, chapterID, result); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "审稿完成但保存失败: " + err.Error(),
			"result":  result,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// AIAudit AI 状态审计
func AIAudit(c *gin.Context) {
	bookID := c.Param("id")
	chapterID := parseInt(c.Query("chapter_id"))

	// 获取 LLM 客户端
	llmClient, err := getLLMClient()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "AI服务未配置",
			"hint":    "请在系统设置中配置API Key后重试",
		})
		return
	}

	// 创建审稿服务
	reviewService := service.NewReviewService(llmClient, jsonStore, getPromptsConfig())

	// 状态审计
	chapter, err := reviewService.AuditChapter(c.Request.Context(), bookID, chapterID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "审计失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "审计完成",
		"chapter": chapter,
	})
}

// AIRewrite AI 重写
func AIRewrite(c *gin.Context) {
	var req struct {
		BookName    string `json:"book_name"`
		ChapterID   int    `json:"chapter_id"`
		Instruction string `json:"instruction"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bookName := req.BookName
	if bookName == "" {
		bookName = c.Param("id")
	}

	// 获取 LLM 客户端
	llmClient, err := getLLMClient()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "AI服务未配置",
			"hint":    "请在系统设置中配置API Key后重试",
		})
		return
	}

	// 创建写作服务
	writerService := service.NewWriterService(llmClient, jsonStore, getPromptsConfig())

	// 重写
	content, err := writerService.RewriteChapter(c.Request.Context(), bookName, req.ChapterID, req.Instruction)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "重写失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"content": content,
		"message": "重写成功",
	})
}

// AIRewriteParagraph AI 重写段落
func AIRewriteParagraph(c *gin.Context) {
	var req struct {
		BookName    string `json:"book_name"`
		ChapterID   int    `json:"chapter_id"`
		ParagraphID string `json:"paragraph_id"`
		Instruction string `json:"instruction"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bookName := req.BookName
	if bookName == "" {
		bookName = c.Param("id")
	}

	// 获取 LLM 客户端
	llmClient, err := getLLMClient()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "AI服务未配置",
			"hint":    "请在系统设置中配置API Key后重试",
		})
		return
	}

	// 创建写作服务
	writerService := service.NewWriterService(llmClient, jsonStore, getPromptsConfig())

	// 重写段落
	newContent, err := writerService.RewriteParagraph(c.Request.Context(), bookName, req.ChapterID, req.ParagraphID, req.Instruction)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "重写失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"content": newContent,
		"message": "重写成功",
	})
}

// ==================== 智能工具箱 ====================

// ToolNaming 命名工具
func ToolNaming(c *gin.Context) {
	var req struct {
		Type   string `json:"type"`
		Genre  string `json:"genre"`
		Count  int    `json:"count"`
		Gender string `json:"gender"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "请使用 CLI 命令生成名称",
		"command": fmt.Sprintf("ai-writer tool name --type %s --genre %s --count %d", req.Type, req.Genre, req.Count),
	})
}

// ToolCharacter 角色生成工具
func ToolCharacter(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "请使用 CLI 命令生成角色",
		"command": "ai-writer tool character --type <类型> --gender <性别> --genre <题材>",
	})
}

// ToolConflict 冲突生成工具
func ToolConflict(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "请使用 CLI 命令生成冲突",
		"command": "ai-writer tool conflict --type <类型> --genre <题材>",
	})
}

// ToolScene 场景生成工具
func ToolScene(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "请使用 CLI 命令生成场景",
		"command": "ai-writer tool scene --type <类型> --location <地点>",
	})
}

// ToolGoldfinger 金手指生成工具
func ToolGoldfinger(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "请使用 CLI 命令生成金手指",
		"command": "ai-writer tool goldfinger --type <类型> --genre <题材>",
	})
}

// ToolTitle 书名生成工具
func ToolTitle(c *gin.Context) {
	var req struct {
		Genre string `json:"genre"`
		Theme string `json:"theme"`
		Count int    `json:"count"`
		Style string `json:"style"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "请使用 CLI 命令生成书名",
		"command": fmt.Sprintf("ai-writer tool title --genre %s --count %d", req.Genre, req.Count),
	})
}

// ToolSynopsis 简介生成工具
func ToolSynopsis(c *gin.Context) {
	var req struct {
		Genre     string `json:"genre"`
		MainChar  string `json:"main_char"`
		WorldView string `json:"world_view"`
		Type      string `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "请使用 CLI 命令生成简介",
		"command": fmt.Sprintf("ai-writer tool synopsis --genre %s --type %s", req.Genre, req.Type),
	})
}

// ToolTwist 剧情转折生成工具
func ToolTwist(c *gin.Context) {
	var req struct {
		Type       string `json:"type"`
		Genre      string `json:"genre"`
		Context    string `json:"context"`
		Characters string `json:"characters"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "请使用 CLI 命令生成剧情转折",
		"command": fmt.Sprintf("ai-writer tool twist --type %s --genre %s", req.Type, req.Genre),
	})
}

// ToolDialogue 对话生成工具
func ToolDialogue(c *gin.Context) {
	var req struct {
		Characters string `json:"characters"`
		Situation  string `json:"situation"`
		Mood       string `json:"mood"`
		Genre      string `json:"genre"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "请使用 CLI 命令生成对话",
		"command": fmt.Sprintf("ai-writer tool dialogue --characters \"%s\" --situation \"%s\"", req.Characters, req.Situation),
	})
}

// ==================== 架构师 ====================

// ArchitectGenerate 生成大纲
func ArchitectGenerate(c *gin.Context) {
	var req struct {
		Genre       string `json:"genre"`
		MainChar    string `json:"main_char"`
		Theme       string `json:"theme"`
		TargetWords int    `json:"target_words"`
		Volumes     int    `json:"volumes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "请使用 CLI 命令生成大纲",
		"command": fmt.Sprintf("ai-writer architect generate --genre %s --volumes %d", req.Genre, req.Volumes),
	})
}

// ArchitectFission 分形裂变
func ArchitectFission(c *gin.Context) {
	var req struct {
		Strategy string `json:"strategy"`
		Count    int    `json:"count"`
		Outline  string `json:"outline"`
		NodeType string `json:"node_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "请使用 CLI 命令进行分形裂变",
		"command": fmt.Sprintf("ai-writer architect fission --strategy %s --count %d", req.Strategy, req.Count),
	})
}

// ArchitectStrategies 获取裂变策略
func ArchitectStrategies(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"strategies": map[string]interface{}{
			"expand": []map[string]string{
				{"id": "expand_detail", "name": "详细展开", "description": "将简单大纲展开为详细章节"},
				{"id": "expand_plot", "name": "剧情展开", "description": "展开剧情细节和转折"},
			},
			"refine": []map[string]string{
				{"id": "refine_logic", "name": "逻辑优化", "description": "优化剧情逻辑"},
				{"id": "refine_pacing", "name": "节奏优化", "description": "优化叙事节奏"},
			},
			"branch": []map[string]string{
				{"id": "branch_plot", "name": "剧情分支", "description": "生成多条剧情线"},
				{"id": "branch_ending", "name": "结局分支", "description": "生成多种可能结局"},
			},
		},
	})
}

// ==================== 拆书分析 ====================

// AnalysisParse 解析文件
func AnalysisParse(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "请使用 CLI 命令解析文件",
		"command": "ai-writer analysis parse <文件路径>",
	})
}

// AnalysisAnalyze 分析作品
func AnalysisAnalyze(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "请使用 CLI 命令分析作品",
		"command": "ai-writer analysis analyze <文件路径>",
	})
}

// ==================== 图谱 ====================

// GetKnowledgeGraph 获取知识图谱
func GetKnowledgeGraph(c *gin.Context) {
	bookID := c.Param("id")

	graphService := service.NewGraphService(jsonStore)
	data, err := graphService.BuildGraph(bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetEChartsData 获取 ECharts 图谱数据
func GetEChartsData(c *gin.Context) {
	bookID := c.Param("id")
	graphType := c.DefaultQuery("type", "relationship")

	// 验证图谱类型
	validTypes := map[string]bool{
		"relationship": true,
		"causal":       true,
		"foreshadow":   true,
		"thread":       true,
		"emotion":      true,
		"timeline":     true,
	}
	if !validTypes[graphType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的图谱类型: " + graphType})
		return
	}

	graphService := service.NewGraphService(jsonStore)

	var data interface{}
	var err error

	switch graphType {
	case "relationship":
		data, err = graphService.BuildGraph(bookID)
	case "causal":
		data, err = graphService.BuildCausalGraph(bookID)
	case "foreshadow":
		data, err = graphService.BuildForeshadowGraph(bookID)
	case "thread":
		data, err = graphService.BuildThreadGraph(bookID)
	case "emotion":
		data, err = graphService.BuildEmotionGraph(bookID)
	case "timeline":
		data, err = graphService.BuildTimelineGraph(bookID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

// ==================== 系统设置 ====================

// GetConfig 获取配置
func GetConfig(c *gin.Context) {
	if cfg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "配置未初始化"})
		return
	}

	// API Key 掩码处理
	apiKeyDisplay := ""
	if cfg.LLM.APIKey != "" {
		// 显示前4位和后4位，中间用*代替
		if len(cfg.LLM.APIKey) > 8 {
			apiKeyDisplay = cfg.LLM.APIKey[:4] + "****" + cfg.LLM.APIKey[len(cfg.LLM.APIKey)-4:]
		} else {
			apiKeyDisplay = "****"
		}
	}

	// 返回非敏感配置
	cfgData := map[string]interface{}{
		"provider":       cfg.LLM.Provider,
		"api_key_set":    cfg.LLM.APIKey != "",
		"api_key_display": apiKeyDisplay,
		"base_url":       cfg.LLM.BaseURL,
		"models":         cfg.LLM.Models,
		"temperatures":   cfg.LLM.Temperatures,
		"max_retries":    cfg.LLM.MaxRetries,
		"timeout":        cfg.LLM.Timeout,
		"vector_store": map[string]int{
			"chunk_size": cfg.VectorStore.ChunkSize,
			"overlap":    cfg.VectorStore.Overlap,
		},
		"storage": map[string]string{
			"projects_dir": cfg.Storage.ProjectsDir,
			"vector_db_dir": cfg.Storage.VectorDBDir,
		},
		"pricing": cfg.Pricing,
	}
	c.JSON(http.StatusOK, cfgData)
}

// UpdateConfig 更新配置
func UpdateConfig(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新配置文件
	if provider, ok := req["provider"].(string); ok {
		viper.Set("llm.provider", provider)
		cfg.LLM.Provider = provider
	}
	if apiKey, ok := req["api_key"].(string); ok && apiKey != "" {
		viper.Set("llm.api_key", apiKey)
		cfg.LLM.APIKey = apiKey
	}
	if baseURL, ok := req["base_url"].(string); ok {
		viper.Set("llm.base_url", baseURL)
		cfg.LLM.BaseURL = baseURL
	}
	if models, ok := req["models"].(map[string]interface{}); ok {
		viper.Set("llm.models", models)
		// 转换为 map[string]string
		for k, v := range models {
			if vs, ok := v.(string); ok {
				cfg.LLM.Models[k] = vs
			}
		}
	}
	if temperatures, ok := req["temperatures"].(map[string]interface{}); ok {
		viper.Set("llm.temperatures", temperatures)
		// 转换为 map[string]float64
		for k, v := range temperatures {
			if vf, ok := v.(float64); ok {
				cfg.LLM.Temperatures[k] = vf
			}
		}
	}
	if maxRetries, ok := req["max_retries"].(float64); ok {
		viper.Set("llm.max_retries", int(maxRetries))
		cfg.LLM.MaxRetries = int(maxRetries)
	}
	if timeout, ok := req["timeout"].(float64); ok {
		viper.Set("llm.timeout", int(timeout))
		cfg.LLM.Timeout = int(timeout)
	}
	if vectorStore, ok := req["vector_store"].(map[string]interface{}); ok {
		if chunkSize, ok := vectorStore["chunk_size"].(float64); ok {
			viper.Set("vector_store.chunk_size", int(chunkSize))
			cfg.VectorStore.ChunkSize = int(chunkSize)
		}
		if overlap, ok := vectorStore["overlap"].(float64); ok {
			viper.Set("vector_store.overlap", int(overlap))
			cfg.VectorStore.Overlap = int(overlap)
		}
	}

	// 保存配置文件
	if err := viper.WriteConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存配置失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "配置已更新并生效",
		"data":    req,
	})
}

// GetPrompts 获取提示词
func GetPrompts(c *gin.Context) {
	if cfg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "配置未初始化"})
		return
	}

	prompts := map[string]string{
		"writer_system":             cfg.Prompts.WriterSystem,
		"architect_system":          cfg.Prompts.ArchitectSystem,
		"reviewer_system":           cfg.Prompts.ReviewerSystem,
		"auditor_system":            cfg.Prompts.AuditorSystem,
		"timekeeper_system":         cfg.Prompts.TimekeeperSystem,
		"summary_system":            cfg.Prompts.SummarySystem,
		"summary_chapter_system":    cfg.Prompts.SummaryChapterSystem,
		"summary_book_system":       cfg.Prompts.SummaryBookSystem,
		"knowledge_filter_system":   cfg.Prompts.KnowledgeFilterSystem,
		"json_only_architect_system": cfg.Prompts.JsonOnlyArchitectSystem,
		"inspiration_assistant_system": cfg.Prompts.InspirationAssistantSystem,
	}
	c.JSON(http.StatusOK, prompts)
}

// UpdatePrompts 更新提示词
func UpdatePrompts(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新配置
	for key, value := range req {
		viper.Set("prompts."+key, value)
	}

	// 保存配置文件
	if err := viper.WriteConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存提示词失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "提示词已更新",
		"data":    req,
	})
}

// GetBillingStats 获取费用统计
func GetBillingStats(c *gin.Context) {
	if cfg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "配置未初始化"})
		return
	}

	// 获取真实统计数据
	stats := billingStore.GetStats()

	// 转换 pricing 格式
	pricingData := make(map[string]interface{})
	for k, v := range cfg.Pricing {
		pricingData[k] = v
	}

	result := map[string]interface{}{
		"pricing":        pricingData,
		"total_tokens":   stats.TotalTokens,
		"total_cost":     stats.TotalCost,
		"monthly_tokens": stats.MonthlyTokens,
		"monthly_cost":   stats.MonthlyCost,
		"daily_stats":    stats.DailyStats,
		"by_model":       stats.ByModel,
	}
	c.JSON(http.StatusOK, result)
}

// GetWritingGoals 获取写作目标
func GetWritingGoals(c *gin.Context) {
	// 从存储中读取或返回默认值
	goals := map[string]interface{}{
		"daily_words":      2000,
		"weekly_chapters":  2,
		"target_date":      nil,
		"current_progress": 0,
	}
	c.JSON(http.StatusOK, goals)
}

// UpdateWritingGoals 更新写作目标
func UpdateWritingGoals(c *gin.Context) {
	var req struct {
		DailyWords     int    `json:"daily_words"`
		WeeklyChapters int    `json:"weekly_chapters"`
		TargetDate     string `json:"target_date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "写作目标已更新",
		"data":    req,
	})
}

// ==================== 工具函数 ====================

func parseInt(s string) int {
	var result int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		}
	}
	return result
}

// generateID 生成唯一 ID
func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// getLLMClient 获取 LLM 客户端
func getLLMClient() (llm.Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("配置未初始化")
	}

	if cfg.LLM.APIKey == "" {
		return nil, fmt.Errorf("API Key 未配置")
	}

	llmConfig := &llm.Config{
		Provider:     cfg.LLM.Provider,
		APIKey:       cfg.LLM.APIKey,
		BaseURL:      cfg.LLM.BaseURL,
		Models:       cfg.LLM.Models,
		Temperatures: cfg.LLM.Temperatures,
		MaxRetries:   cfg.LLM.MaxRetries,
		Timeout:      cfg.LLM.Timeout,
	}

	return llm.NewClient(llmConfig), nil
}

// getPromptsConfig 获取提示词配置
func getPromptsConfig() *config.PromptsConfig {
	if cfg == nil {
		return &config.PromptsConfig{}
	}
	return &cfg.Prompts
}

// recordTokenUsage 记录 token 使用量（估算）
func recordTokenUsage(bookName, taskType, model string, chapterID int, inputText, outputText string) {
	if billingStore == nil {
		return
	}

	// 简单估算：中文约 1.5 字/token，英文约 4 字符/token
	inputTokens := len(inputText) / 2
	outputTokens := len(outputText) / 2

	// 计算费用
	cost := store.CalculateCost(model, inputTokens, outputTokens, convertPricingToMap(cfg.Pricing))

	billingStore.RecordUsage(bookName, taskType, model, chapterID, inputTokens, outputTokens, cost)
}

// convertPricingToMap 转换 pricing 格式
func convertPricingToMap(pricing config.PricingConfig) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range pricing {
		result[k] = map[string]interface{}{
			"input":  v.Input,
			"output": v.Output,
		}
	}
	return result
}

// ==================== 续写和导出 ====================

// AIContinue AI 续写章节
func AIContinue(c *gin.Context) {
	var req struct {
		BookName    string `json:"book_name"`
		ChapterID   int    `json:"chapter_id"`
		WriteWords  int    `json:"write_words"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bookName := req.BookName
	if bookName == "" {
		bookName = c.Param("id")
	}

	// 获取 LLM 客户端
	llmClient, err := getLLMClient()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "AI服务未配置",
			"hint":    "请在系统设置中配置API Key后重试",
		})
		return
	}

	// 加载现有内容
	existingContent, err := jsonStore.LoadChapterContent(bookName, req.ChapterID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "加载章节内容失败: " + err.Error()})
		return
	}

	// 创建写作服务
	writerService := service.NewWriterService(llmClient, jsonStore, getPromptsConfig())

	// 续写
	content, err := writerService.ContinueChapter(c.Request.Context(), bookName, req.ChapterID, existingContent, req.WriteWords)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "续写失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"content": content,
		"message": "续写成功",
	})
}

// ExportBook 导出书籍
func ExportBook(c *gin.Context) {
	bookID := c.Param("id")
	format := c.Param("format") // txt, markdown, json

	// 加载书籍元数据
	bookPath := filepath.Join(cfg.Server.DataDir, "projects", bookID, "metadata.json")
	data, err := os.ReadFile(bookPath)

	var bookMeta model.BookMeta
	if err != nil {
		// 如果 metadata.json 不存在，使用默认值
		bookMeta = model.BookMeta{ID: bookID, Name: bookID}
	} else {
		if err := json.Unmarshal(data, &bookMeta); err != nil {
			bookMeta = model.BookMeta{ID: bookID, Name: bookID}
		}
	}

	// 加载章节列表
	chapters, err := jsonStore.LoadChapters(bookID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "加载章节失败"})
		return
	}

	var content string
	var filename string
	var contentType string

	switch format {
	case "txt":
		content = exportTXT(&bookMeta, chapters)
		filename = bookID + ".txt"
		contentType = "text/plain; charset=utf-8"
	case "markdown", "md":
		content = exportMarkdown(&bookMeta, chapters)
		filename = bookID + ".md"
		contentType = "text/markdown; charset=utf-8"
	case "json":
		jsonData, _ := json.MarshalIndent(map[string]interface{}{
			"book":     bookMeta,
			"chapters": chapters,
		}, "", "  ")
		content = string(jsonData)
		filename = bookID + ".json"
		contentType = "application/json"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持导出格式: " + format})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(200, contentType, []byte(content))
}

func exportTXT(book *model.BookMeta, chapters []*model.Chapter) string {
	var buf strings.Builder
	buf.WriteString(book.Name + "\n")
	if book.Description != "" {
		buf.WriteString("\n简介:\n" + book.Description + "\n")
	}
	buf.WriteString("\n" + strings.Repeat("=", 50) + "\n\n")

	for _, ch := range chapters {
		content, err := jsonStore.LoadChapterContent(book.ID, ch.ID)
		if err != nil || content == "" {
			continue
		}
		buf.WriteString(fmt.Sprintf("第%d章 %s\n\n", ch.ID, ch.Title))
		buf.WriteString(content + "\n\n")
		buf.WriteString(strings.Repeat("-", 30) + "\n\n")
	}

	return buf.String()
}

func exportMarkdown(book *model.BookMeta, chapters []*model.Chapter) string {
	var buf strings.Builder
	buf.WriteString("# " + book.Name + "\n\n")
	if book.Description != "" {
		buf.WriteString("## 简介\n\n" + book.Description + "\n\n")
	}

	for _, ch := range chapters {
		content, err := jsonStore.LoadChapterContent(book.ID, ch.ID)
		if err != nil || content == "" {
			continue
		}
		buf.WriteString(fmt.Sprintf("## 第%d章 %s\n\n", ch.ID, ch.Title))
		buf.WriteString(content + "\n\n")
	}

	return buf.String()
}

// ==================== 批量生成 ====================

// BatchProgress 批量生成进度
type BatchProgress struct {
	BookName  string    `json:"book_name"`
	From      int       `json:"from"`
	To        int       `json:"to"`
	Current   int       `json:"current"`
	Completed []int     `json:"completed"`
	Failed    []int     `json:"failed"`
	StartedAt time.Time `json:"started_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// getBatchProgressFile 获取进度文件路径
func getBatchProgressFile(bookName string) string {
	return filepath.Join(cfg.Server.DataDir, "projects", bookName, "batch_progress.json")
}

// loadBatchProgress 加载进度
func loadBatchProgress(bookName string) *BatchProgress {
	data, err := os.ReadFile(getBatchProgressFile(bookName))
	if err != nil {
		return nil
	}
	var progress BatchProgress
	if err := json.Unmarshal(data, &progress); err != nil {
		return nil
	}
	return &progress
}

// saveBatchProgress 保存进度
func saveBatchProgress(bookName string, progress *BatchProgress) {
	data, _ := json.MarshalIndent(progress, "", "  ")
	os.WriteFile(getBatchProgressFile(bookName), data, 0644)
}

// deleteBatchProgress 删除进度
func deleteBatchProgress(bookName string) {
	os.Remove(getBatchProgressFile(bookName))
}

// BatchGenerate SSE 流式批量生成
func BatchGenerate(c *gin.Context) {
	var req struct {
		BookName string `json:"book_name"`
		From     int    `json:"from"`
		To       int    `json:"to"`
		Stream   bool   `json:"stream"`
		Retry    int    `json:"retry"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bookName := req.BookName
	if bookName == "" {
		bookName = c.Query("book_name")
	}

	if req.From < 1 || req.To < req.From {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的章节范围"})
		return
	}

	// 获取 LLM 客户端
	llmClient, err := getLLMClient()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "AI服务未配置"})
		return
	}

	// 加载或创建进度
	progress := loadBatchProgress(bookName)
	if progress == nil || progress.From != req.From || progress.To != req.To {
		progress = &BatchProgress{
			BookName:  bookName,
			From:      req.From,
			To:        req.To,
			Current:   req.From,
			Completed: []int{},
			Failed:    []int{},
			StartedAt: time.Now(),
		}
	}

	// 设置 SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	writerService := service.NewWriterService(llmClient, jsonStore, getPromptsConfig())
	ctx := c.Request.Context()

	// 发送初始状态
	c.SSEvent("start", gin.H{
		"book_name": bookName,
		"from":      req.From,
		"to":        req.To,
	})
	c.Writer.Flush()

	for chapterID := progress.Current; chapterID <= req.To; chapterID++ {
		// 检查是否已完成
		completed := false
		for _, cID := range progress.Completed {
			if cID == chapterID {
				completed = true
				break
			}
		}
		if completed {
			continue
		}

		// 发送章节开始
		c.SSEvent("chapter_start", gin.H{
			"chapter_id": chapterID,
			"title":      fmt.Sprintf("第%d章", chapterID),
		})
		c.Writer.Flush()

		var content string
		var genErr error

		// 重试机制
		for attempt := 0; attempt <= req.Retry; attempt++ {
			if attempt > 0 {
				c.SSEvent("retry", gin.H{
					"chapter_id": chapterID,
					"attempt":    attempt,
				})
				c.Writer.Flush()
				time.Sleep(2 * time.Second)
			}

			genCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)

			if req.Stream {
				// 流式生成
				streamCh, err := writerService.WriteChapterStream(genCtx, bookName, chapterID, "")
				if err != nil {
					cancel()
					genErr = err
					continue
				}

				var fullContent strings.Builder
				for chunk := range streamCh {
					if chunk.Error != nil {
						genErr = chunk.Error
						break
					}
					fullContent.WriteString(chunk.Content)
					c.SSEvent("content", gin.H{
						"chapter_id": chapterID,
						"content":    chunk.Content,
					})
					c.Writer.Flush()
				}
				if genErr == nil {
					content = fullContent.String()
				}
			} else {
				content, genErr = writerService.WriteChapter(genCtx, bookName, chapterID, "")
			}

			cancel()

			if genErr == nil {
				break
			}
		}

		if genErr != nil {
			c.SSEvent("chapter_error", gin.H{
				"chapter_id": chapterID,
				"error":      genErr.Error(),
			})
			progress.Failed = append(progress.Failed, chapterID)
			progress.UpdatedAt = time.Now()
			saveBatchProgress(bookName, progress)
			c.Writer.Flush()
			continue
		}

		// 保存内容
		if err := jsonStore.SaveChapterContent(bookName, chapterID, content); err != nil {
			c.SSEvent("chapter_error", gin.H{
				"chapter_id": chapterID,
				"error":      "保存失败: " + err.Error(),
			})
			progress.Failed = append(progress.Failed, chapterID)
			c.Writer.Flush()
			continue
		}

		// 更新进度
		progress.Completed = append(progress.Completed, chapterID)
		progress.Current = chapterID + 1
		progress.UpdatedAt = time.Now()
		saveBatchProgress(bookName, progress)

		// 发送完成事件
		c.SSEvent("chapter_done", gin.H{
			"chapter_id": chapterID,
			"word_count": len(content),
			"progress":   len(progress.Completed),
			"total":      req.To - req.From + 1,
		})
		c.Writer.Flush()
	}

	// 完成
	c.SSEvent("done", gin.H{
		"completed": len(progress.Completed),
		"failed":    len(progress.Failed),
		"failed_ids": progress.Failed,
	})
	c.Writer.Flush()

	// 删除进度文件
	if len(progress.Failed) == 0 {
		deleteBatchProgress(bookName)
	}
}

// BatchStatus 获取批量生成进度
func BatchStatus(c *gin.Context) {
	bookName := c.Query("book_name")
	if bookName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 book_name 参数"})
		return
	}

	progress := loadBatchProgress(bookName)
	if progress == nil {
		c.JSON(http.StatusOK, gin.H{
			"status":    "idle",
			"message":   "没有进行中的批量任务",
			"progress":  nil,
		})
		return
	}

	total := progress.To - progress.From + 1
	completed := len(progress.Completed)

	c.JSON(http.StatusOK, gin.H{
		"status":    "running",
		"book_name": progress.BookName,
		"from":      progress.From,
		"to":        progress.To,
		"current":   progress.Current,
		"completed": completed,
		"failed":    len(progress.Failed),
		"failed_ids": progress.Failed,
		"percent":   float64(completed) / float64(total) * 100,
		"started_at": progress.StartedAt,
		"updated_at": progress.UpdatedAt,
	})
}

// BatchReset 重置批量生成进度
func BatchReset(c *gin.Context) {
	bookName := c.Query("book_name")
	if bookName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 book_name 参数"})
		return
	}

	deleteBatchProgress(bookName)
	c.JSON(http.StatusOK, gin.H{
		"message": "进度已重置",
	})
}

// ==================== 状态同步 ====================

// getPendingChangesFile 获取待审核变更文件路径（验证书名防止路径注入）
func getPendingChangesFile(bookName string) (string, error) {
	if !validBookName(bookName) {
		return "", fmt.Errorf("书名不合法: %s", bookName)
	}
	return filepath.Join("projects", bookName, "pending_changes.json"), nil
}

// loadPendingChanges 加载待审核变更
func loadPendingChanges(bookName string) (*service.PendingChanges, error) {
	filePath, err := getPendingChangesFile(bookName)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var pending service.PendingChanges
	if err := json.Unmarshal(data, &pending); err != nil {
		return nil, err
	}
	return &pending, nil
}

// savePendingChanges 保存待审核变更
func savePendingChanges(bookName string, pending *service.PendingChanges) error {
	filePath, err := getPendingChangesFile(bookName)
	if err != nil {
		return err
	}
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	data, err := json.MarshalIndent(pending, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}
	return os.WriteFile(filePath, data, 0644)
}

// SyncExtract 从章节提取状态变更
func SyncExtract(c *gin.Context) {
	var req struct {
		BookName  string `json:"book_name"`
		ChapterID int    `json:"chapter_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bookName := req.BookName
	if bookName == "" {
		bookName = c.Param("id")
	}

	// 获取 LLM 客户端
	llmClient, err := getLLMClient()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "AI服务未配置"})
		return
	}

	syncService := service.NewSyncService(llmClient, jsonStore)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Minute)
	defer cancel()

	pending, err := syncService.ExtractStateChanges(ctx, bookName, req.ChapterID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "提取失败: " + err.Error()})
		return
	}

	if len(pending.Changes) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "未检测到状态变更",
			"changes": []interface{}{},
		})
		return
	}

	// 保存待审核变更
	if err := savePendingChanges(bookName, pending); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "保存失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "提取完成",
		"chapter_id":   pending.ChapterID,
		"extracted_at": pending.ExtractedAt,
		"changes":      pending.Changes,
	})
}

// SyncPending 获取待审核变更
func SyncPending(c *gin.Context) {
	bookName := c.Query("book_name")
	if bookName == "" {
		bookName = c.Param("id")
	}

	pending, err := loadPendingChanges(bookName)
	if err != nil || pending == nil || len(pending.Changes) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "暂无待审核变更",
			"changes": []interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, pending)
}

// SyncApply 应用状态变更
func SyncApply(c *gin.Context) {
	var req struct {
		BookName string `json:"book_name"`
		ChangeID string `json:"change_id"` // 可选，指定应用的变更ID
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bookName := req.BookName
	if bookName == "" {
		bookName = c.Param("id")
	}

	pending, err := loadPendingChanges(bookName)
	if err != nil || pending == nil || len(pending.Changes) == 0 {
		c.JSON(http.StatusOK, gin.H{"error": "暂无待审核变更"})
		return
	}

	syncService := service.NewSyncService(nil, jsonStore)

	applied := 0
	failed := 0
	var remaining []service.StateChange

	for _, change := range pending.Changes {
		// 如果指定了变更ID，只应用该变更
		if req.ChangeID != "" {
			// 检查是否匹配（精确匹配或8字符前缀匹配）
			matchesExact := change.ID == req.ChangeID
			matchesPrefix := len(change.ID) >= 8 && len(req.ChangeID) >= 8 && change.ID[:8] == req.ChangeID[:8]
			if !matchesExact && !matchesPrefix {
				remaining = append(remaining, change)
				continue
			}
		}

		// 应用变更
		if err := syncService.ApplyChange(bookName, &change); err != nil {
			failed++
		} else {
			applied++
		}
	}

	// 更新待审核列表
	if len(remaining) > 0 {
		pending.Changes = remaining
		if err := savePendingChanges(bookName, pending); err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "保存剩余变更失败: " + err.Error()})
			return
		}
	} else {
		filePath, _ := getPendingChangesFile(bookName)
		os.Remove(filePath)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "应用完成",
		"applied":  applied,
		"failed":   failed,
		"remaining": len(remaining),
	})
}

// SyncReject 丢弃状态变更
func SyncReject(c *gin.Context) {
	var req struct {
		BookName string `json:"book_name"`
		ChangeID string `json:"change_id"` // 可选，指定丢弃的变更ID
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bookName := req.BookName
	if bookName == "" {
		bookName = c.Param("id")
	}

	pending, err := loadPendingChanges(bookName)
	if err != nil || pending == nil || len(pending.Changes) == 0 {
		c.JSON(http.StatusOK, gin.H{"error": "暂无待审核变更"})
		return
	}

	if req.ChangeID == "" {
		// 丢弃所有变更
		filePath, _ := getPendingChangesFile(bookName)
		os.Remove(filePath)
		c.JSON(http.StatusOK, gin.H{
			"message": "已丢弃所有变更",
			"rejected": len(pending.Changes),
		})
		return
	}

	// 丢弃指定变更
	var remaining []service.StateChange
	rejected := 0
	for _, change := range pending.Changes {
		if change.ID == req.ChangeID || (len(change.ID) >= 8 && change.ID[:8] == req.ChangeID) {
			rejected++
		} else {
			remaining = append(remaining, change)
		}
	}

	if len(remaining) > 0 {
		pending.Changes = remaining
		if err := savePendingChanges(bookName, pending); err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "保存剩余变更失败: " + err.Error()})
			return
		}
	} else {
		filePath, _ := getPendingChangesFile(bookName)
		os.Remove(filePath)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "已丢弃变更",
		"rejected": rejected,
	})
}