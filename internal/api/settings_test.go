package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yuukiyuuna/MangaMetaManager/internal/models"
)

func setupTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	models.InitDB(filepath.Join(t.TempDir(), "test.db"))

	router := gin.New()
	apiGroup := router.Group("/api")
	NewSettingsHandler().RegisterRoutes(apiGroup)
	NewMangaHandler().RegisterRoutes(apiGroup)

	return router
}

func performJSONRequest(t *testing.T, router http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, bytes.NewReader(data))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	return recorder
}

func TestUpdateAppSettingsPersistsFalseValue(t *testing.T) {
	router := setupTestRouter(t)

	recorder := performJSONRequest(t, router, http.MethodPatch, "/api/settings/app", map[string]bool{"backupBeforeFlatten": true})
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	recorder = performJSONRequest(t, router, http.MethodPatch, "/api/settings/app", map[string]bool{"backupBeforeFlatten": false})
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var settings models.AppSettings
	if err := models.DB.First(&settings).Error; err != nil {
		t.Fatalf("failed to load settings: %v", err)
	}
	if settings.BackupBeforeFlatten {
		t.Fatal("expected backupBeforeFlatten to persist as false")
	}

	mangaHandler := NewMangaHandler()
	if mangaHandler.getBackupSetting() {
		t.Fatal("expected scraper backup setting to read false")
	}

	recorder = performJSONRequest(t, router, http.MethodPatch, "/api/settings/app", map[string]bool{"backupBeforeFlatten": true})
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestAddLibraryFolderIsIdempotent(t *testing.T) {
	router := setupTestRouter(t)
	libraryPath := t.TempDir()

	recorder := performJSONRequest(t, router, http.MethodPost, "/api/library/folders", map[string]string{"path": libraryPath})
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected first add status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	recorder = performJSONRequest(t, router, http.MethodPost, "/api/library/folders", map[string]string{"path": libraryPath})
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected duplicate add status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var count int64
	if err := models.DB.Model(&models.LibraryFolder{}).Where("path = ?", libraryPath).Count(&count).Error; err != nil {
		t.Fatalf("failed to count folders: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one folder record, got %d", count)
	}

	var folder models.LibraryFolder
	if err := models.DB.Where("path = ?", libraryPath).First(&folder).Error; err != nil {
		t.Fatalf("failed to load folder: %v", err)
	}

	recorder = httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodDelete, "/api/library/folders/"+strconv.FormatUint(uint64(folder.ID), 10), nil)
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected delete status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	recorder = performJSONRequest(t, router, http.MethodPost, "/api/library/folders", map[string]string{"path": libraryPath})
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected re-add status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
}
