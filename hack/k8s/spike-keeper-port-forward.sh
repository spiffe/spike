#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

cleanup() {
    echo "Cleaning up port-forward processes..."
    jobs -p | xargs -r kill
    exit 0
}

# Set up trap to catch exit signals
trap cleanup EXIT INT TERM

# Start port-forward commands in background
echo "Starting port-forward for spiffe-spike-keeper-0 on port 8443..."
kubectl -n spike port-forward svc/spiffe-spike-keeper-0 8443:443 --address=0.0.0.0 &

echo "Starting port-forward for spiffe-spike-keeper-1 on port 8543..."
kubectl -n spike port-forward svc/spiffe-spike-keeper-1 8543:443 --address=0.0.0.0 &

echo "Starting port-forward for spiffe-spike-keeper-2 on port 8643..."
kubectl -n spike port-forward svc/spiffe-spike-keeper-2 8643:443 --address=0.0.0.0 &

echo "All port-forwards started. Press Ctrl+C to stop all forwards and exit."

# Wait indefinitely (this keeps the script running)
wait
