#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

spike secret list

echo "listed secrets"

spike secret get test

echo "got secret"

spike policy list

echo "listed policies"

spike policy get --name=workload-can-read

echo "got policy"
