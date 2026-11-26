#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")/.."

export VERSION=${VERSION:-$(sed -nE 's/^var Version = "([^"]+)"/\1/p' cmd/server/main.go)}
export BUILD_TIME=${BUILD_TIME:-$(date -u '+%Y-%m-%dT%H:%M:%SZ')}
export GIT_COMMIT=${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}

docker buildx create --name multiarch-builder --driver docker-container --use 2>/dev/null ||
    docker buildx use multiarch-builder

docker buildx bake -f docker-bake.hcl --push --set 'ipaddress-api-local.pull=true' ipaddress-api-local
