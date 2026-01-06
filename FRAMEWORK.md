# ClaudeD 框架设计文档

## 概述

ClaudeD 是一个命令行工具,用于将本地的 Claude Code 终端会话通过 gotty 和 piko 服务转发到远程服务器,实现远程 Web 访问。

## 核心组件

### 1. Claude Code 核心
- **功能**: AI 驱动的代码助手
- **运行方式**: 作为独立的进程运行
- **每个 Session**: 对应一个独立的 Claude Code 进程实例
- **安装方式**: 通过脚本自动安装,使用标准 claude-code 安装方式

### 2. gotty (重要组件)
- **项目地址**: https://github.com/yudai/gotty
- **作用**: 将 CLI 工具转换为 Web 终端
- **核心功能**:
  - 终端复用和转发
  - WebSocket 连接管理
  - HTTP 基本认证
  - 终端会话保持
- **改造需求**:
  - 实现会话持久化,避免断开重连后丢失上下文
  - 如标准版无法满足,需通过 git submodule 源码修改实现
- **在本项目中的角色**: 作为终端到 Web 的桥梁,是整个系统的核心依赖

### 3. Go Server (统一服务)
- **语言**: Go
- **功能**:
  - HTTP/HTTPS Server (Web 界面)
  - WebSocket Server (终端连接)
  - SSE Server (实时通知推送)
  - Webhook API (订阅管理)
  - 通知服务和队列管理
  - Session 管理和认证
  - 事件监听和处理
- **端口配置**:
  - HTTP: 8088 (Web 服务端口)
  - HTTPS: 8443 (可选)
  - piko 隧道: 8022

### 4. clauded (客户端)
- **语言**: Go
- **功能**:
  - 初始化和管理 Claude Code 进程
  - 配置和管理 gotty 会话
  - 建立 piko 隧道连接
  - 任务完成检测和通知触发
  - 支持参数透传（`--flags`）
  - 支持环境变量透传（`-e`）
- **Session 管理**:
  - 每次执行启动一个新 session
  - 每个 session 独立运行一个 claude-code 进程
  - 支持多 session 并发

### 5. 通知系统 (核心功能)
- **功能**: 实时推送任务完成和状态通知
- **组件**:
  - **Go Server 通知服务**:
    - Webhook 订阅管理
    - SSE (Server-Sent Events) 推送
    - 任务事件监听
    - 通知队列管理
  - **Browser 端通知**:
    - SSE 实时连接
    - 桌面通知显示
    - 通知声音提醒
    - 通知历史记录
  - **Android 端通知**:
    - Webhook 接收服务
    - 系统通知显示
    - 离线通知缓存
    - 点击跳转 WebView

### 6. Android 客户端 (开发中)
- **技术栈**: Capacitor + WebView
- **功能**:
  - 输入连接参数 (host, session, password)
  - 连接到 gotty 终端
  - Webhook 通知接收
  - 系统通知显示
  - 提供移动端访问体验

## 系统架构

