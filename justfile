# justfile - go-filewatcher
# Usage: GOWORK=off just <recipe>

# Default recipe
default: build

# Build the project
build:
    GOWORK=off go build ./...

# Run all tests with race detector
test:
    GOWORK=off go test -race ./...

# Run tests with coverage report
test-cover:
    GOWORK=off go test -race -coverprofile=coverage.out ./...
    GOWORK=off go tool cover -html=coverage.out -o coverage.html

# Run tests with verbose output
test-v:
    GOWORK=off go test -v -race ./...

# Run linter
lint:
    GOWORK=off golangci-lint run ./...

# Run linter with auto-fix
lint-fix:
    GOWORK=off golangci-lint run --fix ./...

# Run go vet
vet:
    GOWORK=off go vet ./...

# Run all quality checks (vet + lint + test)
check: vet lint test

# Clean build cache
clean:
    go clean -cache

# Tidy dependencies
tidy:
    GOWORK=off go mod tidy

# Format code
fmt:
    go fmt ./...

# Run benchmarks
bench:
    GOWORK=off go test -bench=. -benchmem ./...

# Generate test coverage report
coverage:
    GOWORK=off go test -coverprofile=coverage.out ./...
    GOWORK=off go tool cover -func=coverage.out

# Show test coverage summary
coverage-summary:
    GOWORK=off go test -cover ./...

# Run all tests and show detailed output
ci: tidy fmt vet lint test

# Watch mode (requires fswatch or similar)
watch:
    fswatch -r . --exclude=.git | xargs -I{} just test

# Install dependencies
install:
    go mod download

# Build for all platforms
build-all:
    GOOS=darwin GOARCH=arm64 GOWORK=off go build -o bin/darwin-arm64 ./...
    GOOS=darwin GOARCH=amd64 GOWORK=off go build -o bin/darwin-amd64 ./...
    GOOS=linux GOARCH=arm64 GOWORK=off go build -o bin/linux-arm64 ./...
    GOOS=linux GOARCH=amd64 GOWORK=off go build -o bin/linux-amd64 ./...
    GOOS=windows GOARCH=amd64 GOWORK=off go build -o bin/windows-amd64.exe ./...

# Remove built binaries
clean-bin:
    rm -rf bin/

# Release build
release: clean-bin build-all

# Initialize project (first run)
init: tidy install
