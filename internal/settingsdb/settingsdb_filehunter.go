package settingsdb

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ─── Types ───────────────────────────────────────────────────────────────────

// FHFileRow is a single File-Hunter catalog entry returned to the UI.
type FHFileRow struct {
	ID        int64
	Name      string
	FullPath  string
	Extension string
	SHA1      string
}

// FHCategoryItem is a directory component (category) with a matching file count.
type FHCategoryItem struct {
	ID    int64
	Name  string
	Count int
}

// FHFileTypeItem is a file extension with a file count.
type FHFileTypeItem struct {
	ID        int64
	Extension string
	Count     int
}

// ─── Schema ──────────────────────────────────────────────────────────────────

const fileHunterSchema = `
CREATE TABLE IF NOT EXISTS fh_category (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT    NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS fh_file_type (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    extension TEXT    NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS fh_file (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    name         TEXT    NOT NULL,
    full_path    TEXT    NOT NULL UNIQUE,
    file_type_id INTEGER REFERENCES fh_file_type(id),
    sha1         TEXT
);

CREATE TABLE IF NOT EXISTS fh_file_category (
    file_id     INTEGER NOT NULL REFERENCES fh_file(id) ON DELETE CASCADE,
    category_id INTEGER NOT NULL REFERENCES fh_category(id),
    position    INTEGER NOT NULL,
    PRIMARY KEY (file_id, category_id, position)
);

CREATE INDEX IF NOT EXISTS idx_fh_file_name     ON fh_file(name COLLATE NOCASE);
CREATE INDEX IF NOT EXISTS idx_fh_file_fullpath ON fh_file(full_path);
CREATE INDEX IF NOT EXISTS idx_fh_file_sha1     ON fh_file(sha1);
CREATE INDEX IF NOT EXISTS idx_fh_fc_cat_pos    ON fh_file_category(category_id, position);
CREATE INDEX IF NOT EXISTS idx_fh_fc_file       ON fh_file_category(file_id);
`

// InitFileHunterSchema creates the fh_* tables when they don't yet exist.
func (s *Store) InitFileHunterSchema() error {
	if s == nil || s.db == nil {
		return errors.New("nil store")
	}
	if _, err := s.db.Exec(fileHunterSchema); err != nil {
		return fmt.Errorf("init fh schema: %w", err)
	}
	return nil
}

// ─── Import: allfiles.txt ─────────────────────────────────────────────────────

