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
//	passphrase, err = diceware.GenerateWithSeparator(6, "-")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(passphrase)
//
//	// Generate Romanian passphrase
//	passphrase, err = diceware.GenerateWithLanguage(6, diceware.LanguageRomanian)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(passphrase)
//
//	// Generate mixed English and Romanian passphrase
//	passphrase, err = diceware.GenerateWithLanguage(6, diceware.LanguageMixed)
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
	"sort"
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
// count always equals len(wordlistEnglish). Entries that fail isValidWord
// are rerolled past during generation, so a list that has some (the previous
// Romanian list had 241 filler entries) has a usable count below its map
// length. These are the counts that must be used for entropy/size reporting,
// not the raw map length.
var validWordCountEnglish int
var validWordCountRomanian int

// mixedEntry is one distinct word of the Mixed pool and where it came from.
type mixedEntry struct {
	word string
	tag  string // "en" or "ro"
	roll string
}

// mixedPool holds every distinct usable word of both lists. About 280 words
// (Absurd, Album, Atlas, ...) are in both lists and look identical once
// capitalized, so they appear once, as English entries. Mixed mode draws
// uniformly from this pool, which makes its size and entropy exact.
var mixedPool []mixedEntry

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

// String returns the display name of the language.
func (l Language) String() string {
	switch l {
	case LanguageEnglish:
		return "English"
	case LanguageRomanian:
		return "Romanian"
	case LanguageMixed:
		return "Mixed (English + Romanian)"
	default:
		return fmt.Sprintf("Language(%d)", int(l))
	}
}

// ParseLanguage converts a user-supplied name (case-insensitive) to a
// Language. Accepted: "en"/"english", "ro"/"romanian", "mixed"/"mix".
func ParseLanguage(name string) (Language, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "en", "english":
		return LanguageEnglish, nil
	case "ro", "romanian":
		return LanguageRomanian, nil
	case "mixed", "mix":
		return LanguageMixed, nil
	default:
		return 0, fmt.Errorf("unsupported language %q (use: en, ro, or mixed)", name)
	}
}

// MaxWords is the largest word count the Generate functions accept; it keeps
// a mistaken argument from allocating an enormous slice.
const MaxWords = 1024

func init() {
	wordlistEnglish = parseWordlist(wordlistEnglishData)
	wordlistRomanian = parseWordlist(wordlistRomanianData)

	// A missing roll would only show up later, as a runtime error for the
	// unlucky draw, so fail at startup instead.
	for name, list := range map[string]map[string]string{"English": wordlistEnglish, "Romanian": wordlistRomanian} {
		if len(list) != 7776 {
			panic(fmt.Sprintf("%s wordlist has %d entries, want 7776", name, len(list)))
		}
	}

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

	mixedPool = buildMixedPool(wordlistEnglish, wordlistRomanian)
}

// buildMixedPool merges the usable words of both lists, dropping Romanian
// words that match an English word ignoring case. The order is by dice roll,
// English first, so the pool is the same on every run.
func buildMixedPool(english, romanian map[string]string) []mixedEntry {
	pool := make([]mixedEntry, 0, len(english)+len(romanian))
	seen := make(map[string]bool, len(english)+len(romanian))
	add := func(list map[string]string, tag string) {
		rolls := make([]string, 0, len(list))
		for roll := range list {
			rolls = append(rolls, roll)
		}
		sort.Strings(rolls)
		for _, roll := range rolls {
			word := list[roll]
			key := strings.ToLower(word)
			if (tag == "ro" && !isValidWord(word)) || seen[key] {
				continue
			}
			seen[key] = true
			pool = append(pool, mixedEntry{word: word, tag: tag, roll: roll})
		}
	}
	add(english, "en")
	add(romanian, "ro")
	return pool
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

// rollFiveDice rolls five dice and returns the result as a string (e.g., "11111").
// It draws a single uniform value in [0, 6^5) and renders it as five base-6
// digits shifted to 1-6, which is equivalent to five independent d6 rolls
// but needs one read from the CSPRNG instead of five.
func rollFiveDice() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(7776))
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}
	v := int(n.Int64())
	var digits [5]byte
	for i := 4; i >= 0; i-- {
		digits[i] = byte('1' + v%6)
		v /= 6
	}
	return string(digits[:]), nil
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
// filtered/invalid entry (one that fails isValidWord). Returns the raw (uncapitalized) word alongside the winning dice
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
			// Mixed mode ignores the dice roll above and draws one entry
			// uniformly from the pool of distinct words. The roll that is
			// reported is the entry's roll in its own list, tagged with it.
			idx, berr := rand.Int(rand.Reader, big.NewInt(int64(len(mixedPool))))
			if berr != nil {
				return "", "", fmt.Errorf("failed to select word: %w", berr)
			}
			e := mixedPool[idx.Int64()]
			return e.word, e.tag + ":" + e.roll, nil
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
	if wordCount > MaxWords {
		return "", fmt.Errorf("word count must be at most %d, got %d", MaxWords, wordCount)
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
	return GenerateWithRollsLanguageAndSeparator(wordCount, lang, "")
}

// GenerateWithRollsLanguageAndSeparator returns both the passphrase and the dice rolls used to
// generate it, using the specified language(s) and joining the words with the provided separator.
//
// This is the separator-aware counterpart to GenerateWithRollsAndLanguage - use it instead of
// post-processing an already-joined passphrase to insert a separator, since that requires
// guessing word boundaries back out of the capitalized string.
//
// Returns a passphrase, a slice of dice roll strings, and an error.
func GenerateWithRollsLanguageAndSeparator(wordCount int, lang Language, separator string) (passphrase string, rolls []string, err error) {
	if wordCount < 1 {
		return "", nil, fmt.Errorf("word count must be at least 1, got %d", wordCount)
	}
	if wordCount > MaxWords {
		return "", nil, fmt.Errorf("word count must be at most %d, got %d", MaxWords, wordCount)
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

	return strings.Join(words, separator), rolls, nil
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
//   - Romanian: 7,776 words, ~12.925 bits/word
//   - Mixed: 15,280 distinct words (English + Romanian, minus the 272 words
//     that are identical once capitalized), drawn uniformly, ~13.900
//     bits/word
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
// actually produce a word during generation (not the raw entry count).
func WordlistSizeByLanguage(lang Language) int {
	switch lang {
	case LanguageEnglish:
		return validWordCountEnglish
	case LanguageRomanian:
		return validWordCountRomanian
	case LanguageMixed:
		return len(mixedPool)
	default:
		return 0
	}
}

// Options selects what GenerateWithOptions generates. The zero Language is
// English and the zero Separator joins the words with nothing.
type Options struct {
	Words     int
	Language  Language
	Separator string
}

// Result is a generated passphrase and the dice rolls it came from. In Mixed
// mode each roll is prefixed with its list ("en:46122" or "ro:33544").
type Result struct {
	Passphrase string
	Rolls      []string
}

// GenerateWithOptions generates a passphrase as the Options describe. It is
// the single entry point that covers every Generate* variant; those remain
// for compatibility.
func GenerateWithOptions(o Options) (Result, error) {
	passphrase, rolls, err := GenerateWithRollsLanguageAndSeparator(o.Words, o.Language, o.Separator)
	if err != nil {
		return Result{}, err
	}
	return Result{Passphrase: passphrase, Rolls: rolls}, nil
}
