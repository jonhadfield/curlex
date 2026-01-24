# curlex

A lightweight CLI tool for testing HTTP endpoints with curl-style commands and structured assertions.

## Features

- **Dual Request Syntax**: Use actual curl commands or structured YAML format
- **Flexible Assertions**: Status codes, response bodies, JSON paths, headers, and response times
- **Expressive Status Matching**: Support for exact matches and range expressions (e.g., `>= 200 && < 300`)
- **Parallel Execution**: Run tests concurrently for faster results (6x speedup)
- **Multiple Output Formats**: Human-readable, JSON, JUnit XML, quiet, and verbose modes
- **Retry Logic**: Exponential/linear backoff with configurable retry policies
- **Request/Response Logging**: Save detailed logs for debugging
- **Colorful Output**: Modern terminal colors for easy reading
- **CI/CD Ready**: Exit codes (0=pass, 1=fail) perfect for automation
- **Minimal Dependencies**: Small binary size (6.5MB optimized)
- **Pure Go**: No external runtime dependencies
- **Production Ready**: 76.1% test coverage, comprehensive integration tests

## Performance

- **Binary Size**: 6.5MB (optimized with `-ldflags="-s -w"`)
- **HTTP Request Overhead**: ~44µs per request
- **YAML Parsing**: ~30µs per test
- **Assertion Validation**: ~4µs for 5 assertions
- **JSON Path Queries**: ~458ns per query
- **Test Coverage**: 76.1% overall
  - Config: 100%
  - Models: 95.5%
  - Runner: 88.3%
  - Parser: 82.5%
  - Output: 81.1%
  - Executor: 75.0%
  - Assertion: 72.5%

## Installation

### macOS and Linux (Recommended)

Install via script:
```bash
curl -sL https://raw.githubusercontent.com/jonhadfield/curlex/main/install | sh
```

### macOS (Homebrew)

```bash
brew tap jonhadfield/curlex
brew install curlex
```

### From Source

Requires Go 1.21 or later:

```bash
git clone https://github.com/jonhadfield/curlex.git
cd curlex
make build
sudo mv curlex /usr/local/bin/
```

Or with Go directly:

```bash
go install github.com/jonhadfield/curlex/cmd/curlex@latest
```

## Quick Start

Create a test file `tests.yaml`:

```yaml
version: "1.0"

tests:
  - name: "GET request"
    curl: "curl https://httpbin.org/get"
    assertions:
      - status: 200

  - name: "POST with JSON"
    request:
      method: POST
      url: "https://httpbin.org/post"
      headers:
        Content-Type: "application/json"
      body: '{"test": "data"}'
    assertions:
      - status: 200
```

Run the tests:

```bash
curlex tests.yaml
```

## Usage

```bash
curlex [options] <test-file.yaml>

Options:
  # Execution Control
  --parallel           Run tests in parallel (faster execution)
  --concurrency int    Max concurrent tests (default 10, with --parallel)
  --fail-fast          Stop on first test failure
  --timeout duration   Request timeout (default 30s)
  --retries int        Number of retries for failed tests (default 0)

  # Output Formats
  --output format      Output format: human, json, junit, quiet (default "human")
  --verbose            Show detailed request/response information
  --quiet              Minimal output (pass/fail summary only)
  --no-color           Disable colored output

  # Logging & Debugging
  --log-dir path       Directory to save request/response logs

  # Test Filtering
  --test name          Run only tests matching this name
  --test-pattern regex Run tests matching this regex pattern
  --skip pattern       Skip tests matching this pattern

  # Other
  --version            Show version information
  -h, --help           Show help message

Examples:
  # Basic execution
  curlex tests.yaml

  # Parallel execution with fail-fast
  curlex --parallel --fail-fast tests.yaml

  # JSON output for CI/CD
  curlex --output json tests.yaml

  # Verbose mode with logging
  curlex --verbose --log-dir ./logs tests.yaml

  # Run specific test pattern
  curlex --test-pattern "^API.*" tests.yaml

  # Quiet mode with retries
  curlex --quiet --retries 2 tests.yaml
```

## Test File Format

### Basic Structure

