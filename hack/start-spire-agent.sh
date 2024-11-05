#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

TOKEN_FILE=".spire-agent-join-token"

if [ ! -f $TOKEN_FILE ]; then
    echo "Error: token does not exist"
    exit 1
fi

# Verify file was created and is not empty
if [ ! -s $TOKEN_FILE ]; then
    echo "Error: Token file is empty or was not created" >&2
    exit 1
fi

JOIN_TOKEN=$(cat $TOKEN_FILE)
if [ -z "$JOIN_TOKEN" ]; then
    echo "Error: join token is empty"
    exit 1
fi

spire-agent run -config ./config/spire/agent/agent.conf -joinToken "$JOIN_TOKEN"
