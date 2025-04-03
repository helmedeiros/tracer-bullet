.PHONY: test test-coverage build clean lint pre-push

# Build the tracer binary
build:
	go build -o tracer cmd/tracer/main.go

# Run all tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f tracer
	rm -f coverage.out
	rm -f coverage.html

# Run linter
lint:
	golangci-lint run --timeout 5m

# Install development dependencies
dev-deps:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.64.8

# Run all checks (lint + test)
check: lint test

# Run pre-push checks
pre-push: check
