package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Verify144/IcePointCoffee/internal/auth"
	"github.com/Verify144/IcePointCoffee/internal/template"
)

// handleTemplateRoutes 注册模板相关路由
func (s *Server) handleTemplateRoutes(api *http.ServeMux) {
	// 公开：列出公开模板
	api.HandleFunc("/templates", s.handleTemplatesList)
	// 公开：获取模板详情
	api.Handle("/templates/", http.StripPrefix("/templates", s.templateHandlers()))
}

// templateHandlers 返回模板子路由 mux（去掉 /templates 前缀）
func (s *Server) templateHandlers() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleTemplatesItem)
	return mux
}

// handleTemplatesList 列出模板（GET）
// POST 创建模板（需 auth）
func (s *Server) handleTemplatesList(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.templatesList(w, r)
	case http.MethodPost:
		s.templatesCreate(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) templatesList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := template.ListFilter{
		Category: q.Get("category"),
		Search:   q.Get("q"),
		SortBy:   q.Get("sort"),
	}
	if userID := q.Get("user_id"); userID != "" {
		filter.UserID = userID
	}
	if public := q.Get("public"); public == "true" {
		p := true
		filter.Public = &p
	} else if public == "false" {
		p := false
		filter.Public = &p
	}
	if limit := q.Get("limit"); limit != "" {
		var n int
		// simple parse
		for _, c := range limit {
			if c >= '0' && c <= '9' {
				n = n*10 + int(c-'0')
			}
		}
		filter.Limit = n
	}

	list, err := s.templateStore.List(filter)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "list failed", err.Error())
		return
	}

	cats, _ := s.templateStore.Categories()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"templates":  list,
		"categories": cats,
		"count":      len(list),
	})
}

func (s *Server) templatesCreate(w http.ResponseWriter, r *http.Request) {
	if s.authMiddleware == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "auth not enabled", "")
		return
	}
	// 手动校验 token（POST /templates 需要登录）
	token := extractBearerToken(r)
	if token == "" {
		writeJSONError(w, http.StatusUnauthorized, "missing token", "")
		return
	}
	user, _, err := s.authStore.ValidateToken(token)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "invalid token", err.Error())
		return
	}

	var req struct {
		Name         string                 `json:"name"`
		Description  string                 `json:"description"`
		Category     string                 `json:"category"`
		ParamsSchema map[string]interface{} `json:"params_schema"`
		Blocks       map[string]interface{} `json:"blocks"`
		IsPublic     bool                   `json:"is_public"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json", err.Error())
		return
	}
	if req.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "missing name", "")
		return
	}
	if req.Category == "" {
		req.Category = "custom"
	}

	tpl := &template.Template{
		UserID:       user.ID,
		Name:         req.Name,
		Description:  req.Description,
		Category:     req.Category,
		ParamsSchema: req.ParamsSchema,
		Blocks:       req.Blocks,
		IsPublic:     req.IsPublic,
		CreatedAt:    time.Now(),
	}

	if err := s.templateStore.Create(tpl); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "create failed", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, tpl)
}

func (s *Server) handleTemplatesItem(w http.ResponseWriter, r *http.Request) {
	// path is already stripped of /templates prefix, e.g. "/tpl_house_001/like" or "/tpl_house_001"
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
		if action == "use" {
			s.templatesUse(w, r, id)
			return
		}
		if action == "like" {
			s.templatesLike(w, r, id)
			return
		}
		s.templatesGet(w, r, id)
	case http.MethodPost:
		if action == "like" {
			s.templatesLike(w, r, id)
			return
		}
		http.Error(w, "Unknown action: "+action, http.StatusNotFound)
	case http.MethodDelete:
		s.templatesDelete(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) templatesGet(w http.ResponseWriter, r *http.Request, id string) {
	tpl, err := s.templateStore.Get(id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tpl)
}

func (s *Server) templatesDelete(w http.ResponseWriter, r *http.Request, id string) {
	if s.authMiddleware == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "auth not enabled", "")
		return
	}
	token := extractBearerToken(r)
	if token == "" {
		writeJSONError(w, http.StatusUnauthorized, "missing token", "")
		return
	}
	user, _, err := s.authStore.ValidateToken(token)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "invalid token", err.Error())
		return
	}

	if err := s.templateStore.Delete(id, user.ID); err != nil {
		writeJSONError(w, http.StatusNotFound, "delete failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted", "id": id})
}

func (s *Server) templatesUse(w http.ResponseWriter, r *http.Request, id string) {
	tpl, err := s.templateStore.Get(id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found", err.Error())
		return
	}
	s.templateStore.IncrementUses(id)

	// 构造 build 任务
	// 实际执行会交给 builder，这里只是暴露接口
	_ = tpl
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "queued",
		"template": tpl,
		"message":  "Template queued for building",
	})
}

func (s *Server) templatesLike(w http.ResponseWriter, r *http.Request, id string) {
	s.templateStore.IncrementLikes(id)
	tpl, _ := s.templateStore.Get(id)
	writeJSON(w, http.StatusOK, map[string]interface{}{"likes": tpl.Likes})
}

// ==== helpers ====

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	if token := r.Header.Get("X-API-Token"); token != "" {
		return token
	}
	return ""
}

// authUser 尝试从 Authorization 头中获取用户
func (s *Server) authUser(r *http.Request) *auth.User {
	token := extractBearerToken(r)
	if token == "" || s.authStore == nil {
		return nil
	}
	user, _, _ := s.authStore.ValidateToken(token)
	return user
}
