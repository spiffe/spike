#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# ./hack/start.sh
#
# This is a helper script that establishes the following tasks:
#
# 0. Build the binaries for the demo app, SPIKE Keeper, and SPIKE Nexus.
# 1. Start the SPIRE server as a background process.
# 2. Generate an agent token.
# 3. Establish trust between the SPIRE server and SPIRE agent.
# 4. Start the SPIRE agent as a background process.
# 5. Register the SPIRE entries for SPIKE Keeper and SPIKE Nexus.
# 6. Start the SPIKE Keeper as a background process.
# 7. Start the SPIKE Nexus as a background process.
#
# The script also sets up a trap to ensure that all background processes are
# terminated when the script exits.

source ./hack/lib/bg.sh

if ./hack/clear-data.sh; then
    echo "Data cleared successfully"
else
    echo "Failed to clear data"
    exit 1
fi

if ./hack/build-spike.sh; then
    echo "SPIKE binaries built successfully"
else
    echo "Failed to build SPIKE binaries"
    exit 1
fi

# Start SPIRE server in background and save its PID
run_background "./hack/spire-server-start.sh"
# Wait for SPIRE server to initialize
echo "Waiting for SPIRE server to start..."
sleep 5

# Run the registration scripts
echo "Generating agent token..."
if ./hack/spire-server-generate-agent-token.sh; then
    echo "Agent token retrieved successfully"
else
    echo "Failed to retrieve agent token"
    exit 1
fi

echo "Registering SPIRE entries..."
if ./hack/spire-server-entry-register-spike.sh; then
    echo "SPIRE entries registered successfully"
else
    echo "Failed to register SPIRE entries"
    exit 1
fi

echo "Registering SU..."
if ./hack/spire-server-entry-su-register.sh; then
    echo "SU registered successfully"
else
    echo "Failed to register SU"
    exit 1
fi

# First, authenticate sudo and keep the session alive
echo "Please enter sudo password if prompted..."
# TODO: optionally disable this part and also make the `sudo` on spire-agent-start
# optional and dependent on a environment variable.
sudo -v

echo ""
echo "Waiting before starting SPIRE Agent"
sleep 5
# Start SPIRE agent in background and save its PID
run_background "./hack/spire-agent-start.sh"

echo ""
echo "Waiting before keeper"
sleep 5
run_background "./keeper"

echo ""
echo "Waiting before nexus"
sleep 5
run_background "./nexus"

echo ""
echo ""
echo "Everything is set up. Will wait for all processes to finish."
echo "Press Ctrl+C to exit and cleanup all background processes."
echo ""
echo ""

# Wait indefinitely
while true; do
  sleep 1
done