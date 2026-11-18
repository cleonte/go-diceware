# AGENTS.md

Guidance for AI agents (and humans) working in this repository.

## Overview

`go-diceware` is a Go implementation of the Diceware passphrase method with
English and Romanian wordlists, as a library (`diceware.go`) and a CLI
(`cmd/diceware`). Passphrases are CamelCase by default; randomness comes from
`crypto/rand` only.

## Build and test

Only the Go toolchain is needed.

```sh
go build -o diceware ./cmd/diceware
go vet ./...
golangci-lint run ./...   # v2, configured in .golangci.yml
go test -race ./...
```

`just check` runs the same checks plus the examples build.

## Architecture

* `diceware.go`: the library API (generation, `Language`, dice rolls).
* `internal/wordlist`: the embedded wordlists.
* `cmd/diceware`: the cobra CLI; `version` is set at release time by
  `-ldflags "-X main.version=..."`.
* `scripts/genro`: generates `internal/wordlist/ro_diceware.txt` (see its README); do not edit that list by hand.
* `examples/`: small programs that must keep compiling.

## Style

`gofmt`, `go vet` and `golangci-lint` clean. Comments explain *why* only.
`go.mod` keeps the oldest Go that the library supports; CI tests it
(`test-min-go`).

## Commits

Describe the change in a short paragraph and add a `Signed-off-by:` line.

## Releasing

Move "Unreleased" in `CHANGELOG.md` into a `## [X.Y.Z] - date` section (it becomes
the release notes), then `git tag -a vX.Y.Z -m "..." && git push origin vX.Y.Z`.
The `release` workflow runs GoReleaser (`.goreleaser.yaml`): archives,
`checksums.txt` and the Homebrew cask in `cleonte/homebrew-tap`, which needs the
`HOMEBREW_TAP_TOKEN` secret. Check with `goreleaser check`.
