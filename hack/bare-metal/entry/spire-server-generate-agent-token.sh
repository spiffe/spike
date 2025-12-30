#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

TOKEN_FILE=".data/.spire-agent-join-token"

# Generate token and save to file
if ! spire-server token generate -spiffeID spiffe://spike.ist/spire-agent | \
  sed 's/^Token: //' > "$TOKEN_FILE"; then
    echo "Error: Failed to generate token" >&2
    exit 1
fi

# Verify file was created and is not empty
if [ ! -s "$TOKEN_FILE" ]; then
    echo "Error: Token file is empty or was not created" >&2
    exit 1
fi

# Set restrictive permissions
chmod 600 "$TOKEN_FILE"

echo "Token successfully generated and saved to $TOKEN_FILE"
