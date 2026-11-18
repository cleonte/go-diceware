# Contributing to go-diceware

Thank you for considering contributing to go-diceware! This document provides guidelines for contributing to the project.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally
3. Create a new branch for your changes
4. Make your changes
5. Run tests and linting
6. Submit a pull request

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git
- [just](https://github.com/casey/just) (optional but recommended)

### Building

```bash
# Build the CLI (with just)
just build

# Or manually
go build -o diceware ./cmd/diceware
```

### Running Tests

```bash
# Run all tests (with just)
just test

# Run tests with verbose output
go test -v

# Run tests with coverage
just coverage

# Run benchmarks
just bench
```

### Code Quality

Before submitting a pull request, ensure your code passes all checks:

```bash
# Run all checks (with just)
just check

# This runs:
# - go vet and golangci-lint run ./...
# - a gofmt check
# - an examples build
# - tests
```

Linting uses [golangci-lint](https://golangci-lint.run/) v2 (config in `.golangci.yml`); run it directly with:

```bash
golangci-lint run ./...
```

## Contribution Guidelines

### Code Style

- Follow standard Go conventions
- Use `go fmt` to format your code
- Write clear, descriptive variable and function names
- Add comments for exported functions and types
- Keep functions focused and reasonably sized

### Testing

- Add tests for new functionality
- Ensure existing tests pass
- Aim for high test coverage (>80%)
- Include both positive and negative test cases
- Use table-driven tests where appropriate

### Documentation

- Update README.md if adding new features
- Add godoc comments for exported functions
- Include examples for new functionality
- Add an entry under "Unreleased" in CHANGELOG.md

### Commit Messages

- Use clear, descriptive commit messages
- Start with a verb in imperative mood (e.g., "Add", "Fix", "Update")
- Keep the first line under 72 characters
- Add detailed explanation in the body if needed

Example:
```
Add support for custom wordlists

- Allow users to provide their own wordlist files
- Add validation for wordlist format
- Include tests for custom wordlist loading
```

### Pull Requests

- Create a focused PR that addresses a single concern
- Reference any related issues
- Provide a clear description of the changes
- Ensure all tests pass
- Update documentation as needed
- Sign off every commit (`git commit -s`): by doing so you certify the
  [Developer Certificate of Origin](https://developercertificate.org/)
- See `AGENTS.md` for the layout

## What to Contribute

### Good First Issues

- Documentation improvements
- Additional examples
- Test coverage improvements
- Bug fixes

### Feature Ideas

Before working on a major feature, please open an issue to discuss it first. This ensures:
- The feature aligns with project goals
- No duplicate work is being done
- We can discuss the best approach

### Bug Reports

When reporting bugs, please include:
- Go version
- Operating system
- Steps to reproduce
- Expected behavior
- Actual behavior
- Any error messages

## Security Issues

If you discover a security vulnerability, please follow [SECURITY.md](SECURITY.md) and report it privately rather than opening a public issue.

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers
- Accept constructive criticism
- Focus on what's best for the project

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Questions?

Feel free to open an issue if you have questions about contributing!

## Releasing

Pushing a tag `vX.Y.Z` runs the release workflow, which builds the archives (Linux, macOS and Windows, on amd64 and arm64) with [GoReleaser](https://goreleaser.com) and publishes the release. The body of the release is the section of `CHANGELOG.md` for that version (`scripts/release-notes.sh vX.Y.Z` prints it; without a section GoReleaser writes the notes from the commits).

Before tagging:

1. Move what is under "Unreleased" in `CHANGELOG.md` into a section `## [X.Y.Z] - date`.
2. Check locally: `goreleaser check`, then `goreleaser release --snapshot --clean --skip=publish`.
3. Tag and push: `git tag -a vX.Y.Z -m "..." && git push origin vX.Y.Z`.

The release also updates the Homebrew cask in [cleonte/homebrew-tap](https://github.com/cleonte/homebrew-tap) (`homebrew_casks` in `.goreleaser.yaml`). That needs the repository secret `HOMEBREW_TAP_TOKEN`: a fine-grained personal access token with **Contents: read and write on the tap repository only**. Without it the release fails at the publish step, after the archives are built.
