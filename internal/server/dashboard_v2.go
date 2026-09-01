package server

const dashboardHTMLV2 = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>☕ IcePoint Coffee - Dashboard</title>
<style>
:root {
  --bg-primary: #0a0e1a;
  --bg-secondary: #121826;
  --bg-tertiary: #1a2138;
  --border: #1e2740;
  --text-primary: #e8eaf0;
  --text-secondary: #a0a8b8;
  --text-muted: #5a6478;
  --accent-primary: #00d9ff;
  --accent-secondary: #ff6b9d;
  --success: #4ade80;
  --warning: #fbbf24;
  --error: #f87171;
  --info: #60a5fa;
  --shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  --transition: 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

* { margin: 0; padding: 0; box-sizing: border-box; }

html, body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', sans-serif;
  background: var(--bg-primary);
  color: var(--text-primary);
  min-height: 100vh;
  overflow-x: hidden;
}

::-webkit-scrollbar { width: 8px; height: 8px; }
::-webkit-scrollbar-track { background: var(--bg-secondary); }
::-webkit-scrollbar-thumb { background: var(--bg-tertiary); border-radius: 4px; }
::-webkit-scrollbar-thumb:hover { background: var(--accent-primary); }

/* === Header === */
.header {
  background: linear-gradient(135deg, var(--bg-secondary), var(--bg-tertiary));
  padding: 16px 24px;
  border-bottom: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: sticky;
  top: 0;
  z-index: 100;
  backdrop-filter: blur(20px);
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo h1 {
  font-size: 1.4rem;
  font-weight: 700;
  background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.logo-icon {
  font-size: 1.8rem;
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-4px); }
}

.status-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  background: var(--bg-tertiary);
  border-radius: 20px;
  font-size: 0.85rem;
  border: 1px solid var(--border);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--success);
  box-shadow: 0 0 8px var(--success);
  animation: pulse 2s ease-in-out infinite;
}

.status-dot.offline { background: var(--error); box-shadow: 0 0 8px var(--error); }
.status-dot.warning { background: var(--warning); box-shadow: 0 0 8px var(--warning); }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.header-actions {
  display: flex;
  gap: 8px;
}

.icon-btn {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  color: var(--text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: var(--transition);
  font-size: 1rem;
}

.icon-btn:hover {
  background: var(--accent-primary);
  color: var(--bg-primary);
  transform: scale(1.05);
}

/* === Main Layout === */
.main {
  display: grid;
  grid-template-columns: 240px 1fr;
  min-height: calc(100vh - 70px);
}

.sidebar {
  background: var(--bg-secondary);
  padding: 20px 16px;
  border-right: 1px solid var(--border);
}

.sidebar nav { display: flex; flex-direction: column; gap: 4px; }

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  color: var(--text-secondary);
  border-radius: 8px;
  text-decoration: none;
  transition: var(--transition);
  cursor: pointer;
  font-size: 0.9rem;
  user-select: none;
}

.nav-item:hover {
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

.nav-item.active {
  background: linear-gradient(90deg, var(--accent-primary), transparent);
  color: var(--accent-primary);
  border-left: 3px solid var(--accent-primary);
  padding-left: 11px;
}

.nav-icon { font-size: 1.1rem; width: 20px; text-align: center; }

.content {
  padding: 24px;
  overflow-y: auto;
}

/* === Cards === */
.card {
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 20px;
  box-shadow: var(--shadow);
  transition: var(--transition);
}

.card:hover { border-color: var(--bg-tertiary); }

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border);
}

.card-title {
  font-size: 1.1rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

/* === Stats === */
.stat-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}

.stat {
  background: var(--bg-tertiary);
  padding: 20px;
  border-radius: 12px;
  text-align: center;
  border: 1px solid var(--border);
  transition: var(--transition);
  position: relative;
  overflow: hidden;
}

.stat::before {
  content: '';
  position: absolute;
  top: 0; left: 0; right: 0;
  height: 2px;
  background: linear-gradient(90deg, var(--accent-primary), var(--accent-secondary));
  opacity: 0;
  transition: var(--transition);
}

.stat:hover::before { opacity: 1; }
.stat:hover { transform: translateY(-2px); }

