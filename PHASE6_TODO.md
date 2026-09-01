# PHASE6 — 性能压测 + 更多协议包

## 性能压测
- [ ] HTTP 压测（wrk 风格）
- [ ] AI 工具调用压测
- [ ] 事件流并发压测
- [ ] 内存分配优化
- [ ] 字符串拼接优化（bytes.Buffer 复用）
- [ ] JSON 复用（sync.Pool）
- [ ] 基准测试报告

## 更多协议包
- [ ] PlayerSkin (0x9F) - 玩家皮肤
- [ ] PlayerList (0x3F) - 玩家列表
- [ ] BossEvent (0x4A) - Boss 事件
- [ ] MapData (0x86) - 地图数据
- [ ] PlayerFog (0xA0) - 玩家雾
- [ ] SubClient (0x8B) - 子客户端登录
- [ ] Transfer (0x9A) - 服务器传送
- [ ] GameRulesChanged (0xA1) - 游戏规则改变
- [ ] Camera (0x9B) - 相机视角
- [ ] PlayerEnchantOptions (0x9E) - 附魔选项

## 测试
- [ ] 基准测试（go test -bench）
- [ ] 压测报告输出
