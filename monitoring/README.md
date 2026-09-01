# IcePoint Coffee — 监控配置

## 快速启动

```bash
# 启动 Prometheus + Grafana
docker-compose up -d

# 访问
# Grafana:  http://localhost:3000  (admin / admin)
# Prometheus: http://localhost:9090
```

## 架构

```
┌─────────────────┐     ┌──────────────┐     ┌──────────┐
│  IcePoint HTTP  │────▶│ Prometheus   │────▶│ Grafana  │
│  /metrics      │     │  :9090       │     │  :3000   │
└─────────────────┘     └──────────────┘     └──────────┘
```

## 指标清单

### HTTP 请求
| 指标 | 类型 | 描述 |
|------|------|------|
| `icepoint_http_requests_total` | Counter | 总请求数 |
| `icepoint_http_request_duration_seconds` | Histogram | 请求延迟分布 |
| `icepoint_http_errors_total` | Counter | HTTP 错误数 |

### AI 对话
| 指标 | 类型 | 描述 |
|------|------|------|
| `icepoint_ai_chat_total` | Counter | AI 对话总数 |
| `icepoint_ai_chat_success_total` | Counter | 成功次数 |
| `icepoint_ai_chat_failed_total` | Counter | 失败次数 |
| `icepoint_ai_chat_duration_seconds` | Histogram | AI 响应延迟 |
| `icepoint_ai_tokens_used` | Counter | Token 消耗 |
| `icepoint_ai_tool_calls_total` | Counter | 工具调用数 |

### 建筑生成
| 指标 | 类型 | 描述 |
|------|------|------|
| `icepoint_build_total` | Counter | 建筑总数（按类型） |
| `icepoint_build_blocks_total` | Counter | 方块生成总数 |
| `icepoint_build_duration_seconds` | Histogram | 生成延迟 |

### 命令执行
| 指标 | 类型 | 描述 |
|------|------|------|
| `icepoint_commands_total` | Counter | 命令总数 |
| `icepoint_command_success_total` | Counter | 成功命令数 |
| `icepoint_command_failed_total` | Counter | 失败命令数 |

### 事件流
| 指标 | 类型 | 描述 |
|------|------|------|
| `icepoint_event_stream_connections` | Gauge | SSE 连接数 |
| `icepoint_event_stream_messages_total` | Counter | 消息总数 |

### MC 连接
| 指标 | 类型 | 描述 |
|------|------|------|
| `icepoint_mc_connected` | Gauge | MC 连接状态 |
| `icepoint_mc_reconnects_total` | Counter | 重连次数 |

### Runtime
| 指标 | 类型 | 描述 |
|------|------|------|
| `icepoint_go_memstats_alloc_bytes` | Gauge | 堆内存分配 |
| `icepoint_go_memstats_sys_bytes` | Gauge | 系统内存 |
| `icepoint_go_goroutines` | Gauge | Goroutine 数 |

## Prometheus 查询示例

```promql
# QPS
rate(icepoint_http_requests_total[5m])

# P99 延迟
histogram_quantile(0.99,
  sum by (le) (rate(icepoint_http_request_duration_seconds_bucket[5m])))

# 错误率
rate(icepoint_http_errors_total[5m]) / rate(icepoint_http_requests_total[5m])

# 最热端点
topk(5, sum by (path) (rate(icepoint_http_requests_total[5m])))

# AI 成功率
rate(icepoint_ai_chat_success_total[5m]) / 
(rate(icepoint_ai_chat_success_total[5m]) + rate(icepoint_ai_chat_failed_total[5m]))

# 建筑生成速率
rate(icepoint_build_total[5m])

# 每种建筑占比
sum by (type) (rate(icepoint_build_total[5m])) / 
ignoring(type) group_left
sum(rate(icepoint_build_total[5m]))
```

## 告警示例

```yaml
groups:
- name: icepoint
  rules:
  - alert: HighErrorRate
    expr: rate(icepoint_http_errors_total[5m]) / rate(icepoint_http_requests_total[5m]) > 0.05
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "错误率超过 5%"

  - alert: SlowResponse
    expr: histogram_quantile(0.99, rate(icepoint_http_request_duration_seconds_bucket[5m])) > 5
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "P99 延迟超过 5 秒"

  - alert: AIHighFailure
    expr: rate(icepoint_ai_chat_failed_total[5m]) / rate(icepoint_ai_chat_total[5m]) > 0.1
    for: 10m
    labels:
      severity: warning
    annotations:
      summary: "AI 调用失败率超过 10%"
```
