#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

CURRENT_HOST=$(hostname)

# Function to set bundle
set_bundle() {
  local trust_domain=$1
  local bundle_file=$2

  echo "Setting bundle for $trust_domain..."
  echo "bundle file: $bundle_file"

  # Read the bundle content and pass it through stdin
  kubectl exec -i -n spire-server statefulset/spiffe-server -c spire-server -- \
    /opt/spire/bin/spire-server bundle set -format spiffe \
    -socketPath /tmp/spire-server/private/api.sock \
    -id "spiffe://$trust_domain" < "$bundle_file"
}


# Function to list bundles
list_bundles() {
  echo "Current bundles:"
  kubectl exec -n spire-server statefulset/spiffe-server -c spire-server -- \
    /opt/spire/bin/spire-server bundle list \
    -socketPath /tmp/spire-server/private/api.sock
}

cd ~ || exit

case $CURRENT_HOST in
  mgmt)
    echo "=== Running on ${CURRENT_HOST} host ==="

    # Check if spoke bundles exist
    if [[ -f "bundle-workload.json" && -f "bundle-edge-1.json" \
      && -f "bundle-edge-2.json" && -f "bundle-edge-3.json" ]]; then
      echo "Found spoke bundles, setting them up..."
      set_bundle "workload.spike.ist" "bundle-workload.json"
      set_bundle "edge-1.spike.ist" "bundle-edge-1.json"
      set_bundle "edge-2.spike.ist" "bundle-edge-2.json"
      list_bundles
    else
      echo "Missing bundles!"
      exit 1
    fi
    ;;

  workload|edge-1|edge-2|edge-3)
    echo "=== Running on ${CURRENT_HOST} host ==="

    # Set mgmt bundle if it exists
    if [[ -f "bundle-mgmt.json" ]]; then
      echo "Found mgmt bundle, setting it up..."
      set_bundle "mgmt.spike.ist" "bundle-mgmt.json"
    else
      echo "Missing bundles!"
      exit 1
    fi
    ;;

  *)
    echo "ERROR: Unknown hostname: $CURRENT_HOST"
    echo "Expected one of: mgmt, workload, edge-1, edge-2, edge-3"
    exit 1
    ;;
esac

echo ""
echo "Everything is awesome!"
