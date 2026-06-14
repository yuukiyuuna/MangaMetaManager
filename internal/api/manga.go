package api

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yuukiyuuna/MangaMetaManager/internal/core"
	"github.com/yuukiyuuna/MangaMetaManager/internal/metadata"
	"github.com/yuukiyuuna/MangaMetaManager/internal/models"
	"github.com/yuukiyuuna/MangaMetaManager/internal/provider"
	"github.com/yuukiyuuna/MangaMetaManager/internal/scanner"
	"github.com/yuukiyuuna/MangaMetaManager/internal/utils"
)

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
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "20")
	
	var page, size int
	fmt.Sscanf(pageStr, "%d", &page)
	fmt.Sscanf(sizeStr, "%d", &size)
	
	if page < 1 { page = 1 }
	if size < 1 { size = 20 }
	
	offset := (page - 1) * size

	var series []models.MangaSeries
	models.DB.Preload("Books").Order("title asc").Limit(size).Offset(offset).Find(&series)
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

	// To handle camelCase from JSON to snake_case in DB, we need manual mapping or GORM's map support
	// Map frontend names to DB names
	mapping := map[string]string{
		"title":           "title",
		"originalTitle":   "original_title",
		"series":          "series",
		"alternateSeries": "alternate_series",
		"author":         "author",
		"translator":     "translator",
		"publisher":      "publisher",
		"genre":          "genre",
		"tags":           "tags",
		"summary":        "summary",

		"year":            "year",
		"month":           "month",
		"day":             "day",
		"web":             "web",
		"type":            "type",
		"ageRating":       "age_rating",
	}

	dbUpdates := make(map[string]interface{})
	for jsonKey, dbKey := range mapping {
		if val, ok := input[jsonKey]; ok {
			dbUpdates[dbKey] = val
		}
	}

	if err := models.DB.Model(&series).Updates(dbUpdates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Reload and Sync
	models.DB.First(&series, id)
	go scanner.SyncSeriesMetadata(&series)

	c.JSON(http.StatusOK, series)
}

func (h *MangaHandler) DeleteSeries(c *gin.Context) {
	id := c.Param("id")
	models.DB.Unscoped().Where("series_id = ?", id).Delete(&models.MangaBook{})
	if err := models.DB.Unscoped().Delete(&models.MangaSeries{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "series removed from database"})
}

