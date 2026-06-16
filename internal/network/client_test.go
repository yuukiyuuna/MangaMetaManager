package network

import (
	"net/url"
	"testing"
)

func TestShouldBypassProxyDefaults(t *testing.T) {
	factory := NewHTTPClientFactory()
	tests := []string{
		"http://localhost:8080",
		"http://127.0.0.1:8080",
		"http://192.168.1.10",
		"http://10.0.0.5",
		"http://172.16.0.5",
		"http://169.254.1.5",
		"http://[::1]:8080",
	}
	for _, rawURL := range tests {
		parsed, err := url.Parse(rawURL)
		if err != nil {
			t.Fatal(err)
		}
		if !factory.shouldBypassProxy(parsed, "") {
			t.Fatalf("expected %s to bypass proxy", rawURL)
		}
	}
}

func TestShouldBypassProxyNoProxyDomains(t *testing.T) {
	factory := NewHTTPClientFactory()
	parsed, err := url.Parse("https://api.example.com")
	if err != nil {
		t.Fatal(err)
	}
	if !factory.shouldBypassProxy(parsed, "example.com") {
		t.Fatal("expected suffix domain to bypass proxy")
	}
}

func TestShouldNotBypassPublicHostByDefault(t *testing.T) {
	factory := NewHTTPClientFactory()
	parsed, err := url.Parse("https://example.org")
	if err != nil {
		t.Fatal(err)
	}
	if factory.shouldBypassProxy(parsed, "") {
		t.Fatal("expected public host to use proxy")
	}
}
