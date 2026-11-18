package diceware_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/cleonte/go-diceware"
)

func ExampleGenerate() {
	passphrase, err := diceware.Generate(6)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(passphrase) > 0)
	// Output: true
}

func ExampleGenerateWithOptions() {
	res, err := diceware.GenerateWithOptions(diceware.Options{
		Words:     4,
		Language:  diceware.LanguageRomanian,
		Separator: "-",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(strings.Split(res.Passphrase, "-")), len(res.Rolls))
	// Output: 4 4
}

func ExampleParseLanguage() {
	lang, err := diceware.ParseLanguage("ro")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(lang)
	// Output: Romanian
}

func ExampleEntropy() {
	fmt.Printf("%.1f bits\n", diceware.Entropy(6))
	// Output: 77.5 bits
}

func ExampleEntropyForLanguage() {
	fmt.Printf("%.1f bits\n", diceware.EntropyForLanguage(6, diceware.LanguageMixed))
	// Output: 83.4 bits
}

func ExampleWordlistSizeByLanguage() {
	fmt.Println(diceware.WordlistSizeByLanguage(diceware.LanguageEnglish))
	fmt.Println(diceware.WordlistSizeByLanguage(diceware.LanguageMixed))
	// Output:
	// 7776
	// 15280
}
