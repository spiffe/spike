#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

kubectl -n spike port-forward svc/spiffe-spike-nexus 8444:443 --address=0.0.0.0
