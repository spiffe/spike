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
if ./hack/spire-server-entry-spike-register.sh; then
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

if [ "$1" == "--use-sudo" ]; then
  echo "Please enter sudo password if prompted..."
  sudo -v
fi

echo ""
echo "Waiting before starting SPIRE Agent"
sleep 5

# Start SPIRE agent in background and save its PID
if [ "$1" == "--use-sudo" ]; then
  run_background "./hack/spire-agent-start.sh" --use-sudo
else
  run_background "./hack/spire-agent-start.sh"
fi

echo ""
echo "Waiting before SPIKE Keeper 1..."
sleep 5
run_background "./hack/start-keeper-1.sh"
echo ""
echo "Waiting before SPIKE Keeper 2..."
sleep 5
run_background "./hack/start-keeper-2.sh"
echo ""
echo "Waiting before SPIKE Keeper 3..."
sleep 5
run_background "./hack/start-keeper-3.sh"

echo ""
echo "Waiting before SPIKE Nexus..."
sleep 5
run_background "./hack/start-nexus.sh"

echo ""
echo ""
echo "<<"
echo ">"
echo "> Everything is set up."
echo "> You can now experiment with SPIKE."
echo ">"
echo "<<"
echo "> >> To begin, run './spike' on a separate terminal window."
echo "<<"
echo ">"
echo "> When you are done with your experiments, you can press 'Ctrl+C'"
echo "> on this terminal to exit and cleanup all background processes."
echo ">"
echo "<<"
echo ""
echo ""

# Wait indefinitely
while true; do
  sleep 1
done