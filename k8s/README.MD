# SPIKE Kubernetes Resources

This directory contains container and deployment-related files for the SPIKE project.

## Directory Structure

- `./dockerfiles/` - Contains Dockerfiles for all SPIKE components
- `./build.sh` - Build script used by Dockerfiles for cross-compilation

## Building Container Images

### Basic Build

To build a SPIKE component container image:

```bash
# General syntax
docker build -t <component-name>:tag -f kubernetes/dockerfiles/<component-name>.Dockerfile .

# Examples
docker build -t keeper:latest -f kubernetes/dockerfiles/keeper.Dockerfile .
docker build -t nexus:latest -f kubernetes/dockerfiles/nexus.Dockerfile .
docker build -t spike:latest -f kubernetes/dockerfiles/spike.Dockerfile .
```

### Multi-Architecture Builds with Docker Buildx

For building multi-architecture images (e.g., for both amd64 and arm64):

```bash
# Create a new builder instance if you haven't already
docker buildx create --name spike-builder --use

# Build and push multi-arch image
docker buildx build --platform linux/amd64,linux/arm64 \
  -t <registry>/<component-name>:tag \
  -f kubernetes/dockerfiles/<component-name>.Dockerfile \
  --push .

# Example for building and pushing to Docker Hub
docker buildx build --platform linux/amd64,linux/arm64 \
  -t yourusername/spike-keeper:latest \
  -f kubernetes/dockerfiles/keeper.Dockerfile \
  --push .
```

## Running Containers

```bash
# Run with debug logging
docker run --rm -e SPIKE_SYSTEM_LOG_LEVEL=DEBUG keeper:latest

# Run with mounted configuration
docker run --rm -v /path/to/config:/config nexus:latest
```

## Kubernetes Deployment

Sample Kubernetes manifests for deploying SPIKE components will be added in future releases.

## Notes

- All Dockerfiles use distroless base images for minimal attack surface
- The keeper and spike components use the static distroless image
- The nexus component uses the base distroless image due to CGO dependencies
