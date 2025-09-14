#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

docker tag spike-bootstrap:dev localhost:5000/spike-bootstrap:dev
minikube image load localhost:5000/spike-bootstrap:dev
echo "Loaded spike-bootstrap:dev."

docker tag spike-demo:dev localhost:5000/spike-demo:dev
minikube image load localhost:5000/spike-demo:dev
echo "Loaded spike-demo:dev"

docker tag spike-pilot:dev localhost:5000/spike-pilot:dev
minikube image load localhost:5000/spike-pilot:dev
echo "Loaded spike-pilot:dev"

docker tag spike-nexus:dev localhost:5000/spike-nexus:dev
minikube image load localhost:5000/spike-nexus:dev
echo "Loaded spike-nexus:dev"

docker tag spike-keeper:dev localhost:5000/spike-keeper:dev
minikube image load localhost:5000/spike-keeper:dev
echo "Loaded spike-keeper:dev"
echo

echo "All images loaded to the local Minikube registry."
echo "Everything is awesome!"

