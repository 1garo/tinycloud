package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/1garo/tinycloud/internal/app"
)

func TestStorePersistsApps(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	created := app.App{
		ID:        "expense-tracker",
		Name:      "expense-tracker",
		Status:    app.StatusRunning,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	first := New(path)
	if err := first.Add(created); err != nil {
		t.Fatalf("add app: %v", err)
	}

	second := New(path)
	if err := second.Load(); err != nil {
		t.Fatalf("load app: %v", err)
	}
	apps := second.List()
	if len(apps) != 1 {
		t.Fatalf("got %d apps, want 1", len(apps))
	}
	if apps[0].ID != created.ID || apps[0].Status != created.Status {
		t.Fatalf("got %#v, want %#v", apps[0], created)
	}
}

func TestStoreUpdateMissingApp(t *testing.T) {
	store := New(filepath.Join(t.TempDir(), "state.json"))
	if err := store.Update(app.App{ID: "missing"}); err == nil {
		t.Fatal("Update should fail for a missing app")
	}
}
