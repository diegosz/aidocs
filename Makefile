.PHONY: build test lint fmt install clean

# Build the aidocs binary
build:
	go build -o bin/aidocs ./cmd/aidocs

# Install aidocs to GOPATH/bin
install:
	go install ./cmd/aidocs

# Run tests
test:
	go test -v ./...

# Run tests with coverage
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	gofumpt -w .

# Run linter
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html

# Run aidocs on test fixtures
test-blind:
	cd testdata/blind && go run ../../cmd/aidocs/main.go -v

# Download dependencies
deps:
	go mod download
	go mod tidy

# Generate documentation using aidocs (dogfooding)
gendocs:
	go run ./cmd/aidocs -v
