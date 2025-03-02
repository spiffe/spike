# One Dockerfile to rule them all
FROM --platform=$BUILDPLATFORM golang:1.23.3 AS builder

ARG APP
ARG CGOENABLED
ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

ENV GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    GOEXPERIMENT=boringcrypto \
    CGO_ENABLED=$CGOENABLED

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
RUN echo "Building $APP on $BUILDPLATFORM targeting $TARGETPLATFORM"
RUN if [ "$TARGETARCH" = "amd64" ]; then \
        CC=x86_64-linux-gnu-gcc go build -o $APP /workspace/app/$APP/cmd/main.go; \
    elif [ "$TARGETARCH" = "arm64" ]; then \
        CC=aarch64-linux-gnu-gcc go build -o $APP /workspace/app/$APP/cmd/main.go; \
    fi


# Target distroless base image for CGO_ENABLED apps
# This image includes a basic runtime environment with libc and other minimal dependencies
FROM gcr.io/distroless/base AS distroless-base
COPY --from=builder /workspace/$APP /binary
ENTRYPOINT ["/binary"]

# Target distroless static image for CGO_DISABLED apps
# This image includes a static binary and no runtime dependencies
FROM gcr.io/distroless/static AS distroless-static
COPY --from=builder /workspace/$APP /binary
ENTRYPOINT ["/binary"]