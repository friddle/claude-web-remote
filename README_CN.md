# ClauDED

> 通过 piko + gotty 反向代理，在任何地方通过 Web 终端访问本地 Claude Code。

## 🌟 特性

- 🌐 **Web 终端访问** - 从任何浏览器访问 Claude Code
- 🔐 **HTTP 基本认证** - 每个会话的安全密码保护
- 🔑 **会话管理** - 自定义或自动生成的会话 ID
- 🚀 **易于使用** - 简单的命令行界面
- 🔒 **安全隧道** - 通过 piko 加密连接
- ⚙️ **智能检测** - 自动检测 `claude` 或 `claude-code` 命令
- 🔧 **标志传递** - 向 Claude Code 传递自定义标志
- 🌍 **环境变量** - 多级环境变量配置
- 📦 **.env 支持** - 自动加载项目环境变量
- 🏗️ **多架构支持** - 支持 ARM64 和 AMD64 服务器

## 📋 架构

```
客户端                              服务器端
┌─────────────┐                       ┌─────────────────────┐
│             │                       │                     │
│  Claude Code│◄────┐                │  Go HTTP 服务器     │
│             │     │                │  (端口 8088)        │
└─────────────┘     │                │                     │
                    │                │  ↓                  │
┌─────────────┐     │                │  Piko 代理          │
│   clauded   │     └──────────────►│  (端口 8023)        │
│             │ piko upstream       │                     │
│  gotty +    │ 8022                │  ↓                  │
│  piko agent │                     │  Piko Upstream       │
│             │                     │  (端口 8022)         │
└─────────────┘                     └─────────────────────┘
```

**优势**:
- ✅ 无需 nginx 配置
- ✅ 原生 Go 反向代理
- ✅ 统一的进程管理
- ✅ 更小的容器镜像
- ✅ 更简单的部署

## 📦 安装

### 客户端安装

从源代码构建:
```bash
cd client
go build -o clauded .
```

### 服务器部署

#### 使用 Docker Compose（推荐）

`server/docker-compose.yaml` 已预配置:
```yaml
version: "3.8"
services:
  clauded-port-forward:
    image: friddlecopper/clauded-port-forward:latest
    container_name: clauded-port-forward
    environment:
      - PIKO_UPSTREAM_PORT=8022
      - LISTEN_PORT=8088
      - ENABLE_TLS=false
      # - PIKO_TOKEN=your-token-here  # 可选：添加令牌认证
    ports:
      - "8022:8022"  # piko upstream 端口（客户端连接）
      - "8088:8088"  # HTTP 访问端口（浏览器访问）
    restart: unless-stopped
```

启动服务器:
```bash
cd server
docker-compose up -d
```

#### 多架构支持

- **默认** (AMD64): `friddlecopper/clauded-port-forward:latest`
- **AMD64** (Intel/AMD): `friddlecopper/clauded-port-forward:amd64`
- **ARM64** (Apple Silicon): `friddlecopper/clauded-port-forward:arm64`

在 `docker-compose.yaml` 中修改 `image` 标签以选择相应的架构。

## 🚀 使用

### 基本用法

```bash
# 连接到本地服务器（推荐用于测试）
clauded --host=localhost:8022 --session=my-session --password=mypass

# 连接到远程服务器
clauded --host=your-server.com:8022 --session=my-session --password=mypass

# 自动生成会话 ID 和密码
clauded --host=localhost:8022

# 向 claude 传递标志
clauded --host=localhost:8022 \
  --session=my-session \
  --password=mypass \
  --flags='--model opus'

# 传递环境变量（最高优先级）
clauded --host=localhost:8022 \
  --session=my-session \
  --password=mypass \
  --env API_KEY=xxx \
  --env DEBUG=true
```

### 环境变量配置

ClauDED 支持三个级别的环境变量优先级（从低到高）:

1. **系统环境变量** - 现有环境变量
2. **.env 文件** - 项目目录中的 `.env` 或 `.claude.env`
3. **命令行参数** - `--env` 参数（最高优先级）

#### 使用 .env 文件

在项目目录中创建 `.env` 文件:
```bash
# .env 文件示例
ANTHROPIC_API_KEY=your_api_key_here
MODEL=opus
DEBUG=true
HTTP_PROXY=http://proxy.example.com:8080
```

.env 文件在启动 clauded 时会自动加载:
```bash
# .env 文件会自动加载
clauded --host=localhost:8022 --session=my-session --password=mypass

# 命令行参数会覆盖 .env 文件
clauded --host=localhost:8022 --session=my-session --password=mypass \
  --env MODEL=sonnet  # 这将覆盖 .env 中的 MODEL=opus
```

### 访问 Web 终端

启动 clauded 后，在浏览器中访问:

```
http://your-server:8088/your-session-id/
```

