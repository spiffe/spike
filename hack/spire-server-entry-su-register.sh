#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

PILOT_PATH="$(pwd)/spike"
PILOT_SHA=$(sha256sum "$PILOT_PATH" | cut -d' ' -f1)

# Register SPIKE Pilot for the super user
spire-server entry create \
    -spiffeID spiffe://spike.ist/spike/pilot/role/superuser \
    -parentID "spiffe://spike.ist/spire-agent" \
    -selector unix:uid:"$(id -u)" \
    -selector unix:path:"$PILOT_PATH" \
    -selector unix:sha256:"$PILOT_SHA"
