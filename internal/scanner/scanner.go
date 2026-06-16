package scanner

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuukiyuuna/MangaMetaManager/internal/core"
	"github.com/yuukiyuuna/MangaMetaManager/internal/models"
)

func ScanLibrary(task *core.Task) error {
	var folders []models.LibraryFolder
	if err := models.DB.Find(&folders).Error; err != nil {
		return err
	}

	if task != nil {
		core.GlobalTaskManager.UpdateProgress(task, 0, 0, "Starting scan...")
	}

	count := 0
	for _, folder := range folders {
		err := filepath.Walk(folder.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			
			if !info.IsDir() {
				name := filepath.Base(path)
				
				// Cleanup abandoned temp files
				if strings.HasPrefix(name, "mmm-tmp-") || strings.HasPrefix(name, "mmm-raw-tmp-") {
					log.Printf("Cleaning up abandoned temp file: %s", path)
					os.Remove(path)
					return nil // Skip further processing for temp files
				}

				if IsArchive(path) {
					count++
					if task != nil {
						core.GlobalTaskManager.UpdateProgress(task, count, 0, name) // 0 total for indeterminate progress
					}
					processMangaFile(path, info)
				}
			}
			return nil
		})
		if err != nil {
			log.Printf("Error scanning folder %s: %v", folder.Path, err)
		}
	}
	return nil
}

func CleanupTempFiles() {
	var folders []models.LibraryFolder
	if err := models.DB.Find(&folders).Error; err != nil {
		return
	}

	for _, folder := range folders {
		filepath.Walk(folder.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() {
				name := filepath.Base(path)
				if strings.HasPrefix(name, "mmm-tmp-") || strings.HasPrefix(name, "mmm-raw-tmp-") {
					log.Printf("Cleaning up abandoned temp file: %s", path)
					os.Remove(path)
				}
			}
			return nil
		})
	}
}

func CleanLibrary(task *core.Task) error {
	log.Println("Starting database clean...")
	
	// 1. Clean Series
	var seriesList []models.MangaSeries
	if err := models.DB.Find(&seriesList).Error; err == nil {
		total := len(seriesList)
		for i, s := range seriesList {
			if task != nil {
				core.GlobalTaskManager.UpdateProgress(task, i, total, fmt.Sprintf("Checking series: %s", s.Title))
			}
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
		total := len(books)
		for i, b := range books {
			if task != nil {
				core.GlobalTaskManager.UpdateProgress(task, i, total, fmt.Sprintf("Checking book: %s", b.Filename))
			}
			if _, err := os.Stat(b.Path); os.IsNotExist(err) {
				log.Printf("Removing orphaned book: %s", b.Path)
				models.DB.Unscoped().Delete(&b)
			}
		}
	}

	log.Println("Database clean completed.")
	return nil
}

func processMangaFile(path string, info os.FileInfo) {
	dir := filepath.Dir(path)
	filename := filepath.Base(path)
	
	var series models.MangaSeries
	result := models.DB.Where("path = ?", dir).First(&series)
	if result.Error != nil {
		series = models.MangaSeries{
			Path: dir,
			MangaBase: models.MangaBase{
				Title: filepath.Base(dir),
			},
		}
		models.DB.Create(&series)
	}

	var book models.MangaBook
	result = models.DB.Where("path = ?", path).First(&book)
	
	// Incremental scan: check if file has changed
	if result.Error == nil && book.FileModTime == info.ModTime().Unix() && book.FileSize == info.Size() {
		return
	}

	comicInfo, err := ReadComicInfo(path)
	if err != nil {
		log.Printf("Error reading ComicInfo from %s: %v", path, err)
	}

	newBook := models.MangaBook{
		SeriesID:    series.ID,
		Path:        path,
		Filename:    filename,
		FileModTime: info.ModTime().Unix(),
		FileSize:    info.Size(),
		MangaBase: models.MangaBase{
			Title: strings.TrimSuffix(filename, filepath.Ext(filename)),
			Type:  "漫画", // Default
		},
	}

	if comicInfo != nil {
		if comicInfo.Title != "" {
			newBook.Title = comicInfo.Title
		}
		newBook.Volume = float64(comicInfo.Volume)
		newBook.Author = comicInfo.Writer
		newBook.Summary = comicInfo.Summary
		newBook.Status = "Scraped"
		
		// Map back ComicInfo.Manga to our Type
		if comicInfo.Manga == "No" {
			newBook.Type = "小说"
		}
	}

	if result.Error != nil {
		models.DB.Create(&newBook)
	} else {
		models.DB.Model(&book).Updates(newBook)
	}

	// Sync series metadata to DB if available
	if comicInfo != nil && comicInfo.Series != "" {
		series.Series = comicInfo.Series
		series.Author = comicInfo.Writer
		series.Summary = comicInfo.Summary
		series.Genre = comicInfo.Genre
		series.Status = "Scraped"
		models.DB.Save(&series)
	}
}
