package main

import (
	"ai-writer/cmd"
	"ai-writer/internal/config"
)

// Version 版本号，通过 -ldflags 注入
var Version = "dev"

func main() {
	// 设置版本号
	config.SetVersion(Version)
	cmd.Execute()
}