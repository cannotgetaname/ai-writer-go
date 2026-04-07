package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"ai-writer/internal/api"
)

var (
	serverPort int
	serverHost string
	daemonMode bool
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "启动 Web 服务器",
	Long: `启动 AI Writer Web 界面服务。

启动后可通过浏览器访问 http://localhost:8081 进行可视化操作。

示例:
  ai-writer server                    # 前台运行
  ai-writer server -d                 # 后台运行
  ai-writer server -p 9090            # 指定端口
  ai-writer server --host 0.0.0.0     # 允许外部访问`,
	Run: func(cmd *cobra.Command, args []string) {
		if daemonMode {
			startDaemon()
			return
		}

		fmt.Printf("启动 AI Writer Web 服务...\n")
		fmt.Printf("访问地址: http://%s:%d\n", serverHost, serverPort)

		// 启动 Gin 服务器 (存储已在 PersistentPreRun 中初始化)
		router := api.SetupRouter(cfg)
		router.Run(fmt.Sprintf("%s:%d", serverHost, serverPort))
	},
}

// stopCmd 停止服务命令
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "停止 Web 服务器",
	Long: `停止正在运行的 AI Writer Web 服务。

示例:
  ai-writer stop          # 停止服务
  ai-writer stop --force  # 强制停止`,
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")
		stopServer(force)
	},
}

// statusCmd 查看服务状态
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看服务状态",
	Long:  `查看 AI Writer Web 服务的运行状态。`,
	Run: func(cmd *cobra.Command, args []string) {
		checkServerStatus()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statusCmd)

	serverCmd.Flags().IntVarP(&serverPort, "port", "p", 8081, "服务器端口")
	serverCmd.Flags().StringVarP(&serverHost, "host", "H", "localhost", "监听地址")
	serverCmd.Flags().BoolVarP(&daemonMode, "daemon", "d", false, "后台运行")

	stopCmd.Flags().BoolP("force", "f", false, "强制停止")
}

// startDaemon 后台启动服务
func startDaemon() {
	// 获取当前可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取程序路径失败: %v\n", err)
		return
	}

	// 构建启动命令
	args := []string{"server", "--port", strconv.Itoa(serverPort), "--host", serverHost}

	cmd := exec.Command(execPath, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	// Unix 特定的进程属性
	setSysProcAttr(cmd)

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "后台启动失败: %v\n", err)
		return
	}

	// 保存 PID
	pidFile := getPidFile()
	os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0644)

	fmt.Printf("AI Writer 服务已在后台启动 (PID: %d)\n", cmd.Process.Pid)
	fmt.Printf("访问地址: http://%s:%d\n", serverHost, serverPort)
	fmt.Println("使用 'ai-writer stop' 停止服务")
}

// stopServer 停止服务
func stopServer(force bool) {
	// 查找进程
	pids := findServerProcesses()

	if len(pids) == 0 {
		fmt.Println("没有正在运行的 AI Writer 服务")
		return
	}

	for _, pid := range pids {
		var sig os.Signal = syscall.SIGTERM
		if force {
			sig = syscall.SIGKILL
		}

		proc, err := os.FindProcess(pid)
		if err != nil {
			fmt.Printf("查找进程 %d 失败: %v\n", pid, err)
			continue
		}

		if err := proc.Signal(sig); err != nil {
			fmt.Printf("停止进程 %d 失败: %v\n", pid, err)
			continue
		}

		fmt.Printf("已停止服务 (PID: %d)\n", pid)
	}

	// 清理 PID 文件
	os.Remove(getPidFile())
}

// checkServerStatus 检查服务状态
func checkServerStatus() {
	pids := findServerProcesses()

	if len(pids) == 0 {
		fmt.Println("AI Writer 服务状态: 未运行")
		return
	}

	fmt.Printf("AI Writer 服务状态: 运行中\n")
	for _, pid := range pids {
		fmt.Printf("  PID: %d, 端口: %d\n", pid, serverPort)
	}
	fmt.Println("\n使用 'ai-writer stop' 停止服务")
}

// findServerProcesses 查找服务进程
func findServerProcesses() []int {
	var pids []int

	// 先检查 PID 文件
	pidFile := getPidFile()
	if data, err := os.ReadFile(pidFile); err == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
			// 验证进程是否存在
			if proc, _ := os.FindProcess(pid); proc != nil {
				if err := proc.Signal(syscall.Signal(0)); err == nil {
					pids = append(pids, pid)
				}
			}
		}
	}

	// 通过端口查找进程 (Linux/macOS)
	if len(pids) == 0 {
		// 使用 lsof 或 netstat 查找端口占用
		out, err := exec.Command("sh", "-c",
			fmt.Sprintf("lsof -i :%d -t 2>/dev/null || ss -tlnp 2>/dev/null | grep ':%d ' | grep -oP 'pid=\\K[0-9]+'", serverPort, serverPort)).Output()
		if err == nil {
			for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
				if pid, err := strconv.Atoi(strings.TrimSpace(line)); err == nil && pid > 0 {
					pids = append(pids, pid)
				}
			}
		}
	}

	// 也可以通过进程名查找
	if len(pids) == 0 {
		out, err := exec.Command("pgrep", "-f", "ai-writer server").Output()
		if err == nil {
			for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
				if pid, err := strconv.Atoi(strings.TrimSpace(line)); err == nil && pid > 0 {
					pids = append(pids, pid)
				}
			}
		}
	}

	return pids
}

// getPidFile 获取 PID 文件路径
func getPidFile() string {
	return ".ai-writer.pid"
}