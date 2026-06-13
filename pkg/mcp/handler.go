package mcp

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/cyberspacesec/iconhash-skills/pkg/fingerprint"
	"github.com/cyberspacesec/iconhash-skills/pkg/hasher"
	"github.com/cyberspacesec/iconhash-skills/pkg/util"
)

// Pre-compiled regex patterns (avoid recompiling on every request)
var (
	urlPattern    = regexp.MustCompile(`https?://[^\s]+`)
	base64Pattern = regexp.MustCompile(`(?:data:[^;]+;base64,)?([A-Za-z0-9+/=]+)`)
)

// Handler processes MCP requests and generates responses
type Handler struct {
	iconHasher *hasher.IconHasher
	logger     *util.Logger
	options    *hasher.HashOptions
	debug      bool
}

// ToolDefinition describes a tool in the MCP protocol (tools/list response).
type ToolDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema struct {
		Type       string                 `json:"type"`
		Properties map[string]interface{} `json:"properties"`
		Required   []string               `json:"required,omitempty"`
	} `json:"inputSchema"`
}

// ToolResult is the result returned by tools/call.
type ToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ToolContent is a single content item in a tool result.
type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Tools returns the list of available tools (standard MCP tools/list).
func (h *Handler) Tools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "iconhash_url",
			Description: "Calculate the MMH3 favicon hash from a URL. Supports multiple search engine formats (fofa, shodan, censys, quake, zoomeye, hunter).",
			InputSchema: struct {
				Type       string                 `json:"type"`
				Properties map[string]interface{} `json:"properties"`
				Required   []string               `json:"required,omitempty"`
			}{
				Type: "object",
				Properties: map[string]interface{}{
					"url":    map[string]string{"type": "string", "description": "URL of the favicon (e.g. https://example.com/favicon.ico)"},
					"format": map[string]string{"type": "string", "description": "Output format: plain, fofa, shodan, censys, quake, zoomeye, hunter (default: fofa)"},
					"uint32": map[string]string{"type": "boolean", "description": "Use uint32 format instead of int32 (default: false)"},
				},
				Required: []string{"url"},
			},
		},
		{
			Name:        "iconhash_base64",
			Description: "Calculate the MMH3 favicon hash from base64-encoded icon data.",
			InputSchema: struct {
				Type       string                 `json:"type"`
				Properties map[string]interface{} `json:"properties"`
				Required   []string               `json:"required,omitempty"`
			}{
				Type: "object",
				Properties: map[string]interface{}{
					"data":   map[string]string{"type": "string", "description": "Base64-encoded favicon data"},
					"format": map[string]string{"type": "string", "description": "Output format: plain, fofa, shodan, censys, quake, zoomeye, hunter (default: fofa)"},
					"uint32": map[string]string{"type": "boolean", "description": "Use uint32 format instead of int32 (default: false)"},
				},
				Required: []string{"data"},
			},
		},
		{
			Name:        "iconhash_file",
			Description: "Calculate the MMH3 favicon hash from a local file path.",
			InputSchema: struct {
				Type       string                 `json:"type"`
				Properties map[string]interface{} `json:"properties"`
				Required   []string               `json:"required,omitempty"`
			}{
				Type: "object",
				Properties: map[string]interface{}{
					"path":   map[string]string{"type": "string", "description": "Path to the favicon file on disk"},
					"format": map[string]string{"type": "string", "description": "Output format: plain, fofa, shodan, censys, quake, zoomeye, hunter (default: fofa)"},
					"uint32": map[string]string{"type": "boolean", "description": "Use uint32 format instead of int32 (default: false)"},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "iconhash_discover",
			Description: "Discover favicon URLs from a website by parsing HTML link tags and trying common paths, then calculate hashes for each.",
			InputSchema: struct {
				Type       string                 `json:"type"`
				Properties map[string]interface{} `json:"properties"`
				Required   []string               `json:"required,omitempty"`
			}{
				Type: "object",
				Properties: map[string]interface{}{
					"url":    map[string]string{"type": "string", "description": "Website URL to discover favicons from (e.g. https://example.com)"},
					"uint32": map[string]string{"type": "boolean", "description": "Use uint32 format instead of int32 (default: false)"},
				},
				Required: []string{"url"},
			},
		},
		{
			Name:        "iconhash_lookup",
			Description: "Lookup a favicon hash against the built-in fingerprint database to identify associated services and applications.",
			InputSchema: struct {
				Type       string                 `json:"type"`
				Properties map[string]interface{} `json:"properties"`
				Required   []string               `json:"required,omitempty"`
			}{
				Type: "object",
				Properties: map[string]interface{}{
					"hash": map[string]string{"type": "string", "description": "MMH3 favicon hash to look up (e.g. -305179312)"},
				},
				Required: []string{"hash"},
			},
		},
	}
}

