package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// 需要替换的文件列表
var replaceFiles = []string{
	"ai-writer",
	"ai-writer.exe",
	"embedding_server",
	"embedding_server.exe",
	"updater",
	"updater.exe",
	"start_with_embedding.sh",
	"start_with_embedding.bat",
}

// 需要保留的文件/目录列表
var keepFiles = []string{
	"data",
	"config.yaml",
	"embedding_port.txt",
}

// ReplaceFiles 替换文件，保留用户数据
func ReplaceFiles(extractDir string, targetDir string) error {
	// 1. 创建备份目录
	backupDir := filepath.Join(targetDir, ".update_backup")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup dir: %w", err)
	}

	// 2. 备份旧版本二进制文件
	for _, file := range replaceFiles {
		oldPath := filepath.Join(targetDir, file)
		if _, err := os.Stat(oldPath); err == nil {
			if err := backupFile(oldPath, backupDir); err != nil {
				return fmt.Errorf("backup failed for %s: %w", file, err)
			}
		}
	}

	// 3. 找到解压后的实际目录（可能是 ai-writer-{version}-{os}-{arch}/）
	subDir := findExtractedSubDir(extractDir)
	if subDir == "" {
		return fmt.Errorf("cannot find extracted package directory")
	}

	// 4. 替换文件
	for _, file := range replaceFiles {
		srcPath := filepath.Join(subDir, file)
		destPath := filepath.Join(targetDir, file)

		if _, err := os.Stat(srcPath); err != nil {
			continue // 源文件不存在，跳过
		}

		// 删除旧文件
		os.Remove(destPath)

		// 复制新文件
		if err := copyFile(srcPath, destPath); err != nil {
			return fmt.Errorf("replace failed for %s: %w", file, err)
		}

		// 保持可执行权限
		os.Chmod(destPath, 0755)
	}

	// 5. 替换 web/dist 目录
	srcWebDist := filepath.Join(subDir, "web", "dist")
	destWebDist := filepath.Join(targetDir, "web", "dist")

	if _, err := os.Stat(srcWebDist); err == nil {
		// 删除旧的 web/dist
		os.RemoveAll(destWebDist)

		// 复制新的 web/dist
		if err := copyDir(srcWebDist, destWebDist); err != nil {
			return fmt.Errorf("replace web/dist failed: %w", err)
		}
	}

	// 6. 更新 README.txt
	srcReadme := filepath.Join(subDir, "README.txt")
	destReadme := filepath.Join(targetDir, "README.txt")
	if _, err := os.Stat(srcReadme); err == nil {
		copyFile(srcReadme, destReadme)
	}

	return nil
}

// findExtractedSubDir 找到解压后的子目录
func findExtractedSubDir(extractDir string) string {
	entries, err := os.ReadDir(extractDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "ai-writer-") {
			return filepath.Join(extractDir, entry.Name())
		}
	}

	// 如果没有子目录，可能直接解压在当前目录
	return extractDir
}

// copyDir 复制整个目录
func copyDir(src string, dest string) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, destPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, destPath); err != nil {
				return err
			}
		}
	}

	return nil
}