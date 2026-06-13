//
// FilePath    : blog-vim-ime\internal\ime\switcher.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : IME 模式切换执行器, 供 CLI 与 HTTP 复用.
//

package ime

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// SwitchRequest 表示一次模式切换请求参数.
type SwitchRequest struct {
	ModeBefore string
	ModeAfter  string
	WindowHwnd string
}

// SwitchResult 表示模式切换执行结果.
type SwitchResult struct {
	Changed bool
}

// Switcher 封装模式判断与 IME 切换执行流程.
type Switcher struct {
	controller Controller
}

// NewSwitcher 创建可复用的模式切换执行器.
// controller 为底层 IME 控制器实现.
// 返回值为可在 CLI 与 HTTP 调用的执行器实例.
func NewSwitcher(controller Controller) *Switcher {
	return &Switcher{controller: controller}
}

// Execute 执行一次模式切换请求.
// ctx 用于中途取消, req 包含模式与可选窗口句柄.
// 返回结果中 Changed 表示本次是否触发了实际 IME 切换; 参数非法或系统调用失败时返回错误.
func (s *Switcher) Execute(ctx context.Context, req SwitchRequest) (SwitchResult, error) {
	action, err := DetermineAction(req.ModeBefore, req.ModeAfter)
	if err != nil {
		return SwitchResult{}, err
	}

	targetWindow, err := ParseOptionalWindowHandle(req.WindowHwnd)
	if err != nil {
		return SwitchResult{}, err
	}

	switch action {
	case ActionSwitchToChinese:
		if err := s.controller.SetOpenStatus(ctx, true, targetWindow); err != nil {
			return SwitchResult{}, err
		}
		return SwitchResult{Changed: true}, nil
	case ActionSwitchToEnglish:
		if err := s.controller.SetOpenStatus(ctx, false, targetWindow); err != nil {
			return SwitchResult{}, err
		}
		return SwitchResult{Changed: true}, nil
	default:
		return SwitchResult{Changed: false}, nil
	}
}

// ParseOptionalWindowHandle 解析可选窗口句柄参数.
// raw 支持十进制或 0x 前缀十六进制字符串, 空字符串表示不指定窗口.
// 返回解析后的窗口句柄, 空字符串时返回 0; 格式非法时返回错误.
func ParseOptionalWindowHandle(raw string) (uintptr, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, nil
	}

	value, err := strconv.ParseUint(trimmed, 0, strconv.IntSize)
	if err != nil {
		return 0, fmt.Errorf("invalid window-hwnd %q", raw)
	}

	if value == 0 {
		return 0, fmt.Errorf("invalid window-hwnd %q", raw)
	}

	return uintptr(value), nil
}
