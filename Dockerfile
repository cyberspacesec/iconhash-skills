# Builder stage
FROM golang:1.25-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w -X github.com/cyberspacesec/iconhash-skills/cmd.Version=$(git describe --tags --always 2>/dev/null || echo dev) -X github.com/cyberspacesec/iconhash-skills/cmd.BuildHash=$(git rev-parse --short HEAD 2>/dev/null || echo unknown) -X github.com/cyberspacesec/iconhash-skills/cmd.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o iconhash .

# Final stage
FROM alpine:latest

# Add CA certificates
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -h /app appuser

# Set working directory
WORKDIR /app

# Copy binary from the builder stage
COPY --from=builder /app/iconhash /app/iconhash

# Set permissions
RUN chown -R appuser:appuser /app
USER appuser

# Set entrypoint
ENTRYPOINT ["/app/iconhash"]

# Default command
CMD ["--help"]
