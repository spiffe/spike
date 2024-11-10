#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

rm keeper
rm nexus
rm spike

# `boringcrypto` is required for FIPS compliance.

# SPIKE Nexus uses `CGO_ENABLED=1` for sqlite3. If another backend is used,
# set `CGO_ENABLED=0`.

# Linux ARM64
CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o keeper ./app/keeper/cmd/main.go
CGO_ENABLED=1 GOEXPERIMENT=boringcrypto go build -o nexus ./app/nexus/cmd/main.go
CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o spike ./app/spike/cmd/main.go

#commenting out until cross-compilation is figured out.
#GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o keeper-linux-amd64 ./app/keeper/cmd/main.go
#GOOS=linux GOARCH=amd64 CGO_ENABLED=1 GOEXPERIMENT=boringcrypto go build -o nexus-linux-amd64 ./app/nexus/cmd/main.go
#GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o spike-linux-amd64 ./app/spike/cmd/main.go
#
#GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o keeper-darwin-arm64 ./app/keeper/cmd/main.go
#GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 GOEXPERIMENT=boringcrypto go build -o nexus-darwin-arm64 ./app/nexus/cmd/main.go
#GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o spike-darwin-arm64 ./app/spike/cmd/main.go
#
#GOOS=linux GOARCH=arm64 CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o keeper-linux-arm64 ./app/keeper/cmd/main.go
#GOOS=linux GOARCH=arm64 CGO_ENABLED=1 GOEXPERIMENT=boringcrypto go build -o nexus-linux-arm64 ./app/nexus/cmd/main.go
#GOOS=linux GOARCH=arm64 CGO_ENABLED=0 GOEXPERIMENT=boringcrypto go build -o spike-linux-arm64 ./app/spike/cmd/main.go