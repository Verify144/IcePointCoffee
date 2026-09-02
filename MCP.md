# IcePointCoffee MCP 服务器

[Model Context Protocol](https://modelcontextprotocol.io/) 服务器，让 Claude Desktop / Cursor / 其他 AI 框架直接调用 IcePointCoffee 的 28 个 MC 操控工具。

## 协议

- 标准 JSON-RPC 2.0
- MCP protocol version: **2025-06-18**
- 传输层：**STDIO**（Claude Desktop 标准）+ **HTTP+SSE**（HTTP 客户端）

## Claude Desktop 配置

编辑 `~/.config/Claude/claude_desktop_config.json`（macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`）：

```json
{
  "mcpServers": {
    "IcePointCoffee": {
      "command": "/path/to/icepoint",
      "args": ["mcp-server"],
      "env": {
        "VERIFY_SERVER": "hk-8.qwq.fan:19132",
        "VERIFY_TOKEN": "your-fb-master-token",
        "VERIFY_COOKIE": "optional-game-cookie"
      }
    }
  }
}
```

重启 Claude Desktop 即可看到 28 个 `mc_*` 工具。

## Cursor / 其他 AI（HTTP 模式）

启动 HTTP 服务器：

```bash
icepoint --http --mcp  # 同时启用 HTTP + MCP 端点
```

AI 客户端配置：

```json
{
  "mcpServers": {
    "IcePointCoffee": {
      "url": "http://localhost:8080/mcp",
      "headers": {
        "MCP-Protocol-Version": "2025-06-18"
      }
    }
  }
}
```

## 28 个可用工具

### 玩家/通信
- `mc_command` - 执行 MC 命令（带黑名单）
- `mc_chat` - 发送聊天消息
- `mc_teleport` - 传送玩家
- `mc_give` - 给予物品
- `mc_gamemode` - 切换游戏模式
- `mc_swing` - 挥臂动画
- `mc_respawn` - 重生
- `mc_xp` - 经验值
- `mc_effect` - 药水效果
- `mc_clear_inventory` - 清空物品栏
- `mc_dialog` - 富文本/标题

### 实体/世界
- `mc_spawn_entity` - 召唤生物
- `mc_remove_entity` - 移除实体
- `mc_attack` - 攻击实体
- `mc_nametag` - 命名牌
- `mc_setblock` - 放置方块
- `mc_fill` - 填充区域
- `mc_spawnpoint` - 设置出生点

### 环境/视觉
- `mc_world` - 时间/天气
- `mc_weather` - 天气控制
- `mc_worldborder` - 世界边界
- `mc_difficulty` - 难度
- `mc_particle` - 粒子效果
- `mc_play_sound` - 播放声音
- `mc_stop_sound` - 停止声音
- `mc_bossbar` - Boss 血条
- `mc_status` - 连接状态

### AI 智能
- `mc_perceive` - 感知周围环境（玩家/生物/方块 + 自然语言描述）

## 示例对话（Claude Desktop）

**用户**：你周围有什么？

**Claude**：[调用 `mc_perceive`]
> 收到，我看到你站在主城广场中心，坐标 (0, 64, 0)，生命值 20/20，游戏模式生存。附近有 Steve 和 Alex 两名玩家，还有猪和牛各一头。

**用户**：给 Steve 一些钻石

**Claude**：[调用 `mc_give`，target="Steve", item="diamond", count=64]
> 已给 Steve 64 钻石！

**用户**：建一栋 10×10 的玻璃房子

**Claude**：[调用 `mc_fill`，x1=0, y1=64, z1=0, x2=10, y2=70, z2=10, block="glass"]
> 已用玻璃填充 10×7×10 的区域。

## 安全

- 命令黑名单自动生效（`stop`/`kick`/`ban`/`op`/`deop`/`reload`/`publish` 被拒绝）
- `mc_fill` 限制单次 100k 方块
- HTTP 模式支持 `MCP-Protocol-Version` 头校验

## HTTP 端点

| 端点 | 方法 | 说明 |
|---|---|---|
| `/mcp/info` | GET | 服务器信息 + 端点列表 |
| `/mcp` | POST | JSON-RPC 2.0 主入口 |
| `/mcp/stream` | GET | SSE 双向通信 |
| `/mcp/notify` | POST | 触发通知广播 |

## 调试

```bash
# 手动测试 STDIO
printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","clientInfo":{"name":"test"}}}\n' | icepoint mcp-server

# 手动测试 HTTP
curl -X POST http://localhost:8080/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

## 文件

- `internal/mcp/types.go` - JSON-RPC + MCP 协议类型
- `internal/mcp/server.go` - MCP Handler (tools/list, tools/call, initialize)
- `internal/mcp/stdio.go` - STDIO transport
- `internal/mcp/http.go` - HTTP+SSE transport
- `cmd/mcp-server/main.go` - 独立 STDIO 入口
