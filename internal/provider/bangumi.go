package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/yuukiyuuna/MangaMetaManager/internal/metadata"
	"github.com/yuukiyuuna/MangaMetaManager/internal/network"
)

type BangumiProvider struct {
	factory *network.HTTPClientFactory
}

func NewBangumiProvider() *BangumiProvider {
	return &BangumiProvider{
		factory: network.NewHTTPClientFactory(),
	}
}

func (p *BangumiProvider) ID() string   { return "bangumi" }
func (p *BangumiProvider) Name() string { return "Bangumi" }
func (p *BangumiProvider) GetCustomTagName() string { return "CustomBangumi" }

type bgmV0SearchRequest struct {
	Keyword string `json:"keyword"`
	Sort    string `json:"sort"`
	Filter  struct {
		Type []int `json:"type"`
		NSFW bool  `json:"nsfw"`
	} `json:"filter"`
}

type bgmV0SearchResponse struct {
	Data []struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		NameCN string `json:"name_cn"`
		Date   string `json:"date"`
		Image  string `json:"image"`
	} `json:"data"`
}

func (p *BangumiProvider) Search(query string) ([]SearchResult, error) {
	client, err := p.factory.GetClient(p.ID())
	if err != nil {
		return nil, err
	}

	apiURL := "https://api.bgm.tv/v0/search/subjects"

	reqBody := bgmV0SearchRequest{
		Keyword: query,
		Sort:    "rank",
	}
	// type 1: book
	reqBody.Filter.Type = []int{1}
	// Enable nsfw to support R18 doujinshi etc.
	reqBody.Filter.NSFW = true

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bangumi api returned status %d", resp.StatusCode)
	}

	var bgmRes bgmV0SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&bgmRes); err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(bgmRes.Data))
	for _, item := range bgmRes.Data {
		title := item.NameCN
		if title == "" {
			title = item.Name
		}
		
		results = append(results, SearchResult{
			ID:          fmt.Sprintf("%d", item.ID),
			Title:       title,
			Series:      item.Name,
			CoverURL:    item.Image,
			ReleaseDate: item.Date,
		})
	}

	return results, nil
}

func (p *BangumiProvider) GetDetails(id string) (*metadata.ComicInfo, error) {
	client, err := p.factory.GetClient(p.ID())
	if err != nil {
		return nil, err
	}

	// Get subject details
	apiURL := fmt.Sprintf("https://api.bgm.tv/v0/subjects/%s", id)
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("User-Agent", "MangaMetaManager/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var item struct {
		Name    string `json:"name"`
		NameCN  string `json:"name_cn"`
		Summary string `json:"summary"`
		Infobox []struct {
			Key   string      `json:"key"`
			Value interface{} `json:"value"`
		} `json:"infobox"`
		Tags []struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		} `json:"tags"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}

	var title = item.NameCN
	if title == "" {
		title = item.Name
	}

	info := &metadata.ComicInfo{
		Title:         title,
		Series:        item.Name,
		OriginalTitle: item.Name,
		Summary:       item.Summary,
		Web:           fmt.Sprintf("https://bgm.tv/subject/%s", id),
	}

	var authors []string
	var translators []string
	var genres []string
	var tags []string

	for i, t := range item.Tags {
		if i < 8 {
			genres = append(genres, t.Name)
		}
		tags = append(tags, t.Name)
	}
	if len(genres) > 0 {
		info.Genre = strings.Join(genres, ", ")
	}
	if len(tags) > 0 {
		info.Tags = strings.Join(tags, ", ")
	}

	// Parse Infobox
	for _, field := range item.Infobox {
		key := field.Key
		if key == "作者" || key == "原作" || key == "作画" {
			if v, ok := field.Value.(string); ok {
				authors = append(authors, v)
			} else if v, ok := field.Value.([]interface{}); ok && len(v) > 0 {
				if first, ok := v[0].(map[string]interface{}); ok {
					if name, ok := first["v"].(string); ok {
						authors = append(authors, name)
					}
				}
			}
		} else if key == "译者" || key == "翻译" {
			if v, ok := field.Value.(string); ok {
				translators = append(translators, v)
			} else if v, ok := field.Value.([]interface{}); ok && len(v) > 0 {
				if first, ok := v[0].(map[string]interface{}); ok {
					if name, ok := first["v"].(string); ok {
						translators = append(translators, name)
					}
				}
			}
		} else if key == "出版社" {
			if v, ok := field.Value.(string); ok {
				info.Publisher = v
			} else if v, ok := field.Value.([]interface{}); ok && len(v) > 0 {
				if first, ok := v[0].(map[string]interface{}); ok {
					if name, ok := first["v"].(string); ok {
						info.Publisher = name
					}
				}
			}
		} else if key == "发售日" {
			if v, ok := field.Value.(string); ok {
				var y, m, d int
				fmt.Sscanf(v, "%d-%d-%d", &y, &m, &d)
				if y != 0 { info.Year = y }
				if m != 0 { info.Month = m }
				if d != 0 { info.Day = d }
			}
		} else if key == "页数" {
			if v, ok := field.Value.(string); ok {
				var p int
				fmt.Sscanf(v, "%d", &p)
				if p != 0 { info.PageCount = p }
			}
		} else if key == "ISBN" {
			if v, ok := field.Value.(string); ok {
				info.GTIN = v
			}
		}
	}

	if len(authors) > 0 {
		info.Writer = strings.Join(authors, ", ")
	}
	if len(translators) > 0 {
		info.Translator = strings.Join(translators, ", ")
	}

	return info, nil
}

func (p *BangumiProvider) GetRelatedBooks(id string) ([]SearchResult, error) {
	client, err := p.factory.GetClient(p.ID())
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.bgm.tv/v0/subjects/%s/subjects", id)
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("User-Agent", "MangaMetaManager/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var bgmRes []struct {
		ID     int    `json:"id"`
		Type   int    `json:"type"`
		Name   string `json:"name"`
		NameCN string `json:"name_cn"`
		Image  string `json:"image"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&bgmRes); err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0)
	for _, item := range bgmRes {
		// Only Book type (1)
		if item.Type != 1 {
			continue
		}

		title := item.NameCN
		if title == "" {
			title = item.Name
		}
		
		results = append(results, SearchResult{
			ID:       fmt.Sprintf("%d", item.ID),
			Title:    title,
			Series:   item.Name,
			CoverURL: item.Image,
		})
	}

	return results, nil
}

func (p *BangumiProvider) ExtractIDFromURL(urlStr string) string {
	if !strings.Contains(urlStr, "bgm.tv") {
		return ""
	}
	parts := strings.Split(strings.TrimSuffix(urlStr, "/"), "/")
	return parts[len(parts)-1]
}
