# SDK API Capability Expansion Plan

> **For agentic workers:** REQUIRED SUB-SKILL: `superpowers:subagent-driven-development`
> Steps use checkbox (`- [ ]`) syntax.

**Goal:** 补全 go-iconhash SDK 缺失的关键能力，使其从"基础 hash 计算器"升级为"完整的 favicon 情报 SDK"，对齐竞品工具（fav-up、FavFreak、EHole）的核心能力。

**Architecture:** 三层扩展：核心层（hasher 增加 favicon 发现、ctx 支持、Reader 输入）→ 格式层（新增 4 种搜索引擎格式 + 指纹库）→ 接口层（CLI 批量/stdin/文件输出、API CORS/并发、MCP 补全工具）。每层向下依赖，可独立验证。

**Tech Stack:** Go 1.21, net/http, golang.org/x/net/html (HTML 解析), sync/atomic, existing murmur3/cobra

**Risks:**
- Task 2 新增 `golang.org/x/net/html` 依赖，需 `go get` → 缓解：这是 Go 官方扩展库，维护稳定
- Task 3 修改 `HashOptions` 增加 `MaxIconSize` 字段，可能影响现有调用方 → 缓解：有默认值，零值兼容
- Task 6 指纹数据库需要持续维护 → 缓解：先内置 50+ 常见指纹，提供外部 JSON 加载接口

---

### Task 1: 添加 context.Context 支持 — 使所有 hasher 方法支持取消和超时

**Depends on:** None
**Files:**
- Modify: `pkg/hasher/hasher.go:1-180`（所有公共方法签名 + 内部 HTTP 调用）
- Modify: `pkg/hasher/hasher_test.go`（更新测试签名）
- Modify: `pkg/api/server.go`（调用处传入 r.Context()）
- Modify: `pkg/mcp/handler.go`（processURL/processBase64 传入 context）

- [ ] **Step 1: 修改 IconHasher 公共方法签名，添加 ctx 参数**

修改 `pkg/hasher/hasher.go` 中所有公共方法，第一个参数改为 `ctx context.Context`：

```go
// HashFromURL downloads and calculates the hash of an icon from a URL
func (h *IconHasher) HashFromURL(ctx context.Context, url string) (string, error) {
	data, err := h.getContentFromURL(ctx, url)
	if err != nil {
		return "", fmt.Errorf("failed to get content from URL: %w", err)
	}
	return h.HashFromBytes(data)
}

// HashFromURLWithOption downloads and calculates the hash with per-call uint32 override
func (h *IconHasher) HashFromURLWithOption(ctx context.Context, url string, useUint32 bool) (string, error) {
	data, err := h.getContentFromURL(ctx, url)
	if err != nil {
		return "", fmt.Errorf("failed to get content from URL: %w", err)
	}
	return h.hashFromBytesWithOption(data, useUint32)
}
```

同样修改 `HashFromFile`、`HashFromBase64`、`HashFromBase64WithOption`、`HashFromBytes`、`HashFromBytesWithOption` — 非 URL 方法传入 `ctx` 但暂不使用（为 API 一致性预留）。

- [ ] **Step 2: 修改 getContentFromURL 使用 ctx 控制 HTTP 请求**

```go
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
```

- [ ] **Step 3: 在 HashOptions 中添加 MaxIconSize 字段**

```go
type HashOptions struct {
	UseUint32          bool
	RequestTimeout     time.Duration
	InsecureSkipVerify bool
	UserAgent          string
	HTTPClient         *http.Client
	MaxIconSize        int64 // 最大 icon 文件大小（字节），默认 10MB
}
```

修改 `DefaultOptions()` 添加 `MaxIconSize: 10 << 20`。

- [ ] **Step 4: 更新 pkg/api/server.go 中的调用**

所有 `s.iconHasher.HashFromURL(...)` 改为 `s.iconHasher.HashFromURL(r.Context(), ...)`，同样处理所有 `*WithOption` 调用。

- [ ] **Step 5: 更新 pkg/mcp/handler.go 中的调用**

