#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.

./spike secret put /tenants/demo/db/creds \
  username=admin \
  password=SPIKERocks
