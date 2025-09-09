#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Enable alias expansion in non-interactive shells
#
# In Bash, command -v will happily resolve aliases. But in non-interactive
# scripts (i.e., `#!/bin/bash`), aliases are disabled by default, unless the
# script enables them with `shopt -s expand_aliases`.
shopt -s expand_aliases

# Check if bootstrap binary exists in PATH
if ! command -v bootstrap &> /dev/null; then
  echo "Error: 'bootstrap' binary not found in PATH"
  exit 1
fi

# Set log level to debug to see progress.
# Bootstrap needs to know which SPIKE Keeper endpoints to send shards to.
SPIKE_SYSTEM_LOG_LEVEL="debug" \
SPIKE_NEXUS_KEEPER_PEERS='https://localhost:8443,https://localhost:8543,https://localhost:8643' \
exec bootstrap -init