所有 `h.iconHasher.HashFromURL(...)` 改为 `h.iconHasher.HashFromURL(context.Background(), ...)`，import `context`。

- [ ] **Step 6: 更新 cmd/ 中的 CLI 调用**

`cmd/url.go`、`cmd/file.go`、`cmd/base64.go` 中的 `h.HashFromURL` 等调用传入 `context.Background()`。

- [ ] **Step 7: 更新测试并验证**
Run: `go test ./... -count=1`
Expected:
  - Exit code: 0
  - Output contains: "ok" for all packages

- [ ] **Step 8: 提交**
Run: `git add -A && git commit -m "feat(hasher): add context.Context support to all public methods"`

---

### Task 2: 实现 Favicon URL 自动发现 — 从域名自动查找 favicon

**Depends on:** Task 1
**Files:**
- Create: `pkg/hasher/discover.go`
- Create: `pkg/hasher/discover_test.go`
- Modify: `pkg/hasher/hasher.go`（添加 DiscoverFavicon 方法）
- Modify: `pkg/api/server.go`（添加 /hash/discover 端点）

- [ ] **Step 1: 安装 golang.org/x/net 依赖**
Run: `go get golang.org/x/net/html`
Expected:
  - Exit code: 0
  - go.mod contains "golang.org/x/net"

- [ ] **Step 2: 创建 discover.go — favicon URL 自动发现逻辑**

```go
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
	// TryCommonPaths 是否尝试常见 favicon 路径（/favicon.ico 等）
	TryCommonPaths bool
}

// DefaultDiscoverOptions returns default discovery options
func DefaultDiscoverOptions() *DiscoverFaviconOptions {
	return &DiscoverFaviconOptions{
		TryCommonPaths: true,
	}
}

// commonFaviconPaths 是常见的 favicon 路径列表
var commonFaviconPaths = []string{
	"/favicon.ico",
	"/favicon.png",
	"/favicon-32x32.png",
	"/favicon-16x16.png",
	"/apple-touch-icon.png",
	"/mstile-144x144.png",
}

// DiscoverFavicon 从给定域名自动发现 favicon URL
// 策略: 1) 解析 HTML 中的 <link rel="icon"> 标签 2) 尝试常见路径
func (h *IconHasher) DiscoverFavicon(ctx context.Context, siteURL string, opts *DiscoverFaviconOptions) ([]string, error) {
	if opts == nil {
		opts = DefaultDiscoverOptions()
	}

	parsedURL, err := url.Parse(siteURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	// 确保 scheme 存在
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}
	baseURL := parsedURL.Scheme + "://" + parsedURL.Host

	var found []string
	seen := make(map[string]bool)

	// 策略 1: 解析 HTML 中的 favicon 链接
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

	// 策略 2: 尝试常见路径
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

// extractFaviconLinks 从 HTML 中提取 favicon link 标签的 href
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

// parseFaviconLinks 从 HTML token stream 中解析 favicon link 标签
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

// resolveURL 将相对 URL 解析为绝对 URL
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
```

- [ ] **Step 3: 创建 discover_test.go**

```go
package hasher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDiscoverFavicon_HTML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<html><head><link rel="icon" href="/custom-icon.png"><link rel="shortcut icon" href="/favicon.ico"></head><body></body></html>`))
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
	// 即使 HTML 没有 favicon 链接，也应返回常见路径
	if len(urls) == 0 {
		t.Error("Expected common path URLs")
	}
}

