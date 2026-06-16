package scanner

import (
	"archive/zip"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/yuukiyuuna/MangaMetaManager/internal/core"
	"github.com/yuukiyuuna/MangaMetaManager/internal/models"
)

func TestScanLibraryDoesNotModifyArchive(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "scan.db")
	models.InitDB(dbPath)

	archivePath := filepath.Join(tmpDir, "book.cbz")
	createTestArchive(t, archivePath, map[string]string{
		"pages/001.jpg": "page one",
		"pages/002.jpg": "page two",
	})

	beforeInfo, err := os.Stat(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	beforeEntries := zipEntries(t, archivePath)

	if err := models.DB.Create(&models.LibraryFolder{Path: tmpDir}).Error; err != nil {
		t.Fatal(err)
	}
	if err := ScanLibrary(nil); err != nil {
		t.Fatal(err)
	}

	afterInfo, err := os.Stat(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	if beforeInfo.Size() != afterInfo.Size() || !beforeInfo.ModTime().Equal(afterInfo.ModTime()) {
		t.Fatalf("scan modified archive metadata: before size=%d mod=%s after size=%d mod=%s",
			beforeInfo.Size(), beforeInfo.ModTime(), afterInfo.Size(), afterInfo.ModTime())
	}

	afterEntries := zipEntries(t, archivePath)
	if len(beforeEntries) != len(afterEntries) {
		t.Fatalf("scan changed archive entry count: before=%v after=%v", beforeEntries, afterEntries)
	}
	for name := range beforeEntries {
		if !afterEntries[name] {
			t.Fatalf("scan removed archive entry %s; after=%v", name, afterEntries)
		}
	}
	if afterEntries["ComicInfo.xml"] {
		t.Fatal("scan wrote ComicInfo.xml into archive")
	}
	backups, err := filepath.Glob(archivePath + ".*.bak")
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 0 {
		t.Fatalf("scan created backup files unexpectedly: %v", backups)
	}
}

func TestScanLibraryReportsArchiveTotal(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "scan-progress.db")
	models.InitDB(dbPath)
	core.GlobalTaskManager = &core.TaskManager{}

	createTestArchive(t, filepath.Join(tmpDir, "book-1.cbz"), map[string]string{"001.jpg": "one"})
	createTestArchive(t, filepath.Join(tmpDir, "book-2.zip"), map[string]string{"001.jpg": "two"})
	if err := os.WriteFile(filepath.Join(tmpDir, "notes.txt"), []byte("ignored"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := models.DB.Create(&models.LibraryFolder{Path: tmpDir}).Error; err != nil {
		t.Fatal(err)
	}

	task := &core.Task{}
	if err := ScanLibrary(task); err != nil {
		t.Fatal(err)
	}

	if task.Total != 2 {
		t.Fatalf("expected total archive count 2, got %d", task.Total)
	}
	if task.Progress != 2 {
		t.Fatalf("expected progress 2, got %d", task.Progress)
	}
}

func createTestArchive(t *testing.T, archivePath string, files map[string]string) {
	t.Helper()
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	for name, content := range files {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func zipEntries(t *testing.T, archivePath string) map[string]bool {
	t.Helper()
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	entries := make(map[string]bool, len(reader.File))
	for _, file := range reader.File {
		entries[path.Clean(file.Name)] = true
	}
	return entries
}
