#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# The SPIKE Keeper peer address mappings MUST start with the key "1" and they MUST
# increment by 1 for each subsequent SPIKE Keeper.
SPIKE_NEXUS_KEEPER_PEERS='https://localhost:8443,https://localhost:8543,https://localhost:8643' \
exec ./nexus
