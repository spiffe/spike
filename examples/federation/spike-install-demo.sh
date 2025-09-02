#!/bin/bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
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

source ./examples/federation/lib/k8s.sh

pre_install

# Install SPIKE from feature branch until it gets merged to upstream:
# See: https://github.com/spiffe/helm-charts-hardened/pull/591
cd ..
HOSTNAME=$(hostname)
VALUES_FILE="./spike/config/helm/values-demo-mgmt.yaml"

case "$HOSTNAME" in
  "mgmt")
    echo "using values-demo-mgmt.yaml"
    VALUES_FILE="./spike/config/helm/values-demo-mgmt.yaml"
    ;;
  "edge-1")
    echo "using values-demo-edge-1.yaml"
    VALUES_FILE="./spike/config/helm/values-demo-edge-1.yaml"
    ;;
  "edge-2")
    echo "using values-demo-edge-2.yaml"
    VALUES_FILE="./spike/config/helm/values-demo-edge-2.yaml"
    ;;
  "edge-3")
    echo "using values-demo-edge-3.yaml"
    VALUES_FILE="./spike/config/helm/values-demo-edge-3.yaml"
    ;;
  "workload")
    echo "using values-demo-workload.yaml"
    VALUES_FILE="./spike/config/helm/values-demo-workload.yaml"
    ;;
esac

helm upgrade --install -n spire-mgmt spiffe \
  ./helm-charts-hardened/charts/spire \
  -f $VALUES_FILE

echo "Sleeping for 15 secs..."
sleep 15

echo "Everything is awesome!"
