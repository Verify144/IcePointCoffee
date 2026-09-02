# PHASE 10 — 用户认证 + 建筑模板 + 定时任务

## 用户认证系统
- [x] API Token 模型（users 表 + tokens 表）
- [x] Token 生成/验证中间件
- [x] 受保护路由（admin 端点）
- [x] 登录/注册 API（可选，简化为 token 管理）

## 建筑模板系统
- [x] 模板模型（存储结构定义）
- [x] 模板 CRUD API（创建/列表/获取/删除）
- [x] 模板 Marketplace（公开模板）
- [ ] 模板参数化（动态替换变量）
- [ ] 从模板构建（build engine 集成）

## 定时任务（Cron）
- [x] Cron 调度器（基于 Go cron 库）
- [x] 定时任务模型（存储 Cron 表达式）
- [x] Cron CRUD API
- [x] 与任务系统集成（触发 builder/agent）
- [x] Cron 日志记录
