#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# FIXME: Remove -p 1 flag once issue with concurrent test isolation is resolved
go test -race -shuffle=on -coverprofile=coverage.txt -p 1 ./...

go tool cover -html=coverage.txt -o=coverage.html

mv coverage.txt ./docs
mv coverage.html ./docs