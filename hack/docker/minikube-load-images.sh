#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

minikube image load spike-bootstrap:dev
minikube image load spike-demo:dev
minikube image load spike-pilot:dev
minikube image load spike-nexus:dev
minikube image load spike-keeper:dev
