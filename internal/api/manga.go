package api

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yuukiyuuna/MangaMetaManager/internal/core"
	"github.com/yuukiyuuna/MangaMetaManager/internal/metadata"
	"github.com/yuukiyuuna/MangaMetaManager/internal/models"
	"github.com/yuukiyuuna/MangaMetaManager/internal/provider"
	"github.com/yuukiyuuna/MangaMetaManager/internal/scanner"
	"github.com/yuukiyuuna/MangaMetaManager/internal/utils"
	"gorm.io/gorm"
)

const maxPageSize = 100

var seriesUpdateFields = map[string]string{
	"title":           "title",
	"originalTitle":   "original_title",
	"series":          "series",
	"alternateSeries": "alternate_series",
	"author":          "author",
	"translator":      "translator",
	"publisher":       "publisher",
	"genre":           "genre",
	"tags":            "tags",
	"summary":         "summary",
	"year":            "year",
	"month":           "month",
	"day":             "day",
	"web":             "web",
	"language":        "language",
	"type":            "type",
	"ageRating":       "age_rating",
	"gtin":            "gtin",
}

var bookUpdateFields = map[string]string{
	"title":           "title",
	"originalTitle":   "original_title",
	"series":          "series",
	"author":          "author",
	"translator":      "translator",
	"publisher":       "publisher",
	"genre":           "genre",
	"tags":            "tags",
	"summary":         "summary",
	"year":            "year",
	"month":           "month",
	"day":             "day",
	"web":             "web",
	"language":        "language",
	"pageCount":       "page_count",
	"type":            "type",
	"manga":           "type",
	"ageRating":       "age_rating",
	"gtin":            "gtin",
	"volume":          "volume",
	"number":          "number",
	"characters":      "characters",
	"teams":           "teams",
	"seriesGroup":     "series_group",
	"alternateSeries": "alternate_series",
	"alternateNumber": "alternate_number",
	"storyArc":        "story_arc",
}

func parsePositiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}

func filteredUpdates(input map[string]interface{}, allowed map[string]string) map[string]interface{} {
	updates := make(map[string]interface{})
	for jsonKey, dbKey := range allowed {
		if val, ok := input[jsonKey]; ok {
			updates[dbKey] = val
		}
	}
	return updates
}

type MangaHandler struct{}

func NewMangaHandler() *MangaHandler {
	return &MangaHandler{}
}

func (h *MangaHandler) RegisterRoutes(r *gin.RouterGroup) {
	manga := r.Group("/manga")
	{
		manga.GET("", h.ListSeries)
		manga.GET("/:id", h.GetSeries)
		manga.PATCH("/:id", h.UpdateSeries)
		manga.DELETE("/:id", h.DeleteSeries)
		manga.POST("/:id/scrape", h.ScrapeSeries)
		manga.POST("/:id/auto-scrape-books", h.AutoScrapeBooks)

		manga.GET("/books/:bookId", h.GetBook)
		manga.PATCH("/books/:bookId", h.UpdateBook)
		manga.POST("/books/:bookId/scrape", h.ScrapeBook)

		// RAW XML Routes
		manga.GET("/:id/xml", h.GetSeriesXML)
		manga.PUT("/:id/xml", h.UpdateSeriesXML)
		manga.GET("/books/:bookId/xml", h.GetBookXML)
		manga.PUT("/books/:bookId/xml", h.UpdateBookXML)
	}

	library := r.Group("/library")
	{
		library.GET("/folders", h.ListLibraryFolders)
		library.POST("/folders", h.AddLibraryFolder)
		library.DELETE("/folders/:id", h.RemoveLibraryFolder)
		library.POST("/scan", h.ScanLibrary)
		library.POST("/clean", h.CleanLibrary)
		library.GET("/tasks", h.GetTasks)
	}
}

func (h *MangaHandler) getBackupSetting() bool {
	var settings models.AppSettings
	if err := models.DB.First(&settings).Error; err != nil {
		return true // Default
	}
	return settings.BackupBeforeFlatten
}

