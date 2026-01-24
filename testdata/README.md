# curlex Test Examples

This directory contains example YAML files demonstrating various features of curlex.

## Basic Examples

### simple.yaml
Basic test examples demonstrating:
- Simple GET requests
- Status code assertions
- Basic curl command parsing

### advanced.yaml
Advanced features including:
- All assertion types (status, body, json_path, header, response_time)
- Variable substitution
- Structured request format
- Complex test scenarios

## Phase 3 Features

### retry-tests.yaml
Demonstrates retry functionality:
- Default retry settings
- Custom retry configuration
- Exponential vs linear backoff
- Retry on specific status codes
- Retry delay configuration

**Usage:**
```bash
# Run with default settings
curlex testdata/retry-tests.yaml

# Enable request/response logging
curlex testdata/retry-tests.yaml --log-dir ./logs

# Run in verbose mode to see retry attempts
curlex testdata/retry-tests.yaml --verbose
```

### phase3-complete.yaml
Comprehensive example showing all Phase 3 features:
- Retry logic with different strategies
- Defaults merging and override
- Request/response logging (with --log-dir flag)
- Test filtering examples

**Usage:**
```bash
# Run all tests
curlex testdata/phase3-complete.yaml

# Filter by pattern - run only API tests
curlex testdata/phase3-complete.yaml --test-pattern "^API"

# Filter by pattern - run only Auth tests
curlex testdata/phase3-complete.yaml --test-pattern "Auth"

# Skip specific test
curlex testdata/phase3-complete.yaml --skip "Failure Test - Will retry"

# Run specific test by exact name
curlex testdata/phase3-complete.yaml --test "API Test - GET with retries"

# Enable logging to see full request/response details
curlex testdata/phase3-complete.yaml --log-dir ./logs

# Combine filtering and logging
curlex testdata/phase3-complete.yaml --test-pattern "^API" --log-dir ./logs --verbose
```

## Feature Demonstrations

### Retry Logic
```yaml
defaults:
  retries: 2
  retry_delay: "1s"
  retry_backoff: "exponential"  # or "linear"
  retry_on_status: [500, 502, 503, 504]

tests:
  - name: "Flaky endpoint"
    curl: "curl https://api.example.com/flaky"
    retries: 3  # Override default
    assertions:
      - status: 200
```

### Request/Response Logging
```bash
# Enable logging with --log-dir flag
curlex tests.yaml --log-dir ./logs

# Log files are created with format:
# YYYY-MM-DD_HH-MM-SS_test-name.log

# Each log contains:
# - Full request (method, URL, headers, body)
# - Full response (status, headers, body)
# - Assertion results
# - Any errors
```

### Test Filtering
```bash
# Run specific test by exact name
curlex tests.yaml --test "Login test"

# Run tests matching regex pattern
curlex tests.yaml --test-pattern "^API.*"

# Skip specific test
curlex tests.yaml --skip "Slow test"

# Combine pattern matching and skipping
curlex tests.yaml --test-pattern "^API" --skip "API Slow Test"
```

### Defaults Merging
```yaml
defaults:
  timeout: 30s
  retries: 2
  headers:
    User-Agent: "curlex/1.0.0"

tests:
  - name: "Inherits defaults"
    curl: "curl https://example.com"
    # Uses: timeout=30s, retries=2, User-Agent header

  - name: "Overrides timeout"
    curl: "curl https://example.com/slow"
    timeout: 60s
    # Uses: timeout=60s (override), retries=2 (default)

  - name: "Adds custom header"
    request:
      method: GET
      url: "https://example.com"
      headers:
        X-Custom: "value"
    # Headers merged: User-Agent (default) + X-Custom (test-specific)
```

## Running Tests

### Sequential (default)
```bash
curlex testdata/advanced.yaml
```

### Parallel
```bash
curlex testdata/advanced.yaml --parallel --concurrency 5
```

### With Output Formats
```bash
# Human-readable (default)
curlex testdata/advanced.yaml

# JSON output for CI/CD
curlex testdata/advanced.yaml --output json

# JUnit XML for test reporting
curlex testdata/advanced.yaml --output junit

# Quiet mode - minimal output
curlex testdata/advanced.yaml --quiet

# Verbose mode - detailed output
curlex testdata/advanced.yaml --verbose
```

### Fail-Fast Mode
```bash
# Stop on first failure
curlex testdata/advanced.yaml --fail-fast
```

## Tips

1. **Use --verbose to debug**: See full request/response details and retry attempts
2. **Use --log-dir for persistence**: Save full details to files for later analysis
3. **Filter tests during development**: Use --test-pattern to run specific test groups
4. **Combine parallel and fail-fast**: Find failures quickly with --parallel --fail-fast
5. **Use defaults for common settings**: Set shared configuration once in defaults section
