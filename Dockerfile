# One Dockerfile to rule them all
# Example:
# docker build --build-arg APP=nexus -t keeper:latest --target distroless-base .
FROM golang:1.23.3 AS base-builder
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download

FROM golang:1.23.3 AS builder
WORKDIR /workspace
COPY --from=base-builder /go/pkg/mod /go/pkg/mod
COPY . .

ARG APP
ARG TARGETARCH
ENV GOOS=linux \
    GOARCH=$TARGETARCH
RUN if [ "nexus" = "$APP" ]; then \
        export GOEXPERIMENT=boringcrypto && \
        export CGO_ENABLED=1; \
    else \
        export CGO_ENABLED=0; \
    fi

RUN echo "Building $APP for architecture: $TARGETARCH"
RUN go build -o $APP /workspace/app/$APP/cmd/main.go

FROM gcr.io/distroless/base AS distroless-base
COPY --from=builder /workspace/$APP /binary
ENTRYPOINT ["/binary"]

FROM gcr.io/distroless/static AS distroless-static
COPY --from=builder /workspace/$APP /binary
ENTRYPOINT ["/binary"]