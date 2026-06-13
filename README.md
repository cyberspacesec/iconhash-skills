# iconhash-skills

A powerful CLI tool and Go SDK for calculating favicon hashes used in cybersecurity reconnaissance. This tool is an improved implementation of [Becivells/iconhash](https://github.com/Becivells/iconhash).

## Features

- Calculate MMH3 (MurmurHash3) hash of favicons
- Multiple input sources: URL, local file, base64 data, stdin
- 6 search engine formats: Fofa, Shodan, Censys, Quake, ZoomEye, Hunter
- Fingerprint database with 700+ known services for identification
- Batch processing with concurrent workers
- HTTP API server with authentication
- Model Context Protocol (MCP) support for AI integration
- Full Go SDK for embedding in other tools
- Docker support for containerized usage
- Two build variants: Lite (small) and Full (offline-capable)

## Installation

### Download Pre-built Binary (Recommended)

Download from [GitHub Releases](https://github.com/cyberspacesec/iconhash-skills/releases/latest) for your platform.

**Two build variants:**
- **Lite** (`iconhash_lite_*`): Smaller binary, fingerprints auto-downloaded on first use
- **Full** (`iconhash_full_*`): Larger binary with embedded fingerprints, works offline

**Quick install (Linux/macOS):**
```bash
# Linux x86_64 (Lite)
wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_linux_x86_64.tar.gz
tar xzf iconhash_lite_linux_x86_64.tar.gz
chmod +x iconhash && sudo mv iconhash /usr/local/bin/

# macOS Apple Silicon (Lite)
wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_macos_aarch64.tar.gz
tar xzf iconhash_lite_macos_aarch64.tar.gz
chmod +x iconhash && sudo mv iconhash /usr/local/bin/

# Windows (download and extract zip)
# https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_windows_x86_64.zip
```

**Supported platforms:** Linux (x86_64, aarch64, i386, arm, riscv64), macOS (x86_64, aarch64), Windows (x86_64, aarch64, i386), FreeBSD (x86_64, aarch64)

### Install via Go

```bash
go install github.com/cyberspacesec/iconhash-skills/cmd/iconhash@latest
```

### Build from Source

```bash
git clone https://github.com/cyberspacesec/iconhash-skills.git
cd iconhash-skills

# Lite build (default, no embedded fingerprints)
make build

# Full build (with embedded fingerprints, works offline)
make build-full

# Install to $GOPATH/bin
make install
```

### Docker

```bash
docker pull cyberspacesec/iconhash:latest
docker run --rm cyberspacesec/iconhash:latest url https://example.com/favicon.ico
```

## Quick Start

```bash
# Hash from URL
iconhash url https://www.example.com/favicon.ico

# Identify a website (discover + hash + fingerprint lookup)
iconhash identify https://example.com

# Lookup a hash in the fingerprint database
iconhash lookup -- -305179312

# Batch process
iconhash batch -i urls.txt -o results.json

# Start API server
iconhash server -p 8080
```

## CLI Commands

| Command | Purpose |
|---|---|
| `iconhash url <url>` | Calculate hash from a URL |
| `iconhash file <path>` | Calculate hash from a file |
| `iconhash base64 <path>` | Calculate hash from base64 file |
| `iconhash discover <url>` | Discover favicons on a site and hash them |
| `iconhash identify <url>` | Discover + fingerprint identification |
| `iconhash lookup <hash>` | Lookup hash in fingerprint database |
| `iconhash batch` | Batch process URLs from file/stdin |
| `iconhash fingerprints` | Browse/search fingerprint database |
| `iconhash fingerprints update` | Update fingerprint database |
| `iconhash server` | Start HTTP API server |

### Global Flags

| Flag | Short | Description |
|---|---|---|
| `--engine` | `-e` | Format: `plain`, `fofa`, `shodan`, `censys`, `quake`, `zoomeye`, `hunter` |
| `--uint32` | `-n` | Output uint32 instead of int32 |
| `--insecure` | `-k` | Skip TLS verification |
| `--timeout` | `-t` | HTTP timeout (default 30s) |
| `--proxy` | | HTTP/SOCKS5 proxy URL |
| `--debug` | `-d` | Enable debug output |
| `--format` | | Output format: `text`, `json`, `csv` |
| `--output` | `-o` | Output file path |
| `--fingerprint-db` | | Custom fingerprint JSON database |

See [SKILLS.md](./SKILLS.md) for complete documentation of every command and parameter.

## Go SDK

Use iconhash-skills as a Go library in your projects:

```bash
go get github.com/cyberspacesec/iconhash-skills
```

### Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/cyberspacesec/iconhash-skills/pkg/hasher"
    "github.com/cyberspacesec/iconhash-skills/pkg/fingerprint"
    "github.com/cyberspacesec/iconhash-skills/pkg/util"
)

func main() {
    // One-off hash with default options
    hash, _ := hasher.HashURL(context.Background(), "https://example.com/favicon.ico")
    fmt.Println("Hash:", hash)

    // Or create a custom hasher
    h := hasher.New(&hasher.HashOptions{
        InsecureSkipVerify: true,
        RequestTimeout:     30 * time.Second,
    })

    // Hash from various sources
    hash, _ = h.HashFromURL(ctx, "https://example.com/favicon.ico")
    hash, _ = h.HashFromFile("favicon.ico")
    hash, _ = h.HashFromBytes(rawBytes)
    hash, _ = h.HashFromBase64("base64data...")
    hash, _ = h.HashFromReader(reader)

    // Discover + hash
    results := h.DiscoverAndHash(ctx, "https://example.com", nil)

    // Full identification (discover + hash + fingerprint lookup)
    db := fingerprint.DefaultDB()
    results := h.Identify(ctx, "https://example.com", db, nil)

    // Batch processing
    urls := []string{"https://a.com/favicon.ico", "https://b.com/favicon.ico"}
    results := h.BatchHashURLs(ctx, urls, 10) // 10 concurrent workers

    // Format for search engines
    queries := util.FormatAll(hash) // map of all engines
    formatted := util.FormatHash(hash, util.FormatFofa)

    // Fingerprint database operations
    matches := db.Lookup(hash)
    entries := db.LookupByCategory("CMS")
    entries := db.LookupByTag("chinese")
    entries := db.Filter(fingerprint.FilterOptions{Category: "Server", Tag: "java"})
}
```

### With Proxy

```go
// Create options with proxy support
opts, err := hasher.NewOptionsWithProxy("socks5://127.0.0.1:1080", 30*time.Second, true)
h := hasher.New(opts)
```

### API Server SDK

```go
import "github.com/cyberspacesec/iconhash-skills/pkg/api"

config := api.DefaultConfig()
config.Port = 9090
config.AuthToken = "my-secret"
server := api.NewServer(config)
server.Start() // blocks until shutdown
```

### MCP Handler SDK

```go
import "github.com/cyberspacesec/iconhash-skills/pkg/mcp"

handler := mcp.NewHandler(false)
tools := handler.Tools() // List available tools
result := handler.CallTool("iconhash_url", map[string]interface{}{
    "url": "https://example.com/favicon.ico",
})
```

## API Server

```bash
# Start server
iconhash server -p 8080

# With authentication
iconhash server --auth-token secret123 -H 0.0.0.0
```

| Endpoint | Method | Description |
|---|---|---|
| `/health` | GET | Health check |
| `/hash/url` | GET, POST | Hash from URL |
| `/hash/file` | POST | Hash from uploaded file |
| `/hash/base64` | POST | Hash from base64 data |
| `/hash/batch` | POST | Batch hash URLs |
| `/hash/discover` | POST | Discover favicons |
| `/lookup` | GET | Fingerprint lookup |
| `/fingerprints` | GET | Fingerprint database |
| `/mcp` | POST | Model Context Protocol |

## Search Engine Formats

| Engine | Format | Example |
|---|---|---|
| Fofa | `icon_hash="<hash>"` | `icon_hash="-305179312"` |
| Shodan | `http.favicon.hash:<hash>` | `http.favicon.hash:-305179312` |
| Censys | `services.http.response.favicons.md5_hash:<hash>` | |
| Quake | `favicon.hash:"<hash>"` | |
| ZoomEye | `iconhash:"<hash>"` | |
| Hunter | `web.icon="<hash>"` | |

## Docker Usage

```bash
# CLI
docker run --rm cyberspacesec/iconhash:latest url https://example.com/favicon.ico

# API Server
docker run -d -p 8080:8080 cyberspacesec/iconhash:latest server -H 0.0.0.0 -p 8080

# With docker-compose
docker-compose up server
docker-compose run --rm cli url https://example.com/favicon.ico
```

## Development

```bash
make build          # Lite build
make build-full     # Full build (with fingerprints)
make test           # Run tests
make test-coverage  # Run tests with coverage report
make docker-build   # Build Docker image
```

## Use Cases

- **Cybersecurity Reconnaissance**: Search for websites with matching favicons on Fofa, Shodan, etc.
- **Asset Discovery**: Identify web services through their favicon fingerprints
- **Automation**: Batch process URLs, integrate via API or SDK
- **AI Integration**: Use MCP protocol for conversational interfaces

## License

[MIT License](LICENSE)
