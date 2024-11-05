#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

rm .spike-token
rm .spike-agent-join-token

./hack/build-spike.sh
./hack/clear-data.sh
./hack/start-spire-server.sh
