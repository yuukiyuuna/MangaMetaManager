package provider

import (
	"github.com/yuukiyuuna/MangaMetaManager/internal/metadata"
)

type SearchResult struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Series      string `json:"series"`
	Publisher   string `json:"publisher"`
	CoverURL    string `json:"coverUrl"`
	ReleaseDate string `json:"releaseDate"`
}

type Provider interface {
	ID() string
	Name() string
	Search(query string) ([]SearchResult, error)
	GetDetails(id string) (*metadata.ComicInfo, error)
	GetCustomTagName() string
	GetRelatedBooks(id string) ([]SearchResult, error)
	ExtractIDFromURL(url string) string
}
