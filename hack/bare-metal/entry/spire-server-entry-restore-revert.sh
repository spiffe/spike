#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

PILOT_PATH="$(pwd)/bin/spike"
PILOT_SHA=$(sha256sum "$PILOT_PATH" | cut -d' ' -f1)

ENTRY_ID=$(spire-server entry show -spiffeID \
  spiffe://spike.ist/spike/pilot/role/restore \
  | grep "Entry ID" | awk -F: '{print $2}' | xargs)

spire-server entry update \
  -entryID "$ENTRY_ID" \
  -spiffeID spiffe://spike.ist/spike/pilot/role/superuser \
  -parentID "spiffe://spike.ist/spire-agent" \
  -selector unix:uid:"$(id -u)" \
  -selector unix:path:"$PILOT_PATH" \
  -selector unix:sha256:"$PILOT_SHA"

# Wait for the entry to be updated
echo "Waiting for entries to be updated..."
sleep 5
echo "Everything is awesome!"