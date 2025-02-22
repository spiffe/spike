#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

TOKEN_FILE=".spire-agent-join-token"

if ! command -v spire-server &>/dev/null; then
    echo "Error: spire-server binary not found in PATH"
    echo "Please install SPIRE to your system."
    exit 1
fi

if ! command -v spire-agent &>/dev/null; then
    echo "Error: spire-agent binary not found in PATH"
    echo "Please install SPIRE to your system."
    exit 1
fi

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

# Running spire-agent as super user to read meta information of other users'
# processes. If you are using the current user to use SPIKE only, then you
# can run this command without sudo.
if [ "$1" == "--use-sudo" ]; then
  sudo "$WORKSPACE"/spire/bin/spire-agent run \
    -config ./config/spire/agent/agent.conf \
    -joinToken "$JOIN_TOKEN"
else
  "$WORKSPACE"/spire/bin/spire-agent run \
    -config ./config/spire/agent/agent.conf \
    -joinToken "$JOIN_TOKEN"
fi
