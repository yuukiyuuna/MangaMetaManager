package network

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/yuukiyuuna/MangaMetaManager/internal/models"
)

// HTTPClientFactory creates a new http.Client based on proxy settings
type HTTPClientFactory struct{}

func NewHTTPClientFactory() *HTTPClientFactory {
	return &HTTPClientFactory{}
}

func (f *HTTPClientFactory) GetClient(providerID string) (*http.Client, error) {
	var proxySettings *models.ProxySettings
	var providerStrategy models.ProviderProxyStrategy

	// Get Global Proxy
	if err := models.DB.First(&proxySettings).Error; err != nil {
		// If no global proxy settings found, return default client
		return &http.Client{Timeout: 30 * time.Second}, nil
	}

	// Get Provider Strategy
	if providerID != "" {
		if err := models.DB.Where("provider_id = ?", providerID).First(&providerStrategy).Error; err == nil {
			switch providerStrategy.Strategy {
			case "disabled":
				return &http.Client{Timeout: 30 * time.Second}, nil
			case "custom":
				return f.createCustomClient(&providerStrategy)
			case "inherit":
				// Fall through to global
			}
		}
	}

	// Use Global Proxy if enabled
	if proxySettings.Enabled {
		return f.createGlobalClient(proxySettings)
	}

	return &http.Client{Timeout: 30 * time.Second}, nil
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
		Transport: transport,
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

	timeout := time.Duration(s.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &http.Client{
		Transport: transport,
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
