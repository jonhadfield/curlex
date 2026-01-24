package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"curlex/internal/models"
	"curlex/internal/parser"
)

// Executor executes HTTP requests and returns results
type Executor struct {
	client     *http.Client
	curlParser *parser.CurlParser
}

// NewExecutor creates a new HTTP executor with default settings
func NewExecutor(timeout time.Duration) *Executor {
	return &Executor{
		client: &http.Client{
			Timeout: timeout,
			// Default: follow up to 10 redirects
			CheckRedirect: nil,
		},
		curlParser: parser.NewCurlParser(),
	}
}

// Execute runs a single test and returns the result
func (e *Executor) Execute(ctx context.Context, test models.Test) (*models.TestResult, error) {
	result := &models.TestResult{
		Test: test,
	}

	// Prepare the request
	preparedReq, err := e.prepareRequest(test)
	if err != nil {
		result.Error = err
		result.Success = false
		return result, nil
	}

	// Store prepared request for logging
	result.PreparedRequest = preparedReq

	// Create HTTP request
	httpReq, err := e.createHTTPRequest(ctx, preparedReq)
	if err != nil {
		result.Error = err
		result.Success = false
		return result, nil
	}

	// Configure redirect policy if specified
	client := e.client
	if test.MaxRedirects != nil {
		client = e.createClientWithRedirects(*test.MaxRedirects)
	}

	// Execute the request
	start := time.Now()
	resp, err := client.Do(httpReq)
	result.ResponseTime = time.Since(start)

	if err != nil {
		result.Error = fmt.Errorf("request failed: %w", err)
		result.Success = false
		return result, nil
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Errorf("failed to read response body: %w", err)
		result.Success = false
		return result, nil
	}

	// Populate result
	result.StatusCode = resp.StatusCode
	result.ResponseBody = string(body)
	result.Headers = resp.Header

	return result, nil
}

// prepareRequest converts a Test to a PreparedRequest
func (e *Executor) prepareRequest(test models.Test) (*models.PreparedRequest, error) {
	// If curl command is specified, parse it
	if test.Curl != "" {
		return e.curlParser.ParseCurl(test.Curl)
	}

	// Otherwise use structured request
	if test.Request == nil {
		return nil, fmt.Errorf("no request specification found")
	}

	preparedReq := &models.PreparedRequest{
		Method:  test.Request.Method,
		URL:     test.Request.URL,
		Body:    test.Request.Body,
		Headers: make(map[string]string),
	}

	// Copy headers
	if test.Request.Headers != nil {
		for k, v := range test.Request.Headers {
			preparedReq.Headers[k] = v
		}
	}

	return preparedReq, nil
}

// createHTTPRequest creates an http.Request from a PreparedRequest
func (e *Executor) createHTTPRequest(ctx context.Context, preparedReq *models.PreparedRequest) (*http.Request, error) {
	// Create request body reader
	var bodyReader io.Reader
	if preparedReq.Body != "" {
		bodyReader = bytes.NewBufferString(preparedReq.Body)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, preparedReq.Method, preparedReq.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	for key, value := range preparedReq.Headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

// createClientWithRedirects creates an HTTP client with custom redirect policy
func (e *Executor) createClientWithRedirects(maxRedirects int) *http.Client {
	client := &http.Client{
		Timeout: e.client.Timeout,
	}

	if maxRedirects == 0 {
		// No redirects allowed
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else if maxRedirects > 0 {
		// Limit number of redirects
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return fmt.Errorf("stopped after %d redirects", maxRedirects)
			}
			return nil
		}
	}
	// maxRedirects == -1: unlimited redirects (use default behavior, CheckRedirect = nil)

	return client
}
