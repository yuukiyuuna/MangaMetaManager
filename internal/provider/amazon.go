package provider

import (
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

func (p *AmazonProvider) Search(query string) ([]SearchResult, error) {
	// client, err := p.factory.GetClient(p.ID())
	// ... implementation ...
	return []SearchResult{}, nil
}

func (p *AmazonProvider) GetDetails(id string) (*metadata.ComicInfo, error) {
	// client, err := p.factory.GetClient(p.ID())
	// ... implementation ...
	return nil, nil
}
