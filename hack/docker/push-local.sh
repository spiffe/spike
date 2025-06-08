#!/bin/bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Enable alias expansion in non-interactive shells
shopt -s expand_aliases

set -e

export SPIKE_VERSION="dev" # "dev" since this is a local build.
export REGISTRY_PORT=5000

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
