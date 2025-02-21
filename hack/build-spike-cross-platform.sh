#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

VERSION="v0.2.2"

# Create a temporary file
TMP_FILE=$(mktemp)

# Backup original file
cp "$CONFIG_FILE" "${CONFIG_FILE}.bak"

# Update version constants in the config file
sed -E "s/(const [[:alnum:]]+Version = )\"[^\"]+\"/\1\"$VERSION\"/" \
  "$CONFIG_FILE" > "$TMP_FILE"

# Check if any changes were made
if ! diff -q "$CONFIG_FILE" "$TMP_FILE" >/dev/null; then
    mv "$TMP_FILE" "$CONFIG_FILE"
    echo "Updated versions in $CONFIG_FILE to $GO_VERSION"
    echo "Backup saved as ${CONFIG_FILE}.bak"
else
    rm "$TMP_FILE"
    rm "${CONFIG_FILE}.bak"
    echo "No version changes needed in $CONFIG_FILE"
fi

echo "Building for Mac ARM64..."
GOOS=darwin GOARCH=arm64 \
  CGO_ENABLED=0 GOEXPERIMENT=boringcrypto \
  go build -o keeper-v$VERSION-darwin-arm64 \
  ./app/keeper/cmd/main.go
GOOS=darwin GOARCH=arm64 \
  CGO_ENABLED=1 GOEXPERIMENT=boringcrypto \
  go build -o nexus-v$VERSION-darwin-arm64 \
  ./app/nexus/cmd/main.go
GOOS=darwin GOARCH=arm64 \
  CGO_ENABLED=0 GOEXPERIMENT=boringcrypto \
  go build -o spike-v$VERSION-darwin-arm64 \
  ./app/spike/cmd/main.go
echo "Built for Mac ARM64."

# Build for Linux ARM64
echo "Building for Linux ARM64..."
CC=aarch64-linux-musl-gcc \
  GOOS=linux GOARCH=arm64 \
  CGO_ENABLED=0 GOEXPERIMENT=boringcrypto \
  go build -o keeper-v$VERSION-linux-arm64 \
  ./app/keeper/cmd/main.go
CC=aarch64-linux-musl-gcc \
  GOOS=linux GOARCH=arm64 \
  CGO_ENABLED=1 GOEXPERIMENT=boringcrypto \
  go build -o nexus-v$VERSION-linux-arm64 \
  ./app/nexus/cmd/main.go
CC=aarch64-linux-musl-gcc \
  GOOS=linux GOARCH=arm64 CGO_ENABLED=0 GOEXPERIMENT=boringcrypto \
  go build -o spike-v$VERSION-linux-arm64 \
  ./app/spike/cmd/main.go
echo "Built for Linux ARM64."

# Build for Linux AMD64
echo "Building for Linux AMD64..."
CC=x86_64-linux-musl-gcc \
  GOOS=linux GOARCH=amd64 \
  CGO_ENABLED=0 GOEXPERIMENT=boringcrypto \
  go build -o keeper-v$VERSION-linux-amd64 \
  ./app/keeper/cmd/main.go
CC=x86_64-linux-musl-gcc \
  GOOS=linux GOARCH=amd64 \
  CGO_ENABLED=1 GOEXPERIMENT=boringcrypto \
  go build -o nexus-v$VERSION-linux-amd64 \
  ./app/nexus/cmd/main.go
CC=x86_64-linux-musl-gcc \
   GOOS=linux GOARCH=amd64 \
   CGO_ENABLED=0 GOEXPERIMENT=boringcrypto \
   go build -o spike-v$VERSION-linux-amd64 \
   ./app/spike/cmd/main.go
echo "Built for Linux AMD64."

echo "Computing SHA Sums"

for file in keeper-* nexus-* spike-*; do
    shasum -a 256 "$file" > "$file.sum.txt"
done

echo "Everything is awesome!"
