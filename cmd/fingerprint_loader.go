package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cyberspacesec/iconhash-skills/pkg/fingerprint"
	"github.com/fatih/color"
)

const (
	// Multiple download sources for resilience
	defaultFingerprintDownloadURLs = "https://raw.githubusercontent.com/cyberspacesec/iconhash-skills/main/data/fingerprints.json"
	fingerprintMirrorURL           = "https://ghp.ci/https://raw.githubusercontent.com/cyberspacesec/iconhash-skills/main/data/fingerprints.json"
	fingerprintVersionURL          = "https://raw.githubusercontent.com/cyberspacesec/iconhash-skills/main/data/fingerprints.version.json"
	localFingerprintDir            = ".iconhash"
	localFingerprintFile           = "fingerprints.json"
	localFingerprintVersionFile    = "fingerprints.version.json"
)

// FingerprintVersion tracks the version of the fingerprint database
type FingerprintVersion struct {
	Version     string `json:"version"`
	HashCount   int    `json:"hash_count"`
	EntryCount  int    `json:"entry_count"`
	UpdatedAt   string `json:"updated_at"`
	SHA256      string `json:"sha256"`
	Description string `json:"description"`
}

// loadFingerprintDB loads the fingerprint database with the following priority:
// 1. If FingerprintDB flag is set, use it
// 2. If ICONHASH_FINGERPRINT_DB env var is set, use it
// 3. If embedded fingerprints are available, use them
// 4. If local ~/.iconhash/fingerprints.json exists, use it
// 5. Otherwise, try to auto-download
func loadFingerprintDB() *fingerprint.DB {
	db := fingerprint.DefaultDB()

	// Priority 1: Explicit --fingerprint-db flag
	if FingerprintDB != "" {
		if err := db.LoadFromJSON(FingerprintDB); err != nil {
			// Return a DB with error info — the caller can check
			fmt.Fprintf(os.Stderr, "❌ Error loading fingerprint database: %v\n", err)
			return db
		}
		if Debug {
			fmt.Fprintf(os.Stderr, "📂 Loaded custom fingerprint database: %s\n", FingerprintDB)
		}
		return db
	}

	// If embedded data is available, we're good
	if fingerprint.IsEmbedded() {
		return db
	}

	// Priority 2: ICONHASH_FINGERPRINT_DB env var
	if envPath := os.Getenv("ICONHASH_FINGERPRINT_DB"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			if err := db.LoadFromJSON(envPath); err == nil {
				if Debug {
					fmt.Fprintf(os.Stderr, "📂 Loaded fingerprint database from env: %s\n", envPath)
				}
				return db
			}
		}
	}

	// Priority 3: Local default path (~/.iconhash/fingerprints.json)
	localPath := fingerprint.DefaultDBPath()
	if localPath != "" {
		if _, err := os.Stat(localPath); err == nil {
			if err := db.LoadFromJSON(localPath); err == nil {
				if Debug {
					fmt.Fprintf(os.Stderr, "📂 Loaded local fingerprint database: %s\n", localPath)
				}
				return db
			}
		}
	}

	// Priority 4: Auto-download for lightweight builds
	if db.Count() == 0 {
		boldYellow := color.New(color.FgYellow, color.Bold)
		boldCyan := color.New(color.FgCyan, color.Bold)

		boldYellow.Println("\n⚠️  No fingerprint database found!")
		fmt.Println("   This is a lightweight build without embedded fingerprints.")
		boldCyan.Println("   🔄 Auto-downloading fingerprint database...")

		if err := downloadFingerprints(localPath, true); err != nil {
			color.Red("   ❌ Download failed: %v", err)
			fmt.Println("   You can manually download with: iconhash fingerprints update")
			return db
		}

		// Reload the DB after download
		if localPath != "" {
			db.LoadFromJSON(localPath)
		}
	}

	return db
}

// downloadFingerprints downloads the fingerprint database from the best available source
func downloadFingerprints(destPath string, merge bool) error {
	if destPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot determine home directory: %w", err)
		}
		destPath = filepath.Join(homeDir, localFingerprintDir, localFingerprintFile)
	}

	// Try multiple sources
	urls := []string{defaultFingerprintDownloadURLs, fingerprintMirrorURL}
	var body []byte
	var usedURL string

	for _, url := range urls {
		data, err := downloadWithProgress(url)
		if err != nil {
			if Debug {
				fmt.Fprintf(os.Stderr, "   Source %s failed: %v\n", url, err)
			}
			continue
		}
		body = data
		usedURL = url
		break
	}

	if body == nil {
		return fmt.Errorf("all download sources failed")
	}

	// Validate the downloaded data
	var entries []fingerprint.FingerprintEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return fmt.Errorf("invalid fingerprint data: %w", err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("empty fingerprint database")
	}

	// Compute SHA256 for verification
	hash := sha256.Sum256(body)
	sha256Str := hex.EncodeToString(hash[:])
	if Debug {
		fmt.Fprintf(os.Stderr, "   SHA256: %s\n", sha256Str)
	}

	// Create directory
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("cannot create directory %s: %w", destDir, err)
	}

	// Merge or replace
	if merge {
		if _, err := os.Stat(destPath); err == nil {
			// Existing file - merge new data into it
			if err := mergeFingerprintData(destPath, body); err != nil {
				// If merge fails, just overwrite
				if err := os.WriteFile(destPath, body, 0644); err != nil {
					return fmt.Errorf("write failed: %w", err)
				}
			}
		} else {
			// No existing file - just write
			if err := os.WriteFile(destPath, body, 0644); err != nil {
				return fmt.Errorf("write failed: %w", err)
			}
		}
	} else {
		if err := os.WriteFile(destPath, body, 0644); err != nil {
			return fmt.Errorf("write failed: %w", err)
		}
	}

	// Save version info
	saveVersionInfo(destPath, entries, sha256Str, usedURL)

	boldGreen := color.New(color.FgGreen, color.Bold)
	boldGreen.Printf("   ✅ Downloaded %d fingerprints from %s\n", len(entries), shortenURL(usedURL))

	return nil
}

