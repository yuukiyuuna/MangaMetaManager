package provider

import (
	"strings"

	"github.com/yuukiyuuna/MangaMetaManager/internal/metadata"
	"github.com/yuukiyuuna/MangaMetaManager/internal/network"
)

type AmazonProvider struct {
	factory *network.HTTPClientFactory
}

func NewAmazonProvider() *AmazonProvider {
	return &AmazonProvider{
		factory: network.NewHTTPClientFactory(),
	}
}

func (p *AmazonProvider) ID() string   { return "amazon" }
func (p *AmazonProvider) Name() string { return "Amazon" }
func (p *AmazonProvider) GetCustomTagName() string { return "CustomAmazon" }

func (p *AmazonProvider) Search(query string) ([]SearchResult, error) {
	return nil, ErrNotImplemented
}

func (p *AmazonProvider) GetDetails(id string) (*metadata.ComicInfo, error) {
	return nil, ErrNotImplemented
}

func (p *AmazonProvider) GetRelatedBooks(id string) ([]SearchResult, error) {
	return nil, ErrNotImplemented
}

func (p *AmazonProvider) ExtractIDFromURL(urlStr string) string {
	if !strings.Contains(urlStr, "amazon") {
		return ""
	}
	// Simple extraction for /dp/ID or /product/ID
	parts := strings.Split(urlStr, "/")
	for i, part := range parts {
		if part == "dp" || part == "product" {
			if i+1 < len(parts) {
				return strings.Split(parts[i+1], "?")[0]
			}
		}
	}
	return ""
}
