#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

if ! command -v nexus &> /dev/null; then
  echo "Error: 'nexus' command not found. Please ensure nexus is installed and in your PATH."
  exit 1
fi

SPIKE_NEXUS_KEEPER_PEERS='https://localhost:8443,https://localhost:8543,https://localhost:8643' \
exec nexus
