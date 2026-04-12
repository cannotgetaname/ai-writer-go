package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// GitHub Release API
const githubAPI = "https://api.github.com/repos/%s/releases/latest"
const githubDownload = "https://github.com/%s/releases/download/%s/%s"

// Gitee Release API
const giteeAPI = "https://gitee.com/api/v5/repos/%s/releases/latest"
const giteeDownload = "https://gitee.com/%s/releases/download/%s/%s"

// ReleaseInfo 发布信息
type ReleaseInfo struct {
	Version     string
	DownloadURL string
	Changelog   string
}

// FetchLatestRelease 获取最新发布信息
func FetchLatestRelease(repo string, source string) (*ReleaseInfo, error) {
	var apiURL string
	if source == "gitee" {
		apiURL = fmt.Sprintf(giteeAPI, repo)
	} else {
		apiURL = fmt.Sprintf(githubAPI, repo)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	if source == "github" {
		req.Header.Set("Accept", "application/vnd.github.v3+json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// 解析 JSON 响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseReleaseJSON(body, repo, source)
}

// parseReleaseJSON 解析 Release JSON响应
func parseReleaseJSON(body []byte, repo string, source string) (*ReleaseInfo, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	version, _ := data["tag_name"].(string)
	if version == "" {
		version, _ = data["name"].(string) // Gitee 有时用 name
	}
	changelog, _ := data["body"].(string)

	// 查找匹配当前平台的下载链接
	os, arch := getPlatform()
	assets, _ := data["assets"].([]interface{})

	expectedName := fmt.Sprintf("ai-writer-%s-%s-%s", version, os, arch)
	if os == "windows" {
		expectedName += ".zip"
	} else {
		expectedName += ".tar.gz"
	}

	for _, asset := range assets {
		assetMap, _ := asset.(map[string]interface{})
		name, _ := assetMap["name"].(string)
		url, _ := assetMap["browser_download_url"].(string)

		if name == expectedName || strings.Contains(name, expectedName[:len(expectedName)-7]) {
			return &ReleaseInfo{
				Version:     version,
				DownloadURL: url,
				Changelog:   changelog,
			}, nil
		}
	}

	return nil, fmt.Errorf("no matching asset found for %s", expectedName)
}

// DownloadFile 下载文件到指定路径
func DownloadFile(url string, dest string, progress func(int64, int64)) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	// 创建目标文件
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	// 获取文件大小用于进度显示
	total := resp.ContentLength

	// 使用 progressWriter 包装
	pw := &progressWriter{
		total:      total,
		progress:   progress,
		downloaded: 0,
	}

	_, err = io.Copy(out, io.TeeReader(resp.Body, pw))
	return err
}

// progressWriter 进度写入器
type progressWriter struct {
	total      int64
	downloaded int64
	progress   func(int64, int64)
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.downloaded += int64(n)
	if pw.progress != nil {
		pw.progress(pw.downloaded, pw.total)
	}
	return n, nil
}

// CanAccessGitHub 检测是否能访问 GitHub
func CanAccessGitHub() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com", nil)
	if err != nil {
		return false
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode < 500
}