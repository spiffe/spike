#!/bin/bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Demo Installation:
# Installs SPIRE and SPIKE to the Management Cluster
# Uses the local container registry for SPIKE images.
#
# Note that this is NOT a SPIRE production setup.
# Consult SPIRE documentation for production deployment and hardening:
# https://spiffe.io/docs/latest/spire-helm-charts-hardened-about/recommendations/

set -e  # Exit on any error

source ./hack/lib/k8s.sh

pre_install

# Install SPIKE from feature branch until it gets merged to upstream:
# See: https://github.com/spiffe/helm-charts-hardened/pull/591
cd ..
helm upgrade --install -n spire-mgmt spiffe \
  ./helm-charts-hardened/charts/spire \
  -f ./spike/config/helm/values-demo-mgmt.yaml

echo "Sleeping for 15 secs..."
sleep 15

echo "Everything is awesome!"
