#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Terminal 1 - Keeper 0 on port 8443
kubectl -n spike port-forward svc/spiffe-spike-keeper-0 8443:443 --address=0.0.0.0

# Terminal 2 - Keeper 1 on port 8444
kubectl -n spike port-forward svc/spiffe-spike-keeper-1 8543:443 --address=0.0.0.0

# Terminal 3 - Keeper 2 on port 8445
kubectl -n spike port-forward svc/spiffe-spike-keeper-2 8643:443 --address=0.0.0.0