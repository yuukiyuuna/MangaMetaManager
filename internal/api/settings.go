package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yuukiyuuna/MangaMetaManager/internal/models"
)

type SettingsHandler struct{}

func NewSettingsHandler() *SettingsHandler {
	return &SettingsHandler{}
}

func (h *SettingsHandler) RegisterRoutes(r *gin.RouterGroup) {
	settings := r.Group("/settings/app")
	{
		settings.GET("", h.GetAppSettings)
		settings.PATCH("", h.UpdateAppSettings)
	}
}

func (h *SettingsHandler) GetAppSettings(c *gin.Context) {
	var settings models.AppSettings
	if err := models.DB.First(&settings).Error; err != nil {
		// If not found, return default
		c.JSON(http.StatusOK, models.AppSettings{BackupBeforeFlatten: true})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *SettingsHandler) UpdateAppSettings(c *gin.Context) {
	var input models.AppSettings
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var settings models.AppSettings
	result := models.DB.First(&settings)
	if result.Error != nil {
		models.DB.Create(&input)
		settings = input
	} else {
		models.DB.Model(&settings).Updates(input)
	}

	c.JSON(http.StatusOK, settings)
}