```
┌─────────────────────────────────────────────────┐
│                  用户层                          │
│  ┌──────────────┐      ┌──────────────────┐   │
│  │   浏览器      │      │   Android App    │   │
│  │ (Web终端)     │      │ (WebView + 通知)  │   │
│  └──────┬───────┘      └────────┬─────────┘   │
└─────────┼────────────────────────┼─────────────┘
          │ HTTPS                │ HTTPS
          ▼                      ▼
┌─────────────────────────────────────────────────┐
│              远程服务器 (公网)                     │
│  ┌──────────────────────────────────────────┐   │
│  │         Go Server                      │   │
│  │    (端口: 8088, 隧道: 8022)           │   │
│  │  ┌────────────────────────────────────┐  │   │
│  │  │  - HTTP/HTTPS Server             │  │   │
│  │  │  - WebSocket Server               │  │   │
│  │  │  - SSE Server                    │  │   │
│  │  │  - Webhook API                   │  │   │
│  │  │  - 通知服务                      │  │   │
│  │  │  - 任务事件监听器                 │  │   │
│  │  │  - 通知队列管理                   │  │   │
│  │  │  - Session 管理                   │  │   │
│  │  └────────────────────────────────────┘  │   │
│  └──────────────────┬───────────────────────┘   │
└─────────────────────┼────────────────────────────┘
                      │ piko 隧道
┌─────────────────────▼────────────────────────────┐
│              本地机器 (NAT 后)                     │
│  ┌──────────────────────────────────────────┐   │
│  │       clauded 进程                      │   │
│  │   ┌────────────────────────────────┐    │   │
│  │   │  - Session Manager             │    │   │
│  │   │  - 任务完成检测器              │    │   │
│  │   │  - 通知触发器                  │    │   │
│  │   └────────────────────────────────┘    │   │
│  └──────────────────┬───────────────────────┘   │
│                     │                            │
│  ┌──────────────────▼───────────────────────┐   │
│  │         gotty (会话保持)                 │   │
│  └──────────────────┬───────────────────────┘   │
│                     │                            │
│  ┌──────────────────▼───────────────────────┐   │
│  │       Claude Code 进程                    │   │
│  │        (独立实例)                         │   │
│  └──────────────────────────────────────────┘   │
└───────────────────────────────────────────────────┘
```

## 核心工作流程

### 1. 初始化流程
```bash
clauded 执行
  ↓
检测是否已安装 claude-code
  ↓
[未安装] 运行 install.sh (embedfs)
  ↓
自动检测系统类型 (macOS/Debian/Ubuntu/Alpine)
  ↓
安装 Node.js 和 npm
  ↓
使用标准方式安装 claude-code
  ↓
[安装成功] 创建新的 session (UUID 或自定义)
  ↓
启动独立的 claude-code 进程
  ↓
启动 gotty,连接到 claude-code
  ↓
建立 piko 隧道连接
  ↓
等待连接
```

**运行时自动检测机制**:
- clauded 每次启动时自动检测 claude-code 是否存在
- 如不存在,立即启动自动安装流程
- 安装过程透明,无需用户干预
- 安装完成后自动继续启动服务

### 2. 访问流程
```bash
用户浏览器访问
  ↓
[默认 host] https://clauded.friddle.me/<uuid-session>
[自定义 host] https://custom-host/<session>
  ↓
Go Server 处理请求
  ↓
gotty 终端界面 (通过 Go Server 代理)
  ↓
[有密码] HTTP Basic Auth
  ↓
连接到 claude-code 进程
  ↓
[可选] 建立 SSE 连接接收通知
  ↓
使用 Claude Code
```

### 3. 通知流程
```bash
Claude Code 执行任务
  ↓
clauded 任务完成检测器监听输出
  ↓
[检测到任务完成] 触发通知事件
  ↓
Go Server 通知服务推送
  ├─→ Browser (SSE 推送)
  │    ↓
  │  显示桌面通知
  │    ↓
  │  通知声音提醒
  │
  └─→ Android (Webhook)
       ↓
     App 接收 Webhook
       ↓
     显示系统通知
       ↓
     [用户点击] 跳转到 WebView
```

## 安全机制

### 1. 认证策略
- **默认 Host (clauded.friddle.me)**:
  - 强制要求密码认证
  - 自动生成固定 UUID session
  - 显示安全警告: "使用公开服务器有安全风险,请谨慎使用"
- **自定义 Host**:
  - Session 和密码均为可选
  - 用户自行控制安全级别

### 2. 传输安全
- HTTPS 加密传输
- piko 隧道加密
- 支持自定义证书验证

### 3. 会话隔离
- 每个 session 独立的 claude-code 进程
- 进程间完全隔离
- 支持并发多会话

## 安装和部署

### 1. 客户端安装方式