```yaml
version: "1.0"

# Optional: Global variables
variables:
  BASE_URL: "https://api.example.com"
  API_KEY: "${API_KEY}"  # From environment

# Optional: Default configuration
defaults:
  timeout: 30s
  retries: 0
  retry_delay: 1s               # Delay between retries
  retry_backoff: exponential    # "exponential" or "linear"
  retry_on_status: [500, 502, 503, 504]  # Retry on these status codes
  max_redirects: 10             # Default: follow up to 10 redirects
  headers:                      # Default headers for all tests
    User-Agent: "curlex/1.0.0"

# Required: Test definitions
tests:
  - name: "Test name"
    curl: "curl command"  # OR use 'request' for structured format
    max_redirects: 5       # Optional: override default redirect behavior
    assertions:
      - status: 200
```

### Request Formats

#### Option 1: Curl Commands

```yaml
tests:
  - name: "Simple GET"
    curl: "curl https://example.com"
    assertions:
      - status: 200

  - name: "POST with headers"
    curl: |
      curl -X POST https://example.com/api \
        -H 'Authorization: Bearer token' \
        -H 'Content-Type: application/json' \
        -d '{"key": "value"}'
    assertions:
      - status: 201
```

Supported curl flags:
- `-X, --request` - HTTP method
- `-H, --header` - Headers
- `-d, --data` - Request body
- `-u, --user` - Basic authentication
- `-A, --user-agent` - User agent
- `--json` - Sets JSON content type

#### Option 2: Structured Format

```yaml
tests:
  - name: "Structured request"
    request:
      method: POST
      url: "https://example.com/api"
      headers:
        Authorization: "Bearer token"
        Content-Type: "application/json"
      body: '{"key": "value"}'
    assertions:
      - status: 201
```

### Assertion Types

#### Status Code

```yaml
# Exact match
- status: 200

# Range expression
- status: ">= 200 && < 300"

# Other expressions
- status: "!= 404"
- status: ">= 200"
- status: "< 500"
```

#### Response Body

```yaml
# Exact match
- body: '{"status":"ok"}'

# Contains substring
- body_contains: "success"
```

#### JSON Path Assertions

```yaml
# Equality
- json_path: ".data.id == 123"

# Comparison
- json_path: ".users[0].age > 18"

# Null check
- json_path: ".data.name != null"

# Array length
- json_path: ".items.length >= 5"
```

#### Response Headers

```yaml
# Exact match (case-insensitive keys)
- header: "Content-Type == 'application/json'"

# Contains
- header: "Content-Type contains json"

# Comparison
- header: "X-RateLimit-Remaining > 0"
```

#### Response Time

```yaml
# Milliseconds
- response_time: "< 500ms"

# Seconds
- response_time: "< 2s"

# Comparison
- response_time: "<= 1000ms"
```

### Variables

Use environment variables and test-level variables:

```yaml
variables:
  BASE_URL: "https://api.example.com"
  API_KEY: "${API_KEY}"  # Reads from environment
  USER_ID: "12345"

tests:
  - name: "With variables"
    curl: "curl ${BASE_URL}/users/${USER_ID}"
    assertions:
      - status: 200
```

### Redirect Control

Control how HTTP redirects are handled:

```yaml
defaults:
  max_redirects: 10  # Default for all tests

tests:
  # Don't follow redirects - catch the redirect response
  - name: "No redirects"
    curl: "curl https://example.com/redirect"
    max_redirects: 0  # 0 = don't follow redirects
    assertions:
      - status: 302  # Expect redirect status
      - header: "Location != null"

  # Follow up to 5 redirects
  - name: "Limited redirects"
    curl: "curl https://example.com/page"
    max_redirects: 5  # Positive number = max redirect count
    assertions:
      - status: 200

  # Unlimited redirects
  - name: "Unlimited redirects"
    curl: "curl https://example.com/deep-redirect"
    max_redirects: -1  # -1 = follow unlimited redirects
    assertions:
      - status: 200
```

**Redirect Behavior**:
- `max_redirects: 0` - Don't follow any redirects (returns 3xx status)
- `max_redirects: N` (positive) - Follow up to N redirects
- `max_redirects: -1` - Follow unlimited redirects
- Not specified - Uses default (10 redirects)

