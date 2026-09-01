# ☕ IcePoint Coffee

> 自研 Raknet 协议栈 + AI 建筑助手 + Web Dashboard

**MIT License** · Go 1.21+ · 零外部 MC 依赖 · 自研比例 100%

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

---

## ✨ 特性

### 自研协议栈（100% 原创代码）
- 🔐 **Raknet 协议**：帧封装 / 可靠性 / 流控制 / 粘包
- 🔑 **FB 认证**：Facebook Auth 登录流程完整实现
- 🛡️ **ECDH 加密**：X25519 + P-384 + HKDF + AES-256-CTR
- 🎮 **MC 协议**：Login / Text / Command / Move / Inventory / Container / BossBar / Skin 等 **50+ 包**

### AI 建筑助手
- 🤖 多轮对话记忆（50 条上下文）
- 🛠️ 工具调用框架（OpenAI 兼容格式）
- 🏠 内置 7 种建筑生成器：House / Tower / Circle / Sphere / Wall / Floor / Rect
- 📝 Mock AI 模式（无需 API Key）

### Web Dashboard
- 📊 实时状态监控（SSE 事件流）
- 🤖 AI 对话界面（Markdown 渲染）
- 🏠 建筑生成可视化
- 💬 命令执行面板
- 🔌 插件管理
- 🛠️ 工具搜索
- ⌨️ 键盘快捷键（`1-8` 切换 Tab，`Ctrl+R` 刷新，`Ctrl+K` 清空）

### 性能
- 协议编码：**109ns - 852ns / 包**（33M 包/秒）
- 并发写入：**3 亿次/秒**（0 GC 压力）
- 建筑生成：**10万 - 30万座/秒**
- 所有池化操作：**0 GC 分配**

---

## 📦 安装

### 二进制（推荐）

```bash
# 下载最新的 icepoint 二进制
chmod +x icepoint
./icepoint --help
```

### 从源码构建

```bash
git clone https://github.com/Verify144/IcePointCoffee.git
cd IcePointCoffee
go build -o icepoint .
```

---

## ⚙️ 配置

创建配置文件 `~/.icepoint/config.yaml`：

```yaml
# AI 配置（可选，不填则使用 Mock AI）
ai:
  base_url: "https://api.openai.com/v1"  # OpenAI 兼容 API
  api_key: "your-api-key-here"
  model: "gpt-3.5-turbo"

# 服务器配置
server:
  address: "your-server-code"      # 房间号
  fb_token: "your-fb-token"       # Facebook Master Token
  player_name: "IcePoint"          # 玩家名

# 数据库
db:
  path: "~/.icepoint/data.db"

# 插件配置
plugin:
  dir: "./plugins"
  http_port: 8080                  # Dashboard 端口（设为 0 则不启动）
```

---

## 🚀 快速开始

### 1. 启动

```bash
./icepoint
```

输出：
```
    ░█▀▀░█░█░█▀▀░░░█░░░█▀▀░█▀▄░█▄█░█▀█░█▀▄░░░█▀▀░█▀▀░█▀▄░█▀▀░█▀▄
    ░█▀▀░█░█░▀▀█░░░█░░░█░░░█▀▄░█░█░█░█░█▀▄░░░█░░░█▀▀░█▀▄░▀▀█░█▀▄
    ░▀▀▀░▀▀▀░▀▀▀░░░▀░░░▀▀▀░▀░▀░▀░▀░▀░░░▀░░░░░▀▀▀░▀░░░▀░▀░▀░░

冰点咖啡 v1.0.0
AI: Mock 模式
Dashboard: http://127.0.0.1:8080/
❄ >
```

### 2. Web Dashboard

打开 http://127.0.0.1:8080/ 即可访问 Dashboard。

### 3. CLI 命令

```
❄ > build house width:10 height:5
正在分析需求...
任务 ID: task_1234567890
结果: 建造一个 10x10 的橡木房子
状态: completed
```

### 4. HTTP API