#### 方式一: 自动安装脚本
```bash
curl https://xxx.claude.com/install.sh | bash
# 或带参数
curl https://xxx.claude.com/install.sh | bash -s -- --token=xxx --url=xxx
```
- 自动识别系统 (macOS/Linux)
- 下载对应二进制文件
- 自动配置环境

#### 方式二: 本地构建
```bash
git clone <repo>
cd clauded
make build
```

### 2. 服务端部署

#### Docker Compose (推荐)
```yaml
version: "3.8"
services:
  piko:
    image: ghcr.io/friddle/gotty-piko-server:latest
    environment:
      - PIKO_UPSTREAM_PORT=8022
      - LISTEN_PORT=8088
    ports:
      - "8022:8022"
      - "8088:8088"
    restart: unless-stopped
```

#### Docker 直接运行
```bash
docker run -ti --network=host --rm \
  --name=piko-server \
  ghcr.io/friddle/gotty-piko-server
```

## 关键技术点

### 1. gotty 改造 (重点)
- **问题**: 标准 gotty 可能不支持会话持久化
- **解决方案**:
  1. 先尝试配置实现
  2. 如无法实现,使用 git submodule 引入源码修改
  3. 修改点:
     - 会话状态保存
     - 断线重连机制
     - 上下文恢复

### 2. embedfs 使用
```go
// 在 Go 中嵌入安装脚本
//go:embed scripts/install.sh
var installScript []byte
```

### 3. Session 管理
```go
type Session struct {
    ID      string
    Process *exec.Cmd
    Status  SessionStatus
    Config  SessionConfig
}
```

### 4. 多平台支持
- 检测系统: `runtime.GOOS`
- 自动下载对应二进制
- 安装脚本适配不同包管理器
  - macOS: Homebrew
  - Ubuntu/Debian: apt
  - Alpine: apk
- 运行时自动检测系统并执行对应的安装逻辑

## 目录结构

```
clauded/
├── client/                 # 客户端代码 (gottyp)
│   ├── main.go            # 主入口
│   ├── src/
│   │   ├── config.go      # 配置管理
│   │   ├── service.go     # 服务管理
│   │   ├── notifier.go    # 通知触发器
│   │   └── detector.go    # 任务完成检测器
│   ├── go.mod
│   └── Makefile
├── server/                # Go Server 统一服务
│   ├── main.go           # 服务入口
│   ├── handlers/         # HTTP 处理器
│   │   ├── websocket.go   # WebSocket 终端处理
│   │   ├── sse.go         # SSE 通知推送
│   │   ├── api.go         # Webhook API
│   │   └── proxy.go       # HTTP 代理
│   ├── notification/      # 通知服务
│   │   ├── queue.go      # 通知队列
│   │   ├── subscriber.go # 订阅管理
│   │   └── event.go      # 事件处理
│   ├── session/          # Session 管理
│   │   ├── manager.go    # Session 管理器
│   │   └── auth.go       # 认证
│   ├── config.go         # 配置
│   ├── go.mod
│   └── docker-compose.yaml
├── android_client/        # Android 客户端 (Capacitor)
│   ├── android/
│   │   └── services/
│   │       └── WebhookReceiverService.java  # Webhook 接收服务
│   ├── capacitor.config.ts
│   ├── package.json
│   └── src/
│       └── www/
│           └── js/
│               └── notification.js  # 通知管理
├── scripts/               # 安装脚本
│   └── install.sh         # 自动安装脚本 (embedfs)
├── web/                   # Browser 端代码
│   ├── notification.js    # SSE 连接和通知
│   └── ui.js             # UI 交互
├── README.md             # 项目文档
├── FRAMEWORK.md          # 本文档
└── TODO.md               # 任务清单
```

## 使用场景

1. **远程开发**: 在任何地方通过浏览器访问本地 Claude Code
2. **移动端访问**: 通过 Android 客户端在手机上使用
3. **多会话管理**: 同时运行多个 Claude Code 实例
4. **团队协作**: 分享会话进行远程协助