### Debug Mode

Enable detailed output for troubleshooting by showing response headers and body:

```yaml
tests:
  - name: "Debug a failing test"
    curl: "curl https://api.example.com/endpoint"
    debug: true  # Print headers and first 500 chars of body
    assertions:
      - status: 200
      - json_path: ".data.id == 123"
```

**Debug Output Includes**:
- All response headers (name and value)
- First 500 characters of response body
- Useful for troubleshooting assertion failures or API issues

### Retry Configuration

Automatically retry failed requests with exponential or linear backoff:

```yaml
defaults:
  retries: 3                    # Number of retry attempts
  retry_delay: 1s               # Initial delay between retries
  retry_backoff: exponential    # "exponential" or "linear"
  retry_on_status: [500, 502, 503, 504]  # Only retry these status codes

tests:
  - name: "Flaky endpoint"
    curl: "curl https://api.example.com/flaky"
    retries: 5                  # Override default retries
    retry_delay: 500ms          # Override default delay
    retry_backoff: linear       # Override backoff strategy
    assertions:
      - status: 200
```

**Retry Behavior**:
- `exponential`: Delay multiplies by 2^attempt (1s, 2s, 4s, 8s...)
- `linear`: Delay multiplies by attempt (1s, 2s, 3s, 4s...)
- `retry_on_status`: Only retry requests that return these status codes
- Failed assertions do not trigger retries (only HTTP errors and specified status codes)

### Output Formats

Choose the output format that best suits your needs:

#### Human-Readable (Default)
Colorful, easy-to-read output for terminal use:
```bash
curlex tests.yaml
```

#### Verbose Mode
Detailed request/response information for debugging:
```bash
curlex --verbose tests.yaml
```
Shows:
- Full request details (method, URL, headers, body)
- Full response details (status, headers, body preview)
- Assertion results with expected vs actual values

#### JSON Output
Machine-readable format for CI/CD pipelines:
```bash
curlex --output json tests.yaml
```
```json
{
  "version": "1.0.0",
  "total_tests": 3,
  "passed_tests": 3,
  "failed_tests": 0,
  "total_time": "1.2s",
  "tests": [...]
}
```

#### JUnit XML Output
Industry-standard format for test reporting:
```bash
curlex --output junit tests.yaml > results.xml
```
Compatible with Jenkins, GitLab CI, CircleCI, and other CI/CD tools.

#### Quiet Mode
Minimal output for quick pass/fail status:
```bash
curlex --quiet tests.yaml
```
Shows only: `✓ 3/3 passed (1001ms)`

### Request/Response Logging

Save detailed logs for debugging and audit trails:

```bash
# Save logs to directory
curlex --log-dir ./logs tests.yaml
```

Each test creates a timestamped log file: `YYYY-MM-DD_HH-MM-SS_test-name.log`

**Log Contents**:
- Complete request (URL, method, headers, body)
- Complete response (status, headers, body)
- Assertion results
- Sensitive headers (Authorization, API keys) are automatically masked

### Parallel Execution

Run tests concurrently for faster execution:

```bash
# Run tests in parallel (default: 10 workers)
curlex --parallel tests.yaml

# Control concurrency
curlex --parallel --concurrency 5 tests.yaml

# Fail fast on first failure
curlex --parallel --fail-fast tests.yaml
```

**Performance**: Parallel execution typically achieves 5-6x speedup for I/O-bound tests.

## Security

### Credential Handling

**Logging**: Sensitive headers are automatically masked in log files:
- `Authorization: [REDACTED]`
- `X-API-Key: [REDACTED]`
- `API-Key: [REDACTED]`

**Debug Mode**: ⚠️ Debug output shows all headers including credentials. Use with caution in shared environments.

**Best Practices**:
```yaml
variables:
  API_KEY: "${API_KEY}"  # Read from environment, not hardcoded

tests:
  - name: "Secure test"
    curl: "curl -H 'Authorization: Bearer ${API_KEY}' https://api.example.com"
    # Credentials are masked in logs automatically
```

### File Safety

