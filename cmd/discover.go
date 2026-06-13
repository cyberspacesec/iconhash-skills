package cmd

import (
	"context"
	"fmt"

	"github.com/cyberspacesec/iconhash-skills/pkg/hasher"
	"github.com/cyberspacesec/iconhash-skills/pkg/util"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// NewDiscoverCommand creates the discover command
func NewDiscoverCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discover [url]",
		Short: "Discover favicon URLs from a website and calculate their hashes",
		Long: `Discover favicon URLs from a website and calculate their hashes.

This command will scan a website for favicon references by:
1. Parsing HTML <link> tags for favicon URLs
2. Trying common favicon paths (/favicon.ico, /favicon.png, etc.)

Then it calculates the MMH3 hash for each discovered favicon.

Examples:
  iconhash discover https://example.com
  iconhash discover https://example.com --engine fofa
  iconhash discover https://example.com --uint32 --engine shodan`,
		RunE: runDiscover,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				if URL == "" {
					URL = args[0]
				}
			}
			if URL == "" {
				return fmt.Errorf("URL is required. Provide it as an argument or with --url flag")
			}
			return nil
		},
	}

	SilenceUsageOnError(cmd)

	return cmd
}

// runDiscover handles the discover command execution
func runDiscover(cmd *cobra.Command, args []string) error {
	// Create a new hasher with the options from root command
	options := &hasher.HashOptions{
		UseUint32:          Uint32Flag,
		RequestTimeout:     Timeout,
		InsecureSkipVerify: SkipVerify,
		UserAgent:          UserAgent,
	}
	h := hasher.New(options)

	// Debug info if enabled
	if Debug {
		fmt.Fprintf(cmd.ErrOrStderr(), "🔍 Discovering favicons for: %s\n", URL)
		fmt.Fprintf(cmd.ErrOrStderr(), "🔧 Options: uint32=%v, timeout=%.0fs, skip-verify=%v\n",
			options.UseUint32, options.RequestTimeout.Seconds(), options.InsecureSkipVerify)
	}

	// Get format from --engine flag
	format := getEngineFormat()

	fmt.Fprintf(cmd.ErrOrStderr(), "🌐 Discovering favicon URLs from %s...\n", URL)

	// Use the SDK's DiscoverAndHash method
	results := h.DiscoverAndHash(context.Background(), URL, nil)

	boldCyan := color.New(color.FgCyan, color.Bold)
	boldGreen := color.New(color.FgGreen, color.Bold)
	boldYellow := color.New(color.FgYellow, color.Bold)

	// Count successful discoveries
	successCount := 0
	for _, r := range results {
		if r.Err == nil {
			successCount++
		}
	}
	boldGreen.Printf("✅ Found %d favicon URL(s), %d hashed successfully\n\n", len(results), successCount)

	// Display results
	for i, result := range results {
		boldCyan.Printf("[%d/%d] ", i+1, len(results))
		boldYellow.Printf("%s\n", result.URL)

		if result.Err != nil {
			color.Red("  ❌ Error: %v\n", result.Err)
		} else {
			fmt.Printf("  Hash: %s\n", result.Hash)
			if format != util.FormatPlain {
				fmt.Printf("  Formatted: %s\n", util.FormatHash(result.Hash, format))
			}
		}
		fmt.Println()
	}

	return nil
}