## 注意事项

1. **gotty 是核心依赖**: 本项目大量依赖 gotty,如会话保持需要特殊处理,考虑源码修改
2. **安全性**: 使用默认 host 时务必设置强密码
3. **资源管理**: 多 session 会占用更多资源,注意管理
4. **网络环境**: 需要稳定的网络连接,断线可能影响会话

## 测试环境

### 使用 Orb Stack 创建测试虚拟机

Orb Stack 是一个轻量级的虚拟化管理工具,适合快速创建测试环境。

#### 创建 Ubuntu 虚拟机
```bash
# 创建全新的 Ubuntu 虚拟机
orb create ubuntu

# 进入虚拟机
orb shell ubuntu

# 运行测试
./clauded --host=test.example.com --session=test-session --password=test123
```

#### 创建 Alpine 虚拟机
```bash
# 创建全新的 Alpine 虚拟机
orb create alpine

# 进入虚拟机
orb shell alpine

# 运行测试
./clauded --host=test.example.com --session=test-session --password=test123
```

#### 测试脚本
```bash
#!/bin/bash
# test_installation.sh

# 创建并测试不同系统
for distro in ubuntu alpine; do
  echo "=== Testing $distro ==="
  orb create $distro

  # 在新系统中运行 clauded
  orb shell $distro << 'EOF'
    # 下载 clauded
    curl -o clauded https://xxx.claude.com/clauded-linux-amd64
    chmod +x clauded

    # 运行,触发自动安装
    ./clauded --host=test.example.com --session=$distro-test --password=test123

    # 验证 claude-code 是否安装成功
    which claude-code || echo "Installation failed for $distro"
  EOF

  echo "=== $distro test completed ==="
  orb delete $distro
done
```

### 支持的测试平台
- ✅ macOS (包括 Orb Stack macOS 虚拟机)
- ✅ Ubuntu (最新 LTS 版本)
- ✅ Debian (稳定版本)
- ✅ Alpine Linux (最新版本)
- 🚧 Windows (待支持)

### 自动化测试
- 使用 Orb Stack 创建干净的环境
- 每次测试前删除旧虚拟机,确保环境纯净
- 自动验证安装结果
- 生成测试报告

## 通知系统设计

### 通知系统架构

通知系统采用 **Go Server 统一实现**，Go Server 直接处理所有功能：
- **HTTP/HTTPS**: 处理 Web 请求和 API
- **WebSocket**: 处理终端连接
- **SSE**: 处理实时通知推送
- **Webhook API**: 处理订阅管理
- **事件监听**: 监听和处理事件
- **队列管理**: 管理通知队列

通知系统是 ClaudeD 的核心功能之一,用于在任务完成或状态变更时实时通知用户。系统支持多种通知渠道,确保用户不会错过重要信息。

### 核心组件

#### 1. Server 端通知服务

#### Webhook API
```go
// 订阅通知
POST /api/v1/notifications/subscribe
{
  "session_id": "xxx-xxx-xxx",
  "webhook_url": "https://your-server.com/webhook",
  "events": ["task_completed", "error", "progress"]
}

// 取消订阅
DELETE /api/v1/notifications/subscribe
{
  "session_id": "xxx-xxx-xxx",
  "webhook_url": "https://your-server.com/webhook"
}

// 获取订阅列表
GET /api/v1/notifications/subscriptions?session_id=xxx
```

#### SSE (Server-Sent Events)
```go
// SSE 连接端点 (Go Server 直接处理)
GET /api/v1/notifications/stream?session_id=xxx

// SSE 事件格式
event: task_completed
data: {
  "session_id": "xxx",
  "task_name": "代码审查",
  "status": "success",
  "timestamp": "2024-01-06T10:30:00Z"
}

event: error
data: {
  "session_id": "xxx",
  "error": "编译错误",
  "details": "...",
  "timestamp": "2024-01-06T10:31:00Z"
}
```

