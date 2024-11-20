#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

PILOT_PATH="$(pwd)/spike"
PILOT_SHA=$(sha256sum "$PILOT_PATH" | cut -d' ' -f1)

sudo cp "$PILOT_PATH" /usr/local/bin/spike

PILOT_PATH="$(pwd)/spike"
PILOT_SHA=$(sha256sum "$PILOT_PATH" | cut -d' ' -f1)

# Allow Leonardo to use the binary
spire-server entry create \
    -spiffeID spiffe://spike.ist/spike/pilot/role/superuser \
    -parentID "spiffe://spike.ist/spire-agent" \
    -selector unix:uid:"$(id -u leonardo)" \
    -selector unix:path:"/usr/local/bin/spike" \
    -selector unix:sha256:"$PILOT_SHA"
