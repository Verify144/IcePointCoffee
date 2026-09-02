package auth

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Handlers 认证相关的 HTTP handler
type Handlers struct {
	store *Store
}

// NewHandlers 创建认证 handlers
func NewHandlers(store *Store) *Handlers {
	return &Handlers{store: store}
}

// HandleRegister 用户注册
// POST /api/v1/auth/register
func (h *Handlers) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"` // 可选，默认为 "user"
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json", err.Error())
		return
	}

	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "missing fields", "username and password required")
		return
	}
	if len(req.Password) < 6 {
		writeError(w, http.StatusBadRequest, "weak password", "password must be at least 6 chars")
		return
	}

	// 检查用户名是否已存在
	if existing, _ := h.store.GetUserByUsername(req.Username); existing != nil {
		writeError(w, http.StatusConflict, "user exists", "username already taken")
		return
	}

	role := req.Role
	if role == "" {
		role = "user"
	}

	user := &User{
		ID:           GenerateUserID(),
		Username:     req.Username,
		PasswordHash: PasswordHash(req.Password),
		Role:         role,
		CreatedAt:    time.Now(),
	}

	if err := h.store.CreateUser(user); err != nil {
		writeError(w, http.StatusInternalServerError, "create user failed", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"role":       user.Role,
		"created_at": user.CreatedAt,
	})
}

// HandleLogin 用户登录（返回 token）
// POST /api/v1/auth/login
func (h *Handlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json", err.Error())
		return
	}

	user, err := h.store.GetUserByUsername(req.Username)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials", "user not found")
		return
	}

	if !ValidatePassword(req.Password, user.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "invalid credentials", "wrong password")
		return
	}

	// 自动创建 token
	tokenInfo, err := h.store.CreateToken(user.ID, "default", []string{"user"}, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create token failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
		"token": tokenInfo.Token,
	})
}

// HandleCreateToken 创建新的 API token（需登录）
// POST /api/v1/auth/tokens
func (h *Handlers) HandleCreateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "auth required")
		return
	}

	var req struct {
		Name      string   `json:"name"`
		Scopes    []string `json:"scopes"`
		ExpiresIn int      `json:"expires_in"` // 秒，0=不过期
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json", err.Error())
		return
	}

	if req.Name == "" {
		req.Name = "default"
	}
	if len(req.Scopes) == 0 {
		req.Scopes = []string{"user"}
	}

	var expiresAt *time.Time
	if req.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(req.ExpiresIn) * time.Second)
		expiresAt = &t
	}

	tokenInfo, err := h.store.CreateToken(user.ID, req.Name, req.Scopes, expiresAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create token failed", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":     tokenInfo.ID,
		"name":   tokenInfo.Name,
		"token":  tokenInfo.Token, // 仅此次显示
		"scopes": tokenInfo.Scopes,
	})
}

// HandleListTokens 列出当前用户的 tokens
// GET /api/v1/auth/tokens
func (h *Handlers) HandleListTokens(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}

	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "auth required")
		return
	}

	tokens, err := h.store.ListTokens(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tokens": tokens,
	})
}

// HandleDeleteToken 删除 token
// DELETE /api/v1/auth/tokens/{id}
func (h *Handlers) HandleDeleteToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "DELETE only", http.StatusMethodNotAllowed)
		return
	}

	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "auth required")
		return
	}

	// 提取 token ID（已被 StripPrefix 去掉 /auth/tokens/）
	// path 现在是 "/{id}" 或 "/{id}/"
	id := strings.Trim(r.URL.Path, "/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id", "token id required")
		return
	}

	if err := h.store.DeleteToken(id, user.ID); err != nil {
		writeError(w, http.StatusNotFound, "delete failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted", "id": id})
}

// HandleMe 获取当前登录用户
// GET /api/v1/auth/me
func (h *Handlers) HandleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}

	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "auth required")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"role":       user.Role,
		"created_at": user.CreatedAt,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, err, msg string) {
	writeJSON(w, status, map[string]interface{}{
		"error":   err,
		"message": msg,
	})
}
