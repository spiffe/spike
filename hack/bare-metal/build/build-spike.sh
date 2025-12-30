#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

rm ./bin/keeper
rm ./bin/nexus
rm ./bin/spike
rm ./bin/demo
rm ./bin/bootstrap

# Build for the current platform.
echo "Building SPIKE binaries..."
GOFIPS140=v1.0.0 go build -trimpath -ldflags="-s -w" \
  -o ./bin/keeper ./app/keeper/cmd/main.go
# CGO is for sqlite dependency.
CGO_ENABLED=1 GOFIPS140=v1.0.0 go build -trimpath -ldflags="-s -w" \
  -o ./bin/nexus ./app/nexus/cmd/main.go
GOFIPS140=v1.0.0 go build -trimpath -ldflags="-s -w" \
  -o ./bin/spike ./app/spike/cmd/main.go
GOFIPS140=v1.0.0 go build -trimpath -ldflags="-s -w" \
  -o ./bin/demo ./app/demo/cmd/main.go
GOFIPS140=v1.0.0 go build -trimpath -ldflags="-s -w" \
  -o ./bin/bootstrap ./app/bootstrap/cmd/main.go

echo "!!! The symbols have been stripped from binaries for security."
echo "!!! Use 'readelf --symbols #binary_name#' to verify."

echo ""
echo ""

echo "Built SPIKE binaries."
