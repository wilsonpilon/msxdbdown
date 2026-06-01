package settingsdb

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"msxdbdown/internal/uiprefs"

	_ "modernc.org/sqlite"
)

const driverName = "sqlite"

const (
	LocationLocal   = "local"
	LocationAppData = "appdata"
)

// Store persists app settings in a simple key/value SQLite table.
type Store struct {
	db *sql.DB
}

type RomInfoSearchRow struct {
	GameID   int64
	GameName string
	Year     string
	Platform string
	Company  string
}

type RomVersionRow struct {
	RomType   string
	Version   string
	SHA1      string
	Source    string // Dump quality string (GoodMSX, Unknown, etc.)
	FileSize  string
	Active    string
	Preferred string
	RomFound  string
}

type ImportLogFunc func(message string)

type importStats struct {
	backupCreated   int
	tablesRecreated int
	backupsRemoved  int
}

func NormalizeLocation(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case LocationAppData:
		return LocationAppData
	default:
		return LocationLocal
	}
}

func ResolvePath(location string) (string, error) {
	if NormalizeLocation(location) == LocationAppData {
		cfgDir, err := os.UserConfigDir()
		if err != nil || cfgDir == "" {
			return "", fmt.Errorf("resolve config dir: %w", err)
		}
		return filepath.Join(cfgDir, "msxdbdown", "msxdbdown.db"), nil
	}
	return filepath.Join(".", "data", "msxdbdown.db"), nil
}

