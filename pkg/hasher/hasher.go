package hasher

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cyberspacesec/iconhash-skills/pkg/fingerprint"
	"github.com/twmb/murmur3"
)

// HashOptions defines options for icon hashing
type HashOptions struct {
	UseUint32          bool
	RequestTimeout     time.Duration
	InsecureSkipVerify bool
	UserAgent          string
	// HTTPClient allows injecting a custom http.Client for advanced scenarios
	// such as using proxies, custom TLS configs, or connection pooling.
	// When nil, a default client is created from the other options.
	HTTPClient *http.Client
	// MaxIconSize is the maximum size in bytes for icon content (default 10 MB).
	// Responses larger than this will be truncated.
	MaxIconSize int64
}

// DefaultOptions returns a HashOptions with sensible defaults
func DefaultOptions() *HashOptions {
	return &HashOptions{
		UseUint32:          false,
		RequestTimeout:     10 * time.Second,
		InsecureSkipVerify: true,
		UserAgent:          "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_0) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11",
		MaxIconSize:        10 << 20, // 10 MB
	}
}

// NewOptionsWithProxy creates HashOptions with a proxy configured from the given
// proxy URL string (supports http://, https://, socks5://). Returns an error if
// the proxy URL is invalid.
func NewOptionsWithProxy(proxyURL string, timeout time.Duration, insecureSkipVerify bool) (*HashOptions, error) {
	proxy, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	opts := DefaultOptions()
	opts.RequestTimeout = timeout
	opts.InsecureSkipVerify = insecureSkipVerify
	opts.HTTPClient = &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxy),
		},
	}

	return opts, nil
}

// IconHasher provides methods to calculate MMH3 hash of favicons
type IconHasher struct {
	options    *HashOptions
	httpClient *http.Client
}

// New creates a new IconHasher with the given options
func New(options *HashOptions) *IconHasher {
	if options == nil {
		options = DefaultOptions()
	}

	var httpClient *http.Client
	if options.HTTPClient != nil {
		httpClient = options.HTTPClient
	} else {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		if transport.TLSClientConfig != nil {
			transport.TLSClientConfig.InsecureSkipVerify = options.InsecureSkipVerify
		}
		httpClient = &http.Client{
			Timeout:   options.RequestTimeout,
			Transport: transport,
		}
	}

	return &IconHasher{
		options:    options,
		httpClient: httpClient,
	}
}

// Options returns a copy of the hasher's options.
func (h *IconHasher) Options() *HashOptions {
	return h.options
}

// HashFromURL downloads and calculates the hash of an icon from a URL.
// The ctx parameter supports cancellation and timeout propagation.
func (h *IconHasher) HashFromURL(ctx context.Context, url string) (string, error) {
	data, err := h.getContentFromURL(ctx, url)
	if err != nil {
		return "", fmt.Errorf("failed to get content from URL: %w", err)
	}

	return h.HashFromBytes(data)
}

// HashFromURLWithOption downloads and calculates the hash of an icon from a URL,
// allowing per-call override of the UseUint32 option without modifying shared state.
func (h *IconHasher) HashFromURLWithOption(ctx context.Context, url string, useUint32 bool) (string, error) {
	data, err := h.getContentFromURL(ctx, url)
	if err != nil {
		return "", fmt.Errorf("failed to get content from URL: %w", err)
	}

	return h.hashFromBytesWithOption(data, useUint32)
}

// HashFromFile calculates the hash of an icon from a file
func (h *IconHasher) HashFromFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return h.HashFromBytes(data)
}

// HashFromBase64 calculates the hash of an icon from base64 encoded data
func (h *IconHasher) HashFromBase64(base64Data string) (string, error) {
	// Strip prefix if exists
	if strings.HasPrefix(base64Data, "data:image/vnd.microsoft.icon;base64,") {
		base64Data = base64Data[37:]
	}

	// Validate base64 input
	if _, err := base64.StdEncoding.DecodeString(base64Data); err != nil {
		return "", fmt.Errorf("invalid base64 data: %w", err)
	}

	// Format with newlines for every 76 characters
	formattedBytes := h.formatBase64WithNewlines([]byte(base64Data))

	return h.calculateHash(formattedBytes)
}

// HashFromBase64WithOption calculates the hash of an icon from base64 encoded data,
// allowing per-call override of the UseUint32 option without modifying shared state.
func (h *IconHasher) HashFromBase64WithOption(base64Data string, useUint32 bool) (string, error) {
	if strings.HasPrefix(base64Data, "data:image/vnd.microsoft.icon;base64,") {
		base64Data = base64Data[37:]
	}

	// Validate base64 input
	if _, err := base64.StdEncoding.DecodeString(base64Data); err != nil {
		return "", fmt.Errorf("invalid base64 data: %w", err)
	}

	formattedBytes := h.formatBase64WithNewlines([]byte(base64Data))
	return h.calculateHashWithOption(formattedBytes, useUint32)
}

