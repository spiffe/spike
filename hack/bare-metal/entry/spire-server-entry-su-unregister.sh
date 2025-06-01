#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

SPIFFE_ID="spiffe://spike.ist/spike/pilot/role/superuser"

# Find the Entry ID for the given SPIFFE ID
ENTRY_ID=$(spire-server entry show --spiffeID "$SPIFFE_ID" | awk '/Entry ID/ {print $NF}')

# Check if an Entry ID was found
if [ -z "$ENTRY_ID" ]; then
    echo "No entry found for SPIFFE ID: $SPIFFE_ID"
    exit 1
fi

# Delete the entry using the Entry ID
if spire-server entry delete --entryID "$ENTRY_ID"; then
    echo "Successfully deleted entry with SPIFFE ID: $SPIFFE_ID"
else
    echo "Failed to delete entry with SPIFFE ID: $SPIFFE_ID"
    exit 1
fi

# Wait for the entry to be updated
echo "Waiting for entries to be updated..."
sleep 5
echo "Everything is awesome!"