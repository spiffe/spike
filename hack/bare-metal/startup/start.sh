#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

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
#
# Environment variables for skipping startup steps:
# - SPIKE_SKIP_CLEAR_DATA: Skip clearing existing data
# - SPIKE_SKIP_SPIKE_BUILD: Skip building SPIKE binaries
# - SPIKE_SKIP_SPIRE_SERVER_START: Skip starting SPIRE server
# - SPIKE_SKIP_GENERATE_AGENT_TOKEN: Skip generating SPIRE agent token
# - SPIKE_SKIP_REGISTER_ENTRIES: Skip registering SPIRE entries
# - SPIKE_SKIP_SPIRE_AGENT_START: Skip starting SPIRE agent
# - SPIKE_SKIP_KEEPER_INITIALIZATION: Skip initializing SPIKE Keeper instances
# - SPIKE_SKIP_NEXUS_START: Skip starting SPIKE Nexus
#
# Additional environment variables:
# - SPIKE_NEXUS_BACKEND_STORE: When set to "memory", skips Keeper instances

# Enable alias expansion in non-interactive shells
shopt -s expand_aliases

# Domain to check
SPIRE_SERVER_DOMAIN="spire.spike.ist"

check_domain() {
  # First try DNS resolution with dig
  DNS_ANSWER=$(dig +noall +answer "$SPIRE_SERVER_DOMAIN" | grep -v "^;")
    
  # If dig doesn't find it, check if getent exists and try it
  if [ -z "$DNS_ANSWER" ]; then
    echo "No DNS record found for $SPIRE_SERVER_DOMAIN. using other methods..."

    # Check if getent is available
    if command -v getent >/dev/null 2>&1; then
      HOSTS_ANSWER=$(getent hosts "$SPIRE_SERVER_DOMAIN")
    else
      # Fallback for systems without getent (like macOS)
      HOSTS_ANSWER=$(grep "$SPIRE_SERVER_DOMAIN" /etc/hosts | grep -v "^#")
    fi
        
    # If hosts file check also fails, return error
    if [ -z "$HOSTS_ANSWER" ]; then
      echo "Error: Could not resolve $SPIRE_SERVER_DOMAIN through any means."
      return 1
    else
      echo "Found $SPIRE_SERVER_DOMAIN in hosts file:"
      echo "$HOSTS_ANSWER"
    fi
  else
    echo "DNS resolution for $SPIRE_SERVER_DOMAIN:"
    echo "$DNS_ANSWER"
  fi

  return 0
}

# Check domain before proceeding
if ! check_domain; then
  echo "Domain check failed. Exiting..."
  echo "Please add '127.0.0.1 spire.spike.ist' to your /etc/hosts."
  exit 1
fi

# Your existing script continues here
echo "Domain check passed. Continuing with the script..."

# Check for required SPIKE binaries
echo "Checking for required SPIKE binaries..."
REQUIRED_BINARIES=("spike" "nexus" "keeper" "bootstrap")
MISSING_BINARIES=()

for binary in "${REQUIRED_BINARIES[@]}"; do
  if ! command -v "$binary" >/dev/null 2>&1; then
    MISSING_BINARIES+=("$binary")
  fi
done

