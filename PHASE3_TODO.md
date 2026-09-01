# PHASE3 — 事件系统 + HTTP RPC + 异步执行

## 完整 MC 协议包
- [x] Text 包（聊天、提示）
- [x] CommandRequest/Output 包（命令）
- [x] PyRpc 包（网易特有）
- [x] NeteaseJson 包（LOGIN_UID）
- [x] ClientCacheStatus 包
- [x] PlayStatus 包
- [x] WSCommand 包

## 事件系统
- [x] EventBus 事件总线（protocol/events.go）
- [x] MC EventProcessor（mc/event_processor.go）
- [x] 按包 ID 路由
- [x] 聊天事件监听
- [x] 命令输出事件监听

## HTTP RPC 服务
- [x] HTTPServer（netherite/http_server.go）
- [x] /api/v1/plugins 插件管理
- [x] /api/v1/commands 命令执行
- [x] /api/v1/events SSE 事件流
- [x] /api/v1/status 状态查询
- [x] /health 健康检查

## 异步执行
- [x] AsyncExecutor（mc/executor.go）
- [x] 命令队列（1024容量）
- [x] 多工作协程（可配置）
- [x] 自动重试（最多3次）
- [x] 速率限制（50ms间隔）
- [x] 回调钩子

## 集成测试
- [x] Raknet 帧编解码测试
- [x] ECDH 密钥交换测试
- [x] 端到端加密测试
- [x] Varint 编解码测试
- [x] 握手包编解码测试

## 主程序集成
- [x] 事件处理器自动启动
- [x] 异步执行器自动启动
- [x] HTTP RPC 服务自动启动
- [x] 完整信号处理
- [x] 优雅关闭
