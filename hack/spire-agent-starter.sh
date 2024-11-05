#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

./hack/generate-agent-token.sh
./hack/register-spire-entries.sh
./hack/start-spire-agent.sh

