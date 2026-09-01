# PHASE6 — 性能压测 + 更多协议包

## 性能压测
- [x] 协议包编码基准测试
- [x] 内存分配优化
- [x] 字符串拼接优化（bytes.Buffer 复用）
- [x] JSON 复用（sync.Pool）
- [x] 基准测试报告
- [x] 批处理 / 池化基准
- [x] 并发基准测试

## 更多协议包
- [x] PlayerSkin (0x9F) - 玩家皮肤
- [x] PlayerList (0x3F) - 玩家列表
- [x] BossEvent (0x4A) - Boss 事件
- [x] MapData (0x86) - 地图数据
- [x] Transfer (0x9A) - 服务器传送
- [x] GameRulesChanged (0xA1) - 游戏规则改变
- [x] Camera (0x9B) - 相机视角
- [x] PlayerEnchantOptions (0x9E) - 附魔选项

## 测试
- [x] 基准测试（go test -bench）
- [x] 压测报告输出
