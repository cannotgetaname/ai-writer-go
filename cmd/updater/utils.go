package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// extractTarGz 解压 tar.gz 文件
func extractTarGz(src string, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// 安全性检查：防止路径穿越
		target := filepath.Join(dest, header.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dest)) {
			return fmt.Errorf("invalid path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}

// extractZip 解压 zip 文件
func extractZip(src string, dest string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		// 安全性检查：防止路径穿越
		target := filepath.Join(dest, file.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dest)) {
			return fmt.Errorf("invalid path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}

		outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, file.Mode())
		if err != nil {
			return err
		}

		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// backupFile 备份文件
func backupFile(src string, backupDir string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil // 文件不存在，无需备份
	}

	backupPath := filepath.Join(backupDir, filepath.Base(src)+".bak")
	return copyFile(src, backupPath)
}

// copyFile 复制文件
func copyFile(src string, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}

// cleanDir 清理目录（保留指定文件）
func cleanDir(dir string, keepFiles []string) error {
	keepSet := make(map[string]bool)
	for _, f := range keepFiles {
		keepSet[f] = true
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		if keepSet[name] {
			continue
		}

		path := filepath.Join(dir, name)
		os.RemoveAll(path)
	}

	return nil
}

// getExeDir 获取可执行文件所在目录
func getExeDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

// getPlatform 检测操作系统和架构
func getPlatform() (string, string) {
	return runtime.GOOS, runtime.GOARCH
}