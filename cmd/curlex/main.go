package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"curlex/internal/config"
	"curlex/internal/models"
	"curlex/internal/output"
	"curlex/internal/parser"
	"curlex/internal/runner"
)

const version = "1.0.0"

func main() {
	// Parse CLI flags
	cfg, err := config.ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Handle version flag
	if cfg.Version {
		fmt.Printf("curlex version %s\n", version)
		os.Exit(0)
	}

	// Run tests
	exitCode := run(cfg)
	os.Exit(exitCode)
}

func run(cfg *config.Config) int {
	// Create YAML parser
	yamlParser := parser.NewYAMLParser()

	// Parse test suite
	suite, err := yamlParser.Parse(cfg.TestFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse test file: %v\n", err)
		return 1
	}

	// Apply test filtering if configured
	filterConfig := runner.FilterConfig{
		TestName:    cfg.TestFilter,
		TestPattern: cfg.TestPattern,
		SkipTests:   cfg.SkipTests,
	}
	suite.Tests = runner.FilterTests(suite, filterConfig)

	// Check if any tests remain after filtering
	if len(suite.Tests) == 0 {
		fmt.Fprintf(os.Stderr, "No tests match the specified filter criteria\n")
		return 1
	}

	// Create runner
	testRunner := runner.NewRunner(cfg.Timeout, cfg.LogDir)

	// Create progress indicator for human/verbose output (not quiet, json, junit)
	var progress *output.Progress
	showProgress := (cfg.OutputFormat == "human" || cfg.OutputFormat == "" || cfg.Verbose) && !cfg.Quiet && cfg.OutputFormat != "json" && cfg.OutputFormat != "junit"
	if showProgress {
		progress = output.NewProgress(len(suite.Tests), cfg.NoColor)
		testRunner.SetProgress(progress)
		progress.Start()
	}

	// Execute tests (parallel or sequential) with graceful shutdown on SIGINT/SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var suiteResult *models.SuiteResult
	if cfg.Parallel {
		suiteResult, err = testRunner.RunParallel(ctx, suite, cfg.Concurrency, cfg.FailFast)
	} else {
		suiteResult, err = testRunner.Run(ctx, suite)
	}

	// Stop progress indicator
	if progress != nil {
		progress.Stop()
	}

	// Check if execution was interrupted
	if ctx.Err() == context.Canceled {
		fmt.Fprintf(os.Stderr, "\nTest execution interrupted - returning partial results\n")
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to execute tests: %v\n", err)
		return 1
	}

	// Handle output based on format
	if cfg.Quiet || cfg.OutputFormat == "quiet" {
		// Quiet mode - minimal output
		quietFormatter := output.NewQuietFormatter(cfg.NoColor)
		fmt.Print(quietFormatter.FormatSummary(suiteResult.Results, suiteResult.TotalTime))
	} else {
		switch cfg.OutputFormat {
		case "json":
			jsonFormatter := output.NewJSONFormatter()
			fmt.Print(jsonFormatter.Format(suiteResult))
		case "junit":
			junitFormatter := output.NewJUnitFormatter()
			fmt.Print(junitFormatter.Format(suiteResult))
		default: // "human" or verbose
			var formatter interface {
				FormatResult(models.TestResult) string
				FormatSummary([]models.TestResult, time.Duration) string
			}
			if cfg.Verbose {
				formatter = output.NewVerboseFormatter(cfg.NoColor)
			} else {
				formatter = output.NewHumanFormatter(cfg.NoColor)
			}

			// Output results
			for _, result := range suiteResult.Results {
				fmt.Print(formatter.FormatResult(result))
			}

			// Output summary
			fmt.Print(formatter.FormatSummary(suiteResult.Results, suiteResult.TotalTime))
		}
	}

	// Return exit code
	if suiteResult.HasFailures() {
		return 1
	}
	return 0
}