func DetectCurrentPath() (location string, dbPath string, err error) {
	localPath, err := ResolvePath(LocationLocal)
	if err != nil {
		return "", "", err
	}
	localInfo, localErr := os.Stat(localPath)
	if localErr != nil && !errors.Is(localErr, os.ErrNotExist) {
		return "", "", fmt.Errorf("check local db: %w", localErr)
	}

	appDataPath, err := ResolvePath(LocationAppData)
	if err != nil {
		return "", "", err
	}
	appInfo, appErr := os.Stat(appDataPath)
	if appErr != nil && !errors.Is(appErr, os.ErrNotExist) {
		return "", "", fmt.Errorf("check appdata db: %w", appErr)
	}

	if localErr == nil && appErr == nil {
		if appInfo.ModTime().After(localInfo.ModTime()) {
			return LocationAppData, appDataPath, nil
		}
		return LocationLocal, localPath, nil
	}
	if localErr == nil {
		return LocationLocal, localPath, nil
	}
	if appErr == nil {
		return LocationAppData, appDataPath, nil
	}

	legacyPath := filepath.Join(filepath.Dir(appDataPath), "settings.db")
	if _, statErr := os.Stat(legacyPath); statErr == nil {
		return LocationAppData, legacyPath, nil
	}

	return LocationLocal, localPath, nil
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
	if err := s.initDefaults(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return s, nil
}

func (s *Store) initDefaults() error {
	if s == nil || s.db == nil {
		return errors.New("nil store")
	}

	const defaults = `
INSERT OR IGNORE INTO app_settings(key, value) VALUES
('ui.language', 'en'),
('ui.theme', 'System'),
('ui.fontName', 'System'),
('ui.fontSize', '14'),
('ui.density', 'Normal'),
('db.msxromdb.url', 'https://romdb.vampier.net/Archive/sql-msxromdb.zip'),
('db.filehunter.url', 'https://download.file-hunter.com/allfiles.txt'),
('db.filehunter.sha.url', 'https://download.file-hunter.com/sha1sums.txt'),
('db.catalog.location', 'local');
`

	_, err := s.db.Exec(defaults)
	if err != nil {
		return fmt.Errorf("init defaults: %w", err)
	}
	return nil
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

func (s *Store) SearchRomInfoByName(name string, limit int) ([]RomInfoSearchRow, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("nil store")
	}
	if limit <= 0 {
		limit = 200
	}

	queryName := "%" + escapeSQLLike(strings.TrimSpace(name)) + "%"

	const query = `
SELECT
				  COALESCE(r.GameID, 0),
  COALESCE(r.GameName, ''),
  COALESCE(r.Year, ''),
  COALESCE(r.Platform, ''),
  COALESCE(NULLIF(c.ShortName, ''), NULLIF(c.fullname, ''), '') AS company
FROM msxdb_rominfo r
LEFT JOIN msxdb_company c ON c.CompanyID = r.CompanyID1
WHERE COALESCE(r.GameName, '') LIKE ? ESCAPE '\'
ORDER BY r.GameName COLLATE NOCASE, r.Year
LIMIT ?;
`

	rows, err := s.db.Query(query, queryName, limit)
	if err != nil {
		return nil, fmt.Errorf("search msxdb_rominfo: %w", err)
	}
	defer func() { _ = rows.Close() }()

	result := make([]RomInfoSearchRow, 0, 32)
	for rows.Next() {
		var row RomInfoSearchRow
		if err := rows.Scan(&row.GameID, &row.GameName, &row.Year, &row.Platform, &row.Company); err != nil {
			return nil, fmt.Errorf("scan msxdb_rominfo row: %w", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate msxdb_rominfo rows: %w", err)
	}

	return result, nil
}

func (s *Store) GetRomInfoDetailsByGameID(gameID int64) (map[string]string, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("nil store")
	}

	columns, err := romInfoColumnNames(s.db)
	if err != nil {
		return nil, err
	}
	if len(columns) == 0 {
		return nil, errors.New("msxdb_rominfo has no columns")
	}

	query := fmt.Sprintf("SELECT * FROM msxdb_rominfo WHERE GameID = ? LIMIT 1")
	values := make([]any, len(columns))
	scanArgs := make([]any, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	if err := s.db.QueryRow(query, gameID).Scan(scanArgs...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("game not found: %d", gameID)
		}
		return nil, fmt.Errorf("load game details %d: %w", gameID, err)
	}

	details := make(map[string]string, len(columns)+2)
	for i, col := range columns {
		details[col] = sqlValueToString(values[i])
	}

	if companyID := strings.TrimSpace(details["CompanyID1"]); companyID != "" {
		if name, err := lookupCompanyName(s.db, companyID); err == nil && name != "" {
			details["CompanyName1"] = name
		}
	}
	if companyID := strings.TrimSpace(details["CompanyID2"]); companyID != "" {
		if name, err := lookupCompanyName(s.db, companyID); err == nil && name != "" {
			details["CompanyName2"] = name
		}
	}

	return details, nil
}

func (s *Store) GetRomVersionsByGameID(gameID int64) ([]RomVersionRow, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("nil store")
	}

	const query = `
SELECT
  COALESCE(RomType, ''),
  COALESCE(NULLIF(Meta, ''), NULLIF(Remark, ''), ''),
  COALESCE(SHA1, ''),
  COALESCE(Dump, ''),
  COALESCE(CAST(FileSize AS TEXT), ''),
  COALESCE(CAST(Active AS TEXT), '0'),
  COALESCE(CAST(Preferred AS TEXT), '0'),
  COALESCE(CAST(RomFound AS TEXT), '0')
FROM msxdb_romdetails
WHERE GameID = ?
ORDER BY CAST(COALESCE(Preferred, '0') AS INTEGER) DESC,
         CAST(COALESCE(Active, '0') AS INTEGER) DESC,
         HashID ASC;
`

	rows, err := s.db.Query(query, gameID)
	if err != nil {
		return nil, fmt.Errorf("load rom versions %d: %w", gameID, err)
	}
	defer func() { _ = rows.Close() }()

	result := make([]RomVersionRow, 0, 8)
	for rows.Next() {
		var row RomVersionRow
		if err := rows.Scan(&row.RomType, &row.Version, &row.SHA1, &row.Source, &row.FileSize, &row.Active, &row.Preferred, &row.RomFound); err != nil {
			return nil, fmt.Errorf("scan rom version row: %w", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rom version rows: %w", err)
	}

	return result, nil
}

func romInfoColumnNames(db *sql.DB) ([]string, error) {
	rows, err := db.Query("PRAGMA table_info(msxdb_rominfo)")
	if err != nil {
		return nil, fmt.Errorf("read msxdb_rominfo schema: %w", err)
	}
	defer func() { _ = rows.Close() }()

	columns := make([]string, 0, 24)
	for rows.Next() {
		var cid int
		var name string
		var colType string
		var notNull int
		var defaultValue any
		var pk int
		if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultValue, &pk); err != nil {
			return nil, fmt.Errorf("scan msxdb_rominfo schema: %w", err)
		}
		columns = append(columns, name)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate msxdb_rominfo schema: %w", err)
	}

	return columns, nil
}

func lookupCompanyName(db *sql.DB, companyID string) (string, error) {
	var name string
	err := db.QueryRow(`
SELECT COALESCE(NULLIF(ShortName, ''), NULLIF(fullname, ''), '')
FROM msxdb_company
WHERE CompanyID = ?
LIMIT 1
`, companyID).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(name), nil
}

func sqlValueToString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func ResetAtPath(dbPath string, location string) error {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return fmt.Errorf("create database directory: %w", err)
	}

	for _, candidate := range []string{dbPath, dbPath + "-wal", dbPath + "-shm"} {
		if err := os.Remove(candidate); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove %s: %w", candidate, err)
		}
	}

	store, err := Open(dbPath)
	if err != nil {
		return err
	}
	defer func() { _ = store.Close() }()

	if err := store.Set(uiprefs.PrefCatalogDBLocation, NormalizeLocation(location)); err != nil {
		return fmt.Errorf("set db location: %w", err)
	}
	return nil
}

