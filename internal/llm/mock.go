package llm

import (
	"context"
)

// MockClient Mock LLM 客户端，用于测试
type MockClient struct {
	Response string
	Err      error
	Chunks   []StreamChunk // 用于流式响应
}

// Call 调用 LLM
func (m *MockClient) Call(ctx context.Context, prompt string, taskType string) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	return m.Response, nil
}

// CallWithSystem 使用系统提示词调用
func (m *MockClient) CallWithSystem(ctx context.Context, systemPrompt, userPrompt string, taskType string) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	return m.Response, nil
}

// Stream 流式调用
func (m *MockClient) Stream(ctx context.Context, prompt string, taskType string) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk)
	go func() {
		if m.Err != nil {
			ch <- StreamChunk{Error: m.Err}
			close(ch)
			return
		}
		// 如果有预设的 chunks，使用它们
		if len(m.Chunks) > 0 {
			for _, chunk := range m.Chunks {
				ch <- chunk
			}
		} else {
			// 默认行为：发送响应然后结束
			ch <- StreamChunk{Content: m.Response, Done: false}
			ch <- StreamChunk{Done: true}
		}
		close(ch)
	}()
	return ch, nil
}

// StreamWithSystem 使用系统提示词流式调用
func (m *MockClient) StreamWithSystem(ctx context.Context, systemPrompt, userPrompt string, taskType string) (<-chan StreamChunk, error) {
	return m.Stream(ctx, userPrompt, taskType)
}

// CountTokens 计算 token 数量（简单估算）
func (m *MockClient) CountTokens(text string) int {
	// 简单估算：每 4 个字符约 1 token
	return len(text) / 4
}

// NewMockClient 创建 Mock 客户端
func NewMockClient(response string) *MockClient {
	return &MockClient{
		Response: response,
	}
}

// NewMockClientWithError 创建带错误的 Mock 客户端
func NewMockClientWithError(err error) *MockClient {
	return &MockClient{
		Err: err,
	}
}

// NewMockClientWithChunks 创建带流式 chunks 的 Mock 客户端
func NewMockClientWithChunks(chunks []StreamChunk) *MockClient {
	return &MockClient{
		Chunks: chunks,
	}
}