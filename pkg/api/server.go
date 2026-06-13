package api

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"
	neturl "net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cyberspacesec/iconhash-skills/pkg/fingerprint"
	"github.com/cyberspacesec/iconhash-skills/pkg/hasher"
	"github.com/cyberspacesec/iconhash-skills/pkg/mcp"
	"github.com/cyberspacesec/iconhash-skills/pkg/util"
)

// maxRequestBody is the maximum size for request bodies (10 MB).
const maxRequestBody = 10 << 20

// Server represents the HTTP API server
type Server struct {
	config      *Config
	iconHasher  *hasher.IconHasher
	logger      *util.Logger
	mcpHandler  *mcp.Handler
	fingerprint *fingerprint.DB
	debug       bool
	server      *http.Server
}

// Config holds the server configuration
type Config struct {
	Host               string
	Port               int
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	AuthToken          string
	EnableDebug        bool
	InsecureSkipVerify bool
	RequestTimeout     time.Duration
	Proxy              string // HTTP/SOCKS5 proxy URL
	FingerprintDB      string // Path to custom fingerprint JSON database
}

// DefaultConfig returns a default server configuration
func DefaultConfig() *Config {
	return &Config{
		Host:               "127.0.0.1",
		Port:               8080,
		ReadTimeout:        30 * time.Second,
		WriteTimeout:       30 * time.Second,
		AuthToken:          "",
		EnableDebug:        false,
		InsecureSkipVerify: true,
		RequestTimeout:     10 * time.Second,
	}
}

// NewServer creates a new API server
func NewServer(config *Config) *Server {
	if config == nil {
		config = DefaultConfig()
	}

	// Create options for hasher
	options := &hasher.HashOptions{
		UseUint32:          false,
		RequestTimeout:     config.RequestTimeout,
		InsecureSkipVerify: config.InsecureSkipVerify,
		UserAgent:          "IconHash API Server",
	}

	// Configure proxy if specified
	if config.Proxy != "" {
		proxyURL, err := neturl.Parse(config.Proxy)
		if err == nil {
			options.HTTPClient = &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
				},
				Timeout: config.RequestTimeout,
			}
		}
	}

	// Create the server
	logger := util.NewLogger(config.EnableDebug)

	// Create standard icon hasher
	h := hasher.New(options)

	// Create fingerprint database
	fpDB := fingerprint.DefaultDB()
	if config.FingerprintDB != "" {
		if err := fpDB.LoadFromJSON(config.FingerprintDB); err != nil {
			logger.Debugf("Warning: failed to load custom fingerprint DB: %v", err)
		} else {
			logger.Debugf("Loaded custom fingerprint database: %s", config.FingerprintDB)
		}
	}

	return &Server{
		config:      config,
		iconHasher:  h,
		logger:      logger,
		mcpHandler:  mcp.NewHandlerWithHasher(h, config.EnableDebug),
		fingerprint: fpDB,
		debug:       config.EnableDebug,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Create router
	mux := http.NewServeMux()

	// Setup routes
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/hash/url", s.handleHashURL)
	mux.HandleFunc("/hash/file", s.handleHashFile)
	mux.HandleFunc("/hash/base64", s.handleHashBase64)
	mux.HandleFunc("/hash/batch", s.handleHashBatch)
	mux.HandleFunc("/hash/discover", s.handleHashDiscover)
	mux.HandleFunc("/lookup", s.handleLookup)
	mux.HandleFunc("/fingerprints", s.handleFingerprints)
	mux.HandleFunc("/mcp", s.handleMCP)

	// Wrap with CORS middleware first, then auth if token is set
	handler := corsMiddleware(mux)
	if s.config.AuthToken != "" {
		handler = s.authMiddleware(handler)
	}

	// Create server
	addr := s.config.Host + ":" + strconv.Itoa(s.config.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	// Log startup
	if s.debug {
		s.logger.Debugf("Starting server on %s", addr)
		s.logger.Debugf("Auth token: %v", s.config.AuthToken != "")
		s.logger.Debugf("Debug enabled: %v", s.config.EnableDebug)
	}

	// Start server
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// corsMiddleware adds CORS headers to all responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// authMiddleware adds authentication to routes
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health endpoint
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		// Check for token in header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			// Check if it's a Bearer token
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				token := authHeader[7:]
				if subtle.ConstantTimeCompare([]byte(token), []byte(s.config.AuthToken)) == 1 {
					next.ServeHTTP(w, r)
					return
				}
			}
		}

		// Check for token in query parameter
		token := r.URL.Query().Get("token")
		if subtle.ConstantTimeCompare([]byte(token), []byte(s.config.AuthToken)) == 1 {
			next.ServeHTTP(w, r)
			return
		}

		// Unauthorized
		sendErrorResponse(w, "Unauthorized: Invalid or missing authentication token", http.StatusUnauthorized)
	})
}

