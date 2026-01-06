# TODO 任务清单

## 功能开发

### 1. Claude Code 自动安装
- [ ] 创建安装脚本 (install.sh),使用 embedfs 嵌入到 Go 项目
- [ ] 脚本自动检测系统类型 (macOS/Debian/Ubuntu/Alpine)
- [ ] 自动安装 Node.js 和 npm 依赖
- [ ] 使用标准方式安装 claude-code
- [ ] **运行时自动检测**: clauded 启动时检测 claude-code 是否存在
  - [ ] 不存在则自动运行安装脚本
  - [ ] 安装完成后自动启动 claude-code
- [ ] 非支持的系统提示用户手动安装

### 2. 一键安装脚本
- [ ] 创建 curl xxx | bash 安装脚本
- [ ] 脚本自动识别操作系统 (macOS/Linux)
- [ ] 下载对应平台的 clauded 二进制文件
- [ ] 支持通过参数指定 token 和 url
  - `--token=xxx`
  - `--url=xxx`

### 3. 客户端改造 (gottyp)
- [ ] 按 README.md 使用方式重构客户端
- [ ] 实现 gotty 会话保持功能
- [ ] 如无法实现,考虑 git submodule 源码修改
- [ ] 确保每个 session 对应一个独立的 claude-code 进程
- [ ] 支持在程序目录执行,每次执行创建新 session

### 4. 安全认证机制
- [ ] 使用默认 host 时:
  - [ ] 自动生成固定 UUID 作为 session ID
  - [ ] 不允许用户自定义 session
  - [ ] 强制要求设置密码
  - [ ] 显示安全风险警告
- [ ] 使用自定义 host 时:
  - [ ] 允许自定义 session
  - [ ] 密码为可选
- [ ] 更新 README.md 说明上述安全策略

### 5. Android 客户端
- [ ] 创建 capacitor 项目
- [ ] 实现 WebView 界面
- [ ] 输入表单: host, session, password
- [ ] 连接到 gotty 终端
- [ ] 实现 webhook 接收功能
  - [ ] 注册接收 webhook 通知
  - [ ] 显示系统通知
  - [ ] 点击通知跳转到 WebView
- [ ] 打包为 Android APK/AAB

### 6. Session 管理
- [ ] 每个 clauded 执行实例对应一个 session
- [ ] 每个 session 独立运行一个 claude-code 进程
- [ ] 支持多 session 并发运行
- [ ] 进程隔离和资源管理

### 7. 通知系统
- [ ] **Server 端通知服务 (Go 实现)**:
  - [ ] 创建 notification 服务
  - [ ] 实现 SSE Server (支持长连接)
  - [ ] 实现 Webhook API (订阅/取消订阅)
  - [ ] 任务完成事件监听器
  - [ ] 通知消息队列管理 (使用 channel 或 Redis)
  - [ ] Session 订阅管理
  - [ ] WebSocket 支持 (终端连接)
  - [ ] HTTP 代理支持

- [ ] **Browser 端自动提醒**:
  - [ ] 实现 SSE 连接接收实时通知
  - [ ] 显示浏览器桌面通知
  - [ ] 通知声音提醒
  - [ ] 通知历史记录
  - [ ] 通知设置管理

- [ ] **Android 端 Webhook 推送**:
  - [ ] 实现本地 webhook 接收服务器
  - [ ] 订阅 server 端 webhook 通知
  - [ ] 接收通知并显示系统通知
  - [ ] 点击通知打开 WebView
  - [ ] 通知缓存和离线处理

- [ ] **Claude Code 任务完成检测**:
  - [ ] 监听 claude-code 进程输出
  - [ ] 识别任务完成模式
  - [ ] 触发通知事件

- [ ] **通知类型定义**:
  - [ ] 任务完成通知
  - [ ] 错误通知
  - [ ] 进度更新通知
  - [ ] 系统状态通知

## 文档更新

- [ ] 更新主 README.md:
  - [ ] 添加默认 host 安全说明
  - [ ] 添加密码强制要求说明
  - [ ] 添加通知系统说明
  - [ ] 更新使用示例
- [ ] 更新客户端 README.md:
  - [ ] 说明 session 管理机制
  - [ ] 添加安全最佳实践
  - [ ] 添加通知系统使用说明
- [ ] 创建 FRAMEWORK.md 文档
- [ ] 创建 API 文档 (通知系统 API)

## 测试

### 测试环境准备
- [ ] 使用 Orb Stack 创建测试虚拟机
  - [ ] 创建全新的 Ubuntu 虚拟机
  - [ ] 创建全新的 Alpine 虚拟机
  - [ ] 创建全新的 macOS 虚拟机 (如支持)

### 功能测试
- [ ] 测试 macOS 自动安装脚本
- [ ] 测试 Ubuntu 自动安装脚本
- [ ] 测试 Debian 自动安装脚本
- [ ] 测试 Alpine 自动安装脚本
- [ ] 测试运行时自动检测和安装功能
  - [ ] 在全新系统运行 clauded,验证自动安装
  - [ ] 安装后验证 claude-code 正常启动
- [ ] 测试一键安装脚本
- [ ] 测试 session 保持功能
- [ ] 测试安全认证机制
- [ ] 测试 Android 客户端

### 通知系统测试
- [ ] 测试 Server 端 webhook API
- [ ] 测试 Browser 端 SSE 连接和通知
  - [ ] 测试实时通知接收
  - [ ] 测试桌面通知显示
  - [ ] 测试通知声音
- [ ] 测试 Android 端 webhook 接收
  - [ ] 测试通知订阅
  - [ ] 测试系统通知显示
  - [ ] 测试点击跳转
  - [ ] 测试离线通知缓存
- [ ] 测试任务完成检测
  - [ ] 测试不同任务的完成识别
  - [ ] 测试通知触发准确性
- [ ] 测试多种通知类型
  - [ ] 任务完成通知
  - [ ] 错误通知
  - [ ] 进度更新

### 测试脚本
- [ ] 创建 Orb 测试环境准备脚本
- [ ] 创建自动化测试脚本
- [ ] 创建通知系统测试脚本

## 优化

- [ ] 改进错误提示信息
- [ ] 添加详细日志
- [ ] 优化安装体验
- [ ] 性能优化
- [ ] 通知系统性能优化
  - [ ] 批量通知处理
  - [ ] 通知去重
  - [ ] 通知限流

## 运维与部署

- [ ] **服务端镜像更新**:
  - [ ] SSH 连接到 `clauded.friddle.me`
  - [ ] 进入 `clauded` 目录
  - [ ] 使用 `docker-compose` 更新服务 (pull & up -d)
  - [ ] 验证远程服务器镜像更新和脚本测试