package server

import (
	"net/http"
	"strings"
)

// handleTasksRoot 任务根路由
// GET  /api/v1/tasks                 - 列表
// GET  /api/v1/tasks/stats           - 统计
// POST /api/v1/tasks                 - 提交
// GET  /api/v1/tasks/{id}            - 详情
// POST /api/v1/tasks/{id}/cancel     - 取消
func (s *Server) handleTasksRoot(w http.ResponseWriter, r *http.Request) {
	// StripPrefix 之后到达此处的 path 形如：
	//   /                  -> 列表/提交
	//   /stats             -> 统计
	//   /{id}              -> 详情
	//   /{id}/cancel       -> 取消
	path := strings.TrimPrefix(r.URL.Path, "/")
	// 此时 path 应该是 "tasks", "tasks/stats", "tasks/xxx", "tasks/xxx/cancel" 等

	// /tasks 或 /tasks/xxx -> 列表 (GET) 或 提交 (POST)
	if path == "tasks" || path == "tasks/" {
		switch r.Method {
		case http.MethodGet:
			s.handleTaskList(w, r)
		case http.MethodPost:
			s.handleTaskSubmit(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// /tasks/stats -> 统计
	if path == "tasks/stats" {
		s.handleTaskStats(w, r)
		return
	}

	// /tasks/{id}/cancel -> 取消
	if strings.HasSuffix(path, "/cancel") {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimSuffix(path, "/cancel")
		// 去掉 "tasks/" 前缀
		taskID := strings.TrimPrefix(id, "tasks/")
		t, err := s.taskManager.Get(taskID)
		if err != nil {
			SendError(w, http.StatusNotFound, 404, "任务不存在", err.Error())
			return
		}
		if err := s.taskManager.Cancel(t.ID); err != nil {
			SendError(w, http.StatusInternalServerError, 500, "取消失败", err.Error())
			return
		}
		SendSuccess(w, map[string]string{"status": "cancelled", "id": t.ID})
		return
	}

	// /tasks/{id} -> 详情 (GET)
	if strings.HasPrefix(path, "tasks/") {
		rest := strings.TrimPrefix(path, "tasks/")
		if rest == "" || rest == "cancel" {
			// 无效的路径
			return
		}
		if r.Method == http.MethodGet {
			t, err := s.taskManager.Get(rest)
			if err != nil {
				SendError(w, http.StatusNotFound, 404, "任务不存在", err.Error())
				return
			}
			SendSuccess(w, t)
			return
		}
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
