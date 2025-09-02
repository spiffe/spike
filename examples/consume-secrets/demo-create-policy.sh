#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.

if ! command -v spike &> /dev/null; then
  echo "Error: 'spike' command not found. Please add ./spike to your PATH."
  exit 1
fi

spike policy create --name=workload-can-read \
 --path-pattern="^tenants/demo/db/.*$" \
 --spiffeid-pattern="^spiffe://spike\.ist/workload/.*$" \
 --permissions="read"

spike policy create --name=workload-can-write \
 --path-pattern="^tenants/demo/db/.*$" \
 --spiffeid-pattern="^spiffe://spike\.ist/workload/.*$" \
 --permissions="write"

# spike policy create --name=workload-can-rw \
#  --path-pattern="^tenants/demo/db/.*$" \
#  --spiffeid-pattern="^spiffe://spike\.ist/workload/.*$" \
#  --permissions="read,write"