func TestParseFaviconLinks(t *testing.T) {
	html := `<html><head>
		<link rel="icon" type="image/png" href="/favicon.png">
		<link rel="shortcut icon" href="/old-favicon.ico">
		<link rel="stylesheet" href="/style.css">
		<link rel="apple-touch-icon" href="/apple.png">
	</head></html>`
	links, err := parseFaviconLinks(strings.NewReader(html))
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
```

注意：`discover_test.go` 中需 `import "strings"`。

- [ ] **Step 4: 添加 /hash/discover API 端点**

在 `pkg/api/server.go` 的路由注册中添加 `mux.HandleFunc("/hash/discover", s.handleHashDiscover)`，并实现 handler：

```go
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

	type DiscoverResult struct {
		URL  string `json:"url"`
		Hash string `json:"hash,omitempty"`
		Err  string `json:"error,omitempty"`
	}
	type DiscoverResponse struct {
		SiteURL string            `json:"site_url"`
		Results []DiscoverResult `json:"results"`
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
```

- [ ] **Step 5: 验证 discover 功能**
Run: `go test ./pkg/hasher/ -count=1 -run TestDiscover -v`
Expected:
  - Exit code: 0
  - Output contains: "PASS"

- [ ] **Step 6: 提交**
Run: `git add -A && git commit -m "feat(hasher): add favicon URL auto-discovery from domain"`

---

### Task 3: 添加 HashFromReader 方法 — 支持 io.Reader 输入源

**Depends on:** Task 1
**Files:**
- Modify: `pkg/hasher/hasher.go`（添加 HashFromReader 方法）
- Modify: `pkg/hasher/hasher_test.go`（添加测试）

- [ ] **Step 1: 添加 HashFromReader 和 HashFromReaderWithOption 方法**

```go
// HashFromReader calculates the hash of an icon from an io.Reader.
// This is the standard Go idiom for streaming input sources (stdin, pipes, network streams).
func (h *IconHasher) HashFromReader(ctx context.Context, r io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(r, h.options.MaxIconSize))
	if err != nil {
		return "", fmt.Errorf("failed to read from reader: %w", err)
	}
	return h.HashFromBytes(data)
}

// HashFromReaderWithOption calculates the hash from a reader with per-call uint32 override
func (h *IconHasher) HashFromReaderWithOption(ctx context.Context, r io.Reader, useUint32 bool) (string, error) {
	data, err := io.ReadAll(io.LimitReader(r, h.options.MaxIconSize))
	if err != nil {
		return "", fmt.Errorf("failed to read from reader: %w", err)
	}
	return h.HashFromBytesWithOption(data)
}
```

注意：`HashFromReaderWithOption` 内部需调用 `hashFromBytesWithOption` 而非 `HashFromBytesWithOption`，并传入 `useUint32`。

- [ ] **Step 2: 添加测试**

```go
func TestHashFromReader(t *testing.T) {
	data := []byte{0, 0, 1, 0, 1, 0, 16, 16}
	hasher := New(nil)

	hash1, err := hasher.HashFromReader(context.Background(), bytes.NewReader(data))
	if err != nil {
		t.Errorf("HashFromReader() error: %v", err)
	}

	hash2, err := hasher.HashFromBytes(data)
	if err != nil {
		t.Errorf("HashFromBytes() error: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("HashFromReader and HashFromBytes should produce same result: %s vs %s", hash1, hash2)
	}
}
```

- [ ] **Step 3: 验证**
Run: `go test ./pkg/hasher/ -count=1 -run TestHashFromReader -v`
Expected:
  - Exit code: 0
  - Output contains: "PASS"

- [ ] **Step 4: 提交**
Run: `git add -A && git commit -m "feat(hasher): add HashFromReader for io.Reader input support"`

---

### Task 4: 扩展搜索引擎格式 — 支持 Censys、Quake、ZoomEye、Hunter

**Depends on:** None
**Files:**
- Modify: `pkg/util/formatter.go:1-27`（添加新格式常量和格式化逻辑）
- Modify: `pkg/util/formatter_test.go`（添加测试）
- Modify: `pkg/api/server.go`（parseFormatParam 添加新格式）

- [ ] **Step 1: 扩展 OutputFormat 枚举和 FormatHash 函数**

```go
const (
	FormatPlain   OutputFormat = iota
	FormatFofa
	FormatShodan
	FormatCensys
	FormatQuake
	FormatZoomEye
	FormatHunter
)

func FormatHash(hash string, format OutputFormat) string {
	switch format {
	case FormatFofa:
		return fmt.Sprintf("icon_hash=\"%s\"", hash)
	case FormatShodan:
		return fmt.Sprintf("http.favicon.hash:%s", hash)
	case FormatCensys:
		return fmt.Sprintf("services.http.response.favicons.md5_hash:%s", hash)
	case FormatQuake:
		return fmt.Sprintf("favicon.hash:\"%s\"", hash)
	case FormatZoomEye:
		return fmt.Sprintf("iconhash:\"%s\"", hash)
	case FormatHunter:
		return fmt.Sprintf("web.icon=\"%s\"", hash)
	default:
		return hash
	}
}
```

- [ ] **Step 2: 更新 API server 的 parseFormatParam 和 getFormatName**

在 `pkg/api/server.go` 的 `parseFormatParam` switch 中添加：
```go
case "censys":
	return util.FormatCensys
case "quake":
	return util.FormatQuake
case "zoomeye":
	return util.FormatZoomEye
case "hunter":
	return util.FormatHunter
```

同样更新 `getFormatName`。

- [ ] **Step 3: 添加测试**

```go
func TestFormatHash_ExtendedFormats(t *testing.T) {
	tests := []struct {
		name     string
		hash     string
		format   OutputFormat
		expected string
	}{
		{"Censys format", "-12345", FormatCensys, "services.http.response.favicons.md5_hash:-12345"},
		{"Quake format", "-12345", FormatQuake, "favicon.hash:\"-12345\""},
		{"ZoomEye format", "-12345", FormatZoomEye, "iconhash:\"-12345\""},
		{"Hunter format", "-12345", FormatHunter, "web.icon=\"-12345\""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := FormatHash(test.hash, test.format)
			if result != test.expected {
				t.Errorf("FormatHash(%q, %v) = %q, expected %q", test.hash, test.format, result, test.expected)
			}
		})
	}
}
```

- [ ] **Step 4: 验证**
Run: `go test ./... -count=1`
Expected:
  - Exit code: 0

- [ ] **Step 5: 提交**
Run: `git add -A && git commit -m "feat(util): add Censys/Quake/ZoomEye/Hunter search engine formats"`

---

### Task 5: 添加 Favicon 指纹数据库 — hash 到服务名映射

**Depends on:** Task 1
**Files:**
- Create: `pkg/fingerprint/fingerprint.go`
- Create: `pkg/fingerprint/fingerprint_test.go`
- Create: `pkg/fingerprint/data.go`（内嵌指纹数据）
- Modify: `pkg/api/server.go`（添加 /lookup 端点）

- [ ] **Step 1: 创建 fingerprint 数据文件**

```go
package fingerprint

// FingerprintEntry maps a favicon hash to a known service/application
type FingerprintEntry struct {
	Hash    string `json:"hash"`
	Service string `json:"service"`
	Category string `json:"category,omitempty"`
}

// DefaultFingerprints contains well-known favicon hash to service mappings
var DefaultFingerprints = []FingerprintEntry{
	{Hash: "-305179312", Service: "Atlassian Confluence", Category: "CMS"},
	{Hash: "81586312", Service: "Jenkins", Category: "CI/CD"},
	{Hash: "-1275226814", Service: "XAMPP", Category: "Server"},
	{Hash: "-2009722838", Service: "React", Category: "Framework"},
	{Hash: "99395752", Service: "Slack", Category: "Communication"},
	{Hash: "1524374260", Service: "WordPress", Category: "CMS"},
	{Hash: "603314722", Service: "Nginx", Category: "Server"},
	{Hash: "-208326418", Service: "Apache", Category: "Server"},
	{Hash: "903341690", Service: "IIS", Category: "Server"},
	{Hash: "203679399", Service: "GitLab", Category: "VCS"},
	{Hash: "1387751544", Service: "JIRA", Category: "Project Management"},
	{Hash: "116323821", Service: "Spring Boot", Category: "Framework"},
	{Hash: "1500566322", Service: "cPanel", Category: "Hosting"},
	{Hash: "19487.Dialogue", Service: "phpMyAdmin", Category: "Database"},
	{Hash: "-532394952", Service: "Tomcat", Category: "Server"},
	{Hash: "1500566322", Service: "Plesk", Category: "Hosting"},
	{Hash: "-693082529", Service: "Django", Category: "Framework"},
	{Hash: "1540328751", Service: "Grafana", Category: "Monitoring"},
	{Hash: "1118684092", Service: "Prometheus", Category: "Monitoring"},
	{Hash: "1163734395", Service: "Elasticsearch", Category: "Search"},
	{Hash: "1526170209", Service: "Kibana", Category: "Search"},
	{Hash: "1457868128", Service: "Zabbix", Category: "Monitoring"},
	{Hash: "1010493916", Service: "RabbitMQ", Category: "Message Queue"},
	{Hash: "-842192628", Service: "MinIO", Category: "Storage"},
	{Hash: "1585996201", Service: "Kubernetes", Category: "Container"},
	{Hash: "-631094502", Service: "Docker Registry", Category: "Container"},
	{Hash: "-124318229", Service: "Portainer", Category: "Container"},
	{Hash: "-1591257963", Service: "Redmine", Category: "Project Management"},
	{Hash: "163323586", Service: "Moodle", Category: "LMS"},
	{Hash: "-655013097", Service: "Drupal", Category: "CMS"},
	{Hash: "829921184", Service: "Joomla", Category: "CMS"},
	{Hash: "1262005928", Service: "OpenCart", Category: "E-Commerce"},
	{Hash: "1507098592", Service: "Magento", Category: "E-Commerce"},
	{Hash: "766842226", Service: "Shopify", Category: "E-Commerce"},
	{Hash: "-466288462", Service: "Bitbucket", Category: "VCS"},
	{Hash: "1351214924", Service: "SonarQube", Category: "Code Quality"},
	{Hash: "1086177704", Service: "Nexus", Category: "Repository"},
	{Hash: "-254215620", Service: "Artifactory", Category: "Repository"},
	{Hash: "1632760989", Service: "Vaultwarden", Category: "Security"},
	{Hash: "-831781525", Service: "Nextcloud", Category: "Storage"},
	{Hash: "175178466", Service: "Ghost", Category: "CMS"},
	{Hash: "-923690202", Service: "Hadoop", Category: "Big Data"},
	{Hash: "1886732348", Service: "Spark", Category: "Big Data"},
	{Hash: "-1347531381", Service: "Airflow", Category: "Workflow"},
	{Hash: "981670006", Service: "Jupyter", Category: "Data Science"},
	{Hash: "628535358", Service: "RStudio", Category: "Data Science"},
	{Hash: "2006734119", Service: "Metabase", Category: "BI"},
	{Hash: "1594377221", Service: "Superset", Category: "BI"},
	{Hash: "552519518", Service: "Argo CD", Category: "GitOps"},
}
```

- [ ] **Step 2: 创建 fingerprint 查询引擎**

```go
package fingerprint

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// DB provides fingerprint lookup capabilities
type DB struct {
	mu    sync.RWMutex
	byHash map[string][]FingerprintEntry
}

// NewDB creates a fingerprint database from the given entries
func NewDB(entries []FingerprintEntry) *DB {
	db := &DB{
		byHash: make(map[string][]FingerprintEntry),
	}
	for _, e := range entries {
		db.byHash[e.Hash] = append(db.byHash[e.Hash], e)
	}
	return db
}

// DefaultDB returns a DB with built-in fingerprints
func DefaultDB() *DB {
	return NewDB(DefaultFingerprints)
}

// Lookup finds services matching the given hash
func (db *DB) Lookup(hash string) []FingerprintEntry {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.byHash[hash]
}

// LoadFromJSON loads fingerprints from a JSON file
func (db *DB) LoadFromJSON(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read fingerprint file: %w", err)
	}
	var entries []FingerprintEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("failed to parse fingerprint JSON: %w", err)
	}
	db.mu.Lock()
	for _, e := range entries {
		db.byHash[e.Hash] = append(db.byHash[e.Hash], e)
	}
	db.mu.Unlock()
	return nil
}

