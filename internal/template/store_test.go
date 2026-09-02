package template

import (
	"path/filepath"
	"testing"
	"time"

	"database/sql"
	_ "modernc.org/sqlite"
)

func newTemplateDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	db, _ := sql.Open("sqlite", filepath.Join(dir, "tpl.db")+"?_journal=WAL")
	db.SetMaxOpenConns(1)

	schema := `
	CREATE TABLE build_templates (id TEXT PRIMARY KEY, user_id TEXT NOT NULL, name TEXT NOT NULL, description TEXT, category TEXT NOT NULL DEFAULT 'custom', params_schema TEXT NOT NULL DEFAULT '{}', blocks TEXT NOT NULL DEFAULT '{}', is_public INTEGER NOT NULL DEFAULT 0, likes INTEGER NOT NULL DEFAULT 0, uses INTEGER NOT NULL DEFAULT 0, created_at DATETIME NOT NULL, updated_at DATETIME);
	CREATE INDEX idx_templates_user ON build_templates(user_id);
	CREATE INDEX idx_templates_category ON build_templates(category);
	`
	db.Exec(schema)
	return db
}

func TestTemplateCreateAndGet(t *testing.T) {
	db := newTemplateDB(t)
	defer db.Close()

	store := NewStore(db)
	tpl := &Template{
		ID:          "test_001",
		UserID:      "u1",
		Name:        "Test House",
		Description: "A test house",
		Category:    "residential",
		ParamsSchema: map[string]interface{}{
			"width": map[string]interface{}{"type": "int", "default": 10},
		},
		Blocks:     map[string]interface{}{"type": "fill"},
		IsPublic:   false,
		CreatedAt:  time.Now(),
	}

	if err := store.Create(tpl); err != nil {
		t.Fatal(err)
	}

	got, err := store.Get("test_001")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Test House" || got.Category != "residential" {
		t.Errorf("Mismatch: %+v", got)
	}
	if got.ParamsSchema["width"] == nil {
		t.Error("ParamsSchema should be preserved")
	}
}

func TestTemplateList(t *testing.T) {
	db := newTemplateDB(t)
	defer db.Close()

	store := NewStore(db)
	for i := 0; i < 5; i++ {
		store.Create(&Template{
			ID:          "tpl_" + string(rune('a'+i)),
			UserID:      "u1",
			Name:        "tpl",
			Category:    "residential",
			IsPublic:    true,
			CreatedAt:   time.Now(),
		})
	}

	list, err := store.List(ListFilter{Limit: 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 {
		t.Errorf("Expected 3, got %d", len(list))
	}

	categories, err := store.Categories()
	if err != nil {
		t.Fatal(err)
	}
	if len(categories) != 1 {
		t.Errorf("Expected 1 category, got %d", len(categories))
	}
}

func TestTemplateSeed(t *testing.T) {
	db := newTemplateDB(t)
	defer db.Close()

	store := NewStore(db)
	if err := store.SeedPublicTemplates(); err != nil {
		t.Fatal(err)
	}
	if err := store.SeedPublicTemplates(); err != nil {
		t.Fatal(err)
	}

	list, _ := store.List(ListFilter{Public: boolPtr(true)})
	if len(list) != 5 {
		t.Errorf("Expected 5 public templates, got %d", len(list))
	}
}

func TestTemplateLikes(t *testing.T) {
	db := newTemplateDB(t)
	defer db.Close()

	store := NewStore(db)
	store.Create(&Template{ID: "like_t", UserID: "u1", Category: "x", CreatedAt: time.Now()})

	store.IncrementLikes("like_t")
	store.IncrementLikes("like_t")
	tpl, _ := store.Get("like_t")
	if tpl.Likes != 2 {
		t.Errorf("Expected 2 likes, got %d", tpl.Likes)
	}
}

func TestTemplateDelete(t *testing.T) {
	db := newTemplateDB(t)
	defer db.Close()

	store := NewStore(db)
	store.Create(&Template{ID: "del_t", UserID: "u1", Category: "x", CreatedAt: time.Now()})

	if err := store.Delete("del_t", "u1"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get("del_t"); err == nil {
		t.Error("Should be deleted")
	}
}

func boolPtr(b bool) *bool { return &b }
