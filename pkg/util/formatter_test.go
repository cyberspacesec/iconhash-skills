package util

import "testing"

func TestFormatHash(t *testing.T) {
	tests := []struct {
		name     string
		hash     string
		format   OutputFormat
		expected string
	}{
		{"Plain format", "12345", FormatPlain, "12345"},
		{"Fofa format", "12345", FormatFofa, "icon_hash=\"12345\""},
		{"Shodan format", "12345", FormatShodan, "http.favicon.hash:12345"},
		{"Negative number with Plain format", "-12345", FormatPlain, "-12345"},
		{"Negative number with Fofa format", "-12345", FormatFofa, "icon_hash=\"-12345\""},
		{"Negative number with Shodan format", "-12345", FormatShodan, "http.favicon.hash:-12345"},
		{"Censys format", "-12345", FormatCensys, "services.http.response.favicons.md5_hash:-12345"},
		{"Quake format", "-12345", FormatQuake, "favicon.hash:\"-12345\""},
		{"ZoomEye format", "-12345", FormatZoomEye, "iconhash:\"-12345\""},
		{"Hunter format", "-12345", FormatHunter, "web.icon=\"-12345\""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := FormatHash(test.hash, test.format)
			if result != test.expected {
				t.Errorf("FormatHash(%q, %v) = %q, expected %q", test.hash, test.format, result, test.expected)
			}
		})
	}
}
