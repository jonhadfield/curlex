package config

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// Config holds the CLI configuration
type Config struct {
	TestFile     string
	Timeout      time.Duration
	NoColor      bool
	Version      bool
	Verbose      bool
	LogDir       string
	TestFilter   string
	TestPattern  string
	SkipTests    string
	Parallel     bool
	Concurrency  int
	FailFast     bool
	OutputFormat string
	Quiet        bool
}

// ParseFlags parses command-line flags and returns configuration
func ParseFlags() (*Config, error) {
	cfg := &Config{}

	flag.DurationVar(&cfg.Timeout, "timeout", 30*time.Second, "Request timeout (e.g., 30s, 1m)")
	flag.BoolVar(&cfg.NoColor, "no-color", false, "Disable colored output")
	flag.BoolVar(&cfg.Version, "version", false, "Show version information")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Enable verbose output")
	flag.StringVar(&cfg.LogDir, "log-dir", "", "Directory to save request/response logs")
	flag.StringVar(&cfg.TestFilter, "test", "", "Run specific test by exact name")
	flag.StringVar(&cfg.TestPattern, "test-pattern", "", "Run tests matching regex pattern")
	flag.StringVar(&cfg.SkipTests, "skip", "", "Skip tests matching name")
	flag.BoolVar(&cfg.Parallel, "parallel", false, "Run tests in parallel")
	flag.IntVar(&cfg.Concurrency, "concurrency", 10, "Max concurrent tests when using --parallel")
	flag.BoolVar(&cfg.FailFast, "fail-fast", false, "Stop on first test failure")
	flag.StringVar(&cfg.OutputFormat, "output", "human", "Output format: human, json, junit, quiet")
	flag.BoolVar(&cfg.Quiet, "quiet", false, "Minimal output (summary only)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: curlex [options] <test-file.yaml>\n\n")
		fmt.Fprintf(os.Stderr, "A CLI tool for testing HTTP endpoints with curl-style commands and structured assertions.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  curlex tests.yaml\n")
		fmt.Fprintf(os.Stderr, "  curlex --timeout 60s tests.yaml\n")
		fmt.Fprintf(os.Stderr, "  curlex --no-color tests.yaml\n")
	}

	flag.Parse()

	// Handle version flag
	if cfg.Version {
		return cfg, nil
	}

	// Get test file from remaining args
	if flag.NArg() < 1 {
		flag.Usage()
		return nil, fmt.Errorf("missing required argument: test-file.yaml")
	}

	cfg.TestFile = flag.Arg(0)

	// Validate test file exists
	if _, err := os.Stat(cfg.TestFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("test file does not exist: %s", cfg.TestFile)
	}

	return cfg, nil
}
