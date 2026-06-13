package trayassets

import _ "embed"

// defaultTrayIcon 保存项目根目录中的 Windows 托盘 ICO 图标数据.
//
//go:embed app.ico
var defaultTrayIcon []byte

// DefaultTrayIcon 返回嵌入的托盘图标副本.
// 无参数.
// 返回值 []byte, Windows systray 所需的 ICO 图标内容; 当资源为空时返回 nil.
func DefaultTrayIcon() []byte {
	if len(defaultTrayIcon) == 0 {
		return nil
	}

	iconData := make([]byte, len(defaultTrayIcon))
	copy(iconData, defaultTrayIcon)

	return iconData
}
