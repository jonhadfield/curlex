.PHONY: build test clean install run coverage

# Binary name
BINARY_NAME=curlex
VERSION=1.0.1

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/curlex
	@echo "Build complete: ./$(BINARY_NAME)"

# Run all tests
test:
	@echo "Running tests..."
	@go test ./... -v

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Install to /usr/local/bin
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo mv $(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete"

# Run with example test file
run: build
	@echo "Running example tests..."
	@./$(BINARY_NAME) testdata/simple.yaml

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY_NAME)-linux-amd64 ./cmd/curlex
	@GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY_NAME)-darwin-amd64 ./cmd/curlex
	@GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BINARY_NAME)-darwin-arm64 ./cmd/curlex
	@GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY_NAME)-windows-amd64.exe ./cmd/curlex
	@echo "Multi-platform build complete"

# Show version
version:
	@echo "$(BINARY_NAME) version $(VERSION)"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@golangci-lint run

# Show help
help:
	@echo "Available commands:"
	@echo "  make build      - Build the binary"
	@echo "  make test       - Run all tests"
	@echo "  make coverage   - Run tests with coverage report"
	@echo "  make clean      - Remove build artifacts"
	@echo "  make install    - Install to /usr/local/bin"
	@echo "  make run        - Build and run example tests"
	@echo "  make build-all  - Build for multiple platforms"
	@echo "  make fmt        - Format code"
	@echo "  make lint       - Run linter"
	@echo "  make version    - Show version"
