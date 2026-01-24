# Multi-stage build for minimal final image
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binary with optimization flags
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o curlex \
    ./cmd/curlex

# Final minimal image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 curlex && \
    adduser -D -u 1000 -G curlex curlex

# Set working directory
WORKDIR /home/curlex

# Copy binary from builder
COPY --from=builder /build/curlex /usr/local/bin/curlex

# Switch to non-root user
USER curlex

# Set entrypoint
ENTRYPOINT ["curlex"]

# Default command (show help)
CMD ["--help"]