// All returns all fingerprint entries
func (db *DB) All() []FingerprintEntry {
	db.mu.RLock()
	defer db.mu.RUnlock()
	var all []FingerprintEntry
	for _, entries := range db.byHash {
		all = append(all, entries...)
	}
	return all
}
```

- [ ] **Step 3: 创建 fingerprint 测试**

```go
package fingerprint

import "testing"

func TestDefaultDB(t *testing.T) {
	db := DefaultDB()
	if db == nil {
		t.Fatal("DefaultDB() returned nil")
	}

	results := db.Lookup("81586312")
	if len(results) == 0 {
		t.Error("Expected to find Jenkins fingerprint")
	}
	if results[0].Service != "Jenkins" {
		t.Errorf("Expected Jenkins, got %s", results[0].Service)
	}
}

func TestLookupNotFound(t *testing.T) {
	db := DefaultDB()
	results := db.Lookup("999999999999")
	if len(results) != 0 {
		t.Error("Expected empty results for unknown hash")
	}
}

func TestAll(t *testing.T) {
	db := DefaultDB()
	all := db.All()
	if len(all) < 40 {
		t.Errorf("Expected at least 40 fingerprints, got %d", len(all))
	}
}
```

- [ ] **Step 4: 添加 /lookup API 端点**

在 `pkg/api/server.go` 添加路由 `mux.HandleFunc("/lookup", s.handleLookup)`：

```go
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

	db := fingerprint.DefaultDB()
	results := db.Lookup(hash)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"hash":    hash,
		"matches": results,
	})
}
```

添加 import `"github.com/cyberspacesec/iconhash-skills/pkg/fingerprint"`。

- [ ] **Step 5: 验证**
Run: `go test ./... -count=1`
Expected:
  - Exit code: 0

- [ ] **Step 6: 提交**
Run: `git add -A && git commit -m "feat(fingerprint): add favicon hash fingerprint database with lookup API"`

---

### Task 6: CLI 增强 — stdin 支持、批量文件输入、文件输出、代理标志

**Depends on:** Task 1
**Files:**
- Create: `cmd/batch.go`
- Modify: `cmd/root.go`（添加 stdin 检测、--proxy 标志、--output 标志）
- Modify: `cmd/common.go`（添加新全局变量）

- [ ] **Step 1: 添加新的全局变量到 common.go**

```go
var (
	Proxy     string
	OutputFile string
	InputFile  string
)
```

- [ ] **Step 2: 在 root.go 的 Initialize 中注册新 flag**

```go
RootCmd.PersistentFlags().StringVar(&Proxy, "proxy", "", "HTTP/SOCKS5 proxy URL (e.g. socks5://127.0.0.1:1080)")
RootCmd.PersistentFlags().StringVarP(&OutputFile, "output", "o", "", "Output file path (supports .json and .csv)")
RootCmd.PersistentFlags().StringVarP(&InputFile, "input", "i", "", "Input file with URLs (one per line)")
```

修改 `RootCmd.Run` 在开头添加 stdin 检测：

```go
// Check for stdin input
if len(args) == 0 && FilePath == "" && URL == "" && Base64Path == "" && InputFile == "" {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Data is being piped to stdin
		runBatchFromStdin(cmd, args)
		return
	}
	cmd.Help()
	return
}
```

添加 InputFile 处理：

```go
if InputFile != "" {
	runBatchFromFile(cmd, args)
	return
}
```

- [ ] **Step 3: 创建 batch.go — 批量处理命令**

```go
package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cyberspacesec/iconhash-skills/pkg/hasher"
	"github.com/cyberspacesec/iconhash-skills/pkg/util"
	"github.com/fatih/color"
)

