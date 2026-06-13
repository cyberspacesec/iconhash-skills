package cmd

import (
	"context"
	"fmt"
	"sort"

	"github.com/cyberspacesec/iconhash-skills/pkg/hasher"
	"github.com/cyberspacesec/iconhash-skills/pkg/util"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// NewIdentifyCommand creates the identify command
func NewIdentifyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "identify [url]",
		Short: "Identify a website by discovering its favicon and matching against fingerprints",
		Long: `Identify a website by discovering its favicon and matching against fingerprints.

This is the all-in-one command for cyber-space mapping reconnaissance. It:
1. Discovers all favicon URLs on the target website
2. Calculates the MMH3 hash for each favicon
3. Looks up each hash against the fingerprint database
4. Outputs search engine queries for all major platforms (Fofa, Shodan, Censys, Quake, ZoomEye, Hunter)

Examples:
  iconhash identify https://example.com
  iconhash identify https://gitlab.example.com --engine fofa
  iconhash identify https://jenkins.example.com --fingerprint-db custom.json`,
		RunE: runIdentify,
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

// runIdentify handles the identify command execution
func runIdentify(cmd *cobra.Command, args []string) error {
	// Create hasher
	options := &hasher.HashOptions{
		UseUint32:          Uint32Flag,
		RequestTimeout:     Timeout,
		InsecureSkipVerify: SkipVerify,
		UserAgent:          UserAgent,
	}
	h := hasher.New(options)

	// Get fingerprint database
	db := loadFingerprintDB()

	boldCyan := color.New(color.FgCyan, color.Bold)
	boldGreen := color.New(color.FgGreen, color.Bold)
	boldYellow := color.New(color.FgYellow, color.Bold)
	boldRed := color.New(color.FgRed, color.Bold)

	// Step 1: Use the SDK's Identify method (discover + hash + lookup)
	fmt.Fprintf(cmd.ErrOrStderr(), "🌐 Identifying %s...\n", URL)
	results := h.Identify(context.Background(), URL, db, nil)

	// Count successes
	successCount := 0
	for _, r := range results {
		if r.Err == nil {
			successCount++
		}
	}
	boldGreen.Printf("✅ Found %d favicon URL(s), %d identified\n\n", len(results), successCount)

	// Collect unique hashes for search engine queries
	var allHashes []string
	seenHash := make(map[string]bool)

	// Display results
	for i, result := range results {
		boldCyan.Printf("[%d/%d] ", i+1, len(results))

		if result.Err != nil {
			boldRed.Printf("❌ %s: %v\n", result.URL, result.Err)
			continue
		}

		boldYellow.Printf("%s\n", result.URL)
		fmt.Printf("  Hash: %s\n", result.Hash)

		// Track unique hashes
		if !seenHash[result.Hash] {
			seenHash[result.Hash] = true
			allHashes = append(allHashes, result.Hash)
		}

		// Show fingerprint matches
		if len(result.Matches) > 0 {
			boldGreen.Printf("  ✅ Identified: ")
			for j, m := range result.Matches {
				if j > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("%s", m.Service)
				if m.Category != "" {
					fmt.Printf(" [%s]", m.Category)
				}
				if m.Version != "" {
					fmt.Printf(" (%s)", m.Version)
				}
			}
			fmt.Println()

			// Show tags from matches
			var allTags []string
			seenTag := make(map[string]bool)
			for _, m := range result.Matches {
				for _, t := range m.Tags {
					if !seenTag[t] {
						seenTag[t] = true
						allTags = append(allTags, t)
					}
				}
			}
			if len(allTags) > 0 {
				fmt.Printf("  Tags: %s\n", formatTags(allTags))
			}
		} else {
			boldYellow.Println("  ⚠️  No fingerprint match found")
		}

		fmt.Println()
	}

	// Step 2: Output search engine queries for all unique hashes
	if len(allHashes) > 0 {
		boldCyan.Println("🔎 Search Engine Queries (Cyber-Space Mapping):")
		fmt.Println()

		queries := util.FormatAll("")
		engineNames := make([]string, 0, len(queries))
		for name := range queries {
			engineNames = append(engineNames, name)
		}
		sort.Strings(engineNames)

		for _, hash := range allHashes {
			fmt.Printf("  Hash: %s\n", hash)
			for _, name := range engineNames {
				formatted := util.FormatHash(hash, util.ParseFormat(name))
				fmt.Printf("    %-10s %s\n", name+":", formatted)
			}
			fmt.Println()
		}

		// Also show the --engine formatted output if specified
		format := getEngineFormat()
		if format != util.FormatPlain {
			boldCyan.Println("📋 Formatted Output (--engine):")
			for _, hash := range allHashes {
				fmt.Printf("  %s\n", util.FormatHash(hash, format))
			}
		}
	}

	return nil
}
