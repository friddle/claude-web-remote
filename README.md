# ClauDED

> Expose your local Claude Code through web terminal anywhere using piko + gotty reverse proxy.

## 🌟 Features

- 🌐 **Web Terminal Access** - Access Claude Code from any web browser
- 🔐 **HTTP Basic Auth** - Secure password protection for each session
- 🔑 **Session Management** - Custom or auto-generated session IDs
- 🚀 **Easy to Use** - Simple command-line interface
- 🔒 **Secure Tunneling** - Encrypted connection via piko
- ⚙️ **Smart Detection** - Automatically detects `claude` or `claude-code` command
- 🔧 **Flag Passthrough** - Pass custom flags to Claude Code
- 🌍 **Environment Variables** - Multi-level environment variable configuration
- 📦 **.env Support** - Auto-load project environment variables
- 🏗️ **Multi-Architecture** - Support for ARM64 and AMD64 servers

## 📋 Architecture

```
Client Side                              Server Side
┌─────────────┐                       ┌─────────────────────┐
│             │                       │                     │
│  Claude Code│◄────┐                │  Go HTTP Server     │
│             │     │                │  (Port 8088)        │
└─────────────┘     │                │                     │
                    │                │  ↓                  │
┌─────────────┐     │                │  Piko Proxy         │
│   clauded   │     └──────────────►│  (Port 8023)        │
│             │ piko upstream       │                     │
│  gotty +    │ 8022                │  ↓                  │
│  piko agent │                     │  Piko Upstream       │
│             │                     │  (Port 8022)         │
└─────────────┘                     └─────────────────────┘
```

**Advantages**:
- ✅ No nginx configuration needed
- ✅ Native Go reverse proxy
- ✅ Unified process management
- ✅ Smaller container image
- ✅ Simpler deployment

## 📦 Installation

### Client Installation

Build from source:
```bash
cd client
go build -o clauded .
```

### Server Deployment

#### Using Docker Compose (Recommended)

The `server/docker-compose.yaml` comes pre-configured:
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
      # - PIKO_TOKEN=your-token-here  # Optional: add token authentication
    ports:
      - "8022:8022"  # piko upstream port (client connections)
      - "8088:8088"  # HTTP access port (browser access)
    restart: unless-stopped
```

Start the server:
```bash
cd server
docker-compose up -d
```

#### Multi-Architecture Support

- **Default** (AMD64): `friddlecopper/clauded-port-forward:latest`
- **AMD64** (Intel/AMD): `friddlecopper/clauded-port-forward:amd64`
- **ARM64** (Apple Silicon): `friddlecopper/clauded-port-forward:arm64`

Change the `image` tag in `docker-compose.yaml` to select the corresponding architecture.

## 🚀 Usage

### Basic Usage

```bash
# Connect to local server (recommended for testing)
clauded --host=localhost:8022 --session=my-session --password=mypass

# Connect to remote server
clauded --host=your-server.com:8022 --session=my-session --password=mypass

# Auto-generate session ID and password
clauded --host=localhost:8022

# Pass flags to claude
clauded --host=localhost:8022 \
  --session=my-session \
  --password=mypass \
  --flags='--model opus'

# Pass environment variables (highest priority)
clauded --host=localhost:8022 \
  --session=my-session \
  --password=mypass \
  --env API_KEY=xxx \
  --env DEBUG=true
```

### Environment Variable Configuration

ClauDED supports three levels of environment variable priority (low to high):

1. **System Environment Variables** - Existing environment variables
2. **.env File** - `.env` or `.claude.env` in project directory
3. **Command-line Arguments** - `--env` parameters (highest priority)

#### Using .env File

Create a `.env` file in your project directory:
```bash
# .env file example
ANTHROPIC_API_KEY=your_api_key_here
MODEL=opus
DEBUG=true
HTTP_PROXY=http://proxy.example.com:8080
```

The .env file is automatically loaded when starting clauded:
```bash
# .env file will be auto-loaded
clauded --host=localhost:8022 --session=my-session --password=mypass

# Command-line args override .env file
clauded --host=localhost:8022 --session=my-session --password=mypass \
  --env MODEL=sonnet  # This will override MODEL=opus in .env
