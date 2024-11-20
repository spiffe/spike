#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

KEEPER_PATH="$(pwd)/keeper"
KEEPER_SHA=$(sha256sum "$KEEPER_PATH" | cut -d' ' -f1)

NEXUS_PATH="$(pwd)/nexus"
NEXUS_SHA=$(sha256sum "$NEXUS_PATH" | cut -d' ' -f1)

PILOT_PATH="$(pwd)/spike"
PILOT_SHA=$(sha256sum "$PILOT_PATH" | cut -d' ' -f1)

# ## A note for Mac OS users ##
#
# The SPIRE Unix Workload Attestor plugin generates selectors based on 
# Unix-specific attributes of workloads. 
#
# On Darwin (macOS), the following selectors are supported:
# * unix:uid: The user ID of the workload (e.g., unix:uid:1000).
# * unix:user: The username of the workload (e.g., unix:user:nginx).
# * unix:gid: The group ID of the workload (e.g., unix:gid:1000).
# * unix:group: The group name of the workload (e.g., unix:group:www-data).
#
# However, the following selectors are not supported on Darwin:
# * unix:supplementary_gid: The supplementary group ID of the workload.
# * unix:supplementary_group: The supplementary group name of the workload.
#
# ^ These selectors are currently only supported on Linux systems.
#
# Additionally, if the plugin is configured with discover_workload_path = true, 
# it can provide these selectors:
# * unix:path: The path to the workload binary (e.g., unix:path:/usr/bin/nginx).
# * unix:sha256: The SHA256 digest of the workload binary (e.g., unix:sha256:3a6...).

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