// CallTool executes a tool by name with the given arguments (standard MCP tools/call).
func (h *Handler) CallTool(name string, args map[string]interface{}) *ToolResult {
	switch name {
	case "iconhash_url":
		return h.callToolURL(args)
	case "iconhash_base64":
		return h.callToolBase64(args)
	case "iconhash_file":
		return h.callToolFile(args)
	case "iconhash_discover":
		return h.callToolDiscover(args)
	case "iconhash_lookup":
		return h.callToolLookup(args)
	default:
		return &ToolResult{
			Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Unknown tool: %s", name)}},
			IsError: true,
		}
	}
}

func (h *Handler) callToolURL(args map[string]interface{}) *ToolResult {
	urlStr, ok := args["url"].(string)
	if !ok || urlStr == "" {
		return &ToolResult{
			Content: []ToolContent{{Type: "text", Text: "Error: 'url' parameter is required and must be a string"}},
			IsError: true,
		}
	}

	useUint32 := boolArg(args, "uint32")
	format := formatArg(args, "format")

	hash, err := h.iconHasher.HashFromURLWithOption(context.Background(), urlStr, useUint32)
	if err != nil {
		return &ToolResult{
			Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
			IsError: true,
		}
	}

	result := h.formatResult(urlStr, hash, format)
	return &ToolResult{Content: []ToolContent{{Type: "text", Text: result}}}
}

func (h *Handler) callToolBase64(args map[string]interface{}) *ToolResult {
	data, ok := args["data"].(string)
	if !ok || data == "" {
		return &ToolResult{
			Content: []ToolContent{{Type: "text", Text: "Error: 'data' parameter is required and must be a base64 string"}},
			IsError: true,
		}
	}

	useUint32 := boolArg(args, "uint32")
	format := formatArg(args, "format")

	hash, err := h.iconHasher.HashFromBase64WithOption(data, useUint32)
	if err != nil {
		return &ToolResult{
			Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
			IsError: true,
		}
	}

	result := h.formatResult("base64 data", hash, format)
	return &ToolResult{Content: []ToolContent{{Type: "text", Text: result}}}
}

func (h *Handler) callToolFile(args map[string]interface{}) *ToolResult {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return &ToolResult{Content: []ToolContent{{Type: "text", Text: "Error: 'path' parameter is required"}}, IsError: true}
	}
	useUint32 := boolArg(args, "uint32")
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		return &ToolResult{Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error reading file: %v", readErr)}}, IsError: true}
	}
	hash, err := h.iconHasher.HashFromBytesWithOption(data, useUint32)
	if err != nil {
		return &ToolResult{Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}, IsError: true}
	}
	format := formatArg(args, "format")
	result := h.formatResult(path, hash, format)
	return &ToolResult{Content: []ToolContent{{Type: "text", Text: result}}}
}

func (h *Handler) callToolDiscover(args map[string]interface{}) *ToolResult {
	siteURL, ok := args["url"].(string)
	if !ok || siteURL == "" {
		return &ToolResult{Content: []ToolContent{{Type: "text", Text: "Error: 'url' parameter is required"}}, IsError: true}
	}
	useUint32 := boolArg(args, "uint32")
	urls, err := h.iconHasher.DiscoverFavicon(context.Background(), siteURL, nil)
	if err != nil {
		return &ToolResult{Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}}, IsError: true}
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Discovered %d favicon URL(s) for %s:\n\n", len(urls), siteURL))
	for _, u := range urls {
		hash, err := h.iconHasher.HashFromURLWithOption(context.Background(), u, useUint32)
		if err != nil {
			sb.WriteString(fmt.Sprintf("- %s: Error: %v\n", u, err))
		} else {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", u, hash))
		}
	}
	return &ToolResult{Content: []ToolContent{{Type: "text", Text: sb.String()}}}
}

