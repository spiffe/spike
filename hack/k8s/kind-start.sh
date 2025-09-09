#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Enable alias expansion in non-interactive shells
shopt -s expand_aliases

# Check if kind binary is present
if ! command -v kind &> /dev/null
then
  echo "Command 'kind' not found. Please install Kind first."
  echo "You can install it from: https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
  exit 1
fi

if ! command -v kubectl &> /dev/null
then
  echo "Command 'kubectl' not found. Please install kubectl first."
  exit 1
fi

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
  echo "Error: Docker daemon is not running"
  echo "Please start Docker before proceeding"
  exit 1
fi

# Set default values
DEFAULT_CLUSTER_NAME="spike-cluster"
DEFAULT_WORKERS="2"
DEFAULT_VERSION="v1.29.0"

# Use environment variables if set, otherwise use defaults
CLUSTER_NAME="${KIND_CLUSTER_NAME:-$DEFAULT_CLUSTER_NAME}"
WORKERS="${KIND_WORKERS:-$DEFAULT_WORKERS}"
K8S_VERSION="${KIND_K8S_VERSION:-$DEFAULT_VERSION}"

# Display the configuration being used
echo "Starting Kind cluster with the following configuration:"
echo "  Cluster Name: ${CLUSTER_NAME}"
echo "  Workers: ${WORKERS}"
echo "  Kubernetes Version: ${K8S_VERSION}"
echo ""

# Create Kind cluster configuration
cat > /tmp/kind-config.yaml <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF

# Add worker nodes if specified
for ((i=1; i<=WORKERS; i++)); do
  cat >> /tmp/kind-config.yaml <<EOF
- role: worker
EOF
done

echo "Creating Kind cluster..."
kind create cluster \
  --name="${CLUSTER_NAME}" \
  --image="kindest/node:${K8S_VERSION}" \
  --config=/tmp/kind-config.yaml

# Clean up config file
rm -f /tmp/kind-config.yaml

# Set kubectl context
kubectl cluster-info --context kind-${CLUSTER_NAME}

echo ""
echo "Kind cluster '${CLUSTER_NAME}' is ready!"
echo "To switch context: kubectl config use-context kind-${CLUSTER_NAME}"
echo "To install ingress: kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml"

# Wait for the cluster to be ready
echo "Waiting for cluster to be ready..."
kubectl wait --for=condition=Ready nodes --all --timeout=300s

echo ""
echo "Available nodes:"
kubectl get nodes