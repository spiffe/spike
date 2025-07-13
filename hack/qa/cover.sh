#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

go test -race -shuffle=on -coverprofile=coverage.txt ./...

go tool cover -html=coverage.txt -o=coverage.html

mv coverage.txt ./docs
mv coverage.html ./docs