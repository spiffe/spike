#!/bin/bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Installs SPIRE and SPIKE to the cluster.
# Uses the local container registry for SPIKE images.

# Configuration
SPIKE_USE_LOCAL_CHARTS="${SPIKE_USE_LOCAL_CHARTS:-true}" # TODO: should default to false after upstream helm is merged.
SPIKE_LOCAL_CHARTS_PATH="${SPIKE_LOCAL_CHARTS_PATH:-$HOME/WORKSPACE/helm-charts-hardened}"
SPIKE_LOCAL_CHARTS_VALUES_FILE="${SPIKE_LOCAL_CHARTS_VALUES_FILE:-./config/helm/values-local.yaml}"
SPIKE_REMOTE_CHARTS_HELM_REPO="${SPIKE_REMOTE_CHARTS_HELM_REPO:-https://spiffe.github.io/helm-charts-hardened/}"
SPIKE_REMOTE_CHARTS_VALUES_FILE="${SPIKE_REMOTE_CHARTS_VALUES_FILE:-./config/helm/values-dev.yaml}"
SPIKE_REMOTE_CHARTS_CRDS_VERSION="${SPIKE_REMOTE_CHARTS_CRDS_VERSION:-0.5.0}"
SPIKE_REMOTE_CHARTS_SPIRE_VERSION="${SPIKE_REMOTE_CHARTS_SPIRE_VERSION:-0.26.1}"

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

# Function to install a single chart
install_chart() {
  local release_name=$1
  local chart_name=$2
  shift 2
  local extra_args=("$@")

  if [ -n "${SPIKE_USE_LOCAL_CHARTS}" ]; then
    helm upgrade --install -n spire-mgmt "$release_name" \
      "${SPIKE_LOCAL_CHARTS_PATH}/charts/${chart_name}" \
      "${extra_args[@]}"
  else
    # Map chart names to version variables
    case "$chart_name" in
      "spire-crds")
        local version="$SPIKE_REMOTE_CHARTS_CRDS_VERSION"
        ;;
      "spire")
        local version="$SPIKE_REMOTE_CHARTS_SPIRE_VERSION"
        ;;
      *)
        echo "Error: Unknown chart $chart_name"
        return 1
        ;;
    esac

    helm upgrade --install -n spire-mgmt "$release_name" "$chart_name" \
      --repo "$SPIKE_REMOTE_CHARTS_HELM_REPO" \
      --version "$version" \
      "${extra_args[@]}"
  fi
}

# Function to install all charts
install_charts() {
  if [ -n "${SPIKE_USE_LOCAL_CHARTS}" ]; then
    echo "Using local charts from $SPIKE_LOCAL_CHARTS_PATH"
    local values_file="$SPIKE_LOCAL_CHARTS_VALUES_FILE"
  else
    echo "Using upstream charts from $SPIKE_REMOTE_CHARTS_HELM_REPO"
    local values_file="$SPIKE_REMOTE_CHARTS_VALUES_FILE"
  fi

  # Install spire-crds
  install_chart "spire-crds" "spire-crds" --create-namespace

  echo "Sleeping for 15 secs before installing SPIRE and SPIKE..."
  sleep 15

  # Install spire
  install_chart "spiffe" "spire" -f "$values_file"
}

# Install the charts
install_charts

echo "Sleeping for 15 secs..."
sleep 15



echo "Everything is awesome!"
