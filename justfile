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
    @go test -v -race -coverprofile=coverage.txt -covermode=atomic

# Run benchmarks
bench:
    @echo "Running benchmarks..."
    @go test -bench=. -benchmem

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
    @go fmt ./...

# Run all checks before commit
check: lint test
    @echo "✅ All checks passed!"

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
    @go test -cover | grep coverage

# Update dependencies
update:
    @echo "Updating dependencies..."
    @go get -u ./...
    @go mod tidy

# Verify dependencies
verify:
    @echo "Verifying dependencies..."
    @go mod verify

# Run security checks (requires gosec)
security:
    @echo "Running security checks..."
    @which gosec > /dev/null || (echo "Install gosec: go install github.com/securego/gosec/v2/cmd/gosec@latest" && exit 1)
    @gosec ./...

# Build for multiple platforms
build-all:
    @echo "Building for multiple platforms..."
    @GOOS=linux GOARCH=amd64 go build -o dist/diceware-linux-amd64 ./cmd/diceware
    @GOOS=darwin GOARCH=amd64 go build -o dist/diceware-darwin-amd64 ./cmd/diceware
    @GOOS=darwin GOARCH=arm64 go build -o dist/diceware-darwin-arm64 ./cmd/diceware
    @GOOS=windows GOARCH=amd64 go build -o dist/diceware-windows-amd64.exe ./cmd/diceware
    @echo "✅ Built binaries in dist/"
