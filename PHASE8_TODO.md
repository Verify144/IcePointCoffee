# PHASE8 — Prometheus Metrics 可观测性

## Metrics 端点
- [x] /metrics 端点（Prometheus 格式）
- [x] Counter/Gauge/Histogram 指标类型
- [x] 指标注册表

## 埋点指标
- [x] AI 指标（chat 总数/成功/失败/token 消耗/延迟）
- [x] HTTP 请求指标（总数/端点/方法/状态码/延迟）
- [x] 建筑生成指标（总数/类型/方块数/延迟）
- [x] MC 连接指标（连接状态/重连次数）
- [x] 命令指标（总数/成功率/延迟）
- [x] 事件流指标（连接数/消息数）

## 可选：Grafana Dashboard
- [x] Docker Compose 配置（已写入 README）
- [x] PromQL 示例（已写入 README）
- [ ] Grafana Dashboard JSON

## 测试
- [x] 指标端点测试
- [x] 埋点正确性验证
