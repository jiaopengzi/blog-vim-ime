# blog-vim-ime

一个本地运行的 Go HTTP 服务，用于接收本地网络请求信号并自动切换输入法中英文状态, 用于 blog 项目的 vim 模式。

## 功能特性

- ✓ Windows 窗口级输入法状态控制
- ✓ 支持 5 种 Vim 模式切换（normal, insert, visual, replace, cmd）
- ✓ 跨域请求支持（CORS）
- ✓ CLI 独立工具用于本地测试
- ✓ 后台运行，系统托盘显示程序图标，无控制台窗口
- ✓ 托盘菜单支持快捷退出

## 配置端口

默认监听 `127.0.0.1:8765`。可在项目根目录放置 `port.yaml` 覆盖端口：

```yaml
port: 8765
```

## 启动

### 开发模式（显示控制台输出）

```powershell
go run .\cmd\blog-vim-ime
```

### 生产模式（后台运行，系统托盘显示）

直接运行编译后的 EXE 文件，程序启动后在系统托盘显示，点击托盘图标菜单可选择退出：

```powershell
.\bin\blog-vim-ime.exe
```

## 构建与运行

使用统一的 PowerShell 脚本 `run.ps1` 进行所有构建和运行操作：

```powershell
.\run.ps1
```

### 可用操作

| 选项 | 操作描述 |
|------|--------|
| 0 | 格式化代码 (go fmt) |
| 1 | Lint 检查 (golangci-lint) |
| 2 | 运行单元测试 (go test) |
| 3 | 编译构建（含图标）|
| 4 | 编译运行服务 |
| 5 | 运行已编译的服务 EXE |
| 6 | 清理编译产物 |
| 7 | 完整流程（Fmt -> Lint -> Test -> Build）|

### 示例用法

#### 快速开发构建（含图标）

```powershell
.\run.ps1
# 选择 3（编译构建含图标）
```

#### 完整流程构建

```powershell
.\run.ps1
# 选择 7（完整流程）
```

#### 编译并运行

```powershell
.\run.ps1
# 选择 5（编译运行服务）
```

#### 添加应用程序图标

项目根目录包含：

- `app.ico` - 用于 EXE 文件窗口, 文件管理器显示, 以及 Windows 系统托盘图标（via rsrc 工具和 embed 指令）
- `resource.rc` - Windows 资源定义文件

构建过程会自动：

1. 从项目根目录读取 `app.ico`
2. 通过 rsrc 工具将 `app.ico` 嵌入到 EXE 文件
3. 在运行时从嵌入数据中加载托盘图标

运行脚本选择含图标构建：

```powershell
.\run.ps1
# 选择 3（编译构建含图标）
```

### 手动构建（不含图标）

```powershell
go build -o .\bin\blog-vim-ime.exe .\cmd\blog-vim-ime
go build -o .\bin\blog-vim-ime-cli.exe .\cmd\blog-vim-ime-cli
```

## Lint 校验

```bash
golangci-lint run
```

## 接口

### CORS 支持

服务已启用跨域资源共享（CORS），允许浏览器从任何源发起请求。响应头包括：

```text
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type
```

### POST /ime

```http
POST http://127.0.0.1:8765/ime
Content-Type: application/json

{
  "mode-before": "normal",
  "mode-after": "insert",
  "window-hwnd": "0x00123456"
}
```

`window-hwnd` 为可选字段, 用于指定要切换输入法状态的窗口句柄（十进制或 `0x` 十六进制字符串）; 未传时默认使用当前前台窗口。

支持的 Vim 模式：`normal`, `insert`, `visual`, `replace`, `cmd`

切换规则：

1. 进入 `normal` 模式：切到英文输入法（兜底）。
2. 离开 `normal` 模式（进入其他模式）：切到中文输入法。
3. 在同一模式内转换：不做输入法切换（返回 204）。

## 运行测试

```bash
go test ./...
```

## CLI 直接测试（推荐先验证）

支持的模式：`normal`, `insert`, `visual`, `replace`, `cmd`

无参数运行（兜底到英文普通模式）:

```powershell
go run .\cmd\blog-vim-ime-cli
```

切到英文输入法（insert -> normal）:

```powershell
go run .\cmd\blog-vim-ime-cli --mode-before insert --mode-after normal
```

切到中文输入法（normal -> insert）:

```powershell
go run .\cmd\blog-vim-ime-cli --mode-before normal --mode-after insert
```

切到中文输入法（normal -> visual）:

```powershell
go run .\cmd\blog-vim-ime-cli --mode-before normal --mode-after visual
```

切到英文输入法（visual -> normal）:

```powershell
go run .\cmd\blog-vim-ime-cli --mode-before visual --mode-after normal
```

指定窗口句柄测试:

```powershell
go run .\cmd\blog-vim-ime-cli --mode-before normal --mode-after insert --window-hwnd 0x00123456
```

## PowerShell curl 联调示例

切换到中文输入法（normal -> insert）:

```powershell
$body = @{
  "mode-before" = "normal"
  "mode-after"  = "insert"
} | ConvertTo-Json

Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:8765/ime" -ContentType "application/json" -Body $body
```

指定窗口句柄切换（推荐由前端传入目标窗口句柄）:

```powershell
$body = @{
  "mode-before" = "insert"
  "mode-after"  = "normal"
  "window-hwnd" = "0x00123456"
} | ConvertTo-Json

Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:8765/ime" -ContentType "application/json" -Body $body
```

切换到英文输入法（insert -> normal）:

```powershell
$body = @{
  "mode-before" = "insert"
  "mode-after"  = "normal"
} | ConvertTo-Json

Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:8765/ime" -ContentType "application/json" -Body $body
```
