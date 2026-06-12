package scanner

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/openclaw/MangaMetaManager/internal/metadata"
	"github.com/openclaw/MangaMetaManager/internal/models"
)

func ScanLibrary() error {
	var folders []models.LibraryFolder
	if err := models.DB.Find(&folders).Error; err != nil {
		return err
	}

	for _, folder := range folders {
		err := filepath.Walk(folder.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && IsArchive(path) {
				processMangaFile(path)
			}
			return nil
		})
		if err != nil {
			log.Printf("Error scanning folder %s: %v", folder.Path, err)
		}
	}
	return nil
}

func CleanLibrary() error {
	log.Println("Starting database clean...")
	
	// 1. Clean Series
	var seriesList []models.MangaSeries
	if err := models.DB.Find(&seriesList).Error; err == nil {
		for _, s := range seriesList {
			if _, err := os.Stat(s.Path); os.IsNotExist(err) {
				log.Printf("Removing orphaned series: %s", s.Path)
				models.DB.Unscoped().Delete(&s)
				models.DB.Unscoped().Where("series_id = ?", s.ID).Delete(&models.MangaBook{})
			}
		}
	}

	// 2. Clean Books
	var books []models.MangaBook
	if err := models.DB.Find(&books).Error; err == nil {
		for _, b := range books {
			if _, err := os.Stat(b.Path); os.IsNotExist(err) {
				log.Printf("Removing orphaned book: %s", b.Path)
				models.DB.Unscoped().Delete(&b)
			}
		}
	}

	log.Println("Database clean completed.")
	return nil
}

func processMangaFile(path string) {
	dir := filepath.Dir(path)
	filename := filepath.Base(path)
	
	var series models.MangaSeries
	result := models.DB.Where("path = ?", dir).First(&series)
	if result.Error != nil {
		series = models.MangaSeries{
			Path:  dir,
			Title: filepath.Base(dir),
		}
		models.DB.Create(&series)
	}

	var book models.MangaBook
	result = models.DB.Where("path = ?", path).First(&book)
	
	info, err := ReadComicInfo(path)
	if err != nil {
		log.Printf("Error reading ComicInfo from %s: %v", path, err)
	}

	newBook := models.MangaBook{
		SeriesID: series.ID,
		Path:     path,
		Filename: filename,
		Title:    strings.TrimSuffix(filename, filepath.Ext(filename)),
	}

	if info != nil {
		if info.Title != "" {
			newBook.Title = info.Title
		}
		newBook.Volume = info.Volume
		newBook.Author = info.Writer
		newBook.Summary = info.Summary
		newBook.Status = "Scraped"
		
		updateSeriesFromComicInfo(&series, info)
	}

	if result.Error != nil {
		models.DB.Create(&newBook)
	} else {
		models.DB.Model(&book).Updates(newBook)
	}
}

func updateSeriesFromComicInfo(series *models.MangaSeries, info *metadata.ComicInfo) {
	changed := false
	if series.Series == "" && info.Series != "" {
		series.Series = info.Series
		changed = true
	}
	if series.Author == "" && info.Writer != "" {
		series.Author = info.Writer
		changed = true
	}
	if series.Genre == "" && info.Genre != "" {
		series.Genre = info.Genre
		changed = true
	}
	if series.Summary == "" && info.Summary != "" {
		series.Summary = info.Summary
		changed = true
	}
	
	if changed {
		series.Status = "Scraped"
		models.DB.Save(series)
	}
}