如果设置了密码，浏览器会提示进行身份验证:
- **用户名**: 会话 ID
- **密码**: 您的密码

**示例**:
- 会话 ID: `my-session`
- 密码: `mypass`
- URL: `http://localhost:8088/my-session/`
- 认证: 用户名=`my-session`, 密码=`mypass`

### 智能命令检测

ClauDED 自动检测并使用最佳可用命令:

1. **优先级**: 系统 PATH 中的 `claude` 命令
2. **回退**: 系统 PATH 中的 `claude-code` 命令
3. **自动**: `~/.local/bin/claude-code`（自动添加到 PATH）

检测过程:
```bash
🚀 Starting clauded client
✓ Using claude command from: /opt/homebrew/bin/claude
✅ Services started successfully!
```

## ⚙️ 配置

### 服务器环境变量

| 变量 | 默认值 | 描述 |
|----------|---------|-------------|
| `PIKO_UPSTREAM_PORT` | 8022 | Piko upstream 监听端口（用于客户端连接） |
| `LISTEN_PORT` | 8088 | HTTP 服务监听端口（用于浏览器访问） |
| `ENABLE_TLS` | false | 是否启用 TLS |
| `TLS_CERT_FILE` | - | TLS 证书文件路径 |
| `TLS_KEY_FILE` | - | TLS 私钥文件路径 |
| `PIKO_TOKEN` | - | Piko 认证令牌（可选） |

### 客户端参数

| 参数 | 简写 | 默认值 | 描述 |
|----------|-------|---------|-------------|
| `--host` | `-h` | **必需** | 远程服务器地址（格式: host:port） |
| `--session` | `-s` | 自动生成 | 会话 ID |
| `--password` | `-p` | 空 | 认证密码 |
| `--flags` | `-f` | 空 | 传递给 claude 的标志 |
| `--env` | `-e` | 空 | 环境变量（可多次使用） |
| `--auto-exit` | - | true | 24 小时后自动退出 |
| `--insecure-skip-verify` | - | false | 跳过 TLS 证书验证 |
| `--skip-install-check` | - | false | 跳过 claude 安装检查 |

## 🔍 端口说明

- **8022**: Piko upstream 端口 - 用于客户端连接
- **8023**: Piko 代理端口 - 内部使用（容器内）
- **8088**: HTTP 服务端口 - 用于浏览器访问

## ❓ 常见问题

### 连接失败

确保服务器防火墙允许以下端口:
- 客户端需要访问: `8022` 端口
- 浏览器需要访问: `8088` 端口

### 找不到 claude 命令

ClauDED 自动检查以下位置:
- `/opt/homebrew/bin/claude` (Homebrew)
- `/usr/local/bin/claude`
- `~/.local/bin/claude-code`

如果仍然找不到，请确保 claude 已正确安装。

### 多会话支持

您可以同时运行多个 clauded 实例，每个实例使用不同的会话 ID:
```bash
# 终端 1
clauded --host=localhost:8022 --session=session1 --password=pass1

# 终端 2
clauded --host=localhost:8022 --session=session2 --password=pass2

# 终端 3
clauded --host=localhost:8022 --session=session3 --password=pass3
```

## 🛠️ 开发

### 构建客户端

```bash
cd client
go build -o clauded .
```

### 构建服务器 Docker 镜像

```bash
cd server

# AMD64 (Intel/AMD) - 默认
docker build --platform linux/amd64 -t friddlecopper/clauded-port-forward:latest .

# ARM64 (Apple Silicon)
docker build --platform linux/arm64 -t friddlecopper/clauded-port-forward:arm64 .
```

### 项目结构

```
clauded/
├── client/                 # 客户端代码
│   ├── main.go            # 入口点
│   ├── src/
│   │   ├── config.go      # 配置管理
│   │   ├── service.go     # 服务管理（gotty + piko）
│   │   ├── installer.go   # 安装检测
│   │   └── .env           # 环境变量配置
│   └── clauded            # 编译后的二进制文件
├── server/                # 服务器代码
│   ├── cmd/server/        # 服务器入口点
│   ├── config/            # 配置
│   ├── handlers/          # HTTP 处理器
│   ├── proxy/             # 反向代理
│   ├── notification/      # 通知服务
│   ├── session/           # 会话管理
│   ├── Dockerfile         # Docker 镜像构建
│   └── docker-compose.yaml # Docker Compose 配置
└── README.md              # 项目文档
```

## 🤝 贡献

欢迎提交问题和拉取请求！

## 📄 许可证

MIT License

Copyright (c) 2025 ClauDED

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

## 🔗 相关链接

- [Claude Code 官方文档](https://claude.com/claude-code)
- [gotty 项目](https://github.com/yudai/gotty)
- [piko 项目](https://github.com/andydunstall/piko)

