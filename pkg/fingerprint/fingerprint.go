package fingerprint

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// FingerprintEntry maps a favicon hash to a known service/application.
// Enhanced for cybersecurity reconnaissance and cyber-space mapping.
type FingerprintEntry struct {
	Hash        string   `json:"hash"`                  // MMH3 favicon hash
	Service     string   `json:"service"`               // Service/application name
	Category    string   `json:"category,omitempty"`    // Category (CMS, Server, Framework, etc.)
	Version     string   `json:"version,omitempty"`     // Version info if known (e.g. "7.x", "2.x-3.x")
	Description string   `json:"description,omitempty"` // Brief description of the service
	Tags        []string `json:"tags,omitempty"`        // Searchable tags (e.g. ["java", "wiki", "enterprise"])
	Reference   string   `json:"reference,omitempty"`   // Reference URL for the service
	IconURL     string   `json:"icon_url,omitempty"`    // Known favicon URL for verification
	Source      string   `json:"source,omitempty"`      // Data source (shodan, fofa, community, etc.)
}

// DB provides fingerprint lookup capabilities
type DB struct {
	mu     sync.RWMutex
	byHash map[string][]FingerprintEntry
	byCat  map[string][]FingerprintEntry // index by category
	all    []FingerprintEntry
}

// NewDB creates a fingerprint database from the given entries
func NewDB(entries []FingerprintEntry) *DB {
	db := &DB{
		byHash: make(map[string][]FingerprintEntry),
		byCat:  make(map[string][]FingerprintEntry),
	}
	for _, e := range entries {
		db.addEntry(e)
	}
	return db
}

// addEntry adds a single entry to all indexes
func (db *DB) addEntry(e FingerprintEntry) {
	db.byHash[e.Hash] = append(db.byHash[e.Hash], e)
	cat := e.Category
	if cat == "" {
		cat = "Uncategorized"
	}
	db.byCat[cat] = append(db.byCat[cat], e)
	db.all = append(db.all, e)
}

// IsEmbedded returns true if the built-in fingerprint database is embedded
// in the binary (compiled with -tags=embed_fingerprints).
func IsEmbedded() bool {
	return len(DefaultFingerprints) > 0
}

// DefaultDBPath returns the default path where the external fingerprint
// database should be stored (~/.iconhash/fingerprints.json).
func DefaultDBPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".iconhash", "fingerprints.json")
}

// DefaultDB returns a DB with built-in fingerprints.
// In lightweight mode (no embedded fingerprints), it attempts to load
// from the default external path or the ICONHASH_FINGERPRINT_DB env var.
func DefaultDB() *DB {
	db := NewDB(DefaultFingerprints)

	// If embedded data is empty (lightweight build), try to load from external file
	if len(DefaultFingerprints) == 0 {
		if path := resolveFingerprintDBPath(); path != "" {
			if _, err := os.Stat(path); err == nil {
				db.LoadFromJSON(path)
			}
		}
	}

	return db
}

// resolveFingerprintDBPath resolves the fingerprint database path from:
// 1. ICONHASH_FINGERPRINT_DB environment variable
// 2. Default path (~/.iconhash/fingerprints.json)
func resolveFingerprintDBPath() string {
	// Check environment variable first
	if envPath := os.Getenv("ICONHASH_FINGERPRINT_DB"); envPath != "" {
		return envPath
	}
	// Fall back to default path
	return DefaultDBPath()
}

// Lookup finds services matching the given hash (exact match)
func (db *DB) Lookup(hash string) []FingerprintEntry {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.byHash[hash]
}

// LookupByCategory returns all fingerprints in a given category
func (db *DB) LookupByCategory(category string) []FingerprintEntry {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.byCat[category]
}

// Search performs a fuzzy search across service names, categories, tags, and descriptions.
// The query is matched case-insensitively as a substring.
func (db *DB) Search(query string) []FingerprintEntry {
	db.mu.RLock()
	defer db.mu.RUnlock()

	query = strings.ToLower(query)
	var results []FingerprintEntry
	for _, e := range db.all {
		if matchEntry(e, query) {
			results = append(results, e)
		}
	}
	return results
}

