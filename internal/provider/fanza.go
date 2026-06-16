package provider

import (
	"strings"

	"github.com/yuukiyuuna/MangaMetaManager/internal/metadata"
	"github.com/yuukiyuuna/MangaMetaManager/internal/network"
)

type FanzaProvider struct {
	factory *network.HTTPClientFactory
}

func NewFanzaProvider() *FanzaProvider {
	return &FanzaProvider{
		factory: network.NewHTTPClientFactory(),
	}
}

func (p *FanzaProvider) ID() string   { return "fanza" }
func (p *FanzaProvider) Name() string { return "FANZA" }
func (p *FanzaProvider) GetCustomTagName() string { return "CustomFanza" }

func (p *FanzaProvider) Search(query string) ([]SearchResult, error) {
	return nil, ErrNotImplemented
}

func (p *FanzaProvider) GetDetails(id string) (*metadata.ComicInfo, error) {
	return nil, ErrNotImplemented
}

func (p *FanzaProvider) GetRelatedBooks(id string) ([]SearchResult, error) {
	return nil, ErrNotImplemented
}

func (p *FanzaProvider) ExtractIDFromURL(urlStr string) string {
	if !strings.Contains(urlStr, "dmm.co.jp") && !strings.Contains(urlStr, "fanza.com") {
		return ""
	}
	// Extract cid=ID or from path
	if strings.Contains(urlStr, "cid=") {
		parts := strings.Split(urlStr, "cid=")
		return strings.Split(parts[1], "&")[0]
	}
	return ""
}
