#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

KEEPER_PATH="$(pwd)/bin/keeper"
KEEPER_SHA=$(sha256sum "$KEEPER_PATH" | cut -d' ' -f1)

NEXUS_PATH="$(pwd)/bin/nexus"
NEXUS_SHA=$(sha256sum "$NEXUS_PATH" | cut -d' ' -f1)

# Register SPIKE Keeper
spire-server entry create \
    -spiffeID spiffe://spike.ist/spike/keeper \
    -parentID "spiffe://spike.ist/spire-agent" \
    -selector unix:uid:"$(id -u)" \
    -selector unix:path:"$KEEPER_PATH" \
    -selector unix:sha256:"$KEEPER_SHA"

# Register SPIKE Nexus
spire-server entry create \
    -spiffeID spiffe://spike.ist/spike/nexus \
    -parentID "spiffe://spike.ist/spire-agent" \
    -selector unix:uid:"$(id -u)" \
    -selector unix:path:"$NEXUS_PATH" \
    -selector unix:sha256:"$NEXUS_SHA"

# Wait for the entry to be updated
echo "Waiting for entries to be updated..."
sleep 5
echo "Everything is awesome!"