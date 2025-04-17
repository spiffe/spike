#!/bin/bash
set -e

# Check if both arguments are provided
if [ $# -ne 2 ]; then
    echo "Usage: $0 <arch> <app>"
    echo "  arch: amd64 or arm64"
    echo "  app: application name"
    exit 1
fi

TARGETARCH=$1
APP=$2

if [ "$TARGETARCH" = "amd64" ]; then
    CC=x86_64-linux-gnu-gcc go build -ldflags="-s -w" -o $APP /workspace/app/$APP/cmd/main.go
elif [ "$TARGETARCH" = "arm64" ]; then
    CC=aarch64-linux-gnu-gcc go build -ldflags="-s -w" -o $APP /workspace/app/$APP/cmd/main.go
else
    echo "Error: Supported architectures are amd64 and arm64"
    exit 1
fi