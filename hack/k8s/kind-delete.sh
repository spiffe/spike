#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Enable alias expansion in non-interactive shells
shopt -s expand_aliases

command_exists() {
  command -v "$1" >/dev/null 2>&1
}

# Check for Docker
if ! command_exists docker; then
  echo "Error: Docker is not installed or not in PATH"
  echo "Please install Docker from https://docs.docker.com/get-docker/"
  exit 1
fi

# Check for Kind
if ! command_exists kind; then
  echo "Error: Kind is not installed or not in PATH"
  echo "Please install Kind from https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
  exit 1
fi

# Verify Docker is running (optional but recommended)
if ! docker info >/dev/null 2>&1; then
  echo "Warning: Docker daemon is not running"
  echo "Please start Docker before proceeding"
  exit 1
fi

# Set default cluster name
DEFAULT_CLUSTER_NAME="spike-cluster"
CLUSTER_NAME="${KIND_CLUSTER_NAME:-$DEFAULT_CLUSTER_NAME}"

# Check if the cluster exists
if ! kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
  echo "Kind cluster '${CLUSTER_NAME}' does not exist."
  echo "Available clusters:"
  kind get clusters
  exit 1
fi

echo "Deleting Kind cluster: ${CLUSTER_NAME}"
echo ""

# If all checks pass, delete the cluster
kind delete cluster --name="${CLUSTER_NAME}"

echo "Kind cluster '${CLUSTER_NAME}' has been deleted."