func MoveDatabaseFiles(sourcePath, targetPath string) error {
	if sourcePath == targetPath {
		return nil
	}

	if _, err := os.Stat(sourcePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("source database not found: %s", sourcePath)
		}
		return fmt.Errorf("check source database: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}

	for _, suffix := range []string{"", "-wal", "-shm"} {
		from := sourcePath + suffix
		to := targetPath + suffix
		if err := moveFile(from, to); err != nil {
			return err
		}
	}

	return nil
}

func moveFile(from, to string) error {
	content, err := os.ReadFile(from)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read %s: %w", from, err)
	}

	if err := os.WriteFile(to, content, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", to, err)
	}

	if err := os.Remove(from); err != nil {
		return fmt.Errorf("remove %s: %w", from, err)
	}

	return nil
}

// ImportSQLDump executes SQL statements from a dump file into the current store.
// It ignores BEGIN/COMMIT directives from the dump and wraps execution in a
// single SQLite transaction. CREATE TABLE statements are made idempotent.
// When refresh is true, tables found in INSERT INTO statements are atomically
// refreshed by backing up the current table, recreating it, importing new rows,
// and removing the backup only after a successful transaction.
func (s *Store) ImportSQLDump(sqlPath string, refresh bool, logf ImportLogFunc) (int, error) {
	if s == nil || s.db == nil {
		return 0, errors.New("nil store")
	}

	f, err := os.Open(sqlPath)
	if err != nil {
		return 0, fmt.Errorf("open sql dump %s: %w", sqlPath, err)
	}
	defer func() { _ = f.Close() }()

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin import transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)

	var b strings.Builder
	inBlockComment := false
	insertedRows := 0
	createStmts := map[string]string{}
	backupTables := map[string]string{}
	stats := &importStats{}

	flushStatement := func() error {
		stmt := strings.TrimSpace(b.String())
		b.Reset()
		if stmt == "" {
			return nil
		}

		normalized := strings.ToUpper(strings.TrimSpace(stmt))
		if keyword, ok := transactionControlKeyword(stmt); ok {
			if logf != nil {
				logf(fmt.Sprintf("skipped transaction control statement (%s)", keyword))
			}
			return nil
		}

		if strings.HasPrefix(normalized, "CREATE TABLE ") && !strings.Contains(normalized, "IF NOT EXISTS") {
			stmt = strings.Replace(stmt, "CREATE TABLE ", "CREATE TABLE IF NOT EXISTS ", 1)
			normalized = strings.ToUpper(strings.TrimSpace(stmt))
		}

		if strings.HasPrefix(normalized, "CREATE TABLE ") {
			table := parseCreateTableTarget(stmt)
			if table != "" {
				createStmts[table] = stmt
			}
		}

		if strings.HasPrefix(normalized, "INSERT INTO ") {
			table := parseInsertTargetTable(stmt)
			if refresh && table != "" {
				if err := ensureRefreshTable(tx, table, createStmts[table], backupTables, stats, logf); err != nil {
					return err
				}
			}
		}

		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("exec sql statement failed: %w", err)
		}
		if strings.HasPrefix(normalized, "INSERT INTO ") {
			insertedRows++
		}
		return nil
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if inBlockComment {
			if strings.Contains(trimmed, "*/") {
				inBlockComment = false
			}
			continue
		}
		if strings.HasPrefix(trimmed, "/*") {
			if !strings.Contains(trimmed, "*/") {
				inBlockComment = true
			}
			continue
		}
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}

		b.WriteString(line)
		b.WriteByte('\n')

		if strings.HasSuffix(trimmed, ";") {
			if err := flushStatement(); err != nil {
				return 0, err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("read sql dump: %w", err)
	}
	if err := flushStatement(); err != nil {
		return 0, err
	}

	for _, backupTable := range backupTables {
		if backupTable == "" {
			continue
		}
		if _, err := tx.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", quoteIdentifier(backupTable))); err != nil {
			return 0, fmt.Errorf("drop backup table %s: %w", backupTable, err)
		}
		stats.backupsRemoved++
		if logf != nil {
			logf(fmt.Sprintf("backup removed: %s", backupTable))
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit import transaction: %w", err)
	}
	if logf != nil {
		logf(fmt.Sprintf(
			"summary: %d tables recreated, %d backups created, %d backups removed, %d insert statements executed",
			stats.tablesRecreated,
			stats.backupCreated,
			stats.backupsRemoved,
			insertedRows,
		))
	}

	return insertedRows, nil
}

