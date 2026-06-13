//go:build windows
// +build windows

//
// FilePath    : blog-vim-ime\cmd\blog-vim-ime\main.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 程序入口, 负责启动本地 IME HTTP 服务.
//

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"syscall"
	"time"

	trayassets "blog-vim-ime"
	"blog-vim-ime/internal/config"
	"blog-vim-ime/internal/ime"
	"blog-vim-ime/internal/server"

	"github.com/getlantern/systray"
)

var (
	// Windows API 声明
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	user32               = syscall.NewLazyDLL("user32.dll")
	procGetConsoleWindow = kernel32.NewProc("GetConsoleWindow")
	procShowWindow       = user32.NewProc("ShowWindow")
)

const (
	swHide = 0 // Windows API ShowWindow 参数: 隐藏窗口
)

const (
	defaultPort       = 8765
	serverStopTimout  = 5 * time.Second
	readHeaderTimeout = 3 * time.Second
)

// init 在程序启动时隐藏控制台窗口.
// 仅在 Windows 平台下执行; 若获取或隐藏窗口失败, 继续运行.
func init() {
	hideConsoleWindow()
}

// hideConsoleWindow 隐藏当前进程附着的控制台窗口.
// 无参数.
// 无返回值; 当当前进程没有控制台窗口时直接返回.
// nolint: errcheck
func hideConsoleWindow() {
	// 获取当前进程的控制台窗口句柄
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd != 0 {
		// 隐藏窗口
		_, _, _ = procShowWindow.Call(hwnd, swHide)
	}
}

// loadTrayIcon 返回嵌入到二进制中的 Windows 托盘图标数据.
// 无参数.
// 返回值 []byte, systray 所需的 ICO 图像字节数据.
func loadTrayIcon() []byte {
	return trayassets.DefaultTrayIcon()
}

// main 启动本地 IME 服务并在系统托盘显示.
// 无参数.
// 当配置读取失败或服务异常退出时直接终止进程.
// 支持通过托盘菜单退出程序.
func main() {
	logger := log.New(os.Stdout, "[blog-vim-ime] ", log.LstdFlags|log.Lmsgprefix)

	port, err := config.LoadPort("port.yaml", defaultPort)
	if err != nil {
		logger.Fatalf("load port config failed: %v", err)
	}

	controller, err := ime.NewController()
	if err != nil {
		logger.Fatalf("create ime controller failed: %v", err)
	}

	handler := server.New(controller, logger).Routes()
	httpServer := &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", port),
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	logger.Printf("listening on %s", httpServer.Addr)

	// 启动 HTTP 服务
	serverErr := make(chan error, 1)
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// 设置托盘菜单
	systray.Run(
		func() {
			// 托盘启动回调
			systray.SetTitle("blog-vim-ime")
			systray.SetTooltip("输入法切换服务 (IME Switcher)")

			// 加载并设置托盘图标
			if iconData := loadTrayIcon(); len(iconData) > 0 {
				systray.SetIcon(iconData)
			}

			// 创建退出菜单项
			mQuit := systray.AddMenuItem("退出", "退出程序")

			// 监听菜单事件
			go func() {
				<-mQuit.ClickedCh
				systray.Quit()
			}()
		},
		func() {
			// 托盘退出回调，关闭 HTTP 服务
			ctx, cancel := context.WithTimeout(context.Background(), serverStopTimout)
			defer cancel()

			if shutdownErr := httpServer.Shutdown(ctx); shutdownErr != nil && !errors.Is(shutdownErr, context.Canceled) {
				logger.Printf("shutdown failed: %v", shutdownErr)
			}
		},
	)

	// 检查 HTTP 服务错误
	select {
	case err := <-serverErr:
		logger.Fatalf("server exited unexpectedly: %v", err)
	default:
	}
}
