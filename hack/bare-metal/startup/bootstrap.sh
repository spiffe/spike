#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Set log level to debug to see progress.
# Bootstrap needs to know which SPIKE Keeper endpoints to send shards to.
SPIKE_SYSTEM_LOG_LEVEL="debug" \
SPIKE_NEXUS_KEEPER_PEERS='https://localhost:8443,https://localhost:8543,https://localhost:8643' \
exec ./bootstrap -init
