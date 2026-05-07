#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Cross-platform build script for Hikvision NVR Exporter
.DESCRIPTION
    Build Go binaries for Windows, Linux, and macOS from Windows with PowerShell
.PARAMETER Version
    Version string for the binary (default: 0.1.0)
.PARAMETER Target
    Target platform(s) - supported: all, windows, linux, darwin, windows/amd64, linux/amd64, linux/arm64, darwin/amd64, darwin/arm64
.EXAMPLE
    .\scripts\build.ps1
#>

param(
    [string]$Version = "0.1.0",
    [string]$Target = "windows"
)

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RootDir = Resolve-Path (Join-Path $ScriptDir "..")

$APP_NAME = "nvrhikvision-exporter"
$DIST_DIR = Join-Path $RootDir "dist"
$ENTRY_POINT = "./cmd/exporter"

$colors = @{
    'header' = 'Cyan'
    'success' = 'Green'
    'error' = 'Red'
    'info' = 'Blue'
}

function Write-Header {
    param([string]$Message)
    Write-Host ""
    Write-Host "========================================" -ForegroundColor $colors.header
    Write-Host $Message -ForegroundColor $colors.header
    Write-Host "========================================" -ForegroundColor $colors.header
    Write-Host ""
}

function Write-Info {
    param([string]$Message)
    Write-Host "INFO: $Message" -ForegroundColor $colors.info
}

function Write-Success {
    param([string]$Message)
    Write-Host "OK: $Message" -ForegroundColor $colors.success
}

function Write-Error {
    param([string]$Message)
    Write-Host "ERROR: $Message" -ForegroundColor $colors.error
}

function Build-Target {
    param(
        [string]$OS,
        [string]$Arch,
        [string]$Suffix
    )

    Write-Info "Building for $OS/$Arch..."

    $Output = Join-Path $DIST_DIR "$APP_NAME-$OS-$Arch$Suffix"
    $env:GOOS = $OS
    $env:GOARCH = $Arch
    $env:CGO_ENABLED = "0"

    $LdFlags = "-X main.Version=$Version -s -w"
    & go build -o $Output -ldflags $LdFlags $ENTRY_POINT

    if ($LASTEXITCODE -eq 0) {
        $FileSize = (Get-Item $Output).Length / 1MB
        Write-Success "Built: $Output ($([math]::Round($FileSize, 2)) MB)"
    } else {
        Write-Error "Build failed for $OS/$Arch"
        exit 1
    }
}

Write-Header "Hikvision NVR Exporter - Cross-Platform Build"
Write-Host "Version: $Version" -ForegroundColor $colors.info
Write-Host "Target: $Target" -ForegroundColor $colors.info

if (-not (Test-Path $DIST_DIR)) {
    New-Item -ItemType Directory -Force -Path $DIST_DIR | Out-Null
    Write-Success "Created directory: $DIST_DIR"
}

Push-Location $RootDir
try {
    switch ($Target) {
        "all" {
            Build-Target "windows" "amd64" ".exe"
            Build-Target "windows" "arm64" ".exe"
            Build-Target "linux" "amd64" ""
            Build-Target "linux" "arm64" ""
            Build-Target "darwin" "amd64" ""
            Build-Target "darwin" "arm64" ""
        }
        "windows" {
            Build-Target "windows" "amd64" ".exe"
            Build-Target "windows" "arm64" ".exe"
        }
        "windows/amd64" { Build-Target "windows" "amd64" ".exe" }
        "windows/arm64" { Build-Target "windows" "arm64" ".exe" }
        "linux" {
            Build-Target "linux" "amd64" ""
            Build-Target "linux" "arm64" ""
        }
        "linux/amd64" { Build-Target "linux" "amd64" "" }
        "linux/arm64" { Build-Target "linux" "arm64" "" }
        "darwin" {
            Build-Target "darwin" "amd64" ""
            Build-Target "darwin" "arm64" ""
        }
        "darwin/amd64" { Build-Target "darwin" "amd64" "" }
        "darwin/arm64" { Build-Target "darwin" "arm64" "" }
        default {
            Write-Error "Unknown target: $Target"
            Write-Host "Supported targets: all, windows, linux, darwin, windows/amd64, windows/arm64, linux/amd64, linux/arm64, darwin/amd64, darwin/arm64"
            exit 1
        }
    }
}
finally {
    Pop-Location
}

Write-Header "Build Complete!"
Write-Host "Output directory: $DIST_DIR"
Write-Info "Files built:"
Get-ChildItem $DIST_DIR | ForEach-Object {
    $Size = $_.Length / 1MB
    Write-Host "  $($_.Name) ($([math]::Round($Size, 2)) MB)"
}

Write-Host ""
Write-Info "Usage examples:"
Write-Host "  .\dist\nvrhikvision-exporter-windows-amd64.exe -config=config.yaml"
Write-Host "  .\dist\nvrhikvision-exporter-linux-amd64 -config=config.yaml"
Write-Host "  .\dist\nvrhikvision-exporter-darwin-arm64 -config=config.yaml"
Write-Host ""