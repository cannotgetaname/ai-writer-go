package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"ai-writer/internal/model"
)

// characterCmd represents the character command
var characterCmd = &cobra.Command{
	Use:   "character",
	Short: "人物管理",
	Long: `管理书籍人物，包括查看、添加、编辑、删除等操作。

子命令:
  list      列出所有人物
  add       添加人物
  show      查看人物详情
  edit      编辑人物
  delete    删除人物`,
}

var characterListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "列出人物",
	Run: func(cmd *cobra.Command, args []string) {
		name, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		characters, err := jsonStore.LoadCharacters(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		if len(characters) == 0 {
			fmt.Println("暂无人物，使用 'ai-writer character add' 创建人物")
			return
		}

		switch outputFormat {
		case "json":
			printJSON(characters)
		case "markdown":
			fmt.Println("# 人物列表")
			fmt.Println()
			for _, ch := range characters {
				fmt.Printf("- **%s** (%s) - %s\n", ch.Name, ch.Role, ch.Status)
			}
		default:
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "姓名\t角色\t状态\t简介")
			fmt.Fprintln(w, "----\t----\t----\t----")
			for _, ch := range characters {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					ch.Name, ch.Role, ch.Status, truncate(ch.Bio, 30))
			}
			w.Flush()
		}
	},
}

var characterShowCmd = &cobra.Command{
	Use:   "show <角色名>",
	Short: "查看人物详情",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		charName := args[0]
		characters, err := jsonStore.LoadCharacters(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// Find character
		var char *model.Character
		for _, ch := range characters {
			if ch.Name == charName {
				char = ch
				break
			}
		}

		if char == nil {
			fmt.Fprintf(os.Stderr, "错误: 人物 '%s' 不存在\n", charName)
			return
		}

		fmt.Printf("👤 %s\n", char.Name)
		fmt.Println("────────────────────────────────")
		fmt.Printf("  角色: %s\n", char.Role)
		fmt.Printf("  性别: %s\n", char.Gender)
		fmt.Printf("  状态: %s\n", char.Status)
		fmt.Printf("  简介: %s\n", char.Bio)
		fmt.Println()

		if len(char.Relations) > 0 {
			fmt.Println("关系:")
			for _, rel := range char.Relations {
				fmt.Printf("  - %s: %s (%d)\n", rel.TargetName, rel.Type, rel.Value)
			}
		}

		if len(char.EmotionalArc) > 0 {
			fmt.Println("情感弧线:")
			for _, ep := range char.EmotionalArc {
				fmt.Printf("  - 第%d章: %s (强度: %d)\n", ep.ChapterID, ep.Emotion, ep.Intensity)
			}
		}
	},
}

var characterAddCmd = &cobra.Command{
	Use:   "add",
	Short: "添加人物",
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		name, _ := cmd.Flags().GetString("name")
		role, _ := cmd.Flags().GetString("role")
		gender, _ := cmd.Flags().GetString("gender")
		bio, _ := cmd.Flags().GetString("bio")
		status, _ := cmd.Flags().GetString("char-status")

		if name == "" {
			fmt.Fprintf(os.Stderr, "错误: 请指定人物名称 (--name)\n")
			return
		}

		// 加载现有人物
		characters, err := jsonStore.LoadCharacters(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 检查是否已存在
		for _, c := range characters {
			if c.Name == name {
				fmt.Fprintf(os.Stderr, "错误: 人物 '%s' 已存在\n", name)
				return
			}
		}

		// 创建新人物
		newChar := &model.Character{
			ID:        generateID(),
			BookID:    bookName,
			Name:      name,
			Role:      role,
			Gender:    gender,
			Status:    status,
			Bio:       bio,
			Relations: []model.Relation{},
		}

		// 添加到列表
		characters = append(characters, newChar)

		// 保存
		if err := jsonStore.SaveCharacters(bookName, characters); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 已添加人物: %s\n", name)
		if verbose {
			fmt.Printf("   ID: %s\n", newChar.ID)
			fmt.Printf("   角色: %s\n", role)
			fmt.Printf("   性别: %s\n", gender)
			fmt.Printf("   状态: %s\n", status)
			fmt.Printf("   简介: %s\n", bio)
		}
	},
}

