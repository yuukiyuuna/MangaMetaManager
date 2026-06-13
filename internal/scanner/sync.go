package scanner

import (
	"encoding/xml"
	"os"
	"path/filepath"

	"github.com/yuukiyuuna/MangaMetaManager/internal/metadata"
	"github.com/yuukiyuuna/MangaMetaManager/internal/models"
	"github.com/yuukiyuuna/MangaMetaManager/internal/provider"
)

// SyncBookMetadata synchronizes database book record to ComicInfo.xml and performs flattening.
func SyncBookMetadata(book *models.MangaBook, backup bool) error {
	bookPath := book.Path
	
	existing, _ := ReadComicInfo(bookPath)
	if existing == nil {
		existing = &metadata.ComicInfo{}
	}

	// Map DB fields to ComicInfo
	existing.Title = book.Title
	existing.OriginalTitle = book.OriginalTitle
	existing.Series = book.Series
	existing.Number = book.Number
	existing.Writer = book.Author
	existing.Translator = book.Translator
	existing.Publisher = book.Publisher
	existing.Genre = book.Genre
	existing.Tags = book.Tags
	existing.Summary = book.Summary
	existing.Volume = book.Volume
	existing.Year = book.Year
	existing.Month = book.Month
	existing.Day = book.Day
	existing.Web = book.Web
	existing.PageCount = book.PageCount
	existing.AgeRating = book.AgeRating
	existing.Characters = book.Characters
	existing.Teams = book.Teams

	// Unified Type Mapping
	existing.SetMangaType(book.Type)

	// Unified Provider Tag Injection
	if tagName := provider.GetTagNameByURL(book.Web); tagName != "" {
		existing.SetCustomProviderTag(tagName, book.Web)
	}

	err := WriteComicInfo(bookPath, existing, backup)
	
	// Update error status in DB
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	models.DB.Model(&models.MangaBook{}).Where("id = ?", book.ID).Update("last_error", errStr)
	
	return err
}

// SyncSeriesMetadata synchronizes series record to folder-level ComicInfo.xml.
func SyncSeriesMetadata(series *models.MangaSeries) error {
	xmlPath := filepath.Join(series.Path, "ComicInfo.xml")
	existing, _ := ReadComicInfo(xmlPath)
	if existing == nil {
		existing = &metadata.ComicInfo{}
	}

	existing.Title = series.Title
	existing.OriginalTitle = series.OriginalTitle
	existing.Series = series.Series
	existing.Writer = series.Author
	existing.Translator = series.Translator
	existing.Publisher = series.Publisher
	existing.Genre = series.Genre
	existing.Tags = series.Tags
	existing.Summary = series.Summary
	existing.Year = series.Year
	existing.Month = series.Month
	existing.Day = series.Day
	existing.Web = series.Web
	
	existing.SetMangaType(series.Type)

	if tagName := provider.GetTagNameByURL(series.Web); tagName != "" {
		existing.SetCustomProviderTag(tagName, series.Web)
	}

	data, err := xml.MarshalIndent(existing, "", "  ")
	if err != nil {
		return err
	}
	data = append([]byte(xml.Header), data...)
	
	err = os.WriteFile(xmlPath, data, 0644)
	
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	models.DB.Model(&models.MangaSeries{}).Where("id = ?", series.ID).Update("last_error", errStr)
	
	return err
}
