# go-diceware

A Go implementation of the Diceware passphrase generation method using the EFF large wordlist, inspired by [diceware.dmuth.org](https://diceware.dmuth.org/). 

This library provides both a Go package for integration into your applications and a command-line tool for generating secure, memorable passphrases with **capitalized words** in **CamelCase format** (no separators by default).

## Features

- **Cryptographically Secure**: Uses Go's `crypto/rand` for true random number generation
- **EFF Large Wordlist**: Uses the [EFF's improved wordlist](https://www.eff.org/deeplinks/2016/07/new-wordlists-random-passphrases) with 7,776 carefully selected words
- **Capitalized CamelCase**: Words are capitalized and concatenated by default (e.g., `ColtDefaultArousal`)
- **Library and CLI**: Use it as a Go library in your code or as a standalone CLI tool
- **Flexible**: Customize word count and separators
- **Well-Tested**: Comprehensive test suite with >85% coverage
- **Zero Dependencies**: Only uses Go standard library

## Installation

### As a Library

```bash
go get github.com/cleonte/go-diceware
```

### As a CLI Tool

```bash
go install github.com/cleonte/go-diceware/cmd/diceware@latest
```

Or build from source:

```bash
git clone https://github.com/cleonte/go-diceware.git
cd go-diceware
go build -o diceware ./cmd/diceware
```

## Usage

### CLI Tool

Generate a passphrase with default settings (6 capitalized words, no separator):

```bash
$ diceware
ColtDefaultArousalThimbleGaslightYearbook

Entropy: 77.6 bits (6 words from 7776 word list)
```

Specify number of words:

```bash
$ diceware -w 4
EfficientSpottyLaurelPhony

Entropy: 51.7 bits (4 words from 7776 word list)
```

Use a space separator for easier reading:

```bash
$ diceware -w 4 -s " "
Reclining Clapping Frugality Slackness

Entropy: 51.7 bits (4 words from 7776 word list)
```

Use a dash separator:

```bash
$ diceware -w 4 -s "-"
Sterile-Ascent-Barmaid-Plunge

Entropy: 51.7 bits (4 words from 7776 word list)
```

Show dice rolls used to generate the passphrase:

```bash
$ diceware -r -w 3
Dice rolls: [46122 33544 21546]
Passphrase: PuritanHatlessCubicle

Entropy: 38.8 bits (3 words from 7776 word list)
```

### Library Usage

#### Basic Example

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/cleonte/go-diceware"
)

func main() {
    // Generate a 6-word passphrase (capitalized, no separator)
    passphrase, err := diceware.Generate(6)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(passphrase)
    // Output: ColtDefaultArousalThimbleGaslightYearbook
}
```

#### Custom Separator

```go
// Generate with space separator for easier reading
passphrase, err := diceware.GenerateWithSeparator(4, " ")
if err != nil {
    log.Fatal(err)
}
fmt.Println(passphrase)
// Output: Palpable Sandpaper Barber Unmasking

// Generate with dash separator
passphrase, err = diceware.GenerateWithSeparator(4, "-")
if err != nil {
    log.Fatal(err)
}
fmt.Println(passphrase)
// Output: Palpable-Sandpaper-Barber-Unmasking
```

#### With Dice Rolls

```go
// Generate and see the dice rolls used
passphrase, rolls, err := diceware.GenerateWithRolls(4)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Passphrase: %s\n", passphrase)
fmt.Printf("Dice rolls: %v\n", rolls)
// Output:
// Passphrase: PalpableSandpaperBarberUnmasking
// Dice rolls: [43434 52653 13252 62345]
```

#### Calculate Entropy

```go
// Calculate the entropy for a given number of words
entropy := diceware.Entropy(6)
fmt.Printf("Entropy: %.1f bits\n", entropy)
// Output: Entropy: 77.6 bits
```

## Security Considerations

### Recommended Word Counts

The security of your passphrase depends on the number of words:

| Words | Entropy | Use Case |
|-------|---------|----------|
| 4 | ~52 bits | Minimum for low-value accounts |
| 6 | ~78 bits | Recommended for most accounts |
| 8 | ~103 bits | High security accounts |
| 12 | ~155 bits | Cryptocurrency wallets (minimum) |
| 24 | ~310 bits | Maximum paranoia |

### Why Diceware?

- **Memorable**: Humans are better at remembering phrases than random characters
- **Secure**: High entropy when using sufficient words
- **Typeable**: Easier to type than special characters, especially on mobile devices or smart TVs
- **Auditable**: The randomness comes from a well-understood cryptographic source

### When NOT to Use Diceware

- **Against offline attacks**: If an attacker can obtain your encrypted password and perform unlimited cracking attempts, use a password manager with longer random passwords
- **When length matters**: Some systems have password length restrictions that make multi-word passphrases impractical

## How It Works

1. **Rolling Dice**: The library uses Go's `crypto/rand` to simulate rolling five 6-sided dice
2. **Looking Up Words**: Each 5-digit number (e.g., "43434") corresponds to a word in the EFF wordlist
3. **Combining Words**: The words are joined together with your chosen separator
4. **Entropy**: Each word adds ~12.925 bits of entropy (log₂(7776) ≈ 12.925)

## API Reference

### Functions

#### `Generate(wordCount int) (string, error)`

Generates a passphrase with the specified number of words, separated by spaces.

#### `GenerateWithSeparator(wordCount int, separator string) (string, error)`

Generates a passphrase with a custom separator between words.

#### `GenerateWithRolls(wordCount int) (passphrase string, rolls []string, err error)`

Generates a passphrase and returns the dice rolls used to create it.

#### `Entropy(wordCount int) float64`

Calculates the bits of entropy for a given number of words.

#### `WordlistSize() int`

Returns the number of words in the wordlist (7,776).

## Development

This project uses [just](https://github.com/casey/just) as a command runner (modern alternative to make).

### Quick Commands

```bash
# List all available commands
just

# Run tests
just test

# Build the CLI
just build

# Run all checks (lint + test)
just check

# Generate a passphrase quickly
just generate 8

# Run the example program
just example
```

### Manual Commands

If you don't have `just` installed, you can run commands directly:

```bash
# Run tests
go test -v

# Run benchmarks
go test -bench=.

# Test coverage
go test -cover
```

### Installing Just

```bash
# macOS
brew install just

# Linux
cargo install just

# Or download from: https://github.com/casey/just
```

## About Diceware

Diceware is a method for creating passphrases, originally developed by Arnold Reinhold. This implementation uses the [EFF's improved wordlist](https://www.eff.org/deeplinks/2016/07/new-wordlists-random-passphrases), which was designed to be more memorable and easier to type than the original list.

For more information:
- [Original Diceware](https://theworld.com/~reinhold/diceware.html)
- [EFF Wordlist Announcement](https://www.eff.org/deeplinks/2016/07/new-wordlists-random-passphrases)
- [XKCD: Password Strength](https://xkcd.com/936/)

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Acknowledgments

- Arnold Reinhold for creating the Diceware method
- The EFF for creating an improved, more user-friendly wordlist
- Inspired by [diceware.dmuth.org](https://diceware.dmuth.org/)
