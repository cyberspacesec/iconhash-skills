package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cyberspacesec/iconhash-skills/pkg/fingerprint"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	// DefaultFingerprintRepoURL is the default remote source for fingerprint updates
	// Points to the fingerprints.json in the go-iconhash repository's data/ directory
	DefaultFingerprintRepoURL = "https://raw.githubusercontent.com/cyberspacesec/iconhash-skills/main/data/fingerprints.json"
	// DefaultLocalDBDir is the default local directory for downloaded fingerprints
	DefaultLocalDBDir = ".iconhash"
	// DefaultLocalDBFile is the default local filename for downloaded fingerprints
	DefaultLocalDBFile = "fingerprints.json"
)

// NewFingerprintsUpdateCommand creates the fingerprints update subcommand
func NewFingerprintsUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update fingerprint database from remote source",
		Long: `Update fingerprint database from a remote source.

This command downloads the latest community-maintained fingerprint database
and saves it locally for use with the --fingerprint-db flag or ICONHASH_FINGERPRINT_DB env var.

Features:
  - Multiple download sources for resilience (GitHub + mirror)
  - SHA256 verification of downloaded data
  - Incremental merge with existing database (no duplicate entries)
  - Version tracking for update checking
  - Progress indication during download

The default remote source is the official go-iconhash repository on GitHub.
You can also specify a custom URL.

Examples:
  iconhash fingerprints update                           # Update from default source
  iconhash fingerprints update -o my_fingerprints.json   # Save to custom file
  iconhash fingerprints update --url https://example.com/fingerprints.json
  iconhash fingerprints update --replace                 # Replace instead of merge`,
		RunE: runFingerprintsUpdate,
	}

	cmd.Flags().StringVarP(&fpUpdateOutput, "output", "o", "", "Output file path (default: ~/.iconhash/fingerprints.json)")
	cmd.Flags().StringVar(&fpUpdateURL, "url", "", "Custom URL to download fingerprints from")
	cmd.Flags().BoolVar(&fpUpdateReplace, "replace", false, "Replace existing database instead of merging")

	SilenceUsageOnError(cmd)

	return cmd
}

// runFingerprintsUpdate handles the fingerprints update command execution
func runFingerprintsUpdate(cmd *cobra.Command, args []string) error {
	boldCyan := color.New(color.FgCyan, color.Bold)
	boldGreen := color.New(color.FgGreen, color.Bold)

	// Determine the output path
	outputPath := fpUpdateOutput
	if outputPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return wrapError("error getting home directory: %w", err)
		}
		outputPath = filepath.Join(homeDir, DefaultLocalDBDir, DefaultLocalDBFile)
	}

	// If custom URL is specified, download from that URL only
	if fpUpdateURL != "" {
		boldCyan.Println("🔄 Updating fingerprint database...")
		fmt.Printf("  Source: %s\n", fpUpdateURL)
		fmt.Printf("  Target: %s\n", outputPath)
		fmt.Println()

		if err := downloadFromCustomURL(fpUpdateURL, outputPath, !fpUpdateReplace); err != nil {
			return wrapError("download failed: %w", err)
		}

		showUpdateSummary(outputPath)
		return nil
	}

	// Default update flow
	boldCyan.Println("🔄 Updating fingerprint database...")
	fmt.Printf("  Target: %s\n", outputPath)
	fmt.Println()

	// Check local version first
	localVersion := loadLocalVersionInfo()
	if localVersion != nil {
		fmt.Printf("  Current version: %s (%d hashes, updated %s)\n",
			localVersion.Version, localVersion.HashCount, localVersion.UpdatedAt)
	}

	// Try to check for remote version
	fmt.Println("  Checking for updates...")
	remoteVersion, err := checkForUpdates()
	if err == nil && remoteVersion != nil {
		fmt.Printf("  Remote version:  %s (%d hashes, updated %s)\n",
			remoteVersion.Version, remoteVersion.HashCount, remoteVersion.UpdatedAt)

		// Check if we need to update
		if localVersion != nil && localVersion.Version == remoteVersion.Version && localVersion.HashCount == remoteVersion.HashCount {
			boldGreen.Println("\n✅ Fingerprint database is already up to date!")
			return nil
		}
	} else if Debug {
		fmt.Fprintf(os.Stderr, "  Version check failed: %v\n", err)
	}

	// Download
	fmt.Println("  Downloading...")
	if err := downloadFingerprints(outputPath, !fpUpdateReplace); err != nil {
		return wrapError("download failed: %w", err)
	}

	showUpdateSummary(outputPath)

	return nil
}

// downloadFromCustomURL downloads fingerprints from a custom URL
func downloadFromCustomURL(url string, destPath string, merge bool) error {
	data, err := downloadWithProgress(url)
	if err != nil {
		return err
	}

	// Validate
	var entries []fingerprint.FingerprintEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("invalid fingerprint data: %w", err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("empty fingerprint database")
	}

	// Create directory
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("cannot create directory: %w", err)
	}

	// Merge or write
	if merge {
		if _, err := os.Stat(destPath); err == nil {
			if err := mergeFingerprintData(destPath, data); err != nil {
				os.WriteFile(destPath, data, 0644)
			}
		} else {
			os.WriteFile(destPath, data, 0644)
		}
	} else {
		os.WriteFile(destPath, data, 0644)
	}

	return nil
}

// showUpdateSummary displays the summary after an update
func showUpdateSummary(outputPath string) {
	boldGreen := color.New(color.FgGreen, color.Bold)
	boldCyan := color.New(color.FgCyan, color.Bold)

	// Load the updated database and show stats
	db := fingerprint.NewDB(nil)
	if err := db.LoadFromJSON(outputPath); err != nil {
		// Try loading with default embedded data + external
		db = fingerprint.DefaultDB()
		db.LoadFromJSON(outputPath)
	}

	fmt.Println()
	boldGreen.Println("✅ Fingerprint database updated successfully!")
	fmt.Println()
	fmt.Printf("  Unique hashes: %d\n", db.Count())
	fmt.Printf("  Total entries: %d\n", db.TotalEntries())
	fmt.Printf("  Saved to:      %s\n", outputPath)

	// Show version info
	version := loadLocalVersionInfo()
	if version != nil {
		fmt.Printf("  Version:       %s\n", version.Version)
		fmt.Printf("  SHA256:        %s...\n", version.SHA256[:16])
	}

	fmt.Println()
	boldCyan.Println("💡 Usage:")
	fmt.Printf("   iconhash lookup -- -305179312 --fingerprint-db %s\n", outputPath)
	fmt.Printf("   iconhash identify https://example.com --fingerprint-db %s\n", outputPath)
	fmt.Println()
	boldYellow := color.New(color.FgYellow, color.Bold)
	boldYellow.Println("💡 Auto-load (recommended):")
	fmt.Printf("   export ICONHASH_FINGERPRINT_DB=%s\n", outputPath)

	// If embedded mode, show comparison
	if fingerprint.IsEmbedded() {
		embeddedDB := fingerprint.DefaultDB()
		fmt.Println()
		fmt.Printf("  Built-in hashes: %d | Downloaded hashes: %d | Combined: %d\n",
			embeddedDB.Count(), db.Count(), db.Count())
		fmt.Println("  (Downloaded data is merged with built-in data)")
	}
}