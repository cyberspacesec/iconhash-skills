<h1 align="center">🔍 IconHash Skills</h1>

<p align="center">
  <strong>Favicon Hash Calculator for Cyber-Space Mapping</strong><br>
  Calculate MMH3 favicon hashes and identify web services for cybersecurity reconnaissance.
</p>

<p align="center">
  <a href="https://github.com/cyberspacesec/iconhash-skills/releases/latest"><img src="https://img.shields.io/github/v/release/cyberspacesec/iconhash-skills?style=flat-square" alt="Release"></a>
  <a href="https://github.com/cyberspacesec/iconhash-skills/actions"><img src="https://img.shields.io/github/actions/workflow/status/cyberspacesec/iconhash-skills/build.yml?style=flat-square" alt="CI"></a>
  <img src="https://img.shields.io/github/license/cyberspacesec/iconhash-skills?style=flat-square" alt="License">
  <img src="https://img.shields.io/github/go-mod/go-version/cyberspacesec/iconhash-skills?style=flat-square" alt="Go Version">
</p>

<p align="center">
  🇬🇧 <a href="#english">English</a> &nbsp;|&nbsp; 🇨🇳 <a href="#简体中文">简体中文</a>
</p>

---

<a id="english"></a>

## 🚀 Integration Methods

IconHash supports **4 integration methods** — choose the one that fits your workflow:

