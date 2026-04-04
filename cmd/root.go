package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"ai-writer/internal/config"
	"ai-writer/internal/store"
)

var (
	cfgFile     string
	bookName    string
	outputFormat string
	verbose     bool

	cfg       *config.Config
	jsonStore *store.JSONStore
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ai-writer",
	Short: "AI 辅助写作工具",
	Long: `AI Writer - AI 辅助小说创作工具

支持 CLI 命令行和 Web UI 两种操作方式。

示例:
  # 列出所有书籍
  ai-writer book list

  # 创建新书
  ai-writer book create "我的小说"

  # AI 生成章节
  ai-writer write 1 --stream

  # 启动 Web 服务
  ai-writer server

  # 进入交互模式
  ai-writer interactive`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// 初始化配置和存储
		initConfig()
		initStore()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// 全局选项
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "配置文件路径 (默认: ./config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&bookName, "book", "b", "", "指定书籍名称")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "输出格式: table/json/markdown")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "详细输出")

	viper.BindPFlag("book", rootCmd.PersistentFlags().Lookup("book"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("./configs")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	var err error
	cfg, err = config.Load()
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
	}
}

// initStore initializes the JSON store
func initStore() {
	jsonStore = store.NewJSONStore(".")
}

// getBookName returns the book name from flag or last used book
func getBookName() string {
	if bookName != "" {
		return bookName
	}
	// Try to get from config (last open book)
	if cfg != nil {
		lastBook := getLastOpenBook()
		if lastBook != "" {
			return lastBook
		}
	}
	return ""
}

// requireBookName ensures a book name is provided
func requireBookName() (string, error) {
	name := getBookName()
	if name == "" {
		return "", fmt.Errorf("请指定书籍名称: 使用 -b 参数或设置默认书籍")
	}
	return name, nil
}

// getLastOpenBook 获取上次打开的书籍
func getLastOpenBook() string {
	stateFile := filepath.Join(".claude", "state.json")
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return ""
	}

	var state struct {
		LastBook string `json:"last_book"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return ""
	}

	// 验证书籍是否还存在
	if state.LastBook != "" {
		books, _ := jsonStore.ListBooks()
		for _, b := range books {
			if b.Name == state.LastBook {
				return state.LastBook
			}
		}
	}

	return ""
}

// setLastOpenBook 设置上次打开的书籍
func setLastOpenBook(name string) {
	stateDir := filepath.Join(".claude")
	os.MkdirAll(stateDir, 0755)

	state := struct {
		LastBook string `json:"last_book"`
	}{
		LastBook: name,
	}

	data, _ := json.MarshalIndent(state, "", "  ")
	os.WriteFile(filepath.Join(stateDir, "state.json"), data, 0644)
}

// printJSON prints data as JSON
func printJSON(data interface{}) {
	output, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(output))
}

// printJSONToFile writes JSON data to a file
func printJSONToFile(data interface{}, filePath string) error {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, output, 0644)
}

// generateID 生成唯一 ID
func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}