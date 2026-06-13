//
// FilePath    : blog-vim-ime\internal\ime\switcher_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : IME 切换执行器的单元测试.
//

package ime

import (
	"context"
	"errors"
	"testing"
)

type mockController struct {
	lastOpen   bool
	lastWindow uintptr
	calls      int
	err        error
}

// SetOpenStatus 记录调用参数, 便于断言执行器行为.
// open 为目标输入法状态, targetWindow 为目标窗口句柄.
// 返回预设错误, 用于覆盖失败路径.
func (m *mockController) SetOpenStatus(_ context.Context, open bool, targetWindow uintptr) error {
	m.calls++
	m.lastOpen = open
	m.lastWindow = targetWindow
	return m.err
}

func TestSwitcherExecuteSwitchesChinese(t *testing.T) {
	t.Parallel()

	mock := &mockController{}
	switcher := NewSwitcher(mock)

	result, err := switcher.Execute(context.Background(), SwitchRequest{
		ModeBefore: ModeNormal,
		ModeAfter:  ModeInsert,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Changed {
		t.Fatalf("expected changed=true")
	}

	if mock.calls != 1 || !mock.lastOpen {
		t.Fatalf("expected one chinese switch call")
	}
}

func TestSwitcherExecuteSwitchesEnglishWithWindow(t *testing.T) {
	t.Parallel()

	mock := &mockController{}
	switcher := NewSwitcher(mock)

	result, err := switcher.Execute(context.Background(), SwitchRequest{
		ModeBefore: ModeInsert,
		ModeAfter:  ModeNormal,
		WindowHwnd: "0x1234",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Changed {
		t.Fatalf("expected changed=true")
	}

	if mock.calls != 1 || mock.lastOpen {
		t.Fatalf("expected one english switch call")
	}

	if mock.lastWindow != 0x1234 {
		t.Fatalf("expected target window 0x1234, got %d", mock.lastWindow)
	}
}

func TestSwitcherExecuteNoAction(t *testing.T) {
	t.Parallel()

	mock := &mockController{}
	switcher := NewSwitcher(mock)

	result, err := switcher.Execute(context.Background(), SwitchRequest{
		ModeBefore: ModeNormal,
		ModeAfter:  ModeNormal,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Changed {
		t.Fatalf("expected changed=false")
	}

	if mock.calls != 0 {
		t.Fatalf("expected no controller call, got %d", mock.calls)
	}
}

func TestSwitcherExecuteInvalidWindow(t *testing.T) {
	t.Parallel()

	mock := &mockController{}
	switcher := NewSwitcher(mock)

	_, err := switcher.Execute(context.Background(), SwitchRequest{
		ModeBefore: ModeInsert,
		ModeAfter:  ModeNormal,
		WindowHwnd: "invalid",
	})
	if err == nil {
		t.Fatalf("expected error for invalid window handle")
	}
}

func TestSwitcherExecutePropagatesControllerError(t *testing.T) {
	t.Parallel()

	mock := &mockController{err: errors.New("switch failed")}
	switcher := NewSwitcher(mock)

	_, err := switcher.Execute(context.Background(), SwitchRequest{
		ModeBefore: ModeInsert,
		ModeAfter:  ModeNormal,
	})
	if err == nil {
		t.Fatalf("expected controller error")
	}
}

func TestSwitcherExecuteNormalToVisualSwitchChinese(t *testing.T) {
	t.Parallel()

	mock := &mockController{}
	switcher := NewSwitcher(mock)

	result, err := switcher.Execute(context.Background(), SwitchRequest{
		ModeBefore: ModeNormal,
		ModeAfter:  ModeVisual,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Changed {
		t.Fatalf("expected changed=true")
	}

	if mock.calls != 1 || !mock.lastOpen {
		t.Fatalf("expected one chinese switch call")
	}
}

func TestSwitcherExecuteVisualToNormalSwitchEnglish(t *testing.T) {
	t.Parallel()

	mock := &mockController{}
	switcher := NewSwitcher(mock)

	result, err := switcher.Execute(context.Background(), SwitchRequest{
		ModeBefore: ModeVisual,
		ModeAfter:  ModeNormal,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Changed {
		t.Fatalf("expected changed=true")
	}

	if mock.calls != 1 || mock.lastOpen {
		t.Fatalf("expected one english switch call")
	}
}

func TestSwitcherExecuteInsertToReplaceSwitchChinese(t *testing.T) {
	t.Parallel()

	mock := &mockController{}
	switcher := NewSwitcher(mock)

	result, err := switcher.Execute(context.Background(), SwitchRequest{
		ModeBefore: ModeInsert,
		ModeAfter:  ModeReplace,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Changed {
		t.Fatalf("expected changed=true")
	}

	if mock.calls != 1 || !mock.lastOpen {
		t.Fatalf("expected one chinese switch call")
	}
}
