package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server       ServerConfig       `mapstructure:"server"`
	LLM          LLMConfig          `mapstructure:"llm"`
	Embedding    EmbeddingConfig    `mapstructure:"embedding"`
	Storage      StorageConfig      `mapstructure:"storage"`
	VectorStore  VectorStoreConfig  `mapstructure:"vector_store"`
	VectorDB     VectorDBConfig     `mapstructure:"vectordb"`
	Prompts      PromptsConfig      `mapstructure:"prompts"`
	Models       ModelsConfig       `mapstructure:"models"`
	Pricing      PricingConfig      `mapstructure:"pricing"`
	LastOpenBook string             `mapstructure:"last_open_book"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port    string `mapstructure:"port"`
	DataDir string `mapstructure:"data_dir"`
	AuthKey string `mapstructure:"auth_key"` // API 认证密钥（可选）
}

// LLMConfig LLM 配置
type LLMConfig struct {
	Provider     string             `mapstructure:"provider"`
	APIKey       string             `mapstructure:"api_key"`
	BaseURL      string             `mapstructure:"base_url"`
	Models       map[string]string  `mapstructure:"models"`
	Temperatures map[string]float64 `mapstructure:"temperatures"`
	MaxRetries   int                `mapstructure:"max_retries"`
	Timeout      int                `mapstructure:"timeout"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	ProjectsDir string `mapstructure:"projects_dir"`
	VectorDBDir string `mapstructure:"vector_db_dir"`
}

// VectorStoreConfig 向量存储配置
type VectorStoreConfig struct {
	ChunkSize int `mapstructure:"chunk_size"`
	Overlap   int `mapstructure:"overlap"`
}

// VectorDBConfig 向量数据库配置
type VectorDBConfig struct {
	Provider string `mapstructure:"provider"` // sqlite-vec / custom
	BaseURL  string `mapstructure:"base_url"` // custom provider URL
}

// EmbeddingConfig 向量嵌入配置
type EmbeddingConfig struct {
	Provider string `mapstructure:"provider"` // ollama/openai
	Model    string `mapstructure:"model"`    // embeddinggemma:latest / text-embedding-3-small
	BaseURL  string `mapstructure:"base_url"` // http://localhost:11434
	APIKey   string `mapstructure:"api_key"`  // OpenAI API key (if needed)
}

// PromptsConfig 提示词配置
type PromptsConfig struct {
	WriterSystem           string `mapstructure:"writer_system"`
	ArchitectSystem        string `mapstructure:"architect_system"`
	ReviewerSystem         string `mapstructure:"reviewer_system"`
	AuditorSystem          string `mapstructure:"auditor_system"`
	TimekeeperSystem       string `mapstructure:"timekeeper_system"`
	SummarySystem          string `mapstructure:"summary_system"`
	SummaryChapterSystem   string `mapstructure:"summary_chapter_system"`
	SummaryBookSystem      string `mapstructure:"summary_book_system"`
	KnowledgeFilterSystem  string `mapstructure:"knowledge_filter_system"`
	JsonOnlyArchitectSystem string `mapstructure:"json_only_architect_system"`
	InspirationAssistantSystem string `mapstructure:"inspiration_assistant_system"`
}

// ModelsConfig 模型配置
type ModelsConfig struct {
	Writer    string `mapstructure:"writer"`
	Architect string `mapstructure:"architect"`
	Editor    string `mapstructure:"editor"`
	Reviewer  string `mapstructure:"reviewer"`
	Auditor   string `mapstructure:"auditor"`
	Timekeeper string `mapstructure:"timekeeper"`
	Summary   string `mapstructure:"summary"`
}

// PricingConfig 定价配置
type PricingConfig map[string]ModelPricing

// ModelPricing 模型定价
type ModelPricing struct {
	Input  float64 `mapstructure:"input"`
	Output float64 `mapstructure:"output"`
}

