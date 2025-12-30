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

  # Only check for binaries if we're skipping the build step
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
    echo "Or unset SPIKE_SKIP_SPIKE_BUILD to build them automatically."
    exit 1
  fi

  echo "All required SPIKE binaries found."
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

# DEBUG MODE: Log failures instead of exiting to allow investigation
VALIDATION_FAILED=false

if [ $DEMO_EXIT_CODE -ne 0 ]; then
  echo "WARNING: Demo workload failed with exit code $DEMO_EXIT_CODE"
  echo "Output:"
  echo "$DEMO_OUTPUT"
  VALIDATION_FAILED=true
fi

# Validate expected output (warnings only, no exit)
echo "$DEMO_OUTPUT" | grep -q "SPIKE Demo" || \
  { echo "WARNING: Missing 'SPIKE Demo' in output"; VALIDATION_FAILED=true; }
echo "$DEMO_OUTPUT" | grep -q "Connected to SPIKE Nexus." || \
  { echo "WARNING: Missing 'Connected to SPIKE Nexus.' in output"; VALIDATION_FAILED=true; }
echo "$DEMO_OUTPUT" | grep -q "Secret found:" || \
  { echo "WARNING: Missing 'Secret found:' in output"; VALIDATION_FAILED=true; }
echo "$DEMO_OUTPUT" | grep -q "password: SPIKE_Rocks" || \
  { echo "WARNING: Missing expected password in output"; VALIDATION_FAILED=true; }
echo "$DEMO_OUTPUT" | grep -q "username: SPIKE" || \
  { echo "WARNING: Missing expected username in output"; VALIDATION_FAILED=true; }

if [ "$VALIDATION_FAILED" = true ]; then
  echo ""
  echo "=========================================="
  echo "DEMO VALIDATION FAILED (debug mode - continuing anyway)"
  echo "Full demo output:"
  echo "$DEMO_OUTPUT"
  echo "=========================================="
else
  echo "Demo workload verification passed."
fi

echo ""
echo "Verifying policies..."
POLICY_OUTPUT=$(spike policy list 2>&1)
POLICY_EXIT_CODE=$?

POLICY_VALIDATION_FAILED=false

if [ $POLICY_EXIT_CODE -ne 0 ]; then
  echo "WARNING: Policy list failed with exit code $POLICY_EXIT_CODE"
  echo "Output:"
  echo "$POLICY_OUTPUT"
  POLICY_VALIDATION_FAILED=true
fi

# Validate expected policy output (warnings only, no exit)
echo "$POLICY_OUTPUT" | grep -q "POLICIES" || \
  { echo "WARNING: Missing 'POLICIES' header in output"; POLICY_VALIDATION_FAILED=true; }
echo "$POLICY_OUTPUT" | grep -q "workload-can-read" || \
  { echo "WARNING: Missing 'workload-can-read' policy"; POLICY_VALIDATION_FAILED=true; }
echo "$POLICY_OUTPUT" | grep -q "workload-can-write" || \
  { echo "WARNING: Missing 'workload-can-write' policy"; POLICY_VALIDATION_FAILED=true; }
echo "$POLICY_OUTPUT" | grep -q "Permissions: read" || \
  { echo "WARNING: Missing read permission"; POLICY_VALIDATION_FAILED=true; }
echo "$POLICY_OUTPUT" | grep -q "Permissions: write" || \
  { echo "WARNING: Missing write permission"; POLICY_VALIDATION_FAILED=true; }

if [ "$POLICY_VALIDATION_FAILED" = true ]; then
  echo ""
  echo "=========================================="
  echo "POLICY VALIDATION FAILED (debug mode - continuing anyway)"
  echo "Full policy output:"
  echo "$POLICY_OUTPUT"
  echo "=========================================="
else
  echo "Policy verification passed."
fi

