#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

rm keeper
rm nexus
rm spike

# Build for the current platform.
echo "Building SPIKE binaries..."
CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o keeper-darwin-arm64 ./app/keeper/cmd/main.go
CGO_ENABLED=1 GOEXPERIMENT=boringcrypto go build -o nexus-darwin-arm64 ./app/nexus/cmd/main.go
CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o spike-darwin-arm64 ./app/spike/cmd/main.go
echo "Built SPIKE binaries."
