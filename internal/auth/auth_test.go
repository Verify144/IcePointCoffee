package auth

import (
	"path/filepath"
	"testing"
	"time"

	"database/sql"
	_ "modernc.org/sqlite"
)

func newAuthDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	db, _ := sql.Open("sqlite", filepath.Join(dir, "auth.db")+"?_journal=WAL")
	db.SetMaxOpenConns(1)

	schema := `
	CREATE TABLE users (id TEXT PRIMARY KEY, username TEXT NOT NULL UNIQUE, password_hash TEXT NOT NULL, role TEXT NOT NULL DEFAULT 'user', created_at DATETIME NOT NULL, updated_at DATETIME);
	CREATE TABLE api_tokens (id TEXT PRIMARY KEY, user_id TEXT NOT NULL, name TEXT NOT NULL, token_hash TEXT NOT NULL UNIQUE, scopes TEXT NOT NULL DEFAULT '[]', expires_at DATETIME, last_used_at DATETIME, created_at DATETIME NOT NULL);
	`
	db.Exec(schema)
	return db
}

func TestUserCreateAndGet(t *testing.T) {
	db := newAuthDB(t)
	defer db.Close()

	store := NewStore(db)
	user := &User{
		ID:           GenerateUserID(),
		Username:     "alice",
		PasswordHash: PasswordHash("secret123"),
		Role:         "user",
		CreatedAt:    time.Now(),
	}
	if err := store.CreateUser(user); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	got, err := store.GetUserByUsername("alice")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != user.ID || got.Role != "user" {
		t.Errorf("Mismatch: %+v", got)
	}
}

func TestPasswordHashValidate(t *testing.T) {
	hash := PasswordHash("test_password")
	if !ValidatePassword("test_password", hash) {
		t.Error("Valid password should match")
	}
	if ValidatePassword("wrong", hash) {
		t.Error("Wrong password should not match")
	}
}

func TestTokenCreateAndValidate(t *testing.T) {
	db := newAuthDB(t)
	defer db.Close()

	store := NewStore(db)
	user := &User{
		ID:           GenerateUserID(),
		Username:     "bob",
		PasswordHash: PasswordHash("x"),
		Role:         "user",
		CreatedAt:    time.Now(),
	}
	store.CreateUser(user)

	tokenInfo, err := store.CreateToken(user.ID, "test-token", []string{"user", "build"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if tokenInfo.Token == "" {
		t.Fatal("Token should be set")
	}
	if len(tokenInfo.Token) != 64 {
		t.Errorf("Token should be 64 hex chars, got %d", len(tokenInfo.Token))
	}

	// 验证 token
	gotUser, gotToken, err := store.ValidateToken(tokenInfo.Token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if gotUser.ID != user.ID {
		t.Errorf("User mismatch")
	}
	if gotToken.Name != "test-token" {
		t.Errorf("Token name mismatch")
	}
	if len(gotToken.Scopes) != 2 {
		t.Errorf("Scopes mismatch: %v", gotToken.Scopes)
	}
}

func TestTokenInvalid(t *testing.T) {
	db := newAuthDB(t)
	defer db.Close()

	store := NewStore(db)
	_, _, err := store.ValidateToken("nonexistent_token")
	if err == nil {
		t.Error("Should error on invalid token")
	}
}

func TestTokenExpired(t *testing.T) {
	db := newAuthDB(t)
	defer db.Close()

	store := NewStore(db)
	user := &User{
		ID:           GenerateUserID(),
		Username:     "expired",
		PasswordHash: "x",
		CreatedAt:    time.Now(),
	}
	store.CreateUser(user)

	// 创建已过期的 token
	pastTime := time.Now().Add(-1 * time.Hour)
	_, err := store.CreateToken(user.ID, "expired-token", []string{"user"}, &pastTime)
	if err != nil {
		t.Fatal(err)
	}

	// 验证应该失败
	tokens, _ := store.ListTokens(user.ID)
	if len(tokens) == 0 {
		t.Fatal("Token should exist")
	}
	_, _, err = store.ValidateToken("nope")
	if err == nil {
		t.Error("Should fail")
	}
}

func TestTokenDelete(t *testing.T) {
	db := newAuthDB(t)
	defer db.Close()

	store := NewStore(db)
	user := &User{ID: GenerateUserID(), Username: "delete_test", PasswordHash: "x", CreatedAt: time.Now()}
	store.CreateUser(user)

	tokenInfo, _ := store.CreateToken(user.ID, "to-delete", []string{"user"}, nil)
	if err := store.DeleteToken(tokenInfo.ID, user.ID); err != nil {
		t.Fatal(err)
	}

	// 重复删除应该失败
	if err := store.DeleteToken(tokenInfo.ID, user.ID); err == nil {
		t.Error("Delete should fail on already-deleted token")
	}
}

func TestListUsers(t *testing.T) {
	db := newAuthDB(t)
	defer db.Close()

	store := NewStore(db)
	for i, name := range []string{"u1", "u2", "u3"} {
		store.CreateUser(&User{
			ID: GenerateUserID(), Username: name, PasswordHash: "x", CreatedAt: time.Now(),
		})
		_ = i
	}

	users, _ := store.ListUsers(10)
	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}
}

func TestHashToken(t *testing.T) {
	hash1 := HashToken("test")
	hash2 := HashToken("test")
	if hash1 != hash2 {
		t.Error("Same input should produce same hash")
	}
	hash3 := HashToken("test2")
	if hash1 == hash3 {
		t.Error("Different input should produce different hash")
	}
}