// 默认配置
var defaultConfig = Config{
	Server: ServerConfig{
		Port:    "8081",
		DataDir: "data",
	},
	LLM: LLMConfig{
		Provider:   "deepseek",
		BaseURL:    "https://api.deepseek.com",
		MaxRetries: 3,
		Timeout:    120,
		Models: map[string]string{
			"writer":    "deepseek-chat",
			"architect": "deepseek-reasoner",
			"editor":    "deepseek-chat",
			"reviewer":  "deepseek-chat",
			"auditor":   "deepseek-reasoner",
			"timekeeper": "deepseek-chat",
			"summary":   "deepseek-chat",
		},
		Temperatures: map[string]float64{
			"writer":    1.5,
			"architect": 1.0,
			"editor":    0.7,
			"reviewer":  0.5,
			"auditor":   0.6,
			"timekeeper": 0.1,
			"summary":   0.5,
		},
	},
	Embedding: EmbeddingConfig{
		Provider: "python",
		Model:    "",
		BaseURL:  "",
	},
	Storage: StorageConfig{
		ProjectsDir: "data/projects",
		VectorDBDir: "data/vector_db",
	},
	VectorStore: VectorStoreConfig{
		ChunkSize: 500,
		Overlap:   100,
	},
	VectorDB: VectorDBConfig{
		Provider: "sqlite-vec",
		BaseURL:  "",
	},
	Prompts: PromptsConfig{
		WriterSystem: `你是一个顶级网文作家，擅长热血、快节奏、爽点密集的风格。写作要求：
1. 【黄金法则】多用"展示"而非"讲述"（Show, don't tell）。
2. 【对话驱动】通过对话推动剧情和塑造性格，拒绝大段枯燥的心理描写。
3. 【感官描写】调动视觉、听觉、触觉，增加代入感。
4. 【节奏把控】详略得当，战斗场面要干脆利落，日常互动要有趣味。
5. 【字数要求】每章正文必须在 3000-5000 字之间，这是硬性要求，不可偷工减料。
请根据提供的大纲、世界观和上下文，撰写引人入胜的正文。`,
		ArchitectSystem: `你是一个精通起承转合的剧情架构师。你的任务是基于前文和伏笔，规划后续章节。要求：
1. 【逻辑严密】后续剧情必须符合人物性格逻辑，不能机械降神。
2. 【冲突制造】每一章都必须有一个核心冲突或悬念钩子。
3. 【伏笔回收】尝试利用历史记忆中的伏笔。
请严格只返回一个标准的 JSON 列表，不要包含 Markdown 标记或其他废话。`,
		ReviewerSystem: `你是一位资深网文编辑，以专业、犀利的审稿风格著称。审稿要求：
1. 【具体】指出问题必须具体到哪句话、哪个行为，不能泛泛而谈
2. 【可操作】修改建议必须明确、可执行，给出具体方向或示例
3. 【从严】只有严重影响阅读体验才标"严重"，一般问题标"中等"，小瑕疵标"轻微"
4. 【务实】只指出真正影响质量的问题，优秀的部分不要强行挑刺`,
		AuditorSystem: `你是一个世界观数据库管理员。你的任务是分析小说正文，提取状态变更。你需要敏锐地捕捉隐性信息（例如：'他断了一臂' -> 状态: 重伤/残疾）。

请严格按以下 JSON 结构输出（不要使用 Markdown 代码块）：
{
  "char_updates": [{"name": "名字", "field": "属性名", "new_value": "新值"}],
  "item_updates": [{"name": "物品名", "field": "属性名", "new_value": "新值"}],
  "new_chars": [{"name": "名字", "gender": "性别", "role": "角色类型", "status": "状态", "bio": "简介"}],
  "new_items": [{"name": "物品名", "type": "类型", "owner": "持有者", "desc": "描述"}],
  "new_locs": [{"name": "地名", "faction": "所属势力", "desc": "描述"}],
  "relation_updates": [{"source": "主角", "target": "配角", "type": "关系类型"}]
}`,
		TimekeeperSystem: `你是一个精确的时间记录员。你的任务是分析正文，推算时间流逝。输出必须是严格的 JSON 格式：{"label": "当前时间点(如：修仙历10年春)", "duration": "本章经过的时间(如：3天)", "events": ["事件1", "事件2"]}。请只输出 JSON。`,
		SummarySystem: `你是一个专业的网文编辑，擅长提炼剧情精华。请将给定的小说章节压缩成 150 字以内的摘要。`,
		SummaryChapterSystem: `你是一个专业的网文编辑，擅长提炼剧情精华。请将给定的小说章节压缩成 150 字以内的摘要。要求：
1. 保留核心冲突和结果。
2. 记录关键道具或人物的获得/损失。
3. 记录重要的伏笔。
不要写流水账，要写干货。`,
		SummaryBookSystem: `你是一个资深主编，拥有宏观的上帝视角。请根据各章节的摘要，梳理出整本书目前的剧情脉络（Story Arc）。要求：
1. 串联主要故事线，忽略支线细枝末节。
2. 明确主角目前的处境、目标和成长阶段。
3. 篇幅控制在 500 字左右，适合快速回顾。`,
		KnowledgeFilterSystem: `你是一个专业的资料整理助手。你的任务是从检索到的碎片信息中，剔除无关噪音，筛选出对当前章节写作真正有帮助的背景信息（如人物之前的恩怨、物品的特殊设定、地点的具体样貌）。如果片段与当前剧情无关，请忽略。`,
		JsonOnlyArchitectSystem: `你是一个只输出JSON的架构师。`,
		InspirationAssistantSystem: `你是一个网文灵感助手。请只返回请求的内容，不要废话。`,
	},
	Pricing: PricingConfig{
		"deepseek-chat":     {Input: 0.00194, Output: 0.00792},
		"deepseek-reasoner": {Input: 0.00396, Output: 0.01577},
		"gpt-4":             {Input: 0.216, Output: 0.432},
		"gpt-4-turbo":       {Input: 0.072, Output: 0.216},
		"gpt-3.5-turbo":     {Input: 0.0036, Output: 0.0108},
		"claude-3-opus":     {Input: 0.108, Output: 0.54},
		"claude-3-sonnet":   {Input: 0.0216, Output: 0.108},
		"claude-3-haiku":    {Input: 0.0018, Output: 0.009},
	},
}