// handleHealth handles the health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"version": "v1",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// MCPRequest is a generic JSON-RPC style MCP request.
// It supports both the legacy chat-style protocol and the standard tools/* methods.
type MCPRequest struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params,omitempty"`
	// Legacy chat-style fields
	Context interface{} `json:"context,omitempty"`
}

// MCPResponse is a generic JSON-RPC style MCP response.
type MCPResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// handleMCP handles the Model Context Protocol endpoint.
// Supports both standard MCP methods (tools/list, tools/call) and the legacy chat protocol.
func (s *Server) handleMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBody))
	if err != nil {
		sendErrorResponse(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Try to parse as a standard MCP JSON-RPC request first
	var mcpReq MCPRequest
	if err := json.Unmarshal(body, &mcpReq); err == nil && mcpReq.Method != "" {
		switch mcpReq.Method {
		case "tools/list":
			tools := s.mcpHandler.Tools()
			json.NewEncoder(w).Encode(MCPResponse{Result: map[string]interface{}{"tools": tools}})
			return
		case "tools/call":
			if mcpReq.Params == nil {
				json.NewEncoder(w).Encode(MCPResponse{Error: "params is required for tools/call"})
				return
			}
			name, _ := mcpReq.Params["name"].(string)
			arguments, _ := mcpReq.Params["arguments"].(map[string]interface{})
			result := s.mcpHandler.CallTool(name, arguments)
			json.NewEncoder(w).Encode(MCPResponse{Result: result})
			return
		}
	}

	// Fallback: legacy chat-style MCP protocol
	var req mcp.Request
	if err := json.Unmarshal(body, &req); err != nil {
		sendErrorResponse(w, "Invalid MCP request: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.mcpHandler.Process(&req)
	if err != nil {
		sendErrorResponse(w, "Error processing MCP request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respData, err := json.Marshal(resp)
	if err != nil {
		sendErrorResponse(w, "Error serializing response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respData)
}

// HashResponse is the response format for hash endpoints
type HashResponse struct {
	Hash      string `json:"hash"`
	Format    string `json:"format,omitempty"`
	Formatted string `json:"formatted,omitempty"`
	Error     string `json:"error,omitempty"`
}

// handleHashURL handles the hash from URL endpoint.
// Accepts GET with query params, POST with form-encoded body, or POST with JSON body.
func (s *Server) handleHashURL(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Get URL from query or form
	var urlStr string
	if r.Method == http.MethodGet {
		urlStr = r.URL.Query().Get("url")
	} else if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "application/json") {
			body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBody))
			if err != nil {
				sendErrorResponse(w, "Error reading request body", http.StatusBadRequest)
				return
			}
			var jsonBody struct {
				URL string `json:"url"`
			}
			if err := json.Unmarshal(body, &jsonBody); err != nil {
				sendErrorResponse(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
				return
			}
			urlStr = jsonBody.URL
		} else {
			if err := r.ParseForm(); err != nil {
				sendErrorResponse(w, "Error parsing form", http.StatusBadRequest)
				return
			}
			urlStr = r.FormValue("url")
		}
	} else {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate URL
	if urlStr == "" {
		sendErrorResponse(w, "URL parameter is required", http.StatusBadRequest)
		return
	}

	// Check for format parameter
	formatStr := r.URL.Query().Get("format")
	format := parseFormatParam(formatStr)

	// Update hasher options if uint32 is specified
	uint32Param := r.URL.Query().Get("uint32")
	useUint32 := uint32Param == "true" || uint32Param == "1"
	// useUint32 determined per-request (thread-safe)

	// Debug output
	if s.debug {
		s.logger.Debugf("URL hash request: %s", urlStr)
		s.logger.Debugf("Format: %s", getFormatName(format))
		s.logger.Debugf("UseUint32: %v", useUint32)
	}

	// Calculate hash
	hash, err := s.iconHasher.HashFromURLWithOption(r.Context(), urlStr, useUint32)
	if err != nil {
		sendErrorResponse(w, "Error calculating hash: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Format hash based on requested format
	formatted := util.FormatHash(hash, format)
	sendHashResponse(w, hash, getFormatName(format), formatted)
}

// handleHashFile handles the hash from file upload endpoint
func (s *Server) handleHashFile(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Validate method
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // Max 10MB
	if err != nil {
		sendErrorResponse(w, "Error parsing multipart form", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, _, err := r.FormFile("file")
	if err != nil {
		sendErrorResponse(w, "Error getting file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file
	fileData, err := io.ReadAll(file)
	if err != nil {
		sendErrorResponse(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	// Check for format parameter
	formatStr := r.URL.Query().Get("format")
	format := parseFormatParam(formatStr)

	// Update hasher options if uint32 is specified
	uint32Param := r.URL.Query().Get("uint32")
	useUint32 := uint32Param == "true" || uint32Param == "1"
	// useUint32 determined per-request (thread-safe)

	// Debug output
	if s.debug {
		s.logger.Debugf("File hash request: %d bytes", len(fileData))
		s.logger.Debugf("Format: %s", getFormatName(format))
		s.logger.Debugf("UseUint32: %v", useUint32)
	}

	// Calculate hash
	hash, err := s.iconHasher.HashFromBytesWithOption(fileData, useUint32)
	if err != nil {
		sendErrorResponse(w, "Error calculating hash: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Format hash based on requested format
	formatted := util.FormatHash(hash, format)
	sendHashResponse(w, hash, getFormatName(format), formatted)
}

// handleHashBase64 handles the hash from base64 endpoint.
// Accepts both form-encoded (Content-Type: application/x-www-form-urlencoded)
// and JSON (Content-Type: application/json) request bodies.
func (s *Server) handleHashBase64(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Validate method
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var base64Data string

	// Try JSON body first, fallback to form-encoded
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBody))
		if err != nil {
			sendErrorResponse(w, "Error reading request body", http.StatusBadRequest)
			return
		}
		var jsonBody struct {
			Data string `json:"data"`
		}
		if err := json.Unmarshal(body, &jsonBody); err != nil {
			sendErrorResponse(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
			return
		}
		base64Data = jsonBody.Data
	} else {
		// Parse form
		if err := r.ParseForm(); err != nil {
			sendErrorResponse(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		base64Data = r.FormValue("data")
	}

	if base64Data == "" {
		sendErrorResponse(w, "Base64 data is required", http.StatusBadRequest)
		return
	}

	// Check for format parameter
	formatStr := r.URL.Query().Get("format")
	format := parseFormatParam(formatStr)

	// Update hasher options if uint32 is specified
	uint32Param := r.URL.Query().Get("uint32")
	useUint32 := uint32Param == "true" || uint32Param == "1"
	// useUint32 determined per-request (thread-safe)

	// Debug output
	if s.debug {
		s.logger.Debugf("Base64 hash request: %d bytes", len(base64Data))
		s.logger.Debugf("Format: %s", getFormatName(format))
		s.logger.Debugf("UseUint32: %v", useUint32)
	}

	// Calculate hash
	hash, err := s.iconHasher.HashFromBase64WithOption(base64Data, useUint32)
	if err != nil {
		sendErrorResponse(w, "Error calculating hash: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Format hash based on requested format
	formatted := util.FormatHash(hash, format)
	sendHashResponse(w, hash, getFormatName(format), formatted)
}

// BatchHashItem is a single item in a batch hash request.
type BatchHashItem struct {
	URL  string `json:"url"`
	Hash string `json:"hash,omitempty"`
	Err  string `json:"error,omitempty"`
}

// BatchHashResponse is the response format for the batch hash endpoint.
type BatchHashResponse struct {
	Results []BatchHashItem `json:"results"`
}

// handleHashBatch handles the batch hash endpoint.
// Accepts a JSON array of URLs and returns hashes for each.
func (s *Server) handleHashBatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBody))
	if err != nil {
		sendErrorResponse(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var urls []string
	if err := json.Unmarshal(body, &urls); err != nil {
		// Try object form: {"urls": [...]}
		var objReq struct {
			URLs []string `json:"urls"`
		}
		if err2 := json.Unmarshal(body, &objReq); err2 != nil {
			sendErrorResponse(w, "Invalid request body: expected a JSON array of URLs or {\"urls\": [...]}", http.StatusBadRequest)
			return
		}
		urls = objReq.URLs
	}

	if len(urls) == 0 {
		sendErrorResponse(w, "At least one URL is required", http.StatusBadRequest)
		return
	}

	// Limit batch size to prevent resource exhaustion
	if len(urls) > 100 {
		sendErrorResponse(w, "Batch size too large: maximum 100 URLs per request", http.StatusBadRequest)
		return
	}

	// Check for format parameter
	formatStr := r.URL.Query().Get("format")
	format := parseFormatParam(formatStr)

	// Update hasher options if uint32 is specified
	uint32Param := r.URL.Query().Get("uint32")
	useUint32 := uint32Param == "true" || uint32Param == "1"
	// Concurrent processing with worker pool
	results := make([]BatchHashItem, len(urls))
	sem := make(chan struct{}, 10) // 10 concurrent workers
	var wg sync.WaitGroup

	for i, urlStr := range urls {
		wg.Add(1)
		go func(idx int, u string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			item := BatchHashItem{URL: u}
			hash, err := s.iconHasher.HashFromURLWithOption(r.Context(), u, useUint32)
			if err != nil {
				item.Err = err.Error()
			} else {
				item.Hash = util.FormatHash(hash, format)
			}
			results[idx] = item
		}(i, urlStr)
	}
	wg.Wait()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(BatchHashResponse{Results: results})
}

// DiscoverResult is a single result from the discover endpoint
type DiscoverResult struct {
	URL  string `json:"url"`
	Hash string `json:"hash,omitempty"`
	Err  string `json:"error,omitempty"`
}

// DiscoverResponse is the response format for the discover endpoint
type DiscoverResponse struct {
	SiteURL string            `json:"site_url"`
	Results []DiscoverResult `json:"results"`
}

// handleHashDiscover discovers favicon URLs from a domain and calculates their hashes
func (s *Server) handleHashDiscover(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBody))
	if err != nil {
		sendErrorResponse(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var jsonBody struct {
		URL string `json:"url"`
	}
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		if err := json.Unmarshal(body, &jsonBody); err != nil {
			sendErrorResponse(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		r.ParseForm()
		jsonBody.URL = r.FormValue("url")
	}

	if jsonBody.URL == "" {
		sendErrorResponse(w, "URL parameter is required", http.StatusBadRequest)
		return
	}

	uint32Param := r.URL.Query().Get("uint32")
	useUint32 := uint32Param == "true" || uint32Param == "1"

	urls, err := s.iconHasher.DiscoverFavicon(r.Context(), jsonBody.URL, nil)
	if err != nil {
		sendErrorResponse(w, "Error discovering favicon: "+err.Error(), http.StatusInternalServerError)
		return
	}

	results := make([]DiscoverResult, 0, len(urls))
	for _, u := range urls {
		item := DiscoverResult{URL: u}
		hash, err := s.iconHasher.HashFromURLWithOption(r.Context(), u, useUint32)
		if err != nil {
			item.Err = err.Error()
		} else {
			item.Hash = hash
		}
		results = append(results, item)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(DiscoverResponse{SiteURL: jsonBody.URL, Results: results})
}

// handleLookup handles the fingerprint lookup endpoint
func (s *Server) handleLookup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hash := r.URL.Query().Get("hash")
	if hash == "" {
		sendErrorResponse(w, "hash parameter is required", http.StatusBadRequest)
		return
	}

	results := s.fingerprint.Lookup(hash)

	// Generate search engine queries for cyber-space mapping
	searchQueries := map[string]string{
		"fofa":    util.FormatHash(hash, util.FormatFofa),
		"shodan":  util.FormatHash(hash, util.FormatShodan),
		"censys":  util.FormatHash(hash, util.FormatCensys),
		"quake":   util.FormatHash(hash, util.FormatQuake),
		"zoomeye": util.FormatHash(hash, util.FormatZoomEye),
		"hunter":  util.FormatHash(hash, util.FormatHunter),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"hash":           hash,
		"matches":        results,
		"search_queries": searchQueries,
	})
}

// handleFingerprints handles the fingerprints listing endpoint
func (s *Server) handleFingerprints(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Support filtering by category, service, or tag
	category := r.URL.Query().Get("category")
	service := r.URL.Query().Get("service")
	tag := r.URL.Query().Get("tag")

	var entries []fingerprint.FingerprintEntry

	// Use efficient search methods
	if service != "" {
		entries = s.fingerprint.Search(service)
	} else if tag != "" {
		entries = s.fingerprint.Search(tag)
	} else if category != "" {
		entries = s.fingerprint.LookupByCategory(category)
	} else {
		entries = s.fingerprint.All()
	}

	// Apply cross-filters
	if category != "" && (service != "" || tag != "") {
		var filtered []fingerprint.FingerprintEntry
		for _, e := range entries {
			if !strings.EqualFold(e.Category, category) {
				continue
			}
			filtered = append(filtered, e)
		}
		entries = filtered
	}

	// Apply tag filter on top of service search
	if tag != "" && service != "" {
		var filtered []fingerprint.FingerprintEntry
		tagLower := strings.ToLower(tag)
		for _, e := range entries {
			for _, t := range e.Tags {
				if strings.ToLower(t) == tagLower {
					filtered = append(filtered, e)
					break
				}
			}
		}
		entries = filtered
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":        s.fingerprint.Count(),
		"total_entries": s.fingerprint.TotalEntries(),
		"returned":     len(entries),
		"fingerprints": entries,
	})
}

// sendHashResponse sends a hash response
func sendHashResponse(w http.ResponseWriter, hash, formatName, formatted string) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HashResponse{
		Hash:      hash,
		Format:    formatName,
		Formatted: formatted,
	})
}

// sendErrorResponse sends an error response
func sendErrorResponse(w http.ResponseWriter, errMessage string, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(HashResponse{
		Error: errMessage,
	})
}

// parseFormatParam parses the format parameter using the SDK's ParseFormat.
// Defaults to FormatFofa for empty or unrecognized strings.
func parseFormatParam(formatStr string) util.OutputFormat {
	if formatStr == "" {
		return util.FormatFofa
	}
	f := util.ParseFormat(formatStr)
	if f == util.FormatPlain && formatStr != "plain" {
		// Unrecognized format, default to Fofa
		return util.FormatFofa
	}
	return f
}

// getFormatName returns the name of the format using the SDK's FormatName.
func getFormatName(format util.OutputFormat) string {
	return util.FormatName(format)
}
