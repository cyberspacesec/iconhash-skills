package util

import (
	"net/url"
	"strings"
)

// IsURL checks if a string looks like a valid HTTP/HTTPS URL.
// It requires a scheme (http:// or https://) or a www. prefix.
func IsURL(s string) bool {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		_, err := url.ParseRequestURI(s)
		return err == nil
	}
	if strings.HasPrefix(s, "www.") {
		return true
	}
	return false
}
