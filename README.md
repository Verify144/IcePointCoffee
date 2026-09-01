# IcePoint Coffee

**冰点咖啡** - Minecraft 建筑 AI 助手。

> 用自然语言描述你的建筑需求，AI 自动生成并推送到租赁服。

MIT License

## 特性

- **🤖 内置 AI Agent** - 基于 OpenAI 兼容协议，理解自然语言建筑需求
- **🏗️ 建筑生成器** - 支持 house/tower/circle/sphere/wall/floor/rect 等内置图形
- **📋 任务队列** - SQLite 持久化，支持历史查询
- **🔌 HTTP RPC 插件系统** - 用任意语言编写插件，跨进程扩展
- **⚡ 指令执行** - 批量 setblock/fill/structure 推送到租赁服
- **🔒 安全配置** - API Key/FB Token 通过环境变量注入，永不外泄

## 快速开始

### 1. 创建配置

```bash
mkdir -p ~/.icepoint
cp config.example.yaml ~/.icepoint/config.yaml
```

编辑 `~/.icepoint/config.yaml`，填写：

```yaml
ai:
  base_url: "https://api.openai.com/v1"   # 或 DeepSeek / Qwen / Ollama 等
  api_key: "sk-xxxxxxxx"
  model: "gpt-4o-mini"

server:
  address: "your.server.netease.com:25565"
  player_name: "YourBot"
  fb_token: "your_fb_master_token"
```

### 2. 运行

```bash
go build -o icepoint main.go
./icepoint
```

### 3. 使用

```
❄ > 做一个 10x10 的石头地板
❄ > build house width:10 height:5 block:oak_planks center:0,64,0
❄ > 描述: 建一个圆形喷泉半径8格
```

## 配置说明

### AI 配置

| 字段 | 说明 | 默认值 |
|------|------|--------|
| `ai.base_url` | OpenAI 兼容 API 地址 | `https://api.openai.com/v1` |
| `ai.api_key` | API Key | **必填** |
| `ai.model` | 模型名 | `gpt-4o-mini` |
| `ai.temperature` | 温度（0-2） | `0.7` |
| `ai.max_tokens` | 最大 token | `4096` |

### 服务器配置

| 字段 | 说明 |
|------|------|
| `server.address` | 租赁服地址 |
| `server.player_name` | 机器人玩家名 |
| `server.fb_token` | FB Master Token |

### 环境变量覆盖

```bash
export IC_AI_API_KEY="sk-xxx"
export IC_SERVER_ADDRESS="server:25565"
export IC_SERVER_FB_TOKEN="token"
./icepoint
```

## 插件开发

### 创建插件

```
plugins/
  my_plugin/
    plugin.json   # 元信息
    plugin        # 可执行文件（任意语言）
```

### plugin.json

```json
{
  "name": "我的插件",
  "description": "生成复杂建筑",
  "executable": "my_build",
  "port": 8766
}
```

### 插件开发（任意语言）

```go
// plugin/main.go
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

func main() {
    http.HandleFunc("/rpc", func(w http.ResponseWriter, r *http.Request) {
        body, _ := io.ReadAll(r.Body)
        var req map[string]any
        json.Unmarshal(body, &req)

        result := map[string]any{
            "status": "ok",
            "type":   req["type"],
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]any{
            "result": result,
        })
    })
    http.ListenAndServe(":8766", nil)
}
```

## 内置建筑类型

| 类型 | 说明 | 参数 |
|------|------|------|
| `house` | 小屋 | width, height, depth, block |
| `tower` | 高塔 | width, height, block |
| `circle` | 圆形平台 | radius, block |
| `sphere` | 球体 | radius, block |
| `wall` | 墙 | width, height, block |
| `floor` | 地板 | width, depth, block |
| `rect` | 矩形区域 | width, height, depth, block |

## 命令参考

| 命令 | 说明 |
|------|------|
| `/tasks` | 查看最近任务 |
| `/plugins` | 查看插件 |
| `/connect` | 连接服务器 |
| `/disconnect` | 断开连接 |
| `/help` | 帮助 |
| `/quit` | 退出 |

## 架构

```
IcePointCoffee/
├── main.go              # 入口
├── config/              # 配置加载
├── ai/                  # OpenAI 兼容客户端
├── agent/               # AI Agent 引擎
├── builder/             # 建筑生成器
├── importer/            # 指令执行器
├── server/              # 租赁服连接
├── db/                  # SQLite 持久化
└── plugin/              # HTTP RPC 插件系统
```

## 安全注意

- **API Key** 和 **FB Token** 永不写入代码或日志
- 通过环境变量注入：`IC_AI_API_KEY` / `IC_SERVER_FB_TOKEN`
- 插件进程隔离，单个插件崩溃不影响主进程
