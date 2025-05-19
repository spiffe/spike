#!/bin/bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Enable alias expansion in non-interactive shells
shopt -s expand_aliases

set -e

export SPIKE_VERSION="0.4.1"

# Tag the images for the MicroK8s registry (which runs on localhost:32000)
docker tag spike-keeper:$SPIKE_VERSION localhost:32000/spike-keeper:$SPIKE_VERSION
docker tag spike-nexus:$SPIKE_VERSION localhost:32000/spike-nexus:$SPIKE_VERSION
docker tag spike-pilot:$SPIKE_VERSION localhost:32000/spike-pilot:$SPIKE_VERSION

# Push the images to the MicroK8s registry
docker push localhost:32000/spike-keeper:$SPIKE_VERSION
docker push localhost:32000/spike-nexus:$SPIKE_VERSION
docker push localhost:32000/spike-pilot:$SPIKE_VERSION

# Verify the images are in the registry
# The registry API can be queried to list available images
curl http://localhost:32000/v2/_catalog

# To see the tags for a specific image
curl http://localhost:32000/v2/spike-keeper/tags/list
curl http://localhost:32000/v2/spike-nexus/tags/list
curl http://localhost:32000/v2/spike-pilot/tags/list