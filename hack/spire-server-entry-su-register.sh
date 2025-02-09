#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

PILOT_PATH="$(pwd)/spike"
PILOT_SHA=$(sha256sum "$PILOT_PATH" | cut -d' ' -f1)

ENTRY_ID=$(spire-server entry show -spiffeID \
  spiffe://spike.ist/spike/pilot/role/superuser \
  | grep "Entry ID" | awk -F: '{print $2}' | xargs)
  

if [[ -n $ENTRY_ID ]]; then
  echo "Warning: An entry with ID $ENTRY_ID already exists."
  echo "Exiting to avoid duplicate registration."
  exit 1
fi

ENTRY_ID=$(spire-server entry show -spiffeID \
  spiffe://spike.ist/spike/pilot/role/restore \
  | grep "Entry ID" | awk -F: '{print $2}' | xargs)
if [[ -n $ENTRY_ID ]]; then
  echo "Updating existing entry with ID $ENTRY_ID to superuser SPIFFE ID."
  spire-server entry update \
      -entryID "$ENTRY_ID" \
      -spiffeID spiffe://spike.ist/spike/pilot/role/superuser \
      -parentID "spiffe://spike.ist/spire-agent" \
      -selector unix:uid:"$(id -u)" \
      -selector unix:path:"$PILOT_PATH" \
      -selector unix:sha256:"$PILOT_SHA"

  exit 0
fi

# TODO: don't forget to document these special SPIFFEIDs.

ENTRY_ID=$(spire-server entry show -spiffeID \
  spiffe://spike.ist/spike/pilot/role/recover \
  | grep "Entry ID" | awk -F: '{print $2}' | xargs)
if [[ -n $ENTRY_ID ]]; then
  echo "Updating existing entry with ID $ENTRY_ID to superuser SPIFFE ID."
  spire-server entry update \
      -entryID "$ENTRY_ID" \
      -spiffeID spiffe://spike.ist/spike/pilot/role/superuser \
      -parentID "spiffe://spike.ist/spire-agent" \
      -selector unix:uid:"$(id -u)" \
      -selector unix:path:"$PILOT_PATH" \
      -selector unix:sha256:"$PILOT_SHA"

  exit 0
fi

# Register SPIKE Pilot for the super user
spire-server entry create \
    -spiffeID spiffe://spike.ist/spike/pilot/role/superuser \
    -parentID "spiffe://spike.ist/spire-agent" \
    -selector unix:uid:"$(id -u)" \
    -selector unix:path:"$PILOT_PATH" \
    -selector unix:sha256:"$PILOT_SHA"


# TODO:
#{"time":"2025-02-08T15:36:23.839639418-08:00","level":"WARN","msg":"RecoverBackingStoreUsingKeeperShards","msg":"Recovery unsuccessful. Will retry."}
#{"time":"2025-02-08T15:36:23.839643001-08:00","level":"WARN","msg":"RecoverBackingStoreUsingKeeperShards","msg":"Successful keepers: ","!BADKEY":{}}
#{"time":"2025-02-08T15:36:23.839687751-08:00","level":"WARN","msg":"RecoverBackingStoreUsingKeeperShards","msg":"You may need to manually bootstrap."}
#{"time":"2025-02-08T15:36:23.839693709-08:00","level":"INFO","msg":"RecoverBackingStoreUsingKeeperShards","msg":"Waiting for keepers to respond"}
