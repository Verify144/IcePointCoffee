# PHASE 11B — Dashboard 操控标签页 + 命令历史流

## Dashboard 操控标签页
- [ ] 新增"操控" tab（/dashboard/control.html）
- [ ] 实时显示连接状态（connected/disconnected）
- [ ] 命令历史（时间 + 命令 + 输出 + 耗时）
- [ ] AI 执行工具时实时推送结果到前端（SSE / WebSocket）
- [ ] 命令历史持久化（SQLite）
- [ ] 一键清空历史
- [ ] 危险命令红色高亮

## 流式 AI 工具调用回显
- [ ] AI 调用 mc_* 工具时，通过 SSE 推送 tool_result 事件
- [ ] 前端实时显示"🔧 mc_command: time set day → 执行成功"
- [ ] 危险命令警告（fill > 10k / give > 64）

## 命令历史 API
- [ ] GET /api/v1/commands - 列出历史
- [ ] DELETE /api/v1/commands - 清空历史