func (h *Handler) callToolLookup(args map[string]interface{}) *ToolResult {
	hash, ok := args["hash"].(string)
	if !ok || hash == "" {
		return &ToolResult{Content: []ToolContent{{Type: "text", Text: "Error: 'hash' parameter is required"}}, IsError: true}
	}
	db := fingerprint.DefaultDB()
	matches := db.Lookup(hash)
	if len(matches) == 0 {
		return &ToolResult{Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("No known services found for hash: %s", hash)}}}
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d match(es) for hash %s:\n\n", len(matches), hash))
	for _, m := range matches {
		sb.WriteString(fmt.Sprintf("- %s", m.Service))
		if m.Category != "" {
			sb.WriteString(fmt.Sprintf(" [%s]", m.Category))
		}
		sb.WriteString("\n")
	}
	return &ToolResult{Content: []ToolContent{{Type: "text", Text: sb.String()}}}
}

// NewHandler creates a new MCP handler
func NewHandler(debug bool) *Handler {
	options := &hasher.HashOptions{
		UseUint32:          false,
		RequestTimeout:     hasher.DefaultOptions().RequestTimeout,
		InsecureSkipVerify: true,
		UserAgent:          hasher.DefaultOptions().UserAgent,
	}

	return &Handler{
		iconHasher: hasher.New(options),
		logger:     util.NewLogger(debug),
		options:    options,
		debug:      debug,
	}
}

// NewHandlerWithHasher creates a new MCP handler using an existing IconHasher.
// This allows sharing the same hasher configuration (HTTP client, timeouts, TLS) with the API server.
func NewHandlerWithHasher(h *hasher.IconHasher, debug bool) *Handler {
	return &Handler{
		iconHasher: h,
		logger:     util.NewLogger(debug),
		options:    hasher.DefaultOptions(), // used only as fallback reference
		debug:      debug,
	}
}

// Process processes an MCP request and returns a response
func (h *Handler) Process(req *Request) (*Response, error) {
	// Validate the request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Create a response
	resp := NewResponse()

	// Extract the last user message
	var lastUserMessage string
	for i := len(req.Context.Messages) - 1; i >= 0; i-- {
		msg := req.Context.Messages[i]
		if msg.Role == "user" {
			lastUserMessage = msg.Content
			break
		}
	}

	if lastUserMessage == "" {
		resp.Message.Content = "Please provide a URL, file path, or base64 data to calculate the favicon hash."
		resp.Complete()
		return resp, nil
	}

	// Debug logging
	if h.debug {
		h.logger.Debugf("Processing MCP message: %s", lastUserMessage)
	}

	// Process the message and generate a response
	result, err := h.processMessage(lastUserMessage)
	if err != nil {
		resp.Message.Content = fmt.Sprintf("Error: %v", err)
		resp.Message.Meta["error"] = err.Error()
		resp.Complete()
		return resp, nil
	}

	resp.Message.Content = result
	resp.Complete()
	return resp, nil
}

// processMessage processes a user message and returns a result
func (h *Handler) processMessage(message string) (string, error) {
	// Check if the message contains a URL
	if urlPattern.MatchString(message) {
		urls := urlPattern.FindAllString(message, -1)
		if h.debug {
			h.logger.Debugf("Found URL in message: %s", urls[0])
		}
		return h.processURL(urls[0])
	}

	// Check if the message contains base64 data
	if base64Pattern.MatchString(message) {
		matches := base64Pattern.FindStringSubmatch(message)
		if len(matches) > 1 {
			// Check if it's a valid base64 string
			if _, err := base64.StdEncoding.DecodeString(matches[1]); err == nil {
				if h.debug {
					h.logger.Debugf("Found base64 data in message (length: %d)", len(matches[1]))
				}
				return h.processBase64(matches[1])
			}
		}
	}

	// Assume the message contains commands or requests about the tool
	if strings.Contains(strings.ToLower(message), "help") ||
		strings.Contains(strings.ToLower(message), "how to") ||
		strings.Contains(strings.ToLower(message), "example") {
		if h.debug {
			h.logger.Debugf("Detected help request, returning help text")
		}
		return h.getHelpText(), nil
	}

	return "", fmt.Errorf("I couldn't detect a URL, valid base64 data, or a help request in your message. Please provide a favicon URL or base64 data to calculate the hash.")
}

