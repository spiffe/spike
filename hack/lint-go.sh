#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# TODO(strideynet): In Go 1.24, we can leverage the native Go "tools" support
# rather than providing a specific version here.
go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.0 run