package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// worldviewCmd represents the worldview command
var worldviewCmd = &cobra.Command{
	Use:   "worldview",
	Short: "世界观管理",
	Long: `管理书籍世界观设定。

子命令:
  show      查看世界观
  edit      编辑世界观`,
}

var worldviewShowCmd = &cobra.Command{
	Use:   "show",
	Short: "查看世界观",
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			return
		}

		worldview, err := jsonStore.LoadWorldView(bookName)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			return
		}

		fmt.Println("🌍 世界观设定")
		fmt.Println("═══════════════════════════════════════")

		fmt.Println("\n【基本信息】")
		fmt.Println("────────────────────────────────")
		fmt.Printf("  题材: %s\n", worldview.BasicInfo.Genre)
		fmt.Printf("  时代: %s\n", worldview.BasicInfo.Era)
		fmt.Printf("  科技水平: %s\n", worldview.BasicInfo.TechLevel)

		fmt.Println("\n【核心设定】")
		fmt.Println("────────────────────────────────")
		fmt.Printf("  力量体系: %s\n", worldview.CoreSettings.PowerSystem)
		fmt.Printf("  社会结构: %s\n", worldview.CoreSettings.SocialStructure)
		fmt.Printf("  特殊规则: %s\n", worldview.CoreSettings.SpecialRules)

		fmt.Println("\n【关键元素】")
		fmt.Println("────────────────────────────────")
		fmt.Printf("  重要物品: %s\n", worldview.KeyElements.ImportantItems)
		fmt.Printf("  势力组织: %s\n", worldview.KeyElements.Organizations)
		fmt.Printf("  主要地点: %s\n", worldview.KeyElements.Locations)

		fmt.Println("\n【背景故事】")
		fmt.Println("────────────────────────────────")
		fmt.Printf("  历史背景: %s\n", worldview.Background.History)
		fmt.Printf("  主要矛盾: %s\n", worldview.Background.MainConflict)
		fmt.Printf("  发展趋势: %s\n", worldview.Background.Development)
	},
}

var worldviewEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "编辑世界观",
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			return
		}

		// 加载现有世界观
		worldview, err := jsonStore.LoadWorldView(bookName)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			return
		}

		// 获取编辑参数
		genre, _ := cmd.Flags().GetString("genre")
		era, _ := cmd.Flags().GetString("era")
		techLevel, _ := cmd.Flags().GetString("tech-level")
		powerSystem, _ := cmd.Flags().GetString("power-system")
		socialStructure, _ := cmd.Flags().GetString("social-structure")
		specialRules, _ := cmd.Flags().GetString("special-rules")
		importantItems, _ := cmd.Flags().GetString("important-items")
		organizations, _ := cmd.Flags().GetString("organizations")
		locations, _ := cmd.Flags().GetString("locations")
		history, _ := cmd.Flags().GetString("history")
		mainConflict, _ := cmd.Flags().GetString("main-conflict")
		development, _ := cmd.Flags().GetString("development")

		// 更新字段
		if genre != "" {
			worldview.BasicInfo.Genre = genre
		}
		if era != "" {
			worldview.BasicInfo.Era = era
		}
		if techLevel != "" {
			worldview.BasicInfo.TechLevel = techLevel
		}
		if powerSystem != "" {
			worldview.CoreSettings.PowerSystem = powerSystem
		}
		if socialStructure != "" {
			worldview.CoreSettings.SocialStructure = socialStructure
		}
		if specialRules != "" {
			worldview.CoreSettings.SpecialRules = specialRules
		}
		if importantItems != "" {
			worldview.KeyElements.ImportantItems = importantItems
		}
		if organizations != "" {
			worldview.KeyElements.Organizations = organizations
		}
		if locations != "" {
			worldview.KeyElements.Locations = locations
		}
		if history != "" {
			worldview.Background.History = history
		}
		if mainConflict != "" {
			worldview.Background.MainConflict = mainConflict
		}
		if development != "" {
			worldview.Background.Development = development
		}

		// 保存
		if err := jsonStore.SaveWorldView(bookName, worldview); err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: 保存失败: %v\n", err)
			return
		}

		fmt.Println("✅ 已更新世界观设定")
	},
}

func init() {
	rootCmd.AddCommand(worldviewCmd)

	worldviewCmd.AddCommand(worldviewShowCmd)
	worldviewCmd.AddCommand(worldviewEditCmd)

	// 基本信息
	worldviewEditCmd.Flags().String("genre", "", "题材类型")
	worldviewEditCmd.Flags().String("era", "", "时代背景")
	worldviewEditCmd.Flags().String("tech-level", "", "科技水平")

	// 核心设定
	worldviewEditCmd.Flags().String("power-system", "", "力量体系")
	worldviewEditCmd.Flags().String("social-structure", "", "社会结构")
	worldviewEditCmd.Flags().String("special-rules", "", "特殊规则")

	// 关键元素
	worldviewEditCmd.Flags().String("important-items", "", "重要物品")
	worldviewEditCmd.Flags().String("organizations", "", "势力组织")
	worldviewEditCmd.Flags().String("locations", "", "主要地点")

	// 背景故事
	worldviewEditCmd.Flags().String("history", "", "历史背景")
	worldviewEditCmd.Flags().String("main-conflict", "", "主要矛盾")
	worldviewEditCmd.Flags().String("development", "", "发展趋势")
}