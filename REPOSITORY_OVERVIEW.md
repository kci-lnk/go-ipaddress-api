# IP 地址地理定位 API

高性能 IP 地理位置查询服务，支持 IPv4 和 IPv6，基于 [ip2region](https://github.com/lionsoul2014/ip2region) XDB 数据库查询，配合 Redis 缓存和流量限制。

---

## 从 Docker Hub 部署

```bash
# 拉取最新镜像
docker pull kcilnk/go-ipaddress-api

# 启动服务（需要先准备好 Redis）
docker run -d \
  --name ipaddress-api \
  --restart unless-stopped \
  -p 30661:30661 \
  -e REDIS__HOST=your-redis-host \
  -e REDIS__PASSWORD=yourpassword \
  kcilnk/go-ipaddress-api
```

---

## Docker Compose 一键部署（推荐）

### 完整配置示例

创建 `docker-compose.yml`：

```yaml
services:
  ipaddress-api:
    image: kcilnk/go-ipaddress-api
    container_name: ipaddress-api
    restart: unless-stopped
    ports:
      - "30661:30661"
    environment:
      - GIN_MODE=release
      - SERVER__HOST=0.0.0.0
      - SERVER__PORT=30661
      - SERVER__READ_TIMEOUT=10s
      - SERVER__WRITE_TIMEOUT=10s
      - SERVER__IDLE_TIMEOUT=60s
      - SERVER__MAX_HEADER_BYTES=4096
      - REDIS__HOST=redis
      - REDIS__PORT=6379
      - REDIS__PASSWORD=
      - REDIS__DB=0
      - REDIS__POOL_SIZE=10
      - REDIS__DIAL_TIMEOUT=5s
      - REDIS__READ_TIMEOUT=3s
      - REDIS__WRITE_TIMEOUT=3s
      - CACHE__TTL=3600
      - RATELIMIT__ENABLED=true
      - RATELIMIT__REQUESTS_PER_SECOND=100
      - RATELIMIT__BURST=200
      - TRUST_PROXY__ENABLED=true
      - TRUST_PROXY__REAL_IP_HEADER=X-Real-IP
      - LOG__LEVEL=info
      - LOG__FORMAT=json
    depends_on:
      redis:
        condition: service_healthy

  redis:
    image: redis:7-alpine
    container_name: ipaddress-redis
    restart: unless-stopped
    command: redis-server --appendonly yes
    environment:
      - REDIS__PASSWORD=
    volumes:
      - redis-data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 3

volumes:
  redis-data:
```

### 启动服务

```bash
# 启动所有服务
docker compose up -d

# 查看运行状态
docker compose ps

# 查看实时日志
docker compose logs -f
```

### 验证服务

```bash
# 健康检查
curl http://localhost:30661/health

# IP 查询示例
curl "http://localhost:30661/api/v1/ip/lookup?ip=111.67.55.64"
```

---

## 配置说明

所有配置通过环境变量完成，无需配置文件。

### 环境变量列表

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `SERVER__HOST` | `0.0.0.0` | 服务监听地址 |
| `SERVER__PORT` | `30661` | 服务端口 |
| `SERVER__READ_TIMEOUT` | `10s` | HTTP 读取超时 |
| `SERVER__WRITE_TIMEOUT` | `10s` | HTTP 写入超时 |
| `SERVER__IDLE_TIMEOUT` | `60s` | HTTP 空闲超时 |
| `SERVER__MAX_HEADER_BYTES` | `4096` | 最大请求头大小 |
| `REDIS__HOST` | `localhost` | Redis 主机 |
| `REDIS__PORT` | `6379` | Redis 端口 |
| `REDIS__PASSWORD` | (空) | Redis 密码 |
| `REDIS__DB` | `0` | Redis 数据库编号 |
| `REDIS__POOL_SIZE` | `10` | Redis 连接池大小 |
| `REDIS__DIAL_TIMEOUT` | `5s` | Redis 连接超时 |
| `REDIS__READ_TIMEOUT` | `3s` | Redis 读取超时 |
| `REDIS__WRITE_TIMEOUT` | `3s` | Redis 写入超时 |
| `CACHE__TTL` | `3600` | 缓存 TTL（秒） |
| `RATELIMIT__ENABLED` | `true` | 启用流量限制 |
| `RATELIMIT__REQUESTS_PER_SECOND` | `100` | 每秒请求数 |
| `RATELIMIT__BURST` | `200` | 流量限制突发值 |
| `TRUST_PROXY__ENABLED` | `true` | 启用信任代理 |
| `TRUST_PROXY__REAL_IP_HEADER` | `X-Real-IP` | 真实 IP 头 |
| `LOG__LEVEL` | `info` | 日志级别 |
| `LOG__FORMAT` | `json` | 日志格式 (json/text) |

---

## 响应示例

**成功：**
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

**错误：**
```json
{
  "code": 4001,
  "msg": "invalid ip address",
  "ip": "2409:8a74:",
  "result": null
}
```

---

## 错误码

| 代码 | 说明 |
|------|------|
| `0` | 成功 |
| `4001` | 无效的 IP 地址 |
| `4002` | 未找到该 IP 数据 |
| `4290` | 请求频率超限 |
| `4130` | 请求体过大 |
| `5001` | 服务器内部错误 |

---

## 相关链接

- Docker Hub: [kcilnk/go-ipaddress-api](https://hub.docker.com/r/kcilnk/go-ipaddress-api)
- GitHub: [kci-lnk/go-ipaddress-api](https://github.com/kci-lnk/go-ipaddress-api)
