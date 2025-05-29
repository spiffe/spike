#!/bin/bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

set -e  # Exit on any error

# Add Helm repository if it doesn't exist
if ! helm repo list | grep -q "^spiffe\s"; then
    echo "Adding SPIFFE Helm repository..."
    helm repo add spiffe https://spiffe.github.io/helm-charts-hardened/
else
    echo "SPIFFE Helm repository already exists."
fi

helm repo update

# TODO: update manifests to use not root (security context)

# Note that this is NOT a SPIRE production setup.
# Consult SPIRE documentation for production deployment and hardening:
# https://spiffe.io/docs/latest/spire-helm-charts-hardened-about/recommendations/

helm repo update

helm upgrade --install -n spire-mgmt spire-crds spire-crds \
  --repo https://spiffe.github.io/helm-charts-hardened/ --create-namespace

echo "Sleeping for 15 secs..."
sleep 15

helm upgrade --install -n spire-mgmt spire spire \
  --repo https://spiffe.github.io/helm-charts-hardened/ \
  -f ./config/spire/helm/values.yaml

echo "Sleeping for 15 secs..."
sleep 15

echo "Everything is awesome!"
