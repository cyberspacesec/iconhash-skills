package hasher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

// DiscoverFaviconOptions controls the favicon discovery behavior
type DiscoverFaviconOptions struct {
	// TryCommonPaths indicates whether to try common favicon paths (/favicon.ico etc.)
	TryCommonPaths bool
}

// DefaultDiscoverOptions returns default discovery options
func DefaultDiscoverOptions() *DiscoverFaviconOptions {
	return &DiscoverFaviconOptions{
		TryCommonPaths: true,
	}
}

// commonFaviconPaths is the list of common favicon paths to try
var commonFaviconPaths = []string{
	"/favicon.ico",
	"/favicon.png",
	"/favicon-32x32.png",
	"/favicon-16x16.png",
	"/apple-touch-icon.png",
	"/mstile-144x144.png",
}

// DiscoverFavicon discovers favicon URLs from a given domain.
// Strategy: 1) Parse HTML <link rel="icon"> tags  2) Try common paths
func (h *IconHasher) DiscoverFavicon(ctx context.Context, siteURL string, opts *DiscoverFaviconOptions) ([]string, error) {
	if opts == nil {
		opts = DefaultDiscoverOptions()
	}

	parsedURL, err := url.Parse(siteURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	// Ensure scheme exists
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}
	baseURL := parsedURL.Scheme + "://" + parsedURL.Host

	var found []string
	seen := make(map[string]bool)

	// Strategy 1: Parse favicon links from HTML
	links, err := h.extractFaviconLinks(ctx, baseURL)
	if err == nil {
		for _, link := range links {
			resolved := resolveURL(baseURL, link)
			if resolved != "" && !seen[resolved] {
				seen[resolved] = true
				found = append(found, resolved)
			}
		}
	}

	// Strategy 2: Try common paths
	if opts.TryCommonPaths {
		for _, path := range commonFaviconPaths {
			fullURL := baseURL + path
			if !seen[fullURL] {
				seen[fullURL] = true
				found = append(found, fullURL)
			}
		}
	}

	if len(found) == 0 {
		return nil, fmt.Errorf("no favicon found for %s", siteURL)
	}
	return found, nil
}

// extractFaviconLinks fetches HTML and extracts favicon link tag hrefs
func (h *IconHasher) extractFaviconLinks(ctx context.Context, baseURL string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", h.options.UserAgent)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") && !strings.Contains(contentType, "application/xhtml") {
		return nil, fmt.Errorf("response is not HTML: %s", contentType)
	}

	return parseFaviconLinks(resp.Body)
}

// parseFaviconLinks parses favicon link tags from an HTML token stream
func parseFaviconLinks(body io.Reader) ([]string, error) {
	tokenizer := html.NewTokenizer(body)
	var links []string

	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			break
		}
		if tokenType == html.StartTagToken || tokenType == html.SelfClosingTagToken {
			token := tokenizer.Token()
			if token.Data == "link" {
				isIcon := false
				href := ""
				for _, attr := range token.Attr {
					switch attr.Key {
					case "rel":
						val := strings.ToLower(attr.Val)
						if strings.Contains(val, "icon") || strings.Contains(val, "shortcut") {
							isIcon = true
						}
					case "href":
						href = attr.Val
					}
				}
				if isIcon && href != "" {
					links = append(links, href)
				}
			}
		}
	}
	return links, nil
}

// resolveURL resolves a relative URL to an absolute URL
func resolveURL(baseURL, href string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	if strings.HasPrefix(href, "//") {
		parsed, err := url.Parse(baseURL)
		if err != nil {
			return ""
		}
		return parsed.Scheme + ":" + href
	}
	if strings.HasPrefix(href, "/") {
		return baseURL + href
	}
	return baseURL + "/" + href
}
