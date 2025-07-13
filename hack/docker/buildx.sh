#!/bin/bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

set -e

# Check if both arguments are provided
if [ $# -ne 2 ]; then
  echo "Usage: $0 <arch> <app>"
  echo "  arch: amd64 or arm64"
  echo "  app: application name"
  exit 1
fi

TARGET_ARCH=$1
APP=$2

# Set common environment variables
export CGO_ENABLED=1
export GOARCH=$TARGET_ARCH

if [ "$TARGET_ARCH" = "amd64" ]; then
  export CC=x86_64-linux-gnu-gcc
  export CXX=x86_64-linux-gnu-g++
  export AR=x86_64-linux-gnu-ar
elif [ "$TARGET_ARCH" = "arm64" ]; then
  export CC=aarch64-linux-gnu-gcc
  export CXX=aarch64-linux-gnu-g++
  export AR=aarch64-linux-gnu-ar
else
  echo "Error: Supported architectures are amd64 and arm64"
  exit 1
fi

echo "Building $APP for $TARGET_ARCH with CGO_ENABLED=1"
echo "CC=$CC"

# Determine if this app needs CGO
NEEDS_CGO=false
case $APP in
  nexus)
    NEEDS_CGO=true
    ;;
  pilot|keeper)
    NEEDS_CGO=false
    ;;
esac

if [ "$NEEDS_CGO" = "true" ]; then
  export CGO_ENABLED=1
  # Set cross-compile toolchain...
  echo "Building $APP with CGO enabled (for SQLite)"
else
  export CGO_ENABLED=0
  echo "Building $APP as static binary (no CGO needed)"
fi

# Build with proper flags
go build -ldflags="-s -w" -o "$APP" /workspace/app/"$APP"/cmd/main.go

# Verify the binary was created
if [ -f "$APP" ]; then
  echo "Successfully built $APP for $TARGET_ARCH"
else
  echo "Error: Failed to build $APP"
  exit 1
fi
