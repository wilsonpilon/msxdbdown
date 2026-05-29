package settingsdb

import (
	"path/filepath"
	"testing"
)

func TestStoreSetAndGet(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "settings.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() { _ = store.Close() }()

	if err := store.Set("ui.language", "es"); err != nil {
		t.Fatalf("set: %v", err)
	}

	got, err := store.Get("ui.language")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != "es" {
		t.Fatalf("expected es, got %q", got)
	}
}

func TestStoreGetMissingReturnsEmpty(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "settings.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() { _ = store.Close() }()

	got, err := store.Get("not.found")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty value, got %q", got)
	}
}

