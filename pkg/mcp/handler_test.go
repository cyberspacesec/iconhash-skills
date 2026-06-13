package mcp

import (
	"testing"
)

func TestNewRequest(t *testing.T) {
	req := NewRequest()

	if req.Version != ProtocolVersion {
		t.Errorf("Expected version %s, got %s", ProtocolVersion, req.Version)
	}

	if req.Protocol != ProtocolName {
		t.Errorf("Expected protocol %s, got %s", ProtocolName, req.Protocol)
	}

	if len(req.Context.Messages) != 0 {
		t.Errorf("Expected empty messages, got %d", len(req.Context.Messages))
	}
}

func TestNewResponse(t *testing.T) {
	resp := NewResponse()

	if resp.Version != ProtocolVersion {
		t.Errorf("Expected version %s, got %s", ProtocolVersion, resp.Version)
	}

	if resp.Protocol != ProtocolName {
		t.Errorf("Expected protocol %s, got %s", ProtocolName, resp.Protocol)
	}

	if resp.Message.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got %s", resp.Message.Role)
	}

	if resp.Usage == nil {
		t.Error("Expected usage to be non-nil")
	}
}

func TestAddMessage(t *testing.T) {
	req := NewRequest()
	req.AddMessage("user", "Hello")

	if len(req.Context.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(req.Context.Messages))
	}

	if req.Context.Messages[0].Role != "user" {
		t.Errorf("Expected role 'user', got %s", req.Context.Messages[0].Role)
	}

	if req.Context.Messages[0].Content != "Hello" {
		t.Errorf("Expected content 'Hello', got %s", req.Context.Messages[0].Content)
	}
}

func TestRequestValidate(t *testing.T) {
	tests := []struct {
		name          string
		modifyRequest func(*Request)
		expectError   bool
	}{
		{
			name:          "Valid request",
			modifyRequest: func(r *Request) { r.AddMessage("user", "Hello") },
			expectError:   false,
		},
		{
			name:          "Missing version",
			modifyRequest: func(r *Request) { r.Version = ""; r.AddMessage("user", "Hello") },
			expectError:   true,
		},
		{
			name:          "Missing protocol",
			modifyRequest: func(r *Request) { r.Protocol = ""; r.AddMessage("user", "Hello") },
			expectError:   true,
		},
		{
			name:          "No messages",
			modifyRequest: func(r *Request) {},
			expectError:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := NewRequest()
			test.modifyRequest(req)

			err := req.Validate()
			if test.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !test.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestResponseComplete(t *testing.T) {
	resp := NewResponse()
	resp.Message.Content = "Hello world"
	resp.Complete()

	if resp.Usage.ProcessingTime <= 0 {
		t.Errorf("Expected processing time > 0, got %f", resp.Usage.ProcessingTime)
	}

	if resp.Usage.CompletedAt.IsZero() {
		t.Error("Expected completed_at to be set")
	}

	if resp.Usage.TotalTokens <= 0 {
		t.Errorf("Expected total tokens > 0, got %d", resp.Usage.TotalTokens)
	}
}

func TestHandlerTools(t *testing.T) {
	handler := NewHandler(false)
	tools := handler.Tools()

	if len(tools) == 0 {
		t.Fatal("Tools() returned empty list")
	}

	expectedTools := map[string]bool{
		"iconhash_url":      false,
		"iconhash_base64":   false,
		"iconhash_file":     false,
		"iconhash_discover": false,
		"iconhash_lookup":   false,
	}
	for _, tool := range tools {
		if _, ok := expectedTools[tool.Name]; ok {
			expectedTools[tool.Name] = true
		}
	}
	for name, found := range expectedTools {
		if !found {
			t.Errorf("Expected tool %q not found in Tools()", name)
		}
	}
}

func TestCallTool(t *testing.T) {
	handler := NewHandler(false)

	t.Run("unknown tool", func(t *testing.T) {
		result := handler.CallTool("nonexistent", nil)
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		if !result.IsError {
			t.Error("Expected IsError=true for unknown tool")
		}
	})

	t.Run("iconhash_url with missing url", func(t *testing.T) {
		result := handler.CallTool("iconhash_url", map[string]interface{}{})
		if !result.IsError {
			t.Error("Expected error for missing url parameter")
		}
	})

	t.Run("iconhash_file with missing path", func(t *testing.T) {
		result := handler.CallTool("iconhash_file", map[string]interface{}{})
		if !result.IsError {
			t.Error("Expected error for missing path parameter")
		}
	})

	t.Run("iconhash_discover with missing url", func(t *testing.T) {
		result := handler.CallTool("iconhash_discover", map[string]interface{}{})
		if !result.IsError {
			t.Error("Expected error for missing url parameter")
		}
	})

	t.Run("iconhash_lookup with missing hash", func(t *testing.T) {
		result := handler.CallTool("iconhash_lookup", map[string]interface{}{})
		if !result.IsError {
			t.Error("Expected error for missing hash parameter")
		}
	})

	t.Run("iconhash_lookup with known hash", func(t *testing.T) {
		result := handler.CallTool("iconhash_lookup", map[string]interface{}{
			"hash": "-305179312",
		})
		if result.IsError {
			t.Error("Expected success for known hash lookup")
		}
		if len(result.Content) == 0 {
			t.Error("Expected content in lookup result")
		}
	})

	t.Run("iconhash_lookup with unknown hash", func(t *testing.T) {
		result := handler.CallTool("iconhash_lookup", map[string]interface{}{
			"hash": "999999999",
		})
		if result.IsError {
			t.Error("Expected success (not error) for unknown hash lookup, just no matches")
		}
	})

	t.Run("iconhash_base64 with missing data", func(t *testing.T) {
		result := handler.CallTool("iconhash_base64", map[string]interface{}{})
		if !result.IsError {
			t.Error("Expected error for missing data parameter")
		}
	})
}

func TestFormatArg(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]interface{}
		expected string
	}{
		{"default fofa", map[string]interface{}{}, "fofa"},
		{"plain format", map[string]interface{}{"format": "plain"}, "plain"},
		{"shodan format", map[string]interface{}{"format": "shodan"}, "shodan"},
		{"censys format", map[string]interface{}{"format": "censys"}, "censys"},
		{"quake format", map[string]interface{}{"format": "quake"}, "quake"},
		{"zoomeye format", map[string]interface{}{"format": "zoomeye"}, "zoomeye"},
		{"hunter format", map[string]interface{}{"format": "hunter"}, "hunter"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format := formatArg(tt.args, "format")
			// Convert format to string for comparison
			var got string
			switch format {
			case 0:
				got = "plain"
			case 1:
				got = "fofa"
			case 2:
				got = "shodan"
			case 3:
				got = "censys"
			case 4:
				got = "quake"
			case 5:
				got = "zoomeye"
			case 6:
				got = "hunter"
			}
			if got != tt.expected {
				t.Errorf("formatArg() = %s, expected %s", got, tt.expected)
			}
		})
	}
}
