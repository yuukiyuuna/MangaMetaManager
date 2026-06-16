package scanner

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuukiyuuna/MangaMetaManager/internal/metadata"
)

func TestWriteComicInfo_Flattening(t *testing.T) {
	// 1. Create a dummy ZIP with nested structure and duplicate filenames
	tmpZip := filepath.Join(t.TempDir(), "test.zip")
	f, err := os.Create(tmpZip)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)

	// Add files:
	// vol1/01.jpg
	// vol2/01.jpg
	// vol2/02.jpg
	files := []string{"vol1/01.jpg", "vol2/01.jpg", "vol2/02.jpg"}
	for _, name := range files {
		w, _ := zw.Create(name)
		w.Write([]byte("image content"))
	}
	zw.Close()
	f.Close()

	// 2. Call WriteComicInfo
	info := &metadata.ComicInfo{
		Title: "Test Manga",
	}
	err = WriteComicInfo(tmpZip, info, false)
	if err != nil {
		t.Fatalf("WriteComicInfo failed: %v", err)
	}

	// 3. Verify structure
	r, err := zip.OpenReader(tmpZip)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	expectedFiles := map[string]bool{
		"ComicInfo.xml": false,
		"01.jpg":        false, // from vol1
		"vol2_01.jpg":   false, // from vol2 (renamed due to collision)
		"02.jpg":        false, // from vol2
	}

	for _, f := range r.File {
		if _, ok := expectedFiles[f.Name]; ok {
			expectedFiles[f.Name] = true
		} else {
			t.Errorf("Unexpected file in ZIP: %s", f.Name)
		}
	}

	for name, found := range expectedFiles {
		if !found {
			t.Errorf("Expected file %s not found in ZIP", name)
		}
	}
}

func TestWriteComicInfo_Backup(t *testing.T) {
	tmpZip := filepath.Join(t.TempDir(), "test_backup.zip")

	// Create a valid empty ZIP
	f, _ := os.Create(tmpZip)
	zw := zip.NewWriter(f)
	w, _ := zw.Create("test.txt")
	w.Write([]byte("test"))
	zw.Close()
	f.Close()

	info := &metadata.ComicInfo{Title: "Backup Test"}
	err := WriteComicInfo(tmpZip, info, true)
	if err != nil {
		t.Fatalf("WriteComicInfo failed: %v", err)
	}

	backups, err := filepath.Glob(tmpZip + ".*.bak")
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 1 {
		t.Fatalf("expected one timestamped backup, got %d", len(backups))
	}
	if !strings.HasPrefix(filepath.Base(backups[0]), filepath.Base(tmpZip)+".") || !strings.HasSuffix(backups[0], ".bak") {
		t.Fatalf("unexpected backup name: %s", backups[0])
	}
}

func TestWriteComicInfo_BackupDoesNotOverwrite(t *testing.T) {
	tmpZip := filepath.Join(t.TempDir(), "test_backup_multi.zip")
	f, err := os.Create(tmpZip)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("test")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	info := &metadata.ComicInfo{Title: "Backup Test"}
	if err := WriteComicInfo(tmpZip, info, true); err != nil {
		t.Fatalf("first WriteComicInfo failed: %v", err)
	}
	if err := WriteComicInfo(tmpZip, info, true); err != nil {
		t.Fatalf("second WriteComicInfo failed: %v", err)
	}

	backups, err := filepath.Glob(tmpZip + ".*.bak")
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 2 {
		t.Fatalf("expected two backups, got %d: %v", len(backups), backups)
	}
}
