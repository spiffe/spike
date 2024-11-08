#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

if [ -f .spike-admin-token ]; then
  rm .spike-admin-token
fi

if [ -f .spike-agent-join-token ]; then
  rm .spike-agent-join-token
fi

./hack/build-spike.sh
./hack/clear-data.sh
./hack/start-spire-server.sh
