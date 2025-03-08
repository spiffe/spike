#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# TODO: this must start from 1 and monotonically increase and match keeper count -- add to documentation.
SPIKE_NEXUS_KEEPER_PEERS='{"1":"https://localhost:8443","2":"https://localhost:8543","3":"https://localhost:8643"}' \
./nexus
