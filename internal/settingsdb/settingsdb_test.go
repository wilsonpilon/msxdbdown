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

func TestImportSQLDumpIgnoresBeginEndTransactionDirectives(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "msxdbdown.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() { _ = store.Close() }()

	sqlPath := filepath.Join(dir, "import-with-end.sql")
	content := `
BEGIN;
CREATE TABLE tx_table (
  id INTEGER PRIMARY KEY,
  title TEXT NOT NULL
);
INSERT INTO tx_table VALUES (1, 'one');
END;
BEGIN TRANSACTION;
INSERT INTO tx_table VALUES (2, 'two');
COMMIT;
`
	if err := os.WriteFile(sqlPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write sql: %v", err)
	}

	events := []string{}
	inserted, err := store.ImportSQLDump(sqlPath, false, func(message string) {
		events = append(events, message)
	})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if inserted != 2 {
		t.Fatalf("expected 2 insert statements, got %d", inserted)
	}

	hasSkippedEnd := false
	for _, ev := range events {
		if strings.Contains(ev, "skipped transaction control statement (END)") {
			hasSkippedEnd = true
			break
		}
	}
	if !hasSkippedEnd {
		t.Fatalf("expected END skip log entry, got: %v", events)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open raw db: %v", err)
	}
	defer func() { _ = db.Close() }()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM tx_table").Scan(&count); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 rows after import, got %d", count)
	}
}

func TestSearchRomInfoByNameReturnsGridRows(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "msxdbdown.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() { _ = store.Close() }()

	setupSQL := `
CREATE TABLE msxdb_company (
  CompanyID INTEGER,
  ShortName TEXT,
  fullname TEXT
);
CREATE TABLE msxdb_rominfo (
  GameID INTEGER,
  GameName TEXT,
  Year TEXT,
  CompanyID1 INTEGER,
  Platform TEXT
);
INSERT INTO msxdb_company VALUES (1, 'Konami', 'Konami Full');
INSERT INTO msxdb_company VALUES (2, '', 'ASCII Corporation');
INSERT INTO msxdb_rominfo VALUES (10, 'Metal Gear', '1987', 1, 'MSX2');
INSERT INTO msxdb_rominfo VALUES (11, 'Maze of Galious', '1987', 1, 'MSX');
INSERT INTO msxdb_rominfo VALUES (12, 'Alpha', '1984', 2, 'MSX');
`
	if _, err := store.db.Exec(setupSQL); err != nil {
		t.Fatalf("setup romdb tables: %v", err)
	}

	rows, err := store.SearchRomInfoByName("ma", 10)
	if err != nil {
		t.Fatalf("search rom info: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 result, got %d (%v)", len(rows), rows)
	}
	if rows[0].GameID != 11 {
		t.Fatalf("expected GameID 11, got %d", rows[0].GameID)
	}
	if rows[0].GameName != "Maze of Galious" || rows[0].Year != "1987" || rows[0].Platform != "MSX" || rows[0].Company != "Konami" {
		t.Fatalf("unexpected row data: %+v", rows[0])
	}

	rows, err = store.SearchRomInfoByName("", 2)
	if err != nil {
		t.Fatalf("search empty filter: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected limit=2 rows, got %d", len(rows))
	}

	if _, err := store.db.Exec("INSERT INTO msxdb_rominfo VALUES (13, '100% Maze', '1985', 2, 'MSX')"); err != nil {
		t.Fatalf("insert wildcard row: %v", err)
	}
	rows, err = store.SearchRomInfoByName("100%", 10)
	if err != nil {
		t.Fatalf("search escaped wildcard: %v", err)
	}
	if len(rows) != 1 || rows[0].GameName != "100% Maze" || rows[0].Company != "ASCII Corporation" {
		t.Fatalf("unexpected wildcard search rows: %v", rows)
	}
}

func TestGetRomInfoDetailsByGameID(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "msxdbdown.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() { _ = store.Close() }()

	setupSQL := `
CREATE TABLE msxdb_company (
  CompanyID INTEGER,
  ShortName TEXT,
  fullname TEXT
);
CREATE TABLE msxdb_rominfo (
  GameID INTEGER,
  GameName TEXT,
  Year TEXT,
  CompanyID1 INTEGER,
  CompanyID2 INTEGER,
  Platform TEXT,
  Notes TEXT
);
INSERT INTO msxdb_company VALUES (1, 'Konami', 'Konami Full');
INSERT INTO msxdb_company VALUES (2, '', 'ASCII Corporation');
INSERT INTO msxdb_rominfo VALUES (77, 'Metal Gear', '1987', 1, 2, 'MSX2', 'Stealth action');
`
	if _, err := store.db.Exec(setupSQL); err != nil {
		t.Fatalf("setup romdb tables: %v", err)
	}

	details, err := store.GetRomInfoDetailsByGameID(77)
	if err != nil {
		t.Fatalf("load details: %v", err)
	}

	checks := map[string]string{
		"GameID":       "77",
		"GameName":     "Metal Gear",
		"Year":         "1987",
		"Platform":     "MSX2",
		"CompanyID1":   "1",
		"CompanyID2":   "2",
		"CompanyName1": "Konami",
		"CompanyName2": "ASCII Corporation",
	}
	for key, want := range checks {
		if got := details[key]; got != want {
			t.Fatalf("expected %s=%q, got %q", key, want, got)
		}
	}
}

func TestGetRomVersionsByGameID(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "msxdbdown.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() { _ = store.Close() }()

	setupSQL := `
CREATE TABLE msxdb_romdetails (
  HashID INTEGER,
  GameID INTEGER,
  RomType TEXT,
  SHA1 TEXT,
  Remark TEXT,
  Meta TEXT,
  Dump TEXT,
  Active TEXT,
  Preferred TEXT
);
INSERT INTO msxdb_romdetails VALUES (1, 77, 'ASCII8', 'AAA111', '', 'v1.0', 'GoodMSX', '1', '1');
INSERT INTO msxdb_romdetails VALUES (2, 77, 'Konami', 'BBB222', 'beta', '', 'Unknown', '1', '0');
INSERT INTO msxdb_romdetails VALUES (3, 88, 'ASCII8', 'CCC333', '', 'v0.1', 'GoodMSX', '1', '1');
`
	if _, err := store.db.Exec(setupSQL); err != nil {
		t.Fatalf("setup romdetails table: %v", err)
	}

	versions, err := store.GetRomVersionsByGameID(77)
	if err != nil {
		t.Fatalf("load versions: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}

	if versions[0].SHA1 != "AAA111" || versions[0].Version != "v1.0" || versions[0].RomType != "ASCII8" {
		t.Fatalf("unexpected first version row: %+v", versions[0])
	}
	if versions[1].SHA1 != "BBB222" || versions[1].Version != "beta" || versions[1].RomType != "Konami" {
		t.Fatalf("unexpected second version row: %+v", versions[1])
	}
}
