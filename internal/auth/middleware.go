// Package auth 提供用户认证和 API Token 管理。
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// User 用户
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

// APIToken API Token
type APIToken struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	Name       string     `json:"name"`
	TokenHash  string     `json:"-"`
	Scopes     []string   `json:"scopes"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// TokenInfo 创建 token 时返回的完整 token（只显示一次）
type TokenInfo struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Token  string   `json:"token,omitempty"`
	Scopes []string `json:"scopes"`
}

// Store 认证存储
type Store struct {
	db *sql.DB
}

// NewStore 创建认证存储
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// GetUserByUsername 根据用户名获取用户
func (s *Store) GetUserByUsername(username string) (*User, error) {
	var u User
	var updatedAt sql.NullTime
	err := s.db.QueryRowContext(context.Background(),
		`SELECT id, username, password_hash, role, created_at, updated_at FROM users WHERE username=?`, username).
		Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}
	if updatedAt.Valid {
		u.UpdatedAt = updatedAt.Time
	}
	return &u, nil
}

// GetUserByID 根据 ID 获取用户
func (s *Store) GetUserByID(id string) (*User, error) {
	var u User
	var updatedAt sql.NullTime
	err := s.db.QueryRowContext(context.Background(),
		`SELECT id, username, password_hash, role, created_at, updated_at FROM users WHERE id=?`, id).
		Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}
	if updatedAt.Valid {
		u.UpdatedAt = updatedAt.Time
	}
	return &u, nil
}

// ListUsers 列出用户
func (s *Store) ListUsers(limit int) ([]*User, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(context.Background(),
		`SELECT id, username, password_hash, role, created_at, updated_at FROM users LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var u User
		var updatedAt sql.NullTime
		rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt, &updatedAt)
		if updatedAt.Valid {
			u.UpdatedAt = updatedAt.Time
		}
		users = append(users, &u)
	}
	return users, rows.Err()
}

// ValidateToken 验证 Token，返回对应的用户和 token 元信息
func (s *Store) ValidateToken(token string) (*User, *APIToken, error) {
	hash := HashToken(token)

	var t APIToken
	var expiresAt, lastUsed sql.NullTime
	var scopeStr string
	var passwordHash, role string
	var userID string

	err := s.db.QueryRowContext(context.Background(),
		`SELECT t.id, t.user_id, t.name, t.scopes, t.expires_at, t.last_used_at, t.created_at, u.id, u.password_hash, u.role
		 FROM api_tokens t JOIN users u ON t.user_id = u.id
		 WHERE t.token_hash=?`, hash).
		Scan(&t.ID, &t.UserID, &t.Name, &scopeStr, &expiresAt, &lastUsed, &t.CreatedAt, &userID, &passwordHash, &role)
	if err == sql.ErrNoRows {
		return nil, nil, fmt.Errorf("invalid token")
	}
	if err != nil {
		return nil, nil, err
	}

	if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
		return nil, nil, fmt.Errorf("token expired")
	}

	s.db.ExecContext(context.Background(),
		`UPDATE api_tokens SET last_used_at=? WHERE id=?`, time.Now(), t.ID)

	t.Scopes = parseScopes(scopeStr)
	t.ExpiresAt = nullTimeToPtr(expiresAt)
	t.LastUsedAt = nullTimeToPtr(lastUsed)

	user := &User{ID: t.UserID, Username: userID, Role: role, PasswordHash: passwordHash}
	return user, &t, nil
}

