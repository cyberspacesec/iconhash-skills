package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/cyberspacesec/iconhash-skills/pkg/fingerprint"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	fpCategory string
	fpService  string
	fpAll      bool
	fpTag      string
	fpExport   string
)

// NewFingerprintsCommand creates the fingerprints command
func NewFingerprintsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fingerprints",
		Short: "List and search the fingerprint database",
		Long: `List and search the built-in fingerprint database.

This command allows you to browse the fingerprint database, filter by category,
service name, or tags, and export the results.

The database includes fingerprints for popular services like Jenkins, GitLab,
WordPress, Grafana, Kubernetes, and many more — covering international services
and Chinese enterprise systems (BaoTa, Ruiyi, Zhiyuan OA, etc.).

You can load a custom fingerprint database using the --fingerprint-db flag.

Examples:
  iconhash fingerprints                              # Show database statistics
  iconhash fingerprints --all                        # List all fingerprints
  iconhash fingerprints --category CMS               # Filter by category
  iconhash fingerprints --service jenkins            # Search by service name
  iconhash fingerprints --tag chinese                # Filter by tag
  iconhash fingerprints --all --format json          # Export as JSON
  iconhash fingerprints --export fingerprints.json   # Export to file
  iconhash fingerprints --all --fingerprint-db custom.json
  iconhash fingerprints update                       # Update from remote source`,
		RunE: runFingerprints,
	}

	cmd.AddCommand(NewFingerprintsUpdateCommand())

	cmd.Flags().StringVar(&fpCategory, "category", "", "Filter by category (CMS, Server, Framework, OA, etc.)")
	cmd.Flags().StringVar(&fpService, "service", "", "Filter by service name (case-insensitive partial match)")
	cmd.Flags().BoolVar(&fpAll, "all", false, "List all fingerprints (default: show statistics only)")
	cmd.Flags().StringVar(&fpTag, "tag", "", "Filter by tag (e.g. chinese, java, enterprise)")
	cmd.Flags().StringVar(&fpExport, "export", "", "Export fingerprints to file (supports .json and .csv)")

	SilenceUsageOnError(cmd)

	return cmd
}

// runFingerprints handles the fingerprints command execution
func runFingerprints(cmd *cobra.Command, args []string) error {
	// Get the fingerprint database
	db := loadFingerprintDB()

	boldCyan := color.New(color.FgCyan, color.Bold)
	boldGreen := color.New(color.FgGreen, color.Bold)
	boldYellow := color.New(color.FgYellow, color.Bold)

	// If --export is specified, export and exit
	if fpExport != "" {
		return exportFingerprints(db, fpExport)
	}

	// If no filters and not --all, show statistics
	if !fpAll && fpCategory == "" && fpService == "" && fpTag == "" {
		showFingerprintStats(db, boldCyan, boldGreen)
		return nil
	}

	// Get and filter entries
	entries := filterFingerprints(db)

	if len(entries) == 0 {
		boldYellow.Println("⚠️  No fingerprints found matching your criteria.")
		return nil
	}

	// Output based on format
	if OutputFormat == "json" {
		data, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			return wrapError("error marshaling JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	if OutputFormat == "csv" {
		fmt.Println("hash,service,category,version,tags")
		for _, e := range entries {
			tags := strings.Join(e.Tags, ";")
			fmt.Printf("%s,%s,%s,%s,%s\n", e.Hash, e.Service, e.Category, e.Version, tags)
		}
		return nil
	}

	// Default text output
	boldCyan.Printf("📋 Found %d fingerprint(s)\n\n", len(entries))
	for _, e := range entries {
		boldCyan.Printf("  🏷️  %s", e.Service)
		if e.Category != "" {
			boldYellow.Printf(" [%s]", e.Category)
		}
		fmt.Printf(" → %s\n", e.Hash)
		if e.Description != "" {
			fmt.Printf("      %s\n", e.Description)
		}
		if len(e.Tags) > 0 {
			fmt.Printf("      Tags: %s\n", strings.Join(e.Tags, ", "))
		}
	}

	return nil
}

// showFingerprintStats displays database statistics
func showFingerprintStats(db *fingerprint.DB, boldCyan, boldGreen *color.Color) {
	categories := db.Categories()

	// Sort categories by count
	type catEntry struct {
		name  string
		count int
	}
	var sorted []catEntry
	for cat, count := range categories {
		sorted = append(sorted, catEntry{cat, count})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	boldCyan.Println("📊 Fingerprint Database Statistics")
	fmt.Println()
	boldGreen.Printf("  Total unique hashes: %d\n", db.Count())
	boldGreen.Printf("  Total entries: %d\n", db.TotalEntries())
	fmt.Println()

	boldCyan.Println("  Categories:")
	for _, e := range sorted {
		fmt.Printf("    %-24s %d\n", e.name, e.count)
	}

	fmt.Println()
	boldCyan.Println("  Usage:")
	fmt.Println("    iconhash fingerprints --all                # List all fingerprints")
	fmt.Println("    iconhash fingerprints --category CMS       # Filter by category")
	fmt.Println("    iconhash fingerprints --service jenkins    # Search by service name")
	fmt.Println("    iconhash fingerprints --tag chinese        # Filter by tag")
	fmt.Println("    iconhash fingerprints --export out.json    # Export to file")
}

// filterFingerprints returns filtered fingerprint entries
func filterFingerprints(db *fingerprint.DB) []fingerprint.FingerprintEntry {
	// If only --category is specified, use the efficient index
	if fpCategory != "" && fpService == "" && fpTag == "" {
		return db.LookupByCategory(fpCategory)
	}

	// Otherwise, use search or filter all
	var entries []fingerprint.FingerprintEntry
	if fpService != "" || fpTag != "" {
		// Build a combined query
		if fpService != "" && fpTag != "" {
			// Search by service first, then filter by tag
			entries = db.Search(fpService)
			var filtered []fingerprint.FingerprintEntry
			for _, e := range entries {
				if hasTag(e, fpTag) {
					filtered = append(filtered, e)
				}
			}
			entries = filtered
		} else if fpTag != "" {
			// Search by tag only
			entries = db.Search(fpTag)
		} else {
			entries = db.Search(fpService)
		}
	} else if fpAll {
		entries = db.All()
	} else {
		entries = db.All()
	}

	// Apply category filter if specified
	if fpCategory != "" {
		var filtered []fingerprint.FingerprintEntry
		for _, e := range entries {
			if strings.EqualFold(e.Category, fpCategory) {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	return entries
}

// hasTag checks if an entry has a specific tag (case-insensitive)
func hasTag(e fingerprint.FingerprintEntry, tag string) bool {
	tag = strings.ToLower(tag)
	for _, t := range e.Tags {
		if strings.ToLower(t) == tag {
			return true
		}
	}
	return false
}

// exportFingerprints exports the fingerprint database to a file
func exportFingerprints(db *fingerprint.DB, path string) error {
	if strings.HasSuffix(path, ".csv") {
		csvData := db.ExportCSV()
		if err := os.WriteFile(path, []byte(csvData), 0644); err != nil {
			return wrapError("error exporting CSV: %w", err)
		}
	} else {
		// Default to JSON
		jsonData, err := db.ExportJSON()
		if err != nil {
			return wrapError("error exporting JSON: %w", err)
		}
		if err := os.WriteFile(path, jsonData, 0644); err != nil {
			return wrapError("error writing file: %w", err)
		}
	}

	boldGreen := color.New(color.FgGreen, color.Bold)
	boldGreen.Printf("✅ Exported %d fingerprints to %s\n", db.TotalEntries(), path)
	return nil
}
