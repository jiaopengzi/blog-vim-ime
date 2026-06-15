//
// FilePath    : blog-vim-ime\internal\ime\mode.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : Vim 模式切换到 IME 动作的规则定义.
//

package ime

import "fmt"

const (
	ModeNormal  = "normal"
	ModeInsert  = "insert"
	ModeVisual  = "visual"
	ModeReplace = "replace"
	ModeCmd     = "cmd"
)

type TransitionAction int

const (
	ActionNone TransitionAction = iota
	ActionSwitchToEnglish
	ActionSwitchToChinese
)

// DetermineAction 根据目标模式决定输入法切换动作.
// modeBefore 和 modeAfter 支持 normal, insert, visual, replace, cmd 等模式.
// modeAfter 为 normal 或 visual 时切换到英文, 为 insert, replace, cmd 时切换到中文.
// 返回动作枚举; 当模式值不受支持时返回错误.
func DetermineAction(modeBefore, modeAfter string) (TransitionAction, error) {
	if !isSupportedMode(modeBefore) {
		return ActionNone, fmt.Errorf("unsupported mode-before %q", modeBefore)
	}

	if !isSupportedMode(modeAfter) {
		return ActionNone, fmt.Errorf("unsupported mode-after %q", modeAfter)
	}

	if usesEnglishIME(modeAfter) {
		return ActionSwitchToEnglish, nil
	}

	return ActionSwitchToChinese, nil
}

// usesEnglishIME 判断目标模式是否应保持英文输入法.
// mode 表示目标 Vim 模式.
// 返回 true 表示应切换到英文, false 表示应切换到中文.
func usesEnglishIME(mode string) bool {
	switch mode {
	case ModeNormal, ModeVisual:
		return true
	default:
		return false
	}
}

// isSupportedMode 判断模式是否在允许集合内.
// mode 表示待检查模式字符串.
// 返回 true 表示可识别, false 表示非法模式.
func isSupportedMode(mode string) bool {
	switch mode {
	case ModeNormal, ModeInsert, ModeVisual, ModeReplace, ModeCmd:
		return true
	default:
		return false
	}
}
