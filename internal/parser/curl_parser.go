package parser

import (
	"fmt"
	"regexp"
	"strings"

	"curlex/internal/models"
)

// CurlParser parses curl command strings
type CurlParser struct{}

// NewCurlParser creates a new curl parser instance
func NewCurlParser() *CurlParser {
	return &CurlParser{}
}

// ParseCurl converts a curl command string to a PreparedRequest
// Supports common flags: -X, -H, -d, -u, -A, --json
func (p *CurlParser) ParseCurl(curlCmd string) (*models.PreparedRequest, error) {
	req := &models.PreparedRequest{
		Method:  "GET", // default
		Headers: make(map[string]string),
	}

	// Clean up the command
	curlCmd = strings.TrimSpace(curlCmd)

	// Remove 'curl' prefix if present
	curlCmd = strings.TrimPrefix(curlCmd, "curl ")
	curlCmd = strings.TrimPrefix(curlCmd, "curl")
	curlCmd = strings.TrimSpace(curlCmd)

	// Extract URL (first argument that doesn't start with -)
	url, remaining := p.extractURL(curlCmd)
	if url == "" {
		return nil, fmt.Errorf("no URL found in curl command")
	}
	req.URL = url

	// Parse flags
	if err := p.parseFlags(remaining, req); err != nil {
		return nil, err
	}

	return req, nil
}

// extractURL extracts the URL from the curl command
func (p *CurlParser) extractURL(cmd string) (string, string) {
	// Pattern: URL is either quoted or unquoted, not starting with -
	patterns := []string{
		`"([^"]+)"`,                     // Double quoted
		`'([^']+)'`,                     // Single quoted
		`(https?://[^\s]+)`,             // Unquoted http(s) URL
		`([a-zA-Z0-9\.\-_:/@]+[^\s]*)`, // Unquoted URL
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(cmd)
		if len(matches) > 1 {
			url := matches[1]
			// Make sure it looks like a URL
			if strings.Contains(url, "://") || strings.HasPrefix(url, "http") {
				remaining := re.ReplaceAllString(cmd, "")
				return url, remaining
			}
		}
	}

	// Fallback: extract first non-flag argument
	parts := strings.Fields(cmd)
	for i, part := range parts {
		if !strings.HasPrefix(part, "-") {
			remaining := strings.Join(append(parts[:i], parts[i+1:]...), " ")
			return part, remaining
		}
	}

	return "", cmd
}

// parseFlags parses curl flags from the command
func (p *CurlParser) parseFlags(cmd string, req *models.PreparedRequest) error {
	// Parse -X/--request METHOD
	if method := p.extractFlag(cmd, `-X\s+(\w+)`); method != "" {
		req.Method = strings.ToUpper(method)
	} else if method := p.extractFlag(cmd, `--request\s+(\w+)`); method != "" {
		req.Method = strings.ToUpper(method)
	}

	// Parse -H/--header "Header: Value"
	headers := p.extractMultipleFlags(cmd, `-H\s+["']([^"']+)["']`)
	headers = append(headers, p.extractMultipleFlags(cmd, `--header\s+["']([^"']+)["']`)...)
	for _, header := range headers {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			req.Headers[key] = value
		}
	}

	// Parse -d/--data "body"
	// Try double quotes first
	if body := p.extractFlag(cmd, `-d\s+"([^"]+)"`); body != "" {
		req.Body = body
		// -d implies POST if method not specified
		if req.Method == "GET" {
			req.Method = "POST"
		}
	} else if body := p.extractFlag(cmd, `-d\s+'([^']+)'`); body != "" {
		// Try single quotes
		req.Body = body
		if req.Method == "GET" {
			req.Method = "POST"
		}
	} else if body := p.extractFlag(cmd, `--data\s+"([^"]+)"`); body != "" {
		req.Body = body
		if req.Method == "GET" {
			req.Method = "POST"
		}
	} else if body := p.extractFlag(cmd, `--data\s+'([^']+)'`); body != "" {
		req.Body = body
		if req.Method == "GET" {
			req.Method = "POST"
		}
	}

	// Parse -u/--user "username:password" for Basic Auth
	if auth := p.extractFlag(cmd, `-u\s+["']([^"']+)["']`); auth != "" {
		req.Headers["Authorization"] = "Basic " + auth // Note: should be base64 encoded, but keeping simple for now
	} else if auth := p.extractFlag(cmd, `--user\s+["']([^"']+)["']`); auth != "" {
		req.Headers["Authorization"] = "Basic " + auth
	}

	// Parse -A/--user-agent "agent"
	if agent := p.extractFlag(cmd, `-A\s+["']([^"']+)["']`); agent != "" {
		req.Headers["User-Agent"] = agent
	} else if agent := p.extractFlag(cmd, `--user-agent\s+["']([^"']+)["']`); agent != "" {
		req.Headers["User-Agent"] = agent
	}

	// Parse --json (implies -H "Content-Type: application/json")
	if strings.Contains(cmd, "--json") {
		req.Headers["Content-Type"] = "application/json"
		req.Headers["Accept"] = "application/json"
	}

	return nil
}

// extractFlag extracts a single flag value using regex
func (p *CurlParser) extractFlag(cmd, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(cmd)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractMultipleFlags extracts all occurrences of a flag
func (p *CurlParser) extractMultipleFlags(cmd, pattern string) []string {
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(cmd, -1)
	var results []string
	for _, match := range matches {
		if len(match) > 1 {
			results = append(results, match[1])
		}
	}
	return results
}
