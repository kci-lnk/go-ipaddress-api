# Build stage
FROM --platform=$TARGETOS/$TARGETARCH golang:1.23-bookworm AS builder

ARG VERSION=unknown
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

# Copy source code
COPY . .

# Build binary with version info
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH GOTOOLCHAIN=auto go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o /ipaddress-api ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user
RUN adduser -D -u 1000 appuser

# Copy binary from builder
COPY --from=builder /ipaddress-api /app/ipaddress-api

# Copy ipdata (bundled in image)
COPY ipdata /app/ipdata

# Create cache directory
RUN mkdir -p /app/cache && chown -R appuser:appuser /app

USER appuser

EXPOSE 30661

ENV GIN_MODE=release

ENTRYPOINT ["/app/ipaddress-api"]
