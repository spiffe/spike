#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Run tests with coverage
export CGO_ENABLED=0
go test -coverprofile=coverage.out ./...

# Generate HTML coverage report from the coverage data
go tool cover -html=coverage.out -o coverage.html