.stat-value {
  font-size: 2.2rem;
  font-weight: 700;
  background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  line-height: 1.2;
}

.stat-label {
  color: var(--text-muted);
  font-size: 0.85rem;
  margin-top: 6px;
}

.stat-trend {
  font-size: 0.75rem;
  color: var(--success);
  margin-top: 4px;
}

.stat-trend.down { color: var(--error); }

/* === Chat === */
.chat-container {
  height: 500px;
  display: flex;
  flex-direction: column;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  background: var(--bg-primary);
  border-radius: 8px;
  margin-bottom: 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.msg {
  max-width: 75%;
  padding: 10px 14px;
  border-radius: 12px;
  word-wrap: break-word;
  white-space: pre-wrap;
  line-height: 1.5;
  font-size: 0.95rem;
  position: relative;
  animation: msgIn 0.3s ease-out;
}

@keyframes msgIn {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

.msg.user {
  background: var(--accent-primary);
  color: var(--bg-primary);
  margin-left: auto;
  border-bottom-right-radius: 4px;
}

.msg.ai {
  background: var(--bg-tertiary);
  color: var(--text-primary);
  border-bottom-left-radius: 4px;
  border: 1px solid var(--border);
}

.msg.system {
  background: transparent;
  color: var(--text-muted);
  font-size: 0.85rem;
  text-align: center;
  font-style: italic;
  max-width: 100%;
}

.msg.error {
  background: rgba(248, 113, 113, 0.1);
  color: var(--error);
  border: 1px solid var(--error);
}

.msg .timestamp {
  font-size: 0.7rem;
  opacity: 0.6;
  margin-top: 4px;
}

.msg pre {
  background: var(--bg-primary);
  padding: 8px 12px;
  border-radius: 6px;
  margin: 8px 0;
  font-family: 'Fira Code', monospace;
  font-size: 0.85rem;
  overflow-x: auto;
}

.msg code {
  background: var(--bg-primary);
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Fira Code', monospace;
  font-size: 0.85em;
}

.chat-input-row {
  display: flex;
  gap: 8px;
  align-items: flex-end;
}

.chat-input {
  flex: 1;
  padding: 12px 16px;
  background: var(--bg-primary);
  border: 1px solid var(--border);
  border-radius: 12px;
  color: var(--text-primary);
  font-size: 0.95rem;
  font-family: inherit;
  resize: none;
  min-height: 44px;
  max-height: 120px;
  transition: var(--transition);
}

.chat-input:focus {
  outline: none;
  border-color: var(--accent-primary);
  box-shadow: 0 0 0 3px rgba(0, 217, 255, 0.1);
}

.btn-send {
  width: 44px;
  height: 44px;
  background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary));
  border: none;
  border-radius: 12px;
  color: var(--bg-primary);
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.2rem;
  transition: var(--transition);
}

.btn-send:hover { transform: scale(1.05); }
.btn-send:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-send .spinner {
  width: 16px;
  height: 16px;
  border: 2px solid transparent;
  border-top-color: var(--bg-primary);
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }

/* === Forms === */
.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}

.form-group { display: flex; flex-direction: column; gap: 6px; }

.form-label {
  color: var(--text-secondary);
  font-size: 0.85rem;
  font-weight: 500;
}

.form-input, .form-select {
  padding: 10px 12px;
  background: var(--bg-primary);
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--text-primary);
  font-size: 0.9rem;
  font-family: inherit;
  transition: var(--transition);
}

.form-input:focus, .form-select:focus {
  outline: none;
  border-color: var(--accent-primary);
}

.btn {
  padding: 10px 20px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--text-primary);
  font-weight: 500;
  cursor: pointer;
  transition: var(--transition);
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-family: inherit;
  font-size: 0.9rem;
}

.btn:hover { border-color: var(--accent-primary); color: var(--accent-primary); }
.btn:active { transform: scale(0.98); }

