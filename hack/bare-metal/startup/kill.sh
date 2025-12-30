#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# This script kills any dangling SPIKE-related processes that may remain
# after an incomplete or interrupted `make start` run.
#
# Processes killed:
#   - spire-server: SPIRE server
#   - spire-agent: SPIRE agent
#   - nexus: SPIKE Nexus
#   - keeper: SPIKE Keeper instances
#
# Note: bootstrap is not included as it runs briefly and exits on its own.

set -e

echo "Killing SPIKE-related processes..."

# List of process names to kill
PROCESSES=("spire-server" "spire-agent" "nexus" "keeper" "demo")

for proc in "${PROCESSES[@]}"; do
  # Check if process is running
  if pgrep -x "$proc" > /dev/null 2>&1; then
    echo "Killing $proc..."
    pkill -x "$proc" 2>/dev/null || true
  else
    echo "$proc is not running."
  fi
done

echo ""
echo "Done. All SPIKE-related processes have been terminated."