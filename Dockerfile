# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with version info
ARG VERSION=unknown
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o /ipaddress-api ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -u 1000 appuser

# Copy binary from builder
COPY --from=builder /ipaddress-api /app/ipaddress-api

# Copy ipdata
COPY ipdata /app/ipdata

# Create cache directory
RUN mkdir -p /app/cache && chown -R appuser:appuser /app

USER appuser

EXPOSE 30661

ENV GIN_MODE=release
ENV CONFIG_PATH=/app/configs/config.yaml

ENTRYPOINT ["/app/ipaddress-api"]
CMD ["-config", "/app/configs/config.yaml"]
