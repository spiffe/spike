#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# TODO: move all these demo-related scripts to a "demo" folder maybe.

kubectl -n spire-server port-forward svc/spiffe-server 8443:8443 --address=0.0.0.0
