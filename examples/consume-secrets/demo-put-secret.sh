#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.

if ! command -v spike &> /dev/null; then
    echo "Error: 'spike' command not found. Please add ./spike to your PATH."
    exit 1
fi

spike secret put tenants/demo/db/creds \
  username=admin \
  password=SPIKERocks
