package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yuukiyuuna/MangaMetaManager/internal/models"
	"github.com/yuukiyuuna/MangaMetaManager/internal/network"
)

type ProxyHandler struct {
	factory *network.HTTPClientFactory
}

func NewProxyHandler() *ProxyHandler {
	return &ProxyHandler{
		factory: network.NewHTTPClientFactory(),
	}
}

func (h *ProxyHandler) RegisterRoutes(r *gin.RouterGroup) {
	proxy := r.Group("/settings/proxy")
	{
		proxy.GET("", h.GetGlobalProxy)
		proxy.PATCH("", h.UpdateGlobalProxy)
		proxy.POST("/test", h.TestProxy)
	}

	provider := r.Group("/providers")
	{
		provider.GET("/:id/proxy", h.GetProviderProxy)
		provider.PATCH("/:id/proxy", h.UpdateProviderProxy)
	}
}

func (h *ProxyHandler) GetGlobalProxy(c *gin.Context) {
	var settings models.ProxySettings
	if err := models.DB.First(&settings).Error; err != nil {
		c.JSON(http.StatusOK, models.ProxySettings{Enabled: false})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *ProxyHandler) UpdateGlobalProxy(c *gin.Context) {
	var input models.ProxySettings
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validation
	if input.Enabled {
		if input.Host == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Host cannot be empty when proxy is enabled"})
			return
		}
		if input.Port <= 0 || input.Port > 65535 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid port number"})
			return
		}
	}

	var settings models.ProxySettings
	result := models.DB.First(&settings)
	if result.Error != nil {
		models.DB.Create(&input)
		settings = input
	} else {
		// Preserve Password if not provided in input
		if input.Password == "" {
			input.Password = settings.Password
		}
		models.DB.Model(&settings).Updates(input)
	}

	h.factory.InvalidateCache() // Invalidate cache after update

	settings.Password = "" // Hide password
	c.JSON(http.StatusOK, settings)
}

func (h *ProxyHandler) TestProxy(c *gin.Context) {
	var input struct {
		TestURL string `json:"testUrl"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.TestURL == "" {
		input.TestURL = "https://www.google.com"
	}

	client, err := h.factory.GetClient("")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp, err := client.Get(input.TestURL)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		c.JSON(http.StatusOK, gin.H{"success": true, "statusCode": resp.StatusCode})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "statusCode": resp.StatusCode})
	}
}

func (h *ProxyHandler) GetProviderProxy(c *gin.Context) {
	id := c.Param("id")
	var strategy models.ProviderProxyStrategy
	if err := models.DB.Where("provider_id = ?", id).First(&strategy).Error; err != nil {
		c.JSON(http.StatusOK, models.ProviderProxyStrategy{ProviderID: id, Strategy: "inherit"})
		return
	}
	strategy.Password = "" // Hide password
	c.JSON(http.StatusOK, strategy)
}

func (h *ProxyHandler) UpdateProviderProxy(c *gin.Context) {
	id := c.Param("id")
	var input models.ProviderProxyStrategy
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ProviderID = id

	var strategy models.ProviderProxyStrategy
	result := models.DB.Where("provider_id = ?", id).First(&strategy)
	if result.Error != nil {
		models.DB.Create(&input)
		strategy = input
	} else {
		if input.Password == "" {
			input.Password = strategy.Password
		}
		models.DB.Model(&strategy).Updates(input)
	}

	h.factory.InvalidateCache() // Invalidate cache after update

	strategy.Password = "" // Hide password
	c.JSON(http.StatusOK, strategy)
}
