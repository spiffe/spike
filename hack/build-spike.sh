#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

rm keeper
rm nexus
rm spike

# Build for the current platform.
echo "Building SPIKE binaries..."
CGO_ENABLED=1 GOEXPERIMENT=boringcrypto go build -ldflags="-s -w" \
  -o keeper ./app/keeper/cmd/main.go
CGO_ENABLED=1 GOEXPERIMENT=boringcrypto go build -ldflags="-s -w" \
  -o nexus ./app/nexus/cmd/main.go
CGO_ENABLED=1 GOEXPERIMENT=boringcrypto go build -ldflags="-s -w" \
  -o spike ./app/spike/cmd/main.go
CGO_ENABLED=1 GOEXPERIMENT=boringcrypto go build -ldflags="-s -w" \
  -o demo ./app/demo/cmd/main.go
echo "Built SPIKE binaries."
