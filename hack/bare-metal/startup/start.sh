#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# ./hack/bare-metal/startup/start.sh
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
  exit 1
fi

# Your existing script continues here
echo "Domain check passed. Continuing with the script..."

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

# Check if SPIKE_NEXUS_BACKEND_STORE is set to memory:
if [ "$SPIKE_NEXUS_BACKEND_STORE" != "memory" ]; then
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
else
  echo "SPIKE_NEXUS_BACKEND_STORE is set to memory, skipping Keeper instances."
fi

if [ -z "$SPIKE_SKIP_NEXUS_START" ]; then
  echo ""
  echo "Waiting before SPIKE Nexus..."
  sleep 5
  run_background "./hack/bare-metal/startup/start-nexus.sh"
else
  echo "SPIKE_SKIP_NEXUS_START is set, skipping Nexus start."
fi

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
