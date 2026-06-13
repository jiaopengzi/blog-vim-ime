//
// FilePath    : blog-vim-ime\internal\ime\controller.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : IME 控制器接口定义.
//

package ime

import "context"

// Controller 定义输入法开关控制能力.
type Controller interface {
	// SetOpenStatus 设置 IME 打开状态.
	// open 为 true 表示切到中文输入, false 表示切到英文输入, targetWindow 为目标窗口句柄.
	// 当 targetWindow 为 0 时, 自动使用当前前台窗口.
	// 当切换失败或上下文取消时返回错误.
	SetOpenStatus(ctx context.Context, open bool, targetWindow uintptr) error
}
