package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cyberspacesec/iconhash-skills/pkg/util"
)

func TestHealthEndpoint(t *testing.T) {
	// Create a server with default config
	server := NewServer(nil)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Call the handler directly
	server.handleHealth(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Parse response
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check fields
	if status, ok := resp["status"]; !ok || status != "ok" {
		t.Errorf("Expected status to be 'ok', got '%s'", status)
	}

	if _, ok := resp["time"]; !ok {
		t.Error("Expected time field in response")
	}
}

func TestAuthMiddleware(t *testing.T) {
	// Create a server with auth token
	config := DefaultConfig()
	config.AuthToken = "test-token"
	server := NewServer(config)

	// Create a handler that returns success
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Wrap the handler with auth middleware
	authHandler := server.authMiddleware(handler)

	tests := []struct {
		name           string
		path           string
		authorization  string
		queryToken     string
		expectedStatus int
	}{
		{
			name:           "Valid header token",
			path:           "/test",
			authorization:  "Bearer test-token",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid query token",
			path:           "/test?token=test-token",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid token",
			path:           "/test",
			authorization:  "Bearer wrong-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Missing token",
			path:           "/test",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Health endpoint (no auth required)",
			path:           "/health",
			expectedStatus: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest(http.MethodGet, test.path, nil)
			if test.authorization != "" {
				req.Header.Set("Authorization", test.authorization)
			}
			w := httptest.NewRecorder()

			// Call the middleware
			authHandler.ServeHTTP(w, req)

			// Check response
			if w.Code != test.expectedStatus {
				t.Errorf("Expected status code %d, got %d", test.expectedStatus, w.Code)
			}
		})
	}
}

func TestParseFormatParam(t *testing.T) {
	tests := []struct {
		name           string
		formatStr      string
		expectedFormat util.OutputFormat
	}{
		{"Plain format", "plain", util.FormatPlain},
		{"Fofa format", "fofa", util.FormatFofa},
		{"Shodan format", "shodan", util.FormatShodan},
		{"Empty string (defaults to Fofa)", "", util.FormatFofa},
		{"Invalid format (defaults to Fofa)", "invalid", util.FormatFofa},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := parseFormatParam(test.formatStr)
			if result != test.expectedFormat {
				t.Errorf("parseFormatParam(%q) = %d, expected %d",
					test.formatStr, result, test.expectedFormat)
			}
		})
	}
}

func TestGetFormatName(t *testing.T) {
	tests := []struct {
		name         string
		format       util.OutputFormat
		expectedName string
	}{
		{"Plain format", util.FormatPlain, "plain"},
		{"Fofa format", util.FormatFofa, "fofa"},
		{"Shodan format", util.FormatShodan, "shodan"},
		{"Invalid format", util.OutputFormat(99), "unknown"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := getFormatName(test.format)
			if result != test.expectedName {
				t.Errorf("getFormatName(%d) = %q, expected %q",
					test.format, result, test.expectedName)
			}
		})
	}
}

// Helper function to create a multipart form request with a file
func createMultipartRequest(t *testing.T, fieldName, fileName string, fileContent []byte) (*http.Request, *multipart.Writer) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	part.Write(fileContent)

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/hash/file", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, writer
}

// Note: Additional tests for URL, file, and base64 handlers would be similar
// but would require mocking the hasher's functions, which is beyond the scope
// of this example. In a real implementation, you'd use a mock or a test double
// for the hasher.
