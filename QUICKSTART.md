# Quick Start Guide

## Installation

```bash
git clone https://github.com/cleonte/go-diceware.git
cd go-diceware
```

## Using the CLI

### Build the tool

```bash
just build
# or: go build -o diceware ./cmd/diceware
```

### Generate passphrases

```bash
# Default: 6 words, capitalized, no separator (CamelCase)
./diceware
# Output: ChewyMonitorCarelessRoundwormSynapseGuileless

# 4 words for simpler passwords
./diceware -w 4
# Output: EfficientSpottyLaurelPhony

# With space separator for easier reading
./diceware -w 4 -s " "
# Output: Reclining Clapping Frugality Slackness

# With dash separator
./diceware -w 4 -s "-"
# Output: Sterile-Ascent-Barmaid-Plunge

# Show dice rolls used
./diceware -r -w 5
# Output:
# Dice rolls: [46122 33544 21546 12345 54321]
# Passphrase: PuritanHatlessCubicleAcornZebra
```

### Quick generate (with just)

```bash
just generate 8
# Output: UncookedDeceptiveSliverStatusFreezableGoneUndesiredDiscourse
```

## Using as a Library

### Install

```bash
go get github.com/cleonte/go-diceware
```

### Use in your code

```go
package main

import (
    "fmt"
    "log"
    "github.com/cleonte/go-diceware"
)

func main() {
    // Generate passphrase (capitalized, no separator by default)
    passphrase, err := diceware.Generate(6)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(passphrase)
    // Output: ChewyMonitorCarelessRoundwormSynapseGuileless
    
    // With space separator for easier reading
    passphrase, err = diceware.GenerateWithSeparator(4, " ")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(passphrase)
    // Output: Reclining Clapping Frugality Slackness
}
```

## Development

### Run tests

```bash
just test
```

### Run all checks

```bash
just check
```

### See all commands

```bash
just
```

## Output Format

By default, words are **capitalized** and **concatenated without separators** (CamelCase style), matching the diceware.dmuth.org website:

- Default: `ChewyMonitorCareless`
- With spaces: `Chewy Monitor Careless`
- With dashes: `Chewy-Monitor-Careless`

## Security Recommendations

| Words | Entropy | Use Case |
|-------|---------|----------|
| 4 | ~52 bits | Low-value accounts |
| 6 | ~78 bits | Most accounts (recommended) |
| 8 | ~103 bits | High security |
| 12 | ~155 bits | Cryptocurrency wallets |

## Next Steps

- Read the full [README.md](README.md)
- Check out [examples/basic/main.go](examples/basic/main.go)
- Read [CONTRIBUTING.md](CONTRIBUTING.md) to contribute
