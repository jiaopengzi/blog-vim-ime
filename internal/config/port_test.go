//
// FilePath    : blog-vim-ime\internal\config\port_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 端口配置加载器的单元测试.
//

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPortPrefersExecutableDirectoryForRelativePath(t *testing.T) {
	executableDir := t.TempDir()
	workingDir := t.TempDir()

	writePortConfig(t, filepath.Join(executableDir, "port.yaml"), "port: 9001\n")
	writePortConfig(t, filepath.Join(workingDir, "port.yaml"), "port: 9002\n")

	restoreExecutablePath := stubExecutablePath(t, filepath.Join(executableDir, "blog-vim-ime.exe"))
	defer restoreExecutablePath()

	restoreWorkingDir := chdirForTest(t, workingDir)
	defer restoreWorkingDir()

	port, err := LoadPort("port.yaml", 8765)
	if err != nil {
		t.Fatalf("LoadPort returned error: %v", err)
	}

	if port != 9001 {
		t.Fatalf("expected executable directory port 9001, got %d", port)
	}
}

func TestLoadPortMissingFileUsesDefault(t *testing.T) {
	t.Parallel()

	port, err := LoadPort(filepath.Join(t.TempDir(), "missing.yaml"), 8765)
	if err != nil {
		t.Fatalf("LoadPort returned error: %v", err)
	}

	if port != 8765 {
		t.Fatalf("expected default port 8765, got %d", port)
	}
}

func TestLoadPortFromYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "port.yaml")
	err := os.WriteFile(configPath, []byte("port: 9999\n"), 0o644)
	if err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	port, err := LoadPort(configPath, 8765)
	if err != nil {
		t.Fatalf("LoadPort returned error: %v", err)
	}

	if port != 9999 {
		t.Fatalf("expected port 9999, got %d", port)
	}
}

func TestLoadPortRejectsInvalidPort(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "port.yaml")
	err := os.WriteFile(configPath, []byte("port: 70000\n"), 0o644)
	if err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	_, err = LoadPort(configPath, 8765)
	if err == nil {
		t.Fatalf("expected error for invalid port")
	}
}

func TestLoadPortRejectsMissingValue(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "port.yaml")
	err := os.WriteFile(configPath, []byte("# no port here\n"), 0o644)
	if err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	_, err = LoadPort(configPath, 8765)
	if err == nil {
		t.Fatalf("expected error for missing port")
	}
}

// writePortConfig 写入测试用端口配置文件.
// t 表示当前测试上下文, path 表示目标文件路径, content 表示配置文件内容.
// 无返回值; 写入失败时直接终止当前测试.
func writePortConfig(t *testing.T, path string, content string) {
	t.Helper()

	err := os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write config failed: %v", err)
	}
}

// stubExecutablePath 临时替换可执行文件路径解析函数.
// t 表示当前测试上下文, executablePath 表示伪造的可执行文件路径.
// 返回值 func(), 用于恢复原始路径解析函数.
func stubExecutablePath(t *testing.T, executablePath string) func() {
	t.Helper()

	previous := executablePathFunc
	executablePathFunc = func() (string, error) {
		return executablePath, nil
	}

	return func() {
		executablePathFunc = previous
	}
}

// chdirForTest 临时切换当前工作目录.
// t 表示当前测试上下文, directory 表示目标工作目录.
// 返回值 func(), 用于恢复原始工作目录.
func chdirForTest(t *testing.T, directory string) func() {
	t.Helper()

	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory failed: %v", err)
	}

	err = os.Chdir(directory)
	if err != nil {
		t.Fatalf("change working directory failed: %v", err)
	}

	return func() {
		if chdirErr := os.Chdir(previous); chdirErr != nil {
			t.Fatalf("restore working directory failed: %v", chdirErr)
		}
	}
}