**任务事件监听器**
- 监听 Claude Code 进程的 stdout/stderr
- 识别任务完成模式 (如 "✓", "Done", "Completed")
- 解析任务元数据
- 触发通知事件

#### 2. Browser 端通知

**SSE 连接**
```javascript
// 建立 SSE 连接
const eventSource = new EventSource(
  'https://clauded.friddle.me/api/v1/notifications/stream?session_id=xxx'
);

eventSource.addEventListener('task_completed', (event) => {
  const data = JSON.parse(event.data);
  showNotification(data);
});

eventSource.addEventListener('error', (event) => {
  const data = JSON.parse(event.data);
  showErrorNotification(data);
});
```

**桌面通知**
```javascript
function showNotification(data) {
  // 请求通知权限
  Notification.requestPermission().then(permission => {
    if (permission === 'granted') {
      new Notification('Claude Code 任务完成', {
        body: data.task_name + ' 已完成',
        icon: '/icons/claude.png',
        tag: data.session_id, // 相同 tag 会替换旧通知
        requireInteraction: true,
        timestamp: new Date(data.timestamp).getTime()
      });

      // 播放提示音
      playNotificationSound();
    }
  });
}

function playNotificationSound() {
  const audio = new Audio('/sounds/notification.mp3');
  audio.play();
}
```

#### 3. Android 端通知

**Webhook 接收服务**
```java
// WebhookReceiverService.java
public class WebhookReceiverService extends Service {
    private HttpServer server;
    private int port = 8080;

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        startWebhookServer();
        return START_STICKY;
    }

    private void startWebhookServer() {
        server = new HttpServer(port);
        server.addRoute("/webhook", this::handleWebhook);
        server.start();
    }

    private HttpResponse handleWebhook(HttpRequest request) {
        String body = request.getBody();
        NotificationData data = parseNotification(body);
        showSystemNotification(data);
        return new HttpResponse(200, "OK");
    }

    private void showSystemNotification(NotificationData data) {
        NotificationCompat.Builder builder = new NotificationCompat.Builder(this, CHANNEL_ID)
            .setSmallIcon(R.drawable.ic_notification)
            .setContentTitle("Claude Code 任务完成")
            .setContentText(data.getTaskName())
            .setAutoCancel(true)
            .setPriority(NotificationCompat.PRIORITY_HIGH);

        // 点击通知跳转到 WebView
        Intent intent = new Intent(this, MainActivity.class);
        PendingIntent pendingIntent = PendingIntent.getActivity(
            this, 0, intent, PendingIntent.FLAG_IMMUTABLE
        );
        builder.setContentIntent(pendingIntent);

        NotificationManager notificationManager =
            getSystemService(NotificationManager.class);
        notificationManager.notify(NOTIFICATION_ID, builder.build());
    }
}
```

**订阅流程**
```javascript
// 在 WebView 中注册 Webhook
async function subscribeToNotifications(webhookUrl) {
  const response = await fetch(
    'https://clauded.friddle.me/api/v1/notifications/subscribe',
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        session_id: currentSessionId,
        webhook_url: webhookUrl,
        events: ['task_completed', 'error', 'progress']
      })
    }
  );
  return response.json();
}

// 获取本地 Webhook URL
async function getLocalWebhookUrl() {
  // 通过 Capacitor 插件获取
  const { url } = await CapacitorHttp.get({
    url: 'http://localhost:8080/webhook-url'
  });
  return url;
}

// 启动时订阅
document.addEventListener('DOMContentLoaded', async () => {
  const webhookUrl = await getLocalWebhookUrl();
  await subscribeToNotifications(webhookUrl);
});
```

### 通知类型定义

