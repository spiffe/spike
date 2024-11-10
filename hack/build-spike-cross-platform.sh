#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

echo "Building for Mac ARM64..."
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o keeper-darwin-arm64 ./app/keeper/cmd/main.go
GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 GOEXPERIMENT=boringcrypto go build -o nexus-darwin-arm64 ./app/nexus/cmd/main.go
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o spike-darwin-arm64 ./app/spike/cmd/main.go
echo "Built for Mac ARM64."

# Build for Linux ARM64
echo "Building for Linux ARM64..."
CC=aarch64-linux-musl-gcc \
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o keeper-linux-arm64 ./app/keeper/cmd/main.go
CC=aarch64-linux-musl-gcc \
GOOS=linux GOARCH=arm64 CGO_ENABLED=1 GOEXPERIMENT=boringcrypto go build -o nexus-linux-arm64 ./app/nexus/cmd/main.go
CC=aarch64-linux-musl-gcc \
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o spike-linux-arm64 ./app/spike/cmd/main.go
echo "Built for Linux ARM64."

# Build for Linux AMD64
echo "Building for Linux AMD64..."
CC=x86_64-linux-musl-gcc \
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o keeper-linux-amd64 ./app/keeper/cmd/main.go
CC=x86_64-linux-musl-gcc \
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 GOEXPERIMENT=boringcrypto go build -o nexus-linux-amd64 ./app/nexus/cmd/main.go
CC=x86_64-linux-musl-gcc \
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o spike-linux-amd64 ./app/spike/cmd/main.go
echo "Built for Linux AMD64."

echo "Computing SHA Sums"

for file in keeper-* nexus-* spike-*; do
    shasum -a 256 "$file" > "$file.sum.txt"
done

echo "Everything is awesome!"
