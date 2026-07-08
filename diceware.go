// Package diceware provides cryptographically secure passphrase generation
// using the Diceware method with the EFF large wordlist or Romanian wordlist.
//
// The Diceware method generates passphrases by rolling five dice to create
// a 5-digit number, which is then used to look up a word in a wordlist.
// This process is repeated for each word in the passphrase.
//
// Example usage:
//
//	// Generate a passphrase with 6 words (English)
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
//
//	// Generate Romanian passphrase
//	passphrase, err := diceware.GenerateWithLanguage(6, LanguageRomanian)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(passphrase)
//
//	// Generate mixed English and Romanian passphrase
//	passphrase, err := diceware.GenerateWithLanguage(6, LanguageMixed)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(passphrase)
package diceware

import (
	"crypto/rand"
	_ "embed"
	"fmt"
	"math"
	"math/big"
	"strings"
	"unicode"
	"unicode/utf8"
)

//go:embed internal/wordlist/eff_large_wordlist.txt
var wordlistEnglishData string

//go:embed internal/wordlist/ro_diceware.txt
var wordlistRomanianData string

var wordlistEnglish map[string]string
var wordlistRomanian map[string]string

// validWordCountEnglish and validWordCountRomanian track how many entries in
// each wordlist actually get used to produce a word (i.e., how many survive
// isValidWord). English entries are never filtered during generation, so its
// count always equals len(wordlistEnglish). Romanian's raw map includes ~241
// filler entries (digits/symbols used to fill out all 7,776 roll
// combinations) that getWordFromLanguage rerolls past, so its usable count is
// lower than len(wordlistRomanian). These are the counts that must be used
// for entropy/size reporting, not the raw map length.
var validWordCountEnglish int
var validWordCountRomanian int

// Language represents the language for passphrase generation
type Language int

const (
	// LanguageEnglish generates passphrases using only English words
	LanguageEnglish Language = iota
	// LanguageRomanian generates passphrases using only Romanian words
	LanguageRomanian
	// LanguageMixed generates passphrases using a mix of English and Romanian words
	LanguageMixed
)

func init() {
	wordlistEnglish = parseWordlist(wordlistEnglishData)
	wordlistRomanian = parseWordlist(wordlistRomanianData)

	// English words are used as-is (no isValidWord filtering during
	// generation), so every parsed entry is usable.
	validWordCountEnglish = len(wordlistEnglish)

	// Romanian entries that fail isValidWord get rerolled at generation
	// time and can never appear in output, so only count the ones that
	// would actually be selectable.
	for _, word := range wordlistRomanian {
		if isValidWord(word) {
			validWordCountRomanian++
		}
	}
}

// parseWordlist parses the embedded wordlist file into a map
// It validates the format and panics if the wordlist is malformed
func parseWordlist(data string) map[string]string {
	result := make(map[string]string, 7776) // Pre-allocate for expected size
	lines := strings.Split(data, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) != 2 {
			panic(fmt.Sprintf("invalid wordlist format at line %d: expected 2 fields, got %d: %q", i+1, len(parts), line))
		}

		roll := parts[0]
		word := parts[1]

		// Validate roll format (5 digits, each 1-6)
		if !isValidRoll(roll) {
			panic(fmt.Sprintf("invalid dice roll at line %d: %q (expected 5 digits between 1-6)", i+1, roll))
		}

		// Check for duplicate rolls
		if _, exists := result[roll]; exists {
			panic(fmt.Sprintf("duplicate dice roll at line %d: %q", i+1, roll))
		}

		result[roll] = word
	}

	return result
}

// isValidRoll checks if a roll string is valid (5 digits, each 1-6)
func isValidRoll(roll string) bool {
	if len(roll) != 5 {
		return false
	}
	for _, ch := range roll {
		if ch < '1' || ch > '6' {
			return false
		}
	}
	return true
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
	return getWordFromLanguage(LanguageEnglish)
}