```bash
# 健康检查
curl http://localhost:8080/health

# AI 对话
curl -X POST http://localhost:8080/api/v1/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "你好", "use_tools": true}'

# 生成建筑
curl -X POST http://localhost:8080/api/v1/build \
  -H "Content-Type: application/json" \
  -d '{"type": "house", "size": 10, "x": 0, "y": 64, "z": 0}'

# 执行命令
curl -X POST http://localhost:8080/api/v1/commands \
  -H "Content-Type: application/json" \
  -d '{"command": "say Hello from IcePoint!"}'

# SSE 事件流
curl -N http://localhost:8080/api/v1/events
```

---

## 🗂️ 项目结构

```
IcePointCoffee/
├── main.go                          # 入口 + CLI 主循环
├── internal/
│   ├── agent/                       # AI Agent 引擎
│   │   └── engine.go
│   ├── ai/                          # AI 客户端
│   │   ├── client.go               # OpenAI 兼容客户端
│   │   ├── memory.go               # 多轮对话记忆
│   │   ├── tools.go                # 工具注册表
│   │   └── builtin_tools.go        # 内置工具
│   ├── builder/                     # 建筑生成器
│   │   └── builder.go              # 7 种建筑类型
│   ├── config/                      # 配置加载
│   │   └── config.go
│   ├── db/                          # SQLite 数据库
│   │   └── db.go
│   ├── importer/                     # 建筑导入器
│   │   └── importer.go
│   ├── netherite/                   # 核心 Raknet 协议
│   │   ├── raknet/                 # Raknet 帧/连接
│   │   ├── crypto/                 # ECDH 加密
│   │   ├── auth/                   # FB 认证
│   │   ├── mc/                     # MC 客户端
│   │   └── protocol/               # MC 协议包（50+）
│   │       ├── login.go            # 包 ID 定义
│   │       ├── reader.go           # 二进制读写
│   │       ├── text.go             # 聊天
│   │       ├── command.go          # 命令
│   │       ├── player.go           # 移动/动画/标题
│   │       ├── player_ext.go       # 皮肤/Boss/地图/相机
│   │       ├── inventory.go        # 物品/容器
│   │       ├── batch.go            # 命令批处理
│   │       └── perf.go             # 性能优化
│   ├── plugin/                      # 插件管理器
│   │   └── plugin.go
│   └── server/                      # HTTP RPC 服务
│       ├── server.go               # REST API + SSE
│       └── dashboard_v2.go         # Dashboard SPA（33KB）
├── plugins/                          # 插件目录
│   └── example/
│       ├── main.go
│       └── plugin.json
├── BENCHMARK_REPORT.md             # 性能报告
└── PHASE*.md                       # 各阶段任务追踪
```

---

## 🛠️ HTTP API 端点

### 状态
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/api/v1/status` | 系统状态 |
| GET | `/api/v1/events` | SSE 事件流 |

### AI
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/ai/chat` | AI 对话 |
| GET | `/api/v1/ai/tools` | 工具列表 |
| GET | `/api/v1/ai/memory` | 对话记忆 |
| DELETE | `/api/v1/ai/memory` | 清空记忆 |

### 命令 & 建筑
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/commands` | 命令列表 |
| POST | `/api/v1/commands` | 发送命令 |
| POST | `/api/v1/build` | 生成建筑 |

### 插件
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/plugins` | 插件列表 |
| POST | `/api/v1/plugins/register` | 注册插件 |

### 管理员
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/admin/reload` | 重载配置 |
| POST | `/api/admin/restart` | 重启服务 |
| GET | `/api/admin/stats` | 统计信息 |

---

## 📈 Prometheus Metrics

访问 `http://localhost:8080/metrics` 获取 Prometheus 格式指标。

### 可用指标

| 指标名 | 类型 | 说明 |
|--------|------|------|
| `icepoint_http_requests_total` | Counter | HTTP 请求总数（按 method/path/status） |
| `icepoint_http_request_duration_seconds` | Histogram | HTTP 请求延迟分布 |
| `icepoint_ai_chat_total` | Counter | AI 对话总数 |
| `icepoint_ai_chat_success_total` | Counter | AI 成功对话 |
| `icepoint_ai_chat_duration_seconds` | Histogram | AI 响应延迟 |
| `icepoint_build_total` | Counter | 建筑生成总数（按类型） |
| `icepoint_build_blocks_total` | Counter | 生成方块总数 |
| `icepoint_command_total` | Counter | 命令执行总数 |
| `icepoint_mc_connected` | Gauge | MC 连接状态 (1/0) |
| `icepoint_event_stream_connections` | Gauge | SSE 连接数 |
| `icepoint_go_memstats_alloc_bytes` | Gauge | Go 堆内存分配 |
| `icepoint_go_goroutines` | Gauge | Goroutine 数量 |
| `icepoint_rate_limited_total` | Counter | 被限流的请求数 |

