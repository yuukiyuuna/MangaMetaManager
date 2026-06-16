package provider

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

var (
	providers = make(map[string]Provider)
	mu        sync.RWMutex
)

func Register(p Provider) {
	mu.Lock()
	defer mu.Unlock()
	providers[p.ID()] = p
}

func GetProvider(id string) (Provider, error) {
	mu.RLock()
	defer mu.RUnlock()
	p, ok := providers[id]
	if !ok {
		return nil, fmt.Errorf("provider %s not found", id)
	}
	return p, nil
}

func ListProviders() []Provider {
	mu.RLock()
	defer mu.RUnlock()
	list := make([]Provider, 0, len(providers))
	for _, p := range providers {
		if !IsImplemented(p.ID()) {
			continue
		}
		list = append(list, p)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name() < list[j].Name()
	})
	return list
}

func IsImplemented(id string) bool {
	switch id {
	case "bangumi":
		return true
	default:
		return false
	}
}

func InitProviders() {
	Register(NewAmazonProvider())
	Register(NewFanzaProvider())
	Register(NewBangumiProvider())
}

func GetTagNameByURL(url string) string {
	if url == "" {
		return ""
	}
	if strings.Contains(url, "bgm.tv") {
		return "CustomBangumi"
	}
	if strings.Contains(url, "amazon") {
		return "CustomAmazon"
	}
	if strings.Contains(url, "dmm") || strings.Contains(url, "fanza") {
		return "CustomFanza"
	}
	return ""
}
