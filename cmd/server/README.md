# Clauded Server

Claude Code 远程终端服务端，基于 Piko 反向代理实现。

## 快速开始

### Docker 部署

```bash
cd /path/to/server
docker-compose up -d
```

### 手动运行

```bash
# 构建
go build -o server ./cmd/server

# 运行
./server
```

## 访问地址

启动后可以通过以下方式访问：

- **本地访问**: `http://localhost:8088/{session_id}`
- **通过 Nginx**: `http://your-domain.com/{session_id}`

## Nginx 配置

Server 服务可以部署在 Nginx 后面，**只要支持 WebSocket 即可**。

```nginx
server {
    listen 80;
    server_name your-domain.com;

    # 所有路径代理到 clauded-server
    location / {
        proxy_pass http://localhost:8088;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket 支持（必需）
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        # WebSocket 超时设置
        proxy_read_timeout 86400s;
        proxy_send_timeout 86400s;
    }
}
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `LISTEN_PORT` | 8088 | HTTP 服务端口 |
| `PIKO_UPSTREAM_PORT` | 8022 | Piko upstream 端口 |
| `ENABLE_TLS` | false | 是否启用 HTTPS |
| `TLS_CERT_FILE` | - | TLS 证书路径 |
| `TLS_KEY_FILE` | - | TLS 私钥路径 |

## 端口说明

- **8022**: Piko Upstream（客户端连接）
- **8088**: HTTP API 服务
- **8023**: Piko Proxy（内部使用）

## 健康检查

```bash
curl http://localhost:8088/health
```
