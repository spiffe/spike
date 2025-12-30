#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Enable alias expansion in non-interactive shells
shopt -s expand_aliases

TOKEN_FILE=".data/.spire-agent-join-token"

# Enable alias expansion in non-interactive shells
shopt -s expand_aliases

# Enable alias expansion in non-interactive shells
shopt -s expand_aliases

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

# First search in PATH
AGENT_PATH=$(command -v spire-agent)

# If not found in PATH and WORKSPACE is defined, look in WORKSPACE
if [ ! -f "$AGENT_PATH" ] && [ -n "$WORKSPACE" ]; then
  AGENT_PATH="${WORKSPACE}/spire/bin/spire-agent"
  echo "spire-agent not found in PATH, trying WORKSPACE location: $AGENT_PATH"
fi

# If still not found, return error
if [ ! -f "$AGENT_PATH" ]; then
  echo "Error: SPIRE agent not found at $AGENT_PATH"
  echo "Please make sure SPIRE is installed or WORKSPACE is correctly set."
  exit 1
fi

# Running spire-agent as super user to read meta information of other users'
# processes. If you are using the current user to use SPIKE only, then you
# can run this command without sudo.
if [ "$1" == "--use-sudo" ]; then
  exec sudo "$AGENT_PATH" run \
    -config ./config/bare-metal/agent/agent.conf \
    -joinToken "$JOIN_TOKEN"
else
  exec "$AGENT_PATH" run \
    -config ./config/bare-metal/agent/agent.conf \
    -joinToken "$JOIN_TOKEN"
fi
