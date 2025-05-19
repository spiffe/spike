#!/bin/bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Enable alias expansion in non-interactive shells
shopt -s expand_aliases

source ./hack/aliases.sh

helm repo add spiffe https://spiffe.github.io/helm-charts/
helm repo update

# Create a namespace for SPIRE
kubectl create namespace spire-system

# Install SPIRE using Helm
helm install spire spiffe/spire --namespace spire-system \
 -f ./config/spire/helm/values.yaml

