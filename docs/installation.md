# IconHash 安装与部署指南

IconHash 是一个用于计算 favicon MMH3 哈希值的命令行工具，专为网络空间测绘和安全侦察设计。本文档说明如何在不同平台上安装和部署 IconHash。

> **仓库地址：** https://github.com/cyberspacesec/iconhash-skills

---

## 快速安装（推荐）

从 [GitHub Releases](https://github.com/cyberspacesec/iconhash-skills/releases) 下载适合你平台的预编译二进制文件。

### 一键安装脚本

```bash
# Linux / macOS（自动检测平台并下载 lite 版本）
curl -fsSL https://raw.githubusercontent.com/cyberspacesec/iconhash-skills/main/install.sh | bash
```

### 按平台手动下载

#### Linux

| 平台 | 下载命令 |
|------|---------|
| **x86_64**（大多数服务器/PC） | `wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_linux_x86_64.tar.gz` |
| **ARM64**（树莓派5/云ARM实例） | `wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_linux_aarch64.tar.gz` |
| **i386**（32位系统） | `wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_linux_i386.tar.gz` |
| **ARMv6/v7**（树莓派3/旧设备） | `wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_linux_armv6h.tar.gz` |
| **RISC-V 64** | `wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_linux_riscv64.tar.gz` |

```bash
# 安装示例（以 x86_64 为例）
wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_linux_x86_64.tar.gz
tar xzf iconhash_lite_linux_x86_64.tar.gz
chmod +x iconhash
sudo mv iconhash /usr/local/bin/
iconhash --version
```

#### macOS

| 平台 | 下载命令 |
|------|---------|
| **Apple Silicon**（M1/M2/M3/M4） | `wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_darwin_aarch64.tar.gz` |
| **Intel**（x86_64） | `wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_darwin_x86_64.tar.gz` |

```bash
# 安装示例（Apple Silicon）
wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_darwin_aarch64.tar.gz
tar xzf iconhash_lite_darwin_aarch64.tar.gz
chmod +x iconhash
sudo mv iconhash /usr/local/bin/
iconhash --version
```

#### Windows

| 平台 | 下载链接 |
|------|---------|
| **x86_64** | `iconhash_lite_windows_x86_64.zip` |
| **ARM64** | `iconhash_lite_windows_aarch64.zip` |
| **i386**（32位） | `iconhash_lite_windows_i386.zip` |

```powershell
# PowerShell 安装示例（x86_64）
Invoke-WebRequest -Uri "https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_windows_x86_64.zip" -OutFile "iconhash.zip"
Expand-Archive -Path iconhash.zip -DestinationPath iconhash
cd iconhash
.\iconhash.exe --version
```

#### FreeBSD

| 平台 | 下载命令 |
|------|---------|
| **x86_64** | `wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_freebsd_x86_64.tar.gz` |
| **ARM64** | `wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_freebsd_aarch64.tar.gz` |

---

## 首次运行

安装后首次使用 `lookup` 或 `identify` 命令时，lite 版本会自动下载指纹库：

```bash
iconhash lookup -- -305179312
# 如果指纹库为空，会自动提示下载
```

或手动更新指纹库：

```bash
iconhash fingerprints update
```

---

## 构建模式说明

IconHash 提供两种构建模式，所有平台均提供两种版本：

| 特性 | Lite（精简版） | Full（完整版） |
|------|---------------|---------------|
| **二进制体积** | 较小 | 较大（内嵌 719+ 指纹） |
| **指纹库** | 运行时从外部加载 | 编译时内嵌 |
| **离线使用** | 需要先下载指纹库 | 开箱即用 |
| **自动下载** | ✅ 首次使用自动下载 | 不需要 |
| **适用场景** | CI/CD、容器、磁盘敏感 | 离线环境、快速部署 |
| **文件名** | `iconhash_lite_*` | `iconhash_full_*` |
| **二进制名** | `iconhash` | `iconhash-full` |

下载 Full 版本只需将下载链接中的 `lite` 替换为 `full`：

```bash
# Linux x86_64 Full 版本（离线可用）
wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_full_linux_x86_64.tar.gz
tar xzf iconhash_full_linux_x86_64.tar.gz
chmod +x iconhash-full
./iconhash-full --version
```

---

## 平台支持矩阵

| 操作系统 | 架构 | Lite | Full | 适用设备 |
|---------|------|------|------|---------|
| **Linux** | x86_64 | ✅ | ✅ | 服务器、PC、云主机 |
| **Linux** | ARM64 | ✅ | ✅ | 树莓派5、AWS Graviton、Oracle ARM |
| **Linux** | i386 | ✅ | ✅ | 旧式32位服务器 |
| **Linux** | ARMv6/v7 | ✅ | ✅ | 树莓派3、物联网设备 |
| **Linux** | RISC-V 64 | ✅ | ✅ | RISC-V 开发板 |
| **macOS** | x86_64 | ✅ | ✅ | Intel Mac |
| **macOS** | ARM64 | ✅ | ✅ | M1/M2/M3/M4 Mac |
| **Windows** | x86_64 | ✅ | ✅ | 64位 PC/Server |
| **Windows** | ARM64 | ✅ | ✅ | ARM 笔记本 |
| **Windows** | i386 | ✅ | ✅ | 32位 PC |
| **FreeBSD** | x86_64 | ✅ | ✅ | FreeBSD 服务器 |
| **FreeBSD** | ARM64 | ✅ | ✅ | FreeBSD ARM 设备 |

共计 **12 个平台 × 2 种模式 = 24 个二进制文件**。

---

## 从源码构建

如果你出于安全审计原因需要自行编译，或需要针对特定平台构建：

### 前置要求

- Go 1.21 或更高版本（推荐 1.22+）
- Git

### 克隆与构建

```bash
# 1. 克隆仓库
git clone https://github.com/cyberspacesec/iconhash-skills.git
cd iconhash-skills

# 2. 安装依赖
go mod download

# 3a. 构建 Lite 版本（默认，不含内嵌指纹库）
go build -trimpath -ldflags "-s -w" -o iconhash .

# 3b. 构建 Full 版本（含内嵌 719+ 指纹库）
go build -trimpath -tags embed_fingerprints -ldflags "-s -w" -o iconhash-full .

# 4. 验证
./iconhash --version
./iconhash lookup -- -305179312

# 5. 运行测试
go test ./...                           # Lite 模式测试
go test -tags=embed_fingerprints ./...  # Full 模式测试
```

### 使用 Makefile 构建

```bash
make build          # Lite 版本
make build-full     # Full 版本
make build-race     # 带 race detector 的 Full 版本（开发用）
make test           # 运行测试
make test-coverage  # 测试覆盖率报告
```

### 交叉编译

Go 原生支持交叉编译，无需安装额外工具：

```bash
# Linux ARM64（AWS Graviton / 树莓派5）
GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o iconhash .

# Linux ARMv6（树莓派3）
GOOS=linux GOARCH=arm GOARM=6 go build -trimpath -ldflags "-s -w" -o iconhash .

# Linux RISC-V
GOOS=linux GOARCH=riscv64 go build -trimpath -ldflags "-s -w" -o iconhash .

# Linux i386（32位）
GOOS=linux GOARCH=386 go build -trimpath -ldflags "-s -w" -o iconhash .

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o iconhash .

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o iconhash .

# Windows x86_64
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o iconhash.exe .

# Windows ARM64
GOOS=windows GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o iconhash.exe .

# FreeBSD x86_64
GOOS=freebsd GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o iconhash .

# FreeBSD ARM64
GOOS=freebsd GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o iconhash .
```

带版本信息的交叉编译：

```bash
VERSION="1.0.0"
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date +%FT%T%z)

GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "-s -w \
  -X github.com/cyberspacesec/iconhash-skills/cmd.Version=${VERSION} \
  -X github.com/cyberspacesec/iconhash-skills/cmd.BuildDate=${DATE} \
  -X github.com/cyberspacesec/iconhash-skills/cmd.BuildHash=${COMMIT}" \
  -o iconhash .
```

---

## 指纹库管理

### 指纹库加载优先级

IconHash 按以下优先级加载指纹库：

1. `--fingerprint-db` 命令行参数（显式指定路径）
2. `ICONHASH_FINGERPRINT_DB` 环境变量（全局配置）
3. 内嵌指纹库（Full 版本编译时内嵌）
4. `~/.iconhash/fingerprints.json` 本地缓存
5. 自动从 GitHub 下载（lite 版本首次使用时）

### 手动更新指纹库

```bash
# 从默认源更新（增量合并，不丢失已有数据）
iconhash fingerprints update

# 从自定义 URL 更新
iconhash fingerprints update --url https://example.com/fingerprints.json

# 保存到自定义路径
iconhash fingerprints update -o /path/to/fingerprints.json

# 替换而非合并
iconhash fingerprints update --replace
```

### 使用环境变量

```bash
# 设置指纹库路径（推荐，一次设置全局生效）
export ICONHASH_FINGERPRINT_DB=~/.iconhash/fingerprints.json

# 之后所有命令自动使用该指纹库
iconhash lookup -- -305179312
iconhash identify https://example.com
```

### 手动下载指纹库

```bash
# 创建目录
mkdir -p ~/.iconhash

# 下载指纹库
curl -o ~/.iconhash/fingerprints.json \
  https://raw.githubusercontent.com/cyberspacesec/iconhash-skills/main/data/fingerprints.json

# 设置环境变量
export ICONHASH_FINGERPRINT_DB=~/.iconhash/fingerprints.json
```

---

## Docker 部署

```bash
# 拉取镜像
docker pull cyberspacesec/iconhash:latest

# 运行单个命令
docker run --rm cyberspacesec/iconhash:latest --version
docker run --rm cyberspacesec/iconhash:latest lookup -- -305179312

# 启动 API 服务
docker run -p 8080:8080 cyberspacesec/iconhash:latest server -H 0.0.0.0

# 带认证的 API 服务
docker run -p 8080:8080 cyberspacesec/iconhash:latest server -H 0.0.0.0 --auth-token your-secret-token
```

---

## API 服务部署

```bash
# 启动 HTTP API 服务
iconhash server -p 8080 -H 0.0.0.0

# 带认证
iconhash server -p 8080 --auth-token your-secret-token

# 通过代理访问外部 URL
iconhash server -p 8080 --proxy socks5://127.0.0.1:1080

# 自定义指纹库
iconhash server -p 8080 --fingerprint-db /path/to/fingerprints.json
```

### API 端点一览

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |
| POST | `/hash/url?url=...` | URL 哈希计算 |
| POST | `/hash/file` | 文件哈希计算 |
| POST | `/hash/base64` | Base64 哈希计算 |
| POST | `/hash/batch` | 批量哈希计算 |
| POST | `/hash/discover` | 发现 favicon |
| GET | `/lookup?hash=...` | 指纹库反查 |
| GET | `/fingerprints` | 指纹库信息 |
| POST | `/mcp` | Model Context Protocol |

---

## CI/CD 集成

### GitHub Actions

```yaml
- name: Install IconHash
  run: |
    wget https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_linux_x86_64.tar.gz
    tar xzf iconhash_lite_linux_x86_64.tar.gz
    chmod +x iconhash
    sudo mv iconhash /usr/local/bin/

- name: Scan favicon
  run: iconhash identify https://target.com
```

### 通用 Shell

```bash
# 自动检测平台并下载
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
[ "$ARCH" = "x86_64" ] && ARCH="x86_64"
[ "$ARCH" = "aarch64" ] && ARCH="aarch64"
[ "$ARCH" = "armv7l" ] && ARCH="armv6h"

wget "https://github.com/cyberspacesec/iconhash-skills/releases/latest/download/iconhash_lite_${OS}_${ARCH}.tar.gz"
tar xzf iconhash_lite_${OS}_${ARCH}.tar.gz
chmod +x iconhash
```

---

## 验证安装

```bash
# 检查版本
iconhash --version

# 测试基本功能
iconhash lookup -- -305179312    # 应识别为 Atlassian Confluence
iconhash lookup -- 81586312      # 应识别为 Jenkins

# 检查指纹库状态
iconhash fingerprints            # 显示统计信息（719+ 哈希）
iconhash fingerprints update     # 更新到最新版本
```

---

## 发布流程（维护者）

项目使用 GitHub Actions + GoReleaser 自动化发布：

1. **CI 测试**：每次 push 到 main 或 PR 时自动运行（`.github/workflows/build.yml`）
2. **自动发布**：推送 tag 时自动构建并发布 Release（`.github/workflows/release.yml`）

### 发布新版本

```bash
# 1. 确保所有测试通过
make test

# 2. 提交更改
git add .
git commit -m "feat: your changes"

# 3. 创建版本标签
git tag v1.2.0

# 4. 推送代码和标签
git push origin main
git push origin v1.2.0

# 5. GitHub Actions 自动构建并发布 Release
#    - 12 个平台 × 2 种模式 = 24 个二进制文件
#    - SHA256 校验和
#    - 自动生成 changelog
#    - Docker 镜像推送
```

### 手动本地测试 GoReleaser

```bash
# 安装 GoReleaser
go install github.com/goreleaser/goreleaser/v2@latest

# 本地测试（不发布）
goreleaser release --snapshot --clean
```

---

## 常见问题

### Q: 下载指纹库失败怎么办？

网络问题可能导致自动下载失败。你可以：
1. 使用 `iconhash fingerprints update` 重试
2. 手动下载：`curl -o ~/.iconhash/fingerprints.json https://raw.githubusercontent.com/cyberspacesec/iconhash-skills/main/data/fingerprints.json`
3. 使用 Full 版本，无需外部指纹库

### Q: 如何自定义指纹库？

```bash
# 方式1：命令行参数
iconhash lookup -- -305179312 --fingerprint-db /path/to/custom.json

# 方式2：环境变量（推荐）
export ICONHASH_FINGERPRINT_DB=/path/to/custom.json
iconhash lookup -- -305179312
```

### Q: Full 版本和 Lite 版本有什么区别？

- **Lite 版本**：二进制更小，指纹库运行时从外部加载或自动下载。适合 CI/CD、容器等场景。
- **Full 版本**：719+ 指纹内嵌在二进制中，无需网络即可完整使用。适合离线环境或快速部署。
- 功能完全相同，仅指纹库加载方式不同。

### Q: 如何确定我的平台架构？

```bash
# Linux/macOS
uname -m
# x86_64 → 下载 x86_64 版本
# aarch64 → 下载 aarch64 版本
# armv7l → 下载 armv6h 版本

# Windows PowerShell
[Environment]::Is64BitOperatingSystem
# True → 下载 x86_64 版本
# False → 下载 i386 版本
```
