# IconHash Skills Documentation

> **Progressive Disclosure Reference** — This document describes every capability of the `iconhash` CLI tool and Go SDK, organized so that AI Agents (and humans) can quickly find what they need.

---

## Quick Reference

| Command | Purpose | Example |
|---|---|---|
| `iconhash url` | Hash from URL | `iconhash url https://example.com/favicon.ico` |
| `iconhash file` | Hash from file | `iconhash file favicon.ico` |
| `iconhash base64` | Hash from base64 data | `iconhash base64 encoded.txt` |
| `iconhash discover` | Discover favicons on a site | `iconhash discover https://example.com` |
| `iconhash identify` | Discover + fingerprint lookup | `iconhash identify https://example.com` |
| `iconhash lookup` | Lookup hash in fingerprint DB | `iconhash lookup -- -305179312` |
| `iconhash batch` | Batch hash URLs from file/stdin | `iconhash batch -i urls.txt` |
| `iconhash fingerprints` | Browse/search fingerprint DB | `iconhash fingerprints --all` |
| `iconhash fingerprints update` | Update fingerprint DB | `iconhash fingerprints update` |
| `iconhash server` | Start HTTP API server | `iconhash server -p 8080` |

---

## Installation

### Download Pre-built Binary (Recommended)

Download the latest release for your platform from [GitHub Releases](https://github.com/cyberspacesec/iconhash-skills/releases/latest).

**Two build variants:**
- **Lite** (`iconhash_lite_*`): Smaller binary, fingerprints loaded from external file or auto-downloaded at first use
- **Full** (`iconhash_full_*`): Larger binary with 700+ embedded fingerprints, works offline

#### Linux (x86_64)

```bash
# Lite build
wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_linux_x86_64.tar.gz
tar xzf iconhash_lite_linux_x86_64.tar.gz
chmod +x iconhash
sudo mv iconhash /usr/local/bin/

# Full build (with embedded fingerprints, works offline)
wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_full_linux_x86_64.tar.gz
tar xzf iconhash_full_linux_x86_64.tar.gz
chmod +x iconhash-full
sudo mv iconhash-full /usr/local/bin/iconhash
```

#### Linux (ARM64 / aarch64)

```bash
wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_linux_aarch64.tar.gz
tar xzf iconhash_lite_linux_aarch64.tar.gz
chmod +x iconhash
sudo mv iconhash /usr/local/bin/
```

#### macOS (Apple Silicon)

```bash
wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_macos_aarch64.tar.gz
tar xzf iconhash_lite_macos_aarch64.tar.gz
chmod +x iconhash
sudo mv iconhash /usr/local/bin/
```

#### macOS (Intel)

```bash
wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_macos_x86_64.tar.gz
tar xzf iconhash_lite_macos_x86_64.tar.gz
chmod +x iconhash
sudo mv iconhash /usr/local/bin/
```

#### Windows (x86_64)

```powershell
# Download using PowerShell
Invoke-WebRequest -Uri "https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_windows_x86_64.zip" -OutFile "iconhash.zip"
Expand-Archive iconhash.zip
```

#### FreeBSD (x86_64)

```bash
wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_freebsd_x86_64.tar.gz
tar xzf iconhash_lite_freebsd_x86_64.tar.gz
chmod +x iconhash
sudo mv iconhash /usr/local/bin/
```

### Install via Go

```bash
go install github.com/cyberspacesec/iconhash-skills/cmd/iconhash@latest
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/cyberspacesec/iconhash-skills.git
cd iconhash-skills

# Lite build (default, no embedded fingerprints)
make build

# Full build (with embedded fingerprints)
make build-full

# Install to $GOPATH/bin
make install
```

### Docker

```bash
docker pull cyberspacesec/iconhash:latest
docker run --rm cyberspacesec/iconhash:latest url https://example.com/favicon.ico
```

---

## Global Flags

These flags are available for **all** subcommands:

| Flag | Short | Type | Default | Description |
|---|---|---|---|---|
| `--debug` | `-d` | bool | false | Enable debug output |
| `--uint32` | `-n` | bool | false | Output hash as uint32 instead of int32 |
| `--user-agent` | `-a` | string | (built-in) | Custom User-Agent for HTTP requests |
| `--engine` | `-e` | string | plain | Search engine format: `plain`, `fofa`, `shodan`, `censys`, `quake`, `zoomeye`, `hunter` |
| `--insecure` | `-k` | bool | false | Skip TLS certificate verification |
| `--timeout` | `-t` | duration | 30s | HTTP request timeout |
| `--format` | | string | text | Output format: `text`, `json`, `csv` |
| `--proxy` | | string | | HTTP/SOCKS5 proxy URL (e.g. `socks5://127.0.0.1:1080`) |
| `--output` | `-o` | string | | Output file path (supports `.json` and `.csv`) |
| `--input` | `-i` | string | | Input file with URLs (one per line) |
| `--fingerprint-db` | | string | | Path to custom fingerprint JSON database |

---

## Command Details

### `iconhash url` — Hash from URL

Calculate the MMH3 favicon hash from a URL.

```bash
iconhash url https://example.com/favicon.ico
iconhash url https://example.com --engine fofa
iconhash url https://example.com --uint32 --engine shodan
iconhash url https://example.com --proxy socks5://127.0.0.1:1080
```

**Positional argument:** `[url]` — The URL to the favicon (alternative: `--url` flag)

**Output:**
- Plain hash: `-305179312`
- Engine-formatted (e.g. `--engine fofa`): `icon_hash="-305179312"`

---

### `iconhash file` — Hash from File

Calculate the MMH3 favicon hash from a local file.

```bash
iconhash file favicon.ico
iconhash file /path/to/icon.png --engine shodan
iconhash file icon.ico --uint32
```

**Positional argument:** `[filepath]` — Path to the favicon file (alternative: `--file` flag)

---

### `iconhash base64` — Hash from Base64 Data

Calculate the MMH3 favicon hash from a file containing base64-encoded data.

```bash
iconhash base64 encoded.txt
iconhash base64 /path/to/b64.txt --engine fofa
```

**Positional argument:** `[filepath]` — Path to file with base64 data (alternative: `--b64` flag)

The base64 data can optionally include the `data:image/vnd.microsoft.icon;base64,` prefix — it will be stripped automatically.

---

### `iconhash discover` — Discover Favicons

Scan a website for favicon URLs and calculate hashes for each one found.

**Discovery strategy:**
1. Parse HTML `<link rel="icon">` tags
2. Try common favicon paths (`/favicon.ico`, `/favicon.png`, `/apple-touch-icon.png`, etc.)

```bash
iconhash discover https://example.com
iconhash discover https://example.com --engine fofa
iconhash discover https://example.com --insecure
```

**Positional argument:** `[url]` — The website URL to scan (alternative: `--url` flag)

**Output:** For each discovered favicon URL, shows the URL and its hash.

---

### `iconhash identify` — Full Identification

The all-in-one reconnaissance command. It:
1. Discovers all favicon URLs on the target website
2. Calculates the MMH3 hash for each favicon
3. Looks up each hash against the fingerprint database
4. Outputs search engine queries for all major platforms

```bash
iconhash identify https://example.com
iconhash identify https://gitlab.example.com --engine fofa
iconhash identify https://jenkins.example.com --fingerprint-db custom.json
```

**Positional argument:** `[url]` — The website URL to identify (alternative: `--url` flag)

**Output:**
- Discovered favicons with hashes
- Fingerprint matches (service name, category, version, tags)
- Search engine queries for Fofa, Shodan, Censys, Quake, ZoomEye, Hunter

---

### `iconhash lookup` — Fingerprint Lookup

Look up a favicon hash against the built-in fingerprint database to identify associated services.

```bash
# Note: Use "--" before negative hashes to prevent Cobra from interpreting them as flags
iconhash lookup -- -305179312
iconhash lookup 81586312
iconhash lookup -- -305179312 --fingerprint-db custom.json
```

**Positional argument:** `[hash]` — The MMH3 favicon hash to look up

**Important:** Negative hashes (starting with `-`) must be preceded by `--` to prevent Cobra from interpreting them as flags.

**Output:**
- Matching services (name, category, version, description, tags)
- Search engine queries for all platforms

---

### `iconhash batch` — Batch Processing

Process multiple URLs from a file or stdin.

```bash
# From file
iconhash batch -i urls.txt

# From stdin
cat urls.txt | iconhash batch

# With output file
iconhash batch -i urls.txt -o results.json

# With proxy and engine format
iconhash batch -i urls.txt --engine fofa --proxy socks5://127.0.0.1:1080

# CSV output
iconhash batch -i urls.txt -o results.csv
```

**Input format:** One URL per line. Lines starting with `#` are treated as comments and skipped. Blank lines are ignored.

**Output formats:**
- Default (stdout): `<url> <hash>` per line
- JSON (`-o results.json`): Array of `{url, hash, error}` objects
- CSV (`-o results.csv`): Comma-separated values

---

### `iconhash fingerprints` — Fingerprint Database

Browse, search, and export the built-in fingerprint database.

```bash
# Show database statistics
iconhash fingerprints

# List all fingerprints
iconhash fingerprints --all

# Filter by category
iconhash fingerprints --category CMS

# Filter by service name (partial match, case-insensitive)
iconhash fingerprints --service jenkins

# Filter by tag
iconhash fingerprints --tag chinese

# Combine filters
iconhash fingerprints --category Server --tag java

# JSON output
iconhash fingerprints --all --format json

# Export to file
iconhash fingerprints --all --export fingerprints.json
iconhash fingerprints --category CMS --export cms.csv
```

**Flags:**

| Flag | Type | Description |
|---|---|---|
| `--all` | bool | List all fingerprints (default: show stats only) |
| `--category` | string | Filter by category (CMS, Server, Framework, OA, etc.) |
| `--service` | string | Filter by service name (partial match, case-insensitive) |
| `--tag` | string | Filter by tag (e.g. `chinese`, `java`, `enterprise`) |
| `--export` | string | Export to file (supports `.json` and `.csv`) |

#### `iconhash fingerprints update` — Update Fingerprint Database

Download the latest community-maintained fingerprint database.

```bash
# Update from default source
iconhash fingerprints update

# Save to custom file
iconhash fingerprints update -o my_fingerprints.json

# Use custom URL
iconhash fingerprints update --url https://example.com/fingerprints.json

# Replace instead of merge
iconhash fingerprints update --replace
```

**Flags:**

| Flag | Short | Type | Default | Description |
|---|---|---|---|---|
| `--output` | `-o` | string | `~/.iconhash/fingerprints.json` | Output file path |
| `--url` | | string | GitHub source | Custom URL to download from |
| `--replace` | | bool | false | Replace existing DB instead of merging |

**Features:**
- Multiple download sources (GitHub + mirror) for resilience
- SHA256 verification of downloaded data
- Incremental merge with existing database (no duplicates)
- Version tracking for update checking

---

### `iconhash server` — HTTP API Server

Start a REST API server for favicon hash calculation.

```bash
# Start on default port
iconhash server

# Custom host and port
iconhash server -H 0.0.0.0 -p 3000

# With authentication
iconhash server --auth-token secret123

# With debug logging
iconhash server --debug

# With proxy for outbound requests
iconhash server --proxy socks5://127.0.0.1:1080
```

**Server-specific flags:**

| Flag | Short | Type | Default | Description |
|---|---|---|---|---|
| `--host` | `-H` | string | 127.0.0.1 | Host address to bind to |
| `--port` | `-p` | int | 8080 | Port to listen on |
| `--auth-token` | | string | | Authentication token (empty = no auth) |
| `--read-timeout` | | duration | 30s | HTTP server read timeout |
| `--write-timeout` | | duration | 30s | HTTP server write timeout |
| `--proxy` | | string | | HTTP/SOCKS5 proxy for outbound requests |

**API Endpoints:**

| Endpoint | Method | Description |
|---|---|---|
| `/health` | GET | Health check |
| `/hash/url` | GET, POST | Hash from URL |
| `/hash/file` | POST | Hash from uploaded file |
| `/hash/base64` | POST | Hash from base64 data |
| `/hash/batch` | POST | Batch hash URLs |
| `/hash/discover` | POST | Discover favicons and hash |
| `/lookup` | GET | Fingerprint lookup by hash |
| `/fingerprints` | GET | List/search fingerprints |
| `/mcp` | POST | Model Context Protocol endpoint |

**Query Parameters (for hash endpoints):**

| Parameter | Values | Description |
|---|---|---|
| `format` | `fofa`, `shodan`, `censys`, `quake`, `zoomeye`, `hunter`, `plain` | Output format (default: `fofa`) |
| `uint32` | `true`, `false` | Use uint32 format (default: `false`) |

**Authentication:** If `--auth-token` is set, include the token via:
- Header: `Authorization: Bearer <token>`
- Query: `?token=<token>`

---

## Search Engine Formats

IconHash supports 6 cyber-space mapping platforms:

| Engine | Format | Example |
|---|---|---|
| Fofa | `icon_hash="<hash>"` | `icon_hash="-305179312"` |
| Shodan | `http.favicon.hash:<hash>` | `http.favicon.hash:-305179312` |
| Censys | `services.http.response.favicons.md5_hash:<hash>` | `services.http.response.favicons.md5_hash:-305179312` |
| Quake | `favicon.hash:"<hash>"` | `favicon.hash:"-305179312"` |
| ZoomEye | `iconhash:"<hash>"` | `iconhash:"-305179312"` |
| Hunter | `web.icon="<hash>"` | `web.icon="-305179312"` |

---

## Fingerprint Database

The fingerprint database maps well-known MMH3 favicon hashes to services and applications, enabling identification of web services through their favicons.

**Categories include:** CMS, Server, Framework, OA, Router, Security, Mail, Database, Monitoring, Cloud, Wiki, ERP, and more.

**Configuration (for Lite builds):**
1. `--fingerprint-db` flag (highest priority)
2. `ICONHASH_FINGERPRINT_DB` environment variable
3. Embedded fingerprints (Full build only)
4. `~/.iconhash/fingerprints.json` (auto-downloaded)
5. Auto-download on first use

---

## Go SDK API

IconHash can be used as a Go library for integration into security tools, automation pipelines, and other Go projects.

### Installation

```bash
go get github.com/cyberspacesec/iconhash-skills
```

### Quick Start (Package-Level Functions)

For one-off operations with default options, use the package-level convenience functions:

```go
package main

import (
    "context"
    "fmt"
    "github.com/cyberspacesec/iconhash-skills/pkg/hasher"
)

func main() {
    // Hash from URL
    hash, err := hasher.HashURL(context.Background(), "https://example.com/favicon.ico")
    if err != nil {
        panic(err)
    }
    fmt.Println("Hash:", hash)

    // Hash from file
    hash, err = hasher.HashFile("favicon.ico")

    // Hash from bytes
    hash, err = hasher.HashBytes(rawBytes)

    // Hash from base64
    hash, err = hasher.HashBase64("AAABAAEAEBAAAAEAIABoBAAAFgAA...")

    // Hash from reader (stdin, pipes, etc.)
    hash, err = hasher.HashReader(os.Stdin)
}
```

### Custom Options

```go
options := &hasher.HashOptions{
    UseUint32:          false,
    RequestTimeout:     30 * time.Second,
    InsecureSkipVerify: true,
    UserAgent:          "MyApp/1.0",
    MaxIconSize:        10 << 20, // 10 MB
}
h := hasher.New(options)

// Or create with proxy
options, err := hasher.NewOptionsWithProxy("socks5://127.0.0.1:1080", 30*time.Second, true)
```

### Instance Methods

```go
h := hasher.New(options)

// Basic hashing
hash, err := h.HashFromURL(ctx, "https://example.com/favicon.ico")
hash, err := h.HashFromFile("favicon.ico")
hash, err := h.HashFromBase64("base64data...")
hash, err := h.HashFromBytes([]byte{...})
hash, err := h.HashFromReader(reader)

// Per-call uint32 override (thread-safe)
hash, err := h.HashFromURLWithOption(ctx, url, true) // uint32
hash, err := h.HashFromBytesWithOption(data, true)

// Discover + hash in one call
results := h.DiscoverAndHash(ctx, "https://example.com", nil)
for _, r := range results {
    fmt.Printf("%s: %s\n", r.URL, r.Hash)
}

// Full identification (discover + hash + fingerprint lookup)
db := fingerprint.DefaultDB()
results := h.Identify(ctx, "https://example.com", db, nil)
for _, r := range results {
    fmt.Printf("%s → %s\n", r.URL, r.Hash)
    for _, m := range r.Matches {
        fmt.Printf("  Match: %s [%s]\n", m.Service, m.Category)
    }
}

// Batch processing
urls := []string{"https://a.com/favicon.ico", "https://b.com/favicon.ico"}
results := h.BatchHashURLs(ctx, urls, 10) // 10 concurrent workers
for _, r := range results {
    fmt.Printf("%s: %s\n", r.URL, r.Hash)
}

// Batch from reader (stdin, file, etc.)
results, err := h.BatchHashFromReader(ctx, file, 5)
```

### Fingerprint Database SDK

```go
import "github.com/cyberspacesec/iconhash-skills/pkg/fingerprint"

// Load default database
db := fingerprint.DefaultDB()

// Lookup by hash
matches := db.Lookup("-305179312")

// Lookup by category
entries := db.LookupByCategory("CMS")

// Lookup by tag
entries := db.LookupByTag("chinese")

// Multi-criteria filter
entries := db.Filter(fingerprint.FilterOptions{
    Category: "Server",
    Tag:      "java",
})

// Search (fuzzy, across all fields)
results := db.Search("jenkins")

// Statistics
fmt.Printf("Unique hashes: %d\n", db.Count())
fmt.Printf("Total entries: %d\n", db.TotalEntries())
categories := db.Categories()

// Export
jsonData, err := db.ExportJSON()
csvData := db.ExportCSV()
err = db.ExportJSONToFile("fingerprints.json")
err = db.ExportCSVToFile("fingerprints.csv")

// Load custom data
err := db.LoadFromJSON("custom.json")
err := db.LoadFromData(jsonBytes)
```

### Format Utilities

```go
import "github.com/cyberspacesec/iconhash-skills/pkg/util"

// Format hash for a specific engine
formatted := util.FormatHash("-305179312", util.FormatFofa)
// → icon_hash="-305179312"

// Format for all engines at once
queries := util.FormatAll("-305179312")
// → map[string]string{"fofa": ..., "shodan": ..., "censys": ..., "quake": ..., "zoomeye": ..., "hunter": ...}

// Parse format from string
format := util.ParseFormat("shodan") // → util.FormatShodan

// Get format name
name := util.FormatName(format) // → "shodan"
```

### API Server SDK

```go
import "github.com/cyberspacesec/iconhash-skills/pkg/api"

config := api.DefaultConfig()
config.Port = 9090
config.AuthToken = "my-secret"
config.Proxy = "socks5://127.0.0.1:1080"

server := api.NewServer(config)
server.Start() // blocks until shutdown
```

### MCP Handler SDK

```go
import "github.com/cyberspacesec/iconhash-skills/pkg/mcp"

handler := mcp.NewHandler(false)

// List available tools
tools := handler.Tools()

// Call a tool
result := handler.CallTool("iconhash_url", map[string]interface{}{
    "url": "https://example.com/favicon.ico",
    "format": "fofa",
})
```

---

## Supported Platforms

| OS | Architectures | Build IDs |
|---|---|---|
| Linux | x86_64, aarch64, i386, armv6h, riscv64 | `linux_amd64`, `linux_arm64`, `linux_i386`, `linux_armv6h`, `linux_riscv64` |
| macOS | x86_64 (Intel), aarch64 (Apple Silicon) | `macos_x86_64`, `macos_aarch64` |
| Windows | x86_64, aarch64, i386 | `windows_x86_64`, `windows_aarch64`, `windows_i386` |
| FreeBSD | x86_64, aarch64 | `freebsd_x86_64`, `freebsd_aarch64` |

---

## Environment Variables

| Variable | Description |
|---|---|
| `ICONHASH_FINGERPRINT_DB` | Path to custom fingerprint JSON database |
| `HTTP_PROXY` / `HTTPS_PROXY` | System proxy (used by Go's net/http) |

---

## Exit Codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | Error (invalid input, network failure, etc.) |

---

## Common Workflows

### Quick Hash Check

```bash
iconhash url https://target.com/favicon.ico
```

### Identify a Website

```bash
iconhash identify https://target.com
```

### Search for Similar Assets on Fofa

```bash
iconhash url https://target.com/favicon.ico --engine fofa
# Copy the output and paste into fofa.info search
```

### Bulk Reconnaissance

```bash
# Create a URL list
cat > urls.txt << EOF
https://site1.com
https://site2.com
https://site3.com
EOF

# Batch hash and save results
iconhash batch -i urls.txt -o results.json --engine fofa
```

### Offline Usage (Full Build)

```bash
# Download the full build which has embedded fingerprints
# No network needed for fingerprint lookup
iconhash lookup -- -305179312
```
