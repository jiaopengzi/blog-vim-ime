# FilePath    : blog-vim-ime\run.ps1
# Author      : jiaopengzi
# Blog        : https://jiaopengzi.com
# Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
# Description : Windows PowerShell build and run script for blog-vim-ime

# Define executable names
$BINARY_SERVICE = "blog-vim-ime"
$BINARY_CLI = "blog-vim-ime-cli"
$OUTPUT_DIR = ".\bin"
$ICON_FILE = ".\app.ico"

# Display menu
Write-Host ""
Write-Host "blog-vim-ime - Windows Build Script" -ForegroundColor Cyan
Write-Host ""
Write-Host "Choose an operation:"
Write-Host "  0 - Format code (go fmt)"
Write-Host "  1 - Lint check (golangci-lint)"
Write-Host "  2 - Run tests (go test)"
Write-Host "  3 - Build with icon"
Write-Host "  4 - Build and run"
Write-Host "  5 - Run compiled EXE"
Write-Host "  6 - Clean build artifacts"
Write-Host "  7 - Full pipeline (Fmt -> Lint -> Test -> Build)"
Write-Host ""

# Get user choice
$choice = Read-Host "Enter your choice"
Write-Host ""

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

# Build with icon
function buildWithIcon {
    Write-Host "Building with icon..." -ForegroundColor Yellow
    
    if (-not (Test-Path $OUTPUT_DIR)) {
        New-Item -ItemType Directory -Path $OUTPUT_DIR | Out-Null
    }

    # Check rsrc availability and install if needed
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
            return
        }
        Write-Host "rsrc installed successfully" -ForegroundColor Green
    }

    # 检查项目根目录的 app.ico
    if (-not (Test-Path "app.ico")) {
        Write-Host "Error: app.ico not found in project root" -ForegroundColor Red
        return
    }

    # 清理旧版本遗留的 cmd 目录图标资源, 保持项目内仅保留根目录资源文件.
    Remove-Item -Force "cmd\blog-vim-ime\app.png" -ErrorAction SilentlyContinue
    Remove-Item -Force "cmd\blog-vim-ime-cli\app.png" -ErrorAction SilentlyContinue

    # Build service with icon
    Write-Host "Building service with icon..." -ForegroundColor Cyan
    Push-Location .\cmd\blog-vim-ime
    
    Write-Host "Generating icon resource..." -ForegroundColor Cyan
    rsrc -ico "../../app.ico" -o rsrc.syso
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Error: Icon generation failed for service" -ForegroundColor Red
        Remove-Item -Force rsrc.syso -ErrorAction SilentlyContinue
        Pop-Location
        return
    }
    
    go build -o "..\..\$OUTPUT_DIR\$BINARY_SERVICE.exe"
    $buildStatus = $LASTEXITCODE
    Remove-Item -Force rsrc.syso -ErrorAction SilentlyContinue
    Pop-Location
    
    if ($buildStatus -ne 0) {
        Write-Host "Error: Service build failed" -ForegroundColor Red
        return
    }

    # Build CLI with icon
    Write-Host "Building CLI with icon..." -ForegroundColor Cyan
    Push-Location .\cmd\blog-vim-ime-cli
    
    Write-Host "Generating icon resource..." -ForegroundColor Cyan
    rsrc -ico "../../app.ico" -o rsrc.syso
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Error: Icon generation failed for CLI" -ForegroundColor Red
        Remove-Item -Force rsrc.syso -ErrorAction SilentlyContinue
        Pop-Location
        return
    }
    
    go build -o "..\..\$OUTPUT_DIR\$BINARY_CLI.exe"
    $buildStatus = $LASTEXITCODE
    Remove-Item -Force rsrc.syso -ErrorAction SilentlyContinue
    Pop-Location
    
    if ($buildStatus -ne 0) {
        Write-Host "Error: CLI build failed" -ForegroundColor Red
        return
    }

    Write-Host "Build complete with icon" -ForegroundColor Green
    Write-Host "Output:"
    Write-Host "  - $OUTPUT_DIR\$BINARY_SERVICE.exe"
    Write-Host "  - $OUTPUT_DIR\$BINARY_CLI.exe"
}

# Build and run
function buildAndRun {
    Write-Host "Building and running..." -ForegroundColor Yellow
    buildWithIcon
    
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
        Write-Host "Please build first" -ForegroundColor Yellow
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

# Full pipeline
function fullPipeline {
    Write-Host "Running full pipeline..." -ForegroundColor Yellow
    Write-Host ""
    
    formatCode
    Write-Host ""
    
    goLint
    Write-Host ""
    
    runTests
    Write-Host ""
    
    buildWithIcon
    Write-Host ""
    
    Write-Host "Full pipeline complete" -ForegroundColor Green
}

# Execute based on choice
switch ($choice) {
    0 { formatCode }
    1 { goLint }
    2 { runTests }
    3 { buildWithIcon }
    4 { buildAndRun }
    5 { runCompiledBinary }
    6 { clean }
    7 { fullPipeline }
    default { Write-Host "Invalid choice" -ForegroundColor Red }
}

Write-Host ""
