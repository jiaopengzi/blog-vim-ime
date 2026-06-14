# FilePath    : blog-vim-ime\run.ps1
# Author      : jiaopengzi
# Blog        : https://jiaopengzi.com
# Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
# Description : Windows PowerShell build and run script for blog-vim-ime

param(
    [ValidateRange(0, 8)]
    [int]$Choice = -1
)

# Define executable names
$BINARY_SERVICE = "blog-vim-ime"
$BINARY_CLI = "blog-vim-ime-cli"
$OUTPUT_DIR = ".\bin"
$ICON_FILE = ".\app.ico"

# SelectOperation 返回本次脚本执行的菜单选项.
# choice 为可选的非交互输入, 当为 -1 时回退到交互式选择.
# 返回值 int, 表示后续 switch 使用的菜单编号.
function SelectOperation {
    param(
        [int]$Choice
    )

    if ($Choice -ge 0) {
        return $Choice
    }

    Write-Host ""
    Write-Host "blog-vim-ime - Windows Build Script" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Choose an operation:"
    Write-Host "  0 - Format code (go fmt)"
    Write-Host "  1 - Lint check (golangci-lint)"
    Write-Host "  2 - Run tests (go test)"
    Write-Host "  3 - Build service (blog-vim-ime.exe)"
    Write-Host "  4 - Build CLI (blog-vim-ime-cli.exe)"
    Write-Host "  5 - Build service and run"
    Write-Host "  6 - Run compiled service EXE"
    Write-Host "  7 - Full pipeline (Fmt -> Lint -> Test -> Build service)"
    Write-Host "  8 - Clean build artifacts"
    Write-Host ""

    $selected = Read-Host "Enter your choice"
    Write-Host ""

    return [int]$selected
}

$choice = SelectOperation -Choice $Choice

# Format code
function formatCode {
    Write-Host "Formatting code..." -ForegroundColor Yellow
    go fmt ./...
    Write-Host "Format complete" -ForegroundColor Green
}

# Lint check
function goLint {
    Write-Host "Running lint check..." -ForegroundColor Yellow
    golangci-lint run
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Lint check passed" -ForegroundColor Green
    } else {
        Write-Host "Lint check failed" -ForegroundColor Red
    }
}

# Run tests
function runTests {
    Write-Host "Running tests..." -ForegroundColor Yellow
    go test -v ./...
    if ($LASTEXITCODE -eq 0) {
        Write-Host "All tests passed" -ForegroundColor Green
    } else {
        Write-Host "Tests failed" -ForegroundColor Red
    }
}

# Ensure-Rsrc 确保 rsrc 工具可用, 不可用时尝试安装.
# 无参数.
# 无返回值; 安装失败时直接 return 中断调用方.
function Ensure-Rsrc {
    $rsrcAvailable = $false
    try {
        $rsrcOutput = rsrc -h 2>&1
        if ($LASTEXITCODE -eq 0 -or $rsrcOutput -like "*usage*") {
            $rsrcAvailable = $true
        }
    } catch {
        $rsrcAvailable = $false
    }

    if (-not $rsrcAvailable) {
        Write-Host "rsrc tool not found, attempting to install..." -ForegroundColor Yellow
        Write-Host "Installing github.com/akavel/rsrc..." -ForegroundColor Cyan
        go get github.com/akavel/rsrc
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Error: Failed to install rsrc tool" -ForegroundColor Red
            return $false
        }
        Write-Host "rsrc installed successfully" -ForegroundColor Green
    }
    return $true
}

# Prepare-BuildDir 确保输出目录和必要资源文件就绪.
# 无参数.
# 当 app.ico 或 port.yaml 缺失时直接 return 中断调用方.
function Prepare-BuildDir {
    if (-not (Test-Path $OUTPUT_DIR)) {
        New-Item -ItemType Directory -Path $OUTPUT_DIR | Out-Null
    }

    if (-not (Test-Path "app.ico")) {
        Write-Host "Error: app.ico not found in project root" -ForegroundColor Red
        return $false
    }
    if (-not (Test-Path "port.yaml")) {
        Write-Host "Error: port.yaml not found in project root" -ForegroundColor Red
        return $false
    }

    # 清理旧版本遗留的 cmd 目录图标资源.
    Remove-Item -Force "cmd\blog-vim-ime\app.png" -ErrorAction SilentlyContinue
    Remove-Item -Force "cmd\blog-vim-ime-cli\app.png" -ErrorAction SilentlyContinue

    return $true
}

