# ClauDED

> Expose your local Claude Code through web terminal anywhere - perfect for remote access and mobile devices.

## Use Cases

**🌍 Remote Access**
- Access your Claude Code from anywhere in the world
- Work on your projects while traveling
- No need to expose your local machine directly

**📱 Mobile Devices**
- Use Claude Code on your phone or tablet
- Perfect for quick code reviews and responses
- Full terminal experience in mobile browser

## Quick Start

### 1. Deploy Server (one-time setup)

Start the server on your remote machine:

```bash
cd server
docker-compose up -d
```

The server will expose two ports:
- `8022` - for client connections
- `8088` - for browser access

### 2. Connect from Your Local Machine

```bash
# Local testing
clauded --host=localhost:8022 --session=my-session --password=mypass

# Remote server
clauded --host=your-server.com:8022 --session=my-session --password=mypass
```

### 3. Access in Browser

Open your browser and navigate to:

```
http://your-server:8088/my-session/
```

When prompted:
- **Username**: `my-session`
- **Password**: `mypass`

## Usage Examples

### Basic Remote Access

```bash
# Connect to your server
clauded --host=myserver.com:8022 --session=work --password=secure123

# Now access from any browser:
# http://myserver.com:8088/work/
```

### Mobile Device Setup

```bash
# Start session on your local machine
clauded --host=myserver.com:8022 --session=mobile --password=pass123

# Open on your phone:
# http://myserver.com:8088/mobile/
# Login with username: mobile, password: pass123
```

### Auto-generated Credentials

```bash
# Let clauded generate session ID and password
clauded --host=myserver.com:8022

# Output will show:
# ✓ Session ID: abc123
# ✓ Password: xyz789
# URL: http://myserver.com:8088/abc123/
```

### Pass Flags to Claude

```bash
clauded --host=myserver.com:8022 \
  --session=my-session \
  --password=mypass \
  --flags='--model opus'
```

### Use Different AI Tools

```bash
# Use Claude (default)
clauded --host=myserver.com:8022 --session=my-session --password=mypass

# Use OpenCode
clauded --host=myserver.com:8022 --session=my-session --password=mypass --codecmd=opencode

# Use Kimi
clauded --host=myserver.com:8022 --session=my-session --password=mypass --codecmd=kimi

# Use Gemini
clauded --host=myserver.com:8022 --session=my-session --password=mypass --codecmd=gemini
```

### Environment Variables

Create a `.env` file in your project:

```bash
ANTHROPIC_API_KEY=your_key_here
MODEL=opus
DEBUG=true
```

The `.env` file is auto-loaded when you start clauded:

```bash
clauded --host=myserver.com:8022 --session=my-session --password=mypass
```

Override with command-line args:

```bash
clauded --host=myserver.com:8022 --session=my-session \
  --password=mypass --env MODEL=sonnet
```

## Installation

### One-Line Install (Recommended)

⚠️ **Before installing, make sure you have:**
1. Installed Claude Code CLI: `npm install -g @anthropic-ai/claude-code`
2. Configured your API key: `export ANTHROPIC_API_KEY='your-key'`
   Or authenticated: `claude auth login`

**Install with one command:**

```bash
curl -fsSL https://raw.githubusercontent.com/friddle/clauded/main/install.sh | bash
```

Or specify version:

```bash
VERSION=v0.1 curl -fsSL https://raw.githubusercontent.com/friddle/clauded/main/install.sh | bash
```

See [INSTALL.md](INSTALL.md) for detailed installation instructions.

### Build from Source

**Client (Your Local Machine)**

```bash
cd cmd/client
go build -o clauded .
```

**Server (Remote Machine)**

```bash
cd server
docker-compose up -d
```

## Client Parameters

| Parameter | Short | Default | Description |
|----------|-------|---------|-------------|
| `--host` | `-h` | **Required** | Server address (host:port) |
| `--session` | `-s` | Auto-generated | Session ID for URL and auth |
| `--password` | `-p` | Empty | Password for authentication |
| `--codecmd` | - | claude | AI tool to use (claude, opencode, kimi, gemini) |
| `--flags` | `-f` | Empty | Flags to pass to codecmd |
| `--env` | `-e` | Empty | Environment variables (repeatable) |

## Multiple Sessions

Run multiple sessions simultaneously:

```bash
# Terminal 1 - for work
clauded --host=localhost:8022 --session=work --password=workpass

# Terminal 2 - for testing
clauded --host=localhost:8022 --session=test --password=testpass

# Terminal 3 - for mobile
clauded --host=localhost:8022 --session=mobile --password=mobilepass
```

## Troubleshooting

### Connection Failed

Ensure firewall allows:
- Port `8022` - client to server
- Port `8088` - browser to server

### Claude Command Not Found

ClauDED automatically finds:
- `claude` in system PATH
- `claude-code` in system PATH
- `~/.local/bin/claude-code`

If not found, install Claude Code first.

## How It Works

```
Your Local Machine           Remote Server              Browser (Any Device)
┌─────────────┐            ┌──────────────┐           ┌─────────────┐
│  Claude Code│            │              │           │             │
│             │            │  Go Server   │◄──────────│  Web Browser│
│  clauded    │───────────►│  :8088       │           │             │
│  (gotty+    │  piko      │              │           │             │
│   piko)     │  :8022     │  Piko Proxy  │           │             │
└─────────────┘            └──────────────┘           └─────────────┘
```

## License

MIT
