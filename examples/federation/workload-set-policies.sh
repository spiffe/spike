#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

kubectl exec -n spike deploy/spiffe-spike-pilot -- spike policy create --name=workload-can-read-2 \
 --path="tenants/demo/db/*" \
 --spiffeid="^spiffe://workload\\.spike\\.ist/workload/*" \
 --permissions="read"

kubectl exec -n spike deploy/spiffe-spike-pilot -- spike policy create --name=workload-can-write-2 \
 --path="tenants/demo/db/*" \
 --spiffeid="^spiffe://workload\\.spike\\.ist/workload/*" \
 --permissions="write"