// isValidWord checks if a word contains only alphabetic characters (no numbers, symbols, etc.)
func isValidWord(word string) bool {
	if len(word) < 3 {
		return false
	}
	for _, ch := range word {
		if (ch < 'a' || ch > 'z') && (ch < 'A' || ch > 'Z') {
			return false
		}
	}
	return true
}

// rollWord rolls five dice and resolves them to a word for the specified
// language, rerolling internally (up to maxAttempts) if the roll lands on a
// filtered/invalid entry - e.g. Romanian's ~241 numeric/symbol filler
// entries. Returns the raw (uncapitalized) word alongside the winning dice
// roll string. This is the single place the reroll/language-selection logic
// lives; getWordFromLanguage and GenerateWithRollsAndLanguage both build on
// top of it instead of duplicating the switch/reroll logic.
func rollWord(lang Language) (word string, roll string, err error) {
	const maxAttempts = 100 // Prevent infinite loops

	for attempt := 0; attempt < maxAttempts; attempt++ {
		roll, err = rollFiveDice()
		if err != nil {
			return "", "", err
		}

		var exists bool

		switch lang {
		case LanguageEnglish:
			word, exists = wordlistEnglish[roll]
		case LanguageRomanian:
			word, exists = wordlistRomanian[roll]
			// Romanian wordlist contains some non-word entries (numbers, symbols)
			// Re-roll if we get one of those
			if exists && !isValidWord(word) {
				continue
			}
		case LanguageMixed:
			// For mixed mode, randomly choose between English and Romanian
			useBool, berr := rand.Int(rand.Reader, big.NewInt(2))
			if berr != nil {
				return "", "", fmt.Errorf("failed to select language: %w", berr)
			}
			if useBool.Int64() == 0 {
				word, exists = wordlistEnglish[roll]
			} else {
				word, exists = wordlistRomanian[roll]
				// Re-roll if we get a non-word from Romanian wordlist
				if exists && !isValidWord(word) {
					continue
				}
			}
		default:
			return "", "", fmt.Errorf("unsupported language: %v", lang)
		}

		if !exists {
			return "", "", fmt.Errorf("no word found for dice roll: %s", roll)
		}

		return word, roll, nil
	}

	return "", "", fmt.Errorf("failed to generate valid word after %d attempts", maxAttempts)
}

// getWordFromLanguage rolls five dice and returns the corresponding word from the specified language wordlist,
// capitalized to match the Diceware web implementation. For Romanian, it re-rolls if it gets a non-alphabetic
// entry (numbers, symbols, etc.) - see rollWord.
func getWordFromLanguage(lang Language) (string, error) {
	word, _, err := rollWord(lang)
	if err != nil {
		return "", err
	}
	return capitalize(word), nil
}

// capitalize returns the word with the first letter capitalized
// capitalize returns the word with the first letter capitalized. It decodes
// the first rune rather than slicing the first byte, so multi-byte UTF-8
// characters (e.g. accented letters) are capitalized correctly instead of
// being corrupted. Currently a no-op concern for the shipped wordlists (no
// surviving entry starts with a multi-byte rune), but wordlists change.
func capitalize(word string) string {
	if word == "" {
		return word
	}
	r, size := utf8.DecodeRuneInString(word)
	if r == utf8.RuneError && size <= 1 {
		// Invalid encoding at the start of the string - leave untouched
		// rather than risk further corruption.
		return word
	}
	return string(unicode.ToUpper(r)) + word[size:]
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
	return GenerateWithLanguageAndSeparator(wordCount, LanguageEnglish, separator)
}

// GenerateWithLanguage creates a passphrase with the specified number of words
// using the specified language(s). Words are capitalized and concatenated with no separator.
//
// Languages:
//   - LanguageEnglish - English words only
//   - LanguageRomanian - Romanian words only
//   - LanguageMixed - Mix of English and Romanian words
//
// Returns an error if wordCount is less than 1 or if random number generation fails.
func GenerateWithLanguage(wordCount int, lang Language) (string, error) {
	return GenerateWithLanguageAndSeparator(wordCount, lang, "")
}

