// Command diceware generates Diceware passphrases from the command line.
package main

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/cleonte/go-diceware"
	"github.com/spf13/cobra"
)

const (
	defaultWords = 6
	minWords     = 1
	maxWords     = 24
)

// version is set at release time with -ldflags "-X main.version=...";
// "go install" builds fall back to the module version.
var version = "dev"

func buildVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return strings.TrimPrefix(info.Main.Version, "v")
	}
	return version
}

// newRootCmd builds the command with its own flag state, so tests can run it
// repeatedly with their own arguments and output streams.
func newRootCmd(stdout, stderr io.Writer) *cobra.Command {
	var (
		words     int
		separator string
		showRolls bool
		language  string
	)

	cmd := &cobra.Command{
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
  Output: AbajurAbandonAbanosAbatAbatajAbate

  # Generate a mixed English and Romanian passphrase
  diceware -l mixed
  Output: ColtAbandonDefaultAbanosThimbleAbat

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
		Args:          cobra.NoArgs,
		Version:       buildVersion(),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(stdout, stderr, words, separator, showRolls, language)
		},
	}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	cmd.Flags().IntVarP(&words, "words", "w", defaultWords,
		fmt.Sprintf("number of words in the passphrase (%d-%d)", minWords, maxWords))
	cmd.Flags().StringVarP(&separator, "separator", "s", "", "separator between words (default: none)")
	cmd.Flags().BoolVarP(&showRolls, "rolls", "r", false, "show dice rolls used to generate passphrase")
	cmd.Flags().StringVarP(&language, "lang", "l", "en", "language: en (English), ro (Romanian), or mixed")
	_ = cmd.RegisterFlagCompletionFunc("lang", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return []string{"en", "ro", "mixed"}, cobra.ShellCompDirectiveNoFileComp
	})

	cmd.AddCommand(newCompletionCmd())

	cmd.SetHelpTemplate(cmd.HelpTemplate() + `
Recommended word counts for different security levels:
  4 words  - ~52 bits  - Minimum for low-value accounts
  6 words  - ~78 bits  - Recommended for most accounts
  8 words  - ~103 bits - High security accounts
  12 words - ~155 bits - Cryptocurrency wallets (minimum)

For more information about Diceware:
  https://theworld.com/~reinhold/diceware.html
  https://www.eff.org/deeplinks/2016/07/new-wordlists-random-passphrases
  https://github.com/cleonte/go-diceware/blob/main/NOTICE (wordlist sources)
`)
	return cmd
}

// newCompletionCmd adds "diceware completion SHELL". Cobra only adds its own
// completion command to a root command that has subcommands.
func newCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "completion bash|zsh|fish|powershell",
		Short:     "Print a shell completion script",
		Long:      "Print a completion script for the given shell, for example:\n\n  source <(diceware completion bash)",
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		RunE: func(cmd *cobra.Command, args []string) error {
			root, out := cmd.Root(), cmd.OutOrStdout()
			switch args[0] {
			case "bash":
				return root.GenBashCompletionV2(out, true)
			case "zsh":
				return root.GenZshCompletion(out)
			case "fish":
				return root.GenFishCompletion(out, true)
			default:
				return root.GenPowerShellCompletionWithDesc(out)
			}
		},
	}
}

func run(stdout, stderr io.Writer, words int, separator string, showRolls bool, language string) error {
	// Validate word count
	if words < minWords || words > maxWords {
		return fmt.Errorf("word count must be between %d and %d", minWords, maxWords)
	}

	lang, err := diceware.ParseLanguage(language)
	if err != nil {
		return err
	}

	// Generate passphrase
	if showRolls {
		passphrase, rolls, err := diceware.GenerateWithRollsLanguageAndSeparator(words, lang, separator)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintln(stdout, "Dice rolls:", rolls)
		_, _ = fmt.Fprintln(stdout, "Passphrase:", passphrase)
	} else {
		passphrase, err := diceware.GenerateWithLanguageAndSeparator(words, lang, separator)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintln(stdout, passphrase)
	}

	// Show entropy information
	entropy := diceware.EntropyForLanguage(words, lang)
	_, _ = fmt.Fprintf(stderr, "\nEntropy: %.1f bits (%d words, %s wordlist)\n",
		entropy, words, lang)

	return nil
}

func main() {
	if err := newRootCmd(os.Stdout, os.Stderr).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
