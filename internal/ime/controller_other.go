//
// FilePath    : blog-vim-ime\internal\ime\controller_other.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 非 Windows 平台的 IME 控制器占位实现.
//

//go:build !windows

package ime

import (
	"context"
	"fmt"
)

type UnsupportedController struct{}

// NewController 创建非 Windows 平台下的占位控制器.
// 返回值用于统一接口行为.
// 当前实现不会在构造阶段返回错误.
func NewController() (Controller, error) {
	return &UnsupportedController{}, nil
}

// SetOpenStatus 在非 Windows 平台直接返回不支持错误.
// 参数会被忽略, 仅用于满足统一接口, targetWindow 也不会生效.
// 返回固定错误以提示平台限制.
func (c *UnsupportedController) SetOpenStatus(_ context.Context, _ bool, _ uintptr) error {
	return fmt.Errorf("ime switch is only supported on windows")
}
