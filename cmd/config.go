package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"ai-writer/internal/config"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置管理",
	Long: `管理 AI Writer 配置，包括查看、设置、初始化等操作。

子命令:
  show      显示当前配置
  get       获取配置项
  set       设置配置项
  init      初始化配置文件
  check     检查配置有效性`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "显示当前配置",
	Run: func(cmd *cobra.Command, args []string) {
		if cfg == nil {
			fmt.Println("未加载配置文件")
			return
		}

		fmt.Println("当前配置:")
		fmt.Println("────────────────────────────────")
		fmt.Printf("  Provider:    %s\n", cfg.LLM.Provider)
		fmt.Printf("  API Key:     %s\n", maskAPIKey(cfg.LLM.APIKey))
		fmt.Printf("  Base URL:    %s\n", cfg.LLM.BaseURL)
		fmt.Printf("  Data Path:   %s\n", cfg.Server.DataDir)
		fmt.Println()
		fmt.Println("模型配置:")
		for task, model := range cfg.LLM.Models {
			fmt.Printf("  %s: %s\n", task, model)
		}
		fmt.Println()
		fmt.Println("温度配置:")
		for task, temp := range cfg.LLM.Temperatures {
			fmt.Printf("  %s: %.2f\n", task, temp)
		}
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "获取配置项",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		if cfg == nil {
			fmt.Println("未加载配置文件")
			return
		}

		value := getConfigValue(cfg, key)
		if value == "" {
			fmt.Printf("配置项 '%s' 不存在\n", key)
		} else {
			fmt.Println(value)
		}
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "设置配置项",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		if err := setConfigValue(key, value); err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		fmt.Printf("✅ 已设置 %s = %s\n", key, value)
		fmt.Println("配置已保存到文件")
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化配置文件",
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")

		configPath := "config.yaml"

		// 检查文件是否存在
		if _, err := os.Stat(configPath); err == nil && !force {
			fmt.Println("配置文件已存在，使用 --force 覆盖")
			return
		}

		// 创建默认配置
		defaultCfg := getDefaultConfigYAML()

		if err := os.WriteFile(configPath, []byte(defaultCfg), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 创建配置文件失败: %v\n", err)
			return
		}

		fmt.Println("✅ 已创建配置文件: config.yaml")
		fmt.Println()
		fmt.Println("请编辑配置文件设置 API Key:")
		fmt.Println("  vim config.yaml")
		fmt.Println()
		fmt.Println("或使用环境变量:")
		fmt.Println("  export LLM_API_KEY=your-api-key")
	},
}

var configCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "检查配置有效性",
	Run: func(cmd *cobra.Command, args []string) {
		if cfg == nil {
			fmt.Println("❌ 未加载配置文件")
			return
		}

		fmt.Println("检查配置...")
		fmt.Println("────────────────────────────────")

		issues := checkConfig(cfg)
		if len(issues) == 0 {
			fmt.Println("✅ 配置检查通过")
		} else {
			for _, issue := range issues {
				fmt.Printf("❌ %s\n", issue)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configCheckCmd)

	configInitCmd.Flags().BoolP("force", "f", false, "强制覆盖已存在的配置文件")
}

func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func getConfigValue(cfg *config.Config, key string) string {
	switch key {
	case "provider":
		return cfg.LLM.Provider
	case "api_key":
		return cfg.LLM.APIKey
	case "base_url":
		return cfg.LLM.BaseURL
	case "data_path":
		return cfg.Server.DataDir
	default:
		if model, ok := cfg.LLM.Models[key]; ok {
			return model
		}
		return ""
	}
}

func checkConfig(cfg *config.Config) []string {
	var issues []string

	if cfg.LLM.APIKey == "" {
		issues = append(issues, "API Key 未设置")
	}

	if cfg.LLM.Provider == "" {
		issues = append(issues, "Provider 未设置")
	}

	if _, ok := cfg.LLM.Models["writer"]; !ok {
		issues = append(issues, "写作模型未配置")
	}

	return issues
}

// setConfigValue 设置配置项并保存到文件
func setConfigValue(key, value string) error {
	configPath := "config.yaml"

	// 读取现有配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 配置文件不存在，创建默认配置
			data = []byte(getDefaultConfigYAML())
		} else {
			return fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	// 解析 YAML
	var cfgMap map[string]interface{}
	if err := yaml.Unmarshal(data, &cfgMap); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 设置值（支持嵌套键如 llm.api_key）
	setNestedValue(cfgMap, key, value)

	// 序列化回 YAML
	newData, err := yaml.Marshal(cfgMap)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(configPath, newData, 0644); err != nil {
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	// 同时更新 viper
	viper.Set(key, value)

	return nil
}

// setNestedValue 设置嵌套配置值
func setNestedValue(cfgMap map[string]interface{}, key, value string) {
	parts := splitKey(key)
	current := cfgMap

	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			newMap := make(map[string]interface{})
			current[part] = newMap
			current = newMap
		}
	}

	current[parts[len(parts)-1]] = value
}

// splitKey 分割配置键（支持点号和下划线）
func splitKey(key string) []string {
	// 将下划线转换为点号层级
	parts := []string{}
	current := ""

	for _, c := range key {
		if c == '_' || c == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// getDefaultConfigYAML 返回默认配置 YAML
func getDefaultConfigYAML() string {
	return `# AI Writer 配置文件
# 请根据实际情况修改

server:
  port: "8081"
  data_dir: "data"

llm:
  provider: "deepseek"
  api_key: ""  # 请设置你的 API Key
  base_url: "https://api.deepseek.com"

  models:
    writer: "deepseek-chat"
    architect: "deepseek-reasoner"
    reviewer: "deepseek-chat"
    auditor: "deepseek-reasoner"
    timekeeper: "deepseek-chat"
    summary: "deepseek-chat"

  temperatures:
    writer: 1.3
    architect: 1.0
    reviewer: 0.5
    auditor: 0.6
    timekeeper: 0.1
    summary: 0.5

  max_retries: 3
  timeout: 120

storage:
  projects_dir: "data/projects"
  vector_db_dir: "data/vector_db"
`
}

// getConfigFilePath 获取配置文件路径
func getConfigFilePath() string {
	if cfgFile != "" {
		return cfgFile
	}
	return filepath.Join(".", "config.yaml")
}