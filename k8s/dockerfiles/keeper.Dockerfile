FROM --platform=$BUILDPLATFORM golang:1.24.1 AS builder
ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

ENV GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    GOEXPERIMENT=boringcrypto \
    CGO_ENABLED=0

WORKDIR /workspace

# Install cross-compilation tools
RUN apt-get update && apt-get install -y \
    gcc-x86-64-linux-gnu \
    g++-x86-64-linux-gnu \
    gcc-aarch64-linux-gnu \
    g++-aarch64-linux-gnu \
    libc6-dev-arm64-cross \
    libc6-dev-amd64-cross


# Download dependencies first (better layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the app source code
COPY . .

# Build the app for the target architecture
RUN echo "Building keeper on $BUILDPLATFORM targeting $TARGETPLATFORM"
RUN ./k8s/build.sh ${TARGETARCH} keeper

# Target distroless base image for CGO_ENABLED apps
# This image includes a basic runtime environment with libc and other minimal dependencies
FROM gcr.io/distroless/static AS keeper
COPY --from=builder /workspace/keeper /keeper
ENTRYPOINT ["/keeper"]