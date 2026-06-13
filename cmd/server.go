package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cyberspacesec/iconhash-skills/pkg/api"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// NewServerCommand creates the server command
func NewServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the HTTP API server",
		Long: `Start the HTTP API server for favicon hash calculation.

This command starts a REST API server that can calculate favicon hashes from URLs,
files, or base64 encoded data. The server provides endpoints for health checking,
hash calculation, and supports the Model Context Protocol (MCP).

Examples:
  iconhash server
  iconhash server -p 8080 --host 0.0.0.0
  iconhash server --auth-token secret123 --debug`,
		RunE: runServer,
	}

	// Server-specific flags
	cmd.Flags().StringVarP(&Host, "host", "H", "127.0.0.1", "Host to bind server")
	cmd.Flags().IntVarP(&Port, "port", "p", 8080, "Port to bind server")
	cmd.Flags().StringVar(&AuthToken, "auth-token", "", "Authentication token for API requests")
	cmd.Flags().DurationVar(&ReadTimeout, "read-timeout", 30*time.Second, "HTTP server read timeout")
	cmd.Flags().DurationVar(&WriteTimeout, "write-timeout", 30*time.Second, "HTTP server write timeout")
	cmd.Flags().StringVar(&ServerProxy, "proxy", "", "HTTP/SOCKS5 proxy URL for outgoing requests (e.g. socks5://127.0.0.1:1080)")
	cmd.Flags().StringVar(&FingerprintDB, "fingerprint-db", "", "Path to custom fingerprint JSON database")

	SilenceUsageOnError(cmd)

	return cmd
}

// runServer handles the server command execution
func runServer(cmd *cobra.Command, args []string) error {
	// Create server config from flags
	config := &api.Config{
		Host:               Host,
		Port:               Port,
		ReadTimeout:        ReadTimeout,
		WriteTimeout:       WriteTimeout,
		AuthToken:          AuthToken,
		EnableDebug:        Debug,
		InsecureSkipVerify: SkipVerify,
		RequestTimeout:     Timeout,
		Proxy:              ServerProxy,
		FingerprintDB:      FingerprintDB,
	}

	// Create the server
	server := api.NewServer(config)

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop
		fmt.Println("\n🛑 Shutting down server...")
		server.Shutdown(context.Background())
	}()

	// Print startup message
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Println("🚀", cyan("Starting IconHash API Server"))
	fmt.Printf("⚙️  %s: %s\n", yellow("Configuration"), fmt.Sprintf("%s:%d", Host, Port))
	fmt.Printf("🔑 %s: %v\n", yellow("Authentication"), AuthToken != "")
	fmt.Printf("🐛 %s: %v\n", yellow("Debug Mode"), Debug)
	fmt.Printf("⏱️  %s: %v\n", yellow("Request Timeout"), Timeout)
	fmt.Printf("🔐 %s: %v\n", yellow("Insecure Skip Verify"), SkipVerify)
	if ServerProxy != "" {
		fmt.Printf("🌐 %s: %s\n", yellow("Proxy"), ServerProxy)
	}

	// Construct the base URL for easy access
	scheme := "http"
	baseURL := fmt.Sprintf("%s://%s:%d", scheme, Host, Port)
	if Host == "0.0.0.0" {
		baseURL = fmt.Sprintf("%s://localhost:%d", scheme, Port)
	}

	fmt.Println("\n📋", cyan("API Endpoints:"))
	fmt.Printf("  %s: %s/health\n", yellow("Health Check"), baseURL)
	fmt.Printf("  %s: %s/hash/url?url=...\n", yellow("URL Hash"), baseURL)
	fmt.Printf("  %s: %s/hash/file\n", yellow("File Hash"), baseURL)
	fmt.Printf("  %s: %s/hash/base64\n", yellow("Base64 Hash"), baseURL)
	fmt.Printf("  %s: %s/hash/batch\n", yellow("Batch Hash"), baseURL)
	fmt.Printf("  %s: %s/hash/discover\n", yellow("Discover Favicons"), baseURL)
	fmt.Printf("  %s: %s/lookup?hash=...\n", yellow("Fingerprint Lookup"), baseURL)
	fmt.Printf("  %s: %s/fingerprints\n", yellow("Fingerprints DB"), baseURL)
	fmt.Printf("  %s: %s/mcp\n", yellow("Model Context Protocol"), baseURL)

	fmt.Println("\n🔍", cyan("Query Parameters:"))
	fmt.Printf("  %s: uint32=true|false - Use uint32 format\n", yellow("Optional"))
	fmt.Printf("  %s: format=fofa|shodan|censys|quake|zoomeye|hunter|plain - Output format\n", yellow("Optional"))

	if AuthToken != "" {
		fmt.Println("\n🔒", cyan("Authentication:"))
		fmt.Printf("  %s: \"Authorization: Bearer <your-token>\"\n", yellow("Header"))
		fmt.Printf("  %s: \"?token=<your-token>\"\n", yellow("Query"))
		fmt.Printf("  %s: Token is configured (length: %d chars)\n", yellow("Note"), len(AuthToken))
	}

	fmt.Println("\n📢", cyan("Press Ctrl+C to stop the server"))
	fmt.Println(yellow("--------------------------------------------------"))

	// Start the server (blocks until shutdown)
	err := server.Start()
	if err != nil {
		return wrapError("server error: %w", err)
	}

	return nil
}
