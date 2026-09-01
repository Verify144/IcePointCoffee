// Package plugin 实现 HTTP RPC 插件系统。
// 插件以独立进程运行，通过 HTTP 通信。冰点咖啡作为主进程管理插件生命周期。
package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// Manager 插件管理器。
type Manager struct {
	dir      string
	mu       sync.RWMutex
	plugins  map[string]*Plugin
	client   *http.Client
}

// Plugin 插件信息。
type Plugin struct {
	ID          string
	Name        string
	Path        string
	Cmd         *exec.Cmd
	URL         string
	Enabled     bool
	LoadedAt    time.Time
	LastPing    time.Time
	Description string
}

// NewManager 创建插件管理器。
func NewManager(dir string) *Manager {
	return &Manager{
		dir: dir,
		plugins: make(map[string]*Plugin),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Scan 扫描插件目录，识别可用插件。
// 插件目录结构：
//   plugins/
//     my_plugin/
//       plugin.json     # 元信息
//       plugin(.exe)    # 可执行文件
//       config.yaml     # 可选配置
func (m *Manager) Scan() ([]*Plugin, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, err := os.Stat(m.dir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return nil, err
	}

	var plugins []*Plugin
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pluginPath := filepath.Join(m.dir, entry.Name())
		metaPath := filepath.Join(pluginPath, "plugin.json")
		if _, err := os.Stat(metaPath); os.IsNotExist(err) {
			continue
		}

		var meta struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Executable  string `json:"executable"`
			Port        int    `json:"port"`
		}
		data, err := os.ReadFile(metaPath)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}

		exeName := meta.Executable
		if exeName == "" {
			exeName = "plugin"
		}
		exePath := filepath.Join(pluginPath, exeName)
		if _, err := os.Stat(exePath); os.IsNotExist(err) {
			// 尝试带 .exe 后缀
			if _, err := os.Stat(exePath + ".exe"); err == nil {
				exePath = exePath + ".exe"
			} else {
				continue
			}
		}

		p := &Plugin{
			ID:          entry.Name(),
			Name:        meta.Name,
			Path:        pluginPath,
			URL:         fmt.Sprintf("http://127.0.0.1:%d", meta.Port),
			Enabled:     true,
			Description: meta.Description,
		}

		m.plugins[p.ID] = p
		plugins = append(plugins, p)
	}

	return plugins, nil
}

// Start 启动指定插件进程。
func (m *Manager) Start(id string) error {
	m.mu.Lock()
	p, ok := m.plugins[id]
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("插件不存在: %s", id)
	}

	exeName := "plugin"
	if _, err := os.Stat(filepath.Join(p.Path, exeName+".exe")); err == nil {
		exeName += ".exe"
	}
	exePath := filepath.Join(p.Path, exeName)

	cmd := exec.Command(exePath)
	cmd.Dir = p.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动插件失败: %w", err)
	}

	m.mu.Lock()
	p.Cmd = cmd
	p.LoadedAt = time.Now()
	m.mu.Unlock()

	return nil
}

// Stop 停止插件。
func (m *Manager) Stop(id string) error {
	m.mu.Lock()
	p, ok := m.plugins[id]
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("插件不存在: %s", id)
	}
	if p.Cmd == nil || p.Cmd.Process == nil {
		return nil
	}
	return p.Cmd.Process.Kill()
}

// Call 调用插件的 HTTP RPC。
// 端点: POST {url}/rpc
// 请求体: { "method": "...", "params": {...} }
// 响应体: { "result": ..., "error": "..." }
func (m *Manager) Call(ctx context.Context, id, method string, params any) (any, error) {
	m.mu.RLock()
	p, ok := m.plugins[id]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("插件不存在: %s", id)
	}

	reqBody, err := json.Marshal(map[string]any{
		"method": method,
		"params": params,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.URL+"/rpc", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rpcResp struct {
		Result any    `json:"result"`
		Error  string `json:"error"`
	}
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if rpcResp.Error != "" {
		return nil, fmt.Errorf("插件错误: %s", rpcResp.Error)
	}

	return rpcResp.Result, nil
}

// List 列出所有插件。
func (m *Manager) List() []*Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*Plugin
	for _, p := range m.plugins {
		list = append(list, p)
	}
	return list
}

// Get 获取单个插件。
func (m *Manager) Get(id string) (*Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.plugins[id]
	return p, ok
}
