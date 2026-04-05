package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"ai-writer/internal/llm"
)

// TextChunk 文本分块
type TextChunk struct {
	ID        string    `json:"id"`
	BookName  string    `json:"book_name"`
	ChapterID int       `json:"chapter_id"`
	ChunkIndex int      `json:"chunk_index"`
	StartPos  int       `json:"start_pos"`
	EndPos    int       `json:"end_pos"`
	Content   string    `json:"content"`
	Embedding []float64 `json:"embedding"`
}

// VectorIndex 向量索引
type VectorIndex struct {
	BookName   string      `json:"book_name"`
	ChunkSize  int         `json:"chunk_size"`
	Overlap    int         `json:"overlap"`
	Dimensions int         `json:"dimensions"`
	ChunkCount int         `json:"chunk_count"`
	Chunks     []TextChunk `json:"chunks"`
}

// VectorStore 向量存储
type VectorStore struct {
	basePath string
	mu       sync.RWMutex
}

// NewVectorStore 创建向量存储
func NewVectorStore(basePath string) *VectorStore {
	return &VectorStore{
		basePath: basePath,
	}
}

// ensureDir 确保目录存在
func (s *VectorStore) ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// getIndexPath 获取索引文件路径
func (s *VectorStore) getIndexPath(bookName string) string {
	return filepath.Join(s.basePath, bookName, "index.json")
}

// IndexChapter 索引章节内容
func (s *VectorStore) IndexChapter(ctx context.Context, embeddingClient llm.EmbeddingClient, bookName string, chapterID int, content string, chunkSize, overlap int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 加载现有索引
	index, err := s.loadIndex(bookName)
	if err != nil {
		index = &VectorIndex{
			BookName:  bookName,
			ChunkSize: chunkSize,
			Overlap:   overlap,
			Chunks:    []TextChunk{},
		}
	}

	// 删除该章节的旧分块
	var newChunks []TextChunk
	for _, chunk := range index.Chunks {
		if chunk.ChapterID != chapterID {
			newChunks = append(newChunks, chunk)
		}
	}
	index.Chunks = newChunks

	// 分块
	chunks := s.splitText(content, chapterID, chunkSize, overlap)

	// 为每个分块生成 embedding
	for i := range chunks {
		if len(strings.TrimSpace(chunks[i].Content)) < 50 {
			continue // 跳过太短的分块
		}

		embedding, err := embeddingClient.GetEmbedding(ctx, chunks[i].Content)
		if err != nil {
			fmt.Printf("生成 embedding 失败 (章节 %d, 分块 %d): %v\n", chapterID, i, err)
			continue
		}

		chunks[i].Embedding = embedding
		if index.Dimensions == 0 {
			index.Dimensions = len(embedding)
		}

		index.Chunks = append(index.Chunks, chunks[i])
	}

	index.ChunkCount = len(index.Chunks)

	// 保存索引
	return s.saveIndex(bookName, index)
}

// splitText 分割文本
func (s *VectorStore) splitText(content string, chapterID int, chunkSize, overlap int) []TextChunk {
	var chunks []TextChunk

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
				chunks = append(chunks, TextChunk{
					ID:         generateChunkID(),
					BookName:   "",
					ChapterID:  chapterID,
					ChunkIndex: chunkIndex,
					StartPos:   currentStart,
					EndPos:     currentStart + currentChunk.Len(),
					Content:    currentChunk.String(),
				})
				chunkIndex++
			}

			// 开始新块（考虑重叠）
			if overlap > 0 && currentChunk.Len() > overlap {
				// 取最后 overlap 字符作为重叠部分
				overlapText := currentChunk.String()
				if len(overlapText) > overlap {
					overlapText = overlapText[len(overlapText)-overlap:]
				}
				currentChunk.Reset()
				currentChunk.WriteString(overlapText)
				currentStart = currentPos - overlap
			} else {
				currentChunk.Reset()
				currentStart = currentPos
			}
			currentChunk.WriteString(para)
		}

		currentPos += len(para) + 2 // +2 for \n\n
	}

	// 保存最后一个块
	if currentChunk.Len() >= 50 {
		chunks = append(chunks, TextChunk{
			ID:         generateChunkID(),
			BookName:   "",
			ChapterID:  chapterID,
			ChunkIndex: chunkIndex,
			StartPos:   currentStart,
			EndPos:     currentStart + currentChunk.Len(),
			Content:    currentChunk.String(),
		})
	}

	return chunks
}

// Search 搜索相似内容
func (s *VectorStore) Search(queryEmbedding []float64, bookName string, topK int) ([]TextChunk, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	index, err := s.loadIndex(bookName)
	if err != nil {
		return nil, err
	}

	if len(index.Chunks) == 0 {
		return []TextChunk{}, nil
	}

	// 计算相似度
	type scoredChunk struct {
		chunk    TextChunk
		score    float64
	}

	var scored []scoredChunk
	for _, chunk := range index.Chunks {
		if len(chunk.Embedding) == 0 {
			continue
		}

		score := cosineSimilarity(queryEmbedding, chunk.Embedding)
		scored = append(scored, scoredChunk{chunk: chunk, score: score})
	}

	// 按相似度排序
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// 返回 topK
	if topK > len(scored) {
		topK = len(scored)
	}

	results := make([]TextChunk, topK)
	for i := 0; i < topK; i++ {
		results[i] = scored[i].chunk
	}

	return results, nil
}

// DeleteChapter 删除章节的向量
func (s *VectorStore) DeleteChapter(bookName string, chapterID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	index, err := s.loadIndex(bookName)
	if err != nil {
		return nil // 索引不存在，无需删除
	}

	// 过滤掉该章节的分块
	var newChunks []TextChunk
	for _, chunk := range index.Chunks {
		if chunk.ChapterID != chapterID {
			newChunks = append(newChunks, chunk)
		}
	}
	index.Chunks = newChunks
	index.ChunkCount = len(index.Chunks)

	return s.saveIndex(bookName, index)
}

// DeleteBook 删除书籍的向量索引
func (s *VectorStore) DeleteBook(bookName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	bookPath := filepath.Join(s.basePath, bookName)
	return os.RemoveAll(bookPath)
}

// GetStatus 获取向量库状态
func (s *VectorStore) GetStatus(bookName string) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	index, err := s.loadIndex(bookName)
	if err != nil {
		return map[string]interface{}{
			"exists":     false,
			"book_name":  bookName,
			"chunk_count": 0,
		}, nil
	}

	// 统计各章节的分块数量
	chapterChunks := make(map[int]int)
	for _, chunk := range index.Chunks {
		chapterChunks[chunk.ChapterID]++
	}

	return map[string]interface{}{
		"exists":        true,
		"book_name":     bookName,
		"chunk_count":   index.ChunkCount,
		"chunk_size":    index.ChunkSize,
		"overlap":       index.Overlap,
		"dimensions":    index.Dimensions,
		"chapter_chunks": chapterChunks,
	}, nil
}

// loadIndex 加载索引
func (s *VectorStore) loadIndex(bookName string) (*VectorIndex, error) {
	path := s.getIndexPath(bookName)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var index VectorIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, err
	}

	return &index, nil
}

// saveIndex 保存索引
func (s *VectorStore) saveIndex(bookName string, index *VectorIndex) error {
	bookPath := filepath.Join(s.basePath, bookName)
	if err := s.ensureDir(bookPath); err != nil {
		return err
	}

	path := s.getIndexPath(bookName)
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// cosineSimilarity 计算余弦相似度
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// generateChunkID 生成分块 ID
func generateChunkID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}