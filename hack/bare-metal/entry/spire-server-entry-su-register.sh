#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

PILOT_PATH="$(pwd)/bin/spike"
PILOT_SHA=$(sha256sum "$PILOT_PATH" | cut -d' ' -f1)

BOOTSTRAP_PATH="$(pwd)/bin/bootstrap"
BOOTSTRAP_SHA=$(sha256sum "$BOOTSTRAP_PATH" | cut -d' ' -f1)

PILOT_ENTRY_ID=$(spire-server entry show -spiffeID \
  spiffe://spike.ist/spike/pilot/role/superuser \
  | grep "Entry ID" | awk -F: '{print $2}' | xargs)
BOOTSTRAP_ENTRY_ID=$(spire-server entry show -spiffeID \
   spiffe://spike.ist/spike/bootstrap \
   | grep "Entry ID" | awk -F: '{print $2}' | xargs)

if [[ -n $PILOT_ENTRY_ID ]]; then
  echo "Warning: An entry with ID $PILOT_ENTRY_ID already exists."
  echo "Exiting to avoid duplicate registration."
  exit 1
fi
if [[ -n $BOOTSTRAP_ENTRY_ID ]]; then
  echo "Warning: An entry with ID $BOOTSTRAP_ENTRY_ID already exists."
  echo "Exiting to avoid duplicate registration."
  exit 1
fi

# Register SPIKE Pilot for the superuser role.
spire-server entry create \
    -spiffeID spiffe://spike.ist/spike/pilot/role/superuser \
    -parentID "spiffe://spike.ist/spire-agent" \
    -selector unix:uid:"$(id -u)" \
    -selector unix:path:"$PILOT_PATH" \
    -selector unix:sha256:"$PILOT_SHA"

# Register SPIKE Bootstrap for the bootstrap role.
spire-server entry create \
    -spiffeID spiffe://spike.ist/spike/bootstrap \
    -parentID "spiffe://spike.ist/spire-agent" \
    -selector unix:uid:"$(id -u)" \
    -selector unix:path:"$BOOTSTRAP_PATH" \
    -selector unix:sha256:"$BOOTSTRAP_SHA"

# Wait for the entry to be updated
echo "Waiting for entries to be updated..."
sleep 5
echo "Everything is awesome!"
