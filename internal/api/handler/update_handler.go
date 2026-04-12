package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"ai-writer/internal/config"

	"github.com/gin-gonic/gin"
)

// UpdateHandler 更新处理器
type UpdateHandler struct {
	config *config.Config
}

// NewUpdateHandler 创建更新处理器
func NewUpdateHandler(cfg *config.Config) *UpdateHandler {
	return &UpdateHandler{config: cfg}
}

// VersionInfo 版本信息响应
type VersionInfo struct {
	Version string `json:"version"`
	Os      string `json:"os"`
	Arch    string `json:"arch"`
	DataDir string `json:"data_dir"`
}

// CheckUpdateRequest 检查更新请求
type CheckUpdateRequest struct {
	Source string `json:"source"` // github/gitee/auto
}

// CheckUpdateResponse 检查更新响应
type CheckUpdateResponse struct {
	HasUpdate      bool   `json:"has_update"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	DownloadURL    string `json:"download_url"`
	Changelog      string `json:"changelog"`
	SourceUsed     string `json:"source_used"`
}

// StartUpdateRequest 启动更新请求
type StartUpdateRequest struct {
	Source  string `json:"source"`
	Version string `json:"version"`
	URL     string `json:"url"` // 可选直接下载链接
}

// StartUpdateResponse 启动更新响应
type StartUpdateResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// GetVersion 获取当前版本信息
func (h *UpdateHandler) GetVersion(c *gin.Context) {
	c.JSON(http.StatusOK, VersionInfo{
		Version:    config.Version,
		Os:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		DataDir:    h.config.Server.DataDir,
	})
}

// CheckUpdate 检查是否有新版本
func (h *UpdateHandler) CheckUpdate(c *gin.Context) {
	var req CheckUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Source = "auto"
	}

	// 确定使用的源
	source := req.Source
	if source == "auto" {
		source = detectSource()
	}

	// 获取最新版本信息
	repo := h.config.Update.GitHubRepo
	if source == "gitee" {
		repo = h.config.Update.GiteeRepo
	}

	release, err := fetchLatestRelease(repo, source)
	if err != nil {
		c.JSON(http.StatusOK, CheckUpdateResponse{
			HasUpdate:      false,
			CurrentVersion: config.Version,
			LatestVersion:  config.Version,
			SourceUsed:     source,
		})
		return
	}

	// 比较版本
	hasUpdate := compareVersions(config.Version, release.Version)

	c.JSON(http.StatusOK, CheckUpdateResponse{
		HasUpdate:      hasUpdate,
		CurrentVersion: config.Version,
		LatestVersion:  release.Version,
		DownloadURL:    release.DownloadURL,
		Changelog:      release.Changelog,
		SourceUsed:     source,
	})
}

// StartUpdate 启动更新流程
func (h *UpdateHandler) StartUpdate(c *gin.Context) {
	var req StartUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 找到 updater 程序
	exeDir := getExeDir()
	updaterExe := filepath.Join(exeDir, "updater")
	if runtime.GOOS == "windows" {
		updaterExe += ".exe"
	}

	if _, err := os.Stat(updaterExe); os.IsNotExist(err) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "updater not found"})
		return
	}

	// 构造命令行参数
	args := []string{
		"-source", req.Source,
		"-dir", exeDir,
	}

	if req.URL != "" {
		args = append(args, "-url", req.URL, "-version", req.Version)
	}

	// 启动 updater 进程
	cmd := exec.Command(updaterExe, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to start updater: %v", err)})
		return
	}

	c.JSON(http.StatusOK, StartUpdateResponse{
		Status:  "started",
		Message: "更新程序已启动，请重启应用完成更新",
	})
}

// detectSource 检测应该使用的下载源
func detectSource() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com", nil)
	if err != nil {
		return "gitee"
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "gitee"
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return "gitee"
	}
	return "github"
}

// ReleaseInfo 发布信息
type ReleaseInfo struct {
	Version     string
	DownloadURL string
	Changelog   string
}

// fetchLatestRelease 从 Release API 获取最新版本
func fetchLatestRelease(repo string, source string) (*ReleaseInfo, error) {
	var apiURL string
	if source == "gitee" {
		apiURL = fmt.Sprintf("https://gitee.com/api/v5/repos/%s/releases/latest", repo)
	} else {
		apiURL = fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
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
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseReleaseJSON(body, source)
}

// parseReleaseJSON 解析 Release API 响应
func parseReleaseJSON(body []byte, source string) (*ReleaseInfo, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	version, _ := data["tag_name"].(string)
	if version == "" {
		version, _ = data["name"].(string)
	}
	changelog, _ := data["body"].(string)

	// 查找匹配当前平台的 asset
	assets, _ := data["assets"].([]interface{})
	expectedPattern := fmt.Sprintf("ai-writer-%s-%s", runtime.GOOS, runtime.GOARCH)

	for _, asset := range assets {
		assetMap, _ := asset.(map[string]interface{})
		name, _ := assetMap["name"].(string)
		url, _ := assetMap["browser_download_url"].(string)

		if strings.Contains(name, expectedPattern) || strings.Contains(name, runtime.GOOS+"-"+runtime.GOARCH) {
			return &ReleaseInfo{
				Version:     version,
				DownloadURL: url,
				Changelog:   changelog,
			}, nil
		}
	}

	return nil, fmt.Errorf("no matching asset found")
}

// compareVersions 比较版本号，返回是否有更新
func compareVersions(current string, latest string) bool {
	// 去掉 v 前缀
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	// 简单比较：假设版本号格式为 x.y.z
	return latest != current && latest > current
}

// getExeDir 获取可执行文件所在目录
func getExeDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}