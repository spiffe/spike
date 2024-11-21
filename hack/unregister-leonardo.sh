#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Disable Leonardo.
spire-server entry delete \
    -entryID "fa648f59-ee7a-48f0-9f84-3f600eff1362"
