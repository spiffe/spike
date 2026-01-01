#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

FROM --platform=$BUILDPLATFORM golang:1.25.5 AS builder
ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG APPVERSION

ENV GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    APPVERSION=$APPVERSION

WORKDIR /workspace

# Install cross-compilation tools
RUN apt-get update && apt-get install -y \
    gcc-x86-64-linux-gnu \
    g++-x86-64-linux-gnu \
    gcc-aarch64-linux-gnu \
    g++-aarch64-linux-gnu \
    libc6-dev-arm64-cross \
    libc6-dev-amd64-cross \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Download dependencies first (better layer caching)
COPY go.mod go.sum ./
# Copy local SDK (required when using replace directive for development)
# COPY spike-sdk-go/ ./spike-sdk-go/
RUN go mod download

# Copy the app source code
COPY . .

# Build the app for the target architecture
RUN echo "Building SPIKE Keeper on $BUILDPLATFORM targeting $TARGETPLATFORM"
# buildx.sh requires ./app/$appName/cmd/main.go to exist.
# Here, $appName is "keeper":
RUN ./hack/docker/buildx.sh ${TARGETARCH} keeper

# Target distroless base image for CGO_ENABLED apps
# This image includes a basic runtime environment with libc and
# other minimal dependencies
FROM gcr.io/distroless/static AS keeper
# Redefine the ARG in this stage to make it available
ARG APPVERSION

# Copy with numeric UID ownership
COPY --from=builder --chown=1000:1000 /workspace/keeper /keeper

# Run as non-root.
USER 1000

# Apply labels to the final image
LABEL maintainers="SPIKE Maintainers <maintainers@spike.ist>" \
      version="${APPVERSION}" \
      website="https://spike.ist/" \
      repo="https://github.com/spiffe/spike" \
      documentation="https://spike.ist/" \
      contact="https://spike.ist/community/contact/" \
      community="https://spike.ist/community/hello/" \
      changelog="https://spike.ist/tracking/changelog/" \
      org.opencontainers.image.title="SPIKE Keeper" \
      org.opencontainers.image.description="SPIKE Keeper stores encrypted key shares for SPIKE Nexus root key recovery." \
      org.opencontainers.image.source="https://github.com/spiffe/spike" \
      org.opencontainers.image.licenses="Apache-2.0"

ENTRYPOINT ["/keeper"]
