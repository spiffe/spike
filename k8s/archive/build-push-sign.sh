#!/bin/bash
set -e

# This script builds, pushes, and signs Docker images for SPIKE components
# Usage: ./build-push-sign.sh <app> <arch> <version> [registry] [repository]

# Default values
APP=${1:-pilot}
ARCH=${2:-linux/amd64}
VERSION=${3:-latest}
REGISTRY=${4:-ghcr.io}
REPOSITORY=${5:-spiffe}
IMAGE_URL="$REGISTRY/$REPOSITORY/spike-$APP"

# Validate required arguments
if [ -z "$APP" ] || [ -z "$ARCH" ] || [ -z "$VERSION" ]; then
  echo "Usage: ./build-push-sign.sh <app> <arch> <version> [registry] [repository]"
  echo "  <app>: pilot, keeper, or nexus"
  echo "  <arch>: linux/amd64 or linux/arm64"
  echo "  <version>: version tag (e.g., 1.2.3)"
  echo "  [registry]: Docker registry (default: docker.io)"
  echo "  [repository]: Docker repository (default: getspike)"
  exit 1
fi

# Convert architecture format for tags (replace / with -)
ARCH_TAG=$(echo $ARCH | sed 's/\//-/g')

# Generate tags
GIT_SHA=$(git rev-parse --short HEAD)
MAJOR_MINOR=$(echo $VERSION | cut -d. -f1,2)
TAGS=(
  "$IMAGE_URL:${VERSION}-${ARCH_TAG}"
  "$IMAGE_URL:${MAJOR_MINOR}-${ARCH_TAG}"
  "$IMAGE_URL:latest-${ARCH_TAG}"
  "$IMAGE_URL:${GIT_SHA}-${ARCH_TAG}"
)
# Only add the latest tag for amd64
if [[ "$ARCH" == "linux/amd64" ]]; then
  TAGS+=("$IMAGE_URL:latest")
fi

# Build tag arguments
TAG_ARGS=""
for tag in "${TAGS[@]}"; do
  TAG_ARGS="$TAG_ARGS --tag $tag"
done

# Build image
echo "Building image for $APP on $ARCH"
docker buildx build \
  --platform "$ARCH" \
  --file "dockerfiles/$APP.Dockerfile" \
  --cache-from type=gha \
  --cache-to type=gha,mode=max \
  --output type=docker \
  --label "org.opencontainers.image.created=$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
  --label "org.opencontainers.image.version=$VERSION" \
  --label "org.opencontainers.image.revision=$GIT_SHA" \
  --label "org.opencontainers.image.source=https://github.com/spiffe/spike" \
  --label "org.opencontainers.image.licenses=Apache-2.0" \
  --label "org.opencontainers.image.title=spike" \
  --label "org.opencontainers.image.description=SPIKE is a lightweight secrets store that uses SPIFFE as its identity control plane" \
  "$TAG_ARGS" \
  .

# Push images
echo "Pushing images"
for tag in "${TAGS[@]}"; do
  docker push "$tag"
done

echo "Done!"
