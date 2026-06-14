package scanner

import (
	"strings"

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

	// Fetch parent series for context
	var series models.MangaSeries
	if err := models.DB.First(&series, book.SeriesID).Error; err == nil {
		// Use Series' OriginalTitle for the XML Series tag as requested
		if series.OriginalTitle != "" {
			existing.Series = series.OriginalTitle
		} else {
			existing.Series = series.Title
		}
	} else {
		existing.Series = book.Series
	}

	// Map DB fields to ComicInfo
	existing.Title = book.Title
	existing.OriginalTitle = book.OriginalTitle // This is the book's original title
	existing.Number = book.Number
	existing.Writer = book.Author
	existing.Translator = book.Translator
	existing.Publisher = book.Publisher

	// Filtering: "漫画" is not a tag or genre.
	existing.Genre = filterMetadataString(book.Genre)
	existing.Tags = filterMetadataString(book.Tags)

	existing.Summary = book.Summary
	existing.Volume = book.Volume
	existing.Year = book.Year
	existing.Month = book.Month
	existing.Day = book.Day
	existing.Web = book.Web
	existing.PageCount = book.PageCount
	existing.LanguageISO = book.Language
	existing.AgeRating = book.AgeRating
	existing.Characters = book.Characters
	existing.Teams = book.Teams
	existing.SeriesGroup = book.SeriesGroup
	existing.AlternateSeries = book.AlternateSeries
	existing.AlternateNumber = book.AlternateNumber
	existing.StoryArc = book.StoryArc

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

func filterMetadataString(input string) string {
	parts := strings.Split(input, ",")
	var cleaned []string
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t == "" || t == "漫画" || t == "小说" {
			continue
		}
		cleaned = append(cleaned, t)
	}
	return strings.Join(cleaned, ", ")
}

// SyncSeriesMetadata is now a no-op for file synchronization as requested. 
// Series metadata is managed only in the database to keep the library clean.
func SyncSeriesMetadata(series *models.MangaSeries) error {
	// We no longer write ComicInfo.xml to the series folder.
	return nil
}
