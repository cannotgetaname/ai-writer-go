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

// OllamaClient Ollama 本地模型客户端
type OllamaClient struct {
	config     *Config
	httpClient *http.Client
}

// NewOllamaClient 创建 Ollama 客户端
func NewOllamaClient(cfg *Config) *OllamaClient {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:11434"
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 300 // Ollama 本地模型可能较慢
	}

	return &OllamaClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

// OllamaRequest Ollama API 请求
type OllamaRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages,omitempty"`
	Prompt   string        `json:"prompt,omitempty"`    // 简单模式
	System   string        `json:"system,omitempty"`    // 系统提示词
	Stream   bool          `json:"stream"`
	Options  OllamaOptions `json:"options,omitempty"`
}

// OllamaOptions Ollama 选项
type OllamaOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumCtx      int     `json:"num_ctx,omitempty"`     // 上下文长度
	NumPredict  int     `json:"num_predict,omitempty"` // 最大生成 token
}

// OllamaResponse Ollama API 响应
type OllamaResponse struct {
	Model      string       `json:"model"`
	Message    *ChatMessage `json:"message,omitempty"`
	Response   string       `json:"response,omitempty"` // 简单模式
	Done       bool         `json:"done"`
	DoneReason string       `json:"done_reason,omitempty"`
	Context    []int        `json:"context,omitempty"`
}

// Call 调用 Ollama
func (c *OllamaClient) Call(ctx context.Context, prompt string, taskType string) (string, error) {
	return c.CallWithSystem(ctx, "", prompt, taskType)
}

// CallWithSystem 使用系统提示词调用
func (c *OllamaClient) CallWithSystem(ctx context.Context, systemPrompt, userPrompt string, taskType string) (string, error) {
	model := c.config.GetModel(taskType)
	temp := c.config.GetTemperature(taskType)

	req := OllamaRequest{
		Model:  model,
		Stream: false,
		Options: OllamaOptions{
			Temperature: temp,
		},
	}

	// 使用 chat 模式
	if systemPrompt != "" {
		req.Messages = []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		}
	} else {
		req.Messages = []ChatMessage{
			{Role: "user", Content: userPrompt},
		}
	}

	resp, err := c.doChatRequest(ctx, req)
	if err != nil {
		return "", err
	}

	if resp.Message != nil {
		return resp.Message.Content, nil
	}
	return resp.Response, nil
}

// Stream 流式调用
func (c *OllamaClient) Stream(ctx context.Context, prompt string, taskType string) (<-chan StreamChunk, error) {
	return c.StreamWithSystem(ctx, "", prompt, taskType)
}

// StreamWithSystem 使用系统提示词流式调用
func (c *OllamaClient) StreamWithSystem(ctx context.Context, systemPrompt, userPrompt string, taskType string) (<-chan StreamChunk, error) {
	model := c.config.GetModel(taskType)
	temp := c.config.GetTemperature(taskType)

	req := OllamaRequest{
		Model:  model,
		Stream: true,
		Options: OllamaOptions{
			Temperature: temp,
		},
	}

	// 使用 chat 模式
	if systemPrompt != "" {
		req.Messages = []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		}
	} else {
		req.Messages = []ChatMessage{
			{Role: "user", Content: userPrompt},
		}
	}

	return c.doStreamChatRequest(ctx, req)
}

// doChatRequest 发送 chat 请求
func (c *OllamaClient) doChatRequest(ctx context.Context, req OllamaRequest) (*OllamaResponse, error) {
	url := c.config.BaseURL + "/api/chat"

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("Ollama error: %s - %s", httpResp.Status, string(respBody))
	}

	var resp OllamaResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// doStreamChatRequest 发送流式 chat 请求
func (c *OllamaClient) doStreamChatRequest(ctx context.Context, req OllamaRequest) (<-chan StreamChunk, error) {
	url := c.config.BaseURL + "/api/chat"

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		httpResp.Body.Close()
		return nil, fmt.Errorf("Ollama error: %s - %s", httpResp.Status, string(respBody))
	}

	ch := make(chan StreamChunk, 100)

	go func() {
		defer close(ch)
		defer httpResp.Body.Close()

		decoder := json.NewDecoder(httpResp.Body)

		for {
			var resp OllamaResponse
			if err := decoder.Decode(&resp); err != nil {
				if err == io.EOF {
					ch <- StreamChunk{Done: true}
					return
				}
				ch <- StreamChunk{Error: err}
				return
			}

			if resp.Message != nil && resp.Message.Content != "" {
				ch <- StreamChunk{Content: resp.Message.Content}
			}

			if resp.Done {
				ch <- StreamChunk{Done: true}
				return
			}
		}
	}()

	return ch, nil
}

// CountTokens 计算 token 数量
func (c *OllamaClient) CountTokens(text string) int {
	// Ollama 使用不同的分词器，这里简单估算
	return len(text) / 4
}

// CheckConnection 检查 Ollama 服务是否可用
func (c *OllamaClient) CheckConnection(ctx context.Context) error {
	url := c.config.BaseURL

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("Ollama service not available: %v", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama service error: %s", httpResp.Status)
	}

	return nil
}

// ListModels 列出可用模型
func (c *OllamaClient) ListModels(ctx context.Context) ([]string, error) {
	url := c.config.BaseURL + "/api/tags"

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	var resp struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, err
	}

	models := make([]string, len(resp.Models))
	for i, m := range resp.Models {
		models[i] = m.Name
	}

	return models, nil
}

// PullModel 拉取模型
func (c *OllamaClient) PullModel(ctx context.Context, modelName string) error {
	url := c.config.BaseURL + "/api/pull"

	req := struct {
		Name   string `json:"name"`
		Stream bool   `json:"stream"`
	}{
		Name:   modelName,
		Stream: false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		return fmt.Errorf("failed to pull model: %s - %s", httpResp.Status, string(respBody))
	}

	return nil
}