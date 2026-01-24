package config

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Note: Testing ParseFlags is challenging due to global flag state
// These tests focus on validation and error cases

func TestConfig_Defaults(t *testing.T) {
	// Test that Config struct can be created with expected types
	cfg := &Config{
		Timeout:      30 * time.Second,
		NoColor:      false,
		Version:      false,
		Verbose:      false,
		LogDir:       "",
		TestFilter:   "",
		TestPattern:  "",
		SkipTests:    "",
		Parallel:     false,
		Concurrency:  10,
		FailFast:     false,
		OutputFormat: "human",
		Quiet:        false,
	}

	if cfg.Timeout != 30*time.Second {
		t.Errorf("Default timeout = %v, want 30s", cfg.Timeout)
	}
	if cfg.Concurrency != 10 {
		t.Errorf("Default concurrency = %d, want 10", cfg.Concurrency)
	}
	if cfg.OutputFormat != "human" {
		t.Errorf("Default output format = %s, want human", cfg.OutputFormat)
	}
}

func TestParseFlags_MissingFile(t *testing.T) {
	// Reset flag state
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Set args to empty (no test file)
	os.Args = []string{"curlex"}

	cfg, err := ParseFlags()
	if err == nil {
		t.Error("ParseFlags() should error with missing test file")
	}
	if cfg != nil {
		t.Error("ParseFlags() should return nil config on error")
	}
	if err != nil && err.Error() != "missing required argument: test-file.yaml" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestParseFlags_NonExistentFile(t *testing.T) {
	// Reset flag state
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Set args to non-existent file
	os.Args = []string{"curlex", "/nonexistent/path/test.yaml"}

	cfg, err := ParseFlags()
	if err == nil {
		t.Error("ParseFlags() should error with non-existent file")
	}
	if cfg != nil {
		t.Error("ParseFlags() should return nil config on error")
	}
}

func TestParseFlags_ExistingFile(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(testFile, []byte("version: 1.0\ntests: []"), 0644); err != nil {
		t.Fatal(err)
	}

	// Reset flag state
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Set args with valid file
	os.Args = []string{"curlex", testFile}

	cfg, err := ParseFlags()
	if err != nil {
		t.Fatalf("ParseFlags() unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("ParseFlags() returned nil config")
	}
	if cfg.TestFile != testFile {
		t.Errorf("TestFile = %s, want %s", cfg.TestFile, testFile)
	}
}

func TestParseFlags_WithFlags(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(testFile, []byte("version: 1.0\ntests: []"), 0644); err != nil {
		t.Fatal(err)
	}

	// Reset flag state
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Set args with flags
	os.Args = []string{
		"curlex",
		"--timeout", "60s",
		"--no-color",
		"--verbose",
		"--parallel",
		"--concurrency", "5",
		"--fail-fast",
		"--output", "json",
		testFile,
	}

	cfg, err := ParseFlags()
	if err != nil {
		t.Fatalf("ParseFlags() unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("ParseFlags() returned nil config")
	}

	if cfg.Timeout != 60*time.Second {
		t.Errorf("Timeout = %v, want 60s", cfg.Timeout)
	}
	if !cfg.NoColor {
		t.Error("NoColor should be true")
	}
	if !cfg.Verbose {
		t.Error("Verbose should be true")
	}
	if !cfg.Parallel {
		t.Error("Parallel should be true")
	}
	if cfg.Concurrency != 5 {
		t.Errorf("Concurrency = %d, want 5", cfg.Concurrency)
	}
	if !cfg.FailFast {
		t.Error("FailFast should be true")
	}
	if cfg.OutputFormat != "json" {
		t.Errorf("OutputFormat = %s, want json", cfg.OutputFormat)
	}
}

func TestParseFlags_Version(t *testing.T) {
	// Reset flag state
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Set args with version flag
	os.Args = []string{"curlex", "--version"}

	cfg, err := ParseFlags()
	if err != nil {
		t.Fatalf("ParseFlags() unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("ParseFlags() returned nil config")
	}
	if !cfg.Version {
		t.Error("Version should be true")
	}
}

func TestParseFlags_TestFiltering(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(testFile, []byte("version: 1.0\ntests: []"), 0644); err != nil {
		t.Fatal(err)
	}

	// Reset flag state
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Set args with filtering flags
	os.Args = []string{
		"curlex",
		"--test", "MyTest",
		"--test-pattern", "Test.*",
		"--skip", "SkipThis",
		testFile,
	}

	cfg, err := ParseFlags()
	if err != nil {
		t.Fatalf("ParseFlags() unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("ParseFlags() returned nil config")
	}

	if cfg.TestFilter != "MyTest" {
		t.Errorf("TestFilter = %s, want MyTest", cfg.TestFilter)
	}
	if cfg.TestPattern != "Test.*" {
		t.Errorf("TestPattern = %s, want Test.*", cfg.TestPattern)
	}
	if cfg.SkipTests != "SkipThis" {
		t.Errorf("SkipTests = %s, want SkipThis", cfg.SkipTests)
	}
}

func TestParseFlags_LogDir(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(testFile, []byte("version: 1.0\ntests: []"), 0644); err != nil {
		t.Fatal(err)
	}

	logDir := filepath.Join(tmpDir, "logs")

	// Reset flag state
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Set args with log-dir flag
	os.Args = []string{
		"curlex",
		"--log-dir", logDir,
		testFile,
	}

	cfg, err := ParseFlags()
	if err != nil {
		t.Fatalf("ParseFlags() unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("ParseFlags() returned nil config")
	}

	if cfg.LogDir != logDir {
		t.Errorf("LogDir = %s, want %s", cfg.LogDir, logDir)
	}
}
