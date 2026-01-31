package parser

import (
	"testing"
)

func TestCurlParser_ParseCurl(t *testing.T) {
	parser := NewCurlParser()

	tests := []struct {
		name           string
		curl           string
		expectedMethod string
		expectedURL    string
		expectedBody   string
		headerKey      string
		headerValue    string
	}{
		{
			name:           "simple GET",
			curl:           "curl https://example.com",
			expectedMethod: "GET",
			expectedURL:    "https://example.com",
		},
		{
			name:           "GET with -X",
			curl:           "curl -X GET https://example.com",
			expectedMethod: "GET",
			expectedURL:    "https://example.com",
		},
		{
			name:           "POST with data",
			curl:           `curl -X POST -d "test=data" https://example.com`,
			expectedMethod: "POST",
			expectedURL:    "https://example.com",
			expectedBody:   "test=data",
		},
		{
			name:           "with header",
			curl:           `curl -H "Content-Type: application/json" https://example.com`,
			expectedMethod: "GET",
			expectedURL:    "https://example.com",
			headerKey:      "Content-Type",
			headerValue:    "application/json",
		},
		{
			name:           "complex curl",
			curl:           `curl -X POST -H "Content-Type: application/json" -d '{"key":"value"}' https://api.example.com/endpoint`,
			expectedMethod: "POST",
			expectedURL:    "https://api.example.com/endpoint",
			expectedBody:   `{"key":"value"}`,
			headerKey:      "Content-Type",
			headerValue:    "application/json",
		},
		{
			name:           "single cookie with -b",
			curl:           `curl -b "session=abc123" https://example.com`,
			expectedMethod: "GET",
			expectedURL:    "https://example.com",
			headerKey:      "Cookie",
			headerValue:    "session=abc123",
		},
		{
			name:           "single cookie with --cookie",
			curl:           `curl --cookie "user_id=42" https://example.com`,
			expectedMethod: "GET",
			expectedURL:    "https://example.com",
			headerKey:      "Cookie",
			headerValue:    "user_id=42",
		},
		{
			name:           "multiple cookies",
			curl:           `curl -b "session=abc123" -b "user_id=42" https://example.com`,
			expectedMethod: "GET",
			expectedURL:    "https://example.com",
			headerKey:      "Cookie",
			headerValue:    "session=abc123; user_id=42",
		},
		{
			name:           "multiple cookies mixed flags",
			curl:           `curl -b "session=abc123" --cookie "user_id=42" -b "theme=dark" https://example.com`,
			expectedMethod: "GET",
			expectedURL:    "https://example.com",
			headerKey:      "Cookie",
			headerValue:    "session=abc123; theme=dark; user_id=42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := parser.ParseCurl(tt.curl)
			if err != nil {
				t.Fatalf("ParseCurl failed: %v", err)
			}

			if req.Method != tt.expectedMethod {
				t.Errorf("Method = %q, want %q", req.Method, tt.expectedMethod)
			}

			if req.URL != tt.expectedURL {
				t.Errorf("URL = %q, want %q", req.URL, tt.expectedURL)
			}

			if tt.expectedBody != "" && req.Body != tt.expectedBody {
				t.Errorf("Body = %q, want %q", req.Body, tt.expectedBody)
			}

			if tt.headerKey != "" {
				if val, ok := req.Headers[tt.headerKey]; !ok {
					t.Errorf("Header %q not found", tt.headerKey)
				} else if val != tt.headerValue {
					t.Errorf("Header %q = %q, want %q", tt.headerKey, val, tt.headerValue)
				}
			}
		})
	}
}