var characterDeleteCmd = &cobra.Command{
	Use:   "delete <角色名>",
	Short: "删除人物",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		charName := args[0]

		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("确定要删除人物 '%s' 吗？\n", charName)
			fmt.Print("输入 'yes' 确认: ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "yes" {
				fmt.Println("已取消")
				return
			}
		}

		// 加载人物列表
		characters, err := jsonStore.LoadCharacters(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 查找并删除
		found := false
		var newCharacters []*model.Character
		for _, c := range characters {
			if c.Name == charName {
				found = true
				continue
			}
			newCharacters = append(newCharacters, c)
		}

		if !found {
			fmt.Fprintf(os.Stderr, "错误: 人物 '%s' 不存在\n", charName)
			return
		}

		// 保存
		if err := jsonStore.SaveCharacters(bookName, newCharacters); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 已删除人物: %s\n", charName)
	},
}

var characterEditCmd = &cobra.Command{
	Use:   "edit <角色名>",
	Short: "编辑人物",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		charName := args[0]

		// 加载人物列表
		characters, err := jsonStore.LoadCharacters(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 查找人物
		var char *model.Character
		for _, c := range characters {
			if c.Name == charName {
				char = c
				break
			}
		}

		if char == nil {
			fmt.Fprintf(os.Stderr, "错误: 人物 '%s' 不存在\n", charName)
			return
		}

		// 获取编辑参数
		newName, _ := cmd.Flags().GetString("name")
		role, _ := cmd.Flags().GetString("role")
		gender, _ := cmd.Flags().GetString("gender")
		bio, _ := cmd.Flags().GetString("bio")
		status, _ := cmd.Flags().GetString("char-status")

		// 更新字段
		if newName != "" {
			char.Name = newName
		}
		if role != "" {
			char.Role = role
		}
		if gender != "" {
			char.Gender = gender
		}
		if bio != "" {
			char.Bio = bio
		}
		if status != "" {
			char.Status = status
		}

		// 保存
		if err := jsonStore.SaveCharacters(bookName, characters); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 已更新人物: %s\n", char.Name)
		if verbose {
			fmt.Printf("   角色: %s\n", char.Role)
			fmt.Printf("   性别: %s\n", char.Gender)
			fmt.Printf("   状态: %s\n", char.Status)
			fmt.Printf("   简介: %s\n", char.Bio)
		}
	},
}

func init() {
	rootCmd.AddCommand(characterCmd)

	characterCmd.AddCommand(characterListCmd)
	characterCmd.AddCommand(characterShowCmd)
	characterCmd.AddCommand(characterAddCmd)
	characterCmd.AddCommand(characterDeleteCmd)
	characterCmd.AddCommand(characterEditCmd)

	characterAddCmd.Flags().String("name", "", "人物名称")
	characterAddCmd.Flags().String("role", "配角", "角色类型 (主角/配角/反派/路人)")
	characterAddCmd.Flags().String("gender", "男", "性别")
	characterAddCmd.Flags().String("bio", "", "简介")
	characterAddCmd.Flags().String("char-status", "存活", "状态 (存活/死亡/失踪)")

	characterDeleteCmd.Flags().BoolP("force", "f", false, "强制删除，不确认")

	characterEditCmd.Flags().String("name", "", "新名称")
	characterEditCmd.Flags().String("role", "", "角色类型 (主角/配角/反派/路人)")
	characterEditCmd.Flags().String("gender", "", "性别")
	characterEditCmd.Flags().String("bio", "", "简介")
	characterEditCmd.Flags().String("char-status", "", "状态 (存活/死亡/失踪)")
}