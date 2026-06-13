//
// FilePath    : blog-vim-ime\internal\ime\controller_windows.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : Windows 平台 IME API 控制实现.
//

//go:build windows

package ime

import (
	"context"
	"fmt"
	"syscall"
	"time"
)

const (
	wmIMEControl    = 0x0283
	imcGetOpenState = 0x0005
	imcSetOpenState = 0x0006

	checkInterval = 20 * time.Millisecond
	checkTimeout  = 300 * time.Millisecond
)

var (
	user32DLL            = syscall.NewLazyDLL("user32.dll")
	imm32DLL             = syscall.NewLazyDLL("imm32.dll")
	procGetForegroundWnd = user32DLL.NewProc("GetForegroundWindow")
	procSendMessageW     = user32DLL.NewProc("SendMessageW")
	procImmGetDefaultWnd = imm32DLL.NewProc("ImmGetDefaultIMEWnd")
	procImmGetContext    = imm32DLL.NewProc("ImmGetContext")
	procImmReleaseCtx    = imm32DLL.NewProc("ImmReleaseContext")
	procImmSetOpenStatus = imm32DLL.NewProc("ImmSetOpenStatus")
	procImmGetOpenStatus = imm32DLL.NewProc("ImmGetOpenStatus")
)

type WindowsController struct{}

// NewController 创建 Windows 平台的 IME 控制器.
// 返回值为可执行系统 API 切换的控制器实例.
// 当前实现不会在构造阶段返回错误.
func NewController() (Controller, error) {
	return &WindowsController{}, nil
}

// SetOpenStatus 通过系统 IME API 设置输入法开关状态.
// ctx 用于取消等待, open 为目标开关状态, targetWindow 为目标窗口句柄.
// 当 targetWindow 为 0 时, 自动使用当前前台窗口.
// 当状态未达到预期, 或上下文取消时返回错误.
func (c *WindowsController) SetOpenStatus(ctx context.Context, open bool, targetWindow uintptr) error {
	windowHandle, err := resolveTargetWindow(targetWindow)
	if err != nil {
		return err
	}

	if err := setOpenStatusByContext(ctx, windowHandle, open); err == nil {
		return nil
	}

	return setOpenStatusByIMEWindow(ctx, windowHandle, open)
}

// resolveTargetWindow 解析本次切换要操作的窗口句柄.
// targetWindow 为请求中的目标窗口句柄, 当为 0 时自动读取前台窗口.
// 返回值为有效窗口句柄, 若无法获取则返回错误.
func resolveTargetWindow(targetWindow uintptr) (uintptr, error) {
	if targetWindow != 0 {
		return targetWindow, nil
	}

	foreground, _, fgErr := procGetForegroundWnd.Call()
	if foreground == 0 {
		if fgErr != syscall.Errno(0) {
			return 0, fmt.Errorf("GetForegroundWindow failed: %w", fgErr)
		}

		return 0, fmt.Errorf("foreground window not found")
	}

	return foreground, nil
}

// setOpenStatusByContext 通过 ImmSetOpenStatus 设置窗口输入法状态.
// ctx 用于中途取消等待, windowHandle 为目标窗口句柄, open 为目标状态.
// 返回 nil 表示已切换成功; 返回错误表示该路径失败.
func setOpenStatusByContext(ctx context.Context, windowHandle uintptr, open bool) error {
	inputContext, err := getInputContext(windowHandle)
	if err != nil {
		return err
	}
	defer releaseInputContext(windowHandle, inputContext)

	target := uintptr(0)
	if open {
		target = 1
	}

	result, _, callErr := procImmSetOpenStatus.Call(inputContext, target)
	if result == 0 {
		if callErr != nil && callErr != syscall.Errno(0) {
			return fmt.Errorf("ImmSetOpenStatus failed: %w", callErr)
		}

		return fmt.Errorf("ImmSetOpenStatus failed")
	}

	deadline := time.Now().Add(checkTimeout)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		current, err := getOpenStatusByContext(inputContext)
		if err != nil {
			return err
		}

		if current == open {
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("ime open status did not reach target %t", open)
		}

		time.Sleep(checkInterval)
	}
}

