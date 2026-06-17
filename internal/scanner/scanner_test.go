package scanner

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/yuukiyuuna/MangaMetaManager/internal/models"
)

func TestScanLibraryRestoresSoftDeletedSeriesAndBook(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "scan.db")
	models.InitDB(dbPath)

	seriesPath := filepath.Join(tmpDir, "series")
	if err := os.Mkdir(seriesPath, 0755); err != nil {
		t.Fatal(err)
	}
	bookPath := filepath.Join(seriesPath, "book.cbz")
	createScanTestArchive(t, bookPath)

	series := models.MangaSeries{Path: seriesPath, MangaBase: models.MangaBase{Title: "series"}}
	if err := models.DB.Create(&series).Error; err != nil {
		t.Fatal(err)
	}
	book := models.MangaBook{SeriesID: series.ID, Path: bookPath, Filename: "book.cbz"}
	if err := models.DB.Create(&book).Error; err != nil {
		t.Fatal(err)
	}
	if err := models.DB.Delete(&book).Error; err != nil {
		t.Fatal(err)
	}
	if err := models.DB.Delete(&series).Error; err != nil {
		t.Fatal(err)
	}
	if err := models.DB.Create(&models.LibraryFolder{Path: tmpDir}).Error; err != nil {
		t.Fatal(err)
	}

	if err := ScanLibrary(nil); err != nil {
		t.Fatal(err)
	}

	var activeSeriesCount int64
	if err := models.DB.Model(&models.MangaSeries{}).Where("path = ?", seriesPath).Count(&activeSeriesCount).Error; err != nil {
		t.Fatal(err)
	}
	if activeSeriesCount != 1 {
		t.Fatalf("expected one active series after scan, got %d", activeSeriesCount)
	}

	var activeBookCount int64
	if err := models.DB.Model(&models.MangaBook{}).Where("path = ?", bookPath).Count(&activeBookCount).Error; err != nil {
		t.Fatal(err)
	}
	if activeBookCount != 1 {
		t.Fatalf("expected one active book after scan, got %d", activeBookCount)
	}
}

func createScanTestArchive(t *testing.T, archivePath string) {
	t.Helper()
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	entry, err := writer.Create("001.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("page")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}
