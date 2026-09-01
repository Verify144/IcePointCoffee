# ICEPOINT_TODO — IcePoint Coffee 项目

## Phase 1 — 项目初始化
- [ ] 项目结构初始化（go.mod, main.go, cmd/cli）
- [ ] 配置文件加载（config.yaml / env）
- [ ] SQLite 持久化层
- [ ] VerifyBlockMap / VerifyIceStructure / VerifyImporter 集成

## Phase 2 — 核心模块
- [ ] AI 客户端（OpenAI 兼容协议，支持多模型）
- [ ] Agent 引擎（单回合 + 任务队列）
- [ ] 建筑生成器（调用 VerifyBlockMap + VerifyIceStructure）
- [ ] 指令执行器（调用 VerifyImporter）
- [ ] 租赁服连接器（Raknet / WebSocket）

## Phase 3 — 插件系统
- [ ] HTTP RPC 插件协议
- [ ] 插件注册与生命周期管理
- [ ] 内置示例插件

## Phase 4 — 用户交互
- [ ] CLI 主界面
- [ ] 任务状态查询
- [ ] 配置管理命令

## Phase 5 — 测试与文档
- [ ] 单元测试
- [ ] README / 文档
- [ ] GitHub 推送
