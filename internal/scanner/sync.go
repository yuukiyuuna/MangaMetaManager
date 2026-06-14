package scanner

import (
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

// SyncSeriesMetadata is now a no-op for file synchronization as requested. 
// Series metadata is managed only in the database to keep the library clean.
func SyncSeriesMetadata(series *models.MangaSeries) error {
	// We no longer write ComicInfo.xml to the series folder.
	return nil
}
