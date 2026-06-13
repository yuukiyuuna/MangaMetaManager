package scanner

import (
	"archive/zip"
	"os"
	"path/filepath"
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

	// Check if .bak exists
	if _, err := os.Stat(tmpZip + ".bak"); os.IsNotExist(err) {
		t.Errorf("Backup file .bak was not created")
	}
}
