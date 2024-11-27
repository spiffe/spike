#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# TODO: redirect the output of background processes to files.

# Trap script exit and cleanup background processes
cleanup() {
    echo "Cleaning up background processes..."
    # Kill the entire process group to ensure all child processes are terminated
    kill -- -$$
    exit
}

# Set up trap for script termination
trap cleanup EXIT SIGINT SIGTERM

# Start SPIRE server in background and save its PID
./hack/spire-server-start.sh &
SPIRE_PID=$!

# Wait for SPIRE server to initialize
echo "Waiting for SPIRE server to start..."
sleep 5

# Run the registration scripts
echo "Generating agent token..."
./hack/generate-agent-token.sh

echo "Registering SPIRE entries..."
./hack/register-spire-entries.sh

echo "Registering SU..."
./hack/register-su.sh

echo "Waiting before starting SPIRE Agent"
sleep 5
# Start SPIRE agent in background and save its PID
echo "Starting SPIRE agent..."
./hack/spire-agent-start.sh &
SPIRE_AGENT_PID=$!

echo "Waiting before keeper"
sleep 5
./keeper &
KEEPER_PID=$!

echo "Waiting before nexus"
sleep 5
./nexus &
NEXUS_PID=$!

# Wait for any remaining processes
wait $SPIRE_PID $SPIRE_AGENT_PID $KEEPER_PID $NEXUS_PID

echo "Everything is awesome!"