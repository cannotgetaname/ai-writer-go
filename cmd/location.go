package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"ai-writer/internal/model"
)

// locationCmd represents the location command
var locationCmd = &cobra.Command{
	Use:   "location",
	Short: "地点管理",
	Long: `管理书籍地点，包括查看、添加、编辑、删除等操作。

子命令:
  list      列出所有地点
  add       添加地点
  show      查看地点详情
  edit      编辑地点
  delete    删除地点`,
}

var locationListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "列出地点",
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		locations, err := jsonStore.LoadLocations(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		if len(locations) == 0 {
			fmt.Println("暂无地点，使用 'ai-writer location add' 创建地点")
			return
		}

		switch outputFormat {
		case "json":
			printJSON(locations)
		case "markdown":
			fmt.Println("# 地点列表")
			fmt.Println()
			for _, loc := range locations {
				fmt.Printf("- **%s** (%s)\n", loc.Name, loc.Faction)
			}
		default:
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "名称\t势力\t危险\t简介")
			fmt.Fprintln(w, "----\t----\t----\t----")
			for _, loc := range locations {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					loc.Name, loc.Faction, loc.Danger, truncate(loc.Description, 30))
			}
			w.Flush()
		}
	},
}

var locationShowCmd = &cobra.Command{
	Use:   "show <地点名>",
	Short: "查看地点详情",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		locName := args[0]
		locations, err := jsonStore.LoadLocations(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 查找地点
		var location *model.Location
		for _, loc := range locations {
			if loc.Name == locName {
				location = loc
				break
			}
		}

		if location == nil {
			fmt.Fprintf(os.Stderr, "错误: 地点 '%s' 不存在\n", locName)
			return
		}

		fmt.Printf("📍 %s\n", location.Name)
		fmt.Println("────────────────────────────────")
		if location.Parent != "" {
			fmt.Printf("  上级: %s\n", location.Parent)
		}
		fmt.Printf("  势力: %s\n", location.Faction)
		fmt.Printf("  危险: %s\n", location.Danger)
		fmt.Printf("  描述: %s\n", location.Description)
		if len(location.Neighbors) > 0 {
			fmt.Printf("  相邻: %v\n", location.Neighbors)
		}
	},
}

var locationAddCmd = &cobra.Command{
	Use:   "add",
	Short: "添加地点",
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		name, _ := cmd.Flags().GetString("name")
		parent, _ := cmd.Flags().GetString("parent")
		faction, _ := cmd.Flags().GetString("faction")
		danger, _ := cmd.Flags().GetString("danger")
		description, _ := cmd.Flags().GetString("description")

		if name == "" {
			fmt.Fprintf(os.Stderr, "错误: 请指定地点名称 (--name)\n")
			return
		}

		// 加载现有地点
		locations, err := jsonStore.LoadLocations(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 检查是否已存在
		for _, loc := range locations {
			if loc.Name == name {
				fmt.Fprintf(os.Stderr, "错误: 地点 '%s' 已存在\n", name)
				return
			}
		}

		// 创建新地点
		newLocation := &model.Location{
			ID:          generateID(),
			BookID:      bookName,
			Name:        name,
			Parent:      parent,
			Neighbors:   []string{},
			Description: description,
			Faction:     faction,
			Danger:      danger,
		}

		locations = append(locations, newLocation)

		// 保存
		if err := jsonStore.SaveLocations(bookName, locations); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 已添加地点: %s\n", name)
		if verbose {
			fmt.Printf("   势力: %s\n", faction)
			fmt.Printf("   危险: %s\n", danger)
			fmt.Printf("   描述: %s\n", description)
		}
	},
}

var locationEditCmd = &cobra.Command{
	Use:   "edit <地点名>",
	Short: "编辑地点",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		locName := args[0]

		// 加载地点列表
		locations, err := jsonStore.LoadLocations(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 查找地点
		var location *model.Location
		for _, loc := range locations {
			if loc.Name == locName {
				location = loc
				break
			}
		}

		if location == nil {
			fmt.Fprintf(os.Stderr, "错误: 地点 '%s' 不存在\n", locName)
			return
		}

		// 获取编辑参数
		newName, _ := cmd.Flags().GetString("name")
		faction, _ := cmd.Flags().GetString("faction")
		danger, _ := cmd.Flags().GetString("danger")
		description, _ := cmd.Flags().GetString("description")

		// 更新字段
		if newName != "" {
			location.Name = newName
		}
		if faction != "" {
			location.Faction = faction
		}
		if danger != "" {
			location.Danger = danger
		}
		if description != "" {
			location.Description = description
		}

		// 保存
		if err := jsonStore.SaveLocations(bookName, locations); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 已更新地点: %s\n", location.Name)
		if verbose {
			fmt.Printf("   势力: %s\n", location.Faction)
			fmt.Printf("   危险: %s\n", location.Danger)
			fmt.Printf("   描述: %s\n", location.Description)
		}
	},
}

var locationDeleteCmd = &cobra.Command{
	Use:   "delete <地点名>",
	Short: "删除地点",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		locName := args[0]

		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("确定要删除地点 '%s' 吗？\n", locName)
			fmt.Print("输入 'yes' 确认: ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "yes" {
				fmt.Println("已取消")
				return
			}
		}

		// 加载地点列表
		locations, err := jsonStore.LoadLocations(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 查找并删除
		found := false
		var newLocations []*model.Location
		for _, loc := range locations {
			if loc.Name == locName {
				found = true
				continue
			}
			newLocations = append(newLocations, loc)
		}

		if !found {
			fmt.Fprintf(os.Stderr, "错误: 地点 '%s' 不存在\n", locName)
			return
		}

		// 保存
		if err := jsonStore.SaveLocations(bookName, newLocations); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 已删除地点: %s\n", locName)
	},
}

func init() {
	rootCmd.AddCommand(locationCmd)

	locationCmd.AddCommand(locationListCmd)
	locationCmd.AddCommand(locationShowCmd)
	locationCmd.AddCommand(locationAddCmd)
	locationCmd.AddCommand(locationEditCmd)
	locationCmd.AddCommand(locationDeleteCmd)

	locationAddCmd.Flags().String("name", "", "地点名称")
	locationAddCmd.Flags().String("parent", "", "上级地点")
	locationAddCmd.Flags().String("faction", "", "所属势力")
	locationAddCmd.Flags().String("danger", "", "危险等级")
	locationAddCmd.Flags().String("description", "", "描述")

	locationEditCmd.Flags().String("name", "", "新名称")
	locationEditCmd.Flags().String("faction", "", "所属势力")
	locationEditCmd.Flags().String("danger", "", "危险等级")
	locationEditCmd.Flags().String("description", "", "描述")

	locationDeleteCmd.Flags().BoolP("force", "f", false, "强制删除，不确认")
}