func (h *MangaHandler) ListSeries(c *gin.Context) {
	page := parsePositiveInt(c.DefaultQuery("page", "1"), 1)
	size := parsePositiveInt(c.DefaultQuery("size", "20"), 20)
	if size > maxPageSize {
		size = maxPageSize
	}

	offset := (page - 1) * size

	var series []models.MangaSeries
	if err := models.DB.Preload("Books").Order("title asc").Limit(size).Offset(offset).Find(&series).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, series)
}

func (h *MangaHandler) GetSeries(c *gin.Context) {
	id := c.Param("id")
	var series models.MangaSeries
	if err := models.DB.Preload("Books").First(&series, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Series not found"})
		return
	}
	c.JSON(http.StatusOK, series)
}

func (h *MangaHandler) UpdateSeries(c *gin.Context) {
	id := c.Param("id")
	var series models.MangaSeries
	if err := models.DB.First(&series, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Series not found"})
		return
	}

	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("[API] Manual Update Series %s: %v\n", id, input)

	dbUpdates := filteredUpdates(input, seriesUpdateFields)
	if len(dbUpdates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No supported fields to update"})
		return
	}

	if err := models.DB.Model(&series).Updates(dbUpdates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Reload and Sync
	if err := models.DB.First(&series, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, series)
}

func (h *MangaHandler) DeleteSeries(c *gin.Context) {
	id := c.Param("id")
	if err := models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("series_id = ?", id).Delete(&models.MangaBook{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Delete(&models.MangaSeries{}, id).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "series removed from database"})
}

type ScrapeRequest struct {
	Title           string  `json:"title"`
	OriginalTitle   string  `json:"originalTitle"`
	Series          string  `json:"series"`
	AlternateSeries string  `json:"alternateSeries"`
	Author          string  `json:"author"`
	Translator      string  `json:"translator"`
	Publisher       string  `json:"publisher"`
	Genre           string  `json:"genre"`
	Tags            string  `json:"tags"`
	Summary         string  `json:"summary"`
	Year            int     `json:"year"`
	Month           int     `json:"month"`
	Day             int     `json:"day"`
	Web             string  `json:"web"`
	Language        string  `json:"language"`
	PageCount       int     `json:"pageCount"`
	Type            string  `json:"type"`
	AgeRating       string  `json:"ageRating"`
	GTIN            string  `json:"gtin"`
	Volume          float64 `json:"volume"`
}

type AutoScrapeOptions struct {
	UpdateTitle      bool `json:"updateTitle"`
	UpdateAuthor     bool `json:"updateAuthor"`
	UpdateTranslator bool `json:"updateTranslator"`
	UpdateSummary    bool `json:"updateSummary"`
	UpdatePublisher  bool `json:"updatePublisher"`
	UpdateGenre      bool `json:"updateGenre"`
	UpdateDate       bool `json:"updateDate"`
	UpdateWeb        bool `json:"updateWeb"`
	UpdateLanguage   bool `json:"updateLanguage"`
	UpdatePageCount  bool `json:"updatePageCount"`
	UpdateType       bool `json:"updateType"`
	UpdateAgeRating  bool `json:"updateAgeRating"`
	UpdateGTIN       bool `json:"updateGtin"`
}

func defaultAutoScrapeOptions() AutoScrapeOptions {
	return AutoScrapeOptions{
		UpdateTitle:      true,
		UpdateAuthor:     true,
		UpdateTranslator: true,
		UpdateSummary:    true,
		UpdatePublisher:  true,
		UpdateGenre:      true,
		UpdateDate:       true,
		UpdateWeb:        true,
		UpdateLanguage:   true,
		UpdatePageCount:  true,
		UpdateType:       true,
		UpdateAgeRating:  true,
		UpdateGTIN:       true,
	}
}

func (h *MangaHandler) ScrapeSeries(c *gin.Context) {
	id := c.Param("id")
	var input ScrapeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var series models.MangaSeries
	if err := models.DB.First(&series, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Series not found"})
		return
	}

	// Manual assignment to handle all fields
	series.Title = input.Title
	series.OriginalTitle = input.OriginalTitle
	series.Series = input.Series
	series.AlternateSeries = input.AlternateSeries
	series.Author = input.Author
	series.Translator = input.Translator
	series.Publisher = input.Publisher
	series.Summary = input.Summary
	series.Genre = input.Genre
	series.Tags = input.Tags
	series.Year = input.Year
	series.Month = input.Month
	series.Day = input.Day
	series.Web = input.Web
	series.Language = input.Language
	series.Type = input.Type
	series.AgeRating = input.AgeRating
	series.GTIN = input.GTIN
	series.Status = "Scraped"

	if err := models.DB.Save(&series).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, series)
}

