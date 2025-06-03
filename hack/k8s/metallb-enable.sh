#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

set -e  # Exit on any error

# Enable alias expansion in non-interactive shells
shopt -s expand_aliases

source ./hack/lib/aliases.sh
source ./hack/lib/env-k8s.sh

microk8s enable metallb:10.211.55.200-10.211.55.210
