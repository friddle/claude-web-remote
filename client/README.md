# Gotty-Piko 客户端

一个基于终端的高效远程协助工具，集成了 gotty 和 piko 服务。专为复杂网络环境下的远程协助而设计，避免传统远程桌面对高带宽的依赖。

## 功能特性

- 🌐 基于 Web 的终端访问
- 🔐 HTTP 基本认证保护
- 🔄 自动端口分配
- ⏰ 可选的24小时自动退出
- 🖥️ 多终端支持 (zsh, bash, sh, powershell)
- 🚀 简单易用的命令行界面

## 安装

```bash
# 克隆仓库
git clone <repository-url>
cd gotty-piko-client

# 编译
make build

# 或者直接使用 go build
go build -o gottyp ./main.go
```

## 使用方法

### 基本用法

```bash
# 启用HTTP认证
gottyp --name=my-server --remote=192.168.1.100:8088 --pass=mypassword

# 不启用HTTP认证（不推荐用于生产环境）
gottyp --name=my-server --remote=192.168.1.100:8088
```

### 高级用法

```bash
# 指定终端类型
gottyp --name=local --remote=192.168.1.100:8088 --terminal=zsh --pass=localpass

# 禁用24小时自动退出
gottyp --name=server --remote=192.168.1.100:8088 --auto-exit=false --pass=serverpass

# 指定服务器端口
gottyp --name=client1 --remote=piko.example.com:8022 --pass=secret123
```

## 参数说明

| 参数 | 必需 | 说明 | 默认值 |
|------|------|------|--------|
| `--name` | ✅ | piko 客户端标识名称 | - |
| `--remote` | ✅ | 远程 piko 服务器地址 (格式: host:port) | - |
| `--pass` | ❌ | HTTP认证密码 | - |
| `--terminal` | ❌ | 指定要使用的终端类型 (zsh, bash, sh, powershell 等) | 自动检测 |
| `--auto-exit` | ❌ | 是否启用24小时自动退出 | true |
| `--server-port` | ❌ | piko 服务器端口 | 8022 |

## HTTP认证

当提供 `--pass` 参数时，系统会自动启用HTTP基本认证：

- **用户名**: 使用 `--name` 参数的值
- **密码**: 使用 `--pass` 参数的值

访问Web界面时，浏览器会弹出认证对话框，输入上述用户名和密码即可访问。

## 安全建议

1. **始终使用HTTP认证**: 在生产环境中，强烈建议使用 `--pass` 参数启用HTTP认证
2. **使用强密码**: 选择复杂且唯一的密码
3. **限制访问**: 确保只有授权用户可以访问服务器
4. **定期更换密码**: 定期更新认证密码

## 环境变量

支持通过环境变量设置参数：

```bash
export NAME="my-server"
export REMOTE="192.168.1.100:8088"
export PASS="mypassword"
export TERMINAL="zsh"
export AUTO_EXIT="true"

gottyp
```

## 故障排除

### 常见问题

1. **端口被占用**: 程序会自动查找可用端口，从8080开始
2. **认证失败**: 确保用户名和密码正确，用户名就是 `--name` 参数的值
3. **连接失败**: 检查远程服务器地址和端口是否正确

### 日志信息

程序启动时会显示以下信息：
- 客户端名称
- 远程服务器地址
- 使用的终端类型
- 本地监听端口
- HTTP认证状态（如果启用）

## 许可证

[许可证信息] 