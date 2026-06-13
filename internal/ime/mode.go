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

// DetermineAction 根据模式变化计算需要执行的输入法动作.
// modeBefore 和 modeAfter 支持 normal, insert, visual, replace, cmd 等模式.
// 切换规则: normal 模式使用英文输入法; 其他模式使用中文输入法.
// 返回动作枚举; 当模式值不受支持时返回错误.
func DetermineAction(modeBefore, modeAfter string) (TransitionAction, error) {
	if !isSupportedMode(modeBefore) {
		return ActionNone, fmt.Errorf("unsupported mode-before %q", modeBefore)
	}

	if !isSupportedMode(modeAfter) {
		return ActionNone, fmt.Errorf("unsupported mode-after %q", modeAfter)
	}

	if modeBefore == modeAfter {
		return ActionNone, nil
	}

	// normal 模式下切换到英文; 其他模式下切换到中文.
	if modeAfter == ModeNormal {
		return ActionSwitchToEnglish, nil
	}

	return ActionSwitchToChinese, nil
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
