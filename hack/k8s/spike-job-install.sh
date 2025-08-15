#!/bin/bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Installs a Kubernetes Job to boostrap SPIKE Nexus.

# Wait for the spike keeper statefulset to be ready
echo "Waiting for spiffe-spike-keeper statefulset to be ready..."
kubectl wait --for=jsonpath='{.status.readyReplicas}'=3 \
  statefulset/spiffe-spike-keeper -n spike \
  --timeout=300s

kubectl apply -f ./hack/k8s/Bootstrap.yaml

echo "Everything is awesome!"