| # | Method | Best For | Quick Start |
|---|--------|----------|-------------|
| 1 | **🧠 SKILLS** | AI Agents (Claude, GPT, etc.) | One-click copy below ↓ |
| 2 | **💻 CLI** | Terminal users, scripting | [Install CLI](#installation) |
| 3 | **📦 Go SDK** | Go developers, embedding | `go get github.com/cyberspacesec/iconhash-skills` |
| 4 | **🔌 MCP / API** | AI tools, web services, automation | [API & MCP](#api-server--mcp) |

### 🧠 SKILLS Integration (for AI Agents)

Add this to your AI Agent's skills configuration to give it favicon hashing and web service identification capabilities:

```
# IconHash Skills — Favicon Hash Calculator for Cyber-Space Mapping

## What it does
Calculate MMH3 favicon hashes and identify web services from their favicons. Output search queries for Fofa, Shodan, Censys, Quake, ZoomEye, Hunter.

## How to install
### Option A: Download binary (recommended)
Linux x86_64: `wget -qO- https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_linux_x86_64.tar.gz | tar xz && chmod +x iconhash && sudo mv iconhash /usr/local/bin/`
macOS Apple Silicon: `wget -qO- https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_macos_aarch64.tar.gz | tar xz && chmod +x iconhash && sudo mv iconhash /usr/local/bin/`
macOS Intel: `wget -qO- https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_macos_x86_64.tar.gz | tar xz && chmod +x iconhash && sudo mv iconhash /usr/local/bin/`

### Option B: Build from source
`git clone https://github.com/cyberspacesec/iconhash-skills.git && cd iconhash-skills && make build && sudo mv iconhash /usr/local/bin/`

### Option C: Go install
`go install github.com/cyberspacesec/iconhash-skills/cmd/iconhash@latest`

## Available commands
- `iconhash url <url>` — Calculate hash from favicon URL
- `iconhash file <path>` — Calculate hash from local file
- `iconhash base64 <path>` — Calculate hash from base64 file
- `iconhash discover <url>` — Discover all favicons on a website and hash them
- `iconhash identify <url>` — Full identification: discover + hash + fingerprint lookup
- `iconhash lookup <hash>` — Lookup a hash in the fingerprint database (700+ services)
- `iconhash batch -i urls.txt` — Batch hash URLs from file or stdin
- `iconhash fingerprints` — Browse/search fingerprint database
- `iconhash fingerprints update` — Update fingerprint database from remote
- `iconhash server -p 8080` — Start HTTP API server

## Key flags
- `--engine fofa|shodan|censys|quake|zoomeye|hunter` — Output format for search engines
- `--insecure` — Skip TLS verification
- `--proxy socks5://host:port` — Use proxy
- `--fingerprint-db <path>` — Custom fingerprint database
- `--uint32` — Output as uint32 instead of int32

## Examples
- `iconhash url https://example.com/favicon.ico` → hash: -305179312
- `iconhash identify https://gitlab.example.com` → discover favicons, hash them, match fingerprints
- `iconhash lookup -- -305179312` → identify service from hash
- `iconhash url https://example.com/favicon.ico --engine fofa` → icon_hash="-305179312"

## Full documentation
See https://github.com/cyberspacesec/iconhash-skills/blob/main/SKILLS.md
```

> 💡 **Tip:** Copy the text above directly into your Claude Code SKILLS file, Cursor rules, or any AI Agent configuration.

## Features

- 🧮 Calculate MMH3 (MurmurHash3) hash of favicons
- 🌐 Multiple input sources: URL, local file, base64 data, stdin
- 🔎 6 search engine formats: Fofa, Shodan, Censys, Quake, ZoomEye, Hunter
- 🏷️ Fingerprint database with 700+ known services for identification
- ⚡ Batch processing with concurrent workers
- 🖥️ HTTP API server with authentication
- 🤖 Model Context Protocol (MCP) support for AI integration
- 📦 Full Go SDK for embedding in other tools
- 🐳 Docker support for containerized usage
- 📋 SKILLS documentation for AI Agent integration
- 🏗️ Two build variants: Lite (small) and Full (offline-capable)

## Installation

### Download Pre-built Binary (Recommended)

Download from [GitHub Releases](https://github.com/cyberspacesec/iconhash-skills/releases/latest).

**Build variants:**
- **Lite** (`iconhash_lite_*`): Smaller binary, fingerprints auto-downloaded on first use
- **Full** (`iconhash_full_*`): Larger binary with embedded fingerprints, works offline

```bash
# Linux x86_64 (Lite)
wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_linux_x86_64.tar.gz
tar xzf iconhash_lite_linux_x86_64.tar.gz
chmod +x iconhash && sudo mv iconhash /usr/local/bin/

# macOS Apple Silicon (Lite)
wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_macos_aarch64.tar.gz
tar xzf iconhash_lite_macos_aarch64.tar.gz
chmod +x iconhash && sudo mv iconhash /usr/local/bin/

# Windows x86_64 (download and extract zip)
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

# Lite build (default)
make build

# Full build (with embedded fingerprints, works offline)
make build-full
```

### Docker

```bash
docker pull cyberspacesec/iconhash:latest
docker run --rm cyberspacesec/iconhash:latest url https://example.com/favicon.ico
```

## Quick Start

```bash
iconhash url https://www.example.com/favicon.ico      # Hash from URL
iconhash identify https://example.com                   # Identify a website
iconhash lookup -- -305179312                           # Lookup fingerprint
iconhash batch -i urls.txt -o results.json              # Batch process
iconhash server -p 8080                                 # Start API server
```

## 💻 CLI Reference

| Command | Purpose |
|---|---|
| `iconhash url <url>` | Calculate hash from a URL |
| `iconhash file <path>` | Calculate hash from a file |
| `iconhash base64 <path>` | Calculate hash from base64 file |
| `iconhash discover <url>` | Discover favicons on a site |
| `iconhash identify <url>` | Full identification (discover + hash + fingerprint) |
| `iconhash lookup <hash>` | Lookup hash in fingerprint DB |
| `iconhash batch -i <file>` | Batch process URLs |
| `iconhash fingerprints` | Browse/search fingerprint DB |
| `iconhash fingerprints update` | Update fingerprint DB |
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
| `--format` | | Output: `text`, `json`, `csv` |
| `--output` | `-o` | Output file path |
| `--fingerprint-db` | | Custom fingerprint database |

📖 See [SKILLS.md](./SKILLS.md) for complete documentation.

## 📦 Go SDK

```bash
go get github.com/cyberspacesec/iconhash-skills
```

```go
import (
    "context"
    "github.com/cyberspacesec/iconhash-skills/pkg/hasher"
    "github.com/cyberspacesec/iconhash-skills/pkg/fingerprint"
    "github.com/cyberspacesec/iconhash-skills/pkg/util"
)

// One-off hash with defaults
hash, _ := hasher.HashURL(context.Background(), "https://example.com/favicon.ico")

// Custom hasher
h := hasher.New(&hasher.HashOptions{InsecureSkipVerify: true})
hash, _ = h.HashFromURL(ctx, "https://example.com/favicon.ico")

// Full identification
db := fingerprint.DefaultDB()
results := h.Identify(ctx, "https://example.com", db, nil)

// Batch processing
results := h.BatchHashURLs(ctx, urls, 10)

// Format for all search engines
queries := util.FormatAll(hash)

// With proxy
opts, _ := hasher.NewOptionsWithProxy("socks5://127.0.0.1:1080", 30*time.Second, true)
```

## 🔌 API Server & MCP

### HTTP API Server

```bash
iconhash server -p 8080 --auth-token secret123
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

### MCP Integration

```go
import "github.com/cyberspacesec/iconhash-skills/pkg/mcp"

handler := mcp.NewHandler(false)
tools := handler.Tools()
result := handler.CallTool("iconhash_url", map[string]interface{}{
    "url": "https://example.com/favicon.ico",
})
```

**MCP Tools:** `iconhash_url`, `iconhash_base64`, `iconhash_file`, `iconhash_discover`, `iconhash_lookup`

## Search Engine Formats

| Engine | Format | Example |
|---|---|---|
| Fofa | `icon_hash="<hash>"` | `icon_hash="-305179312"` |
| Shodan | `http.favicon.hash:<hash>` | `http.favicon.hash:-305179312` |
| Censys | `services.http.response.favicons.md5_hash:<hash>` | |
| Quake | `favicon.hash:"<hash>"` | |
| ZoomEye | `iconhash:"<hash>"` | |
| Hunter | `web.icon="<hash>"` | |

## Docker

```bash
docker run --rm cyberspacesec/iconhash:latest url https://example.com/favicon.ico
docker run -d -p 8080:8080 cyberspacesec/iconhash:latest server -H 0.0.0.0 -p 8080
```

## Development

```bash
make build          # Lite build
make build-full     # Full build (with fingerprints)
make test           # Run tests
make test-coverage  # Run tests with coverage
```

## License

[MIT License](LICENSE)

---

<a id="简体中文"></a>

## 🚀 接入方式

IconHash 支持 **4 种接入方式** —— 根据你的工作流选择：

| # | 方式 | 适用场景 | 快速开始 |
|---|------|---------|---------|
| 1 | **🧠 SKILLS** | AI Agent（Claude、GPT 等） | 一键复制 ↓ |
| 2 | **💻 CLI** | 终端用户、脚本 | [安装 CLI](#安装) |
| 3 | **📦 Go SDK** | Go 开发者、嵌入式集成 | `go get github.com/cyberspacesec/iconhash-skills` |
| 4 | **🔌 MCP / API** | AI 工具、Web 服务、自动化 | [API 与 MCP](#api-服务与-mcp) |

### 🧠 SKILLS 接入（面向 AI Agent）

将以下内容添加到你的 AI Agent 技能配置中，为其赋予 favicon 哈希计算和 Web 服务识别能力：

```
# IconHash Skills — 网络空间测绘 Favicon 哈希计算工具

## 功能描述
计算 favicon 的 MMH3 哈希，并通过指纹库识别 Web 服务。输出适用于 Fofa、Shodan、Censys、Quake、ZoomEye、Hunter 的搜索语法。

## 安装方式
### 方式 A：下载二进制文件（推荐）
Linux x86_64: `wget -qO- https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_linux_x86_64.tar.gz | tar xz && chmod +x iconhash && sudo mv iconhash /usr/local/bin/`
macOS Apple Silicon: `wget -qO- https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_macos_aarch64.tar.gz | tar xz && chmod +x iconhash && sudo mv iconhash /usr/local/bin/`
macOS Intel: `wget -qO- https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_macos_x86_64.tar.gz | tar xz && chmod +x iconhash && sudo mv iconhash /usr/local/bin/`

### 方式 B：源码编译
`git clone https://github.com/cyberspacesec/iconhash-skills.git && cd iconhash-skills && make build && sudo mv iconhash /usr/local/bin/`

### 方式 C：Go 安装
`go install github.com/cyberspacesec/iconhash-skills/cmd/iconhash@latest`

## 可用命令
- `iconhash url <url>` — 从 URL 计算 favicon 哈希
- `iconhash file <path>` — 从本地文件计算哈希
- `iconhash base64 <path>` — 从 base64 文件计算哈希
- `iconhash discover <url>` — 发现网站所有 favicon 并计算哈希
- `iconhash identify <url>` — 完整识别：发现 + 哈希 + 指纹匹配
- `iconhash lookup <hash>` — 在指纹库中查找哈希（700+ 服务）
- `iconhash batch -i urls.txt` — 批量计算 URL 列表
- `iconhash fingerprints` — 浏览/搜索指纹数据库
- `iconhash fingerprints update` — 更新指纹数据库
- `iconhash server -p 8080` — 启动 HTTP API 服务器

## 关键参数
- `--engine fofa|shodan|censys|quake|zoomeye|hunter` — 搜索引擎输出格式
- `--insecure` — 跳过 TLS 证书验证
- `--proxy socks5://host:port` — 使用代理
- `--fingerprint-db <path>` — 自定义指纹数据库
- `--uint32` — 输出 uint32 而非 int32

## 示例
- `iconhash url https://example.com/favicon.ico` → 哈希: -305179312
- `iconhash identify https://gitlab.example.com` → 发现 favicon、计算哈希、匹配指纹
- `iconhash lookup -- -305179312` → 从哈希识别服务
- `iconhash url https://example.com/favicon.ico --engine fofa` → icon_hash="-305179312"

## 完整文档
查看 https://github.com/cyberspacesec/iconhash-skills/blob/main/SKILLS.md
```

> 💡 **提示：** 直接将上方文本复制到你的 Claude Code SKILLS 文件、Cursor rules 或任何 AI Agent 配置中即可。

## 功能特性

- 🧮 计算 favicon 的 MMH3（MurmurHash3）哈希
- 🌐 多种输入源：URL、本地文件、base64 数据、标准输入
- 🔎 6 种搜索引擎格式：Fofa、Shodan、Censys、Quake、ZoomEye、Hunter
- 🏷️ 指纹数据库包含 700+ 已知服务，支持识别
- ⚡ 批量处理，支持并发工作池
- 🖥️ HTTP API 服务器，支持认证
- 🤖 支持模型上下文协议（MCP），便于 AI 集成
- 📦 完整的 Go SDK，可嵌入其他工具
- 🐳 支持 Docker 容器化部署
- 📋 SKILLS 文档，方便 AI Agent 接入
- 🏗️ 两种构建变体：Lite（小体积）和 Full（离线可用）

<a id="安装"></a>

## 安装

### 下载预构建二进制文件（推荐）

从 [GitHub Releases](https://github.com/cyberspacesec/iconhash-skills/releases/latest) 下载。

**构建变体：**
- **Lite**（`iconhash_lite_*`）：更小的二进制，指纹首次使用时自动下载
- **Full**（`iconhash_full_*`）：更大的二进制，内嵌指纹，离线可用

```bash
# Linux x86_64（Lite）
wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_linux_x86_64.tar.gz
tar xzf iconhash_lite_linux_x86_64.tar.gz
chmod +x iconhash && sudo mv iconhash /usr/local/bin/

# macOS Apple Silicon（Lite）
wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_macos_aarch64.tar.gz
tar xzf iconhash_lite_macos_aarch64.tar.gz
chmod +x iconhash && sudo mv iconhash /usr/local/bin/

# Windows x86_64（下载并解压 zip）
# https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_windows_x86_64.zip
```

**支持平台：** Linux（x86_64、aarch64、i386、arm、riscv64）、macOS（x86_64、aarch64）、Windows（x86_64、aarch64、i386）、FreeBSD（x86_64、aarch64）

### 通过 Go 安装

```bash
go install github.com/cyberspacesec/iconhash-skills/cmd/iconhash@latest
```

### 源码编译

```bash
git clone https://github.com/cyberspacesec/iconhash-skills.git
cd iconhash-skills

# Lite 构建（默认）
make build

# Full 构建（内嵌指纹，离线可用）
make build-full
```

### Docker

```bash
docker pull cyberspacesec/iconhash:latest
docker run --rm cyberspacesec/iconhash:latest url https://example.com/favicon.ico
```

## 快速开始

```bash
iconhash url https://www.example.com/favicon.ico      # 从 URL 计算哈希
iconhash identify https://example.com                   # 识别网站
iconhash lookup -- -305179312                           # 查找指纹
iconhash batch -i urls.txt -o results.json              # 批量处理
iconhash server -p 8080                                 # 启动 API 服务器
```

## 💻 CLI 命令

| 命令 | 功能 |
|---|---|
| `iconhash url <url>` | 从 URL 计算哈希 |
| `iconhash file <path>` | 从文件计算哈希 |
| `iconhash base64 <path>` | 从 base64 文件计算哈希 |
| `iconhash discover <url>` | 发现网站 favicon |
| `iconhash identify <url>` | 完整识别（发现 + 哈希 + 指纹） |
| `iconhash lookup <hash>` | 在指纹库中查找 |
| `iconhash batch -i <file>` | 批量处理 URL |
| `iconhash fingerprints` | 浏览/搜索指纹库 |
| `iconhash fingerprints update` | 更新指纹库 |
| `iconhash server` | 启动 HTTP API 服务器 |

### 全局参数

| 参数 | 缩写 | 说明 |
|---|---|---|
| `--engine` | `-e` | 格式：`plain`、`fofa`、`shodan`、`censys`、`quake`、`zoomeye`、`hunter` |
| `--uint32` | `-n` | 输出 uint32 而非 int32 |
| `--insecure` | `-k` | 跳过 TLS 验证 |
| `--timeout` | `-t` | HTTP 超时（默认 30s） |
| `--proxy` | | HTTP/SOCKS5 代理地址 |
| `--debug` | `-d` | 启用调试输出 |
| `--format` | | 输出：`text`、`json`、`csv` |
| `--output` | `-o` | 输出文件路径 |
| `--fingerprint-db` | | 自定义指纹数据库 |

📖 完整文档见 [SKILLS.md](./SKILLS.md)。

## 📦 Go SDK

```bash
go get github.com/cyberspacesec/iconhash-skills
```

```go
import (
    "context"
    "github.com/cyberspacesec/iconhash-skills/pkg/hasher"
    "github.com/cyberspacesec/iconhash-skills/pkg/fingerprint"
    "github.com/cyberspacesec/iconhash-skills/pkg/util"
)

// 一次性计算（默认选项）
hash, _ := hasher.HashURL(context.Background(), "https://example.com/favicon.ico")

// 自定义 hasher
h := hasher.New(&hasher.HashOptions{InsecureSkipVerify: true})
hash, _ = h.HashFromURL(ctx, "https://example.com/favicon.ico")

// 完整识别
db := fingerprint.DefaultDB()
results := h.Identify(ctx, "https://example.com", db, nil)

// 批量处理
results := h.BatchHashURLs(ctx, urls, 10)

// 格式化为所有搜索引擎
queries := util.FormatAll(hash)

// 使用代理
opts, _ := hasher.NewOptionsWithProxy("socks5://127.0.0.1:1080", 30*time.Second, true)
```

<a id="api-服务与-mcp"></a>

## 🔌 API 服务与 MCP

### HTTP API 服务器

```bash
iconhash server -p 8080 --auth-token secret123
```

| 端点 | 方法 | 说明 |
|---|---|---|
| `/health` | GET | 健康检查 |
| `/hash/url` | GET, POST | 从 URL 计算哈希 |
| `/hash/file` | POST | 从文件计算哈希 |
| `/hash/base64` | POST | 从 base64 计算哈希 |
| `/hash/batch` | POST | 批量计算 |
| `/hash/discover` | POST | 发现 favicon |
| `/lookup` | GET | 指纹查找 |
| `/fingerprints` | GET | 指纹数据库 |
| `/mcp` | POST | 模型上下文协议 |

### MCP 集成

```go
import "github.com/cyberspacesec/iconhash-skills/pkg/mcp"

handler := mcp.NewHandler(false)
tools := handler.Tools()
result := handler.CallTool("iconhash_url", map[string]interface{}{
    "url": "https://example.com/favicon.ico",
})
```

**MCP 工具：** `iconhash_url`、`iconhash_base64`、`iconhash_file`、`iconhash_discover`、`iconhash_lookup`

## 搜索引擎格式

| 引擎 | 格式 | 示例 |
|---|---|---|
| Fofa | `icon_hash="<hash>"` | `icon_hash="-305179312"` |
| Shodan | `http.favicon.hash:<hash>` | `http.favicon.hash:-305179312` |
| Censys | `services.http.response.favicons.md5_hash:<hash>` | |
| Quake | `favicon.hash:"<hash>"` | |
| ZoomEye | `iconhash:"<hash>"` | |
| Hunter | `web.icon="<hash>"` | |

## Docker

```bash
docker run --rm cyberspacesec/iconhash:latest url https://example.com/favicon.ico
docker run -d -p 8080:8080 cyberspacesec/iconhash:latest server -H 0.0.0.0 -p 8080
```

## 开发

```bash
make build          # Lite 构建
make build-full     # Full 构建（含指纹）
make test           # 运行测试
make test-coverage  # 运行测试并生成覆盖率
```

## 许可证

[MIT License](LICENSE)
