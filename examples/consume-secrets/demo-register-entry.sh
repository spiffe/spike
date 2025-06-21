#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.

DEMO_PATH="$(pwd)/demo"
DEMO_SHA=$(sha256sum "$DEMO_PATH" | cut -d' ' -f1)

# Register Demo Workload
spire-server entry create \
    -spiffeID spiffe://spike.ist/workload/demo \
    -parentID "spiffe://spike.ist/spire-agent" \
    -selector unix:uid:"$(id -u)" \
    -selector unix:path:"$DEMO_PATH" \
    -selector unix:sha256:"$DEMO_SHA"
