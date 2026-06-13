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
