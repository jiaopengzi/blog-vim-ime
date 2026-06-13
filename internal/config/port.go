//
// FilePath    : blog-vim-ime\internal\config\port.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 端口配置加载器, 从本地 YAML 文件读取端口.
//

package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	minPort = 1
	maxPort = 65535
)

// LoadPort 从指定 YAML 文件读取端口配置.
// path 表示配置文件路径, defaultPort 表示文件不存在时使用的默认端口.
// 返回值为最终端口; 当文件读取失败或端口格式非法时返回错误.
func LoadPort(path string, defaultPort int) (int, error) {
	if !isValidPort(defaultPort) {
		return 0, fmt.Errorf("default port %d out of range", defaultPort)
	}

	// #nosec G304 -- path 由主程序固定传入本地 port.yaml, 不接受远程输入.
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultPort, nil
		}

		return 0, fmt.Errorf("read %s: %w", path, err)
	}

	port, err := parsePortYAML(string(data))
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", path, err)
	}

	return port, nil
}

// parsePortYAML 解析 YAML 文本中的 port 字段.
// content 表示配置文件内容.
// 返回解析得到的端口; 当缺少有效值或格式错误时返回错误.
func parsePortYAML(content string) (int, error) {
	lines := strings.SplitSeq(content, "\n")
	for line := range lines {
		trimmed := strings.TrimSpace(strings.TrimPrefix(line, "\ufeff"))
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if strings.Contains(trimmed, ":") {
			key, value, found := strings.Cut(trimmed, ":")
			if !found {
				continue
			}

			if strings.TrimSpace(key) != "port" {
				continue
			}

			return parsePortValue(value)
		}

		return parsePortValue(trimmed)
	}

	return 0, fmt.Errorf("missing port value")
}

// parsePortValue 将字符串端口值转换为整数并校验范围.
// value 表示原始端口字符串.
// 返回合法端口; 当转换失败或超范围时返回错误.
func parsePortValue(value string) (int, error) {
	portText := strings.TrimSpace(value)
	port, err := strconv.Atoi(portText)
	if err != nil {
		return 0, fmt.Errorf("invalid port %q", portText)
	}

	if !isValidPort(port) {
		return 0, fmt.Errorf("port %d out of range", port)
	}

	return port, nil
}

// isValidPort 判断端口是否位于 TCP 可用范围内.
// port 表示待校验端口.
// 返回值为 true 表示合法, false 表示非法.
func isValidPort(port int) bool {
	return port >= minPort && port <= maxPort
}