// HashFromBytes calculates the hash of an icon from bytes
func (h *IconHasher) HashFromBytes(data []byte) (string, error) {
	encodedBytes := h.standardBase64Encode(data)
	return h.calculateHash(encodedBytes)
}

// HashFromBytesWithOption calculates the hash of an icon from bytes,
// allowing per-call override of the UseUint32 option without modifying shared state.
func (h *IconHasher) HashFromBytesWithOption(data []byte, useUint32 bool) (string, error) {
	encodedBytes := h.standardBase64Encode(data)
	return h.calculateHashWithOption(encodedBytes, useUint32)
}

// hashFromBytesWithOption is an internal helper for per-call useUint32.
func (h *IconHasher) hashFromBytesWithOption(data []byte, useUint32 bool) (string, error) {
	encodedBytes := h.standardBase64Encode(data)
	return h.calculateHashWithOption(encodedBytes, useUint32)
}

// HashFromReader calculates the hash of an icon from an io.Reader.
// This is the standard Go idiom for streaming input sources (stdin, pipes, network streams).
func (h *IconHasher) HashFromReader(r io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(r, h.options.MaxIconSize))
	if err != nil {
		return "", fmt.Errorf("failed to read from reader: %w", err)
	}
	return h.HashFromBytes(data)
}

// HashFromReaderWithOption calculates the hash from a reader with per-call uint32 override
func (h *IconHasher) HashFromReaderWithOption(r io.Reader, useUint32 bool) (string, error) {
	data, err := io.ReadAll(io.LimitReader(r, h.options.MaxIconSize))
	if err != nil {
		return "", fmt.Errorf("failed to read from reader: %w", err)
	}
	return h.hashFromBytesWithOption(data, useUint32)
}

// ---------------------------------------------------------------------------
// DiscoverAndHash — combines DiscoverFavicon + HashFromURL
// ---------------------------------------------------------------------------

// DiscoverResult holds the result of discovering and hashing a single favicon.
type DiscoverResult struct {
	URL  string // The discovered favicon URL
	Hash string // The calculated MMH3 hash (empty if Err is set)
	Err  error  // Non-nil if this URL failed
}

// DiscoverAndHash discovers all favicon URLs for a site and calculates the hash
// for each. This combines DiscoverFavicon and HashFromURL into a single call,
// which is what the "iconhash discover" CLI command does.
func (h *IconHasher) DiscoverAndHash(ctx context.Context, siteURL string, opts *DiscoverFaviconOptions) []DiscoverResult {
	urls, err := h.DiscoverFavicon(ctx, siteURL, opts)
	if err != nil {
		return []DiscoverResult{{URL: siteURL, Err: err}}
	}

	results := make([]DiscoverResult, len(urls))
	for i, u := range urls {
		results[i] = DiscoverResult{URL: u}
		hash, err := h.HashFromURLWithOption(ctx, u, h.options.UseUint32)
		if err != nil {
			results[i].Err = err
		} else {
			results[i].Hash = hash
		}
	}
	return results
}

// ---------------------------------------------------------------------------
// Identify — discover + hash + fingerprint lookup
// ---------------------------------------------------------------------------

// IdentifyResult holds the result of identifying a website via its favicons.
type IdentifyResult struct {
	URL     string                         // The discovered favicon URL
	Hash    string                         // The calculated MMH3 hash (empty if Err is set)
	Matches []fingerprint.FingerprintEntry // Fingerprint matches (may be empty)
	Err     error                          // Non-nil if this URL failed
}

// Identify discovers all favicons for a site, hashes each, and looks up
// each hash in the provided fingerprint database. This is the all-in-one
// reconnaissance method equivalent to the "iconhash identify" CLI command.
func (h *IconHasher) Identify(ctx context.Context, siteURL string, db *fingerprint.DB, opts *DiscoverFaviconOptions) []IdentifyResult {
	urls, err := h.DiscoverFavicon(ctx, siteURL, opts)
	if err != nil {
		return []IdentifyResult{{URL: siteURL, Err: err}}
	}

	results := make([]IdentifyResult, len(urls))
	for i, u := range urls {
		results[i] = IdentifyResult{URL: u}
		hash, err := h.HashFromURLWithOption(ctx, u, h.options.UseUint32)
		if err != nil {
			results[i].Err = err
		} else {
			results[i].Hash = hash
			if db != nil {
				results[i].Matches = db.Lookup(hash)
			}
		}
	}
	return results
}

// ---------------------------------------------------------------------------
// Batch processing
// ---------------------------------------------------------------------------

// BatchResult holds the result of hashing a single URL in a batch operation.
type BatchResult struct {
	URL  string // The URL that was hashed
	Hash string // The calculated MMH3 hash (empty if Err is set)
	Err  error  // Non-nil if this URL failed
}

