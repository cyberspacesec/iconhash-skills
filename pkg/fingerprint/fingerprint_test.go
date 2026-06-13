package fingerprint

import (
	"encoding/json"
	"os"
	"testing"
)

func TestDefaultDB(t *testing.T) {
	db := DefaultDB()
	if db == nil {
		t.Fatal("DefaultDB() returned nil")
	}

	// In embedded mode, test that built-in data is loaded
	if IsEmbedded() {
		results := db.Lookup("81586312")
		if len(results) == 0 {
			t.Error("Expected to find Jenkins fingerprint in embedded mode")
		} else if results[0].Service != "Jenkins" {
			t.Errorf("Expected Jenkins, got %s", results[0].Service)
		}
	} else {
		// In lite mode, DefaultFingerprints should be empty
		if len(DefaultFingerprints) != 0 {
			t.Errorf("Expected empty DefaultFingerprints in lite mode, got %d", len(DefaultFingerprints))
		}
	}
}

func TestLookupNotFound(t *testing.T) {
	db := NewDB(nil)
	results := db.Lookup("999999999999")
	if len(results) != 0 {
		t.Error("Expected empty results for unknown hash")
	}
}

func TestAll(t *testing.T) {
	if IsEmbedded() {
		db := DefaultDB()
		all := db.All()
		if len(all) < 40 {
			t.Errorf("Expected at least 40 fingerprints in embedded mode, got %d", len(all))
		}
	}
}

func TestCount(t *testing.T) {
	if IsEmbedded() {
		db := DefaultDB()
		count := db.Count()
		if count < 40 {
			t.Errorf("Expected at least 40 unique hashes in embedded mode, got %d", count)
		}
	}
}

func TestLoadFromJSON(t *testing.T) {
	// Create a temp JSON file with custom fingerprints
	entries := []FingerprintEntry{
		{Hash: "12345", Service: "TestService", Category: "TestCategory"},
	}
	data, _ := json.Marshal(entries)
	tmpFile, err := os.CreateTemp("", "fingerprint-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Write(data)
	tmpFile.Close()

	db := NewDB(nil)
	err = db.LoadFromJSON(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadFromJSON() error: %v", err)
	}

	results := db.Lookup("12345")
	if len(results) == 0 {
		t.Error("Expected to find custom fingerprint after LoadFromJSON")
	}
	if results[0].Service != "TestService" {
		t.Errorf("Expected TestService, got %s", results[0].Service)
	}
}

func TestLoadFromJSON_InvalidPath(t *testing.T) {
	db := NewDB(nil)
	err := db.LoadFromJSON("/nonexistent/path.json")
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

func TestLoadFromJSON_InvalidJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fingerprint-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("not valid json")
	tmpFile.Close()

	db := NewDB(nil)
	err = db.LoadFromJSON(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestSearch(t *testing.T) {
	entries := []FingerprintEntry{
		{Hash: "111", Service: "Jenkins", Category: "CI/CD", Tags: []string{"java", "automation"}},
		{Hash: "222", Service: "GitLab", Category: "CI/CD", Tags: []string{"ruby", "git"}},
		{Hash: "333", Service: "WordPress", Category: "CMS", Description: "Blog CMS"},
	}
	db := NewDB(entries)

	// Search by service name
	results := db.Search("jenkins")
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'jenkins', got %d", len(results))
	}

	// Search by category
	results = db.Search("ci/cd")
	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'ci/cd', got %d", len(results))
	}

	// Search by tag
	results = db.Search("java")
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'java', got %d", len(results))
	}

	// Search by description
	results = db.Search("blog")
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'blog', got %d", len(results))
	}
}

func TestLoadFromData(t *testing.T) {
	entries := []FingerprintEntry{
		{Hash: "99999", Service: "DataTestService", Category: "Test"},
	}
	data, _ := json.Marshal(entries)

	db := NewDB(nil)
	err := db.LoadFromData(data)
	if err != nil {
		t.Fatalf("LoadFromData() error: %v", err)
	}

	results := db.Lookup("99999")
	if len(results) == 0 {
		t.Error("Expected to find fingerprint after LoadFromData")
	}
	if results[0].Service != "DataTestService" {
		t.Errorf("Expected DataTestService, got %s", results[0].Service)
	}
}

func TestIsEmbedded(t *testing.T) {
	// IsEmbedded should return a consistent bool
	result := IsEmbedded()
	_ = result // just verify it doesn't panic
}

func TestCategories(t *testing.T) {
	entries := []FingerprintEntry{
		{Hash: "111", Service: "A", Category: "CMS"},
		{Hash: "222", Service: "B", Category: "CMS"},
		{Hash: "333", Service: "C", Category: "Server"},
	}
	db := NewDB(entries)

	cats := db.Categories()
	if cats["CMS"] != 2 {
		t.Errorf("Expected 2 CMS entries, got %d", cats["CMS"])
	}
	if cats["Server"] != 1 {
		t.Errorf("Expected 1 Server entry, got %d", cats["Server"])
	}
}

func TestExportJSON(t *testing.T) {
	entries := []FingerprintEntry{
		{Hash: "111", Service: "TestExport", Category: "Test"},
	}
	db := NewDB(entries)

	data, err := db.ExportJSON()
	if err != nil {
		t.Fatalf("ExportJSON() error: %v", err)
	}

	var exported []FingerprintEntry
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("Failed to parse exported JSON: %v", err)
	}
	if len(exported) != 1 {
		t.Errorf("Expected 1 exported entry, got %d", len(exported))
	}
}

func TestExportCSV(t *testing.T) {
	entries := []FingerprintEntry{
		{Hash: "111", Service: "TestCSV", Category: "Test", Tags: []string{"a", "b"}},
	}
	db := NewDB(entries)

	csv := db.ExportCSV()
	if csv == "" {
		t.Error("ExportCSV() returned empty string")
	}
	if !contains(csv, "TestCSV") {
		t.Error("ExportCSV() output missing service name")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
