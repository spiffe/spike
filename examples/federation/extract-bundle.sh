#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Get the current hostname
HOSTNAME=$(hostname)

# Define the management cluster hostname
MGMT_HOST="mgmt"

# Function to extract bundle and save it
extract_bundle() {
  local bundle_name=$1
  echo "Extracting bundle as ${bundle_name}..."

  kubectl exec -n spire-server statefulset/spiffe-server -c spire-server -- \
    /opt/spire/bin/spire-server bundle show \
    -socketPath /tmp/spire-server/private/api.sock \
    -format spiffe > "${bundle_name}"

  # shellcheck disable=SC2181
  if [ $? -eq 0 ]; then
    echo "Bundle extracted successfully to ${bundle_name}"
    return 0
  else
    echo "Error: Failed to extract bundle"
    return 1
  fi
}

# Function to copy bundle to management cluster
copy_to_mgmt() {
  local bundle_file=$1
  echo "Copying ${bundle_file} to ${MGMT_HOST}..."

  if scp "${bundle_file}" "${MGMT_HOST}@spiffe-management-cluster:~/"; then
    echo "Bundle copied successfully to ${MGMT_HOST}"
    return 0
  else
    echo "Error: Failed to copy bundle to ${MGMT_HOST}"
    return 1
  fi
}

# Main logic based on hostname
case "${HOSTNAME}" in
  "mgmt")
    echo "Running on management cluster"
    extract_bundle "bundle-mgmt.json"
    if mv "bundle-mgmt.json" ~/; then
      echo "Bundle moved to ~/"
    else
      echo "Error: Failed to move bundle to home directory"
    fi

    # Distribute to all edge and workload clusters
    echo "Distributing management bundle to edge and workload clusters..."

    # Copy to edge-1
    echo "Copying to edge-1..."
    if scp "$HOME/bundle-mgmt.json" "edge-1@spiffe-edge-cluster-1:~/"; then
      echo "✓ Bundle copied to edge-1"
    else
      echo "✗ Failed to copy bundle to edge-1"
    fi

    # Copy to edge-2
    echo "Copying to edge-2..."
    if scp "$HOME/bundle-mgmt.json" "edge-2@spiffe-edge-cluster-2:~/"; then
      echo "✓ Bundle copied to edge-2"
    else
      echo "✗ Failed to copy bundle to edge-2"
    fi

    # Copy to edge-3
    echo "Copying to edge-3..."
    if scp "$HOME/bundle-mgmt.json" "edge-3@spiffe-edge-cluster-3:~/"; then
      echo "✓ Bundle copied to edge-2"
    else
      echo "✗ Failed to copy bundle to edge-3"
    fi

    # Copy to workload
    echo "Copying to workload..."
    if scp "$HOME/bundle-mgmt.json" "workload@spiffe-workload-cluster:~/"; then
      echo "✓ Bundle copied to workload"
    else
      echo "✗ Failed to copy bundle to workload"
    fi
    ;;

  "edge-1")
    echo "Running on edge-1 cluster"
    extract_bundle "bundle-edge-1.json" && \
      copy_to_mgmt "bundle-edge-1.json"
    ;;

  "edge-2")
    echo "Running on edge-2 cluster"
    extract_bundle "bundle-edge-2.json" && \
      copy_to_mgmt "bundle-edge-2.json"
    ;;

  "edge-3")
    echo "Running on edge-3 cluster"
    extract_bundle "bundle-edge-3.json" && \
      copy_to_mgmt "bundle-edge-3.json"
    ;;

  "workload")
    echo "Running on workload cluster"
    extract_bundle "bundle-workload.json" && \
      copy_to_mgmt "bundle-workload.json"
    ;;

  *)
    echo "Error: Unknown hostname '${HOSTNAME}'"
    echo "Expected: mgmt, edge-1, edge-2, edge-3, or workload"
    exit 1
    ;;
esac

# Check if everything completed successfully
# shellcheck disable=SC2181
if [ $? -eq 0 ]; then
  echo "Everything is awesome!"
else
  echo "Something went wrong during the process"
  exit 1
fi
