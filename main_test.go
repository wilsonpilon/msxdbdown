package main

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/ulikunitz/xz/lzma"
)

func TestExtractZipFlatExtractsFilesIntoDestination(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "sql-msxromdb.zip")
	destinationDir := filepath.Join(tempDir, "download")

	file, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}

	writer := zip.NewWriter(file)
	entry, err := writer.Create("nested/sql-msxromdb.sql")
	if err != nil {
		t.Fatalf("create zip entry: %v", err)
	}
	if _, err := entry.Write([]byte("CREATE TABLE demo(id INTEGER);")); err != nil {
		t.Fatalf("write zip entry: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close zip file: %v", err)
	}

	extracted, err := extractZipFlat(zipPath, destinationDir)
	if err != nil {
		t.Fatalf("extract zip: %v", err)
	}

	if len(extracted) != 1 {
		t.Fatalf("expected 1 extracted file, got %d", len(extracted))
	}

	gotPath := extracted[0]
	wantPath := filepath.Join(destinationDir, "sql-msxromdb.sql")
	if gotPath != wantPath {
		t.Fatalf("expected extracted path %q, got %q", wantPath, gotPath)
	}

	content, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatalf("read extracted file: %v", err)
	}
	if string(content) != "CREATE TABLE demo(id INTEGER);" {
		t.Fatalf("unexpected extracted content: %q", string(content))
	}
}

func TestExtractZipFlatFailsWhenArchiveHasNoFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "empty.zip")
	destinationDir := filepath.Join(tempDir, "download")

	file, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}

	writer := zip.NewWriter(file)
	if _, err := writer.Create("folder/"); err != nil {
		t.Fatalf("create dir entry: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close zip file: %v", err)
	}

	if _, err := extractZipFlat(zipPath, destinationDir); err == nil {
		t.Fatal("expected extractZipFlat to fail for archive without files")
	}
}

func TestOpenZipLZMAReaderDecodesPayload(t *testing.T) {
	t.Parallel()

	want := []byte("CREATE TABLE demo(id INTEGER);")

	var lzmaStream bytes.Buffer
	writer, err := lzma.NewWriter(&lzmaStream)
	if err != nil {
		t.Fatalf("create lzma writer: %v", err)
	}
	if _, err := writer.Write(want); err != nil {
		t.Fatalf("write lzma payload: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close lzma writer: %v", err)
	}

	rawLZMA := lzmaStream.Bytes()
	if len(rawLZMA) < lzma.HeaderLen {
		t.Fatalf("lzma stream too short: %d", len(rawLZMA))
	}

	zipPayload := make([]byte, 0, 4+5+len(rawLZMA)-lzma.HeaderLen)
	zipPayload = append(zipPayload, 0x01, 0x00, 0x05, 0x00)
	zipPayload = append(zipPayload, rawLZMA[:5]...)
	zipPayload = append(zipPayload, rawLZMA[lzma.HeaderLen:]...)

	reader, err := openZipLZMAReader(bytes.NewReader(zipPayload), uint64(len(want)))
	if err != nil {
		t.Fatalf("open zip lzma reader: %v", err)
	}
	t.Cleanup(func() { _ = reader.Close() })

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read decoded payload: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("decoded payload mismatch: got %q want %q", string(got), string(want))
	}
}

func TestExtractZipFlatSupportsBundledLZMAArchive(t *testing.T) {
	t.Parallel()

	zipPath := filepath.Join("download", "sql-msxromdb.zip")
	if _, err := os.Stat(zipPath); err != nil {
		t.Skipf("bundled archive unavailable: %v", err)
	}

	destinationDir := filepath.Join(t.TempDir(), "download")
	extracted, err := extractZipFlat(zipPath, destinationDir)
	if err != nil {
		t.Fatalf("extract bundled lzma zip: %v", err)
	}

	if len(extracted) == 0 {
		t.Fatal("expected at least one extracted file")
	}

	foundSQL := false
	for _, p := range extracted {
		if filepath.Base(p) == "sql-romdb.sql" || filepath.Base(p) == "sql-msxromdb.sql" {
			foundSQL = true
			break
		}
	}
	if !foundSQL {
		t.Fatalf("expected extracted SQL dump in %v", extracted)
	}
}

func TestCompareYearValuesUsesNumericOrdering(t *testing.T) {
	t.Parallel()

	if got := compareYearValues("1999", "2001"); got >= 0 {
		t.Fatalf("expected 1999 < 2001, got %d", got)
	}
	if got := compareYearValues("2001", "1999"); got <= 0 {
		t.Fatalf("expected 2001 > 1999, got %d", got)
	}
	if got := compareYearValues("1987 (JP)", "1988"); got >= 0 {
		t.Fatalf("expected parsed 1987 < 1988, got %d", got)
	}
	if got := compareYearValues("unknown", "1988"); got <= 0 {
		t.Fatalf("expected non-numeric year to sort after numeric, got %d", got)
	}
}

func TestParseYearValue(t *testing.T) {
	t.Parallel()

	year, ok := parseYearValue(" 1986 ")
	if !ok || year != 1986 {
		t.Fatalf("expected 1986, got year=%d ok=%v", year, ok)
	}

	year, ok = parseYearValue("1987 (JP)")
	if !ok || year != 1987 {
		t.Fatalf("expected parsed 1987, got year=%d ok=%v", year, ok)
	}

	if _, ok := parseYearValue("N/A"); ok {
		t.Fatal("expected parse failure for non-numeric value")
	}
}