type batchResult struct {
	URL  string `json:"url"`
	Hash string `json:"hash,omitempty"`
	Err  string `json:"error,omitempty"`
}

func runBatchFromStdin(cmd *cobra.Command, args []string) {
	runBatchFromReader(cmd, os.Stdin)
}

func runBatchFromFile(cmd *cobra.Command, args []string) {
	f, err := os.Open(InputFile)
	if err != nil {
		color.Red("❌ Error opening input file: %v", err)
		os.Exit(1)
	}
	defer f.Close()
	runBatchFromReader(cmd, f)
}

func runBatchFromReader(cmd *cobra.Command, r io.Reader) {
	options := buildHashOptions()
	h := hasher.New(options)

	var format util.OutputFormat
	if ShodanFormat {
		format = util.FormatShodan
	} else if FofaFormat {
		format = util.FormatFofa
	} else {
		format = util.FormatPlain
	}

	var results []batchResult
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		hash, err := h.HashFromURL(context.Background(), line)
		if err != nil {
			results = append(results, batchResult{URL: line, Err: err.Error()})
		} else {
			results = append(results, batchResult{URL: line, Hash: util.FormatHash(hash, format)})
		}
	}

	if OutputFile != "" {
		writeBatchOutput(results)
	} else {
		for _, r := range results {
			if r.Err != "" {
				fmt.Fprintf(os.Stderr, "❌ %s: %s\n", r.URL, r.Err)
			} else {
				fmt.Printf("%s %s\n", r.URL, r.Hash)
			}
		}
	}
}

