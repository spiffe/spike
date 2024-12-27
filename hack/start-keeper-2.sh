#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

SPIKE_KEEPER_PEERS='{"1":"https://localhost:8443","2":"https://localhost:8543","3":"https://localhost:8643"}' \
SPIKE_KEEPER_RANDOM_SEED='SHARED_RANDOM_SEED--CHANGE_THIS_FOR_PRODUCTION_USE--7f1b8eU1BJS0UgUm9ja3MK22736' \
SPIKE_KEEPER_TLS_PORT=':8543' \
SPIKE_KEEPER_ID="2" \
./keeper
