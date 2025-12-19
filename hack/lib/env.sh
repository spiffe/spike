#!/usr/bin/env bash

#    \ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Source this file in your ~/.profile (or ~/.zprofile, or the profile file
# of your active shell) for your convenience.
#
# NOTE: This file contains example environment variables for local development
# and debugging. Some values intentionally deviate from production defaults
# (e.g., SPIKE_SYSTEM_LOG_LEVEL is set to "debug" here).
#
# For the authoritative list of all environment variables and their default
# values, see: docs-src/content/usage/configuration.md
# Or online at: https://spike.ist/usage/configuration/

# Development/debugging overrides:
export SPIKE_SYSTEM_LOG_LEVEL="debug"
export SPIKE_STACK_TRACES_ON_LOG_FATAL="false"

# Backend configuration:
export SPIKE_NEXUS_BACKEND_STORE="sqlite"
# export SPIKE_NEXUS_BACKEND_STORE="memory"

# Network configuration:
export SPIKE_NEXUS_TLS_PORT=":8553"
export SPIKE_NEXUS_API_URL="https://localhost:8553"
export SPIKE_KEEPER_TLS_PORT=":8443"
export SPIKE_NEXUS_KEEPER_PEERS='https://localhost:8443,https://localhost:8543,https://localhost:8643'

# Trust configuration:
export SPIKE_TRUST_ROOT="spike.ist"

# Shamir secret sharing:
export SPIKE_NEXUS_SHAMIR_SHARES="3"
export SPIKE_NEXUS_SHAMIR_THRESHOLD="2"

# To modify bare-metal startup behavior:
# . ~/.zshrc
# export SPIKE_SKIP_REGISTER_ENTRIES="true"
# export SPIKE_SKIP_CLEAR_DATA="true"
# export SPIKE_SKIP_SPIKE_BUILD="true"
# export SPIKE_SKIP_SPIRE_SERVER_START="true"
# export SPIKE_SKIP_GENERATE_AGENT_TOKEN="true"
# export SPIKE_SKIP_SPIRE_AGENT_START="true"
# export SPIKE_SKIP_KEEPER_INITIALIZATION="true"
# export SPIKE_SKIP_NEXUS_START="true"
