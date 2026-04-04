package llm

import (
	"context"
)

// Client LLM 客户端接口
type Client interface {
	// Call 调用 LLM，返回完整响应
	Call(ctx context.Context, prompt string, taskType string) (string, error)

	// CallWithSystem 使用系统提示词调用
	CallWithSystem(ctx context.Context, systemPrompt, userPrompt string, taskType string) (string, error)

	// Stream 流式调用，返回 channel
	Stream(ctx context.Context, prompt string, taskType string) (<-chan StreamChunk, error)

	// StreamWithSystem 使用系统提示词流式调用
	StreamWithSystem(ctx context.Context, systemPrompt, userPrompt string, taskType string) (<-chan StreamChunk, error)

	// CountTokens 计算 token 数量（用于费用统计）
	CountTokens(text string) int
}

// StreamChunk 流式输出块
type StreamChunk struct {
	Content string
	Done    bool
	Error   error
}

// Config LLM 配置
type Config struct {
	Provider     string            `yaml:"provider"`      // openai/deepseek/ollama
	APIKey       string            `yaml:"api_key"`
	BaseURL      string            `yaml:"base_url"`
	Models       map[string]string `yaml:"models"`        // task_type -> model
	Temperatures map[string]float64 `yaml:"temperatures"`
	MaxRetries   int               `yaml:"max_retries"`
	Timeout      int               `yaml:"timeout"`       // seconds
}

// NewClient 创建客户端
func NewClient(cfg *Config) Client {
	switch cfg.Provider {
	case "openai":
		return NewOpenAIClient(cfg)
	case "deepseek":
		return NewDeepSeekClient(cfg) // DeepSeek 兼容 OpenAI 接口
	case "ollama":
		return NewOllamaClient(cfg)
	default:
		return NewOpenAIClient(cfg)
	}
}

// GetModel 获取指定任务类型的模型
func (c *Config) GetModel(taskType string) string {
	if model, ok := c.Models[taskType]; ok {
		return model
	}
	// 默认模型
	return c.Models["writer"]
}

// GetTemperature 获取指定任务类型的温度
func (c *Config) GetTemperature(taskType string) float64 {
	if temp, ok := c.Temperatures[taskType]; ok {
		return temp
	}
	return 1.0 // 默认温度
}