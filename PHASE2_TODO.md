# PHASE2 — 加密连接层

## 加密握手
- [x] ECDH X25519 密钥对生成
- [x] ECDH P-384 备用（网易）
- [x] 共享密钥派生（Bedrock SHA-256 双 hash）
- [x] 网易专用密钥派生（Challenge-based）
- [x] 加密握手包编解码

## 帧加密
- [x] AES-256-CTR 流加密器
- [x] 双向加密器（send/recv 独立 IV）
- [x] GCM 认证加密（备用）
- [x] EncryptedFrame 接收（解密）
- [x] EncryptedFrame 发送（加密）

## 握手协议
- [x] EncryptedHandshake (0x91) 编码/解码
- [x] ServerHandshake (0x93) 编码/解码
- [x] NewConnectionAccepted (0x92) 编码
- [x] HKDF 密钥派生
- [x] Challenge 签名验证
- [x] Raknet 层加密钩子（Conn.EnableEncryption）

## 测试
- [x] ECDH 密钥交换端到端测试
- [x] AES-256-CTR 加解密往返测试
- [x] 密钥派生确定性测试
- [x] 握手包编解码测试
- [x] 端到端加密测试（Alice→Bob）

## Phase 1 完成情况
- [x] Raknet 帧层 ✅
- [x] FB 认证 ✅
- [x] MC 登录 + PyRpc 握手 ✅
- [x] 加密层 ✅