func ensureRefreshTable(tx *sql.Tx, table, createStmt string, backupTables map[string]string, stats *importStats, logf ImportLogFunc) error {
	if _, prepared := backupTables[table]; prepared {
		return nil
	}

	exists, err := tableExists(tx, table)
	if err != nil {
		return fmt.Errorf("check table %s existence: %w", table, err)
	}
	if !exists {
		backupTables[table] = ""
		return nil
	}

	backupTable := "__msxdb_backup_" + sanitizeIdentifier(table)
	if _, err := tx.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", quoteIdentifier(backupTable))); err != nil {
		return fmt.Errorf("drop stale backup table %s: %w", backupTable, err)
	}

	if _, err := tx.Exec(fmt.Sprintf("ALTER TABLE %s RENAME TO %s", quoteIdentifier(table), quoteIdentifier(backupTable))); err != nil {
		return fmt.Errorf("backup table %s: %w", table, err)
	}
	if stats != nil {
		stats.backupCreated++
	}
	if logf != nil {
		logf(fmt.Sprintf("backup created: %s -> %s", table, backupTable))
	}

	if createStmt != "" {
		if _, err := tx.Exec(createStmt); err != nil {
			return fmt.Errorf("recreate table %s: %w", table, err)
		}
		if stats != nil {
			stats.tablesRecreated++
		}
		if logf != nil {
			logf(fmt.Sprintf("table recreated from dump schema: %s", table))
		}
	} else {
		// Fallback only used when dump does not include CREATE TABLE for this target.
		if _, err := tx.Exec(fmt.Sprintf("CREATE TABLE %s AS SELECT * FROM %s WHERE 1=0", quoteIdentifier(table), quoteIdentifier(backupTable))); err != nil {
			return fmt.Errorf("recreate table %s fallback: %w", table, err)
		}
		if stats != nil {
			stats.tablesRecreated++
		}
		if logf != nil {
			logf(fmt.Sprintf("table recreated from backup fallback: %s", table))
		}
	}

	backupTables[table] = backupTable
	return nil
}

func tableExists(tx *sql.Tx, table string) (bool, error) {
	var count int
	err := tx.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name = ?", table).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func quoteIdentifier(name string) string {
	return "\"" + strings.ReplaceAll(name, "\"", "\"\"") + "\""
}

func sanitizeIdentifier(name string) string {
	if name == "" {
		return "unknown"
	}

	var b strings.Builder
	b.Grow(len(name))
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
			continue
		}
		b.WriteByte('_')
	}
	return b.String()
}

func parseCreateTableTarget(stmt string) string {
	t := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(stmt, "\n", " "), "\t", " "))
	if t == "" {
		return ""
	}

	upper := strings.ToUpper(t)
	if !strings.HasPrefix(upper, "CREATE TABLE ") {
		return ""
	}

	rest := strings.TrimSpace(t[len("CREATE TABLE "):])
	restUpper := strings.ToUpper(rest)
	if strings.HasPrefix(restUpper, "IF NOT EXISTS ") {
		rest = strings.TrimSpace(rest[len("IF NOT EXISTS "):])
	}
	if rest == "" {
		return ""
	}

	end := len(rest)
	for i, r := range rest {
		if r == ' ' || r == '(' {
			end = i
			break
		}
	}
	table := strings.TrimSpace(rest[:end])
	table = strings.Trim(table, "`\"")
	return table
}

func parseInsertTargetTable(stmt string) string {
	t := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(stmt, "\n", " "), "\t", " "))
	if t == "" {
		return ""
	}

	upper := strings.ToUpper(t)
	if !strings.HasPrefix(upper, "INSERT INTO ") {
		return ""
	}

	rest := strings.TrimSpace(t[len("INSERT INTO "):])
	if rest == "" {
		return ""
	}

	end := len(rest)
	for i, r := range rest {
		if r == ' ' || r == '(' {
			end = i
			break
		}
	}
	table := strings.TrimSpace(rest[:end])
	table = strings.Trim(table, "`\"")
	return table
}

func transactionControlKeyword(stmt string) (string, bool) {
	normalized := strings.TrimSpace(stmt)
	normalized = strings.TrimSuffix(normalized, ";")
	normalized = strings.ToUpper(strings.Join(strings.Fields(normalized), " "))

	switch normalized {
	case "BEGIN", "BEGIN TRANSACTION", "BEGIN IMMEDIATE", "BEGIN EXCLUSIVE", "BEGIN DEFERRED", "COMMIT", "END", "END TRANSACTION", "ROLLBACK", "ROLLBACK TRANSACTION":
		if strings.HasPrefix(normalized, "BEGIN") {
			return "BEGIN", true
		}
		if strings.HasPrefix(normalized, "ROLLBACK") {
			return "ROLLBACK", true
		}
		if strings.HasPrefix(normalized, "END") {
			return "END", true
		}
		if strings.HasPrefix(normalized, "COMMIT") {
			return "COMMIT", true
		}
		return normalized, true
	default:
		return "", false
	}
}

func escapeSQLLike(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "%", "\\%")
	value = strings.ReplaceAll(value, "_", "\\_")
	return value
}
