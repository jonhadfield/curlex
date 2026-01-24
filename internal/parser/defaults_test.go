package parser

import (
	"testing"
	"time"

	"curlex/internal/models"
)

func TestMergeDefaults_Timeout(t *testing.T) {
	defaults := models.DefaultConfig{
		Timeout: 60 * time.Second,
	}

	test := &models.Test{
		Name: "Test without timeout",
	}

	MergeDefaults(test, defaults)

	if test.Timeout != 60*time.Second {
		t.Errorf("Expected timeout 60s, got %v", test.Timeout)
	}
}

func TestMergeDefaults_TimeoutOverride(t *testing.T) {
	defaults := models.DefaultConfig{
		Timeout: 60 * time.Second,
	}

	test := &models.Test{
		Name:    "Test with timeout",
		Timeout: 30 * time.Second,
	}

	MergeDefaults(test, defaults)

	if test.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s (override), got %v", test.Timeout)
	}
}

func TestMergeDefaults_Retries(t *testing.T) {
	defaults := models.DefaultConfig{
		Retries: 3,
	}

	test := &models.Test{
		Name: "Test without retries",
	}

	MergeDefaults(test, defaults)

	if test.Retries != 3 {
		t.Errorf("Expected retries 3, got %d", test.Retries)
	}
}

func TestMergeDefaults_RetryDelay(t *testing.T) {
	defaults := models.DefaultConfig{
		RetryDelay: 2 * time.Second,
	}

	test := &models.Test{
		Name: "Test without retry delay",
	}

	MergeDefaults(test, defaults)

	if test.RetryDelay != 2*time.Second {
		t.Errorf("Expected retry delay 2s, got %v", test.RetryDelay)
	}
}

func TestMergeDefaults_RetryBackoff(t *testing.T) {
	defaults := models.DefaultConfig{
		RetryBackoff: "exponential",
	}

	test := &models.Test{
		Name: "Test without backoff",
	}

	MergeDefaults(test, defaults)

	if test.RetryBackoff != "exponential" {
		t.Errorf("Expected backoff 'exponential', got %s", test.RetryBackoff)
	}
}

func TestMergeDefaults_RetryOnStatus(t *testing.T) {
	defaults := models.DefaultConfig{
		RetryOnStatus: []int{500, 502, 503},
	}

	test := &models.Test{
		Name: "Test without retry status",
	}

	MergeDefaults(test, defaults)

	if len(test.RetryOnStatus) != 3 {
		t.Errorf("Expected 3 retry statuses, got %d", len(test.RetryOnStatus))
	}

	for i, status := range test.RetryOnStatus {
		if status != defaults.RetryOnStatus[i] {
			t.Errorf("Expected status %d at index %d, got %d", defaults.RetryOnStatus[i], i, status)
		}
	}
}

func TestMergeDefaults_MaxRedirects(t *testing.T) {
	maxRedir := 5
	defaults := models.DefaultConfig{
		MaxRedirects: &maxRedir,
	}

	test := &models.Test{
		Name: "Test without max redirects",
	}

	MergeDefaults(test, defaults)

	if test.MaxRedirects == nil {
		t.Fatal("Expected max_redirects to be set")
	}

	if *test.MaxRedirects != 5 {
		t.Errorf("Expected max_redirects 5, got %d", *test.MaxRedirects)
	}
}

func TestMergeDefaults_Headers(t *testing.T) {
	defaults := models.DefaultConfig{
		Headers: map[string]string{
			"User-Agent": "curlex/1.0",
			"Accept":     "application/json",
		},
	}

	test := &models.Test{
		Name: "Test with structured request",
		Request: &models.StructuredRequest{
			Method: "GET",
			URL:    "https://example.com",
		},
	}

	MergeDefaults(test, defaults)

	if test.Request.Headers == nil {
		t.Fatal("Expected headers to be set")
	}

	if test.Request.Headers["User-Agent"] != "curlex/1.0" {
		t.Errorf("Expected User-Agent 'curlex/1.0', got %s", test.Request.Headers["User-Agent"])
	}

	if test.Request.Headers["Accept"] != "application/json" {
		t.Errorf("Expected Accept 'application/json', got %s", test.Request.Headers["Accept"])
	}
}

func TestMergeDefaults_HeadersOverride(t *testing.T) {
	defaults := models.DefaultConfig{
		Headers: map[string]string{
			"User-Agent": "curlex/1.0",
			"Accept":     "application/json",
		},
	}

	test := &models.Test{
		Name: "Test with headers",
		Request: &models.StructuredRequest{
			Method: "GET",
			URL:    "https://example.com",
			Headers: map[string]string{
				"User-Agent": "custom/2.0", // Override
			},
		},
	}

	MergeDefaults(test, defaults)

	// User-Agent should be overridden
	if test.Request.Headers["User-Agent"] != "custom/2.0" {
		t.Errorf("Expected User-Agent 'custom/2.0' (override), got %s", test.Request.Headers["User-Agent"])
	}

	// Accept should be inherited
	if test.Request.Headers["Accept"] != "application/json" {
		t.Errorf("Expected Accept 'application/json' (inherited), got %s", test.Request.Headers["Accept"])
	}
}

func TestApplyDefaults(t *testing.T) {
	suite := &models.TestSuite{
		Defaults: models.DefaultConfig{
			Timeout: 30 * time.Second,
			Retries: 2,
		},
		Tests: []models.Test{
			{Name: "Test 1"},
			{Name: "Test 2"},
			{Name: "Test 3", Timeout: 60 * time.Second}, // Override
		},
	}

	ApplyDefaults(suite)

	if suite.Tests[0].Timeout != 30*time.Second {
		t.Errorf("Test 1: Expected timeout 30s, got %v", suite.Tests[0].Timeout)
	}

	if suite.Tests[1].Retries != 2 {
		t.Errorf("Test 2: Expected retries 2, got %d", suite.Tests[1].Retries)
	}

	if suite.Tests[2].Timeout != 60*time.Second {
		t.Errorf("Test 3: Expected timeout 60s (override), got %v", suite.Tests[2].Timeout)
	}
}
