package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
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
