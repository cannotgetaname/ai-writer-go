package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PythonEmbeddingClient 调用独立 Python embedding 服务
type PythonEmbeddingClient struct {
	baseURL    string
	httpClient *http.Client
}

// pythonEmbedRequest Python 服务请求
type pythonEmbedRequest struct {
	Texts []string `json:"texts"`
}

// pythonEmbedResponse Python 服务响应
type pythonEmbedResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
	Model      string      `json:"model"`
	Dimension  int         `json:"dimension"`
}

// pythonHealthResponse Python 健康检查响应
type pythonHealthResponse struct {
	Status      string `json:"status"`
	ModelLoaded bool   `json:"model_loaded"`
}

// NewPythonEmbeddingClient 创建 Python embedding 客户端
// portFile: 端口文件路径，如 ./embedding_port.txt
func NewPythonEmbeddingClient(portFile string) *PythonEmbeddingClient {
	baseURL := readPortFromFile(portFile)
	return &PythonEmbeddingClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// NewPythonEmbeddingClientFromURL 直接从 URL 创建客户端
func NewPythonEmbeddingClientFromURL(baseURL string) *PythonEmbeddingClient {
	return &PythonEmbeddingClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// readPortFromFile 从端口文件读取端口并构建 URL
func readPortFromFile(portFile string) string {
	// 尝试多个路径
	paths := []string{
		portFile,
		"./embedding_port.txt",
		filepath.Join(filepath.Dir(os.Args[0]), "embedding_port.txt"),
	}

	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err == nil {
			port := strings.TrimSpace(string(content))
			return fmt.Sprintf("http://127.0.0.1:%s", port)
		}
	}

	// 默认端口
	return "http://127.0.0.1:8082"
}

// WaitForReady 等待服务就绪
func (c *PythonEmbeddingClient) WaitForReady(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("embedding service not ready after %v", timeout)
		case <-ticker.C:
			if c.checkHealth(ctx) {
				return nil
			}
		}
	}
}

// checkHealth 检查健康状态
func (c *PythonEmbeddingClient) checkHealth(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
	if err != nil {
		return false
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	var health pythonHealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return false
	}

	return health.Status == "ok" && health.ModelLoaded
}

// GetEmbedding 获取单个文本的向量表示
func (c *PythonEmbeddingClient) GetEmbedding(ctx context.Context, text string) ([]float64, error) {
	req := pythonEmbedRequest{Texts: []string{text}}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/embed", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("Python embedding error: %s - %s", httpResp.Status, string(respBody))
	}

	var resp pythonEmbedResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(resp.Embeddings) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	return resp.Embeddings[0], nil
}

// GetEmbeddings 批量获取向量（可选优化）
func (c *PythonEmbeddingClient) GetEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	req := pythonEmbedRequest{Texts: texts}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/embed", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("Python embedding error: %s - %s", httpResp.Status, string(respBody))
	}

	var resp pythonEmbedResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return resp.Embeddings, nil
}