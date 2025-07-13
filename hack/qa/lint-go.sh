#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Requires a recent version of go runtime.
# Requires `sudo apt-get install binutils-gold gcc`
go run github.com/golangci/golangci-lint/cmd/golangci-lint run -v
