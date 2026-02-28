// Package diceware provides cryptographically secure passphrase generation
// using the Diceware method with the EFF large wordlist.
//
// The Diceware method generates passphrases by rolling five dice to create
// a 5-digit number, which is then used to look up a word in a wordlist.
// This process is repeated for each word in the passphrase.
//
// Example usage:
//
//	// Generate a passphrase with 6 words
//	passphrase, err := diceware.Generate(6)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(passphrase)
//
//	// Generate with custom separator
//	passphrase, err := diceware.GenerateWithSeparator(6, "-")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(passphrase)
package diceware

import (
	"crypto/rand"
	_ "embed"
	"fmt"
	"math/big"
	"strings"
)

//go:embed internal/wordlist/eff_large_wordlist.txt
var wordlistData string

var wordlist map[string]string

func init() {
	wordlist = parseWordlist(wordlistData)
}

// parseWordlist parses the embedded wordlist file into a map
func parseWordlist(data string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(data, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}

	return result
}

// rollDice simulates rolling a single die (1-6) using cryptographically secure random numbers
func rollDice() (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(6))
	if err != nil {
		return 0, fmt.Errorf("failed to generate random number: %w", err)
	}
	return int(n.Int64()) + 1, nil
}

// rollFiveDice rolls five dice and returns the result as a string (e.g., "11111")
func rollFiveDice() (string, error) {
	var result strings.Builder
	for i := 0; i < 5; i++ {
		roll, err := rollDice()
		if err != nil {
			return "", err
		}
		result.WriteString(fmt.Sprintf("%d", roll))
	}
	return result.String(), nil
}

// getWord rolls five dice and returns the corresponding word from the wordlist,
// capitalized to match the Diceware web implementation
func getWord() (string, error) {
	roll, err := rollFiveDice()
	if err != nil {
		return "", err
	}

	word, exists := wordlist[roll]
	if !exists {
		return "", fmt.Errorf("no word found for dice roll: %s", roll)
	}

	// Capitalize first letter
	return capitalize(word), nil
}

// capitalize returns the word with the first letter capitalized
func capitalize(word string) string {
	if len(word) == 0 {
		return word
	}
	return strings.ToUpper(word[:1]) + word[1:]
}

// Generate creates a passphrase with the specified number of words.
// Words are capitalized and concatenated with no separator by default,
// matching the diceware.dmuth.org implementation.
// The number of words should be at least 4 for reasonable security,
// with 6-8 words recommended for most use cases.
//
// Example output: "ColtDefaultArousalThimble"
//
// Entropy:
//   - 4 words: ~51.7 bits
//   - 5 words: ~64.6 bits
//   - 6 words: ~77.5 bits (recommended minimum)
//   - 7 words: ~90.4 bits
//   - 8 words: ~103.3 bits
//
// Returns an error if wordCount is less than 1 or if random number generation fails.
func Generate(wordCount int) (string, error) {
	return GenerateWithSeparator(wordCount, "")
}

// GenerateWithSeparator creates a passphrase with the specified number of words
// and joins them with the provided separator. Words are capitalized.
//
// Common separators:
//   - "" (empty) - no separator, CamelCase style (default)
//   - " " (space) - easier to read
//   - "-" (dash) - good for URLs
//
// Returns an error if wordCount is less than 1 or if random number generation fails.
func GenerateWithSeparator(wordCount int, separator string) (string, error) {
	if wordCount < 1 {
		return "", fmt.Errorf("word count must be at least 1, got %d", wordCount)
	}

	words := make([]string, wordCount)
	for i := 0; i < wordCount; i++ {
		word, err := getWord()
		if err != nil {
			return "", fmt.Errorf("failed to generate word %d: %w", i+1, err)
		}
		words[i] = word
	}

	return strings.Join(words, separator), nil
}

// GenerateWithRolls returns both the passphrase and the dice rolls used to generate it.
// Words are capitalized and concatenated with no separator.
// This can be useful for verification or debugging purposes.
//
// Returns a passphrase, a slice of dice roll strings, and an error.
func GenerateWithRolls(wordCount int) (passphrase string, rolls []string, err error) {
	if wordCount < 1 {
		return "", nil, fmt.Errorf("word count must be at least 1, got %d", wordCount)
	}

	words := make([]string, wordCount)
	rolls = make([]string, wordCount)

	for i := 0; i < wordCount; i++ {
		roll, err := rollFiveDice()
		if err != nil {
			return "", nil, fmt.Errorf("failed to generate dice roll %d: %w", i+1, err)
		}

		word, exists := wordlist[roll]
		if !exists {
			return "", nil, fmt.Errorf("no word found for dice roll: %s", roll)
		}

		words[i] = capitalize(word)
		rolls[i] = roll
	}

	return strings.Join(words, ""), rolls, nil
}

// Entropy calculates the bits of entropy for a given number of words.
// The EFF large wordlist has 7,776 words (6^5), providing ~12.925 bits per word.
func Entropy(wordCount int) float64 {
	// log2(7776) ≈ 12.925 bits per word
	const bitsPerWord = 12.925
	return float64(wordCount) * bitsPerWord
}

// WordlistSize returns the number of words in the wordlist
func WordlistSize() int {
	return len(wordlist)
}