// Load 加载配置
func Load() (*Config, error) {
	cfg := defaultConfig

	// 设置配置文件搜索路径
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("./data")

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// 配置文件不存在，使用默认配置
		return &cfg, nil
	}

	// 解析配置
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &cfg, nil
}

// GetProjectsPath 获取项目目录路径
func (c *Config) GetProjectsPath() string {
	if c.Storage.ProjectsDir == "" {
		return "data/projects"
	}
	return c.Storage.ProjectsDir
}

// GetVectorDBPath 获取向量数据库路径
func (c *Config) GetVectorDBPath() string {
	if c.Storage.VectorDBDir == "" {
		return "data/vector_db"
	}
	return c.Storage.VectorDBDir
}

// EnsureDataDirs 确保数据目录存在
func (c *Config) EnsureDataDirs() error {
	dirs := []string{
		c.GetProjectsPath(),
		c.GetVectorDBPath(),
		filepath.Join(c.GetProjectsPath(), "global"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// GetModel 获取指定任务类型的模型
func (c *Config) GetModel(taskType string) string {
	if model, ok := c.LLM.Models[taskType]; ok {
		return model
	}
	// 默认模型
	return c.LLM.Models["writer"]
}

// GetTemperature 获取指定任务类型的温度
func (c *Config) GetTemperature(taskType string) float64 {
	if temp, ok := c.LLM.Temperatures[taskType]; ok {
		return temp
	}
	return 1.0 // 默认温度
}

// GetPrompt 获取指定任务类型的系统提示词
func (c *Config) GetPrompt(taskType string) string {
	switch taskType {
	case "writer":
		return c.Prompts.WriterSystem
	case "architect":
		return c.Prompts.ArchitectSystem
	case "reviewer":
		return c.Prompts.ReviewerSystem
	case "auditor":
		return c.Prompts.AuditorSystem
	case "timekeeper":
		return c.Prompts.TimekeeperSystem
	case "summary":
		return c.Prompts.SummarySystem
	case "summary_chapter":
		return c.Prompts.SummaryChapterSystem
	case "summary_book":
		return c.Prompts.SummaryBookSystem
	case "knowledge_filter":
		return c.Prompts.KnowledgeFilterSystem
	case "json_architect":
		return c.Prompts.JsonOnlyArchitectSystem
	case "inspiration":
		return c.Prompts.InspirationAssistantSystem
	default:
		return c.Prompts.WriterSystem
	}
}

// CalculateCost 计算 token 成本
func (c *Config) CalculateCost(model string, inputTokens, outputTokens int) float64 {
	pricing, ok := c.Pricing[model]
	if !ok {
		return 0
	}
	return pricing.Input*float64(inputTokens) + pricing.Output*float64(outputTokens)
}