// BatchHashURLs hashes multiple favicon URLs concurrently. The concurrency
// parameter controls the maximum number of parallel HTTP requests (0 or 1 means
// sequential). This is the SDK equivalent of the "iconhash batch" CLI command.
func (h *IconHasher) BatchHashURLs(ctx context.Context, urls []string, concurrency int) []BatchResult {
	if concurrency < 1 {
		concurrency = 1
	}

	results := make([]BatchResult, len(urls))
	if concurrency <= 1 {
		// Sequential
		for i, u := range urls {
			results[i] = BatchResult{URL: u}
			hash, err := h.HashFromURL(ctx, u)
			if err != nil {
				results[i].Err = err
			} else {
				results[i].Hash = hash
			}
		}
		return results
	}

	// Concurrent with worker pool
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, u := range urls {
		wg.Add(1)
		go func(idx int, rawURL string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			results[idx] = BatchResult{URL: rawURL}
			hash, err := h.HashFromURL(ctx, rawURL)
			if err != nil {
				results[idx].Err = err
			} else {
				results[idx].Hash = hash
			}
		}(i, u)
	}
	wg.Wait()
	return results
}

// BatchHashFromReader reads URLs from a reader (one per line, skipping blanks
// and # comments), hashes each, and returns results. This is the SDK equivalent
// of piping URLs to the batch CLI command.
func (h *IconHasher) BatchHashFromReader(ctx context.Context, r io.Reader, concurrency int) ([]BatchResult, error) {
	var urls []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		urls = append(urls, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}

	return h.BatchHashURLs(ctx, urls, concurrency), nil
}

// ---------------------------------------------------------------------------
// Package-level convenience functions (default options, no IconHasher required)
// ---------------------------------------------------------------------------

// defaultHasher returns a shared IconHasher with default options for convenience functions.
var defaultHasher = New(nil)

// HashURL calculates the MMH3 hash of a favicon from a URL using default options.
// This is a convenience wrapper for one-off hash calculations.
func HashURL(ctx context.Context, rawURL string) (string, error) {
	return defaultHasher.HashFromURL(ctx, rawURL)
}

// HashFile calculates the MMH3 hash of a favicon from a local file using default options.
func HashFile(filePath string) (string, error) {
	return defaultHasher.HashFromFile(filePath)
}

// HashBase64 calculates the MMH3 hash from base64-encoded icon data using default options.
func HashBase64(base64Data string) (string, error) {
	return defaultHasher.HashFromBase64(base64Data)
}

// HashBytes calculates the MMH3 hash from raw icon bytes using default options.
func HashBytes(data []byte) (string, error) {
	return defaultHasher.HashFromBytes(data)
}

// HashReader calculates the MMH3 hash from an io.Reader using default options.
func HashReader(r io.Reader) (string, error) {
	return defaultHasher.HashFromReader(r)
}

// DiscoverFavicons discovers favicon URLs from a website using default options.
func DiscoverFavicons(ctx context.Context, siteURL string) ([]string, error) {
	return defaultHasher.DiscoverFavicon(ctx, siteURL, nil)
}

// IdentifySite discovers favicons, hashes them, and looks up fingerprints
// using default options.
func IdentifySite(ctx context.Context, siteURL string, db *fingerprint.DB) []IdentifyResult {
	return defaultHasher.Identify(ctx, siteURL, db, nil)
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func (h *IconHasher) getContentFromURL(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
		return nil, fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
	}

	return io.ReadAll(io.LimitReader(resp.Body, h.options.MaxIconSize))
}

// standardBase64Encode encodes bytes to base64 and formats with newlines
func (h *IconHasher) standardBase64Encode(data []byte) []byte {
	encodedStr := base64.StdEncoding.EncodeToString(data)
	return h.formatBase64WithNewlines([]byte(encodedStr))
}

// formatBase64WithNewlines formats base64 data with newlines every 76 characters
func (h *IconHasher) formatBase64WithNewlines(data []byte) []byte {
	var buffer bytes.Buffer

	for i := 0; i < len(data); i++ {
		buffer.WriteByte(data[i])
		if (i+1)%76 == 0 {
			buffer.WriteByte('\n')
		}
	}

	// Add final newline if not already present
	if len(data)%76 != 0 {
		buffer.WriteByte('\n')
	}

	return buffer.Bytes()
}

// calculateHash computes the MMH3 hash using the hasher's default UseUint32 option
func (h *IconHasher) calculateHash(data []byte) (string, error) {
	return h.calculateHashWithOption(data, h.options.UseUint32)
}

// calculateHashWithOption computes the MMH3 hash with an explicit useUint32 parameter.
// This is safe for concurrent use since it does not read shared mutable state.
func (h *IconHasher) calculateHashWithOption(data []byte, useUint32 bool) (string, error) {
	var h32 hash.Hash32 = murmur3.New32()
	_, err := h32.Write(data)
	if err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	if useUint32 {
		return fmt.Sprintf("%d", h32.Sum32()), nil
	}
	return fmt.Sprintf("%d", int32(h32.Sum32())), nil
}
