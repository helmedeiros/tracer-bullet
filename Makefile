.PHONY: test test-coverage build clean lint

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
	golangci-lint run

# Install development dependencies
dev-deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run all checks (lint + test)
check: lint test
