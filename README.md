# IP Address Location API

High-performance IP geolocation API service built with Go, supporting both IPv4 and IPv6 lookups using [ip2region](https://github.com/lionsoul2014/ip2region) database with Redis caching and rate limiting.

## Features

- **IPv4/IPv6 Support** - Full support for both address families
- **ip2region Database** - High-performance local XDB database lookup (bundled in image)
- **Redis Caching** - Configurable TTL cache to reduce database load
- **Rate Limiting** - Token bucket algorithm via Redis Lua scripts
- **Trust Proxy** - Extract real client IP from Cloudflare/X-Forwarded-For headers with strict IP validation
- **Structured Logging** - JSON format logs via slog for production observability
- **Request Size Limit** - Protects against large payload attacks
- **Multi-arch Build** - Compile for Linux (amd64/arm64), macOS (amd64/arm64), Windows
- **Docker Support** - Multi-stage build with docker-compose orchestration
- **Env Var Config** - All configuration via environment variables with defaults

## Quick Start

### Prerequisites

- Go 1.23+
- Redis 7+
- [Task](https://taskfile.dev/) (optional, for task automation)
- Docker & Docker Compose

### Run with Docker Compose

```bash
# Start services (Redis + API)
docker compose up -d

# View logs
docker compose logs -f

# Stop services
docker compose down
```

### Build from Source

```bash
# Build for current platform
task build

# Build for all platforms
task build-multi
```

### Run

```bash
# Run standalone (requires Redis running)
redis-server &
task run
```

### API Usage

```bash
# Lookup IP (default port 30661)
curl "http://localhost:30661/api/v1/ip/lookup?ip=111.67.55.64"

# IPv6 lookup
curl "http://localhost:30661/api/v1/ip/lookup?ip=2408:844f:d46:ea27:35fe:19fe:d293:d4ef"

# Health check
curl "http://localhost:30661/health"
```

### Response Format

**Success:**
```json
{
  "code": 0,
  "msg": "success",
  "ip": "111.67.55.64",
  "result": {
    "version": "ipv4",
    "continent": "亚洲",
    "country": "中国",
    "province": "台湾",
    "city": "台北",
    "district": "",
    "isp": "威迈思电信",
    "country_code": "21",
    "fields": ["亚洲","中国","台湾","台北","","威迈思电信","121.565170","25.037798","710100","02","222","Asia/Shanghai","21","TWD"],
    "raw": "亚洲|中国|台湾|台北||威迈思电信|121.565170|25.037798|710100|02|222|Asia/Shanghai|TWD|21||CN"
  }
}
```

**Error:**
```json
{
  "code": 4001,
  "msg": "invalid ip address",
  "ip": "2409:8a74:",
  "result": null
}
```

## Configuration

All configuration is via environment variables. No config file required.

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER__HOST` | `0.0.0.0` | Server bind host |
| `SERVER__PORT` | `30661` | Server port |
| `SERVER__READ_TIMEOUT` | `10s` | HTTP read timeout |
| `SERVER__WRITE_TIMEOUT` | `10s` | HTTP write timeout |
| `SERVER__IDLE_TIMEOUT` | `60s` | HTTP idle timeout |
| `SERVER__MAX_HEADER_BYTES` | `4096` | Max header bytes |
| `REDIS__HOST` | `localhost` | Redis host |
| `REDIS__PORT` | `6379` | Redis port |
| `REDIS__PASSWORD` | (empty) | Redis password |
| `REDIS__DB` | `0` | Redis database |
| `REDIS__POOL_SIZE` | `10` | Redis connection pool size |
| `REDIS__DIAL_TIMEOUT` | `5s` | Redis dial timeout |
| `REDIS__READ_TIMEOUT` | `3s` | Redis read timeout |
| `REDIS__WRITE_TIMEOUT` | `3s` | Redis write timeout |
| `CACHE__TTL` | `3600` | Cache TTL in seconds |
| `RATELIMIT__ENABLED` | `true` | Enable rate limiting |
| `RATELIMIT__REQUESTS_PER_SECOND` | `100` | Requests per second |
| `RATELIMIT__BURST` | `200` | Rate limit burst size |
| `TRUST_PROXY__ENABLED` | `true` | Enable trust proxy |
| `TRUST_PROXY__REAL_IP_HEADER` | `X-Real-IP` | Real IP header |
| `LOG__LEVEL` | `info` | Log level |
| `LOG__FORMAT` | `json` | Log format (json/text) |

### Redis Key Prefix

All Redis keys use a fixed global prefix: `ipaddress:`
- Cache: `ipaddress:cache:<ip>`
- Rate limit: `ipaddress:ratelimit:<ip>`

## Project Structure

```
.
├── cmd/server/          # Application entry point
│   └── main.go         # Run function, version constants
├── internal/
│   ├── cache/          # Redis cache layer
│   ├── config/         # Configuration loading (env vars)
│   ├── handler/        # HTTP handlers
│   ├── ipdata/         # ip2region XDB database wrapper
│   ├── middleware/
│   │   ├── ratelimit.go    # Rate limiting middleware
│   │   ├── trust_proxy.go  # Real IP extraction with strict validation
│   │   └── body_size.go    # Request body size limit
│   └── ratelimit/      # Token bucket implementation
├── pkg/response/       # Unified response format
├── scripts/           # Build and deployment scripts
├── ipdata/            # ip2region XDB database files (bundled in image)
│   ├── base_full_v4.xdb
│   └── base_full_v6.xdb
├── Taskfile.yml       # Task automation
├── Dockerfile         # Multi-stage Docker build
└── docker-compose.yml # Docker orchestration with Redis
```

## Available Tasks

```bash
task --list
```

| Task | Description |
|------|-------------|
| `build` | Build binary for current platform |
| `build-multi` | Build binaries for all architectures |
| `run` | Run the server locally |
| `docker:build` | Build Docker image |
| `docker:push` | Push Docker image to registry |
| `docker:compose:up` | Start services with docker compose |
| `docker:compose:down` | Stop services |
| `publish:local` | Publish binaries via SSH |
| `publish:docker` | Publish multi-arch Docker images to Hub |
| `tidy` | Tidy go modules |

## Deployment

### Docker Compose

```bash
# Start services (Redis + API)
docker compose up -d

# View logs
docker compose logs -f

# Stop services
docker compose down
```

### Remote Server

```bash
# Publish binaries
SSH_HOST=192.168.31.135 task publish:local
```

### Docker Hub

```bash
# Publish multi-arch images (amd64 + arm64) to Docker Hub
# Version is automatically read from main.go
task publish:docker

# This will:
# 1. Read version from main.go (currently 1.0.1)
# 2. Build linux/amd64 and linux/arm64 images
# 3. Tag as kcilnk/go-ipaddress-api:<version> and kcilnk/go-ipaddress-api:latest
# 4. Create and push multi-arch manifests for both tags
```

## Docker

### Build Image

```bash
# Without registry
docker build -t ipaddress-api:latest .

# With build args
docker build \
  --build-arg VERSION=1.0.1 \
  --build-arg BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ') \
  -t ipaddress-api:latest .
```

### Run Container

```bash
docker run -d \
  -p 30661:30661 \
  -e REDIS__HOST=your-redis-host \
  -e REDIS__PASSWORD=yourpassword \
  --link redis \
  kcilnk/go-ipaddress-api:latest
```

### Multi-platform Build

```bash
# Use task (recommended)
task publish:docker

# Or manually with buildx
docker buildx create --use
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t kcilnk/go-ipaddress-api:latest \
  --push .
```

## Error Codes

| Code | Message | Description |
|------|---------|-------------|
| `0` | success | Successful lookup |
| `4001` | invalid ip address | Invalid IP format |
| `4002` | ip data not found | No data for this IP |
| `4130` | request body too large | Request exceeds size limit |
| `4290` | rate limit exceeded | Too many requests |
| `5001` | internal error | Server error |

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- [ip2region](https://github.com/lionsoul2014/ip2region) - High-performance IP geolocation database
- [Gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [go-redis](https://github.com/redis/go-redis) - Redis client
