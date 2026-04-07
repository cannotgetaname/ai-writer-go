package store

import (
	"context"
	"database/sql"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"

	_ "modernc.org/sqlite"
)

// sqliteScoredChunk 用于搜索结果的排序
type sqliteScoredChunk struct {
	chunk TextChunk
	score float64
}

// SQLiteVectorDB SQLite 向量存储实现
// 使用 pure Go SQLite 驱动，在 Go 中实现向量相似度搜索
type SQLiteVectorDB struct {
	basePath    string
	mu          sync.RWMutex
	connections map[string]*sql.DB // bookName -> db connection
}

// NewSQLiteVectorDB 创建 SQLite 向量存储
func NewSQLiteVectorDB(basePath string) (*SQLiteVectorDB, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create vector db directory: %w", err)
	}

	return &SQLiteVectorDB{
		basePath:    basePath,
		connections: make(map[string]*sql.DB),
	}, nil
}

// getDB 获取或创建数据库连接
func (s *SQLiteVectorDB) getDB(bookName string) (*sql.DB, error) {
	s.mu.RLock()
	db, exists := s.connections[bookName]
	s.mu.RUnlock()

	if exists {
		return db, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 双重检查
	if db, exists = s.connections[bookName]; exists {
		return db, nil
	}

	dbPath := filepath.Join(s.basePath, bookName+".db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(1) // SQLite 单写限制
	db.SetMaxIdleConns(1)

	// 创建向量存储表
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS vec_chunks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chapter_id INTEGER NOT NULL,
		chunk_index INTEGER NOT NULL,
		content TEXT NOT NULL,
		embedding BLOB NOT NULL,
		dimensions INTEGER NOT NULL DEFAULT 768,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_vec_chunks_chapter ON vec_chunks(chapter_id);
	CREATE INDEX IF NOT EXISTS idx_vec_chunks_book_chapter ON vec_chunks(chapter_id, chunk_index);
	`
	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create vector table: %w", err)
	}

	s.connections[bookName] = db
	return db, nil
}

// serializeEmbedding 将 float64 向量序列化为 BLOB
func serializeEmbedding(embedding []float64) []byte {
	// 转换为 float32 以节省空间
	buf := make([]byte, len(embedding)*4)
	for i, v := range embedding {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(float32(v)))
	}
	return buf
}

// deserializeEmbedding 从 BLOB 反序列化为 float64 向量
func deserializeEmbedding(data []byte, dimensions int) []float64 {
	embedding := make([]float64, dimensions)
	for i := 0; i < dimensions; i++ {
		embedding[i] = float64(math.Float32frombits(binary.LittleEndian.Uint32(data[i*4:])))
	}
	return embedding
}

// Index 索引文本分块
func (s *SQLiteVectorDB) Index(ctx context.Context, bookName string, chunks []TextChunk) error {
	db, err := s.getDB(bookName)
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO vec_chunks(chapter_id, chunk_index, content, embedding, dimensions)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, chunk := range chunks {
		if len(chunk.Embedding) == 0 {
			continue
		}

		embeddingBlob := serializeEmbedding(chunk.Embedding)
		dimensions := len(chunk.Embedding)

		_, err = stmt.ExecContext(ctx, chunk.ChapterID, chunk.ChunkIndex, chunk.Content, embeddingBlob, dimensions)
		if err != nil {
			return fmt.Errorf("failed to insert chunk (chapter %d, index %d): %w", chunk.ChapterID, chunk.ChunkIndex, err)
		}
	}

	return tx.Commit()
}

// Search 搜索相似内容
func (s *SQLiteVectorDB) Search(ctx context.Context, bookName string, queryEmbedding []float64, topK int) ([]TextChunk, error) {
	db, err := s.getDB(bookName)
	if err != nil {
		return nil, err
	}

	// 查询所有分块，在 Go 中计算相似度
	query := `
		SELECT id, chapter_id, chunk_index, content, embedding, dimensions
		FROM vec_chunks
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query chunks: %w", err)
	}
	defer rows.Close()

	var scored []sqliteScoredChunk
	for rows.Next() {
		var id int
		var chapterID, chunkIndex, dimensions int
		var content string
		var embeddingBlob []byte

		if err := rows.Scan(&id, &chapterID, &chunkIndex, &content, &embeddingBlob, &dimensions); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		embedding := deserializeEmbedding(embeddingBlob, dimensions)

		// 计算余弦相似度
		score := sqliteCosineSimilarity(queryEmbedding, embedding)
		if score > 0.3 { // 只保留相似度大于 0.3 的结果
			scored = append(scored, sqliteScoredChunk{
				chunk: TextChunk{
					ID:         fmt.Sprintf("%d", id),
					ChapterID:  chapterID,
					ChunkIndex: chunkIndex,
					Content:    content,
					Embedding:  embedding,
				},
				score: score,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
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

// sqliteCosineSimilarity 计算余弦相似度
func sqliteCosineSimilarity(a, b []float64) float64 {
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

// DeleteBook 删除书籍的所有向量
func (s *SQLiteVectorDB) DeleteBook(ctx context.Context, bookName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if db, exists := s.connections[bookName]; exists {
		db.Close()
		delete(s.connections, bookName)
	}

	dbPath := filepath.Join(s.basePath, bookName+".db")
	return os.Remove(dbPath)
}

// DeleteChapter 删除章节的向量
func (s *SQLiteVectorDB) DeleteChapter(ctx context.Context, bookName string, chapterID int) error {
	db, err := s.getDB(bookName)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "DELETE FROM vec_chunks WHERE chapter_id = ?", chapterID)
	return err
}

// GetStatus 获取状态信息
func (s *SQLiteVectorDB) GetStatus(ctx context.Context, bookName string) (map[string]interface{}, error) {
	dbPath := filepath.Join(s.basePath, bookName+".db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return map[string]interface{}{
			"exists":      false,
			"book_name":   bookName,
			"chunk_count": 0,
		}, nil
	}

	db, err := s.getDB(bookName)
	if err != nil {
		return nil, err
	}

	var chunkCount int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM vec_chunks").Scan(&chunkCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count chunks: %w", err)
	}

	// 统计各章节数量
	rows, err := db.QueryContext(ctx, "SELECT chapter_id, COUNT(*) FROM vec_chunks GROUP BY chapter_id")
	if err != nil {
		return nil, fmt.Errorf("failed to group by chapter: %w", err)
	}
	defer rows.Close()

	chapterChunks := make(map[int]int)
	for rows.Next() {
		var chapterID, count int
		if err := rows.Scan(&chapterID, &count); err != nil {
			return nil, err
		}
		chapterChunks[chapterID] = count
	}

	// 获取向量维度信息
	var avgDimensions float64
	err = db.QueryRowContext(ctx, "SELECT AVG(dimensions) FROM vec_chunks").Scan(&avgDimensions)
	if err != nil {
		avgDimensions = 768 // 默认维度
	}

	return map[string]interface{}{
		"exists":         true,
		"book_name":      bookName,
		"chunk_count":    chunkCount,
		"chapter_chunks": chapterChunks,
		"dimensions":     int(avgDimensions),
		"backend":        "sqlite-pure-go",
	}, nil
}

// Close 关闭所有连接
func (s *SQLiteVectorDB) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var lastErr error
	for _, db := range s.connections {
		if err := db.Close(); err != nil {
			lastErr = err
		}
	}
	s.connections = make(map[string]*sql.DB)
	return lastErr
}