```go
type NotificationType string

const (
    TaskCompleted NotificationType = "task_completed"
    Error          NotificationType = "error"
    ProgressUpdate NotificationType = "progress"
    SystemStatus   NotificationType = "system_status"
)

type Notification struct {
    ID        string                 `json:"id"`
    SessionID string                 `json:"session_id"`
    Type      NotificationType        `json:"type"`
    Data      map[string]interface{} `json:"data"`
    Timestamp time.Time              `json:"timestamp"`
}
```

### 任务完成检测机制

```go
// 任务完成检测器
type TaskDetector struct {
    patterns []string
    timeout  time.Duration
}

func (d *TaskDetector) Detect(output string) bool {
    // 检测常见任务完成模式
    completionPatterns := []string{
        "✓",        // Checkmark
        "Done",      // Done
        "Completed", // Completed
        "Finished",  // Finished
        "Success",   // Success
        "Build successful",
        "Tests passed",
    }

    for _, pattern := range completionPatterns {
        if strings.Contains(output, pattern) {
            return true
        }
    }

    return false
}

func (d *TaskDetector) DetectError(output string) bool {
    errorPatterns := []string{
        "Error:",
        "ERROR",
        "Failed",
        "Exception",
        "fatal:",
    }

    for _, pattern := range errorPatterns {
        if strings.Contains(output, pattern) {
            return true
        }
    }

    return false
}
```

### 通知队列管理

```go
// 通知队列
type NotificationQueue struct {
    queue     chan Notification
    subscribers map[string][]chan Notification
    mu        sync.RWMutex
}

func (q *NotificationQueue) Publish(sessionID string, notification Notification) {
    q.mu.RLock()
    defer q.mu.RUnlock()

    // 发送给该 session 的所有订阅者
    for _, ch := range q.subscribers[sessionID] {
        select {
        case ch <- notification:
        default:
            // 队列满,丢弃通知
            log.Printf("Notification dropped for session %s", sessionID)
        }
    }
}

func (q *NotificationQueue) Subscribe(sessionID string) chan Notification {
    q.mu.Lock()
    defer q.mu.Unlock()

    ch := make(chan Notification, 100)
    q.subscribers[sessionID] = append(q.subscribers[sessionID], ch)
    return ch
}
```

### 离线通知缓存

```java
// 离线通知缓存管理
public class NotificationCache {
    private SharedPreferences prefs;
    private static final String CACHE_KEY = "cached_notifications";

    public void cacheNotification(NotificationData data) {
        JSONArray array = getCachedNotifications();
        array.put(data.toJSON());

        // 限制缓存数量 (最多 50 条)
        if (array.length() > 50) {
            array.remove(0);
        }

        prefs.edit().putString(CACHE_KEY, array.toString()).apply();
    }

    public JSONArray getCachedNotifications() {
        String json = prefs.getString(CACHE_KEY, "[]");
        try {
            return new JSONArray(json);
        } catch (JSONException e) {
            return new JSONArray();
        }
    }

    public void clearCache() {
        prefs.edit().remove(CACHE_KEY).apply();
    }
}
```

### 通知安全机制

1. **Webhook 验证**
   - 使用 HMAC 签名验证请求来源
   - 定期轮换 webhook token

2. **Session 隔离**
   - 每个通知只发送给对应 session 的订阅者
   - 防止跨 session 通知泄露

3. **限流保护**
   - 单用户每分钟最多 10 条通知
   - 相同通知 5 分钟内只发送一次

### 通知性能优化

1. **批量处理**
   - 将多个小通知合并为一条
   - 使用延迟批处理减少网络请求

2. **去重机制**
   - 相同内容的通知只保留最新一条
   - 使用 message-id 标识

3. **优先级队列**
   - 错误通知优先级最高
   - 进度通知优先级最低

## 后续优化方向

1. 增强会话恢复能力
2. 添加会话持久化存储
3. 优化移动端体验
4. 添加监控和日志
5. 支持更多平台 (Windows)
6. 提供插件机制
7. 通知系统增强
   - 支持更多通知渠道 (邮件、短信等)
   - 智能通知聚合和摘要
   - 用户自定义通知规则

