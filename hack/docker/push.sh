#!/bin/bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# eval $(minikube -p minikube docker-env) -> use Minikube docker instead of host docker
# kubectl port-forward -n kube-system svc/registry 5000:80 -> port forward registry

# Enable alias expansion in non-interactive shells
shopt -s expand_aliases

set -e

# shellcheck disable=SC2046
# eval $(minikube -p minikube docker-env)

export SPIKE_VERSION="0.4.0"

REGISTRY_PORT=5000

# Tag the images for the MicroK8s registry (which runs on localhost:$REGISTRY_PORT)
docker tag spike-keeper:$SPIKE_VERSION localhost:$REGISTRY_PORT/spike-keeper:$SPIKE_VERSION
docker tag spike-nexus:$SPIKE_VERSION localhost:$REGISTRY_PORT/spike-nexus:$SPIKE_VERSION
docker tag spike-pilot:$SPIKE_VERSION localhost:$REGISTRY_PORT/spike-pilot:$SPIKE_VERSION

# Push the images to the MicroK8s registry
docker push localhost:$REGISTRY_PORT/spike-keeper:$SPIKE_VERSION
docker push localhost:$REGISTRY_PORT/spike-nexus:$SPIKE_VERSION
docker push localhost:$REGISTRY_PORT/spike-pilot:$SPIKE_VERSION

# Verify the images are in the registry
# The registry API can be queried to list available images
curl http://localhost:$REGISTRY_PORT/v2/_catalog

# To see the tags for a specific image
curl http://localhost:$REGISTRY_PORT/v2/spike-keeper/tags/list
curl http://localhost:$REGISTRY_PORT/v2/spike-nexus/tags/list
curl http://localhost:$REGISTRY_PORT/v2/spike-pilot/tags/list
