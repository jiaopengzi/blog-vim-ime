//
// FilePath    : blog-vim-ime\cmd\blog-vim-ime-cli\main.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : CLI 入口, 用于本地直接验证 IME 切换能力.
//

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"blog-vim-ime/internal/ime"
)

type cliOptions struct {
	modeBefore string
	modeAfter  string
	windowHwnd string
}

// parseFlags 解析 CLI 参数.
// 无参数, 通过标准 flag 包读取命令行.
// mode-before 和 mode-after 为可选参数; 缺省时默认切换到英文普通模式.
// 返回值为解析后的参数对象.
func parseFlags() cliOptions {
	options := cliOptions{}
	flag.StringVar(&options.modeBefore, "mode-before", "", "mode before transition: normal|insert|visual|replace|cmd, optional")
	flag.StringVar(&options.modeAfter, "mode-after", "", "mode after transition: normal|insert|visual|replace|cmd, optional")
	flag.StringVar(&options.windowHwnd, "window-hwnd", "", "target window handle, optional, support decimal or 0x hex")
	flag.Parse()
	return options
}

// main 运行 CLI 切换流程.
// 无参数.
// 参数非法或切换失败时退出码为 1.
// 当 mode-before 或 mode-after 缺省时, 兜底为切换到英文普通模式.
func main() {
	logger := log.New(os.Stdout, "[blog-vim-ime-cli] ", log.LstdFlags|log.Lmsgprefix)
	options := parseFlags()

	// 兜底逻辑: 如果参数缺省, 默认切换到英文普通模式.
	if options.modeBefore == "" || options.modeAfter == "" {
		options.modeBefore = ime.ModeInsert
		options.modeAfter = ime.ModeNormal
		logger.Println("mode-before or mode-after not specified, fallback to: insert -> normal (switch to English)")
	}

	controller, err := ime.NewController()
	if err != nil {
		logger.Fatalf("create ime controller failed: %v", err)
	}

	switcher := ime.NewSwitcher(controller)
	result, err := switcher.Execute(context.Background(), ime.SwitchRequest{
		ModeBefore: options.modeBefore,
		ModeAfter:  options.modeAfter,
		WindowHwnd: options.windowHwnd,
	})
	if err != nil {
		logger.Fatalf("execute switch failed: %v", err)
	}

	if result.Changed {
		fmt.Println("ok: switched")
		return
	}

	fmt.Println("ok: no-op")
}
