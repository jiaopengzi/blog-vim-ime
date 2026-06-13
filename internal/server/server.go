//
// FilePath    : blog-vim-ime\internal\server\server.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : IME HTTP 服务实现.
//

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"blog-vim-ime/internal/ime"
)

const (
	responseStatusError = "error"
	responseStatusOK    = "ok"
)

type Server struct {
	switcher *ime.Switcher
	logger   *log.Logger
	mux      *http.ServeMux
}

// IMETransitionRequest 定义模式切换请求体.
type IMETransitionRequest struct {
	ModeBefore string `json:"mode-before"`
	ModeAfter  string `json:"mode-after"`
	WindowHwnd string `json:"window-hwnd,omitempty"`
}

// IMEResponse 定义接口统一返回结构.
type IMEResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// New 创建并注册 IME HTTP 路由.
// controller 负责执行系统输入法切换, logger 用于日志输出.
// 返回值为可直接挂载的服务实例.
func New(controller ime.Controller, logger *log.Logger) *Server {
	s := &Server{
		switcher: ime.NewSwitcher(controller),
		logger:   logger,
		mux:      http.NewServeMux(),
	}

	s.mux.HandleFunc("/ime", s.handleIME)

	return s
}

// Routes 返回服务路由处理器.
// 无参数.
// 返回值为包含 /ime 路由与 CORS 中间件的 HTTP Handler.
func (s *Server) Routes() http.Handler {
	return corsMiddleware(s.mux)
}

// corsMiddleware 为所有请求添加 CORS 响应头, 允许跨域访问.
// handler 为内层路由处理器.
// 返回值为包装后的 HTTP Handler, 支持所有 Origin 的跨域请求.
func corsMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 允许来自任何源的跨域请求
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// 处理 OPTIONS 预检请求
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

// handleIME 处理模式切换请求并触发输入法动作.
// w 和 r 分别表示响应写入器与请求对象.
// 根据请求结果返回 200, 204, 400 或 500.
func (s *Server) handleIME(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, IMEResponse{
			Status:  responseStatusError,
			Message: "method not allowed",
		})
		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			s.logger.Printf("close request body failed: %v", err)
		}
	}()

	var req IMETransitionRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, IMEResponse{
			Status:  responseStatusError,
			Message: fmt.Sprintf("invalid request body: %v", err),
		})
		return
	}

	result, err := s.switcher.Execute(context.Background(), ime.SwitchRequest{
		ModeBefore: req.ModeBefore,
		ModeAfter:  req.ModeAfter,
		WindowHwnd: req.WindowHwnd,
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if isRequestValidationError(err) {
			statusCode = http.StatusBadRequest
		}

		writeJSON(w, statusCode, IMEResponse{
			Status:  responseStatusError,
			Message: err.Error(),
		})
		return
	}

	if !result.Changed {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	writeJSON(w, http.StatusOK, IMEResponse{Status: responseStatusOK})
}

// isRequestValidationError 判断错误是否属于请求参数校验失败.
// err 为执行层返回错误.
// 返回 true 表示应返回 400, false 表示应返回 500.
func isRequestValidationError(err error) bool {
	message := err.Error()
	return strings.Contains(message, "unsupported mode-before") ||
		strings.Contains(message, "unsupported mode-after") ||
		strings.Contains(message, "invalid window-hwnd")
}

// writeJSON 以统一格式写入 JSON 响应.
// status 为 HTTP 状态码, payload 为返回体结构.
// 该函数不返回错误, 编码失败时由 net/http 处理连接写入异常.
func writeJSON(w http.ResponseWriter, status int, payload IMEResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("encode response body failed: %v", err)
	}
}
