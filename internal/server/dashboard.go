package server

const dashboardHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>IcePoint Coffee - Dashboard</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#0a0a0f;color:#e0e0e0;min-height:100vh}
.header{background:linear-gradient(135deg,#1a1a2e,#16213e);padding:20px 30px;border-bottom:2px solid #0f3460;display:flex;align-items:center;justify-content:space-between}
.header h1{color:#e94560;font-size:1.5rem;font-weight:700}
.header .badge{background:#0f3460;color:#00d9ff;padding:4px 12px;border-radius:20px;font-size:0.8rem}
.main{display:grid;grid-template-columns:280px 1fr;min-height:calc(100vh - 70px)}
.sidebar{background:#12121a;padding:20px;border-right:1px solid #1a1a2e}
.sidebar nav a{display:block;padding:12px 16px;color:#888;border-radius:8px;margin-bottom:4px;text-decoration:none;transition:.2s}
.sidebar nav a:hover,.sidebar nav a.active{background:#1a1a2e;color:#00d9ff}
.content{padding:24px;overflow-y:auto}
.card{background:#12121a;border:1px solid #1a1a2e;border-radius:12px;padding:20px;margin-bottom:20px}
.card h2{color:#00d9ff;font-size:1.1rem;margin-bottom:16px;display:flex;align-items:center;gap:8px}
.stat-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(150px,1fr));gap:16px;margin-bottom:20px}
.stat{background:#1a1a2e;padding:16px;border-radius:8px;text-align:center}
.stat .value{font-size:2rem;font-weight:700;color:#e94560}
.stat .label{color:#888;font-size:0.85rem;margin-top:4px}
.chat-container{height:400px;display:flex;flex-direction:column}
.chat-messages{flex:1;overflow-y:auto;padding:12px;background:#0a0a0f;border-radius:8px;margin-bottom:12px}
.chat-messages .msg{margin-bottom:12px;padding:10px 14px;border-radius:8px;max-width:85%}
.chat-messages .user{background:#0f3460;margin-left:auto;color:#00d9ff}
.chat-messages .ai{background:#1a1a2e;color:#e0e0e0}
.chat-messages .system{background:#2a1a1a;color:#e94560;font-size:0.85rem}
.chat-input{display:flex;gap:8px}
.chat-input input{flex:1;padding:12px 16px;background:#0a0a0f;border:1px solid #1a1a2e;border-radius:8px;color:#e0e0e0;font-size:1rem}
.chat-input input:focus{outline:none;border-color:#00d9ff}
.chat-input button{padding:12px 24px;background:linear-gradient(135deg,#e94560,#ff6b6b);border:none;border-radius:8px;color:#fff;font-weight:600;cursor:pointer}
.chat-input button:hover{opacity:.9}
.events-list{max-height:300px;overflow-y:auto}
.event-item{padding:10px 12px;border-bottom:1px solid #1a1a2e;display:flex;justify-content:space-between;font-size:0.85rem}
.event-item .type{color:#00d9ff;font-weight:600}
.event-item .time{color:#555}
.form-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:16px;margin-bottom:16px}
.form-group{display:flex;flex-direction:column;gap:6px}
.form-group label{color:#888;font-size:0.85rem}
.form-group input,.form-group select{padding:10px 12px;background:#0a0a0f;border:1px solid #1a1a2e;border-radius:6px;color:#e0e0e0}
.btn{background:linear-gradient(135deg,#00d9ff,#0099cc);padding:10px 20px;border:none;border-radius:6px;color:#fff;font-weight:600;cursor:pointer;margin-right:8px}
.btn:hover{opacity:.9}
.output{background:#0a0a0f;padding:12px;border-radius:6px;font-family:'Fira Code',monospace;font-size:0.9rem;min-height:100px;white-space:pre-wrap}
.hidden{display:none}
</style>
</head>
<body>
<header class="header">
<h1>☕ IcePoint Coffee</h1>
<span class="badge" id="statusBadge">● 初始化中</span>
</header>
<div class="main">
<aside class="sidebar">
<nav>
<a href="#" class="active" data-tab="status">📊 状态</a>
<a href="#" data-tab="ai">🤖 AI 对话</a>
<a href="#" data-tab="commands">💬 命令</a>
<a href="#" data-tab="events">📡 事件</a>
<a href="#" data-tab="build">🏠 建筑</a>
<a href="#" data-tab="plugins">🔌 插件</a>
<a href="#" data-tab="tools">🛠️ 工具</a>
</nav>
</aside>
<main class="content" id="mainContent">
<div id="tabStatus" class="tab-content">
<div class="stat-grid">
<div class="stat"><div class="value" id="statUptime">--</div><div class="label">运行时间</div></div>
<div class="stat"><div class="value" id="statCmds">0</div><div class="label">命令数</div></div>
<div class="stat"><div class="value" id="statEvents">0</div><div class="label">事件数</div></div>
<div class="stat"><div class="value" id="statPlugins">0</div><div class="label">插件数</div></div>
<div class="stat"><div class="value" id="statTools">0</div><div class="label">工具数</div></div>
<div class="stat"><div class="value" id="statAI">❌</div><div class="label">AI 状态</div></div>
</div>
<div class="card"><h2>🔄 快速操作</h2>
<button class="btn" onclick="reload()">🔄 重载配置</button>
<button class="btn" onclick="restart()">🔄 重启服务</button>
<button class="btn" onclick="refreshStatus()">🔍 刷新状态</button>
</div>
</div>

<div id="tabAI" class="tab-content hidden">
<div class="card">
<h2>🤖 AI 对话</h2>
<div class="chat-container">
<div class="chat-messages" id="chatMessages">
<div class="msg system">欢迎使用 IcePoint Coffee AI！当前为 Mock 模式。</div>
</div>
<div class="chat-input">
<input type="text" id="chatInput" placeholder="输入消息..." onkeypress="if(event.key==='Enter')sendChat()">
<button onclick="sendChat()">发送</button>
</div>
</div>
</div>
</div>

<div id="tabCommands" class="tab-content hidden">
<div class="card">
<h2>💬 命令执行</h2>
<div class="form-grid">
<div class="form-group"><label>服务器地址</label><input id="cmdHost" value="localhost"></div>
<div class="form-group"><label>命令</label><input id="cmdText" placeholder="say Hello" onkeypress="if(event.key==='Enter')sendCommand()"></div>
</div>
<button class="btn" onclick="sendCommand()">执行命令</button>
<div class="output" id="cmdOutput">等待执行...</div>
</div>
</div>

<div id="tabEvents" class="tab-content hidden">
<div class="card">
<h2>📡 实时事件</h2>
<div class="events-list" id="eventsList"></div>
</div>
</div>

<div id="tabBuild" class="tab-content hidden">
<div class="card">
<h2>🏠 建筑生成器</h2>
<div class="form-grid">
<div class="form-group"><label>类型</label><select id="buildType"><option value="house">房子 House</option><option value="tower">塔 Tower</option><option value="circle">圆形 Circle</option><option value="sphere">球体 Sphere</option><option value="wall">墙 Wall</option><option value="floor">地板 Floor</option><option value="rect">矩形 Rect</option></select></div>
<div class="form-group"><label>尺寸</label><input id="buildSize" type="number" value="10"></div>
<div class="form-group"><label>X 坐标</label><input id="buildX" type="number" value="0"></div>
<div class="form-group"><label>Y 坐标</label><input id="buildY" type="number" value="64"></div>
<div class="form-group"><label>Z 坐标</label><input id="buildZ" type="number" value="0"></div>
</div>
<button class="btn" onclick="buildStructure()">生成建筑</button>
<div class="output" id="buildOutput">等待生成...</div>
</div>
</div>

<div id="tabPlugins" class="tab-content hidden">
<div class="card">
<h2>🔌 插件管理</h2>
<div class="form-grid">
<div class="form-group"><label>插件名称</label><input id="pluginName" placeholder="my-plugin"></div>
</div>
<button class="btn" onclick="registerPlugin()">注册插件</button>
<div id="pluginList" class="output">暂无插件</div>
</div>
</div>

<div id="tabTools" class="tab-content hidden">
<div class="card">
<h2>🛠️ 可用工具</h2>
<div class="output" id="toolsList">加载中...</div>
</div>
</div>
</div>
</div>

<script>
const API = '/api/v1';
let eventSource;

async function api(path, method, body) {
  method = method || 'GET';
  const opts = {method: method, headers:{'Content-Type':'application/json'}};
  if (body) opts.body = JSON.stringify(body);
  const r = await fetch(API + path, opts);
  return r.json();
}

function showTab(name) {
  document.querySelectorAll('.tab-content').forEach(function(el) { el.classList.add('hidden'); });
  document.querySelectorAll('.sidebar nav a').forEach(function(el) { el.classList.remove('active'); });
  document.getElementById('tab' + name.charAt(0).toUpperCase() + name.slice(1)).classList.remove('hidden');
}

document.querySelectorAll('.sidebar nav a').forEach(function(a) {
  a.addEventListener('click', function(e) {
    e.preventDefault();
    const tab = a.dataset.tab;
    showTab(tab);
    if (tab === 'status') refreshStatus();
    if (tab === 'tools') loadTools();
    if (tab === 'plugins') loadPlugins();
  });
});

async function refreshStatus() {
  const s = await api('/status');
  document.getElementById('statUptime').textContent = s.uptime || '--';
  document.getElementById('statCmds').textContent = s.commands || 0;
  document.getElementById('statEvents').textContent = s.events || 0;
  document.getElementById('statPlugins').textContent = s.plugins || 0;
  document.getElementById('statTools').textContent = s.tools_count || 0;
  document.getElementById('statAI').textContent = s.ai_enabled ? '✅' : '❌';
  document.getElementById('statusBadge').textContent = s.ai_enabled ? '● 运行中 (AI)' : '● 运行中';
}

async function sendChat() {
  const input = document.getElementById('chatInput');
  const msgs = document.getElementById('chatMessages');
  const text = input.value.trim();
  if (!text) return;
  msgs.innerHTML += '<div class="msg user">' + escapeHtml(text) + '</div>';
  input.value = '';
  const r = await api('/ai/chat', 'POST', {message: text, use_tools: true});
  msgs.innerHTML += '<div class="msg ai">' + escapeHtml(r.response || JSON.stringify(r)) + '</div>';
  msgs.scrollTop = msgs.scrollHeight;
}

async function sendCommand() {
  const cmd = document.getElementById('cmdText').value;
  const out = document.getElementById('cmdOutput');
  out.textContent = '执行中...';
  const r = await api('/commands', 'POST', {command: cmd});
  out.textContent = JSON.stringify(r, null, 2);
}

async function buildStructure() {
  const out = document.getElementById('buildOutput');
  out.textContent = '生成中...';
  const r = await api('/build', 'POST', {
    type: document.getElementById('buildType').value,
    size: parseInt(document.getElementById('buildSize').value),
    x: parseInt(document.getElementById('buildX').value),
    y: parseInt(document.getElementById('buildY').value),
    z: parseInt(document.getElementById('buildZ').value)
  });
  out.textContent = r.result || JSON.stringify(r, null, 2);
}

async function registerPlugin() {
  const name = document.getElementById('pluginName').value;
  const r = await api('/plugins/register', 'POST', {name: name, type: 'default'});
  alert(r.status + ': ' + name);
  loadPlugins();
}

async function loadPlugins() {
  const r = await api('/plugins');
  const el = document.getElementById('pluginList');
  if (!r.plugins || r.plugins.length === 0) {
    el.textContent = '暂无插件';
  } else {
    el.textContent = r.plugins.map(function(p) { return '• ' + p.name + ' (' + (p.enabled ? '启用' : '禁用') + ')'; }).join('\n');
  }
}

async function loadTools() {
  const r = await api('/ai/tools');
  const el = document.getElementById('toolsList');
  const tools = r.openai || [];
  el.textContent = tools.map(function(t) { return '🛠️ ' + t.function.name + '\n   ' + t.function.description; }).join('\n\n');
}

async function reload() { alert(JSON.stringify(await api('/admin/reload'))); }
async function restart() { alert(JSON.stringify(await api('/admin/restart'))); }

function startSSE() {
  eventSource = new EventSource(API + '/events');
  eventSource.onmessage = function(e) {
    try {
      const ev = JSON.parse(e.data);
      if (ev.type === 'ping') return;
      const list = document.getElementById('eventsList');
      if (!list) return;
      const time = new Date(ev.timestamp).toLocaleTimeString();
      list.innerHTML = '<div class="event-item"><span class="type">' + ev.type + '</span><span>' + time + '</span></div>' + list.innerHTML;
      if (list.children.length > 50) list.lastChild.remove();
    } catch(err) {}
  };
}

function escapeHtml(s) {
  return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
}

refreshStatus();
startSSE();
setInterval(refreshStatus, 5000);
</script>
</body>
</html>`
