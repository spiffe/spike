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

DEMO_PATH="$(pwd)/demo"
DEMO_SHA=$(sha256sum "$DEMO_PATH" | cut -d' ' -f1)

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

# Register Demo Workload
spire-server entry create \
    -spiffeID spiffe://spike.ist/workload/demo \
    -parentID "spiffe://spike.ist/spire-agent" \
    -selector unix:uid:"$(id -u)" \
    -selector unix:path:"$DEMO_PATH" \
    -selector unix:sha256:"$DEMO_SHA"

# DEBU[0140] PID attested to have selectors
# pid=10440 selectors="[type:\"unix\"  value:\"uid:1001\" type:\"unix\"  value:\"user:volkan\" type:\"unix\"
# value:\"gid:1001\" type:\"unix\"  value:\"group:volkan\" type:\"unix\"
# value:\"supplementary_gid:4\" type:\"unix\"  value:\"supplementary_group:adm\" type:\"unix\"
# value:\"supplementary_gid:24\" type:\"unix\"  value:\"supplementary_group:cdrom\" type:\"unix\"
# value:\"supplementary_gid:27\" type:\"unix\"  value:\"supplementary_group:sudo\" type:\"unix\"
# value:\"supplementary_gid:30\" type:\"unix\"  value:\"supplementary_group:dip\" type:\"unix\"
# value:\"supplementary_gid:46\" type:\"unix\"  value:\"supplementary_group:plugdev\" type:\"unix\"
# value:\"supplementary_gid:100\" type:\"unix\"  value:\"supplementary_group:users\" type:\"unix\"
# value:\"supplementary_gid:121\" type:\"unix\"  value:\"supplementary_group:lpadmin\" type:\"unix\"
# value:\"supplementary_gid:1001\" type:\"unix\"  value:\"supplementary_group:volkan\" type:\"unix\"
# value:\"path:/home/volkan/Desktop/WORKSPACE/spike/demo\" type:\"unix\"  value:\"sha256:111501b4abf0f8996621f64631548386f086f96a800fef995864f63fa9fd0362\"]" subsystem_name=workload_attestor

