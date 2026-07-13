package main

import (
	"fmt"
	"os"

	"github.com/cleonte/go-diceware"
	"github.com/spf13/cobra"
)

const (
	defaultWords = 6
	minWords     = 1
	maxWords     = 20
)

var (
	words     int
	separator string
	showRolls bool
	language  string
)

var rootCmd = &cobra.Command{
	Use:   "diceware",
	Short: "Diceware Passphrase Generator",
	Long: `Generate cryptographically secure passphrases using the Diceware method
with the EFF large wordlist (7,776 English words) or Romanian wordlist (7,776 words).

Words are capitalized and concatenated by default (like "ColtDefaultArousal").`,
	Example: `  # Generate a 6-word English passphrase (default, no separator)
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
  diceware -w 10 -l ro -s "_"`,
	RunE:          run,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.Flags().IntVarP(&words, "words", "w", defaultWords,
		fmt.Sprintf("number of words in the passphrase (%d-%d)", minWords, maxWords))
	rootCmd.Flags().StringVarP(&separator, "separator", "s", "", "separator between words (default: none)")
	rootCmd.Flags().BoolVarP(&showRolls, "rolls", "r", false, "show dice rolls used to generate passphrase")
	rootCmd.Flags().StringVarP(&language, "lang", "l", "en", "language: en (English), ro (Romanian), or mixed")

	rootCmd.SetHelpTemplate(rootCmd.HelpTemplate() + fmt.Sprintf(`
Recommended word counts for different security levels:
  4 words  - ~52 bits  - Minimum for low-value accounts
  6 words  - ~78 bits  - Recommended for most accounts
  8 words  - ~103 bits - High security accounts
  12 words - ~155 bits - Cryptocurrency wallets (minimum)

For more information about Diceware:
  https://theworld.com/~reinhold/diceware.html
  https://www.eff.org/deeplinks/2016/07/new-wordlists-random-passphrases
  https://github.com/danciu/diceware.ro (Romanian wordlist)
`))
}

func run(cmd *cobra.Command, args []string) error {
	// Validate word count
	if words < minWords || words > maxWords {
		return fmt.Errorf("word count must be between %d and %d", minWords, maxWords)
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
		return fmt.Errorf("unsupported language '%s'. Use: en, ro, or mixed", language)
	}

	// Generate passphrase
	if showRolls {
		passphrase, rolls, err := diceware.GenerateWithRollsLanguageAndSeparator(words, lang, separator)
		if err != nil {
			return err
		}

		fmt.Println("Dice rolls:", rolls)
		fmt.Println("Passphrase:", passphrase)
	} else {
		passphrase, err := diceware.GenerateWithLanguageAndSeparator(words, lang, separator)
		if err != nil {
			return err
		}

		fmt.Println(passphrase)
	}

	// Show entropy information
	entropy := diceware.EntropyForLanguage(words, lang)
	langName := "English"
	if lang == diceware.LanguageRomanian {
		langName = "Romanian"
	} else if lang == diceware.LanguageMixed {
		langName = "Mixed (English + Romanian)"
	}
	fmt.Fprintf(os.Stderr, "\nEntropy: %.1f bits (%d words, %s wordlist)\n",
		entropy, words, langName)

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
