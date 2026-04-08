// internal/api/handler/vector_handler.go
package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"ai-writer/internal/llm"
	"ai-writer/internal/store"
)

var (
	vectorDBClient  store.VectorDBClient
	embeddingClient llm.EmbeddingClient
)

// InitVectorDB 初始化向量数据库客户端
func InitVectorDB(client store.VectorDBClient) {
	vectorDBClient = client
}

// InitEmbeddingClient 初始化 embedding 客户端
func InitEmbeddingClient(client llm.EmbeddingClient) {
	embeddingClient = client
}

// VectorIndexBook 索引整本书
func VectorIndexBook(c *gin.Context) {
	if vectorDBClient == nil || embeddingClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "向量服务未初始化"})
		return
	}

	var req struct {
		BookName string `json:"book_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !validBookName(req.BookName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "书名不合法"})
		return
	}

	// 获取所有章节
	chapters, err := jsonStore.LoadChapters(req.BookName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载章节失败: " + err.Error()})
		return
	}

	// 索引每个章节
	var totalChunks int
	for _, ch := range chapters {
		content, err := jsonStore.LoadChapterContent(req.BookName, ch.ID)
		if err != nil || content == "" {
			continue
		}

		// 分块
		chunks := splitText(content, ch.ID, cfg.VectorStore.ChunkSize, cfg.VectorStore.Overlap)

		// 生成向量
		for i := range chunks {
			if len(chunks[i].Content) < 50 {
				continue
			}

			embedding, err := embeddingClient.GetEmbedding(c.Request.Context(), chunks[i].Content)
			if err != nil {
				continue
			}
			chunks[i].Embedding = embedding
		}

		// 存储
		if err := vectorDBClient.Index(c.Request.Context(), req.BookName, chunks); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "索引章节失败: " + err.Error()})
			return
		}

		totalChunks += len(chunks)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "索引完成",
		"book_name":   req.BookName,
		"chunk_count": totalChunks,
	})
}

// VectorIndexChapter 索引单个章节
func VectorIndexChapter(c *gin.Context) {
	if vectorDBClient == nil || embeddingClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "向量服务未初始化"})
		return
	}

	var req struct {
		BookName  string `json:"book_name" binding:"required"`
		ChapterID int    `json:"chapter_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !validBookName(req.BookName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "书名不合法"})
		return
	}

	// 加载章节内容
	content, err := jsonStore.LoadChapterContent(req.BookName, req.ChapterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载章节失败: " + err.Error()})
		return
	}

	// 分块
	chunks := splitText(content, req.ChapterID, cfg.VectorStore.ChunkSize, cfg.VectorStore.Overlap)

	// 生成向量
	for i := range chunks {
		if len(chunks[i].Content) < 50 {
			continue
		}

		embedding, err := embeddingClient.GetEmbedding(c.Request.Context(), chunks[i].Content)
		if err != nil {
			continue
		}
		chunks[i].Embedding = embedding
	}

	// 先删除旧数据
	vectorDBClient.DeleteChapter(c.Request.Context(), req.BookName, req.ChapterID)

	// 存储新数据
	if err := vectorDBClient.Index(c.Request.Context(), req.BookName, chunks); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "索引失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "索引完成",
		"chunk_count": len(chunks),
	})
}

// VectorSearch 搜索相似内容
func VectorSearch(c *gin.Context) {
	if vectorDBClient == nil || embeddingClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "向量服务未初始化"})
		return
	}

	var req struct {
		BookName string `json:"book_name" binding:"required"`
		Query    string `json:"query" binding:"required"`
		TopK     int    `json:"top_k"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !validBookName(req.BookName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "书名不合法"})
		return
	}

	if req.TopK == 0 {
		req.TopK = 5
	}

	// 生成查询向量
	queryEmbedding, err := embeddingClient.GetEmbedding(c.Request.Context(), req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成查询向量失败: " + err.Error()})
		return
	}

	// 搜索
	results, err := vectorDBClient.Search(c.Request.Context(), req.BookName, queryEmbedding, req.TopK)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results":   results,
		"book_name": req.BookName,
		"query":     req.Query,
	})
}

// VectorStatus 获取向量库状态
func VectorStatus(c *gin.Context) {
	if vectorDBClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "向量服务未初始化"})
		return
	}

	bookName := c.Query("book_name")
	if bookName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "book_name required"})
		return
	}

	if !validBookName(bookName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "书名不合法"})
		return
	}

	status, err := vectorDBClient.GetStatus(c.Request.Context(), bookName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// VectorDeleteBook 删除书籍向量
func VectorDeleteBook(c *gin.Context) {
	if vectorDBClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "向量服务未初始化"})
		return
	}

	bookName := c.Param("book_name")
	if bookName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "book_name required"})
		return
	}

	if !validBookName(bookName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "书名不合法"})
		return
	}

	if err := vectorDBClient.DeleteBook(c.Request.Context(), bookName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功", "book_name": bookName})
}

// splitText 分割文本
func splitText(content string, chapterID int, chunkSize, overlap int) []store.TextChunk {
	var chunks []store.TextChunk

	// 按段落分割
	paragraphs := strings.Split(content, "\n\n")

	currentPos := 0
	chunkIndex := 0
	var currentChunk strings.Builder
	currentStart := 0

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// 如果当前块加上新段落不超过限制，添加到当前块
		if currentChunk.Len()+len(para) <= chunkSize {
			if currentChunk.Len() > 0 {
				currentChunk.WriteString("\n\n")
			} else {
				currentStart = currentPos
			}
			currentChunk.WriteString(para)
		} else {
			// 保存当前块
			if currentChunk.Len() >= 50 {
				chunks = append(chunks, store.TextChunk{
					ID:         generateChunkID(),
					ChapterID:  chapterID,
					ChunkIndex: chunkIndex,
					StartPos:   currentStart,
					EndPos:     currentStart + currentChunk.Len(),
					Content:    currentChunk.String(),
				})
				chunkIndex++
			}

			// 开始新块
			currentChunk.Reset()
			currentStart = currentPos
			currentChunk.WriteString(para)
		}

		currentPos += len(para) + 2 // +2 for \n\n
	}

	// 保存最后一个块
	if currentChunk.Len() >= 50 {
		chunks = append(chunks, store.TextChunk{
			ID:         generateChunkID(),
			ChapterID:  chapterID,
			ChunkIndex: chunkIndex,
			StartPos:   currentStart,
			EndPos:     currentStart + currentChunk.Len(),
			Content:    currentChunk.String(),
		})
	}

	return chunks
}

// generateChunkID 生成唯一 ID
func generateChunkID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}