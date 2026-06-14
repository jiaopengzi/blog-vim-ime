//
// FilePath    : blog-vim-ime\internal\autostart\autostart_windows.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : Windows 注册表开机启动管理.
//

//go:build windows

package autostart

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const (
	runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
	appName    = "blog-vim-ime"
)

// IsEnabled 查询当前用户注册表中是否已设置开机启动.
// 无参数.
// 返回 true 表示已启用, false 表示未启用; 注册表读取失败时返回错误.
func IsEnabled() (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false, fmt.Errorf("open registry key: %w", err)
	}
	defer func() {
		if closeErr := key.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	_, _, err = key.GetStringValue(appName)
	if err == registry.ErrNotExist {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read registry value: %w", err)
	}

	return true, nil
}

// Enable 向当前用户注册表写入开机启动项, 指向当前 EXE 的绝对路径.
// 无参数.
// 注册表写入失败时返回错误.
func Enable() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	absPath, err := filepath.Abs(exePath)
	if err != nil {
		return fmt.Errorf("resolve absolute path: %w", err)
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open registry key: %w", err)
	}
	defer func() {
		if closeErr := key.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	if err := key.SetStringValue(appName, absPath); err != nil {
		return fmt.Errorf("set registry value: %w", err)
	}

	return nil
}

// Disable 从当前用户注册表中删除开机启动项.
// 无参数.
// 当注册表中不存在对应键值时视为成功; 注册表操作失败时返回错误.
func Disable() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open registry key: %w", err)
	}
	defer func() {
		if closeErr := key.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	if err := key.DeleteValue(appName); err != nil {
		if err == registry.ErrNotExist {
			// 已不存在, 视为成功
			return nil
		}

		return fmt.Errorf("delete registry value: %w", err)
	}

	return nil
}
