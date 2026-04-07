package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TEIEmbeddingClient TEI (Text Embeddings Inference) 客户端
type TEIEmbeddingClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewTEIEmbeddingClient 创建 TEI 客户端
func NewTEIEmbeddingClient(baseURL string) *TEIEmbeddingClient {
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8081"
	}

	return &TEIEmbeddingClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// TEIEmbedRequest TEI 请求
type TEIEmbedRequest struct {
	Inputs []string `json:"inputs"`
}

// TEIEmbedResponse TEI 响应
type TEIEmbedResponse [][]float64

// GetEmbedding 获取单个文本的向量
func (c *TEIEmbeddingClient) GetEmbedding(ctx context.Context, text string) ([]float64, error) {
	embeddings, err := c.GetEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}
	return embeddings[0], nil
}

// GetEmbeddings 批量获取向量
func (c *TEIEmbeddingClient) GetEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	req := TEIEmbedRequest{Inputs: texts}

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
		return nil, fmt.Errorf("TEI error: %s - %s", httpResp.Status, string(respBody))
	}

	var resp TEIEmbedResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return resp, nil
}

// HealthCheck 检查服务是否可用
func (c *TEIEmbeddingClient) HealthCheck(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
	if err != nil {
		return err
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("TEI service not available: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("TEI service unhealthy: %s", httpResp.Status)
	}

	return nil
}