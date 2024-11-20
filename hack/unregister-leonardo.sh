#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Disable Leonardo.
spire-server entry delete \
    -spiffeID spiffe://spike.ist/spike/pilot/role/superadmin
