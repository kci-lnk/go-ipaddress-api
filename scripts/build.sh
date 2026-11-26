#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")/.."

VERSION=${VERSION:-$(sed -nE 's/^var Version = "([^"]+)"/\1/p' cmd/server/main.go)}
BUILD_TIME=${BUILD_TIME:-$(date -u '+%Y-%m-%dT%H:%M:%SZ')}
GIT_COMMIT=${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}
OUTPUT_DIR=${OUTPUT_DIR:-bin}

if [[ -z "$VERSION" ]]; then
    echo "Unable to determine the build version." >&2
    exit 1
fi

echo "Building version: $VERSION"
echo "Build time: $BUILD_TIME"
echo "Git commit: $GIT_COMMIT"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Build for multiple architectures
build() {
    local os=$1
    local arch=$2
    local suffix=$3

    echo "Building ${os}/${arch}..."

    GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build \
        -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o "$OUTPUT_DIR/ipaddress-api-${os}-${arch}${suffix}" \
        ./cmd/server
}

# Linux builds
build linux amd64 ""
build linux arm64 ""

# macOS builds
build darwin amd64 ""
build darwin arm64 ""

# Windows build
build windows amd64 ".exe"

echo "Build complete. Binaries:"
ls -la "$OUTPUT_DIR"
