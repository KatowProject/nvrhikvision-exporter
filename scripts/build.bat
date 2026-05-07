@echo off
REM Cross-platform build script for Windows (and all platforms)
REM Usage: scripts\build.bat [version] [target]
REM Examples:
REM   scripts\build.bat                    # Build current platform (Windows)
REM   scripts\build.bat 1.0.0              # Build with version 1.0.0
REM   scripts\build.bat 1.0.0 all          # Build all platforms

setlocal enabledelayedexpansion

set "SCRIPT_DIR=%~dp0"
for %%I in ("%SCRIPT_DIR%..") do set "ROOT_DIR=%%~fI"

set VERSION=%1
if "!VERSION!"=="" set VERSION=0.1.0

set TARGET=%2
if "!TARGET!"=="" set TARGET=windows

set APP_NAME=nvrhikvision-exporter
set DIST_DIR=%ROOT_DIR%\dist
set ENTRY_POINT=./cmd/exporter

if not exist "!DIST_DIR!" mkdir "!DIST_DIR!"

pushd "!ROOT_DIR!" >nul

echo.
echo ===============================================
echo  Hikvision NVR Exporter - Cross-Platform Build
echo ===============================================
echo Version: !VERSION!
echo Target: !TARGET!
echo.

goto !TARGET!

:windows
echo [BUILD] Building for Windows/amd64...
set "OUTPUT=!DIST_DIR!\!APP_NAME!-windows-amd64.exe"
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0
go build -o "!OUTPUT!" -ldflags "-X main.Version=!VERSION! -s -w" "!ENTRY_POINT!"
if !ERRORLEVEL! EQU 0 (
    echo [SUCCESS] Built: !OUTPUT!
) else (
    echo [ERROR] Build failed
    popd >nul
    exit /b 1
)
goto end

:linux
echo [BUILD] Building for Linux/amd64...
set "OUTPUT=!DIST_DIR!\!APP_NAME!-linux-amd64"
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0
go build -o "!OUTPUT!" -ldflags "-X main.Version=!VERSION! -s -w" "!ENTRY_POINT!"
if !ERRORLEVEL! EQU 0 (
    echo [SUCCESS] Built: !OUTPUT!
) else (
    echo [ERROR] Build failed
    popd >nul
    exit /b 1
)
goto end

:darwin
echo [BUILD] Building for macOS/amd64 (Intel)...
set "OUTPUT1=!DIST_DIR!\!APP_NAME!-darwin-amd64"
set GOOS=darwin
set GOARCH=amd64
set CGO_ENABLED=0
go build -o "!OUTPUT1!" -ldflags "-X main.Version=!VERSION! -s -w" "!ENTRY_POINT!"
if !ERRORLEVEL! EQU 0 (
    echo [SUCCESS] Built: !OUTPUT1!
) else (
    echo [ERROR] Build failed
    popd >nul
    exit /b 1
)

echo [BUILD] Building for macOS/arm64 (Apple Silicon)...
set "OUTPUT2=!DIST_DIR!\!APP_NAME!-darwin-arm64"
set GOOS=darwin
set GOARCH=arm64
set CGO_ENABLED=0
go build -o "!OUTPUT2!" -ldflags "-X main.Version=!VERSION! -s -w" "!ENTRY_POINT!"
if !ERRORLEVEL! EQU 0 (
    echo [SUCCESS] Built: !OUTPUT2!
) else (
    echo [ERROR] Build failed
    popd >nul
    exit /b 1
)
goto end

:all
echo [BUILD] Building for Windows/amd64...
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0
go build -o "!DIST_DIR!\!APP_NAME!-windows-amd64.exe" -ldflags "-X main.Version=!VERSION! -s -w" "!ENTRY_POINT!"
if !ERRORLEVEL! NEQ 0 (echo [ERROR] Windows build failed & popd >nul & exit /b 1)
echo [SUCCESS] Built Windows/amd64

echo [BUILD] Building for Linux/amd64...
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0
go build -o "!DIST_DIR!\!APP_NAME!-linux-amd64" -ldflags "-X main.Version=!VERSION! -s -w" "!ENTRY_POINT!"
if !ERRORLEVEL! NEQ 0 (echo [ERROR] Linux build failed & popd >nul & exit /b 1)
echo [SUCCESS] Built Linux/amd64

echo [BUILD] Building for Linux/arm64...
set GOOS=linux
set GOARCH=arm64
set CGO_ENABLED=0
go build -o "!DIST_DIR!\!APP_NAME!-linux-arm64" -ldflags "-X main.Version=!VERSION! -s -w" "!ENTRY_POINT!"
if !ERRORLEVEL! NEQ 0 (echo [ERROR] Linux ARM build failed & popd >nul & exit /b 1)
echo [SUCCESS] Built Linux/arm64

echo [BUILD] Building for macOS/amd64 (Intel)...
set GOOS=darwin
set GOARCH=amd64
set CGO_ENABLED=0
go build -o "!DIST_DIR!\!APP_NAME!-darwin-amd64" -ldflags "-X main.Version=!VERSION! -s -w" "!ENTRY_POINT!"
if !ERRORLEVEL! NEQ 0 (echo [ERROR] macOS Intel build failed & popd >nul & exit /b 1)
echo [SUCCESS] Built macOS/amd64

echo [BUILD] Building for macOS/arm64 (Apple Silicon)...
set GOOS=darwin
set GOARCH=arm64
set CGO_ENABLED=0
go build -o "!DIST_DIR!\!APP_NAME!-darwin-arm64" -ldflags "-X main.Version=!VERSION! -s -w" "!ENTRY_POINT!"
if !ERRORLEVEL! NEQ 0 (echo [ERROR] macOS ARM build failed & popd >nul & exit /b 1)
echo [SUCCESS] Built macOS/arm64

goto end

:end
echo.
echo ===============================================
echo Build process completed!
echo Output directory: !DIST_DIR!
echo ===============================================
echo.
echo Files built:
dir /b "!DIST_DIR!"
echo.
echo Usage examples:
echo   !DIST_DIR!\!APP_NAME!-windows-amd64.exe -config=config.yaml
echo   !DIST_DIR!\!APP_NAME!-linux-amd64 -config=config.yaml
echo.

popd >nul
endlocal