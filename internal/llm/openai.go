package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAIClient OpenAI/DeepSeek 客户端（兼容 OpenAI API）
type OpenAIClient struct {
	config     *Config
	httpClient *http.Client
}

// NewOpenAIClient 创建 OpenAI 客户端
func NewOpenAIClient(cfg *Config) *OpenAIClient {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 120
	}

	return &OpenAIClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

// NewDeepSeekClient 创建 DeepSeek 客户端（兼容 OpenAI API）
func NewDeepSeekClient(cfg *Config) *OpenAIClient {
	// DeepSeek 使用相同的 OpenAI 兼容接口
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.deepseek.com"
	}
	return NewOpenAIClient(cfg)
}

// ChatRequest OpenAI Chat API 请求
type ChatRequest struct {
	Model       string          `json:"model"`
	Messages    []ChatMessage   `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Role    string `json:"role"`    // system/user/assistant
	Content string `json:"content"`
}

// ChatResponse OpenAI Chat API 响应
type ChatResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []ChatChoice   `json:"choices"`
	Usage   *ChatUsage     `json:"usage,omitempty"`
}

// ChatChoice 选择项
type ChatChoice struct {
	Index        int          `json:"index"`
	Message      ChatMessage  `json:"message"`
	Delta        *ChatDelta   `json:"delta,omitempty"`
	FinishReason string       `json:"finish_reason"`
}

// ChatDelta 流式增量
type ChatDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// ChatUsage Token 使用量
type ChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Call 调用 LLM
func (c *OpenAIClient) Call(ctx context.Context, prompt string, taskType string) (string, error) {
	return c.CallWithSystem(ctx, "", prompt, taskType)
}

// CallWithSystem 使用系统提示词调用
func (c *OpenAIClient) CallWithSystem(ctx context.Context, systemPrompt, userPrompt string, taskType string) (string, error) {
	model := c.config.GetModel(taskType)
	temp := c.config.GetTemperature(taskType)

	messages := []ChatMessage{}
	if systemPrompt != "" {
		messages = append(messages, ChatMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}
	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: userPrompt,
	})

	req := ChatRequest{
		Model:       model,
		Messages:    messages,
		Temperature: temp,
		Stream:      false,
	}

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices")
	}

	return resp.Choices[0].Message.Content, nil
}

// Stream 流式调用
func (c *OpenAIClient) Stream(ctx context.Context, prompt string, taskType string) (<-chan StreamChunk, error) {
	return c.StreamWithSystem(ctx, "", prompt, taskType)
}

// StreamWithSystem 使用系统提示词流式调用
func (c *OpenAIClient) StreamWithSystem(ctx context.Context, systemPrompt, userPrompt string, taskType string) (<-chan StreamChunk, error) {
	model := c.config.GetModel(taskType)
	temp := c.config.GetTemperature(taskType)

	messages := []ChatMessage{}
	if systemPrompt != "" {
		messages = append(messages, ChatMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}
	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: userPrompt,
	})

	req := ChatRequest{
		Model:       model,
		Messages:    messages,
		Temperature: temp,
		Stream:      true,
	}

	return c.doStreamRequest(ctx, req)
}

// doRequest 发送请求
func (c *OpenAIClient) doRequest(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	baseURL := c.config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	url := baseURL + "/chat/completions"

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("API error: %s - %s", httpResp.Status, string(respBody))
	}

	var resp ChatResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// doStreamRequest 发送流式请求
func (c *OpenAIClient) doStreamRequest(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error) {
	baseURL := c.config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	url := baseURL + "/chat/completions"

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		httpResp.Body.Close()
		return nil, fmt.Errorf("API error: %s - %s", httpResp.Status, string(respBody))
	}

	ch := make(chan StreamChunk, 100)

	go func() {
		defer close(ch)
		defer httpResp.Body.Close()

		reader := httpResp.Body
		buf := make([]byte, 4096)

		for {
			n, err := reader.Read(buf)
			if err != nil {
				if err == io.EOF {
					ch <- StreamChunk{Done: true}
					return
				}
				ch <- StreamChunk{Error: err}
				return
			}

			data := string(buf[:n])
			lines := strings.Split(data, "\n")

			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || line == "data: [DONE]" {
					if line == "data: [DONE]" {
						ch <- StreamChunk{Done: true}
						return
					}
					continue
				}

				if strings.HasPrefix(line, "data: ") {
					jsonStr := strings.TrimPrefix(line, "data: ")
					var resp ChatResponse
					if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
						continue
					}

					if len(resp.Choices) > 0 && resp.Choices[0].Delta != nil {
						content := resp.Choices[0].Delta.Content
						if content != "" {
							ch <- StreamChunk{Content: content}
						}
					}

					if resp.Choices[0].FinishReason == "stop" {
						ch <- StreamChunk{Done: true}
						return
					}
				}
			}
		}
	}()

	return ch, nil
}

// CountTokens 计算 token 数量（简单估算）
func (c *OpenAIClient) CountTokens(text string) int {
	// 简单估算：中文约 1.5 字符/token，英文约 4 字符/token
	// 这里用更简单的方法：每 4 个字符约 1 token
	return len(text) / 4
}