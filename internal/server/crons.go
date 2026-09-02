package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Verify144/IcePointCoffee/internal/cron"
	"github.com/Verify144/IcePointCoffee/internal/task"
)

// handleCronRoutes 注册 Cron 相关路由
func (s *Server) handleCronRoutes(api *http.ServeMux) {
	// 所有 cron 端点都需要认证
	api.HandleFunc("/crons", s.handleCronsList)
	api.Handle("/crons/", http.StripPrefix("/crons", s.cronHandlers()))
}

// cronHandlers 返回 cron 子路由 mux
func (s *Server) cronHandlers() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleCronsItem)
	return mux
}

func (s *Server) handleCronsList(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.cronsList(w, r)
	case http.MethodPost:
		s.cronsCreate(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) cronsList(w http.ResponseWriter, r *http.Request) {
	user := s.authUser(r)
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized", "auth required")
		return
	}

	jobs, err := s.cronStore.List(user.ID, false)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "list failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"jobs":  jobs,
		"count": len(jobs),
	})
}

func (s *Server) cronsCreate(w http.ResponseWriter, r *http.Request) {
	user := s.authUser(r)
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized", "auth required")
		return
	}

	var req struct {
		Name     string                 `json:"name"`
		CronExpr string                 `json:"cron_expr"`
		TaskType string                 `json:"task_type"`
		Payload  map[string]interface{} `json:"payload"`
		Enabled  *bool                  `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json", err.Error())
		return
	}
	if req.Name == "" || req.CronExpr == "" || req.TaskType == "" {
		writeJSONError(w, http.StatusBadRequest, "missing fields", "name, cron_expr, task_type required")
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	job := &cron.CronJob{
		UserID:    user.ID,
		Name:      req.Name,
		CronExpr:  req.CronExpr,
		TaskType:  req.TaskType,
		Payload:   req.Payload,
		Enabled:   enabled,
		CreatedAt: time.Now(),
	}

	if err := s.cronStore.Create(job); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "create failed", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, job)
}

func (s *Server) handleCronsItem(w http.ResponseWriter, r *http.Request) {
	// path is already stripped of /crons prefix
	path := r.URL.Path
	rest := strings.TrimPrefix(path, "/")
	parts := strings.Split(rest, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	id := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch r.Method {
	case http.MethodGet:
		s.cronsGet(w, r, id)
	case http.MethodDelete:
		s.cronsDelete(w, r, id)
	case http.MethodPost:
		switch action {
		case "enable":
			s.cronsSetEnabled(w, r, id, true)
		case "disable":
			s.cronsSetEnabled(w, r, id, false)
		case "trigger":
			s.cronsTrigger(w, r, id)
		default:
			http.Error(w, "Unknown action", http.StatusNotFound)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) cronsGet(w http.ResponseWriter, r *http.Request, id string) {
	user := s.authUser(r)
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	job, err := s.cronStore.Get(id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found", err.Error())
		return
	}
	if job.UserID != user.ID {
		writeJSONError(w, http.StatusForbidden, "forbidden", "not owner")
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) cronsDelete(w http.ResponseWriter, r *http.Request, id string) {
	user := s.authUser(r)
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	if err := s.cronStore.Delete(id, user.ID); err != nil {
		writeJSONError(w, http.StatusNotFound, "delete failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted", "id": id})
}

func (s *Server) cronsSetEnabled(w http.ResponseWriter, r *http.Request, id string, enabled bool) {
	user := s.authUser(r)
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	job, err := s.cronStore.Get(id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found", err.Error())
		return
	}
	if job.UserID != user.ID {
		writeJSONError(w, http.StatusForbidden, "forbidden", "not owner")
		return
	}
	job.Enabled = enabled
	if err := s.cronStore.Update(job); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "update failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) cronsTrigger(w http.ResponseWriter, r *http.Request, id string) {
	user := s.authUser(r)
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}
	job, err := s.cronStore.Get(id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found", err.Error())
		return
	}
	if job.UserID != user.ID {
		writeJSONError(w, http.StatusForbidden, "forbidden", "not owner")
		return
	}

	// 触发任务（提交到 task manager）
	if err := s.taskManager.Submit(task.NewTask(job.TaskType, job.Payload)); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "submit failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "triggered", "id": id})
}
