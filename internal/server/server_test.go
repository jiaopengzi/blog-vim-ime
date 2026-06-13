//
// FilePath    : blog-vim-ime\internal\server\server_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : IME HTTP 服务的单元测试.
//

package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockController struct {
	lastOpen   bool
	lastWindow uintptr
	calls      int
	err        error
}

// SetOpenStatus 记录调用参数, 用于断言服务层传参.
// open 为目标输入法开关状态, targetWindow 为目标窗口句柄.
// 返回预设错误, 用于覆盖错误分支测试.
func (m *mockController) SetOpenStatus(_ context.Context, open bool, targetWindow uintptr) error {
	m.calls++
	m.lastOpen = open
	m.lastWindow = targetWindow
	return m.err
}

func TestHandleIME_NormalToInsertSwitchesChinese(t *testing.T) {
	t.Parallel()

	mock := &mockController{}
	s := New(mock, log.New(io.Discard, "", 0))

	req := httptest.NewRequest(http.MethodPost, "/ime", bytes.NewBufferString(`{"mode-before":"normal","mode-after":"insert"}`))
	rec := httptest.NewRecorder()

	s.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if mock.calls != 1 {
		t.Fatalf("expected controller to be called once, got %d", mock.calls)
	}

	if !mock.lastOpen {
		t.Fatalf("expected chinese(open=true), got open=false")
	}

	if mock.lastWindow != 0 {
		t.Fatalf("expected default window handle 0, got %d", mock.lastWindow)
	}
}

func TestHandleIME_InsertToNormalSwitchesEnglish(t *testing.T) {
	t.Parallel()

	mock := &mockController{}
	s := New(mock, log.New(io.Discard, "", 0))

	req := httptest.NewRequest(http.MethodPost, "/ime", bytes.NewBufferString(`{"mode-before":"insert","mode-after":"normal"}`))
	rec := httptest.NewRecorder()

	s.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if mock.calls != 1 {
		t.Fatalf("expected controller to be called once, got %d", mock.calls)
	}

	if mock.lastOpen {
		t.Fatalf("expected english(open=false), got open=true")
	}

	if mock.lastWindow != 0 {
		t.Fatalf("expected default window handle 0, got %d", mock.lastWindow)
	}
}

func TestHandleIME_NoTransitionReturnsNoContent(t *testing.T) {
	t.Parallel()

	mock := &mockController{}
	s := New(mock, log.New(io.Discard, "", 0))

	req := httptest.NewRequest(http.MethodPost, "/ime", bytes.NewBufferString(`{"mode-before":"normal","mode-after":"normal"}`))
	rec := httptest.NewRecorder()

	s.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	if mock.calls != 0 {
		t.Fatalf("expected controller not to be called, got %d", mock.calls)
	}
}

func TestHandleIME_RejectsInvalidMethod(t *testing.T) {
	t.Parallel()

	mock := &mockController{}
	s := New(mock, log.New(io.Discard, "", 0))

	req := httptest.NewRequest(http.MethodGet, "/ime", nil)
	rec := httptest.NewRecorder()

	s.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleIME_RejectsInvalidBody(t *testing.T) {
	t.Parallel()

	mock := &mockController{}
	s := New(mock, log.New(io.Discard, "", 0))

	req := httptest.NewRequest(http.MethodPost, "/ime", bytes.NewBufferString(`{"mode-before":"normal"`))
	rec := httptest.NewRecorder()

	s.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleIME_RejectsUnsupportedMode(t *testing.T) {
	t.Parallel()

	mock := &mockController{}
	s := New(mock, log.New(io.Discard, "", 0))

	req := httptest.NewRequest(http.MethodPost, "/ime", bytes.NewBufferString(`{"mode-before":"unknown","mode-after":"insert"}`))
	rec := httptest.NewRecorder()

	s.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleIME_PropagatesControllerError(t *testing.T) {
	t.Parallel()

	mock := &mockController{err: errors.New("switch failed")}
	s := New(mock, log.New(io.Discard, "", 0))

	req := httptest.NewRequest(http.MethodPost, "/ime", bytes.NewBufferString(`{"mode-before":"normal","mode-after":"insert"}`))
	rec := httptest.NewRecorder()

	s.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestHandleIME_WithWindowHandlePassesTarget(t *testing.T) {
	t.Parallel()

	mock := &mockController{}
	s := New(mock, log.New(io.Discard, "", 0))

	req := httptest.NewRequest(http.MethodPost, "/ime", bytes.NewBufferString(`{"mode-before":"insert","mode-after":"normal","window-hwnd":"0x1234"}`))
	rec := httptest.NewRecorder()

	s.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if mock.lastWindow != 0x1234 {
		t.Fatalf("expected window handle 0x1234, got %d", mock.lastWindow)
	}
}

func TestHandleIME_RejectsInvalidWindowHandle(t *testing.T) {
	t.Parallel()

	mock := &mockController{}
	s := New(mock, log.New(io.Discard, "", 0))

	req := httptest.NewRequest(http.MethodPost, "/ime", bytes.NewBufferString(`{"mode-before":"insert","mode-after":"normal","window-hwnd":"bad-hwnd"}`))
	rec := httptest.NewRecorder()

	s.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCORSHeaders(t *testing.T) {
	t.Parallel()

	mock := &mockController{}
	s := New(mock, log.New(io.Discard, "", 0))

	req := httptest.NewRequest(http.MethodPost, "/ime", bytes.NewBufferString(`{"mode-before":"normal","mode-after":"insert"}`))
	rec := httptest.NewRecorder()

	s.Routes().ServeHTTP(rec, req)

	if origin := rec.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Fatalf("expected Access-Control-Allow-Origin *, got %q", origin)
	}

	if methods := rec.Header().Get("Access-Control-Allow-Methods"); methods != "GET, POST, OPTIONS" {
		t.Fatalf("expected Access-Control-Allow-Methods 'GET, POST, OPTIONS', got %q", methods)
	}

	if headers := rec.Header().Get("Access-Control-Allow-Headers"); headers != "Content-Type" {
		t.Fatalf("expected Access-Control-Allow-Headers 'Content-Type', got %q", headers)
	}
}

func TestCORSPreflight(t *testing.T) {
	t.Parallel()

	mock := &mockController{}
	s := New(mock, log.New(io.Discard, "", 0))

	req := httptest.NewRequest(http.MethodOptions, "/ime", nil)
	rec := httptest.NewRecorder()

	s.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for OPTIONS, got %d", rec.Code)
	}

	if origin := rec.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Fatalf("expected Access-Control-Allow-Origin *, got %q", origin)
	}
}
