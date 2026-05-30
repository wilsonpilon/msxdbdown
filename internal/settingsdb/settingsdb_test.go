package settingsdb

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
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

func TestOpenSeedsDefaultSettings(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "msxdbdown.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() { _ = store.Close() }()

	checks := map[string]string{
		"ui.language":         "en",
		"ui.theme":            "System",
		"ui.fontName":         "System",
		"ui.fontSize":         "14",
		"ui.density":          "Normal",
		"db.catalog.location": "local",
	}

	for key, want := range checks {
		got, err := store.Get(key)
		if err != nil {
			t.Fatalf("get %s: %v", key, err)
		}
		if got != want {
			t.Fatalf("expected %s=%q, got %q", key, want, got)
		}
	}
}

func TestImportSQLDumpCreatesAndInserts(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "msxdbdown.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() { _ = store.Close() }()

	sqlPath := filepath.Join(dir, "import.sql")
	content := `
/* comment */
CREATE TABLE my_table (
  id INTEGER,
  title VARCHAR(255)
);
INSERT INTO my_table VALUES (1, 'one');
INSERT INTO my_table VALUES (2, 'two');
`
	if err := os.WriteFile(sqlPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write sql: %v", err)
	}

	inserted, err := store.ImportSQLDump(sqlPath, false, nil)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if inserted != 2 {
		t.Fatalf("expected 2 insert statements, got %d", inserted)
	}

	// Re-import in refresh mode should avoid duplicates.
	if _, err := store.ImportSQLDump(sqlPath, true, nil); err != nil {
		t.Fatalf("re-import: %v", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open raw db: %v", err)
	}
	defer func() { _ = db.Close() }()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM my_table").Scan(&count); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 rows after refresh re-import, got %d", count)
	}
}

func TestImportSQLDumpRefreshIsAtomicPerTable(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "msxdbdown.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() { _ = store.Close() }()

	seedSQL := filepath.Join(dir, "seed.sql")
	seedContent := `
CREATE TABLE games (
  id INTEGER PRIMARY KEY,
  title TEXT NOT NULL
);
INSERT INTO games VALUES (1, 'old');
`
	if err := os.WriteFile(seedSQL, []byte(seedContent), 0o644); err != nil {
		t.Fatalf("write seed sql: %v", err)
	}
	if _, err := store.ImportSQLDump(seedSQL, false, nil); err != nil {
		t.Fatalf("seed import: %v", err)
	}

	refreshSQL := filepath.Join(dir, "refresh.sql")
	refreshContent := `
CREATE TABLE games (
  id INTEGER PRIMARY KEY,
  title TEXT NOT NULL
);
INSERT INTO games VALUES (2, 'new');
`
	if err := os.WriteFile(refreshSQL, []byte(refreshContent), 0o644); err != nil {
		t.Fatalf("write refresh sql: %v", err)
	}
	events := []string{}
	if _, err := store.ImportSQLDump(refreshSQL, true, func(message string) {
		events = append(events, message)
	}); err != nil {
		t.Fatalf("refresh import: %v", err)
	}

	hasBackupCreated := false
	hasRecreated := false
	hasBackupRemoved := false
	hasSummary := false
	for _, ev := range events {
		if strings.Contains(ev, "backup created:") {
			hasBackupCreated = true
		}
		if strings.Contains(ev, "table recreated") {
			hasRecreated = true
		}
		if strings.Contains(ev, "backup removed:") {
			hasBackupRemoved = true
		}
		if strings.Contains(ev, "summary:") && strings.Contains(ev, "tables recreated") && strings.Contains(ev, "backups removed") {
			hasSummary = true
		}
	}
	if !hasBackupCreated || !hasRecreated || !hasBackupRemoved || !hasSummary {
		t.Fatalf("expected detailed refresh logs, got: %v", events)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open raw db: %v", err)
	}
	defer func() { _ = db.Close() }()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM games").Scan(&count); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row after refresh import, got %d", count)
	}

	var title string
	if err := db.QueryRow("SELECT title FROM games WHERE id = 2").Scan(&title); err != nil {
		t.Fatalf("read refreshed row: %v", err)
	}
	if title != "new" {
		t.Fatalf("expected refreshed title 'new', got %q", title)
	}

	var backupCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name LIKE '__msxdb_backup_%'").Scan(&backupCount); err != nil {
		t.Fatalf("count backup tables: %v", err)
	}
	if backupCount != 0 {
		t.Fatalf("expected no backup tables after successful import, got %d", backupCount)
	}
}
