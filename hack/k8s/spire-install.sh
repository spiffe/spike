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
    helm repo add spiffe https://spiffe.github.io/helm-charts-hardened/
else
    echo "spiffe Helm repository already exists."
fi

# Note that this is NOT a SPIRE production set up.
# Consult SPIRE documentation for production deployment and hardening.

# Better to create namespace manually instead of helm doing the guesswork.
kubectl create ns spire-system
kubectl create ns spire-server

helm repo update
helm upgrade --install -n spire-server spire-crds spire-crds --repo https://spiffe.github.io/helm-charts-hardened/ --create-namespace
helm upgrade --install -n spire-server spire spire --repo https://spiffe.github.io/helm-charts-hardened/ -f ./config/spire/helm/values.yaml



#

#
## Update Helm repositories
#echo "Updating Helm repositories..."
#helm repo update
#
## Install/Upgrade CRDs (let Helm create the namespace)
#echo "Installing/Upgrading SPIRE CRDs..."
#helm upgrade --install -n spire-system spire-crds spire-crds \
#  --repo https://spiffe.github.io/helm-charts-hardened/ \
#  --create-namespace
#
## Wait for CRDs to be established
#echo "Waiting for CRDs to be ready..."
#kubectl wait --for condition=established --timeout=60s \
#  crd/clusterspiffeids.spiffe.io \
#  crd/clusterfederatedtrustdomains.spiffe.io \
#  crd/controllermanagerconfigs.spire.spiffe.io 2>/dev/null || true
#
## Force update namespace metadata for the main spire release
#echo "Updating namespace metadata..."
#kubectl patch namespace spire-system --type=merge -p '{
#  "metadata": {
#    "annotations": {
#      "meta.helm.sh/release-name": "spire",
#      "meta.helm.sh/release-namespace": "spire-system"
#    },
#    "labels": {
#      "app.kubernetes.io/managed-by": "Helm"
#    }
#  }
#}'
#
## Verify the metadata was added
#echo "Verifying namespace metadata..."
#kubectl get namespace spire-system -o jsonpath='{.metadata}' | jq '.annotations, .labels'
#
## Install/Upgrade SPIRE
#echo "Installing/Upgrading SPIRE..."
#helm upgrade --install -n spire-system spire spire \
#  --repo https://spiffe.github.io/helm-charts-hardened/ \
#  -f ./config/spire/helm/values.yaml \
#  --set controllerManager.validatingWebhookConfiguration.failurePolicy=Ignore \
#  --wait
#
#echo "SPIRE setup completed successfully."


#set -e  # Exit on any error
#
## Enable alias expansion in non-interactive shells
#shopt -s expand_aliases
#
#source ./hack/lib/aliases.sh
#
## Add Helm repository if it doesn't exist
#if ! helm repo list | grep -q "^spiffe\s"; then
#    echo "Adding spiffe Helm repository..."
#    helm repo add spiffe https://spiffe.github.io/helm-charts/
#else
#    echo "spiffe Helm repository already exists."
#fi
#
## Update Helm repositories
#echo "Updating Helm repositories..."
#helm repo update
#
## Install/Upgrade CRDs (let Helm create the namespace)
#echo "Installing/Upgrading SPIRE CRDs..."
#helm upgrade --install -n spire-system spire-crds spire-crds \
#  --repo https://spiffe.github.io/helm-charts-hardened/ \
#  --create-namespace
#
## Wait for CRDs to be established
#echo "Waiting for CRDs to be ready..."
#kubectl wait --for condition=established --timeout=60s \
#  crd/clusterspiffeids.spiffe.io \
#  crd/clusterfederatedtrustdomains.spiffe.io \
#  crd/controllermanagerconfigs.spire.spiffe.io 2>/dev/null || true
#
## Force update namespace metadata for the main spire release
#echo "Updating namespace metadata..."
#kubectl patch namespace spire-system --type=merge -p '{
#  "metadata": {
#    "annotations": {
#      "meta.helm.sh/release-name": "spire",
#      "meta.helm.sh/release-namespace": "spire-system"
#    },
#    "labels": {
#      "app.kubernetes.io/managed-by": "Helm"
#    }
#  }
#}'
#
## Verify the metadata was added
#echo "Verifying namespace metadata..."
#kubectl get namespace spire-system -o jsonpath='{.metadata}' | jq '.annotations, .labels'
#
## Install/Upgrade SPIRE
#echo "Installing/Upgrading SPIRE..."
#helm upgrade --install -n spire-system spire spire \
#  --repo https://spiffe.github.io/helm-charts-hardened/ \
#  -f ./config/spire/helm/values.yaml
##   --wait
#
#echo "SPIRE setup completed successfully."

