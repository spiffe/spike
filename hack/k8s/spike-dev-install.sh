#!/bin/bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Installs SPIRE and SPIKE to the cluster.
# Uses the local container registry for SPIKE images.

set -e  # Exit on any error

# Add Helm repository if it doesn't exist
if ! helm repo list | grep -q "^spiffe\s"; then
  echo "Adding SPIFFE Helm repository..."
  helm repo add spiffe https://spiffe.github.io/helm-charts-hardened/
else
  echo "SPIFFE Helm repository already exists."
fi

# Note that this is NOT a SPIRE production setup.
# Consult SPIRE documentation for production deployment and hardening:
# https://spiffe.io/docs/latest/spire-helm-charts-hardened-about/recommendations/

helm repo update

echo "SPIKE install..."
echo "Current context: $(kubectl config current-context)"

create_namespace_if_not_exists() {
  local ns=$1
  echo "Checking namespace '$ns'..."

  # More explicit check
  if kubectl get namespace "$ns" 2>/dev/null | grep -q "$ns"; then
    echo "Namespace '$ns' already exists, skipping..."
  else
    echo "Creating namespace '$ns'..."
    kubectl create namespace "$ns"
    # shellcheck disable=SC2181
    if [ $? -eq 0 ]; then
      echo "Successfully created namespace '$ns'"
    else
      echo "Failed to create namespace '$ns'"
      return 1
    fi
  fi
}

create_namespace_if_not_exists "spike" # Pilot/Nexus/Keepers/Bootstrap

# List all namespaces after creation
echo "SPIKE namespaces:"
kubectl get namespaces | grep spike || echo "No spike namespaces found"

helm upgrade --install -n spire-mgmt spire-crds spire-crds \
  --repo https://spiffe.github.io/helm-charts-hardened/ \
  --version 0.5.0 --create-namespace

echo "Sleeping for 15 secs before installing SPIRE and SPIKE..."
sleep 15

helm upgrade --install -n spire-mgmt spiffe spire \
  --repo https://spiffe.github.io/helm-charts-hardened/ \
  --version 0.26.1 \
  -f ./config/helm/values-dev.yaml

echo "Sleeping for 15 secs..."
sleep 15

echo "Everything is awesome!"
