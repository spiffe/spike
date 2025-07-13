#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

source ./hack/lib/bg.sh

run_background "./hack/bare-metal/startup/start.sh"
sleep 5

echo "Starting tests..."
go build -ldflags="-s -w" -o ci-test ./ci/test/main.go
./ci-test

echo "Tests completed successfully!"
echo "Cleaning up background processes..."

cleanup
