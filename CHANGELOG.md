# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - 2026-11-18

### Added

- `GenerateWithOptions(Options) (Result, error)`: one entry point for words, language and separator, returning the passphrase and its rolls.
- `diceware completion bash|zsh|fish|powershell`, completion of `--lang`, and a man page (`man/diceware.1`, installed by the Homebrew cask and in the release archives).
- `NOTICE` file with the licences of the embedded wordlists, shipped in the release archives.
- `MaxWords` (1024): the Generate functions reject larger word counts.
- Godoc examples, fuzz tests for `ParseLanguage` and the wordlist parser, CLI tests, and a chi-square test of the dice rolls.
- CodeQL workflow, issue and pull request templates, build provenance for release archives.

### Changed

- **Mixed mode draws uniformly from the 15,030 distinct words of both lists.** About 280 words (Absurd, Album, Atlas, ...) are in both lists and look the same once capitalized; they now count once, so a mixed passphrase has slightly less entropy than reported before (83.3 bits for 6 words, not 83.4) and every word is equally likely.
- In Mixed mode the dice rolls that `-r` and the `Rolls` results report are prefixed with their list, `en:46122` or `ro:33544`, because the same digits mean a different word in each list.
- The wordlists are checked for all 7,776 rolls at startup.
- Release archives are named `diceware_...` (were `go-diceware_...`).
- A release without `HOMEBREW_TAP_TOKEN` publishes the archives and only skips the cask.
- `just build-all` runs GoReleaser in snapshot mode; `just vuln` runs govulncheck and `just check` includes it.

### Removed

- `IMPACT.md` and `examples/impact-analysis`: outdated, and the example wrote to fixed paths in `/tmp` without using the library.
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

## [0.2.4] - 2026-07-13

### Changed

- The CLI uses cobra instead of the `flag` package.

## [0.2.3] - 2026-07-08

### Added

- A separator-aware API, `GenerateWithRollsLanguageAndSeparator`, so the CLI no longer guesses word boundaries.

### Fixed

- `Entropy` ignored the language, which gave wrong bits for Romanian and Mixed.
- `capitalize` is rune-safe instead of slicing the first byte.
- A panic risk and a precision bug in the impact-analysis example.

## [0.2.2] - 2026-02-28

### Fixed

- Several bugs and missing input validation.

## [0.2.1] - 2026-02-28

### Fixed

- The Romanian wordlist no longer yields non-alphabetic entries (numbers and symbols).

## [0.2.0] - 2026-02-28

### Added

- Romanian wordlist and multi-language passphrase generation, including Mixed mode.

## [0.1.0] - 2026-02-28

### Added

- Diceware passphrase generator: library and CLI with the EFF large wordlist.

[Unreleased]: https://github.com/cleonte/go-diceware/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/cleonte/go-diceware/compare/v0.2.6...v0.3.0
[0.2.6]: https://github.com/cleonte/go-diceware/compare/v0.2.5...v0.2.6
[0.2.5]: https://github.com/cleonte/go-diceware/compare/v0.2.4...v0.2.5
[0.2.4]: https://github.com/cleonte/go-diceware/compare/v0.2.3...v0.2.4
[0.2.3]: https://github.com/cleonte/go-diceware/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/cleonte/go-diceware/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/cleonte/go-diceware/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/cleonte/go-diceware/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/cleonte/go-diceware/releases/tag/v0.1.0
