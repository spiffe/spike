#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# see: https://spike.ist/#/configuration

export SPIKE_NEXUS_POLL_INTERVAL="5s"
export SPIKE_NEXUS_MAX_SECRET_VERSIONS="10"

export SPIKE_NEXUS_TLS_PORT=":8553"
export SPIKE_KEEPER_TLS_PORT=":8443"

export SPIKE_SYSTEM_LOG_LEVEL="debug"