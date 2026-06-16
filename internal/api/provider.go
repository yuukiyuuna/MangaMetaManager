package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yuukiyuuna/MangaMetaManager/internal/provider"
	"github.com/yuukiyuuna/MangaMetaManager/internal/utils"
)

type ProviderHandler struct{}

func NewProviderHandler() *ProviderHandler {
	return &ProviderHandler{}
}

func (h *ProviderHandler) RegisterRoutes(r *gin.RouterGroup) {
	pGroup := r.Group("/metadata/providers")
	{
		pGroup.GET("", h.ListProviders)
		pGroup.GET("/:id/search", h.Search)
		pGroup.GET("/:id/details/:metadataID", h.GetDetails)
	}
}

func (h *ProviderHandler) ListProviders(c *gin.Context) {
	providers := provider.ListProviders()
	type ProviderInfo struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	res := make([]ProviderInfo, 0, len(providers))
	for _, p := range providers {
		res = append(res, ProviderInfo{ID: p.ID(), Name: p.Name()})
	}
	c.JSON(http.StatusOK, res)
}

func (h *ProviderHandler) Search(c *gin.Context) {
	id := c.Param("id")
	query := c.Query("q")

	// Pre-clean query
	query = utils.BuildBookSearchQuery("", query)

	p, err := provider.GetProvider(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	results, err := p.Search(query)
	if err != nil {
		if errors.Is(err, provider.ErrNotImplemented) {
			c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

func (h *ProviderHandler) GetDetails(c *gin.Context) {
	id := c.Param("id")
	metadataID := c.Param("metadataID")

	p, err := provider.GetProvider(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	details, err := p.GetDetails(metadataID)
	if err != nil {
		if errors.Is(err, provider.ErrNotImplemented) {
			c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, details)
}
