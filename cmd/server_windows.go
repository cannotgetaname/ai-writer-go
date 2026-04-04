//go:build windows

package cmd

import "os/exec"

func setSysProcAttr(cmd *exec.Cmd) {
	// Windows 不支持 Setsid，无需设置
}