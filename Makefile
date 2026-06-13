# Build variables
BINARY_NAME=iconhash
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(shell date +%FT%T%z)

# Docker variables
DOCKER_IMAGE=cyberspacesec/iconhash
DOCKER_TAG=$(VERSION)

# Go build flags
LDFLAGS=-ldflags "-X github.com/cyberspacesec/iconhash-skills/cmd.Version=${VERSION} -X github.com/cyberspacesec/iconhash-skills/cmd.BuildDate=${DATE} -X github.com/cyberspacesec/iconhash-skills/cmd.BuildHash=${COMMIT}"

# Cross-compilation targets
PLATFORMS=darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64

.PHONY: all build build-lite build-full clean install test \
        release release-lite release-full release-all \
        cross-compile docker-build docker-push

all: clean build

# ============================================
# Build targets
# ============================================

# Default: lightweight build without embedded fingerprints
build: build-lite

# Lightweight build - smaller binary, fingerprints loaded externally
build-lite:
	@echo "Building ${BINARY_NAME} (lite - no embedded fingerprints)..."
	go build ${LDFLAGS} -o ${BINARY_NAME} .

# Full build - larger binary with embedded fingerprints
build-full:
	@echo "Building ${BINARY_NAME} (full - with embedded fingerprints)..."
	go build ${LDFLAGS} -tags embed_fingerprints -o ${BINARY_NAME} .

# Build with race detector for development
build-race:
	@echo "Building ${BINARY_NAME} with race detector..."
	go build -race ${LDFLAGS} -tags embed_fingerprints -o ${BINARY_NAME} .

clean:
	@echo "Cleaning up..."
	rm -f ${BINARY_NAME}
	rm -f ${BINARY_NAME}-*

install: build
	@echo "Installing ${BINARY_NAME}..."
	go install ${LDFLAGS} .

test:
	@echo "Running tests..."
	go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# ============================================
# Release targets
# ============================================

# Release lite version (default distribution - no embedded fingerprints)
release-lite:
	@echo "Building release: lite (no embedded fingerprints)..."
	@mkdir -p dist
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*}; \
		GOARCH=$$($${platform#*/}); \
		OUTPUT=dist/${BINARY_NAME}-$${GOOS}-$${GOARCH}; \
		if [ "$$GOOS" = "windows" ]; then OUTPUT=$${OUTPUT}.exe; fi; \
		echo "  Building $$OUTPUT..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build ${LDFLAGS} -o $$OUTPUT .; \
	done
	@echo "Lite release builds complete in dist/"

# Release full version (with embedded fingerprints)
release-full:
	@echo "Building release: full (with embedded fingerprints)..."
	@mkdir -p dist
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*}; \
		GOARCH=$$($${platform#*/}); \
		OUTPUT=dist/${BINARY_NAME}-full-$${GOOS}-$${GOARCH}; \
		if [ "$$GOOS" = "windows" ]; then OUTPUT=$${OUTPUT}.exe; fi; \
		echo "  Building $$OUTPUT..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build ${LDFLAGS} -tags embed_fingerprints -o $$OUTPUT .; \
	done
	@echo "Full release builds complete in dist/"

# Build both lite and full releases
release-all: release-lite release-full
	@echo "All release builds complete."
	@ls -lh dist/

# Shorthand for release-all
release: release-all

# ============================================
# Development targets
# ============================================

# Create a sample favicon.ico for testing
sample:
	@echo "Creating sample favicon.ico for testing..."
	mkdir -p test
	curl -s -o test/favicon.ico https://www.baidu.com/favicon.ico

# Run the tool with test favicon
test-sample: build sample
	@echo "Testing with sample favicon.ico..."
	./${BINARY_NAME} -f test/favicon.ico

# Run the tool with a URL
test-url: build
	@echo "Testing with URL..."
	./${BINARY_NAME} -u https://www.baidu.com/favicon.ico

# Export the fingerprint database to JSON
export-fingerprints:
	@echo "Exporting fingerprint database to data/fingerprints.json..."
	@mkdir -p data
	go run -tags=embed_fingerprints ${LDFLAGS} -exec "echo" . 2>/dev/null; \
	echo "Use: go run -tags=embed_fingerprints /tmp/export_fp/main.go > data/fingerprints.json"

# ============================================
# Docker targets
# ============================================

docker-build:
	@echo "Building Docker image ${DOCKER_IMAGE}:${DOCKER_TAG}..."
	docker build -t ${DOCKER_IMAGE}:${DOCKER_TAG} .
	docker tag ${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_IMAGE}:latest

docker-push: docker-build
	@echo "Pushing Docker image ${DOCKER_IMAGE}:${DOCKER_TAG}..."
	docker push ${DOCKER_IMAGE}:${DOCKER_TAG}
	docker push ${DOCKER_IMAGE}:latest

# Run the tool inside a Docker container
docker-run:
	@echo "Running in Docker container..."
	docker run --rm ${DOCKER_IMAGE}:latest $(ARGS)
