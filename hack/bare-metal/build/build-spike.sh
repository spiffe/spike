#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

rm keeper
rm nexus
rm spike
rm demo
rm bootstrap

# Build for the current platform.
echo "Building SPIKE binaries..."
GOFIPS140=v1.0.0 go build -ldflags="-s -w" \
  -o keeper ./app/keeper/cmd/main.go
# CGO is for sqlite dependency.
CGO_ENABLED=1 GOFIPS140=v1.0.0 go build -ldflags="-s -w" \
  -o nexus ./app/nexus/cmd/main.go
GOFIPS140=v1.0.0 go build -ldflags="-s -w" \
  -o spike ./app/spike/cmd/main.go
GOFIPS140=v1.0.0 go build -ldflags="-s -w" \
  -o demo ./app/demo/cmd/main.go
GOFIPS140=v1.0.0 go build -ldflags="-s -w" \
  -o bootstrap ./app/bootstrap/cmd/main.go

echo "!!! The symbols have been stripped from binaries for security."
echo "!!! Use 'readelf --symbols #binary_name#' to verify."

echo ""
echo ""

echo "Built SPIKE binaries."
