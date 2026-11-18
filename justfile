# go-diceware justfile
# See https://github.com/casey/just for more info

# Default recipe to display help
default:
    @just --list

# Build the CLI binary
build:
    @echo "Building diceware..."
    @go build -o diceware ./cmd/diceware

# Run tests with race detection and coverage
test:
    @echo "Running tests..."
    @go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Run benchmarks
bench:
    @echo "Running benchmarks..."
    @go test -bench=. -benchmem ./...

# Generate and open coverage report
coverage: test
    @echo "Generating coverage report..."
    @go tool cover -html=coverage.txt -o coverage.html
    @echo "Coverage report saved to coverage.html"

# Clean build artifacts
clean:
    @echo "Cleaning..."
    @rm -f diceware
    @rm -f coverage.txt coverage.html
    @rm -f *.test

# Install the CLI tool
install:
    @echo "Installing diceware..."
    @go install ./cmd/diceware

# Run linters and formatters
lint:
    @echo "Running linters..."
    @go vet ./...
    @golangci-lint run ./...
    @test -z "$(gofmt -l .)" || (echo "Unformatted files:"; gofmt -l .; exit 1)

# Run all checks before commit
check: lint vuln build-examples test
    @echo "✅ All checks passed!"

# Ensure all examples compile
build-examples:
    @go build -o /dev/null ./examples/...

# Run example program
example:
    @cd examples/basic && go run main.go

# Generate a passphrase (shortcut)
generate words="6":
    @go run ./cmd/diceware -w {{words}}

# Run tests in watch mode (requires entr)
watch:
    @echo "Watching for changes... (requires 'entr' to be installed)"
    @find . -name '*.go' | entr -c just test

# Show project statistics
stats:
    @echo "📊 Project Statistics:"
    @echo "Lines of code:"
    @find . -name '*.go' -not -path "*/vendor/*" | xargs wc -l | tail -1
    @echo ""
    @echo "Test coverage:"
    @go test -cover ./... | grep coverage

# Update dependencies (keeps the go directive in go.mod; check it afterwards)
update:
    @echo "Updating dependencies..."
    @go get -u ./...
    @go mod tidy
    @git diff --stat go.mod

# Verify dependencies
verify:
    @echo "Verifying dependencies..."
    @go mod verify

# Check dependencies for known vulnerabilities (same as CI)
vuln:
    @go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...

# Build the release archives locally, without publishing (needs goreleaser)
build-all:
    @HOMEBREW_TAP_TOKEN= goreleaser release --snapshot --clean --skip=publish
