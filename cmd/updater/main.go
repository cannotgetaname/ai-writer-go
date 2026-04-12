package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// 解析命令行参数
	var (
		repo        string
		version     string
		source      string
		workDir     string
		downloadURL string
	)

	flag.StringVar(&repo, "repo", "cannotgetaname/ai-writer-go", "Repository path")
	flag.StringVar(&version, "version", "", "Target version (empty for latest)")
	flag.StringVar(&source, "source", "auto", "Download source: github/gitee/auto")
	flag.StringVar(&workDir, "dir", "", "Working directory (default: exe dir)")
	flag.StringVar(&downloadURL, "url", "", "Direct download URL (skip API)")
	flag.Parse()

	// 确定工作目录
	if workDir == "" {
		workDir = getExeDir()
	}
	fmt.Printf("Working directory: %s\n", workDir)

	// 确定下载源
	if source == "auto" {
		if CanAccessGitHub() {
			source = "github"
			fmt.Println("Using GitHub as download source")
		} else {
			source = "gitee"
			fmt.Println("Using Gitee as download source (GitHub unreachable)")
		}
	}

	// 获取版本信息
	var release *ReleaseInfo
	var err error

	if downloadURL != "" {
		// 直接使用提供的下载链接
		release = &ReleaseInfo{
			Version:     version,
			DownloadURL: downloadURL,
		}
	} else {
		// 从 API 获取
		fmt.Printf("Fetching latest release from %s...\n", source)
		release, err = FetchLatestRelease(repo, source)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching release: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Latest version: %s\n", release.Version)
	}

	// 创建临时目录
	tempDir := filepath.Join(workDir, ".update_temp")
	os.RemoveAll(tempDir)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir) // 清理临时文件

	// 下载发布包
	packageFile := filepath.Join(tempDir, "package.tar.gz")
	if source == "gitee" && !strings.Contains(release.DownloadURL, "gitee.com") {
		// 转换 GitHub 链接到 Gitee（如果镜像）
		// 这里假设用户会在 Gitee 上同步发布
	}

	fmt.Printf("Downloading from: %s\n", release.DownloadURL)
	err = DownloadFile(release.DownloadURL, packageFile, func(downloaded, total int64) {
		if total > 0 {
			percent := float64(downloaded) / float64(total) * 100
			fmt.Printf("Progress: %.1f%% (%d/%d bytes)\n", percent, downloaded, total)
		} else {
			fmt.Printf("Downloaded: %d bytes\n", downloaded)
		}
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Download complete!")

	// 解压
	fmt.Println("Extracting package...")
	osType, _ := getPlatform()
	if osType == "windows" {
		err = extractZip(packageFile, tempDir)
	} else {
		err = extractTarGz(packageFile, tempDir)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Extract failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Extract complete!")

	// 替换文件
	fmt.Println("Replacing files...")
	err = ReplaceFiles(tempDir, workDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Replace failed: %v\n", err)
		// 尝试从备份恢复
		restoreBackup(workDir)
		os.Exit(1)
	}
	fmt.Println("Replace complete!")

	// 清理备份（可选保留）
	backupDir := filepath.Join(workDir, ".update_backup")
	fmt.Printf("Backup stored at: %s\n", backupDir)
	fmt.Println("You can delete it after verifying the update.")

	// 完成
	fmt.Println("\n=== Update Complete ===")
	fmt.Printf("Updated to version: %s\n", release.Version)
	fmt.Println("Please restart the application.")
}

// restoreBackup 从备份恢复
func restoreBackup(workDir string) {
	backupDir := filepath.Join(workDir, ".update_backup")
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, ".bak") {
			originalName := name[:len(name)-4]
			src := filepath.Join(backupDir, name)
			dest := filepath.Join(workDir, originalName)
			copyFile(src, dest)
			fmt.Printf("Restored: %s\n", originalName)
		}
	}
}