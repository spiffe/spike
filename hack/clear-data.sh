#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

mkdir -p ./.data

cd ./.data || exit

# shellcheck disable=SC2035
rm -rf *

mkdir -p "$HOME"/.spike/data
rm -rf "$HOME"/.spike/data/spike.nexus.bootstrap.tombstone