echo ""
echo "Verifying cipher (streaming mode - stdin/stdout)..."
CIPHER_TEST_INPUT="Hello SPIKE Cipher Streaming Test"
CIPHER_STREAM_OUTPUT=$(echo "$CIPHER_TEST_INPUT" | spike cipher encrypt | spike cipher decrypt 2>&1)
CIPHER_STREAM_EXIT_CODE=$?

CIPHER_STREAM_VALIDATION_FAILED=false

if [ $CIPHER_STREAM_EXIT_CODE -ne 0 ]; then
  echo "WARNING: Cipher streaming test failed with exit code $CIPHER_STREAM_EXIT_CODE"
  echo "Output:"
  echo "$CIPHER_STREAM_OUTPUT"
  CIPHER_STREAM_VALIDATION_FAILED=true
fi

if [ "$CIPHER_STREAM_OUTPUT" != "$CIPHER_TEST_INPUT" ]; then
  echo "WARNING: Cipher streaming decrypt output doesn't match input"
  echo "Expected: $CIPHER_TEST_INPUT"
  echo "Got: $CIPHER_STREAM_OUTPUT"
  CIPHER_STREAM_VALIDATION_FAILED=true
fi

if [ "$CIPHER_STREAM_VALIDATION_FAILED" = true ]; then
  echo ""
  echo "=========================================="
  echo "CIPHER STREAMING VALIDATION FAILED (debug mode - continuing anyway)"
  echo "=========================================="
else
  echo "Cipher streaming mode verification passed."
fi

echo ""
echo "Verifying cipher (file mode)..."
CIPHER_FILE_INPUT="Hello SPIKE Cipher File Test with special chars: @#$%"
CIPHER_TEMP_IN=$(mktemp)
CIPHER_TEMP_ENC=$(mktemp)
CIPHER_TEMP_DEC=$(mktemp)

# Write input to file
echo -n "$CIPHER_FILE_INPUT" > "$CIPHER_TEMP_IN"

CIPHER_FILE_VALIDATION_FAILED=false

# Encrypt file to file
spike cipher encrypt -f "$CIPHER_TEMP_IN" -o "$CIPHER_TEMP_ENC" 2>&1
CIPHER_FILE_ENCRYPT_EXIT_CODE=$?

if [ $CIPHER_FILE_ENCRYPT_EXIT_CODE -ne 0 ]; then
  echo "WARNING: Cipher file encrypt failed with exit code $CIPHER_FILE_ENCRYPT_EXIT_CODE"
  CIPHER_FILE_VALIDATION_FAILED=true
fi

# Decrypt file to file
spike cipher decrypt -f "$CIPHER_TEMP_ENC" -o "$CIPHER_TEMP_DEC" 2>&1
CIPHER_FILE_DECRYPT_EXIT_CODE=$?

if [ $CIPHER_FILE_DECRYPT_EXIT_CODE -ne 0 ]; then
  echo "WARNING: Cipher file decrypt failed with exit code $CIPHER_FILE_DECRYPT_EXIT_CODE"
  CIPHER_FILE_VALIDATION_FAILED=true
fi

# Compare output
CIPHER_FILE_OUTPUT=$(cat "$CIPHER_TEMP_DEC")
if [ "$CIPHER_FILE_OUTPUT" != "$CIPHER_FILE_INPUT" ]; then
  echo "WARNING: Cipher file decrypt output doesn't match input"
  echo "Expected: $CIPHER_FILE_INPUT"
  echo "Got: $CIPHER_FILE_OUTPUT"
  CIPHER_FILE_VALIDATION_FAILED=true
fi

# Clean up temp files
rm -f "$CIPHER_TEMP_IN" "$CIPHER_TEMP_ENC" "$CIPHER_TEMP_DEC"

if [ "$CIPHER_FILE_VALIDATION_FAILED" = true ]; then
  echo ""
  echo "=========================================="
  echo "CIPHER FILE VALIDATION FAILED (debug mode - continuing anyway)"
  echo "=========================================="
else
  echo "Cipher file mode verification passed."
fi

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
