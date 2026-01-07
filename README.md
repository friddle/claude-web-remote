# ClauDED

> Expose your local Claude Code through web terminal anywhere - perfect for remote access and mobile devices.

⚠️ **Security Notice**: The quick start uses a demo server (clauded.friddle.me) for testing. For production use, we strongly recommend deploying your own self-hosted server to ensure full control over your data and security.

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

### Option 1: Quick Demo (One-Line Install)

⚠️ **Before installing, make sure you have:**
- Installed Claude Code CLI: `npm install -g @anthropic-ai/claude-code`
- Configured your API key: `export ANTHROPIC_API_KEY='your-key'` or `claude auth login`

Install and start in one command:

```bash
curl -fsSL https://raw.githubusercontent.com/friddle/claude-web-remote/main/install.sh | bash
```

This will:
1. Download the `clauded` binary to `/usr/local/bin`
2. Connect to the demo server at `clauded.friddle.me:8022`
3. Generate a random session and password
4. Print the browser URL to access your Claude Code

### Option 2: Self-Hosted Server (Recommended)

For better security and control, deploy your own server:

**1. Deploy the server on your remote machine:**

```bash
cd server
docker-compose up -d
```

The server will expose:
- Port `8022` - for client connections
- Port `8088` - for browser access

**2. Connect your local machine:**

```bash
# Set your API key
export ANTHROPIC_API_KEY='your-key'

# Start clauded
clauded --host=your-server.com:8022 --session=my-session --password=mypass
```

**3. Access in browser:**

```
http://your-server.com:8088/my-session/
```

When prompted, enter your password.

**Client running example:**

![Run Client](pic/run_client.png)

## Usage

### Basic Connection

```bash
# Local testing
clauded --host=localhost:8022 --session=my-session --password=mypass

# Remote server
clauded --host=myserver.com:8022 --session=my-session --password=mypass

# Demo server
clauded --host=clauded.friddle.me:8022 --session=my-session --password=mypass
```

### Set API Key

```bash
# Method 1: Environment variable
export ANTHROPIC_API_KEY='your-key'
clauded --host=myserver.com:8022 --session=my-session --password=mypass

# Method 2: Pass via --env
clauded --host=myserver.com:8022 --session=my-session --password=mypass \
  --env ANTHROPIC_API_KEY='your-key'

# Method 3: Use .env file in your project directory
echo "ANTHROPIC_API_KEY=your-key" > .env
clauded --host=myserver.com:8022 --session=my-session --password=mypass
```

### Pass Flags to Claude

```bash
# Use specific model
clauded --host=myserver.com:8022 --session=my-session --password=mypass \
  --flags='--model opus'

# Multiple flags
clauded --host=myserver.com:8022 --session=my-session --password=mypass \
  --flags='--model opus --max-tokens 4096'
```

### Use Different AI Tools

```bash
# Claude (default)
clauded --host=myserver.com:8022 --session=my-session --password=mypass

# OpenCode
clauded --host=myserver.com:8022 --session=my-session --password=mypass \
  --codecmd=opencode

# Kimi
clauded --host=myserver.com:8022 --session=my-session --password=mypass \
  --codecmd=kimi

# Gemini
clauded --host=myserver.com:8022 --session=my-session --password=mypass \
  --codecmd=gemini
```

### Mobile Device Access

```bash
# Start session on your local machine
clauded --host=myserver.com:8022 --session=mobile --password=pass123

# Access on your phone at: http://myserver.com:8088/mobile/
```

### Multiple Sessions

Run multiple sessions simultaneously:

```bash
# Terminal 1 - work session
clauded --host=localhost:8022 --session=work --password=workpass

# Terminal 2 - test session
clauded --host=localhost:8022 --session=test --password=testpass

# Terminal 3 - mobile session
clauded --host=localhost:8022 --session=mobile --password=mobilepass
```

**Web interface example:**

![Web Usage](pic/web_usage.png)

## Client Parameters

| Parameter | Short | Default | Description |
|----------|-------|---------|-------------|
| `--host` | `-h` | **Required** | Server address (host:port) |
| `--session` | `-s` | Auto-generated | Session ID for URL and auth |
| `--password` | `-p` | Empty | Password for authentication |
| `--codecmd` | - | claude | AI tool to use (claude, opencode, kimi, gemini) |
| `--flags` | `-f` | Empty | Flags to pass to codecmd |
| `--env` | `-e` | Empty | Environment variables (repeatable) |

## Troubleshooting

### Connection Failed

**Access URL:** All browser access goes through `http://your-server-ip:8088/`

Ensure firewall allows:
- Port `8022` - client to server (for clauded connections)
- Port `8088` - browser to server (default, can be modified in docker-compose.yaml)

### Claude Command Not Found

ClauDED automatically finds:
- `claude` in system PATH
- `claude-code` in system PATH
- `~/.local/bin/claude-code`

If not found, install Claude Code first:
```bash
npm install -g @anthropic-ai/claude-code
```

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
