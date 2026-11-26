# Build stage
FROM --platform=$BUILDPLATFORM golang:1.26-bookworm AS builder

ARG TARGETOS
ARG TARGETARCH
ARG VERSION
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown
ARG GOPROXY=https://proxy.golang.org,direct
ARG HTTP_PROXY=""
ARG HTTPS_PROXY=""

WORKDIR /app

# Unset proxy vars if empty, otherwise set them
RUN if [ -n "$HTTP_PROXY" ]; then \
      export HTTP_PROXY=$HTTP_PROXY HTTPS_PROXY=$HTTPS_PROXY; \
    fi

# Copy go mod files
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    GOPROXY=${GOPROXY} GOTOOLCHAIN=auto go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg

# Build binary with version info
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GOTOOLCHAIN=auto go build \
    -ldflags="-s -w ${VERSION:+-X main.Version=$VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o /ipaddress-api ./cmd/server

# Runtime stage
FROM alpine:3.23

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user
RUN adduser -D -u 1000 appuser \
    && mkdir -p /app/cache \
    && chown appuser:appuser /app /app/cache

# Copy binary from builder
COPY --from=builder --chown=appuser:appuser /ipaddress-api /app/ipaddress-api

# Copy ipdata (bundled in image)
COPY --chown=appuser:appuser ipdata /app/ipdata

USER appuser

EXPOSE 30661

ENV GIN_MODE=release

ENTRYPOINT ["/app/ipaddress-api"]