func buildHashOptions() *hasher.HashOptions {
	opts := &hasher.HashOptions{
		UseUint32:          Uint32Flag,
		RequestTimeout:     Timeout,
		InsecureSkipVerify: SkipVerify,
		UserAgent:          UserAgent,
	}
	if Proxy != "" {
		proxyURL, err := url.Parse(Proxy)
		if err == nil {
			opts.HTTPClient = &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
				},
				Timeout: Timeout,
			}
		}
	}
	return opts
}

func writeBatchOutput(results []batchResult) {
	if strings.HasSuffix(OutputFile, ".csv") {
		f, err := os.Create(OutputFile)
		if err != nil {
			color.Red("❌ Error creating output file: %v", err)
			return
		}
		defer f.Close()
		f.WriteString("url,hash,error\n")
		for _, r := range results {
			fmt.Fprintf(f, "%q,%q,%q\n", r.URL, r.Hash, r.Err)
		}
	} else {
		data, _ := json.MarshalIndent(results, "", "  ")
		os.WriteFile(OutputFile, data, 0644)
	}
}
```

注意：`batch.go` 需要 import `context`, `io`, `net/http`, `net/url`。

- [ ] **Step 4: 验证**
Run: `go build ./...`
Expected:
  - Exit code: 0

- [ ] **Step 5: 提交**
Run: `git add -A && git commit -m "feat(cli): add stdin/batch/file input, proxy flag, and file output support"`

---

### Task 7: API Server 增强 — CORS、Graceful Shutdown、并发 Batch

**Depends on:** Task 1
**Files:**
- Modify: `pkg/api/server.go`（CORS 中间件、graceful shutdown、并发 batch）

- [ ] **Step 1: 添加 CORS 中间件**

```go
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
```

在 `Start()` 中将 CORS middleware 放在 auth middleware 之前：

```go
handler := corsMiddleware(mux)
if s.config.AuthToken != "" {
	handler = s.authMiddleware(handler)
}
```

- [ ] **Step 2: 替换 os.Exit 为 Graceful Shutdown**

修改 `Start()` 方法返回 `http.Server` 引用并支持 Shutdown：

```go
func (s *Server) Start() error {
	mux := http.NewServeMux()
	// ... route registration ...

	handler := corsMiddleware(mux)
	if s.config.AuthToken != "" {
		handler = s.authMiddleware(handler)
	}

	addr := s.config.Host + ":" + strconv.Itoa(s.config.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	// Channel for server errors
	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.ListenAndServe()
	}()

	// Wait for error or context cancellation
	select {
	case err := <-errChan:
		if err != http.ErrServerClosed {
			return err
		}
		return nil
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	// srv needs to be stored on the Server struct
	return s.server.Shutdown(ctx)
}
```

需要在 `Server` struct 中添加 `server *http.Server` 字段。

- [ ] **Step 3: 修改 batch 端点使用并发处理**

```go
func (s *Server) handleHashBatch(w http.ResponseWriter, r *http.Request) {
	// ... existing parsing code ...

	// Concurrent processing with worker pool
	type batchResult struct {
		index int
		item  BatchHashItem
	}
	
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
```

添加 import `"sync"`。

- [ ] **Step 4: 验证**
Run: `go build ./... && go test ./... -count=1`
Expected:
  - Exit code: 0

- [ ] **Step 5: 提交**
Run: `git add -A && git commit -m "feat(api): add CORS, graceful shutdown, and concurrent batch processing"`

---

### Task 8: MCP 层补全 — 添加 iconhash_file、iconhash_discover、iconhash_lookup 工具

**Depends on:** Task 2, Task 5
**Files:**
- Modify: `pkg/mcp/handler.go`（扩展 Tools 列表和 CallTool 路由）

- [ ] **Step 1: 在 Tools() 中添加 iconhash_file、iconhash_discover、iconhash_lookup 工具定义**

```go
{
	Name:        "iconhash_file",
	Description: "Calculate the MMH3 favicon hash from a file path on the server.",
	InputSchema: struct {
		Type       string                 `json:"type"`
		Properties map[string]interface{} `json:"properties"`
		Required   []string               `json:"required,omitempty"`
	}{
		Type: "object",
		Properties: map[string]interface{}{
			"path":   map[string]string{"type": "string", "description": "Absolute path to the favicon file"},
			"format": map[string]string{"type": "string", "description": "Output format: plain, fofa, shodan (default: fofa)"},
			"uint32": map[string]string{"type": "boolean", "description": "Use uint32 format (default: false)"},
		},
		Required: []string{"path"},
	},
},
{
	Name:        "iconhash_discover",
	Description: "Discover favicon URLs from a domain and calculate their hashes.",
	InputSchema: struct {
		Type       string                 `json:"type"`
		Properties map[string]interface{} `json:"properties"`
		Required   []string               `json:"required,omitempty"`
	}{
		Type: "object",
		Properties: map[string]interface{}{
			"url":    map[string]string{"type": "string", "description": "Domain or URL to discover favicons from"},
			"format": map[string]string{"type": "string", "description": "Output format: plain, fofa, shodan (default: fofa)"},
			"uint32": map[string]string{"type": "boolean", "description": "Use uint32 format (default: false)"},
		},
		Required: []string{"url"},
	},
},
{
	Name:        "iconhash_lookup",
	Description: "Look up a favicon hash in the fingerprint database to identify the service.",
	InputSchema: struct {
		Type       string                 `json:"type"`
		Properties map[string]interface{} `json:"properties"`
		Required   []string               `json:"required,omitempty"`
	}{
		Type: "object",
		Properties: map[string]interface{}{
			"hash": map[string]string{"type": "string", "description": "Favicon hash value to look up"},
		},
		Required: []string{"hash"},
	},
},
```

- [ ] **Step 2: 在 CallTool 中添加新工具的处理分支**

```go
case "iconhash_file":
	return h.callToolFile(args)
case "iconhash_discover":
	return h.callToolDiscover(args)
case "iconhash_lookup":
	return h.callToolLookup(args)
```

实现三个新方法：

```go
func (h *Handler) callToolFile(args map[string]interface{}) *ToolResult {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return &ToolResult{Content: []ToolContent{{Type: "text", Text: "Error: 'path' parameter is required"}}, IsError: true}
	}
	useUint32 := boolArg(args, "uint32")
	hash, err := h.iconHasher.HashFromFile(path)
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
```

添加 import `"context"` 和 `"github.com/cyberspacesec/iconhash-skills/pkg/fingerprint"`。

- [ ] **Step 3: 验证**
Run: `go test ./... -count=1`
Expected:
  - Exit code: 0

- [ ] **Step 4: 提交**
Run: `git add -A && git commit -m "feat(mcp): add iconhash_file, iconhash_discover, iconhash_lookup tools"`