// processURL processes a URL and returns the hash
func (h *Handler) processURL(urlStr string) (string, error) {
	// Validate URL
	_, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %v", err)
	}

	if h.debug {
		h.logger.Debugf("Processing URL: %s", urlStr)
	}

	// Calculate hash
	hash, err := h.iconHasher.HashFromURL(context.Background(), urlStr)
	if err != nil {
		return "", fmt.Errorf("error calculating hash: %v", err)
	}

	if h.debug {
		h.logger.Debugf("Calculated hash: %s", hash)
	}

	return h.formatResult(urlStr, hash, util.FormatFofa), nil
}

// processBase64 processes base64 data and returns the hash
func (h *Handler) processBase64(data string) (string, error) {
	if h.debug {
		h.logger.Debugf("Processing base64 data (length: %d)", len(data))
	}

	// Calculate hash
	hash, err := h.iconHasher.HashFromBase64(data)
	if err != nil {
		return "", fmt.Errorf("error calculating hash: %v", err)
	}

	if h.debug {
		h.logger.Debugf("Calculated hash: %s", hash)
	}

	return h.formatResult("base64 data", hash, util.FormatFofa), nil
}

// formatResult formats a hash result for display
func (h *Handler) formatResult(source string, hash string, format util.OutputFormat) string {
	result := fmt.Sprintf("Favicon Hash for %s:\n\n", source)
	result += fmt.Sprintf("Plain hash: %s\n", hash)
	result += fmt.Sprintf("Fofa format: %s\n", util.FormatHash(hash, util.FormatFofa))
	result += fmt.Sprintf("Shodan format: %s\n", util.FormatHash(hash, util.FormatShodan))
	result += fmt.Sprintf("Censys format: %s\n", util.FormatHash(hash, util.FormatCensys))
	result += fmt.Sprintf("Quake format: %s\n", util.FormatHash(hash, util.FormatQuake))
	result += fmt.Sprintf("ZoomEye format: %s\n", util.FormatHash(hash, util.FormatZoomEye))
	result += fmt.Sprintf("Hunter format: %s\n", util.FormatHash(hash, util.FormatHunter))
	return result
}

// getHelpText returns help text for the MCP
func (h *Handler) getHelpText() string {
	return `# IconHash - Favicon Hash Calculator

This tool calculates the MMH3 hash of favicons for use in cybersecurity reconnaissance.

## Available Tools:

1. iconhash_url - Calculate hash from a URL
2. iconhash_base64 - Calculate hash from base64 data
3. iconhash_file - Calculate hash from a local file
4. iconhash_discover - Discover favicon URLs from a website
5. iconhash_lookup - Lookup a hash against the fingerprint database

## Examples:

1. Calculate hash from a URL:
   "Calculate the hash for https://example.com/favicon.ico"

2. Calculate hash from base64 data:
   "Calculate the hash for this base64: AAABAAEAEBAAAAEAIABoBAAAFgAAA..."

3. Discover favicons from a website:
   "Discover favicons on https://example.com"

4. Lookup a hash:
   "What service is hash -305179312?"

## How to use the results:

The hash can be used in search engines like:
- Fofa: icon_hash="123456789"
- Shodan: http.favicon.hash:123456789
- Censys: services.http.response.favicons.md5_hash:123456789
- Quake: favicon.hash:"123456789"
- ZoomEye: iconhash:"123456789"
- Hunter: web.icon="123456789"

For more information, visit: https://github.com/cyberspacesec/iconhash-skills
`
}

// boolArg extracts a boolean argument from the tool call arguments map.
func boolArg(args map[string]interface{}, key string) bool {
	val, ok := args[key]
	if !ok {
		return false
	}
	switch v := val.(type) {
	case bool:
		return v
	case string:
		return strings.ToLower(v) == "true" || v == "1"
	default:
		return false
	}
}

// formatArg extracts and parses the format argument from the tool call arguments map.
func formatArg(args map[string]interface{}, key string) util.OutputFormat {
	val, ok := args[key]
	if !ok {
		return util.FormatFofa
	}
	str, ok := val.(string)
	if !ok {
		return util.FormatFofa
	}
	return util.ParseFormat(strings.ToLower(str))
}
