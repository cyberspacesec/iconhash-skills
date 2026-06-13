package hasher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDiscoverFavicon_HTML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<html><head><link rel="icon" href="/custom-icon.png"><link rel="shortcut icon" href="/favicon.ico"></head><body></body></html>`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	h := New(nil)
	urls, err := h.DiscoverFavicon(context.Background(), server.URL, nil)
	if err != nil {
		t.Fatalf("DiscoverFavicon() error: %v", err)
	}
	if len(urls) < 2 {
		t.Errorf("Expected at least 2 URLs, got %d: %v", len(urls), urls)
	}
}

func TestDiscoverFavicon_CommonPaths(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	h := New(nil)
	urls, err := h.DiscoverFavicon(context.Background(), server.URL, nil)
	if err != nil {
		t.Fatalf("DiscoverFavicon() error: %v", err)
	}
	// Even without HTML favicon links, should return common paths
	if len(urls) == 0 {
		t.Error("Expected common path URLs")
	}
}

func TestDiscoverFavicon_NoCommonPaths(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	h := New(nil)
	opts := &DiscoverFaviconOptions{TryCommonPaths: false}
	_, err := h.DiscoverFavicon(context.Background(), server.URL, opts)
	if err == nil {
		t.Error("Expected error when no favicons found and TryCommonPaths=false")
	}
}

func TestDiscoverFavicon_ContextCancel(t *testing.T) {
	// When TryCommonPaths is false and HTML fetch fails due to cancelled context,
	// DiscoverFavicon should return an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {} // block forever
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	h := New(nil)
	opts := &DiscoverFaviconOptions{TryCommonPaths: false}
	_, err := h.DiscoverFavicon(ctx, server.URL, opts)
	if err == nil {
		t.Error("Expected error with cancelled context and TryCommonPaths=false")
	}
}

func TestParseFaviconLinks(t *testing.T) {
	htmlStr := `<html><head>
		<link rel="icon" type="image/png" href="/favicon.png">
		<link rel="shortcut icon" href="/old-favicon.ico">
		<link rel="stylesheet" href="/style.css">
		<link rel="apple-touch-icon" href="/apple.png">
	</head></html>`
	links, err := parseFaviconLinks(strings.NewReader(htmlStr))
	if err != nil {
		t.Fatalf("parseFaviconLinks() error: %v", err)
	}
	if len(links) != 3 {
		t.Errorf("Expected 3 favicon links, got %d: %v", len(links), links)
	}
}

func TestResolveURL(t *testing.T) {
	tests := []struct {
		base, href, expected string
	}{
		{"https://example.com", "https://cdn.example.com/icon.png", "https://cdn.example.com/icon.png"},
		{"https://example.com", "//cdn.example.com/icon.png", "https://cdn.example.com/icon.png"},
		{"https://example.com", "/favicon.ico", "https://example.com/favicon.ico"},
		{"https://example.com", "icons/favicon.ico", "https://example.com/icons/favicon.ico"},
	}
	for _, test := range tests {
		result := resolveURL(test.base, test.href)
		if result != test.expected {
			t.Errorf("resolveURL(%q, %q) = %q, expected %q", test.base, test.href, result, test.expected)
		}
	}
}