// matchEntry checks if a FingerprintEntry matches the query
func matchEntry(e FingerprintEntry, query string) bool {
	if strings.Contains(strings.ToLower(e.Service), query) {
		return true
	}
	if strings.Contains(strings.ToLower(e.Category), query) {
		return true
	}
	if strings.Contains(strings.ToLower(e.Description), query) {
		return true
	}
	if strings.Contains(strings.ToLower(e.Version), query) {
		return true
	}
	for _, tag := range e.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

// Categories returns a sorted list of all categories with their counts
func (db *DB) Categories() map[string]int {
	db.mu.RLock()
	defer db.mu.RUnlock()

	result := make(map[string]int, len(db.byCat))
	for cat, entries := range db.byCat {
		result[cat] = len(entries)
	}
	return result
}

// LoadFromJSON loads fingerprints from a JSON file and merges them into the database
func (db *DB) LoadFromJSON(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read fingerprint file: %w", err)
	}
	var entries []FingerprintEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("failed to parse fingerprint JSON: %w", err)
	}
	db.mu.Lock()
	for _, e := range entries {
		db.addEntry(e)
	}
	db.mu.Unlock()
	return nil
}

// LoadFromData loads fingerprints from raw JSON bytes and merges them into the database
func (db *DB) LoadFromData(data []byte) error {
	var entries []FingerprintEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("failed to parse fingerprint JSON: %w", err)
	}
	db.mu.Lock()
	for _, e := range entries {
		db.addEntry(e)
	}
	db.mu.Unlock()
	return nil
}

// All returns all fingerprint entries
func (db *DB) All() []FingerprintEntry {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.all
}

// Count returns the total number of unique hashes in the database
func (db *DB) Count() int {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return len(db.byHash)
}

// TotalEntries returns the total number of entries (including duplicates by hash)
func (db *DB) TotalEntries() int {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return len(db.all)
}

// ExportJSON exports all fingerprints as JSON
func (db *DB) ExportJSON() ([]byte, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// Sort by category then service for consistent output
	sorted := make([]FingerprintEntry, len(db.all))
	copy(sorted, db.all)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Category != sorted[j].Category {
			return sorted[i].Category < sorted[j].Category
		}
		return sorted[i].Service < sorted[j].Service
	})

	return json.MarshalIndent(sorted, "", "  ")
}

// ExportCSV exports all fingerprints as CSV
func (db *DB) ExportCSV() string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString("hash,service,category,version,description,tags,reference\n")
	for _, e := range db.all {
		tags := strings.Join(e.Tags, ";")
		desc := strings.ReplaceAll(e.Description, ",", "，")
		sb.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s\n",
			e.Hash, e.Service, e.Category, e.Version, desc, tags, e.Reference))
	}
	return sb.String()
}

// ExportJSONToFile exports all fingerprints as formatted JSON and writes to the given file path.
func (db *DB) ExportJSONToFile(path string) error {
	data, err := db.ExportJSON()
	if err != nil {
		return fmt.Errorf("error exporting JSON: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}
	return nil
}

// ExportCSVToFile exports all fingerprints as CSV and writes to the given file path.
func (db *DB) ExportCSVToFile(path string) error {
	csvData := db.ExportCSV()
	if err := os.WriteFile(path, []byte(csvData), 0644); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}
	return nil
}

// LookupByTag returns all fingerprints that contain the given tag (case-insensitive match).
func (db *DB) LookupByTag(tag string) []FingerprintEntry {
	tagLower := strings.ToLower(tag)
	db.mu.RLock()
	defer db.mu.RUnlock()

	var results []FingerprintEntry
	for _, e := range db.all {
		for _, t := range e.Tags {
			if strings.ToLower(t) == tagLower {
				results = append(results, e)
				break
			}
		}
	}
	return results
}

// FilterOptions specifies criteria for filtering fingerprint entries.
type FilterOptions struct {
	Category string // Exact category match (case-insensitive)
	Service  string // Substring match on service name (case-insensitive)
	Tag      string // Tag match (case-insensitive)
}

// Filter returns entries matching all non-empty criteria in the FilterOptions.
// This is the SDK equivalent of the "iconhash fingerprints --category X --service Y --tag Z" CLI command.
func (db *DB) Filter(opts FilterOptions) []FingerprintEntry {
	// If only --category is specified, use the efficient index
	if opts.Category != "" && opts.Service == "" && opts.Tag == "" {
		return db.LookupByCategory(opts.Category)
	}

	var entries []FingerprintEntry
	if opts.Service != "" || opts.Tag != "" {
		query := opts.Service
		if query == "" {
			query = opts.Tag
		}
		entries = db.Search(query)
	} else {
		entries = db.All()
	}

	// Apply all filters
	var filtered []FingerprintEntry
	for _, e := range entries {
		if opts.Category != "" && !strings.EqualFold(e.Category, opts.Category) {
			continue
		}
		if opts.Service != "" && !strings.Contains(strings.ToLower(e.Service), strings.ToLower(opts.Service)) {
			continue
		}
		if opts.Tag != "" {
			found := false
			for _, t := range e.Tags {
				if strings.EqualFold(t, opts.Tag) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		filtered = append(filtered, e)
	}
	return filtered
}
