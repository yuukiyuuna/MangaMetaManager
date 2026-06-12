package provider

import (
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

func (p *FanzaProvider) Search(query string) ([]SearchResult, error) {
	return []SearchResult{}, nil
}

func (p *FanzaProvider) GetDetails(id string) (*metadata.ComicInfo, error) {
	return nil, nil
}
