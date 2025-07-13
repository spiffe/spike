#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

command_exists() {
  command -v "$1" >/dev/null 2>&1
}

# Check for Docker
if ! command_exists docker; then
  echo "Error: Docker is not installed or not in PATH"
  echo "Please install Docker from https://docs.docker.com/get-docker/"
  exit 1
fi

# Check for Minikube
if ! command_exists minikube; then
  echo "Error: Minikube is not installed or not in PATH"
  echo "Please install Minikube from https://minikube.sigs.k8s.io/docs/start/"
  exit 1
fi

# Verify Docker is running (optional but recommended)
if ! docker info >/dev/null 2>&1; then
  echo "Warning: Docker daemon is not running"
  echo "Please start Docker before proceeding"
  exit 1
fi

# If all checks pass, run `minikube delete`
minikube delete