// GenerateWithLanguageAndSeparator creates a passphrase with the specified number of words
// using the specified language(s) and joins them with the provided separator.
//
// Returns an error if wordCount is less than 1 or if random number generation fails.
func GenerateWithLanguageAndSeparator(wordCount int, lang Language, separator string) (string, error) {
	if wordCount < 1 {
		return "", fmt.Errorf("word count must be at least 1, got %d", wordCount)
	}

	words := make([]string, wordCount)
	for i := 0; i < wordCount; i++ {
		word, err := getWordFromLanguage(lang)
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
	return GenerateWithRollsAndLanguage(wordCount, LanguageEnglish)
}

// GenerateWithRollsAndLanguage returns both the passphrase and the dice rolls used to generate it
// using the specified language(s). Words are capitalized and concatenated with no separator.
//
// Returns a passphrase, a slice of dice roll strings, and an error.
func GenerateWithRollsAndLanguage(wordCount int, lang Language) (passphrase string, rolls []string, err error) {
	if wordCount < 1 {
		return "", nil, fmt.Errorf("word count must be at least 1, got %d", wordCount)
	}

	words := make([]string, wordCount)
	rolls = make([]string, wordCount)

	for i := 0; i < wordCount; i++ {
		word, roll, werr := rollWord(lang)
		if werr != nil {
			return "", nil, fmt.Errorf("failed to generate word %d: %w", i+1, werr)
		}
		words[i] = capitalize(word)
		rolls[i] = roll
	}

	return strings.Join(words, ""), rolls, nil
}

// Entropy calculates the bits of entropy for a given number of words,
// assuming the English wordlist. Equivalent to
// EntropyForLanguage(wordCount, LanguageEnglish).
//
// The EFF large wordlist has 7,776 usable words (6^5), providing ~12.925
// bits per word.
func Entropy(wordCount int) float64 {
	return EntropyForLanguage(wordCount, LanguageEnglish)
}

// EntropyForLanguage calculates the bits of entropy for a given number of
// words in the specified language. Unlike Entropy, this accounts for the
// fact that different wordlists have different usable sizes:
//
//   - English: 7,776 words, ~12.925 bits/word
//   - Romanian: 7,535 usable words (241 filler entries are skipped during
//     generation), ~12.879 bits/word
//   - Mixed: 15,311 usable words combined (English + valid Romanian),
//     ~13.902 bits/word, since each word also carries the extra bit from
//     the English/Romanian coin flip
func EntropyForLanguage(wordCount int, lang Language) float64 {
	size := WordlistSizeByLanguage(lang)
	if size < 2 {
		return 0
	}
	return float64(wordCount) * math.Log2(float64(size))
}

// WordlistSize returns the number of usable words in the English wordlist
func WordlistSize() int {
	return validWordCountEnglish
}

// WordlistSizeByLanguage returns the number of usable words in the wordlist
// for the specified language, i.e., the number of distinct dice rolls that
// actually produce a word during generation (not the raw entry count -
// Romanian's raw wordlist includes ~241 filler entries that are skipped).
func WordlistSizeByLanguage(lang Language) int {
	switch lang {
	case LanguageEnglish:
		return validWordCountEnglish
	case LanguageRomanian:
		return validWordCountRomanian
	case LanguageMixed:
		// Mixed mode selects with a fair coin flip between the two
		// wordlists and rerolls the whole attempt (coin + dice) if it
		// lands on an invalid Romanian entry. That rejection sampling
		// preserves uniformity, so the combined usable space really is
		// just the sum of both usable counts.
		return validWordCountEnglish + validWordCountRomanian
	default:
		return 0
	}
}
