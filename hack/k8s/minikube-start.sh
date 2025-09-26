#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Enable alias expansion in non-interactive shells
shopt -s expand_aliases

# Check if minikube binary is present
if ! command -v minikube &> /dev/null
then
  echo "Command 'minikube' not found. Please install Minikube first."
  exit 1
fi

if ! command -v kubectl &> /dev/null
then
  echo "Command 'kubectl' not found. Please install kubectl first."
  exit 1
fi

# Set default values
SPIKE_DEFAULT_MEMORY="4096"  # 4GB
SPIKE_DEFAULT_CPU="2"        # 2 CPUs
SPIKE_DEFAULT_NODES="1"      # 1 node

# Use environment variables if set, otherwise use defaults
MEMORY="${MEMORY:-$SPIKE_DEFAULT_MEMORY}"
CPU="${CPU:-$SPIKE_DEFAULT_CPU}"
NODES="${NODES:-$SPIKE_DEFAULT_NODES}"

# Display the configuration being used
echo "Starting Minikube with the following configuration:"
echo "  Memory: ${MEMORY}MB"
echo "  CPUs: ${CPU}"
echo "  Nodes: ${NODES}"
echo ""
echo "eval #(minikube -p minikube docker-env)"

# Minikube might need additional flags for SPIRE to work properly.
# A bare-metal or cloud Kubernetes cluster will not need these extra configs.
minikube start \
  --memory="$MEMORY" \
  --cpus="$CPU" \
  --nodes="$NODES" \
  --insecure-registry "10.0.0.0/24"

echo "waiting 10 secs before enabling registry"

sleep 10
minikube addons enable registry
minikube addons enable csi-hostpath-driver

kubectl get node
