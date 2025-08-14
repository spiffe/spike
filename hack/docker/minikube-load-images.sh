#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

docker tag spike-bootstrap:dev localhost:5000/spike-bootstrap:dev
minikube image load localhost:5000/spike-bootstrap:dev

docker tag spike-bootstrap:dev localhost:5000/spike-bootstrap:dev
minikube image load localhost:5000/spike-bootstrap:dev

docker tag spike-bootstrap:dev localhost:5000/spike-bootstrap:dev
minikube image load localhost:5000/spike-bootstrap:dev

docker tag spike-bootstrap:dev localhost:5000/spike-bootstrap:dev
minikube image load localhost:5000/spike-bootstrap:dev

docker tag spike-bootstrap:dev localhost:5000/spike-bootstrap:dev
minikube image load localhost:5000/spike-bootstrap:dev

minikube image load spike-demo:dev
minikube image load spike-pilot:dev
minikube image load spike-nexus:dev
minikube image load spike-keeper:dev


