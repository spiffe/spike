#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# This is a simple way to create a single-node SPIRE development setup.
git clone --single-branch --branch v1.11.2 https://github.com/spiffe/spire.git
cd spire || exit
go build ./cmd/spire-server
go build ./cmd/spire-agent
sudo mv spire-server /usr/local/bin
sudo mv spire-agent /usr/local/bin
