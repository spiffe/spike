#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.

./spike policy create --name=workload-can-read \
  --path="/tenants/demo/db/*" \
  --spiffeid="^spiffe://spike.ist/workload/*" \
  --permissions="read"

./spike policy create --name=workload-can-write \
  --path="/tenants/demo/db/*" \
  --spiffeid="^spiffe://spike.ist/workload/*" \
  --permissions="write"

#./spike policy create --name=workload-can-rw \
#  --path="/tenants/demo/db/*" \
#  --spiffeid="^spiffe://spike.ist/workload/*" \
#  --permissions="read,write"