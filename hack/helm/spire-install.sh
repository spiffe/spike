#!/bin/bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

set -e  # Exit on any error

# Enable alias expansion in non-interactive shells
shopt -s expand_aliases

source ./hack/lib/aliases.sh

# Add Helm repository if it doesn't exist
if ! helm repo list | grep -q "^spiffe\s"; then
    echo "Adding spiffe Helm repository..."
    helm repo add spiffe https://spiffe.github.io/helm-charts/
else
    echo "spiffe Helm repository already exists."
fi

# Update Helm repositories
echo "Updating Helm repositories..."
helm repo update

# Create namespace if it doesn't exist
if ! kubectl get namespace spire-system &> /dev/null; then
    echo "Creating spire-system namespace..."
    kubectl create namespace spire-system
else
    echo "spire-system namespace already exists."
fi

# Check if SPIRE is already installed
if helm list -n spire-system | grep -q "^spire\s"; then
    echo "SPIRE is already installed. Upgrading with the latest values..."
    helm upgrade spire spiffe/spire --namespace spire-system \
         -f ./config/spire/helm/values.yaml
else
    echo "Installing SPIRE..."
    helm install spire spiffe/spire --namespace spire-system \
         -f ./config/spire/helm/values.yaml
fi

echo "SPIRE setup completed successfully."
