#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

#!/bin/bash

# Configuration
NAMESPACE="spike"
TIMEOUT=300  # 5 minutes timeout
RETRY_INTERVAL=5  # Check every 5 seconds

# Color codes for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

cleanup() {
  echo -e "\n${YELLOW}Cleaning up port-forward processes...${NC}"
  jobs -p | xargs -r kill 2>/dev/null
  exit 0
}

# Set up trap to catch exit signals
trap cleanup EXIT INT TERM

# Function to check if a pod is ready
check_pod_ready() {
  local pod_name=$1
  local namespace=$2

  # Check if pod exists
  if ! kubectl get pod "$pod_name" -n "$namespace" &>/dev/null; then
    return 1
  fi

  # Check if pod is ready
  local ready
  ready=$(kubectl get pod "$pod_name" -n "$namespace" -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}' 2>/dev/null)

  if [[ "$ready" == "True" ]]; then
    return 0
  else
    return 1
  fi
}

# Function to wait for pod to be ready
wait_for_pod() {
  local pod_name=$1
  local namespace=$2
  local start_time
  start_time=$(date +%s)

  echo -e "${YELLOW}Waiting for pod $pod_name to be ready...${NC}"

  while true; do
    if check_pod_ready "$pod_name" "$namespace"; then
      echo -e "${GREEN}✓ Pod $pod_name is ready${NC}"
      return 0
    fi

    local current_time
    current_time=$(date +%s)
    local elapsed=$((current_time - start_time))

    if [[ $elapsed -gt $TIMEOUT ]]; then
      echo -e "${RED}✗ Timeout waiting for pod $pod_name to be ready${NC}"
      return 1
    fi

    # Show pod status
    local phase
    phase=$(kubectl get pod "$pod_name" -n "$namespace" -o jsonpath='{.status.phase}' 2>/dev/null || echo "Not Found")
    echo -e "  Pod status: $phase (${elapsed}s elapsed)"

    sleep $RETRY_INTERVAL
  done
}

# Function to check if service exists
check_service_exists() {
  local service_name=$1
  local namespace=$2

  if kubectl get svc "$service_name" -n "$namespace" &>/dev/null; then
    return 0
  else
    echo -e "${RED}✗ Service $service_name not found in namespace $namespace${NC}"
    return 1
  fi
}

# Main execution
echo -e "${YELLOW}Starting port-forward setup for SPIFFE Spike Keepers...${NC}\n"

# Define pods and their corresponding services and ports
declare -A pod_configs=(
  ["spiffe-spike-keeper-0"]="8444"
  ["spiffe-spike-keeper-1"]="8543"
  ["spiffe-spike-keeper-2"]="8643"
)

# First, check all services exist
echo -e "${YELLOW}Checking services...${NC}"
all_services_exist=true
for pod_name in "${!pod_configs[@]}"; do
  if ! check_service_exists "$pod_name" "$NAMESPACE"; then
    all_services_exist=false
  fi
done

if [[ "$all_services_exist" != "true" ]]; then
  echo -e "${RED}Not all required services exist. Please check your configuration.${NC}"
  exit 1
fi

echo -e "${GREEN}✓ All services exist${NC}\n"

# Wait for all pods to be ready
echo -e "${YELLOW}Checking pod readiness...${NC}"
all_pods_ready=true
for pod_name in "${!pod_configs[@]}"; do
  if ! wait_for_pod "$pod_name" "$NAMESPACE"; then
    all_pods_ready=false
    break
  fi
done

if [[ "$all_pods_ready" != "true" ]]; then
  echo -e "${RED}Not all pods are ready. Exiting.${NC}"
  exit 1
fi

echo -e "\n${GREEN}✓ All pods are ready. Starting port forwards...${NC}\n"

# Start port-forward commands
for pod_name in "${!pod_configs[@]}"; do
  local_port=${pod_configs[$pod_name]}
  echo -e "${YELLOW}Starting port-forward for $pod_name on port $local_port...${NC}"
  kubectl -n "$NAMESPACE" port-forward "svc/$pod_name" "$local_port:443" --address=0.0.0.0 &

  # Give it a moment to start
  sleep 1

  # Check if the port-forward process is still running
  if kill -0 $! 2>/dev/null; then
    echo -e "${GREEN}✓ Port-forward for $pod_name started successfully (PID: $!)${NC}"
  else
    echo -e "${RED}✗ Failed to start port-forward for $pod_name${NC}"
  fi
done

echo -e "\n${GREEN}All port-forwards started successfully!${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop all forwards and exit.${NC}\n"

# Optional: Show connection information
echo -e "${YELLOW}Connection information:${NC}"
for pod_name in "${!pod_configs[@]}"; do
  local_port=${pod_configs[$pod_name]}
  echo -e "  • $pod_name: https://localhost:$local_port"
done
echo ""

# Wait indefinitely (this keeps the script running)
wait
