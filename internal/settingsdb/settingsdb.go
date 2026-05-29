package settingsdb

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const driverName = "sqlite"

// Store persists app settings in a simple key/value SQLite table.
type Store struct {
	db *sql.DB
}

func OpenDefault() (*Store, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil || cfgDir == "" {
		cfgDir = "."
	}

	dir := filepath.Join(cfgDir, "msxdbdown")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}

	return Open(filepath.Join(dir, "settings.db"))
}

func Open(dbPath string) (*Store, error) {
	db, err := sql.Open(driverName, dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	s := &Store{db: db}
	if err := s.initSchema(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return s, nil
}

func (s *Store) initSchema() error {
	if s == nil || s.db == nil {
		return errors.New("nil store")
	}

	const schema = `
CREATE TABLE IF NOT EXISTS app_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
`
	_, err := s.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("init schema: %w", err)
	}
	return nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Get(key string) (string, error) {
	if s == nil || s.db == nil {
		return "", errors.New("nil store")
	}

	var value string
	err := s.db.QueryRow("SELECT value FROM app_settings WHERE key = ?", key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get %s: %w", key, err)
	}
	return value, nil
}

func (s *Store) Set(key, value string) error {
	if s == nil || s.db == nil {
		return errors.New("nil store")
	}

	_, err := s.db.Exec(`
INSERT INTO app_settings(key, value)
VALUES(?, ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value
`, key, value)
	if err != nil {
		return fmt.Errorf("set %s: %w", key, err)
	}
	return nil
}
