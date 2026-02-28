package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/cleonte/go-diceware"
)

const (
	defaultWords = 6
	minWords     = 1
	maxWords     = 20
)

func main() {
	var (
		words     int
		separator string
		showRolls bool
		showHelp  bool
		language  string
	)

	flag.IntVar(&words, "words", defaultWords, "number of words in the passphrase (1-20)")
	flag.IntVar(&words, "w", defaultWords, "number of words (shorthand)")
	flag.StringVar(&separator, "separator", "", "separator between words (default: none)")
	flag.StringVar(&separator, "s", "", "separator between words (shorthand)")
	flag.BoolVar(&showRolls, "rolls", false, "show dice rolls used to generate passphrase")
	flag.BoolVar(&showRolls, "r", false, "show dice rolls (shorthand)")
	flag.StringVar(&language, "lang", "en", "language: en (English), ro (Romanian), or mixed")
	flag.StringVar(&language, "l", "en", "language (shorthand)")
	flag.BoolVar(&showHelp, "help", false, "show help message")
	flag.BoolVar(&showHelp, "h", false, "show help message (shorthand)")

	flag.Usage = usage
	flag.Parse()

	if showHelp {
		usage()
		os.Exit(0)
	}

	// Validate word count
	if words < minWords || words > maxWords {
		fmt.Fprintf(os.Stderr, "Error: word count must be between %d and %d\n", minWords, maxWords)
		os.Exit(1)
	}

	// Parse language
	var lang diceware.Language
	switch language {
	case "en", "english":
		lang = diceware.LanguageEnglish
	case "ro", "romanian":
		lang = diceware.LanguageRomanian
	case "mixed", "mix":
		lang = diceware.LanguageMixed
	default:
		fmt.Fprintf(os.Stderr, "Error: unsupported language '%s'. Use: en, ro, or mixed\n", language)
		os.Exit(1)
	}

	// Generate passphrase
	if showRolls {
		passphrase, rolls, err := diceware.GenerateWithRollsAndLanguage(words, lang)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Apply separator if specified (split and rejoin to avoid regenerating)
		if separator != "" {
			wordList := splitCapitalizedWords(passphrase)
			passphrase = joinWithSeparator(wordList, separator)
		}

		fmt.Println("Dice rolls:", rolls)
		fmt.Println("Passphrase:", passphrase)
	} else {
		passphrase, err := diceware.GenerateWithLanguageAndSeparator(words, lang, separator)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(passphrase)
	}

	// Show entropy information
	entropy := diceware.Entropy(words)
	langName := "English"
	if lang == diceware.LanguageRomanian {
		langName = "Romanian"
	} else if lang == diceware.LanguageMixed {
		langName = "Mixed (English + Romanian)"
	}
	fmt.Fprintf(os.Stderr, "\nEntropy: %.1f bits (%d words, %s wordlist)\n",
		entropy, words, langName)
}

func usage() {
	fmt.Fprintf(os.Stderr, `Diceware Passphrase Generator

Generate cryptographically secure passphrases using the Diceware method
with the EFF large wordlist (7,776 English words) or Romanian wordlist (7,776 words).

Words are capitalized and concatenated by default (like "ColtDefaultArousal").

Usage:
  diceware [options]

Options:
  -w, --words N       Number of words in passphrase (default: %d, range: %d-%d)
  -s, --separator S   Separator between words (default: none)
  -l, --lang LANG     Language: en (English), ro (Romanian), or mixed (default: en)
  -r, --rolls         Show dice rolls used to generate passphrase
  -h, --help          Show this help message

Examples:
  # Generate a 6-word English passphrase (default, no separator)
  diceware
  Output: ColtDefaultArousalThimbleGaslightYearbook

  # Generate a Romanian passphrase
  diceware -l ro
  Output: AbaAbagerAbajurAbatajAbateAbator

  # Generate a mixed English and Romanian passphrase
  diceware -l mixed
  Output: ColtAbagerDefaultAbatajThimbleAbator

  # Generate an 8-word passphrase
  diceware -w 8

  # Generate with space separator (easier to read)
  diceware -s " "
  Output: Colt Default Arousal Thimble

  # Generate with dash separator
  diceware -s "-"
  Output: Colt-Default-Arousal-Thimble

  # Show dice rolls used
  diceware -r

  # Generate 10-word Romanian passphrase with underscores
  diceware -w 10 -l ro -s "_"

Recommended word counts for different security levels:
  4 words  - ~52 bits  - Minimum for low-value accounts
  6 words  - ~78 bits  - Recommended for most accounts
  8 words  - ~103 bits - High security accounts
  12 words - ~155 bits - Cryptocurrency wallets (minimum)

For more information about Diceware:
  https://theworld.com/~reinhold/diceware.html
  https://www.eff.org/deeplinks/2016/07/new-wordlists-random-passphrases
  https://github.com/danciu/diceware.ro (Romanian wordlist)

`, defaultWords, minWords, maxWords)
}

// splitCapitalizedWords splits a string of concatenated capitalized words
// e.g., "HelloWorldTest" -> ["Hello", "World", "Test"]
func splitCapitalizedWords(s string) []string {
	if s == "" {
		return nil
	}

	var words []string
	var currentWord strings.Builder

	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			// Start of a new word
			words = append(words, currentWord.String())
			currentWord.Reset()
		}
		currentWord.WriteRune(r)
	}

	// Add the last word
	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}

// joinWithSeparator joins words with the specified separator
func joinWithSeparator(words []string, separator string) string {
	return strings.Join(words, separator)
}
