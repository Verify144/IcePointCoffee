# PHASE1 — Raknet + FB 认证 + 登录到 token

## Raknet 帧层
- [x] Raknet 包结构定义（8 位 ID 头）
- [x] UDP socket 封装
- [x] Connection Request (0x01) 发送/接收
- [x] Connection Accept (0x10) 解析
- [x] New Connection (0x13) 处理
- [x] Disconnect (0x15) 处理
- [x] Keep-Alive 机制
- [x] MTU 探测

## 可靠传输
- [x] ACK (0xC0) 帧
- [x] NAK (0xA0) 帧
- [x] 发送队列 + 重传
- [x] 接收窗口
- [x] 分片重组 (0x84/0x85/0x86/0x87)

## FB 认证 (HTTP)
- [x] /api/new 拿 secret
- [x] /api/phoenix/login 提交 token
- [x] 解析 chain_info / ip_address / identity_data

## MC 登录
- [x] ECDH P-384 密钥对生成
- [x] Login 包构造
- [x] PlayStatus 等待
- [x] PyRpc 握手序列（5 个包）
- [x] NeteaseJson (LOGIN_UID)

## 测试
- [x] UDP socket 单测
- [x] Raknet 帧解析测试
- [x] Varint 往返测试
- [x] PyRpc 编码解码测试
- [x] FB 认证参数测试
