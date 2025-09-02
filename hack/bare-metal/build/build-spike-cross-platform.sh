#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

VERSION=$(cat ./app/VERSION.txt)

echo "Building for Mac ARM64..."
GOOS=darwin GOARCH=arm64 GOFIPS140=v1.0.0 \
  go build -ldflags="-s -w" -o keeper-v"$VERSION"-darwin-arm64 \
  ./app/keeper/cmd/main.go
GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 GOFIPS140=v1.0.0 \
  go build -ldflags="-s -w" -o nexus-v"$VERSION"-darwin-arm64 \
  ./app/nexus/cmd/main.go
GOOS=darwin GOARCH=arm64 GOFIPS140=v1.0.0 \
  go build -ldflags="-s -w" -o spike-v"$VERSION"-darwin-arm64 \
  ./app/spike/cmd/main.go
GOOS=darwin GOARCH=arm64 GOFIPS140=v1.0.0 \
  go build -ldflags="-s -w" -o bootstrap-v"$VERSION"-darwin-arm64 \
  ./app/bootstrap/cmd/main.go
echo "Built for Mac ARM64."

# Build for Linux ARM64
echo "Building for Linux ARM64..."
CC=aarch64-linux-musl-gcc \
  GOOS=linux GOARCH=arm64 GOFIPS140=v1.0.0 \
  go build -ldflags="-s -w" -o keeper-v"$VERSION"-linux-arm64 \
  ./app/keeper/cmd/main.go
CC=aarch64-linux-musl-gcc \
  GOOS=linux GOARCH=arm64 CGO_ENABLED=1 GOFIPS140=v1.0.0 \
  go build -ldflags="-s -w" -o nexus-v"$VERSION"-linux-arm64 \
  ./app/nexus/cmd/main.go
CC=aarch64-linux-musl-gcc \
  GOOS=linux GOARCH=arm64 GOFIPS140=v1.0.0 \
  go build -ldflags="-s -w" -o spike-v"$VERSION"-linux-arm64 \
  ./app/spike/cmd/main.go
CC=aarch64-linux-musl-gcc \
  GOOS=linux GOARCH=arm64 GOFIPS140=v1.0.0 \
  go build -ldflags="-s -w" -o bootstrap-v"$VERSION"-linux-arm64 \
  ./app/bootstrap/cmd/main.go
echo "Built for Linux ARM64."

# Build for Linux AMD64
echo "Building for Linux AMD64..."
CC=x86_64-linux-musl-gcc \
  GOOS=linux GOARCH=amd64 GOFIPS140=v1.0.0 \
  go build -ldflags="-s -w" -o keeper-v"$VERSION"-linux-amd64 \
  ./app/keeper/cmd/main.go
CC=x86_64-linux-musl-gcc \
  GOOS=linux GOARCH=amd64 CGO_ENABLED=1 GOFIPS140=v1.0.0 \
  go build -ldflags="-s -w" -o nexus-v"$VERSION"-linux-amd64 \
  ./app/nexus/cmd/main.go
CC=x86_64-linux-musl-gcc \
   GOOS=linux GOARCH=amd64 GOFIPS140=v1.0.0 \
   go build -ldflags="-s -w" -o spike-v"$VERSION"-linux-amd64 \
   ./app/spike/cmd/main.go
CC=x86_64-linux-musl-gcc \
   GOOS=linux GOARCH=amd64 GOFIPS140=v1.0.0 \
   go build -ldflags="-s -w" -o bootstrap-v"$VERSION"-linux-amd64 \
   ./app/bootstrap/cmd/main.go
echo "Built for Linux AMD64."

echo "Computing SHA Sums"

for file in keeper-* nexus-* spike-* bootstrap-*; do
  shasum -a 256 "$file" > "$file.sum.txt"
done

echo "!!! The symbols have been stripped from binaries for security."
echo "!!! Use 'readelf --symbols #binary_name#' to verify."

echo ""
echo ""

echo "Everything is awesome!"