### Grafana + Prometheus 集成

```yaml
# docker-compose.yml
version: '3.8'
services:
  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
```

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'icepoint'
    static_configs:
      - targets: ['host.docker.internal:8080']
```

### 示例查询（Grafana）

```promql
# QPS
rate(icepoint_http_requests_total[5m])

# P99 延迟
histogram_quantile(0.99,
  rate(icepoint_http_request_duration_seconds_bucket[5m]))

# AI 成功率
sum(rate(icepoint_ai_chat_success_total[5m]))
  / sum(rate(icepoint_ai_chat_total[5m]))

# 当前连接数
icepoint_event_stream_connections
```

## 📊 MC 协议包清单

| ID | 包名 | 用途 |
|----|------|------|
| 0x01 | Login | 登录 |
| 0x02 | PlayStatus | 状态 |
| 0x05 | Disconnect | 断开 |
| 0x09 | Text | 聊天/提示 |
| 0x0A | SetTime | 时间 |
| 0x0B | StartGame | 开始游戏 |
| 0x0D | AddActor | 添加实体 |
| 0x0F | RemoveActor | 移除实体 |
| 0x3F | PlayerList | 玩家列表 |
| 0x4A | BossEvent | Boss 血条 |
| 0x5C | SetTitle | 标题 |
| 0x86 | MapData | 地图数据 |
| 0x8B | SubClientLogin | 子客户端 |
| 0x90 | WebSocketCommand | WebSocket |
| 0x9A | Transfer | 服务器传送 |
| 0x9B | Camera | 相机视角 |
| 0x9F | PlayerSkin | 玩家皮肤 |
| 0xA1 | GameRulesChanged | 游戏规则 |
| 0xD2 | CommandRequest | 命令请求 |
| 0xD3 | CommandOutput | 命令输出 |
| 0xE1 | Animate | 动画 |
| 0xE2 | UpdateAttributes | 属性更新 |
| 0xF2 | InventoryTransaction | 物品交互 |
| 0xF3 | InventoryContent | 背包内容 |
| 0xF5 | MovePlayer | 玩家移动 |
| 0xF8 | ContainerOpen | 打开容器 |
| 0xF9 | ContainerClose | 关闭容器 |

---

## ⌨️ 键盘快捷键

| 快捷键 | 功能 |
|--------|------|
| `1` - `8` | 切换 Dashboard Tab |
| `Enter` | 发送 AI 消息 |
| `Shift+Enter` | 换行 |
| `Ctrl+R` | 刷新状态 |
| `Ctrl+K` | 清空聊天 |

---

## 🔧 故障排查

### 连接失败
```
连接服务器失败: context deadline exceeded
```
- 检查 `server.address`（房间号）是否正确
- 检查 `server.fb_token` 是否有效
- 确认网络可以访问网易服务器

### AI 无响应
```
AI: Mock 模式
```
- 配置 `ai.base_url` 和 `ai.api_key`
- 或使用内置 Mock AI（无需配置）

### Dashboard 打不开
```
HTTP RPC 插件服务: http://127.0.0.1:8080/
```
- 检查 `plugin.http_port` 是否为 `0`
- 确认端口未被占用：`lsof -i :8080`

---

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE)

---

## 🙏 致谢

- 自研 Raknet 参考 [neomega-core](https://github.com/Neilgravity/neomega-core) 架构
- ECDH 参考 RFC 7748 / RFC 8442
- MC 协议参考 [protocol MaineCraft](https://github.com/Marcussacapuces) 文档

---

**Made with ❤️ by Verify144 · © 2026**
