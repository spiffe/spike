#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

FROM --platform=$BUILDPLATFORM golang:1.24.1 AS builder
ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG APPVERSION

ENV GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    APPVERSION=$APPVERSION \
    GOEXPERIMENT=boringcrypto \
    CGO_ENABLED=1

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
RUN go mod download

# Copy the app source code
COPY . .

# Build the app for the target architecture
RUN echo "Building SPIKE Pilot on $BUILDPLATFORM targeting $TARGETPLATFORM"
RUN ./dockerfiles/build.sh ${TARGETARCH} spike

# Use BusyBox as the base image
FROM busybox:1.36-musl AS spike
# Redefine the ARG in this stage to make it available
ARG APPVERSION

# Create necessary directories and users
RUN adduser -D -H -u 1000 spike

# Copy the binary from builder
COPY --from=builder /workspace/spike /spike

# Ensure the binary is executable
RUN chmod +x /spike

# Switch to non-root user
USER spike

# Apply labels to the final image
LABEL maintainers="SPIKE Maintainers <maintainers@spike.ist>" \
      version="${APPVERSION}" \
      website="https://spike.ist/" \
      repo="https://github.com/spiffe/spike" \
      documentation="https://spike.ist/" \
      contact="https://spike.ist/community/contact/" \
      community="https://spike.ist/community/hello/" \
      changelog="https://spike.ist/tracking/changelog/"

ENTRYPOINT ["/spike"]
