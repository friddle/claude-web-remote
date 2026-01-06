# clauded

将本地 Claude Code 通过 piko + gotty 转发到远程服务器，实现远程 Web 访问。

## 简介

clauded 是一个命令行工具，用于将本地的 Claude Code 终端会话通过 gotty 和 piko 服务转发到远程服务器，让你可以通过 Web 浏览器在任何地方访问和使用 Claude Code。

### 核心功能

- 🌐 **Web 终端访问** - 将 Claude Code 暴露为 Web 终端
- 🔐 **密码认证** - 支持 HTTP Basic 认证保护
- 🔑 **会话管理** - 自定义或自动生成会话ID
- 🚀 **简单易用** - 开箱即用的命令行接口
- 🔒 **安全访问** - 通过 piko 加密隧道连接
- ⚙️ **智能检测** - 自动检测并使用 `claude` 或 `claude-code` 命令
- 🔧 **参数透传** - 支持自定义 Claude Code 参数
- 🌍 **环境变量** - 支持多级环境变量配置
- 📦 **.env 支持** - 自动加载项目环境变量
- 🏗️ **多架构** - 支持 ARM64 和 AMD64 服务器

## 系统架构

```
客户端                      服务器
┌─────────────┐            ┌─────────────────────┐
│             │            │                     │
│  Claude Code│◄────┐     │  Go HTTP Server     │
│             │     │     │  (端口 8088)        │
└─────────────┘     │     │                     │
                    │     │  ↓                  │
┌─────────────┐     │     │  Piko Proxy         │
│   clauded   │     └────►│  (端口 8023)        │
│             │ piko     │                     │
│  gotty +    │ upstream│  ↓                  │
│  piko agent │ 8022    │  Piko Upstream       │
│             │         │  (端口 8022)         │
└─────────────┘         └─────────────────────┘
```

**优势**：
- ✅ 无需 nginx 配置
- ✅ Go 原生反向代理
- ✅ 统一的进程管理
- ✅ 更小的容器镜像
- ✅ 更简单的部署

## 安装

### 客户端安装

从源码构建：
```bash
cd client
go build -o clauded .
```

### 服务端部署

#### 使用 Docker Compose（推荐）

在 `server/docker-compose.yaml` 中已配置好：
```yaml
version: "3.8"
services:
  clauded-server:
    image: friddlecopper/claued-server:latest
    container_name: clauded-server
    environment:
      - PIKO_UPSTREAM_PORT=8022
      - LISTEN_PORT=8088
      - ENABLE_TLS=false
      # - PIKO_TOKEN=your-token-here  # 可选：添加 token 认证
    ports:
      - "8022:8022"  # piko upstream port (客户端连接)
      - "8088:8088"  # HTTP access port (浏览器访问)
    restart: unless-stopped
```

启动服务：
```bash
cd server
docker-compose up -d
```

#### 多架构支持

- **ARM64** (Apple Silicon): `friddlecopper/claued-server:latest`
- **AMD64** (Intel/AMD): `friddlecopper/claued-server:amd64`

修改 `docker-compose.yaml` 中的 image 标签即可选择对应架构。

## 使用方法

### 基本用法

```bash
# 连接到本地服务器（推荐用于测试）
clauded --host=localhost:8022 --session=my-session --password=mypass

# 连接到远程服务器
clauded --host=your-server.com:8022 --session=my-session --password=mypass

# 自动生成会话ID和密码
clauded --host=localhost:8022

# 透传 claude 参数
clauded --host=localhost:8022 \
  --session=my-session \
  --password=mypass \
  --flags='--model opus'

# 透传环境变量（最高优先级）
clauded --host=localhost:8022 \
  --session=my-session \
  --password=mypass \
  --env API_KEY=xxx \
  --env DEBUG=true
```

### 环境变量配置

clauded 支持三级环境变量优先级（从低到高）：

1. **系统环境变量** - 已存在的环境变量
2. **.env 文件** - 项目目录中的 `.env` 或 `.claude.env`
3. **命令行参数** - `--env` 参数（最高优先级）

#### 使用 .env 文件

在项目目录创建 `.env` 文件：
```bash
# .env 文件示例
ANTHROPIC_API_KEY=your_api_key_here
MODEL=opus
DEBUG=true
HTTP_PROXY=http://proxy.example.com:8080
```

启动 clauded 时会自动加载：
```bash
# .env 文件会被自动加载
clauded --host=localhost:8022 --session=my-session --password=mypass

# 命令行参数会覆盖 .env 文件
clauded --host=localhost:8022 --session=my-session --password=mypass \
  --env MODEL=sonnet  # 会覆盖 .env 中的 MODEL=opus
```