// downloadWithProgress downloads from a URL with progress indication
func downloadWithProgress(url string) ([]byte, error) {
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Read with progress for large files
	total := resp.ContentLength
	var body []byte
	buf := make([]byte, 32*1024)
	var downloaded int64
	lastPct := -1

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			body = append(body, buf[:n]...)
			downloaded += int64(n)
			if total > 0 {
				pct := int(float64(downloaded) / float64(total) * 100)
				if pct != lastPct && pct%20 == 0 {
					fmt.Printf("   📥 Downloading... %d%%\n", pct)
					lastPct = pct
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return body, nil
}

// mergeFingerprintData merges new fingerprint data into an existing file
func mergeFingerprintData(destPath string, newData []byte) error {
	// Read existing data
	existingData, err := os.ReadFile(destPath)
	if err != nil {
		return err
	}

	var existingEntries []fingerprint.FingerprintEntry
	if err := json.Unmarshal(existingData, &existingEntries); err != nil {
		return err
	}

	var newEntries []fingerprint.FingerprintEntry
	if err := json.Unmarshal(newData, &newEntries); err != nil {
		return err
	}

	// Build a set of existing hashes for dedup
	existingHashes := make(map[string]bool)
	for _, e := range existingEntries {
		key := e.Hash + "|" + e.Service
		existingHashes[key] = true
	}

	// Add only new entries
	merged := existingEntries
	added := 0
	for _, e := range newEntries {
		key := e.Hash + "|" + e.Service
		if !existingHashes[key] {
			merged = append(merged, e)
			added++
		}
	}

	// Write merged data
	mergedData, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(destPath, mergedData, 0644); err != nil {
		return err
	}

	if Debug && added > 0 {
		fmt.Fprintf(os.Stderr, "   Merged: added %d new entries (total: %d)\n", added, len(merged))
	}

	return nil
}

// saveVersionInfo saves version information for the fingerprint database
func saveVersionInfo(destPath string, entries []fingerprint.FingerprintEntry, sha256Str string, sourceURL string) {
	// Count unique hashes
	hashSet := make(map[string]bool)
	for _, e := range entries {
		hashSet[e.Hash] = true
	}

	version := FingerprintVersion{
		Version:     time.Now().Format("20060102"),
		HashCount:   len(hashSet),
		EntryCount:  len(entries),
		UpdatedAt:   time.Now().Format(time.RFC3339),
		SHA256:      sha256Str,
		Description: fmt.Sprintf("go-iconhash fingerprint database (%d hashes)", len(hashSet)),
	}

	versionPath := strings.Replace(destPath, localFingerprintFile, localFingerprintVersionFile, 1)
	versionData, _ := json.MarshalIndent(version, "", "  ")
	os.WriteFile(versionPath, versionData, 0644)
}

// loadLocalVersionInfo loads the local version info if it exists
func loadLocalVersionInfo() *FingerprintVersion {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	versionPath := filepath.Join(homeDir, localFingerprintDir, localFingerprintVersionFile)
	data, err := os.ReadFile(versionPath)
	if err != nil {
		return nil
	}
	var version FingerprintVersion
	if err := json.Unmarshal(data, &version); err != nil {
		return nil
	}
	return &version
}

// checkForUpdates checks if a newer version of the fingerprint database is available
func checkForUpdates() (*FingerprintVersion, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(fingerprintVersionURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var remoteVersion FingerprintVersion
	if err := json.NewDecoder(resp.Body).Decode(&remoteVersion); err != nil {
		return nil, err
	}

	return &remoteVersion, nil
}

// shortenURL returns a shorter version of the URL for display
func shortenURL(url string) string {
	if strings.HasPrefix(url, "https://raw.githubusercontent.com/") {
		parts := strings.Split(url, "/")
		if len(parts) >= 5 {
			return fmt.Sprintf("GitHub/%s/%s", parts[4], parts[6])
		}
	}
	if len(url) > 40 {
		return url[:40] + "..."
	}
	return url
}
