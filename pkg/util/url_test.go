package util

import "testing"

func TestIsURL(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"https://example.com/favicon.ico", true},
		{"http://example.com/favicon.ico", true},
		{"https://example.com", true},
		{"www.example.com", true},
		{"example.com/favicon.ico", false},
		{"foo://bar", false},
		{"", false},
		{"not a url", false},
		{"  https://example.com  ", true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := IsURL(test.input)
			if result != test.expected {
				t.Errorf("IsURL(%q) = %v, expected %v", test.input, result, test.expected)
			}
		})
	}
}
