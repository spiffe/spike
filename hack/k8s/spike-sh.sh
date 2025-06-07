#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Script to find and shell into spike pod

NAMESPACE="spike"
POD_PREFIX="spike-pilot"

# Find the pod
echo "Looking for pod with prefix '$POD_PREFIX' in namespace '$NAMESPACE'..."
POD=$(kubectl get pods -n $NAMESPACE -o name | grep $POD_PREFIX | head -1 | cut -d'/' -f2)

if [ -z "$POD" ]; then
    echo "Error: No pod found with prefix '$POD_PREFIX' in namespace '$NAMESPACE'"
    exit 1
fi

echo "Found pod: $POD"
echo "Connecting to shell..."

# Shell into the pod using sh (for busybox)
kubectl exec -it -n $NAMESPACE $POD -- sh