type ScrapeRequest struct {
	Title           string `json:"title"`
	OriginalTitle   string `json:"originalTitle"`
	Series          string `json:"series"`
	AlternateSeries string `json:"alternateSeries"`
	Author          string `json:"author"`
	Translator      string `json:"translator"`
	Publisher       string `json:"publisher"`
	Genre           string `json:"genre"`
	Tags            string `json:"tags"`
	Summary         string `json:"summary"`
	Year            int    `json:"year"`
	Month           int    `json:"month"`
	Day             int    `json:"day"`
	Web             string `json:"web"`
	PageCount       int    `json:"pageCount"`
	Type            string `json:"type"`
	AgeRating       string `json:"ageRating"`
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
	series.Type = input.Type
	series.AgeRating = input.AgeRating
	series.Status = "Scraped"

	models.DB.Save(&series)
	go scanner.SyncSeriesMetadata(&series)

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

	mapping := map[string]string{
		"title":         "title",
		"originalTitle": "original_title",
		"series":        "series",
		"number":        "number",
		"author":         "author",
		"translator":     "translator",
		"publisher":      "publisher",
		"genre":          "genre",
		"tags":           "tags",
		"summary":        "summary",

		"volume":        "volume",
		"year":          "year",
		"month":         "month",
		"day":           "day",
		"web":           "web",
		"pageCount":     "page_count",
		"manga":         "manga",
		"ageRating":     "age_rating",
		"characters":    "characters",
		"teams":         "teams",
	}
	dbUpdates := make(map[string]interface{})
	if val, ok := input["manga"]; ok {
		dbUpdates["type"] = val
	}
	if val, ok := input["type"]; ok {
		dbUpdates["type"] = val
	}

	for jsonKey, dbKey := range mapping {
		if jsonKey == "manga" { continue } // Handled manually
		if val, ok := input[jsonKey]; ok {
			dbUpdates[dbKey] = val
		}
	}

	if err := models.DB.Model(&book).Updates(dbUpdates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Reload and Sync
	models.DB.First(&book, id)
	go scanner.SyncBookMetadata(&book, h.getBackupSetting())

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
	book.PageCount = input.PageCount
	book.Type = input.Type
	book.AgeRating = input.AgeRating
	book.Status = "Scraped"

	models.DB.Save(&book)
	go scanner.SyncBookMetadata(&book, h.getBackupSetting())

	c.JSON(http.StatusOK, book)
	}

func (h *MangaHandler) AutoScrapeBooks(c *gin.Context) {
	seriesId := c.Param("id")
	var input struct {
		ProviderID string `json:"providerId"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	core.GlobalTaskManager.AddTask(&core.Task{
		ID:   fmt.Sprintf("auto-scrape-%s-%d", seriesId, time.Now().Unix()),
		Type: core.TaskScrape,
		Work: func() {
			p, err := provider.GetProvider(input.ProviderID)
			if err != nil { return }

			var series models.MangaSeries
			if err := models.DB.First(&series, seriesId).Error; err != nil { return }

			var books []models.MangaBook
			if err := models.DB.Where("series_id = ?", seriesId).Find(&books).Error; err != nil { return }

			for _, b := range books {
				// Combined context for more accurate book search
				cleanedTitle := utils.BuildBookSearchQuery(series.Title, b.Filename)
				results, err := p.Search(cleanedTitle)
				if err == nil && len(results) > 0 {
					details, err := p.GetDetails(results[0].ID)
					if err == nil {
						b.Title = details.Title
						b.OriginalTitle = details.OriginalTitle
						b.Author = details.Writer
						b.Translator = details.Translator
						b.Publisher = details.Publisher
						b.Genre = details.Genre
						b.Tags = details.Tags
						b.Summary = details.Summary
						b.Year = details.Year
						b.Month = details.Month
						b.Day = details.Day
						b.Web = details.Web
						b.PageCount = details.PageCount
						
						// Default to 漫画 for auto scrape
						if details.Manga == "No" {
							b.Type = "小说"
						} else {
							b.Type = "漫画"
						}
						
						b.AgeRating = details.AgeRating
						b.Status = "Scraped"

						models.DB.Save(&b)
						scanner.SyncBookMetadata(&b, h.getBackupSetting())

						// Small delay to prevent rate limit
						time.Sleep(500 * time.Millisecond)
					}
				}
			}
		},
	})

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
		data, _ = xml.MarshalIndent(info, "", "  ")
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
	models.DB.Save(&series)
	c.JSON(http.StatusOK, gin.H{"status": "saved"})
}

func (h *MangaHandler) GetBookXML(c *gin.Context) {
	id := c.Param("bookId")
	var book models.MangaBook
	if err := models.DB.First(&book, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}
	existing, _ := scanner.ReadComicInfo(book.Path)
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

	data, _ := xml.MarshalIndent(existing, "", "  ")
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
	
	book.Title = info.Title
	book.OriginalTitle = info.OriginalTitle
	book.Series = info.Series
	book.Number = info.Number
	book.Author = info.Writer
	book.Publisher = info.Publisher
	book.Genre = info.Genre
	book.Volume = info.Volume
	book.Year = info.Year
	book.Month = info.Month
	book.Day = info.Day
	book.Web = info.Web
	book.PageCount = info.PageCount
	if info.Manga == "No" {
		book.Type = "小说"
	} else {
		book.Type = "漫画"
	}
	book.AgeRating = info.AgeRating
	book.Characters = info.Characters
	book.Teams = info.Teams
	book.Summary = info.Summary
	models.DB.Save(&book)
	c.JSON(http.StatusOK, gin.H{"status": "saved"})
}

func (h *MangaHandler) ListLibraryFolders(c *gin.Context) {
	var folders []models.LibraryFolder
	models.DB.Find(&folders)
	c.JSON(http.StatusOK, folders)
}

func (h *MangaHandler) AddLibraryFolder(c *gin.Context) {
	var input models.LibraryFolder
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := models.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, input)
}

func (h *MangaHandler) RemoveLibraryFolder(c *gin.Context) {
	id := c.Param("id")
	if err := models.DB.Delete(&models.LibraryFolder{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *MangaHandler) CleanLibrary(c *gin.Context) {
	core.GlobalTaskManager.AddTask(&core.Task{
		ID:   fmt.Sprintf("clean-%d", time.Now().Unix()),
		Type: "Database Clean",
		Work: func() {
			scanner.CleanLibrary()
		},
	})
	c.JSON(http.StatusOK, gin.H{"status": "clean task added to queue"})
}

func (h *MangaHandler) ScanLibrary(c *gin.Context) {
	core.GlobalTaskManager.AddTask(&core.Task{
		ID:   fmt.Sprintf("scan-%d", time.Now().Unix()),
		Type: core.TaskScan,
		Work: func() {
			scanner.ScanLibrary()
		},
	})
	c.JSON(http.StatusOK, gin.H{"status": "scan task added to queue"})
}

func (h *MangaHandler) GetTasks(c *gin.Context) {
	c.JSON(http.StatusOK, core.GlobalTaskManager.GetTasks())
}
