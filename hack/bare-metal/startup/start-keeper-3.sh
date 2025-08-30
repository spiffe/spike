#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

if ! command -v keeper &> /dev/null; then
  echo "Error: 'keeper' command not found. Please ensure keeper is installed and in your PATH."
  exit 1
fi

SPIKE_KEEPER_TLS_PORT=':8643' \
exec ./keeper
