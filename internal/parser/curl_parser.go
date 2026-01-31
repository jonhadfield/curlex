package parser

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"curlex/internal/models"
)

// Pre-compiled regex patterns for performance
var (
	// URL extraction patterns
	urlPatternDoubleQuote = regexp.MustCompile(`"([^"]+)"`)
	urlPatternSingleQuote = regexp.MustCompile(`'([^']+)'`)
	urlPatternHTTP        = regexp.MustCompile(`(https?://[^\s]+)`)
	urlPatternGeneric     = regexp.MustCompile(`([a-zA-Z0-9\.\-_:/@]+[^\s]*)`)

	// Flag parsing patterns
	flagMethodShort     = regexp.MustCompile(`-X\s+(\w+)`)
	flagMethodLong      = regexp.MustCompile(`--request\s+(\w+)`)
	flagHeaderShort     = regexp.MustCompile(`-H\s+["']([^"']+)["']`)
	flagHeaderLong      = regexp.MustCompile(`--header\s+["']([^"']+)["']`)
	flagDataShortDouble = regexp.MustCompile(`-d\s+"([^"]+)"`)
	flagDataShortSingle = regexp.MustCompile(`-d\s+'([^']+)'`)
	flagDataLongDouble  = regexp.MustCompile(`--data\s+"([^"]+)"`)
	flagDataLongSingle  = regexp.MustCompile(`--data\s+'([^']+)'`)
	flagUserShort       = regexp.MustCompile(`-u\s+["']([^"']+)["']`)
	flagUserLong        = regexp.MustCompile(`--user\s+["']([^"']+)["']`)
	flagUserAgentShort  = regexp.MustCompile(`-A\s+["']([^"']+)["']`)
	flagUserAgentLong   = regexp.MustCompile(`--user-agent\s+["']([^"']+)["']`)
	flagCookieShort     = regexp.MustCompile(`-b\s+["']([^"']+)["']`)
	flagCookieLong      = regexp.MustCompile(`--cookie\s+["']([^"']+)["']`)
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
	// Use pre-compiled regex patterns for performance
	patterns := []*regexp.Regexp{
		urlPatternDoubleQuote,
		urlPatternSingleQuote,
		urlPatternHTTP,
		urlPatternGeneric,
	}

	for _, re := range patterns {
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
	if method := p.extractFlagRe(cmd, flagMethodShort); method != "" {
		req.Method = strings.ToUpper(method)
	} else if method := p.extractFlagRe(cmd, flagMethodLong); method != "" {
		req.Method = strings.ToUpper(method)
	}

	// Parse -H/--header "Header: Value"
	headers := p.extractMultipleFlagsRe(cmd, flagHeaderShort)
	headers = append(headers, p.extractMultipleFlagsRe(cmd, flagHeaderLong)...)
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
	if body := p.extractFlagRe(cmd, flagDataShortDouble); body != "" {
		req.Body = body
		// -d implies POST if method not specified
		if req.Method == "GET" {
			req.Method = "POST"
		}
	} else if body := p.extractFlagRe(cmd, flagDataShortSingle); body != "" {
		// Try single quotes
		req.Body = body
		if req.Method == "GET" {
			req.Method = "POST"
		}
	} else if body := p.extractFlagRe(cmd, flagDataLongDouble); body != "" {
		req.Body = body
		if req.Method == "GET" {
			req.Method = "POST"
		}
	} else if body := p.extractFlagRe(cmd, flagDataLongSingle); body != "" {
		req.Body = body
		if req.Method == "GET" {
			req.Method = "POST"
		}
	}

	// Parse -u/--user "username:password" for Basic Auth
	if auth := p.extractFlagRe(cmd, flagUserShort); auth != "" {
		encoded := base64.StdEncoding.EncodeToString([]byte(auth))
		req.Headers["Authorization"] = "Basic " + encoded
	} else if auth := p.extractFlagRe(cmd, flagUserLong); auth != "" {
		encoded := base64.StdEncoding.EncodeToString([]byte(auth))
		req.Headers["Authorization"] = "Basic " + encoded
	}

	// Parse -A/--user-agent "agent"
	if agent := p.extractFlagRe(cmd, flagUserAgentShort); agent != "" {
		req.Headers["User-Agent"] = agent
	} else if agent := p.extractFlagRe(cmd, flagUserAgentLong); agent != "" {
		req.Headers["User-Agent"] = agent
	}

	// Parse -b/--cookie "name=value"
	// Multiple cookies are combined with semicolons
	cookies := p.extractMultipleFlagsRe(cmd, flagCookieShort)
	cookies = append(cookies, p.extractMultipleFlagsRe(cmd, flagCookieLong)...)
	if len(cookies) > 0 {
		// Combine multiple cookies with semicolons
		req.Headers["Cookie"] = strings.Join(cookies, "; ")
	}

	// Parse --json (implies -H "Content-Type: application/json")
	if strings.Contains(cmd, "--json") {
		req.Headers["Content-Type"] = "application/json"
		req.Headers["Accept"] = "application/json"
	}

	return nil
}

// extractFlagRe extracts a single flag value using pre-compiled regex
func (p *CurlParser) extractFlagRe(cmd string, re *regexp.Regexp) string {
	matches := re.FindStringSubmatch(cmd)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractMultipleFlagsRe extracts all occurrences of a flag using pre-compiled regex
func (p *CurlParser) extractMultipleFlagsRe(cmd string, re *regexp.Regexp) []string {
	matches := re.FindAllStringSubmatch(cmd, -1)
	var results []string
	for _, match := range matches {
		if len(match) > 1 {
			results = append(results, match[1])
		}
	}
	return results
}
