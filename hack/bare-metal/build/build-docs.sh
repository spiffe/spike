#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Check if zola exists
if ! command -v zola &> /dev/null; then
	echo "Error: zola is not installed or not in PATH"
	exit 1
fi

cd ./docs-src || exit

zola build

# Move built content from ./docs-src/public to ./docs, merging content
cp -r ./public/* ../docs/