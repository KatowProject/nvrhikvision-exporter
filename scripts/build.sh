#!/usr/bin/env bash
# Cross-platform build script for Linux, macOS, and Windows
# Usage: ./scripts/build.sh [version] [target]
# Examples:
#   ./scripts/build.sh                    # Build all platforms, v0.1.0
#   ./scripts/build.sh 1.0.0              # Build all platforms, v1.0.0
#   ./scripts/build.sh 1.0.0 linux/amd64  # Build a specific platform

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

VERSION="${1:-0.1.0}"
TARGET="${2:-all}"
APP_NAME="nvrhikvision-exporter"
DIST_DIR="$ROOT_DIR/dist"
ENTRY_POINT="./cmd/exporter"

# ANSI Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_header() {
    echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║${NC} $1"
    echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

mkdir -p "$DIST_DIR"
cd "$ROOT_DIR"

print_header "Hikvision NVR Exporter - Cross-Platform Build"
echo "Version: $VERSION"
echo "Target: $TARGET"
echo ""

build_target() {
    local os=$1
    local arch=$2
    local suffix=$3

    print_info "Building for $os/$arch..."

    local output="$DIST_DIR/${APP_NAME}-${os}-${arch}${suffix}"

    GOOS=$os GOARCH=$arch CGO_ENABLED=0 \
        go build \
            -o "$output" \
            -ldflags "-X main.Version=$VERSION -s -w" \
            "$ENTRY_POINT"

    print_success "Built: $output"
}

case "$TARGET" in
    "all")
        build_target "linux" "amd64" ""
        build_target "linux" "arm64" ""
        build_target "windows" "amd64" ".exe"
        build_target "windows" "arm64" ".exe"
        build_target "darwin" "amd64" ""
        build_target "darwin" "arm64" ""
        ;;
    "linux")
        build_target "linux" "amd64" ""
        build_target "linux" "arm64" ""
        ;;
    "linux/amd64")
        build_target "linux" "amd64" ""
        ;;
    "linux/arm64")
        build_target "linux" "arm64" ""
        ;;
    "windows")
        build_target "windows" "amd64" ".exe"
        build_target "windows" "arm64" ".exe"
        ;;
    "windows/amd64")
        build_target "windows" "amd64" ".exe"
        ;;
    "windows/arm64")
        build_target "windows" "arm64" ".exe"
        ;;
    "darwin"|"macos")
        build_target "darwin" "amd64" ""
        build_target "darwin" "arm64" ""
        ;;
    "darwin/amd64"|"macos/amd64")
        build_target "darwin" "amd64" ""
        ;;
    "darwin/arm64"|"macos/arm64")
        build_target "darwin" "arm64" ""
        ;;
    *)
        print_error "Unknown target: $TARGET"
        echo "Supported targets: all, linux, linux/amd64, linux/arm64, windows, windows/amd64, windows/arm64, darwin/amd64, darwin/arm64"
        exit 1
        ;;
esac

echo ""
print_header "Build Complete!"
echo "Output directory: $DIST_DIR"
echo ""
print_info "Files built:"
ls -lh "$DIST_DIR" | tail -n +2 | awk '{printf "  %-40s %8s\n", $9, $5}'

echo ""
print_info "Usage examples:"
echo "  ./dist/nvrhikvision-exporter-linux-amd64 -config=config.yaml"
echo "  ./dist/nvrhikvision-exporter-windows-amd64.exe -config=config.yaml"
echo "  ./dist/nvrhikvision-exporter-darwin-arm64 -config=config.yaml"