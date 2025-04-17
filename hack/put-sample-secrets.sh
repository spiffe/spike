#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

spike secret put test test=test

echo "Put secret"

spike secret list

echo "listed secrets"

spike secret get test

echo "got secret"

./examples/consume-secrets/demo-create-policy.sh

echo "created policy"

spike policy list

echo "listed policies"

spike policy get --name=workload-can-read

echo "got policy"
