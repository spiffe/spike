#!/bin/bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Docker Complete Cleanup Script

echo "====== Starting Docker Cleanup ======"

# Stop all running containers
echo "Stopping all running containers..."
docker stop "$(docker ps -aq)" 2>/dev/null \
  || echo "No running containers to stop."

# Remove all containers
echo "Removing all containers..."
docker rm "$(docker ps -aq)" 2>/dev/null \
  || echo "No containers to remove."

# Remove all images
echo "Removing all Docker images..."
docker rmi "$(docker images -q)" --force 2>/dev/null \
  || echo "No images to remove."

# Remove all volumes
echo "Removing all Docker volumes..."
docker volume rm "$(docker volume ls -q)" 2>/dev/null \
  || echo "No volumes to remove."

# Remove all networks (except default ones)
echo "Removing all custom Docker networks..."
docker network rm \
  "$(docker network ls | grep -v "bridge\|host\|none" | awk '{print $1}')" \
  2>/dev/null || echo "No custom networks to remove."

# Remove build cache (Docker 17.06.0 and later)
echo "Pruning build cache..."
docker builder prune --all --force

# System prune (removes unused data)
echo "Running system prune to remove any remaining unused data..."
docker system prune --all --volumes --force

echo "====== Docker Cleanup Complete ======"
echo "Your Docker environment has been reset."
