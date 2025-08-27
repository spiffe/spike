#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# see: https://spike.ist/#/configuration

# source this file in your ~/.profile for your convenience.
# Note that some of these values might deviate from the defaults to make
# debugging easier.

export SPIKE_NEXUS_MAX_SECRET_VERSIONS="10"
export SPIKE_NEXUS_BACKEND_STORE="sqlite"
# export SPIKE_NEXUS_BACKEND_STORE="memory"
export SPIKE_NEXUS_TLS_PORT=":8553"
export SPIKE_NEXUS_SHA_HASH_LENGTH="32"
export SPIKE_NEXUS_PBKDF2_ITERATION_COUNT="600000"

export SPIKE_NEXUS_DB_OPERATION_TIMEOUT="15s"
export SPIKE_NEXUS_DB_INITIALIZATION_TIMEOUT="30s"
export SPIKE_NEXUS_DB_JOURNAL_MODE="WAL"
export SPIKE_NEXUS_DB_BUSY_TIMEOUT_MS="1000"
export SPIKE_NEXUS_DB_MAX_OPEN_CONNS="10"
export SPIKE_NEXUS_DB_MAX_IDLE_CONNS="5"
export SPIKE_NEXUS_DB_CONN_MAX_LIFETIME="1h"

export SPIKE_NEXUS_RECOVERY_TIMEOUT="0"
export SPIKE_NEXUS_RECOVER_MAX_INTERVAL="30s"
export SPIKE_NEXUS_RECOVERY_POLL_INTERVAL="5s"
export SPIKE_NEXUS_KEEPER_UPDATE_INTERVAL="1m"

export SPIKE_NEXUS_SHAMIR_SHARES="3"
export SPIKE_NEXUS_SHAMIR_THRESHOLD="2"

export SPIKE_NEXUS_KEEPER_PEERS='https://localhost:8443,https://localhost:8543,https://localhost:8643'
export SPIKE_KEEPER_TLS_PORT=":8443"

export SPIKE_SYSTEM_LOG_LEVEL="debug"

export SPIKE_TRUST_ROOT="spike.ist"
export SPIKE_NEXUS_API_URL="https://localhost:8553"
