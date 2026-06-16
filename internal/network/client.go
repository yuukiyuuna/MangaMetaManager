package network

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/yuukiyuuna/MangaMetaManager/internal/models"
)

var (
	globalFactory     *HTTPClientFactory
	globalFactoryOnce sync.Once
)

// HTTPClientFactory creates a new http.Client based on proxy settings
type HTTPClientFactory struct {
	mu      sync.RWMutex
	clients map[string]*http.Client
}

func NewHTTPClientFactory() *HTTPClientFactory {
	globalFactoryOnce.Do(func() {
		globalFactory = &HTTPClientFactory{
			clients: make(map[string]*http.Client),
		}
	})
	return globalFactory
}

func (f *HTTPClientFactory) InvalidateCache() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.clients = make(map[string]*http.Client)
}

type userAgentTransport struct {
	rt http.RoundTripper
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "MangaMetaManager/1.0")
	}
	return t.rt.RoundTrip(req)
}

func (f *HTTPClientFactory) GetClient(providerID string) (*http.Client, error) {
	cacheKey := providerID
	if cacheKey == "" {
		cacheKey = "global"
	}

	f.mu.RLock()
	if client, ok := f.clients[cacheKey]; ok {
		f.mu.RUnlock()
		return client, nil
	}
	f.mu.RUnlock()

	var proxySettings models.ProxySettings
	var providerStrategy models.ProviderProxyStrategy

	var client *http.Client
	var err error

	// Get Global Proxy
	if errDB := models.DB.First(&proxySettings).Error; errDB != nil {
		// If no global proxy settings found, return default client
		client = &http.Client{
			Transport: &userAgentTransport{http.DefaultTransport},
			Timeout:   30 * time.Second,
		}
	} else {
		// Get Provider Strategy
		strategyFound := false
		if providerID != "" {
			if errDB := models.DB.Where("provider_id = ?", providerID).First(&providerStrategy).Error; errDB == nil {
				strategyFound = true
				switch providerStrategy.Strategy {
				case "disabled":
					client = &http.Client{
						Transport: &userAgentTransport{http.DefaultTransport},
						Timeout:   30 * time.Second,
					}
				case "custom":
					client, err = f.createCustomClient(&providerStrategy)
				case "inherit":
					// Fall through to global
					strategyFound = false
				}
			}
		}

		if !strategyFound {
			// Use Global Proxy if enabled
			if proxySettings.Enabled {
				client, err = f.createGlobalClient(&proxySettings)
			} else {
				client = &http.Client{
					Transport: &userAgentTransport{http.DefaultTransport},
					Timeout:   30 * time.Second,
				}
			}
		}
	}

	if err != nil {
		return nil, err
	}

	f.mu.Lock()
	f.clients[cacheKey] = client
	f.mu.Unlock()

	return client, nil
}

func (f *HTTPClientFactory) createGlobalClient(s *models.ProxySettings) (*http.Client, error) {
	proxyURL, err := f.buildProxyURL(s.Type, s.Host, s.Port, s.Username, s.Password)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: false},
	}

	// Handle NoProxy logic
	if s.NoProxy != "" {
		originalProxy := transport.Proxy
		transport.Proxy = func(req *http.Request) (*url.URL, error) {
			if f.shouldBypassProxy(req.URL, s.NoProxy) {
				return nil, nil
			}
			return originalProxy(req)
		}
	}

	timeout := time.Duration(s.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &http.Client{
		Transport: &userAgentTransport{transport},
		Timeout:   timeout,
	}, nil
}

func (f *HTTPClientFactory) createCustomClient(s *models.ProviderProxyStrategy) (*http.Client, error) {
	proxyURL, err := f.buildProxyURL(s.Type, s.Host, s.Port, s.Username, s.Password)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// Fetch global noProxy settings to apply to custom proxies
	var proxySettings models.ProxySettings
	if err := models.DB.First(&proxySettings).Error; err == nil && proxySettings.NoProxy != "" {
		originalProxy := transport.Proxy
		transport.Proxy = func(req *http.Request) (*url.URL, error) {
			if f.shouldBypassProxy(req.URL, proxySettings.NoProxy) {
				return nil, nil
			}
			return originalProxy(req)
		}
	} else {
		// Even without global noProxy, apply default bypass (localhost, etc.)
		originalProxy := transport.Proxy
		transport.Proxy = func(req *http.Request) (*url.URL, error) {
			if f.shouldBypassProxy(req.URL, "") {
				return nil, nil
			}
			return originalProxy(req)
		}
	}

	timeout := time.Duration(s.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &http.Client{
		Transport: &userAgentTransport{transport},
		Timeout:   timeout,
	}, nil
}

func (f *HTTPClientFactory) buildProxyURL(pType, host string, port int, user, pass string) (*url.URL, error) {
	var scheme string
	switch strings.ToLower(pType) {
	case "socks5":
		scheme = "socks5"
	case "https":
		scheme = "https"
	default:
		scheme = "http"
	}

	userInfo := ""
	if user != "" {
		if pass != "" {
			userInfo = fmt.Sprintf("%s:%s@", url.QueryEscape(user), url.QueryEscape(pass))
		} else {
			userInfo = fmt.Sprintf("%s@", url.QueryEscape(user))
		}
	}

	rawURL := fmt.Sprintf("%s://%s%s:%d", scheme, userInfo, host, port)
	return url.Parse(rawURL)
}

func (f *HTTPClientFactory) shouldBypassProxy(targetURL *url.URL, noProxy string) bool {
	host := strings.ToLower(targetURL.Hostname())
	bypassList := strings.Split(noProxy, ",")
	
	// Default bypass
	defaultBypass := []string{"localhost", "127.0.0.1", "::1"}
	for _, b := range defaultBypass {
		if host == b {
			return true
		}
	}

	for _, b := range bypassList {
		b = strings.ToLower(strings.TrimSpace(b))
		if b == "" {
			continue
		}
		
		// If b starts with a dot, match suffix
		if strings.HasPrefix(b, ".") {
			if strings.HasSuffix(host, b) || host == b[1:] {
				return true
			}
		} else {
			// Exact match for domains or IPs
			if host == b {
				return true
			}
			// Suffix match for domains (common in NoProxy)
			if strings.HasSuffix(host, "."+b) {
				return true
			}
		}
	}
	return false
}
