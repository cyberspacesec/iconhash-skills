package cmd

import (
	"fmt"
	"strings"

	"github.com/cyberspacesec/iconhash-skills/pkg/util"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// NewLookupCommand creates the lookup command
func NewLookupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lookup [hash]",
		Short: "Lookup a favicon hash against the fingerprint database",
		Long: `Lookup a favicon hash against the built-in fingerprint database.

This command checks the provided MMH3 favicon hash against a database of
well-known service fingerprints. This helps identify what service or application
is associated with a given favicon hash.

The output includes search syntax for all major cyber-space mapping platforms
(Fofa, Shodan, Censys, Quake, ZoomEye, Hunter) so you can directly
copy-paste to search for similar assets.

Note: Negative hashes start with "-" which Cobra may interpret as a flag.
Use "--" before the hash or use the --hash flag instead.

Examples:
  iconhash lookup -- -305179312
  iconhash lookup 81586312
  iconhash lookup --fingerprint-db custom_fingerprints.json -- -305179312
  iconhash lookup -- -305179312 --engine fofa`,
		RunE: runLookup,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("hash value is required. Provide it as an argument: iconhash lookup -- <hash>")
			}
			return nil
		},
	}

	// Disable interspersed args so negative hashes (e.g. -305179312) aren't
	// interpreted as flags. Users can still pass flags BEFORE the hash, or
	// use "--" to separate flags from positional args.
	cmd.Flags().SetInterspersed(false)

	SilenceUsageOnError(cmd)

	return cmd
}

// runLookup handles the lookup command execution
func runLookup(cmd *cobra.Command, args []string) error {
	hashValue := args[0]

	// Get the fingerprint database
	db := loadFingerprintDB()

	// Lookup the hash
	matches := db.Lookup(hashValue)

	boldCyan := color.New(color.FgCyan, color.Bold)
	boldGreen := color.New(color.FgGreen, color.Bold)
	boldYellow := color.New(color.FgYellow, color.Bold)

	boldCyan.Printf("🔍 Looking up hash: ")
	fmt.Println(hashValue)
	fmt.Println()

	if len(matches) == 0 {
		boldYellow.Println("⚠️  No known services found for this hash.")
		fmt.Println()
		fmt.Println("This hash doesn't match any known service in the fingerprint database.")
	} else {
		boldGreen.Printf("✅ Found %d match(es)\n\n", len(matches))

		// Display matches with rich info
		for _, m := range matches {
			boldCyan.Printf("  🏷️  Service: ")
			fmt.Println(m.Service)
			if m.Category != "" {
				boldYellow.Printf("  📂 Category: ")
				fmt.Println(m.Category)
			}
			if m.Version != "" {
				boldYellow.Printf("  🔖 Version: ")
				fmt.Println(m.Version)
			}
			if m.Description != "" {
				fmt.Printf("  📝 %s\n", m.Description)
			}
			if len(m.Tags) > 0 {
				fmt.Printf("  🏷️  Tags: %s\n", formatTags(m.Tags))
			}
			fmt.Printf("  🔢 Hash: %s\n", m.Hash)
			fmt.Println()
		}
	}

	// Always show search engine syntax — this is the core use case for cyber-space mapping
	boldCyan.Println("🔎 Search Engine Queries (Cyber-Space Mapping):")
	fmt.Println()

	queries := util.FormatAll(hashValue)
	engineNames := []string{"Fofa", "Shodan", "Censys", "Quake", "ZoomEye", "Hunter"}
	for _, name := range engineNames {
		fmt.Printf("  %-10s %s\n", name+":", queries[strings.ToLower(name)])
	}

	return nil
}

// formatTags formats tags as colored string
func formatTags(tags []string) string {
	var result string
	for i, tag := range tags {
		if i > 0 {
			result += ", "
		}
		result += tag
	}
	return result
}
