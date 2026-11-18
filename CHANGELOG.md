# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `NOTICE` file with the licences of the embedded wordlists, shipped in the release archives.
- `MaxWords` (1024): the Generate functions reject larger word counts.
- Fuzz tests for `ParseLanguage` and the wordlist parser; a chi-square test of the dice rolls.
- CodeQL workflow, issue and pull request templates, build provenance for release archives.

### Changed

- The wordlists are checked for all 7,776 rolls at startup.
- `just build-all` runs GoReleaser in snapshot mode; `just vuln` runs govulncheck; `just check` includes it.

### Removed

- The unused `rollDice` helper and its test.

## [0.2.6] - 2026-11-18

### Changed

- The CLI accepts up to 24 words (was 20), as the README recommends.
- `--version` falls back to the module version for builds from `go install`.
- Updated pflag to v1.1.0.

### Fixed

- Two tests split passphrases on `-`, which occurs inside some EFF words (`yo-yo`, `t-shirt`), so they failed at random.
- README: CI badge URL, Romanian word count (7,535 usable), coverage and dependency claims; the package documentation example now compiles.

## [0.2.5] - 2026-11-18

### Added

- `--version` flag.
- Release archives for Linux, macOS and Windows (amd64 and arm64) built by GoReleaser, and a Homebrew cask (`brew install cleonte/tap/diceware`).
- golangci-lint v2 in CI, and badges in the README.
