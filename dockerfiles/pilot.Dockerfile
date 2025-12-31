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
    APPVERSION=$APPVERSION \
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
# Copy local SDK (required when using replace directive for development)
# COPY spike-sdk-go/ ./spike-sdk-go/
RUN go mod download

# Copy the app source code
COPY . .

# Build the app for the target architecture
RUN echo "Building SPIKE Pilot on $BUILDPLATFORM targeting $TARGETPLATFORM"
# buildx.sh requires ./app/$appName/cmd/main.go to exist.
# Here, $appName is "pilot":
RUN ./hack/docker/buildx.sh ${TARGETARCH} spike

# Use BusyBox as the base image
# While using a distroless base reduces the attack surface,
# SPIKE Pilot is an operational tool that does not actively run
# in a production network, and does not expose any outbound network
# endpoints. Therefore, including a minimal base image that has
# a shell is the right trade-off between security and convenience.
FROM busybox:1.36 AS spike
# Redefine the ARG in this stage to make it available
ARG APPVERSION

# Create necessary directories and users
RUN adduser -D -H -u 1000 spike

# Copy the binary from builder
COPY --from=builder /workspace/spike /usr/local/bin/spike

# Change ownership to spike user
RUN chown spike:spike /usr/local/bin/spike

# Ensure the binary is executable
RUN chmod +x /usr/local/bin/spike

# Apply labels to the final image
LABEL maintainers="SPIKE Maintainers <maintainers@spike.ist>" \
      version="${APPVERSION}" \
      website="https://spike.ist/" \
      repo="https://github.com/spiffe/spike" \
      documentation="https://spike.ist/" \
      contact="https://spike.ist/community/contact/" \
      community="https://spike.ist/community/hello/" \
      changelog="https://spike.ist/tracking/changelog/" \
      org.opencontainers.image.title="SPIKE Pilot" \
      org.opencontainers.image.description="SPIKE Pilot is the CLI tool for managing secrets in SPIKE." \
      org.opencontainers.image.source="https://github.com/spiffe/spike" \
      org.opencontainers.image.licenses="Apache-2.0"

ENTRYPOINT ["/usr/local/bin/spike"]