if [ ${#MISSING_BINARIES[@]} -gt 0 ]; then
  echo "Error: The following required binaries are not found in PATH:"
  for missing in "${MISSING_BINARIES[@]}"; do
    echo "  - $missing"
  done
  echo ""
  echo "Please build SPIKE binaries first by running:"
  echo "  ./hack/bare-metal/build/build-spike.sh"
  echo ""
  echo "If you have the binaries built, ensure the binaries are in your PATH."
  exit 1
fi

echo "All required SPIKE binaries found."

# Helpers
source ./hack/lib/bg.sh

if [ -z "$SPIKE_SKIP_CLEAR_DATA" ]; then
  if ./hack/bare-metal/build/clear-data.sh; then
    echo "Data cleared successfully"
  else
    echo "Failed to clear data"
    exit 1
  fi
else
  echo "SPIKE_SKIP_CLEAR_DATA is set, skipping data clear."
fi

if [ -z "$SPIKE_SKIP_SPIKE_BUILD" ]; then
  if ./hack/bare-metal/build/build-spike.sh; then
    echo "SPIKE binaries built successfully"
  else
    echo "Failed to build SPIKE binaries"
    exit 1
  fi
else
  echo "SPIKE_SKIP_SPIKE_BUILD is set, skipping SPIKE build."
fi

# Start SPIRE server in background and save its PID
if [ -z "$SPIKE_SKIP_SPIRE_SERVER_START" ]; then
  run_background "./hack/bare-metal/startup/spire-server-start.sh"
  # Wait for SPIRE server to initialize
  echo "Waiting for SPIRE server to start..."
  sleep 5
else
  echo "SPIKE_SKIP_SPIRE_SERVER_START is set, skipping SPIRE server start."
fi


# Run the registration scripts
if [ -z "$SPIKE_SKIP_GENERATE_AGENT_TOKEN" ]; then
  echo "Generating agent token..."
  if ./hack/bare-metal/entry/spire-server-generate-agent-token.sh; then
    echo "Agent token retrieved successfully"
  else
    echo "Failed to retrieve agent token"
    exit 1
  fi
else
  echo "SPIKE_SKIP_GENERATE_AGENT_TOKEN is set, skipping agent token generation."
fi

if [ -z "$SPIKE_SKIP_REGISTER_ENTRIES" ]; then
  echo "Registering SPIRE entries..."
  if ./hack/bare-metal/entry/spire-server-entry-spike-register.sh; then
    echo "SPIRE entries registered successfully"
  else
    echo "Failed to register SPIRE entries"
    exit 1
  fi

  echo "Registering SU..."
  if ./hack/bare-metal/entry/spire-server-entry-su-register.sh; then
    echo "SU registered successfully"
  else
    echo "Failed to register SU"
    exit 1
  fi
else
  echo "SPIKE_SKIP_REGISTER_ENTRIES is set, skipping entries registration."
fi

if [ "$1" == "--use-sudo" ]; then
  echo "Please enter sudo password if prompted..."
  sudo -v
fi

# Start SPIRE agent in background and save its PID
if [ -z "$SPIKE_SKIP_SPIRE_AGENT_START" ]; then
  echo ""
  echo "Waiting before starting SPIRE Agent"
  sleep 5
  
  if [ "$1" == "--use-sudo" ]; then
    run_background "./hack/bare-metal/startup/spire-agent-start.sh" --use-sudo
  else
    run_background "./hack/bare-metal/startup/spire-agent-start.sh"
  fi
else
  echo "SPIKE_SKIP_SPIRE_AGENT_START is set, skipping SPIRE agent start."
fi

# No SPIKE Keeper initialization is required for the "in-memory" backing store:
if [ "$SPIKE_NEXUS_BACKEND_STORE" == "memory" ]; then
  echo "SPIKE_NEXUS_BACKEND_STORE is set to memory, skipping Keeper instances."
  SPIKE_SKIP_KEEPER_INITIALIZATION="true"
fi

# Check if we want to skip the keeper initialization step:
if [ -z "$SPIKE_SKIP_KEEPER_INITIALIZATION" ]; then
  echo ""
  echo "Waiting before SPIKE Keeper 1..."
  sleep 5
  run_background "./hack/bare-metal/startup/start-keeper-1.sh"
  echo ""
  echo "Waiting before SPIKE Keeper 2..."
  sleep 5
  run_background "./hack/bare-metal/startup/start-keeper-2.sh"
  echo ""
  echo "Waiting before SPIKE Keeper 3..."
  sleep 5
  run_background "./hack/bare-metal/startup/start-keeper-3.sh"
else
  echo "SPIKE_SKIP_KEEPER_INITIALIZATION is set, skipping Keeper instances."

fi

if [ -z "$SPIKE_SKIP_NEXUS_START" ]; then
  echo ""
  echo "Waiting before SPIKE Nexus..."
  sleep 5
  run_background "./hack/bare-metal/startup/start-nexus.sh"
else
  echo "SPIKE_SKIP_NEXUS_START is set, skipping Nexus start."
  echo ""
fi

echo ""
echo "Waiting before SPIKE Bootstrap..."
sleep 5
run_background ./hack/bare-metal/startup/bootstrap.sh

echo ""
echo "Registering entries for the demo workload..."
./examples/consume-secrets/demo-register-entry.sh

echo ""
echo "Waiting a bit more for the entries to marinate..."
sleep 5

echo ""
echo "Creating policies for the demo workload..."
./examples/consume-secrets/demo-create-policy.sh

echo ""
echo "Running demo workload to verify setup..."
DEMO_OUTPUT=$(demo 2>&1)
DEMO_EXIT_CODE=$?

if [ $DEMO_EXIT_CODE -ne 0 ]; then
  echo "Error: Demo workload failed with exit code $DEMO_EXIT_CODE"
  echo "Output:"
  echo "$DEMO_OUTPUT"
  exit 1
fi

# Validate expected output
echo "$DEMO_OUTPUT" | grep -q "SPIKE Demo" || \
  { echo "Error: Missing 'SPIKE Demo' in output"; exit 1; }
echo "$DEMO_OUTPUT" | grep -q "Connected to SPIKE Nexus." || \
  { echo "Error: Missing 'Connected to SPIKE Nexus.' in output"; exit 1; }
echo "$DEMO_OUTPUT" | grep -q "Secret found:" || \
  { echo "Error: Missing 'Secret found:' in output"; exit 1; }
echo "$DEMO_OUTPUT" | grep -q "password: SPIKE_Rocks" || \
  { echo "Error: Missing expected password in output"; exit 1; }
echo "$DEMO_OUTPUT" | grep -q "username: SPIKE" || \
  { echo "Error: Missing expected username in output"; exit 1; }

echo "Demo workload verification passed."

echo ""
echo "Verifying policies..."
POLICY_OUTPUT=$(spike policy list 2>&1)
POLICY_EXIT_CODE=$?

if [ $POLICY_EXIT_CODE -ne 0 ]; then
  echo "Error: Policy list failed with exit code $POLICY_EXIT_CODE"
  echo "Output:"
  echo "$POLICY_OUTPUT"
  exit 1
fi

# Validate expected policy output
echo "$POLICY_OUTPUT" | grep -q "POLICIES" || \
  { echo "Error: Missing 'POLICIES' header in output"; exit 1; }
echo "$POLICY_OUTPUT" | grep -q "workload-can-read" || \
  { echo "Error: Missing 'workload-can-read' policy"; exit 1; }
echo "$POLICY_OUTPUT" | grep -q "workload-can-write" || \
  { echo "Error: Missing 'workload-can-write' policy"; exit 1; }
echo "$POLICY_OUTPUT" | \
  grep -qF "SPIFFE ID Pattern: ^spiffe://spike\.ist/workload/.*$" || \
  { echo "Error: Missing expected SPIFFE ID pattern"; exit 1; }
echo "$POLICY_OUTPUT" | \
  grep -qF "Path Pattern: ^tenants/demo/db/.*$" || \
  { echo "Error: Missing expected path pattern"; exit 1; }
echo "$POLICY_OUTPUT" | grep -q "Permissions: read" || \
  { echo "Error: Missing read permission"; exit 1; }
echo "$POLICY_OUTPUT" | grep -q "Permissions: write" || \
  { echo "Error: Missing write permission"; exit 1; }

echo "Policy verification passed."

echo "Done. Will sleep a bit..."
sleep 5

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
