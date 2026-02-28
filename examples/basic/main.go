package main

import (
	"fmt"
	"log"

	"github.com/cleonte/go-diceware"
)

func main() {
	fmt.Println("=== Diceware Library Examples ===")
	fmt.Println()

	// Example 1: Basic usage (capitalized, no separator)
	fmt.Println("1. Generate a passphrase with 6 words (CamelCase):")
	passphrase, err := diceware.Generate(6)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   %s\n\n", passphrase)

	// Example 2: With space separator for easier reading
	fmt.Println("2. Generate with space separator:")
	passphrase, err = diceware.GenerateWithSeparator(4, " ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   %s\n\n", passphrase)

	// Example 3: With dash separator
	fmt.Println("3. Generate with dash separator:")
	passphrase, err = diceware.GenerateWithSeparator(4, "-")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   %s\n\n", passphrase)

	// Example 4: Show dice rolls
	fmt.Println("4. Generate with dice rolls:")
	passphrase, rolls, err := diceware.GenerateWithRolls(5)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Passphrase: %s\n", passphrase)
	fmt.Printf("   Dice rolls: %v\n\n", rolls)

	// Example 5: Calculate entropy for different word counts
	fmt.Println("5. Entropy for different word counts:")
	for _, words := range []int{4, 6, 8, 12} {
		entropy := diceware.Entropy(words)
		fmt.Printf("   %2d words: %.1f bits\n", words, entropy)
	}
	fmt.Println()

	// Example 6: Wordlist information
	fmt.Println("6. Wordlist information:")
	fmt.Printf("   Total words in list: %d\n", diceware.WordlistSize())
	fmt.Printf("   Entropy per word: %.3f bits\n\n", diceware.Entropy(1))

	// Example 7: Generate multiple passphrases
	fmt.Println("7. Generate 3 passphrases for comparison:")
	for i := 1; i <= 3; i++ {
		passphrase, err = diceware.Generate(5)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("   %d. %s\n", i, passphrase)
	}
	fmt.Println()

	// Example 8: Different separators (all capitalized words)
	fmt.Println("8. Capitalized words with different separators:")
	separators := []struct {
		sep  string
		name string
	}{
		{"", "none"},
		{" ", "space"},
		{"-", "dash"},
		{"_", "underscore"},
		{".", "dot"},
	}
	for _, s := range separators {
		passphrase, err = diceware.GenerateWithSeparator(3, s.sep)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("   %-10s: %s\n", s.name, passphrase)
	}
}