func (h *MangaHandler) GetBook(c *gin.Context) {
	id := c.Param("bookId")
	var book models.MangaBook
	if err := models.DB.First(&book, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}
	c.JSON(http.StatusOK, book)
}

func (h *MangaHandler) UpdateBook(c *gin.Context) {
	id := c.Param("bookId")
	var book models.MangaBook
	if err := models.DB.First(&book, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("[API] Manual Update Book %s: %v\n", id, input)

	dbUpdates := filteredUpdates(input, bookUpdateFields)
	if len(dbUpdates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No supported fields to update"})
		return
	}

	if err := models.DB.Model(&book).Updates(dbUpdates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Reload and Sync
	if err := models.DB.First(&book, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	core.SyncQueue <- func() {
		scanner.SyncBookMetadata(&book, h.getBackupSetting())
	}

	c.JSON(http.StatusOK, book)
}

func (h *MangaHandler) ScrapeBook(c *gin.Context) {
	id := c.Param("bookId")
	var input ScrapeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var book models.MangaBook
	if err := models.DB.First(&book, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	book.Title = input.Title
	book.OriginalTitle = input.OriginalTitle
	book.Series = input.Series
	book.Author = input.Author
	book.Translator = input.Translator
	book.Publisher = input.Publisher
	book.Genre = input.Genre
	book.Tags = input.Tags
	book.Summary = input.Summary
	book.Year = input.Year
	book.Month = input.Month
	book.Day = input.Day
	book.Web = input.Web
	book.Language = input.Language
	book.PageCount = input.PageCount
	book.Type = input.Type
	book.AgeRating = input.AgeRating
	book.GTIN = input.GTIN
	book.Volume = input.Volume
	book.Status = "Scraped"

	if err := models.DB.Save(&book).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	core.SyncQueue <- func() {
		scanner.SyncBookMetadata(&book, h.getBackupSetting())
	}

	c.JSON(http.StatusOK, book)
}

func (h *MangaHandler) AutoScrapeBooks(c *gin.Context) {
	seriesId := c.Param("id")
	var input struct {
		ProviderID string             `json:"providerId"`
		Options    *AutoScrapeOptions `json:"options"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var series models.MangaSeries
	if err := models.DB.First(&series, seriesId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Series not found"})
		return
	}

	// Prerequisite: Series must have been manually scraped first (indicated by Status and Web link)
	if series.Status != "Scraped" || series.Web == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please scrape series metadata manually first to establish a provider link."})
		return
	}

	p, err := provider.GetProvider(input.ProviderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider"})
		return
	}
	if !provider.IsImplemented(input.ProviderID) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Provider is not implemented yet"})
		return
	}

	options := defaultAutoScrapeOptions()
	if input.Options != nil {
		options = *input.Options
	}

	task := &core.Task{
		ID:   fmt.Sprintf("auto-scrape-%s-%d", seriesId, time.Now().Unix()),
		Type: core.TaskScrape,
	}

	task.Work = func() error {
		// Use Provider interface to extract ID
		subjectID := p.ExtractIDFromURL(series.Web)
		if subjectID == "" {
			return fmt.Errorf("could not extract provider ID from %s", series.Web)
		}

		// Fetch Ground Truth related books with retries
		var relatedBooks []provider.SearchResult
		var err error
		for attempt := 1; attempt <= 3; attempt++ {
			relatedBooks, err = p.GetRelatedBooks(subjectID)
			if err == nil {
				break
			}
			fmt.Printf("[AutoScrape] Attempt %d failed for related books of %s: %v\n", attempt, series.Title, err)
			if attempt < 3 {
				time.Sleep(2000 * time.Millisecond)
			}
		}

		if err != nil {
			return fmt.Errorf("fetch related books for %s after 3 attempts: %w", series.Title, err)
		}

		// Delay to respect API limits before starting individual book scrapes
		time.Sleep(1000 * time.Millisecond)

		var books []models.MangaBook
		if err := models.DB.Where("series_id = ?", seriesId).Find(&books).Error; err != nil {
			return err
		}

		total := len(books)
		core.GlobalTaskManager.UpdateProgress(task, 0, total, "Starting...")

		for i, b := range books {
			core.GlobalTaskManager.UpdateProgress(task, i+1, total, fmt.Sprintf("Processing %s", b.Filename))

			localVol := utils.ParseVolumeNumber(b.Filename)
			var matchedID string

			// 1. Try numeric match with ambiguity handling
			if localVol != -1 {
				var matches []provider.SearchResult
				for _, rb := range relatedBooks {
					remoteVol := utils.ParseVolumeNumber(rb.Title)
					if localVol == remoteVol {
						matches = append(matches, rb)
					}
				}

				if len(matches) == 1 {
					matchedID = matches[0].ID
				} else if len(matches) > 1 {
					// Handle ambiguity: look for keywords like "特装版", "限定版"
					keywords := []string{"特装版", "限定版", "特装", "限定", "Special Edition", "Limited Edition", "Drama CD"}

					localHasKeyword := false
					for _, k := range keywords {
						if strings.Contains(strings.ToLower(b.Filename), strings.ToLower(k)) {
							localHasKeyword = true
							break
						}
					}

					if !localHasKeyword {
						// Prefer the one WITHOUT keywords
						for _, m := range matches {
							hasRemoteKeyword := false
							for _, k := range keywords {
								if strings.Contains(strings.ToLower(m.Title), strings.ToLower(k)) {
									hasRemoteKeyword = true
									break
								}
							}
							if !hasRemoteKeyword {
								matchedID = m.ID
								break
							}
						}
					}

					// Fallback to first if still no match
					if matchedID == "" {
						matchedID = matches[0].ID
					}
				}
			}

			// 2. Fallback to similarity match for irregular names
			if matchedID == "" {
				maxSim := 0.0
				for _, rb := range relatedBooks {
					sim := utils.SimpleSimilarity(b.Filename, rb.Title)
					if sim > 0.8 && sim > maxSim {
						maxSim = sim
						matchedID = rb.ID
					}
				}
			}

			if matchedID != "" {
				var details *metadata.ComicInfo
				var err error
				for attempt := 1; attempt <= 3; attempt++ {
					details, err = p.GetDetails(matchedID)
					if err == nil {
						break
					}
					fmt.Printf("[AutoScrape] Attempt %d failed for book %s: %v\n", attempt, b.Filename, err)
					if attempt < 3 {
						time.Sleep(2000 * time.Millisecond)
					}
				}

				if err == nil {
					if options.UpdateTitle {
						b.Title = details.Title
						b.OriginalTitle = details.OriginalTitle
					}
					if options.UpdateAuthor {
						b.Author = details.Writer
					}
					if options.UpdateTranslator {
						b.Translator = details.Translator
					}
					if options.UpdatePublisher {
						b.Publisher = details.Publisher
					}
					if options.UpdateGenre {
						b.Genre = details.Genre
						b.Tags = details.Tags
					}
					if options.UpdateSummary {
						b.Summary = details.Summary
					}
					if options.UpdateDate {
						b.Year = details.Year
						b.Month = details.Month
						b.Day = details.Day
					}
					if options.UpdateWeb {
						b.Web = details.Web
					}
					if options.UpdateLanguage {
						b.Language = details.LanguageISO
					}
					if options.UpdatePageCount {
						b.PageCount = details.PageCount
					}

					if options.UpdateType {
						if details.Manga == "No" {
							b.Type = "小说"
						} else {
							b.Type = "漫画"
						}
					}

					if options.UpdateAgeRating {
						b.AgeRating = details.AgeRating
					}
					if options.UpdateGTIN {
						b.GTIN = details.GTIN
					}
					b.Status = "Scraped"
					b.LastError = ""

					if err := models.DB.Save(&b).Error; err != nil {
						return err
					}
					if err := scanner.SyncBookMetadata(&b, h.getBackupSetting()); err != nil {
						return err
					}
				} else {
					b.LastError = fmt.Sprintf("Failed to fetch details after 3 attempts: %v", err)
					if err := models.DB.Save(&b).Error; err != nil {
						return err
					}
				}
				// Moderate delay for API limits after network attempts
				time.Sleep(1000 * time.Millisecond)
			} else {
				fmt.Printf("[AutoScrape] No match found for: %s\n", b.Filename)
				b.LastError = "No match found during auto-scrape"
				if err := models.DB.Save(&b).Error; err != nil {
					return err
				}
			}
		}
		return nil
	}

	core.GlobalTaskManager.AddTask(task)

	c.JSON(http.StatusOK, gin.H{"status": "auto scrape task added to queue"})
}

func (h *MangaHandler) GetSeriesXML(c *gin.Context) {
	id := c.Param("id")
	var series models.MangaSeries
	if err := models.DB.First(&series, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Series not found"})
		return
	}

	xmlPath := filepath.Join(series.Path, "ComicInfo.xml")
	data, err := os.ReadFile(xmlPath)
	if err != nil {
		info := &metadata.ComicInfo{XmlnsXsi: "http://www.w3.org/2001/XMLSchema-instance", XmlnsXsd: "http://www.w3.org/2001/XMLSchema"}
		data, err = xml.MarshalIndent(info, "", "  ")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		data = append([]byte(xml.Header), data...)
	}
	c.Data(http.StatusOK, "application/xml", data)
}

func (h *MangaHandler) UpdateSeriesXML(c *gin.Context) {
	id := c.Param("id")
	var series models.MangaSeries
	if err := models.DB.First(&series, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Series not found"})
		return
	}
	rawData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}
	var info metadata.ComicInfo
	if err := xml.Unmarshal(rawData, &info); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid XML format: " + err.Error()})
		return
	}
	xmlPath := filepath.Join(series.Path, "ComicInfo.xml")
	if err := os.WriteFile(xmlPath, rawData, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Map fields manually for sync
	series.Title = info.Title
	series.OriginalTitle = info.OriginalTitle
	series.Series = info.Series
	series.AlternateSeries = info.AlternateSeries
	series.Author = info.Writer
	series.Publisher = info.Publisher
	series.Summary = info.Summary
	series.Genre = info.Genre
	series.Year = info.Year
	series.Month = info.Month
	series.Day = info.Day
	series.Web = info.Web
	if info.Manga == "No" {
		series.Type = "小说"
	} else {
		series.Type = "漫画"
	}
	series.AgeRating = info.AgeRating
	if err := models.DB.Save(&series).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "saved"})
}

func (h *MangaHandler) GetBookXML(c *gin.Context) {
	id := c.Param("bookId")
	var book models.MangaBook
	if err := models.DB.First(&book, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}
	existing, err := scanner.ReadComicInfo(book.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		existing = &metadata.ComicInfo{}
	}
	// Always ensure namespaces for Komga compatibility
	existing.XmlnsXsi = "http://www.w3.org/2001/XMLSchema-instance"
	existing.XmlnsXsd = "http://www.w3.org/2001/XMLSchema"

	// Ensure Manga type is set if empty (for new files or files without the field)
	if existing.Manga == "" {
		existing.SetMangaType(book.Type)
	}

	data, err := xml.MarshalIndent(existing, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	data = append([]byte(xml.Header), data...)
	c.Data(http.StatusOK, "application/xml", data)
}

func (h *MangaHandler) UpdateBookXML(c *gin.Context) {
	id := c.Param("bookId")
	var book models.MangaBook
	if err := models.DB.First(&book, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}
	rawData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}
	var info metadata.ComicInfo
	if err := xml.Unmarshal(rawData, &info); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid XML format: " + err.Error()})
		return
	}
	if err := scanner.WriteRawComicInfo(book.Path, rawData, h.getBackupSetting()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write ZIP: " + err.Error()})
		return
	}

	if stat, err := os.Stat(book.Path); err == nil {
		book.FileModTime = stat.ModTime().Unix()
		book.FileSize = stat.Size()
	}

	book.Title = info.Title
	book.OriginalTitle = info.OriginalTitle
	book.Series = info.Series
	book.Number = info.Number
	book.Author = info.Writer
	book.Publisher = info.Publisher
	book.Genre = info.Genre
	book.Volume = float64(info.Volume)
	book.Year = info.Year
	book.Month = info.Month
	book.Day = info.Day
	book.Web = info.Web
	book.Language = info.LanguageISO
	book.PageCount = info.PageCount
	if info.Manga == "No" {
		book.Type = "小说"
	} else {
		book.Type = "漫画"
	}
	book.AgeRating = info.AgeRating
	book.GTIN = info.GTIN
	book.Characters = info.Characters
	book.Teams = info.Teams
	book.Summary = info.Summary
	if err := models.DB.Save(&book).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "saved"})
}

func (h *MangaHandler) ListLibraryFolders(c *gin.Context) {
	var folders []models.LibraryFolder
	if err := models.DB.Find(&folders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, folders)
}

func (h *MangaHandler) AddLibraryFolder(c *gin.Context) {
	var input models.LibraryFolder
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate path
	absPath, err := filepath.Abs(input.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid path format"})
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Directory does not exist"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	if !info.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is not a directory"})
		return
	}

	input.Path = absPath
	if err := models.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, input)
}

func (h *MangaHandler) RemoveLibraryFolder(c *gin.Context) {
	id := c.Param("id")
	var folder models.LibraryFolder
	if err := models.DB.First(&folder, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	// Delete all series and books associated with this folder path
	// We use a transaction for safety
	err := models.DB.Transaction(func(tx *gorm.DB) error {
		// Find series in this folder. Match exactly or as a subfolder.
		pattern := folder.Path + "/%"

		var seriesIDs []uint
		tx.Model(&models.MangaSeries{}).Where("path = ? OR path LIKE ?", folder.Path, pattern).Pluck("id", &seriesIDs)

		if len(seriesIDs) > 0 {
			// Delete books belonging to these series
			if err := tx.Where("series_id IN ?", seriesIDs).Delete(&models.MangaBook{}).Error; err != nil {
				return err
			}
			// Delete the series themselves
			if err := tx.Where("id IN ?", seriesIDs).Delete(&models.MangaSeries{}).Error; err != nil {
				return err
			}
		}

		// 2. Delete the LibraryFolder record
		if err := tx.Delete(&models.LibraryFolder{}, id).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted and cleaned up"})
}

func (h *MangaHandler) CleanLibrary(c *gin.Context) {
	task := &core.Task{
		ID:   fmt.Sprintf("clean-%d", time.Now().Unix()),
		Type: "Database Clean",
	}
	task.Work = func() error {
		return scanner.CleanLibrary(task)
	}
	core.GlobalTaskManager.AddTask(task)
	c.JSON(http.StatusOK, gin.H{"status": "clean task added to queue"})
}

func (h *MangaHandler) ScanLibrary(c *gin.Context) {
	task := &core.Task{
		ID:   fmt.Sprintf("scan-%d", time.Now().Unix()),
		Type: core.TaskScan,
	}
	task.Work = func() error {
		return scanner.ScanLibrary(task)
	}
	core.GlobalTaskManager.AddTask(task)
	c.JSON(http.StatusOK, gin.H{"status": "scan task added to queue"})
}

func (h *MangaHandler) GetTasks(c *gin.Context) {
	c.JSON(http.StatusOK, core.GlobalTaskManager.GetTasks())
}
