package store

import (
	"context"
)

// VectorDBClient 向量数据库客户端接口
// 定义向量数据库操作的抽象接口，支持多种后端实现（如内存存储、sqlite-vec、Milvus等）
type VectorDBClient interface {
	// Index 索引文本分块
	// 将文本分块及其向量嵌入存储到向量数据库中
	Index(ctx context.Context, bookName string, chunks []TextChunk) error

	// Search 搜索相似内容
	// 根据查询向量在指定书籍中搜索最相似的topK个文本分块
	Search(ctx context.Context, bookName string, queryEmbedding []float64, topK int) ([]TextChunk, error)

	// DeleteBook 删除书籍的所有向量
	// 删除指定书籍的所有向量数据
	DeleteBook(ctx context.Context, bookName string) error

	// DeleteChapter 删除章节的向量
	// 删除指定书籍中特定章节的所有向量数据
	DeleteChapter(ctx context.Context, bookName string, chapterID int) error

	// GetStatus 获取状态信息
	// 返回向量数据库的状态信息，如分块数量、维度等
	GetStatus(ctx context.Context, bookName string) (map[string]interface{}, error)

	// Close 关闭连接
	// 释放资源并关闭数据库连接
	Close() error
}