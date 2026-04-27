#!/bin/bash
set -e

cd "$(dirname "$0")/.."

# Load version
VERSION=$(grep -E 'Version\s*=' cmd/server/version.go | head -1 | sed 's/.*=\s*"\([^"]*\)".*/\1/')
export VERSION

if [ -z "$SSH_HOST" ]; then
    echo "Error: SSH_HOST not set"
    exit 1
fi

if [ ! -d "bin" ]; then
    echo "Error: bin directory not found. Run 'task build-multi' first"
    exit 1
fi

echo "Publishing binaries to $SSH_HOST..."

# Sync binaries
rsync -avz --progress \
    --chmod=755 \
    bin/ \
    root@${SSH_HOST}:/opt/ipaddress-api/bin/

echo "Binaries published successfully"
