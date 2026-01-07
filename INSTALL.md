# Clauded 安装脚本

一键安装 clauded 工具，支持自动检测系统环境并下载对应版本。

## 快速安装

### 使用默认版本 (v0.1)

```bash
curl -fsSL https://raw.githubusercontent.com/friddle/claude-web-remote/main/install.sh | bash
```

### 指定版本

```bash
VERSION=v0.2 curl -fsSL https://raw.githubusercontent.com/friddle/claude-web-remote/main/install.sh | bash
```

### 下载后执行

```bash
# 下载
wget https://raw.githubusercontent.com/friddle/claude-web-remote/main/install.sh
# 或
curl -O https://raw.githubusercontent.com/friddle/claude-web-remote/main/install.sh

# 执行
chmod +x install.sh
./install.sh
```

## 环境要求

⚠️ **安装前请确保已配置好 Claude Code 环境**

### 1. 安装 Claude Code CLI

```bash
npm install -g @anthropic-ai/claude-code
```

### 2. 配置 API 密钥

**方式一：使用 Anthropic 官方 API**

```bash
export ANTHROPIC_API_KEY='your-api-key-here'
```

**方式二：使用自定义 API 端点**

```bash
export ANTHROPIC_BASE_URL='https://your-endpoint.com'
export ANTHROPIC_AUTH_TOKEN='your-token'
```

**方式三：使用 claude auth 登录**

```bash
claude auth login
```

## 使用方法

### 基本用法

```bash
clauded --session mysession --remote localhost:8022
```

### 完整参数

```bash
clauded \
  --session mysession \
  --remote your-server.com:8022 \
  --password mypassword \
  --auto-exit=false
```

### 参数说明

| 参数 | 说明 | 示例 |
|------|------|------|
| `--session` | 会话名称 | `--session friddle` |
| `--remote` | 远程服务器地址 | `--remote localhost:8022` |
| `--password` | 认证密码（可选） | `--password secret` |
| `--auto-exit` | 24小时自动退出（默认 true） | `--auto-exit=false` |
| `--daemon` | 后台运行（默认 true） | `--daemon=false` |

## 会话管理

### 查看运行中的会话

```bash
clauded session list
```

### 停止会话

```bash
clauded session kill <session-id>
```

## 故障排除

### 权限问题

如果安装到 `/usr/local/bin` 失败，脚本会自动安装到 `~/.local/bin`。

确保该目录在 PATH 中：

```bash
export PATH="$PATH:$HOME/.local/bin"
```

### 无法连接服务器

1. 检查服务器地址是否正确
2. 确认服务器端口是否开放
3. 查看防火墙设置

### Claude Code 无法运行

1. 确认 Claude Code 已安装：`claude --version`
2. 检查环境变量：`env | grep ANTHROPIC`
3. 重新认证：`claude auth login`

## 高级用法

### 使用环境变量配置

```bash
export CLAUDED_SESSION="mysession"
export CLAUDED_REMOTE="localhost:8022"
export CLAUDED_PASSWORD="mypass"

clauded
```

### 后台运行

```bash
# 启动后台会话
clauded --session mysession --remote localhost:8022 --daemon=true

# 查看日志
tail -f ~/.clauded.log
```

### 多会话管理

```bash
# 启动多个会话
clauded --session work --remote localhost:8022
clauded --session personal --remote localhost:8022

# 列出所有会话
clauded session list

# 停止特定会话
clauded session kill work
```

## 更新到最新版本

```bash
VERSION=latest curl -fsSL https://raw.githubusercontent.com/friddle/claude-web-remote/main/install.sh | bash
```

## 卸载

```bash
# 删除二进制文件
sudo rm /usr/local/bin/clauded
# 或
rm ~/.local/bin/clauded

# 删除配置和数据
rm -rf ~/.clauded
```
