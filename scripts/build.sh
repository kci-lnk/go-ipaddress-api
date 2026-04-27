#!/bin/bash
set -e

cd "$(dirname "$0")/.."

# Load version from cmd/server/version.go
VERSION=$(grep -E 'Version\s*=' cmd/server/version.go | head -1 | sed 's/.*=\s*"\([^"]*\)".*/\1/')
BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

export VERSION
export BUILD_TIME
export GIT_COMMIT

echo "Building version: $VERSION"
echo "Build time: $BUILD_TIME"
echo "Git commit: $GIT_COMMIT"

# Create output directory
mkdir -p bin

# Build for multiple architectures
build() {
    local os=$1
    local arch=$2
    local suffix=$3

    echo "Building ${os}/${arch}..."

    GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build \
        -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o "bin/ipaddress-api-${os}-${arch}${suffix}" \
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
ls -la bin/