// ListTokens 列出用户的 tokens
func (s *Store) ListTokens(userID string) ([]*APIToken, error) {
	rows, err := s.db.QueryContext(context.Background(),
		`SELECT id, user_id, name, scopes, expires_at, last_used_at, created_at
		 FROM api_tokens WHERE user_id=? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*APIToken
	for rows.Next() {
		var t APIToken
		var expiresAt, lastUsed sql.NullTime
		var scopeStr string
		rows.Scan(&t.ID, &t.UserID, &t.Name, &scopeStr, &expiresAt, &lastUsed, &t.CreatedAt)
		t.Scopes = parseScopes(scopeStr)
		t.ExpiresAt = nullTimeToPtr(expiresAt)
		t.LastUsedAt = nullTimeToPtr(lastUsed)
		tokens = append(tokens, &t)
	}
	return tokens, rows.Err()
}

// DeleteToken 删除 Token
func (s *Store) DeleteToken(tokenID, userID string) error {
	res, err := s.db.ExecContext(context.Background(),
		`DELETE FROM api_tokens WHERE id=? AND user_id=?`, tokenID, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("token not found")
	}
	return nil
}

// CreateToken 创建 API Token
func (s *Store) CreateToken(userID, name string, scopes []string, expiresAt *time.Time) (*TokenInfo, error) {
	token := generateToken()
	hash := HashToken(token)
	id := generateID("tk_", 16)
	scopeStr := joinScopes(scopes)

	_, err := s.db.ExecContext(context.Background(),
		`INSERT INTO api_tokens (id, user_id, name, token_hash, scopes, expires_at, created_at)
		 VALUES (?,?,?,?,?,?,?)`,
		id, userID, name, hash, scopeStr, expiresAt, time.Now())
	if err != nil {
		return nil, err
	}

	return &TokenInfo{
		ID:     id,
		Name:   name,
		Token:  token,
		Scopes: scopes,
	}, nil
}

// CreateUser 创建用户
func (s *Store) CreateUser(user *User) error {
	if user.ID == "" {
		user.ID = GenerateUserID()
	}
	_, err := s.db.ExecContext(context.Background(),
		`INSERT INTO users (id, username, password_hash, role, created_at) VALUES (?,?,?,?,?)`,
		user.ID, user.Username, user.PasswordHash, user.Role, user.CreatedAt)
	return err
}

// generateToken 生成随机 token（32 字节 hex）
func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// generateID 生成带前缀的随机 ID
func generateID(prefix string, n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return prefix + hex.EncodeToString(b)
}

// GenerateUserID 生成用户 ID
func GenerateUserID() string {
	return generateID("u_", 16)
}

func joinScopes(scopes []string) string {
	return strings.Join(scopes, ",")
}

func parseScopes(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			result = append(result, p)
		}
	}
	return result
}

func nullTimeToPtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

// HashToken 对 token 做 SHA256 哈希
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// PasswordHash 简单哈希（生产环境请用 bcrypt）
func PasswordHash(password string) string {
	h := sha256.Sum256([]byte("icepoint_salt_" + password))
	return hex.EncodeToString(h[:])
}

// ValidatePassword 验证密码
func ValidatePassword(password, hash string) bool {
	return PasswordHash(password) == hash
}

// ==== 中间件 ====

// ContextKey context 中存储用户信息的 key
type contextKey string

const UserContextKey contextKey = "user"

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	store  *Store
	scopes []string // 需要的 scopes，为空表示只需要有效 token
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(store *Store) *AuthMiddleware {
	return &AuthMiddleware{store: store}
}

// RequireAuth 需要有效认证
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			http.Error(w, `{"error":"unauthorized","message":"Missing API token"}`, http.StatusUnauthorized)
			return
		}

		user, _, err := m.store.ValidateToken(token)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"unauthorized","message":"%s"}`, err.Error()), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireScope 需要特定 scope
func (m *AuthMiddleware) RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				http.Error(w, `{"error":"unauthorized","message":"Missing API token"}`, http.StatusUnauthorized)
				return
			}

			user, tk, err := m.store.ValidateToken(token)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"unauthorized","message":"%s"}`, err.Error()), http.StatusUnauthorized)
				return
			}

			if !hasScope(tk.Scopes, scope) {
				http.Error(w, fmt.Sprintf(`{"error":"forbidden","message":"Missing scope: %s"}`, scope), http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractToken 从请求中提取 token
func extractToken(r *http.Request) string {
	// Authorization: Bearer <token>
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	// X-API-Token
	if token := r.Header.Get("X-API-Token"); token != "" {
		return token
	}
	// Query param
	return r.URL.Query().Get("api_token")
}

// hasScope 检查是否有指定 scope
func hasScope(scopes []string, need string) bool {
	for _, s := range scopes {
		if s == need || s == "*" {
			return true
		}
	}
	return false
}

// GetUserFromContext 从 context 获取当前用户
func GetUserFromContext(ctx context.Context) *User {
	if u, ok := ctx.Value(UserContextKey).(*User); ok {
		return u
	}
	return nil
}