// setOpenStatusByIMEWindow 通过默认 IME 窗口消息设置输入法状态.
// ctx 用于中途取消等待, windowHandle 为目标窗口句柄, open 为目标状态.
// 返回 nil 表示已切换成功; 返回错误表示该路径失败.
func setOpenStatusByIMEWindow(ctx context.Context, windowHandle uintptr, open bool) error {
	imeWindow, err := getDefaultIMEWindow(windowHandle)
	if err != nil {
		return err
	}

	target := uintptr(0)
	if open {
		target = 1
	}

	sendIMEControlMessage(imeWindow, imcSetOpenState, target)

	deadline := time.Now().Add(checkTimeout)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		current, readErr := getOpenStatusByIMEWindow(imeWindow)
		if readErr != nil {
			return readErr
		}

		if current == open {
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("ime open status did not reach target %t", open)
		}

		time.Sleep(checkInterval)
	}
}

// getDefaultIMEWindow 获取目标窗口关联的默认 IME 窗口句柄.
// windowHandle 为目标窗口句柄.
// 返回 IME 窗口句柄; 当 IME 窗口不存在时返回错误.
func getDefaultIMEWindow(windowHandle uintptr) (uintptr, error) {
	imeWindow, _, imeErr := procImmGetDefaultWnd.Call(windowHandle)
	if imeWindow == 0 {
		if imeErr != syscall.Errno(0) {
			return 0, fmt.Errorf("ImmGetDefaultIMEWnd failed: %w", imeErr)
		}

		return 0, fmt.Errorf("default ime window not found")
	}

	return imeWindow, nil
}

// getOpenStatusByIMEWindow 查询指定默认 IME 窗口当前的开关状态.
// imeWindow 表示默认 IME 窗口句柄.
// 返回 true 表示中文开启, false 表示英文开启.
func getOpenStatusByIMEWindow(imeWindow uintptr) (bool, error) {
	result := sendIMEControlMessage(imeWindow, imcGetOpenState, 0)

	return result != 0, nil
}

// getInputContext 获取目标窗口的输入上下文句柄.
// windowHandle 为目标窗口句柄.
// 返回输入上下文句柄; 当获取失败时返回错误.
func getInputContext(windowHandle uintptr) (uintptr, error) {
	inputContext, _, callErr := procImmGetContext.Call(windowHandle)
	if inputContext == 0 {
		if callErr != nil && callErr != syscall.Errno(0) {
			return 0, fmt.Errorf("ImmGetContext failed: %w", callErr)
		}

		return 0, fmt.Errorf("input context not found")
	}

	return inputContext, nil
}

// releaseInputContext 释放窗口输入上下文句柄.
// windowHandle 为目标窗口句柄, inputContext 为输入上下文句柄.
// 无返回值, 释放失败时仅消费错误返回值以满足静态检查.
func releaseInputContext(windowHandle, inputContext uintptr) {
	_, _, callErr := procImmReleaseCtx.Call(windowHandle, inputContext)
	if callErr != nil && callErr != syscall.Errno(0) {
		_ = callErr.Error()
	}
}

// getOpenStatusByContext 查询输入上下文的开关状态.
// inputContext 为输入上下文句柄.
// 返回 true 表示中文开启, false 表示英文开启.
func getOpenStatusByContext(inputContext uintptr) (bool, error) {
	result, _, callErr := procImmGetOpenStatus.Call(inputContext)
	if callErr != nil && callErr != syscall.Errno(0) {
		return false, fmt.Errorf("ImmGetOpenStatus failed: %w", callErr)
	}

	return result != 0, nil
}

// sendIMEControlMessage 向默认 IME 窗口发送控制消息.
// imeWindow 为 IME 窗口句柄, wParam 和 lParam 为 IME 控制参数.
// 返回值为 SendMessageW 的消息处理结果.
func sendIMEControlMessage(imeWindow, wParam, lParam uintptr) uintptr {
	result, _, callErr := procSendMessageW.Call(imeWindow, wmIMEControl, wParam, lParam)
	if callErr != nil && callErr != syscall.Errno(0) {
		// NOTE: SendMessageW 的失败不通过 GetLastError 判定, 这里仅显式消费 error 返回值.
		_ = callErr.Error()
	}

	return result
}
