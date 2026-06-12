package provider

import (
	"fmt"
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
		list = append(list, p)
	}
	return list
}

func InitProviders() {
	Register(NewAmazonProvider())
	Register(NewFanzaProvider())
	Register(NewBangumiProvider())
}