### 访问 Web 终端

启动 clauded 后，在浏览器中访问：

```
http://your-server:8088/your-session-id/
```

如果设置了密码，浏览器会提示输入认证信息：
- **用户名**：会话ID（session）
- **密码**：你设置的密码

**示例**：
- 会话ID: `my-session`
- 密码: `mypass`
- 访问地址: `http://localhost:8088/my-session/`
- 认证信息: 用户名=`my-session`, 密码=`mypass`

### 智能命令检测

clauded 会自动检测并使用最佳命令：

1. **优先**：系统 PATH 中的 `claude` 命令
2. **降级**：系统 PATH 中的 `claude-code` 命令
3. **自动**：`~/.local/bin/claude-code`（自动添加到 PATH）

检测过程：
```bash
🚀 Starting clauded client
✓ Using claude command from: /opt/homebrew/bin/claude
✅ Services started successfully!
```

## 配置说明

### 服务端环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `PIKO_UPSTREAM_PORT` | 8022 | Piko upstream 监听端口（客户端连接） |
| `LISTEN_PORT` | 8088 | HTTP 服务监听端口（浏览器访问） |
| `ENABLE_TLS` | false | 是否启用 TLS |
| `TLS_CERT_FILE` | - | TLS 证书文件路径 |
| `TLS_KEY_FILE` | - | TLS 私钥文件路径 |
| `PIKO_TOKEN` | - | Piko 认证 token（可选） |

### 客户端参数

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--host` | `-h` | 必填 | 远程服务器地址（格式：host:port） |
| `--session` | `-s` | 自动生成 | 会话ID |
| `--password` | `-p` | 空 | 认证密码 |
| `--flags` | `-f` | 空 | 透传给 claude 的参数 |
| `--env` | `-e` | 空 | 环境变量（可多次使用） |
| `--auto-exit` | - | true | 24小时后自动退出 |
| `--insecure-skip-verify` | - | false | 跳过 TLS 证书验证 |
| `--skip-install-check` | - | false | 跳过 claude 安装检查 |

## 常见问题

### 1. 端口说明

- **8022**: Piko upstream 端口，客户端连接使用
- **8023**: Piko proxy 端口，内部使用（容器内部）
- **8088**: HTTP 服务端口，浏览器访问使用

### 2. 连接失败

确保服务器防火墙开放以下端口：
- 客户端需要访问：`8022` 端口
- 浏览器需要访问：`8088` 端口

### 3. claude 命令未找到

clauded 会自动检测以下位置：
- `/opt/homebrew/bin/claude` (Homebrew)
- `/usr/local/bin/claude`
- `~/.local/bin/claude-code`

如果仍未找到，请确保 claude 已正确安装。

### 4. 多会话支持

可以同时运行多个 clauded 实例，每个实例使用不同的 session ID：
```bash
# 终端 1
clauded --host=localhost:8022 --session=session1 --password=pass1

# 终端 2
clauded --host=localhost:8022 --session=session2 --password=pass2

# 终端 3
clauded --host=localhost:8022 --session=session3 --password=pass3
```

## 开发

### 构建客户端

```bash
cd client
go build -o clauded .
```

### 构建服务端 Docker 镜像

```bash
cd server

# ARM64 (Apple Silicon)
docker build --platform linux/arm64 -t friddlecopper/claued-server:latest .

# AMD64 (Intel/AMD)
docker build --platform linux/amd64 -t friddlecopper/claued-server:amd64 .
```

### 项目结构

```
clauded/
├── client/                 # 客户端代码
│   ├── main.go            # 入口文件
│   ├── src/
│   │   ├── config.go      # 配置管理
│   │   ├── service.go     # 服务管理（gotty + piko）
│   │   ├── installer.go   # 安装检测
│   │   └── .env           # 环境变量配置
│   └── clauded            # 编译后的可执行文件
├── server/                # 服务端代码
│   ├── cmd/server/        # 服务器入口
│   ├── config/            # 配置管理
│   ├── handlers/          # HTTP 处理器
│   ├── proxy/             # 反向代理
│   ├── notification/      # 通知服务
│   ├── session/           # 会话管理
│   ├── Dockerfile         # Docker 镜像构建
│   └── docker-compose.yaml # Docker Compose 配置
└── README.md              # 项目文档
```

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

[待添加许可证信息]

## 相关链接

- [Claude Code 官方文档](https://claude.com/claude-code)
- [gotty 项目](https://github.com/yudai/gotty)
- [piko 项目](https://github.com/andydunstall/piko)
