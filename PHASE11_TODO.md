# PHASE 11 — AI 实时操控租赁服

## 目标
让 AI 能够在对话中动态生成并执行 MC 命令，实时影响租赁服。
无需断开重连，类似 IDE 中"执行代码片段"。

## 方案
新增 AI Tool 集合（MC 内置能力），每个工具映射到具体的 mc.Client 方法。

### 工具集（mc_tools.go）
- [x] `mc_command` - 执行任意命令（带白名单检查）
- [x] `mc_chat` - 发送聊天消息
- [x] `mc_teleport` - 传送到坐标
- [x] `mc_give` - 给予物品
- [x] `mc_setblock` - 设置单个方块
- [x] `mc_fill` - 填充区域
- [x] `mc_dialog` - 发送 tellraw/title 给玩家
- [x] `mc_gamemode` - 切换游戏模式
- [x] `mc_world` - 设置时间/天气
- [x] `mc_status` - 查询连接状态

### 安全保障
- [x] 命令黑名单（禁用 `stop` / `kick` / `ban` / `op` / `deop` / `reload`）
- [ ] 命令白名单（仅允许常见命令前缀）— 用黑名单更简单
- [ ] 危险命令确认（红色警告：fill/give 大量）
- [x] fill 体积限制 100,000 方块

### 集成
- [x] AI 工具自动注册到 server
- [ ] Dashboard 新增"操控"标签页（实时显示连接状态 + 命令历史）
- [ ] 流式 AI 对话中调用工具并实时回显结果

### 测试
- [x] 单元测试：每个工具的参数解析
- [x] 集成测试：Mock 客户端验证命令发送
- [x] 黑名单测试：危险命令被拒绝

