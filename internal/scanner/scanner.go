package scanner

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuukiyuuna/MangaMetaManager/internal/core"
	"github.com/yuukiyuuna/MangaMetaManager/internal/models"
	"gorm.io/gorm"
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
					if err := os.Remove(path); err != nil {
						log.Printf("Failed to remove abandoned temp file %s: %v", path, err)
					}
					return nil // Skip further processing for temp files
				}

				if IsArchive(path) {
					count++
					if task != nil {
						core.GlobalTaskManager.UpdateProgress(task, count, 0, name) // 0 total for indeterminate progress
					}
					if err := processMangaFile(path, info); err != nil {
						return err
					}
				}
			}
			return nil
		})
		if err != nil {
			log.Printf("Error scanning folder %s: %v", folder.Path, err)
			return err
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
		if err := filepath.Walk(folder.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() {
				name := filepath.Base(path)
				if strings.HasPrefix(name, "mmm-tmp-") || strings.HasPrefix(name, "mmm-raw-tmp-") {
					log.Printf("Cleaning up abandoned temp file: %s", path)
					if err := os.Remove(path); err != nil {
						log.Printf("Failed to remove abandoned temp file %s: %v", path, err)
					}
				}
			}
			return nil
		}); err != nil {
			log.Printf("Error cleaning temp files in %s: %v", folder.Path, err)
		}
	}
}

func CleanLibrary(task *core.Task) error {
	log.Println("Starting database clean...")

	// 1. Clean Series
	var seriesList []models.MangaSeries
	if err := models.DB.Find(&seriesList).Error; err != nil {
		return err
	} else {
		total := len(seriesList)
		for i, s := range seriesList {
			if task != nil {
				core.GlobalTaskManager.UpdateProgress(task, i, total, fmt.Sprintf("Checking series: %s", s.Title))
			}
			if _, err := os.Stat(s.Path); os.IsNotExist(err) {
				log.Printf("Removing orphaned series: %s", s.Path)
				if err := models.DB.Transaction(func(tx *gorm.DB) error {
					if err := tx.Unscoped().Where("series_id = ?", s.ID).Delete(&models.MangaBook{}).Error; err != nil {
						return err
					}
					return tx.Unscoped().Delete(&s).Error
				}); err != nil {
					return err
				}
			}
		}
	}

	// 2. Clean Books
	var books []models.MangaBook
	if err := models.DB.Find(&books).Error; err != nil {
		return err
	} else {
		total := len(books)
		for i, b := range books {
			if task != nil {
				core.GlobalTaskManager.UpdateProgress(task, i, total, fmt.Sprintf("Checking book: %s", b.Filename))
			}
			if _, err := os.Stat(b.Path); os.IsNotExist(err) {
				log.Printf("Removing orphaned book: %s", b.Path)
				if err := models.DB.Unscoped().Delete(&b).Error; err != nil {
					return err
				}
			}
		}
	}

	log.Println("Database clean completed.")
	return nil
}

func processMangaFile(path string, info os.FileInfo) error {
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
		if err := models.DB.Create(&series).Error; err != nil {
			return err
		}
	}

	var book models.MangaBook
	result = models.DB.Where("path = ?", path).First(&book)

	// Incremental scan: check if file has changed
	if result.Error == nil && book.FileModTime == info.ModTime().Unix() && book.FileSize == info.Size() {
		return nil
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
		if err := models.DB.Create(&newBook).Error; err != nil {
			return err
		}
	} else {
		if err := models.DB.Model(&book).Updates(newBook).Error; err != nil {
			return err
		}
	}

	// Sync series metadata to DB if available
	if comicInfo != nil && comicInfo.Series != "" {
		series.Series = comicInfo.Series
		series.Author = comicInfo.Writer
		series.Summary = comicInfo.Summary
		series.Genre = comicInfo.Genre
		series.Status = "Scraped"
		if err := models.DB.Save(&series).Error; err != nil {
			return err
		}
	}

	return nil
}