// ImportFileHunterAllFiles parses download/allfiles.txt and populates the
// fh_file, fh_category, fh_file_type and fh_file_category tables.
// All existing fh_* data is replaced on each import.
func (s *Store) ImportFileHunterAllFiles(path string, logf ImportLogFunc) (int, error) {
	if s == nil || s.db == nil {
		return 0, errors.New("nil store")
	}
	if err := s.InitFileHunterSchema(); err != nil {
		return 0, err
	}

	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open allfiles: %w", err)
	}
	defer func() { _ = f.Close() }()

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin fh import: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Truncate existing data (order respects FK constraints)
	for _, tbl := range []string{"fh_file_category", "fh_file", "fh_file_type", "fh_category"} {
		if _, err := tx.Exec("DELETE FROM " + tbl); err != nil {
			return 0, fmt.Errorf("truncate %s: %w", tbl, err)
		}
	}

	// In-memory caches to avoid repeated round-trips for repeated category/type names
	catCache := map[string]int64{}
	typeCache := map[string]int64{}

	getCatID := func(name string) (int64, error) {
		if id, ok := catCache[name]; ok {
			return id, nil
		}
		res, err := tx.Exec("INSERT OR IGNORE INTO fh_category(name) VALUES(?)", name)
		if err != nil {
			return 0, err
		}
		id, _ := res.LastInsertId()
		if id == 0 {
			_ = tx.QueryRow("SELECT id FROM fh_category WHERE name = ?", name).Scan(&id)
		}
		catCache[name] = id
		return id, nil
	}

	getTypeID := func(ext string) (int64, error) {
		if id, ok := typeCache[ext]; ok {
			return id, nil
		}
		res, err := tx.Exec("INSERT OR IGNORE INTO fh_file_type(extension) VALUES(?)", ext)
		if err != nil {
			return 0, err
		}
		id, _ := res.LastInsertId()
		if id == 0 {
			_ = tx.QueryRow("SELECT id FROM fh_file_type WHERE extension = ?", ext).Scan(&id)
		}
		typeCache[ext] = id
		return id, nil
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)

	inserted := 0
	lineNum := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNum++
		if line == "" {
			continue
		}

		// Normalize path separators to backslash
		line = strings.ReplaceAll(line, "/", `\`)

		parts := strings.Split(line, `\`)
		filename := parts[len(parts)-1]
		categories := parts[:len(parts)-1]

		// Derive clean name (no extension) and extension
		extWithDot := filepath.Ext(filename)
		ext := strings.ToLower(strings.TrimPrefix(extWithDot, "."))
		name := strings.TrimSuffix(filename, extWithDot)
		if name == "" {
			name = filename
		}

		var fileTypeID *int64
		if ext != "" {
			tid, err := getTypeID(ext)
			if err != nil {
				return inserted, fmt.Errorf("line %d get type %q: %w", lineNum, ext, err)
			}
			fileTypeID = &tid
		}

		res, err := tx.Exec(
			"INSERT OR IGNORE INTO fh_file(name, full_path, file_type_id) VALUES(?,?,?)",
			name, line, fileTypeID,
		)
		if err != nil {
			return inserted, fmt.Errorf("line %d insert file: %w", lineNum, err)
		}
		fileID, _ := res.LastInsertId()
		if fileID == 0 {
			// Duplicate full_path — skip
			continue
		}

		for pos, cat := range categories {
			cat = strings.TrimSpace(cat)
			if cat == "" {
				continue
			}
			catID, err := getCatID(cat)
			if err != nil {
				return inserted, fmt.Errorf("line %d get cat %q: %w", lineNum, cat, err)
			}
			if _, err := tx.Exec(
				"INSERT OR IGNORE INTO fh_file_category(file_id, category_id, position) VALUES(?,?,?)",
				fileID, catID, pos,
			); err != nil {
				return inserted, fmt.Errorf("line %d insert cat rel: %w", lineNum, err)
			}
		}

		inserted++
		if logf != nil && inserted%5000 == 0 {
			logf(fmt.Sprintf("  ... %d files imported so far", inserted))
		}
	}

	if err := scanner.Err(); err != nil {
		return inserted, fmt.Errorf("read allfiles: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return inserted, fmt.Errorf("commit fh import: %w", err)
	}
	if logf != nil {
		logf(fmt.Sprintf("import complete: %d files inserted (%d lines read)", inserted, lineNum))
	}
	return inserted, nil
}

// ─── Import: sha1sums.txt ─────────────────────────────────────────────────────

// ImportFileHunterSHA1Sums reads sha1sums.txt and updates fh_file.sha1.
// Format per line:  <SHA1HEX>  .\path\to\file.ext
func (s *Store) ImportFileHunterSHA1Sums(path string, logf ImportLogFunc) (int, error) {
	if s == nil || s.db == nil {
		return 0, errors.New("nil store")
	}
	if err := s.InitFileHunterSchema(); err != nil {
		return 0, err
	}

	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open sha1sums: %w", err)
	}
	defer func() { _ = f.Close() }()

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin sha1 update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)

	updated := 0
	lineNum := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNum++
		if line == "" {
			continue
		}

		// Split on first double-space (standard sha1sum format) or tab
		var sha1, rawPath string
		if idx := strings.Index(line, "  "); idx > 0 {
			sha1 = strings.TrimSpace(line[:idx])
			rawPath = strings.TrimSpace(line[idx+2:])
		} else if idx := strings.Index(line, "\t"); idx > 0 {
			sha1 = strings.TrimSpace(line[:idx])
			rawPath = strings.TrimSpace(line[idx+1:])
		} else {
			continue
		}

		// Normalize path: remove leading .\  or  ./
		rawPath = strings.TrimPrefix(rawPath, `.\`)
		rawPath = strings.TrimPrefix(rawPath, `./`)
		rawPath = strings.ReplaceAll(rawPath, "/", `\`)

		res, err := tx.Exec("UPDATE fh_file SET sha1 = ? WHERE full_path = ?", sha1, rawPath)
		if err != nil {
			return updated, fmt.Errorf("line %d update sha1: %w", lineNum, err)
		}
		if n, _ := res.RowsAffected(); n > 0 {
			updated++
		}
	}

	if err := scanner.Err(); err != nil {
		return updated, fmt.Errorf("read sha1sums: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return updated, fmt.Errorf("commit sha1 update: %w", err)
	}
	if logf != nil {
		logf(fmt.Sprintf("sha1 update complete: %d files updated (%d lines read)", updated, lineNum))
	}
	return updated, nil
}

// ─── Query helpers ────────────────────────────────────────────────────────────

// CountFHFiles returns the total number of imported File-Hunter files.
// Returns 0 if the table doesn't exist yet.
func (s *Store) CountFHFiles() int {
	if s == nil || s.db == nil {
		return 0
	}
	var n int
	_ = s.db.QueryRow("SELECT COUNT(*) FROM fh_file").Scan(&n)
	return n
}

// ListFHCategories returns available category names at the next depth level
// given a path filter.  pathFilter = nil → return top-level categories (position 0).
func (s *Store) ListFHCategories(pathFilter []string) ([]FHCategoryItem, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("nil store")
	}

	nextPos := len(pathFilter)

	var query string
	var args []any

	if nextPos == 0 {
		query = `
SELECT c.id, c.name, COUNT(DISTINCT fc.file_id)
FROM fh_category c
JOIN fh_file_category fc ON fc.category_id = c.id AND fc.position = 0
GROUP BY c.id, c.name
ORDER BY c.name COLLATE NOCASE`
	} else {
		subQ, subArgs := buildFHPathQuery(pathFilter)
		query = fmt.Sprintf(`
SELECT c.id, c.name, COUNT(DISTINCT fc_n.file_id)
FROM fh_file_category fc_n
JOIN fh_category c ON c.id = fc_n.category_id
WHERE fc_n.position = ? AND fc_n.file_id IN (%s)
GROUP BY c.id, c.name
ORDER BY c.name COLLATE NOCASE`, subQ)
		args = append([]any{nextPos}, subArgs...)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list fh categories at depth %d: %w", nextPos, err)
	}
	defer func() { _ = rows.Close() }()

	var result []FHCategoryItem
	for rows.Next() {
		var item FHCategoryItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Count); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

// ListFHFileTypes returns available file extensions, optionally filtered by path.
func (s *Store) ListFHFileTypes(pathFilter []string) ([]FHFileTypeItem, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("nil store")
	}

	var query string
	var args []any

	if len(pathFilter) == 0 {
		query = `
SELECT ft.id, ft.extension, COUNT(*)
FROM fh_file_type ft
JOIN fh_file f ON f.file_type_id = ft.id
GROUP BY ft.id, ft.extension
ORDER BY ft.extension COLLATE NOCASE`
	} else {
		subQ, subArgs := buildFHPathQuery(pathFilter)
		query = fmt.Sprintf(`
SELECT ft.id, ft.extension, COUNT(*)
FROM fh_file_type ft
JOIN fh_file f ON f.file_type_id = ft.id
WHERE f.id IN (%s)
GROUP BY ft.id, ft.extension
ORDER BY ft.extension COLLATE NOCASE`, subQ)
		args = subArgs
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list fh file types: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []FHFileTypeItem
	for rows.Next() {
		var item FHFileTypeItem
		if err := rows.Scan(&item.ID, &item.Extension, &item.Count); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

// SearchFHFiles returns files matching the given path filter, name pattern and/or extension.
func (s *Store) SearchFHFiles(pathFilter []string, nameQuery, extension string, limit int) ([]FHFileRow, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("nil store")
	}
	if limit <= 0 {
		limit = 500
	}

	var conditions []string
	var args []any

	if len(pathFilter) > 0 {
		subQ, subArgs := buildFHPathQuery(pathFilter)
		conditions = append(conditions, "f.id IN ("+subQ+")")
		args = append(args, subArgs...)
	}

	if q := strings.TrimSpace(nameQuery); q != "" {
		conditions = append(conditions, `f.name LIKE ? ESCAPE '\'`)
		args = append(args, "%"+escapeSQLLike(q)+"%")
	}

	if ext := strings.ToLower(strings.TrimSpace(extension)); ext != "" {
		conditions = append(conditions, "LOWER(COALESCE(ft.extension,'')) = ?")
		args = append(args, ext)
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`
SELECT f.id, f.name, f.full_path,
       COALESCE(ft.extension, ''),
       COALESCE(f.sha1, '')
FROM fh_file f
LEFT JOIN fh_file_type ft ON ft.id = f.file_type_id
%s
ORDER BY f.name COLLATE NOCASE
LIMIT ?`, where)

	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("search fh files: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []FHFileRow
	for rows.Next() {
		var row FHFileRow
		if err := rows.Scan(&row.ID, &row.Name, &row.FullPath, &row.Extension, &row.SHA1); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

// buildFHPathQuery returns a SELECT that returns file IDs matching the given path.
// pathFilter = ["Games", "MSX1"] → files that have Games at position 0 AND MSX1 at position 1.
func buildFHPathQuery(pathFilter []string) (string, []any) {
	if len(pathFilter) == 0 {
		return "SELECT id FROM fh_file", nil
	}

	var sb strings.Builder
	var args []any

	sb.WriteString("SELECT DISTINCT f0.id FROM fh_file f0")
	for i, cat := range pathFilter {
		sb.WriteString(fmt.Sprintf(
			"\nJOIN fh_file_category ffc%d ON ffc%d.file_id = f0.id AND ffc%d.position = %d"+
				"\nJOIN fh_category      fc%d  ON fc%d.id = ffc%d.category_id AND fc%d.name = ?",
			i, i, i, i,
			i, i, i, i,
		))
		args = append(args, cat)
	}

	return sb.String(), args
}