## Enable alias expansion in non-interactive shells
#shopt -s expand_aliases
#
#source ./hack/lib/aliases.sh
#
## Add Helm repository if it doesn't exist
#if ! helm repo list | grep -q "^spiffe\s"; then
#    echo "Adding spiffe Helm repository..."
#    helm repo add spiffe https://spiffe.github.io/helm-charts/
#else
#    echo "spiffe Helm repository already exists."
#fi
#
## Update Helm repositories
#echo "Updating Helm repositories..."
#helm repo update
#
## Install/Upgrade CRDs (let Helm create the namespace)
#echo "Installing/Upgrading SPIRE CRDs..."
#helm upgrade --install -n spire-system spire-crds spire-crds \
#  --repo https://spiffe.github.io/helm-charts-hardened/ \
#  --create-namespace
#
## Wait for CRDs to be established
#echo "Waiting for CRDs to be ready..."
#kubectl wait --for condition=established --timeout=60s \
#  crd/clusterspiffeids.spiffe.io \
#  crd/clusterfederatedtrustdomains.spiffe.io \
#  crd/controllermanagerconfigs.spire.spiffe.io 2>/dev/null || true
#
## Install/Upgrade SPIRE
#echo "Installing/Upgrading SPIRE..."
#helm upgrade --install -n spire-system spire spire \
#  --repo https://spiffe.github.io/helm-charts-hardened/ \
#  -f ./config/spire/helm/values.yaml \
#  --wait
#
#echo "SPIRE setup completed successfully."

##!/bin/bash
#set -e  # Exit on any error
#
## Enable alias expansion in non-interactive shells
#shopt -s expand_aliases
#
#source ./hack/lib/aliases.sh
#
## Add Helm repository if it doesn't exist
#if ! helm repo list | grep -q "^spiffe\s"; then
#    echo "Adding spiffe Helm repository..."
#    helm repo add spiffe https://spiffe.github.io/helm-charts/
#else
#    echo "spiffe Helm repository already exists."
#fi
#
## Update Helm repositories
#echo "Updating Helm repositories..."
#helm repo update
#
## Install CRDs first
#echo "Installing/Upgrading SPIRE CRDs..."
#helm upgrade --install spire-crds spiffe/spire-crds \
#  --namespace spire-server \
#  --create-namespace
#
## Wait for CRDs to be established
#echo "Waiting for CRDs to be ready..."
#kubectl wait --for condition=established --timeout=60s \
#  crd/clusterspiffeids.spiffe.io \
#  crd/clusterfederatedtrustdomains.spiffe.io \
#  crd/controllermanagerconfigs.spire.spiffe.io 2>/dev/null || true
#
## Install SPIRE
#echo "Installing/Upgrading SPIRE..."
#helm upgrade --install spire spiffe/spire \
#  --namespace spire-server \
#  -f ./config/spire/helm/values.yaml \
#  --wait \
#  --timeout 10m
#
#echo "SPIRE setup completed successfully."


#set -e  # Exit on any error
#
## Enable alias expansion in non-interactive shells
#shopt -s expand_aliases
#
#source ./hack/lib/aliases.sh
#
## Add Helm repository if it doesn't exist
#if ! helm repo list | grep -q "^spiffe\s"; then
#    echo "Adding spiffe Helm repository..."
#    helm repo add spiffe https://spiffe.github.io/helm-charts/
#else
#    echo "spiffe Helm repository already exists."
#fi
#
## Update Helm repositories
#echo "Updating Helm repositories..."
#helm repo update
#
## Create namespace with proper metadata first
#if ! kubectl get namespace spire-server &> /dev/null; then
#    echo "Creating spire-server namespace..."
#    kubectl create namespace spire-server
#    kubectl annotate namespace spire-server meta.helm.sh/release-name=spire --overwrite
#    kubectl annotate namespace spire-server meta.helm.sh/release-namespace=spire-server --overwrite
#    kubectl label namespace spire-server app.kubernetes.io/managed-by=Helm --overwrite
#fi
#
### Create namespace if it doesn't exist
##if ! kubectl get namespace spire-system &> /dev/null; then
##    echo "Creating spire-system namespace..."
##    kubectl create namespace spire-system
##else
##    echo "spire-system namespace already exists."
##fi
#
## Install/Upgrade CRDs
#echo "Installing/Upgrading SPIRE CRDs..."
#helm upgrade --install -n spire-server spire-crds spire-crds \
#  --repo https://spiffe.github.io/helm-charts-hardened/
#
## Wait for CRDs to be established
#kubectl wait --for condition=established --timeout=60s \
#  crd/clusterspiffeids.spiffe.io || true
#
## Install/Upgrade SPIRE
#echo "Installing/Upgrading SPIRE..."
#helm upgrade --install -n spire-server spire spire \
#  --repo https://spiffe.github.io/helm-charts-hardened/ \
#  -f ./config/spire/helm/values.yaml \
#  --wait
#
#echo "SPIRE setup completed successfully."