```

### Accessing Web Terminal

After starting clauded, access in your browser:

```
http://your-server:8088/your-session-id/
```

If you set a password, the browser will prompt for authentication:
- **Username**: Session ID
- **Password**: Your password

**Example**:
- Session ID: `my-session`
- Password: `mypass`
- URL: `http://localhost:8088/my-session/`
- Auth: Username=`my-session`, Password=`mypass`

### Smart Command Detection

ClauDED automatically detects and uses the best available command:

1. **Priority**: `claude` command in system PATH
2. **Fallback**: `claude-code` command in system PATH
3. **Auto**: `~/.local/bin/claude-code` (auto-added to PATH)

Detection process:
```bash
🚀 Starting clauded client
✓ Using claude command from: /opt/homebrew/bin/claude
✅ Services started successfully!
```

## ⚙️ Configuration

### Server Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PIKO_UPSTREAM_PORT` | 8022 | Piko upstream listen port (for client connections) |
| `LISTEN_PORT` | 8088 | HTTP service listen port (for browser access) |
| `ENABLE_TLS` | false | Whether to enable TLS |
| `TLS_CERT_FILE` | - | TLS certificate file path |
| `TLS_KEY_FILE` | - | TLS private key file path |
| `PIKO_TOKEN` | - | Piko authentication token (optional) |

### Client Parameters

| Parameter | Short | Default | Description |
|----------|-------|---------|-------------|
| `--host` | `-h` | **Required** | Remote server address (format: host:port) |
| `--session` | `-s` | Auto-generated | Session ID |
| `--password` | `-p` | Empty | Authentication password |
| `--flags` | `-f` | Empty | Flags to pass to claude |
| `--env` | `-e` | Empty | Environment variables (can be used multiple times) |
| `--auto-exit` | - | true | Auto exit after 24 hours |
| `--insecure-skip-verify` | - | false | Skip TLS certificate verification |
| `--skip-install-check` | - | false | Skip claude installation check |

## 🔍 Port Explanation

- **8022**: Piko upstream port - used by client to connect
- **8023**: Piko proxy port - internal use (inside container)
- **8088**: HTTP service port - used by browser to access

## ❓ FAQ

### Connection Failed

Make sure your server firewall allows the following ports:
- Client needs access to: `8022` port
- Browser needs access to: `8088` port

### claude Command Not Found

ClauDED automatically checks the following locations:
- `/opt/homebrew/bin/claude` (Homebrew)
- `/usr/local/bin/claude`
- `~/.local/bin/claude-code`

If still not found, please ensure claude is properly installed.

### Multiple Sessions Support

You can run multiple clauded instances simultaneously, each with a different session ID:
```bash
# Terminal 1
clauded --host=localhost:8022 --session=session1 --password=pass1

# Terminal 2
clauded --host=localhost:8022 --session=session2 --password=pass2

# Terminal 3
clauded --host=localhost:8022 --session=session3 --password=pass3
```

## 🛠️ Development

### Build Client

```bash
cd client
go build -o clauded .
```

### Build Server Docker Image

```bash
cd server

# AMD64 (Intel/AMD) - default
docker build --platform linux/amd64 -t friddlecopper/clauded-port-forward:latest .

# ARM64 (Apple Silicon)
docker build --platform linux/arm64 -t friddlecopper/clauded-port-forward:arm64 .
```

### Project Structure

```
clauded/
├── client/                 # Client code
│   ├── main.go            # Entry point
│   ├── src/
│   │   ├── config.go      # Configuration management
│   │   ├── service.go     # Service management (gotty + piko)
│   │   ├── installer.go   # Installation detection
│   │   └── .env           # Environment variables config
│   └── clauded            # Compiled binary
├── server/                # Server code
│   ├── cmd/server/        # Server entry point
│   ├── config/            # Configuration
│   ├── handlers/          # HTTP handlers
│   ├── proxy/             # Reverse proxy
│   ├── notification/      # Notification service
│   ├── session/           # Session management
│   ├── Dockerfile         # Docker image build
│   └── docker-compose.yaml # Docker Compose config
└── README.md              # Project documentation
```

## 🤝 Contributing

Issues and Pull Requests are welcome!

## 📄 License

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

## 🔗 Related Links

- [Claude Code Official Documentation](https://claude.com/claude-code)
- [gotty Project](https://github.com/yudai/gotty)
- [piko Project](https://github.com/andydunstall/piko)