# Build-Single 编译单个 Go 入口目录并嵌入图标.
# CmdDir 为 cmd 子目录名 (如 blog-vim-ime), BinaryName 为输出 EXE 名称.
# 编译失败时直接 return.
function Build-Single {
    param(
        [string]$CmdDir,
        [string]$BinaryName
    )

    Push-Location ".\cmd\$CmdDir"

    Write-Host "Generating icon resource..." -ForegroundColor Cyan
    rsrc -ico "../../app.ico" -o rsrc.syso
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Error: Icon generation failed for $CmdDir" -ForegroundColor Red
        Remove-Item -Force rsrc.syso -ErrorAction SilentlyContinue
        Pop-Location
        return $false
    }

    go build -o "..\..\$OUTPUT_DIR\$BinaryName.exe"
    $buildStatus = $LASTEXITCODE
    Remove-Item -Force rsrc.syso -ErrorAction SilentlyContinue
    Pop-Location

    if ($buildStatus -ne 0) {
        Write-Host "Error: Build failed for $CmdDir" -ForegroundColor Red
        return $false
    }

    return $true
}

# Build service (blog-vim-ime.exe)
function buildService {
    Write-Host "Building service (blog-vim-ime.exe)..." -ForegroundColor Yellow

    if (-not (Ensure-Rsrc)) { return }
    if (-not (Prepare-BuildDir)) { return }

    if (-not (Build-Single -CmdDir "blog-vim-ime" -BinaryName $BINARY_SERVICE)) { return }

    # 复制运行期配置到输出目录.
    Copy-Item -Path "port.yaml" -Destination "$OUTPUT_DIR\port.yaml" -Force

    Write-Host "Service build complete" -ForegroundColor Green
    Write-Host "Output: $OUTPUT_DIR\$BINARY_SERVICE.exe"
    Write-Host "        $OUTPUT_DIR\port.yaml"
}

# Build CLI (blog-vim-ime-cli.exe)
function buildCLI {
    Write-Host "Building CLI (blog-vim-ime-cli.exe)..." -ForegroundColor Yellow

    if (-not (Ensure-Rsrc)) { return }
    if (-not (Prepare-BuildDir)) { return }

    if (-not (Build-Single -CmdDir "blog-vim-ime-cli" -BinaryName $BINARY_CLI)) { return }

    Write-Host "CLI build complete" -ForegroundColor Green
    Write-Host "Output: $OUTPUT_DIR\$BINARY_CLI.exe"
}

# Build service and run
function buildAndRun {
    Write-Host "Building and running..." -ForegroundColor Yellow
    buildService

    if ($LASTEXITCODE -eq 0) {
        Write-Host ""
        Write-Host "Starting service..." -ForegroundColor Cyan
        & "$OUTPUT_DIR\$BINARY_SERVICE.exe"
    }
}

# Run compiled binary
function runCompiledBinary {
    $exePath = "$OUTPUT_DIR\$BINARY_SERVICE.exe"

    if (-not (Test-Path $exePath)) {
        Write-Host "Error: EXE not found: $exePath" -ForegroundColor Red
        Write-Host "Please build first (option 3)" -ForegroundColor Yellow
        return
    }

    Write-Host "Starting service..." -ForegroundColor Cyan
    Write-Host "Service listening on: http://127.0.0.1:8765" -ForegroundColor Cyan
    Write-Host ""
    & $exePath
}

# Clean artifacts
function clean {
    Write-Host "Cleaning artifacts..." -ForegroundColor Yellow

    go clean

    if (Test-Path $OUTPUT_DIR) {
        Remove-Item -Recurse -Force $OUTPUT_DIR
        Write-Host "Removed $OUTPUT_DIR directory" -ForegroundColor Green
    }

    Remove-Item -Force rsrc_windows_amd64.syso -ErrorAction SilentlyContinue
    Remove-Item -Force "cmd\blog-vim-ime\app.png" -ErrorAction SilentlyContinue
    Remove-Item -Force "cmd\blog-vim-ime-cli\app.png" -ErrorAction SilentlyContinue

    Write-Host "Cleanup complete" -ForegroundColor Green
}

# Full pipeline (builds service only)
function fullPipeline {
    Write-Host "Running full pipeline..." -ForegroundColor Yellow
    Write-Host ""

    formatCode
    Write-Host ""

    goLint
    Write-Host ""

    runTests
    Write-Host ""

    buildService
    Write-Host ""

    Write-Host "Full pipeline complete" -ForegroundColor Green
}

# Execute based on choice
switch ($choice) {
    0 { formatCode }
    1 { goLint }
    2 { runTests }
    3 { buildService }
    4 { buildCLI }
    5 { buildAndRun }
    6 { runCompiledBinary }
    7 { fullPipeline }
    8 { clean }
    default { Write-Host "Invalid choice" -ForegroundColor Red }
}

Write-Host ""
