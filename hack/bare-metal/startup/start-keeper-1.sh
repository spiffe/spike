#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Enable alias expansion in non-interactive shells
shopt -s expand_aliases

if ! command -v keeper &> /dev/null; then
  echo "Error: 'keeper' command not found. Please ensure keeper is installed and in your PATH."
  exit 1
fi

SPIKE_KEEPER_TLS_PORT=':8443' \
exec keeper
