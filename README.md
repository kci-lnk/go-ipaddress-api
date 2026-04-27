# IP Address Location API

High-performance IP geolocation API service built with Go, supporting both IPv4 and IPv6 lookups using [ip2region](https://github.com/lionsoul2014/ip2region) database with Redis caching and rate limiting.

## Features

- **IPv4/IPv6 Support** - Full support for both address families
- **ip2region Database** - High-performance local XDB database lookup
- **Redis Caching** - 1-hour TTL cache to reduce database load
- **Rate Limiting** - Token bucket algorithm via Redis Lua scripts
- **Trust Proxy** - Extract real client IP from Cloudflare/X-Forwarded-For headers with strict IP validation
- **Structured Logging** - JSON format logs via slog for production observability
- **Request Size Limit** - Protects against large payload attacks
- **Multi-arch Build** - Compile for Linux (amd64/arm64), macOS (amd64/arm64), Windows
- **Docker Support** - Multi-stage build with docker-compose orchestration

## Quick Start

### Prerequisites

- Go 1.23+
- Redis 7+
- [Task](https://taskfile.dev/) (optional, for task automation)

### Build

```bash
# Build for current platform
task build

# Build for all platforms
task build-multi
```

### Run

```bash
# With Docker Compose (includes Redis)
task docker:compose:up

# Or run standalone (requires Redis running)
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

Configuration is managed via `configs/config.yaml` and environment variables.

### config.yaml

```yaml
server:
  host: "0.0.0.0"
  port: 30661
  mode: "release"
  read_timeout: 10      # HTTP read timeout (seconds)
  write_timeout: 10    # HTTP write timeout (seconds)
  idle_timeout: 60     # HTTP idle timeout (seconds)
  max_header_bytes: 4096

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  pool_size: 10
  dial_timeout: 5      # Redis dial timeout (seconds)
  read_timeout: 3      # Redis read timeout (seconds)
  write_timeout: 3     # Redis write timeout (seconds)

ratelimit:
  enabled: true
  requests_per_second: 100
  burst: 200

cache:
  ttl: 3600

trust_proxy:
  enabled: true
  real_ip_header: "X-Real-IP"
  real_ip_headers:
    - "X-Real-IP"
    - "X-Forwarded-For"
    - "CF-Connecting-IP"

log:
  level: "info"
  format: "json"
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `REDIS__HOST` | Redis host | `localhost` |
| `REDIS__PORT` | Redis port | `6379` |
| `REDIS__PASSWORD` | Redis password | (empty) |
| `REDIS__DB` | Redis database | `0` |
| `REDIS__DIAL_TIMEOUT` | Redis dial timeout (seconds) | `5` |
| `REDIS__READ_TIMEOUT` | Redis read timeout (seconds) | `3` |
| `REDIS__WRITE_TIMEOUT` | Redis write timeout (seconds) | `3` |
| `CACHE__TTL` | Cache TTL in seconds | `3600` |
| `RATELIMIT__ENABLED` | Enable rate limiting | `true` |
| `RATELIMIT__REQUESTS_PER_SECOND` | Requests per second | `100` |
| `RATELIMIT__BURST` | Burst size | `200` |
| `TRUST_PROXY__ENABLED` | Enable trust proxy | `true` |
| `TRUST_PROXY__REAL_IP_HEADER` | Primary real IP header | `X-Real-IP` |
| `SERVER_PORT` | API server port (docker compose) | `30661` |
| `REDIS_PORT` | Redis port (docker compose) | `6379` |

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
│   ├── config/         # Configuration loading (viper)
│   ├── handler/        # HTTP handlers
│   ├── ipdata/         # ip2region XDB database wrapper
│   ├── middleware/
│   │   ├── ratelimit.go    # Rate limiting middleware
│   │   ├── trust_proxy.go  # Real IP extraction with strict validation
│   │   └── body_size.go    # Request body size limit
│   └── ratelimit/      # Token bucket implementation
├── pkg/response/       # Unified response format
├── configs/            # Configuration files
├── scripts/           # Build and deployment scripts
├── ipdata/            # ip2region XDB database files
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
| `publish:docker` | Publish to Docker Hub |
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
# Login and publish
export DOCKER_USERNAME=yourusername
task docker:login
task publish:docker
```

## Docker

### Build Image

```bash
# Without registry
docker build -t ipaddress-api:latest .

# With build args
docker build \
  --build-arg VERSION=1.0.0 \
  --build-arg BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ') \
  -t ipaddress-api:latest .
```

### Run Container

```bash
docker run -d \
  -p 30661:30661 \
  -v $(pwd)/ipdata:/app/ipdata \
  -v $(pwd)/configs:/app/configs \
  -e REDIS__HOST=redis \
  -e REDIS__PASSWORD=yourpassword \
  --link redis \
  ipaddress-api:latest
```

### Multi-platform Build

```bash
docker buildx create --use
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t yourusername/ipaddress-api:latest \
  --push .
```

## Database Files

Place XDB database files in `ipdata/` directory:

```
ipdata/
├── base_full_v4.xdb   # IPv4 database
└── base_full_v6.xdb   # IPv6 database
```

Obtain from [ip2region releases](https://github.com/lionsoul2014/ip2region/releases) or build your own.

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
