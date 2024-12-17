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
./hack/spire-server-start.sh > spire-server-out.log 2>&1 spire-server-err.log &
SPIRE_SERVER_PID=$!

echo "SPIRE_SERVER_PID: $SPIRE_SERVER_PID"

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

echo ""
echo "Waiting before starting SPIRE Agent"
sleep 5
# Start SPIRE agent in background and save its PID
echo "Starting SPIRE agent..."
./hack/spire-agent-start.sh > spire-agent-out.log 2> spire-agent-err.log &
SPIRE_AGENT_PID=$!

echo "SPIRE_AGENT_PID: $SPIRE_AGENT_PID"

echo ""
echo "Waiting before keeper"
sleep 5
./keeper > keeper-out.log 2> keeper-err.log &
KEEPER_PID=$!

echo "KEEPER_PID: $KEEPER_PID"

echo ""
echo "Waiting before nexus"
sleep 5
./nexus > nexus-out.log 2> nexus-err.log &
NEXUS_PID=$!

echo "NEXUS_PID: $NEXUS_PID"

# Wait for any remaining processes

echo ""
echo ""
echo "Everything is set up. Will wait for all processes to finish."
echo "Press Ctrl+C to exit and cleanup all background processes."
echo ""
echo ""
wait $SPIRE_SERVER_PID $SPIRE_AGENT_PID $KEEPER_PID $NEXUS_PID