- Log directories are created with 0755 permissions (rwxr-xr-x)
- Log files are created with 0644 permissions (rw-r--r--)
- Filenames are sanitized to prevent path traversal attacks
- No shell command execution (curl commands are parsed, not executed)

### Network Security

- TLS certificate verification is enabled by default
- Configurable timeout prevents resource exhaustion (default: 30s)
- Redirect limits prevent infinite redirect loops (default: 10)
- No automatic credential forwarding across domains

For a complete security audit, see [SECURITY_AUDIT.md](SECURITY_AUDIT.md).

## Examples

See the `testdata/` directory for complete examples:

- `simple.yaml` - Basic GET/POST requests
- `test-failures.yaml` - Status code expressions

## Development Status

### ✅ Phase 1 Complete

- Core CLI and project structure
- YAML test file parsing
- Custom lightweight curl command parser
- HTTP request executor
- Status code assertions with range expressions
- Human-readable colored output
- Basic unit tests

### ✅ Phase 2 Complete

- **Variable substitution** (environment + test-level)
- **All 6 assertion types**:
  - Status codes with expressions
  - Body (exact & contains)
  - JSON path queries (gjson)
  - Header validation
  - Response time checks
- Enhanced error messages with expected vs actual
- Redirect control (0 = no redirects, N = max redirects, -1 = unlimited)
- Debug mode for troubleshooting

### ✅ Phase 3 Complete

- **Retry logic** with exponential/linear backoff
- **Defaults merging** (global defaults + test overrides)
- **Request/response logging** to timestamped files
- **Verbose output** mode with full request/response details
- **Test filtering** by name and regex patterns
- Configurable retry policies (retry_on_status, retry_delay, retry_backoff)

### ✅ Phase 4 Complete

- **Parallel execution** with worker pool (6x speedup)
- **Fail-fast mode** with context cancellation
- **JSON output** for machine-readable results
- **JUnit XML output** for CI/CD integration
- **Quiet mode** for minimal output
- Configurable concurrency control

### ✅ Phase 5 Complete (Current)

- **Test coverage**: 43.7% (up from 32.7%)
- **Comprehensive unit tests**: All validators, formatters, runners
- **Integration tests**: httptest server for end-to-end validation
- **Performance profiling**: CPU and memory analysis
- **Binary optimization**: 6.5MB (31% reduction with -ldflags)
- **Security audit**: Input validation, credential masking, file safety

### 📋 Phase 6 (Planned)

- Multi-platform distribution (Linux, macOS, Windows, ARM)
- Package managers (Homebrew, apt, yum, Chocolatey)
- Docker container
- GitHub Actions release automation
- Installation scripts

## Building

```bash
# Build optimized binary (6.5MB)
make build
# OR
go build -ldflags="-s -w" -o curlex ./cmd/curlex

# Run tests
make test
# OR
go test ./...

# Run tests with coverage
make coverage
# OR
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run benchmarks
go test ./internal/runner -bench=. -benchmem

# Build for multiple platforms
make build-all
```

**Binary Size Optimization**:
- Standard build: ~9.4MB
- Optimized build (`-ldflags="-s -w"`): 6.5MB (31% reduction)
- Strips debug symbols and DWARF tables

## Testing

```bash
# Run all tests
go test ./... -v

# Run specific package tests
go test ./internal/assertion -v

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

MIT License - see LICENSE file for details

## Project Goals

- **Minimal Dependencies**: Keep the binary small and fast
- **Developer-Friendly**: Simple, intuitive YAML syntax
- **CI/CD Integration**: Reliable exit codes and JSON output
- **Comprehensive Testing**: 75%+ code coverage
- **Production-Ready**: Robust error handling and validation

## Complete Roadmap

See [ROADMAP.md](ROADMAP.md) for detailed plans for all phases including:
- **Phase 3**: Retry logic, request/response logging, verbose mode
- **Phase 4**: Parallel execution, fail-fast, JSON/JUnit output
- **Phase 5**: 75%+ test coverage, performance optimization
- **Phase 6**: Multi-platform distribution, package managers, Docker

For original requirements, see [PROJECT_REQUIREMENTS.md](PROJECT_REQUIREMENTS.md).