.btn-primary {
  background: linear-gradient(135deg, var(--accent-primary), #0099cc);
  color: var(--bg-primary);
  border: none;
}

.btn-primary:hover { color: var(--bg-primary); }

.btn-danger {
  background: rgba(248, 113, 113, 0.1);
  border-color: var(--error);
  color: var(--error);
}

.btn-danger:hover { background: var(--error); color: var(--bg-primary); }

/* === Output === */
.output {
  background: var(--bg-primary);
  padding: 14px;
  border-radius: 8px;
  font-family: 'Fira Code', 'Monaco', monospace;
  font-size: 0.85rem;
  min-height: 100px;
  max-height: 300px;
  overflow: auto;
  white-space: pre-wrap;
  word-wrap: break-word;
  border: 1px solid var(--border);
  color: var(--text-primary);
  line-height: 1.5;
}

.output-line {
  padding: 2px 0;
  display: flex;
  gap: 8px;
  align-items: center;
}

.output-line .linenum { color: var(--text-muted); user-select: none; min-width: 30px; }

/* === Events === */
.events-list {
  max-height: 400px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.event-item {
  padding: 10px 14px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 0.85rem;
  transition: var(--transition);
  animation: msgIn 0.3s ease-out;
}

.event-item:hover { background: var(--bg-secondary); }

.event-type {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  background: var(--accent-primary);
  color: var(--bg-primary);
  text-transform: uppercase;
}

.event-type.info { background: var(--info); }
.event-type.success { background: var(--success); }
.event-type.warning { background: var(--warning); }
.event-type.error { background: var(--error); }

.event-time {
  color: var(--text-muted);
  font-size: 0.75rem;
  font-family: 'Fira Code', monospace;
}

/* === Tools === */
.tool-search {
  margin-bottom: 16px;
}

.tools-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 12px;
}

.tool-card {
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 14px;
  transition: var(--transition);
  cursor: pointer;
}

.tool-card:hover {
  border-color: var(--accent-primary);
  transform: translateY(-2px);
  box-shadow: var(--shadow);
}

.tool-name {
  font-weight: 600;
  color: var(--accent-primary);
  margin-bottom: 4px;
  font-family: 'Fira Code', monospace;
}

.tool-desc {
  color: var(--text-secondary);
  font-size: 0.85rem;
  line-height: 1.4;
}

/* === Toast === */
.toast-container {
  position: fixed;
  top: 80px;
  right: 20px;
  z-index: 1000;
  display: flex;
  flex-direction: column;
  gap: 8px;
  pointer-events: none;
}

.toast {
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 12px 16px;
  min-width: 250px;
  box-shadow: var(--shadow);
  display: flex;
  align-items: center;
  gap: 10px;
  animation: slideIn 0.3s ease-out;
  pointer-events: auto;
}

.toast.success { border-left: 3px solid var(--success); }
.toast.error { border-left: 3px solid var(--error); }
.toast.warning { border-left: 3px solid var(--warning); }
.toast.info { border-left: 3px solid var(--info); }

@keyframes slideIn {
  from { opacity: 0; transform: translateX(100%); }
  to { opacity: 1; transform: translateX(0); }
}

.toast.hide { animation: slideOut 0.3s ease-in forwards; }

@keyframes slideOut {
  to { opacity: 0; transform: translateX(100%); }
}

/* === Loading === */
.skeleton {
  background: linear-gradient(90deg, var(--bg-tertiary) 25%, var(--bg-secondary) 50%, var(--bg-tertiary) 75%);
  background-size: 200% 100%;
  animation: skeleton 1.5s ease-in-out infinite;
  border-radius: 6px;
  height: 16px;
}

@keyframes skeleton {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

/* === Responsive === */
@media (max-width: 768px) {
  .main { grid-template-columns: 1fr; }
  .sidebar { display: none; }
  .header { padding: 12px 16px; }
  .content { padding: 16px; }
  .stat-grid { grid-template-columns: repeat(2, 1fr); }
  .stat-value { font-size: 1.6rem; }
  .msg { max-width: 90%; }
  .chat-container { height: 60vh; }
}

.hidden { display: none !important; }

/* === Mobile Nav === */
.mobile-nav-toggle {
  display: none;
  background: none;
  border: none;
  color: var(--text-primary);
  font-size: 1.4rem;
  cursor: pointer;
}

@media (max-width: 768px) {
  .mobile-nav-toggle { display: block; }
  .sidebar.mobile-open {
    display: block;
    position: fixed;
    top: 60px;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 50;
  }
}
</style>
</head>
<body>
<header class="header">
  <div class="logo">
    <span class="logo-icon">☕</span>
    <h1>IcePoint Coffee</h1>
  </div>
  <div class="header-actions">
    <span class="status-badge" id="statusBadge">
      <span class="status-dot" id="statusDot"></span>
      <span id="statusText">连接中...</span>
    </span>
    <button class="icon-btn" onclick="refreshStatus()" title="刷新">🔄</button>
    <button class="icon-btn" onclick="toggleTheme()" title="主题">🌓</button>
    <button class="mobile-nav-toggle" onclick="toggleMobileNav()">☰</button>
  </div>
</header>

<div class="toast-container" id="toastContainer"></div>

<div class="main">
<aside class="sidebar" id="sidebar">
  <nav>
    <a class="nav-item active" data-tab="status"><span class="nav-icon">📊</span><span>状态</span></a>
    <a class="nav-item" data-tab="ai"><span class="nav-icon">🤖</span><span>AI 对话</span></a>
    <a class="nav-item" data-tab="commands"><span class="nav-icon">💬</span><span>命令</span></a>
    <a class="nav-item" data-tab="events"><span class="nav-icon">📡</span><span>事件</span></a>
    <a class="nav-item" data-tab="build"><span class="nav-icon">🏠</span><span>建筑</span></a>
    <a class="nav-item" data-tab="plugins"><span class="nav-icon">🔌</span><span>插件</span></a>
    <a class="nav-item" data-tab="tools"><span class="nav-icon">🛠️</span><span>工具</span></a>
    <a class="nav-item" data-tab="about"><span class="nav-icon">ℹ️</span><span>关于</span></a>
  </nav>
</aside>
<main class="content">
  <div id="tabStatus" class="tab-content">
    <div class="stat-grid" id="statGrid">
      <div class="stat"><div class="stat-value" id="statUptime">--</div><div class="stat-label">运行时间</div></div>
      <div class="stat"><div class="stat-value" id="statCmds">0</div><div class="stat-label">命令数</div></div>
      <div class="stat"><div class="stat-value" id="statEvents">0</div><div class="stat-label">事件数</div></div>
      <div class="stat"><div class="stat-value" id="statPlugins">0</div><div class="stat-label">插件数</div></div>
      <div class="stat"><div class="stat-value" id="statTools">0</div><div class="stat-label">工具数</div></div>
      <div class="stat"><div class="stat-value" id="statAI">❌</div><div class="stat-label">AI 状态</div></div>
    </div>
    <div class="card">
      <div class="card-header">
        <div class="card-title">🔄 快速操作</div>
      </div>
      <div style="display:flex;gap:8px;flex-wrap:wrap">
        <button class="btn btn-primary" onclick="reload()">🔄 重载配置</button>
        <button class="btn" onclick="restart()">🔄 重启服务</button>
        <button class="btn btn-danger" onclick="confirmAction('确认清空事件?', clearEvents)">🗑️ 清空事件</button>
      </div>
    </div>
  </div>

  <div id="tabAI" class="tab-content hidden">
    <div class="card">
      <div class="card-header">
        <div class="card-title">🤖 AI 对话</div>
        <button class="btn" onclick="clearChat()">清空</button>
      </div>
      <div class="chat-container">
        <div class="chat-messages" id="chatMessages">
          <div class="msg system">欢迎使用 IcePoint Coffee AI！</div>
        </div>
        <div class="chat-input-row">
          <textarea class="chat-input" id="chatInput" placeholder="输入消息... (Enter 发送, Shift+Enter 换行)" rows="1"></textarea>
          <button class="btn-send" id="sendBtn" onclick="sendChat()">➤</button>
        </div>
      </div>
    </div>
  </div>

  <div id="tabCommands" class="tab-content hidden">
    <div class="card">
      <div class="card-header">
        <div class="card-title">💬 命令执行</div>
      </div>
      <div class="form-grid">
        <div class="form-group">
          <label class="form-label">目标</label>
          <input class="form-input" id="cmdHost" value="localhost" placeholder="服务器地址">
        </div>
        <div class="form-group" style="grid-column: span 2">
          <label class="form-label">命令</label>
          <input class="form-input" id="cmdText" placeholder="say Hello">
        </div>
      </div>
      <div style="display:flex;gap:8px">
        <button class="btn btn-primary" onclick="sendCommand()">执行</button>
        <button class="btn" onclick="document.getElementById('cmdText').value=''">清空</button>
      </div>
      <div class="output" id="cmdOutput" style="margin-top:12px">等待执行...</div>
    </div>
  </div>

  <div id="tabEvents" class="tab-content hidden">
    <div class="card">
      <div class="card-header">
        <div class="card-title">📡 实时事件</div>
        <button class="btn" onclick="clearEvents()">清空</button>
      </div>
      <div class="events-list" id="eventsList"></div>
    </div>
  </div>

  <div id="tabBuild" class="tab-content hidden">
    <div class="card">
      <div class="card-header">
        <div class="card-title">🏠 建筑生成器</div>
      </div>
      <div class="form-grid">
        <div class="form-group">
          <label class="form-label">类型</label>
          <select class="form-select" id="buildType">
            <option value="house">🏠 房子 House</option>
            <option value="tower">🗼 塔 Tower</option>
            <option value="circle">⭕ 圆形 Circle</option>
            <option value="sphere">🔵 球体 Sphere</option>
            <option value="wall">🧱 墙 Wall</option>
            <option value="floor">🟫 地板 Floor</option>
            <option value="rect">⬜ 矩形 Rect</option>
          </select>
        </div>
        <div class="form-group">
          <label class="form-label">尺寸</label>
          <input class="form-input" id="buildSize" type="number" value="10" min="1" max="100">
        </div>
        <div class="form-group">
          <label class="form-label">X 坐标</label>
          <input class="form-input" id="buildX" type="number" value="0">
        </div>
        <div class="form-group">
          <label class="form-label">Y 坐标</label>
          <input class="form-input" id="buildY" type="number" value="64">
        </div>
        <div class="form-group">
          <label class="form-label">Z 坐标</label>
          <input class="form-input" id="buildZ" type="number" value="0">
        </div>
      </div>
      <div style="display:flex;gap:8px">
        <button class="btn btn-primary" onclick="buildStructure()">⚡ 生成建筑</button>
        <button class="btn" onclick="randomBuild()">🎲 随机生成</button>
      </div>
      <div class="output" id="buildOutput" style="margin-top:12px">等待生成...</div>
    </div>
  </div>

  <div id="tabPlugins" class="tab-content hidden">
    <div class="card">
      <div class="card-header">
        <div class="card-title">🔌 插件管理</div>
      </div>
      <div class="form-grid">
        <div class="form-group">
          <label class="form-label">插件名</label>
          <input class="form-input" id="pluginName" placeholder="my-plugin">
        </div>
      </div>
      <div style="display:flex;gap:8px">
        <button class="btn btn-primary" onclick="registerPlugin()">➕ 注册</button>
        <button class="btn" onclick="loadPlugins()">🔄 刷新</button>
      </div>
      <div class="output" id="pluginList" style="margin-top:12px">暂无插件</div>
    </div>
  </div>

  <div id="tabTools" class="tab-content hidden">
    <div class="card">
      <div class="card-header">
        <div class="card-title">🛠️ 可用工具</div>
      </div>
      <div class="tool-search">
        <input class="form-input" id="toolSearch" placeholder="🔍 搜索工具..." oninput="filterTools()">
      </div>
      <div class="tools-grid" id="toolsList">加载中...</div>
    </div>
  </div>

  <div id="tabAbout" class="tab-content hidden">
    <div class="card">
      <div class="card-header">
        <div class="card-title">☕ 关于 IcePoint Coffee</div>
      </div>
      <div style="line-height:1.8">
        <p><strong>版本:</strong> 1.0.0</p>
        <p><strong>协议:</strong> MIT</p>
        <p><strong>自研组件:</strong> Raknet / FB Auth / ECDH / MC Protocol / Builder / AI Agent / Dashboard</p>
        <p><strong>键盘快捷键:</strong></p>
        <ul style="margin-left:20px;color:var(--text-secondary)">
          <li><kbd>1</kbd>~<kbd>8</kbd> 切换 Tab</li>
          <li><kbd>Enter</kbd> 发送消息（AI）</li>
          <li><kbd>Shift</kbd>+<kbd>Enter</kbd> 换行</li>
          <li><kbd>Ctrl</kbd>+<kbd>R</kbd> 刷新状态</li>
          <li><kbd>Ctrl</kbd>+<kbd>K</kbd> 清空聊天</li>
        </ul>
        <p style="margin-top:16px;color:var(--text-muted);font-size:0.85rem">
          Made with ❤️ by Verify144 · © 2026
        </p>
      </div>
    </div>
  </div>
</main>
</div>

<script>
const API = '/api/v1';
let eventSource = null;
let allTools = [];
let aiBusy = false;
let isConnected = true;

const $=id=>document.getElementById(id);

async function api(path, method, body) {
  method = method || 'GET';
  const opts = {method: method, headers: {'Content-Type': 'application/json'}};
  if (body) opts.body = JSON.stringify(body);
  try {
    const r = await fetch(API + path, opts);
    if (!r.ok) throw new Error('HTTP ' + r.status);
    return await r.json();
  } catch (e) {
    showToast('请求失败: ' + e.message, 'error');
    throw e;
  }
}

function showToast(message, type) {
  type = type || 'info';
  const container = $('toastContainer');
  const toast = document.createElement('div');
  toast.className = 'toast ' + type;
  const icons = {success: '✓', error: '✕', warning: '⚠', info: 'ℹ'};
  toast.innerHTML = '<span style="font-size:1.2rem">' + (icons[type] || 'ℹ') + '</span><span>' + escapeHtml(message) + '</span>';
  container.appendChild(toast);
  setTimeout(function() {
    toast.classList.add('hide');
    setTimeout(function() { toast.remove(); }, 300);
  }, 3000);
}

function confirmAction(message, callback) {
  if (confirm(message)) callback();
}

function showTab(name) {
  document.querySelectorAll('.tab-content').forEach(function(el) { el.classList.add('hidden'); });
  document.querySelectorAll('.nav-item').forEach(function(el) { el.classList.remove('active'); });
  $('tab' + name.charAt(0).toUpperCase() + name.slice(1)).classList.remove('hidden');
  const navItem = document.querySelector('.nav-item[data-tab="' + name + '"]');
  if (navItem) navItem.classList.add('active');
  $('sidebar').classList.remove('mobile-open');
}

document.querySelectorAll('.nav-item').forEach(function(a) {
  a.addEventListener('click', function(e) {
    e.preventDefault();
    const tab = a.dataset.tab;
    showTab(tab);
    if (tab === 'status') refreshStatus();
    else if (tab === 'tools') loadTools();
    else if (tab === 'plugins') loadPlugins();
  });
});

async function refreshStatus() {
  try {
    const s = await api('/status');
    $('statUptime').textContent = s.uptime || '--';
    $('statCmds').textContent = s.commands || 0;
    $('statEvents').textContent = s.events || 0;
    $('statPlugins').textContent = s.plugins || 0;
    $('statTools').textContent = s.tools_count || 0;
    $('statAI').textContent = s.ai_enabled ? '✅' : '❌';
    $('statusText').textContent = s.ai_enabled ? '运行中 (AI)' : '运行中';
    $('statusDot').className = 'status-dot';
    isConnected = true;
  } catch (e) {
    $('statusText').textContent = '离线';
    $('statusDot').className = 'status-dot offline';
    isConnected = false;
  }
}

function getTimestamp() {
  const now = new Date();
  return now.toLocaleTimeString('zh-CN', {hour12: false});
}

function escapeHtml(s) {
  if (s === null || s === undefined) return '';
  return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
}

function renderMarkdown(text) {
  if (!text) return '';
  text = escapeHtml(text);
  text = text.replace(/\u0060([^\u0060]+)\u0060/g, '<code>$1</code>');
  text = text.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
  text = text.replace(/\n/g, '<br>');
  return text;
}

async function sendChat() {
  if (aiBusy) return;
  const input = $('chatInput');
  const text = input.value.trim();
  if (!text) return;

  aiBusy = true;
  const sendBtn = $('sendBtn');
  sendBtn.disabled = true;
  sendBtn.innerHTML = '<span class="spinner"></span>';

  const msgs = $('chatMessages');
  const userMsg = document.createElement('div');
  userMsg.className = 'msg user';
  userMsg.innerHTML = renderMarkdown(text) + '<div class="timestamp">' + getTimestamp() + '</div>';
  msgs.appendChild(userMsg);
  input.value = '';
  input.style.height = 'auto';
  msgs.scrollTop = msgs.scrollHeight;

  try {
    const r = await api('/ai/chat', 'POST', {message: text, use_tools: true});
    const aiMsg = document.createElement('div');
    aiMsg.className = 'msg ai';
    aiMsg.innerHTML = renderMarkdown(r.response || JSON.stringify(r)) + '<div class="timestamp">' + getTimestamp() + '</div>';
    msgs.appendChild(aiMsg);
    msgs.scrollTop = msgs.scrollHeight;
  } catch (e) {
    const errMsg = document.createElement('div');
    errMsg.className = 'msg error';
    errMsg.textContent = '错误: ' + e.message;
    msgs.appendChild(errMsg);
  } finally {
    aiBusy = false;
    sendBtn.disabled = false;
    sendBtn.textContent = '➤';
  }
}

function clearChat() {
  $('chatMessages').innerHTML = '<div class="msg system">聊天已清空</div>';
  showToast('聊天已清空', 'success');
}

function clearEvents() {
  $('eventsList').innerHTML = '';
  showToast('事件已清空', 'success');
}

async function sendCommand() {
  const cmd = $('cmdText').value.trim();
  if (!cmd) {
    showToast('请输入命令', 'warning');
    return;
  }
  $('cmdOutput').textContent = '执行中...';
  try {
    const r = await api('/commands', 'POST', {command: cmd});
    $('cmdOutput').textContent = JSON.stringify(r, null, 2);
    showToast('命令已发送', 'success');
  } catch (e) {
    $('cmdOutput').textContent = '执行失败: ' + e.message;
  }
}

async function buildStructure() {
  const out = $('buildOutput');
  out.textContent = '生成中...';
  try {
    const r = await api('/build', 'POST', {
      type: $('buildType').value,
      size: parseInt($('buildSize').value),
      x: parseInt($('buildX').value),
      y: parseInt($('buildY').value),
      z: parseInt($('buildZ').value)
    });
    out.textContent = r.result || JSON.stringify(r, null, 2);
    showToast('建筑已生成', 'success');
  } catch (e) {
    out.textContent = '生成失败: ' + e.message;
  }
}

function randomBuild() {
  const types = ['house', 'tower', 'circle', 'sphere', 'wall', 'floor', 'rect'];
  $('buildType').value = types[Math.floor(Math.random() * types.length)];
  $('buildSize').value = Math.floor(Math.random() * 15) + 5;
  $('buildX').value = Math.floor(Math.random() * 200) - 100;
  $('buildY').value = 64;
  $('buildZ').value = Math.floor(Math.random() * 200) - 100;
  buildStructure();
}

async function registerPlugin() {
  const name = $('pluginName').value.trim();
  if (!name) {
    showToast('请输入插件名', 'warning');
    return;
  }
  try {
    const r = await api('/plugins/register', 'POST', {name: name, type: 'default'});
    showToast(r.status + ': ' + name, 'success');
    $('pluginName').value = '';
    loadPlugins();
  } catch (e) {}
}

async function loadPlugins() {
  try {
    const r = await api('/plugins');
    const el = $('pluginList');
    if (!r.plugins || r.plugins.length === 0) {
      el.textContent = '暂无插件';
    } else {
      el.textContent = r.plugins.map(function(p) { return '• ' + p.name + ' (' + (p.enabled ? '启用' : '禁用') + ')'; }).join('\n');
    }
  } catch (e) {}
}

async function loadTools() {
  try {
    const r = await api('/ai/tools');
    allTools = r.openai || [];
    renderTools(allTools);
  } catch (e) {
    $('toolsList').innerHTML = '<div class="tool-card"><div class="tool-desc">加载失败</div></div>';
  }
}

function renderTools(tools) {
  const el = $('toolsList');
  if (tools.length === 0) {
    el.innerHTML = '<div class="tool-card"><div class="tool-desc">暂无工具</div></div>';
    return;
  }
  el.innerHTML = tools.map(function(t) {
    return '<div class="tool-card"><div class="tool-name">' + t.function.name + '</div><div class="tool-desc">' + escapeHtml(t.function.description) + '</div></div>';
  }).join('');
}

function filterTools() {
  const q = $('toolSearch').value.toLowerCase();
  if (!q) {
    renderTools(allTools);
    return;
  }
  const filtered = allTools.filter(function(t) {
    return t.function.name.toLowerCase().includes(q) || (t.function.description || '').toLowerCase().includes(q);
  });
  renderTools(filtered);
}

async function reload() {
  try {
    const r = await api('/admin/reload');
    showToast('配置已重载', 'success');
  } catch (e) {}
}

async function restart() {
  if (!confirm('确认重启服务?')) return;
  try {
    await api('/admin/restart');
    showToast('服务正在重启...', 'info');
  } catch (e) {}
}

function startSSE() {
  if (eventSource) eventSource.close();
  eventSource = new EventSource(API + '/events');
  eventSource.onmessage = function(e) {
    try {
      const ev = JSON.parse(e.data);
      if (ev.type === 'ping') return;
      addEvent(ev);
    } catch (err) {}
  };
  eventSource.onerror = function() {
    if (isConnected) {
      showToast('事件流断开，3秒后重连...', 'warning');
      isConnected = false;
    }
    setTimeout(startSSE, 3000);
  };
}

function addEvent(ev) {
  const list = $('eventsList');
  if (!list) return;
  const item = document.createElement('div');
  item.className = 'event-item';
  const typeClass = (ev.type || 'info').toLowerCase();
  item.innerHTML = '<span class="event-type ' + typeClass + '">' + ev.type + '</span><span class="event-time">' + getTimestamp() + '</span>';
  list.insertBefore(item, list.firstChild);
  if (list.children.length > 100) list.lastChild.remove();
}

function toggleTheme() {
  const root = document.documentElement;
  const isLight = root.style.getPropertyValue('--bg-primary') === '#f5f7fa';
  if (isLight) {
    root.style.setProperty('--bg-primary', '#0a0e1a');
    root.style.setProperty('--bg-secondary', '#121826');
    root.style.setProperty('--bg-tertiary', '#1a2138');
    root.style.setProperty('--text-primary', '#e8eaf0');
  } else {
    root.style.setProperty('--bg-primary', '#f5f7fa');
    root.style.setProperty('--bg-secondary', '#ffffff');
    root.style.setProperty('--bg-tertiary', '#eef2f7');
    root.style.setProperty('--text-primary', '#1a1a1a');
  }
  showToast('主题已切换', 'info');
}

function toggleMobileNav() {
  $('sidebar').classList.toggle('mobile-open');
}

// Chat input auto-resize
$('chatInput').addEventListener('input', function() {
  this.style.height = 'auto';
  this.style.height = Math.min(this.scrollHeight, 120) + 'px';
});

// Keyboard shortcuts
document.addEventListener('keydown', function(e) {
  if (e.ctrlKey && e.key === 'r') { e.preventDefault(); refreshStatus(); }
  if (e.ctrlKey && e.key === 'k') { e.preventDefault(); clearChat(); }
  if (!e.ctrlKey && !e.altKey && !e.metaKey && e.target.tagName !== 'INPUT' && e.target.tagName !== 'TEXTAREA') {
    const num = parseInt(e.key);
    if (num >= 1 && num <= 8) {
      const tabs = ['status', 'ai', 'commands', 'events', 'build', 'plugins', 'tools', 'about'];
      if (tabs[num - 1]) showTab(tabs[num - 1]);
    }
  }
});

$('chatInput').addEventListener('keydown', function(e) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    sendChat();
  }
});

$('cmdText').addEventListener('keydown', function(e) {
  if (e.key === 'Enter') { e.preventDefault(); sendCommand(); }
});

refreshStatus();
startSSE();
setInterval(refreshStatus, 5000);
setTimeout(function() { showToast('欢迎使用 IcePoint Coffee!', 'info'); }, 500);
</script>
</body>
</html>`
