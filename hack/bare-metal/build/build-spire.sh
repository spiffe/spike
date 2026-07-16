#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# This is a simple way to create a single-node SPIRE development setup.
# It builds the SPIRE server and agent from source and installs them into
# /usr/local/bin. The source is cloned into a temporary directory that is
# removed on exit, so the repository working tree stays clean.

set -euo pipefail

SPIRE_VERSION="v1.11.2"

# Clone into a temporary directory and always clean it up on exit, even on
# failure. Leaving the clone behind would pollute the working tree and break
# `make audit` (gofmt scans it).
CLONE_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "$CLONE_DIR"
}
trap cleanup EXIT

git clone --single-branch --branch "$SPIRE_VERSION" \
  https://github.com/spiffe/spire.git "$CLONE_DIR"

cd "$CLONE_DIR"
go build ./cmd/spire-server
go build ./cmd/spire-agent
sudo mv spire-server /usr/local/bin
sudo mv spire-agent /usr/local/bin
