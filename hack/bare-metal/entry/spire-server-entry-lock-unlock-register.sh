#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Set path to the SPIKE binary
PILOT_PATH="$(pwd)/spike"
PILOT_SHA=$(sha256sum "$PILOT_PATH" | cut -d' ' -f1)

# Register a new SPIFFE ID for lock/unlock
spire-server entry create \
  -spiffeID "spiffe://spike.ist/spike/pilot/role/watchdog" \
  -parentID "spiffe://spike.ist/spire-agent" \
  -selector "unix:uid:$(id -u)" \
  -selector "unix:path:$PILOT_PATH" \
  -selector "unix:sha256:$PILOT_SHA" \
  -ttl 3600

# Wait for propagation
echo "Waiting for SPIFFE entry to propagate..."
sleep 5
echo "Watchdog SPIFFE ID registered successfully!"
