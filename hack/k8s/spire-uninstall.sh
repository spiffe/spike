#!/bin/bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Enable alias expansion in non-interactive shells
shopt -s expand_aliases

source ./hack/lib/aliases.sh

helm uninstall spire -n spire-system

kubectl delete pvc -n spire-system -l app.kubernetes.io/instance=spire

kubectl delete clusterspiffeids --all

echo "SPIRE uninstalled."