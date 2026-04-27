#!/bin/bash
set -e

cd "$(dirname "$0")/.."

# Load version
VERSION=$(grep -E 'Version\s*=' cmd/server/version.go | head -1 | sed 's/.*=\s*"\([^"]*\)".*/\1/')
export VERSION

echo "Building Docker images for version: $VERSION"

BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
export BUILD_TIME

GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
export GIT_COMMIT

# Build for multiple platforms
docker buildx create --use || true

docker buildx build \
    --pull \
    --platform linux/amd64,linux/arm64 \
    -t ipaddress-api:latest \
    -t ipaddress-api:${VERSION} \
    --build-arg VERSION=${VERSION} \
    --build-arg BUILD_TIME=${BUILD_TIME} \
    --build-arg GIT_COMMIT=${GIT_COMMIT} \
    --push \
    .

echo "Docker images pushed successfully"
