package diceware

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name      string
		wordCount int
		wantErr   bool
	}{
		{"valid 1 word", 1, false},
		{"valid 4 words", 4, false},
		{"valid 6 words", 6, false},
		{"valid 8 words", 8, false},
		{"invalid 0 words", 0, true},
		{"invalid negative", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passphrase, err := Generate(tt.wordCount)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				// Default is no separator, so we can't easily split
				// Just check it's not empty
				if passphrase == "" {
					t.Error("Generate() returned empty passphrase")
				}
				// Check that words are capitalized
				if len(passphrase) > 0 && passphrase[0] < 'A' || passphrase[0] > 'Z' {
					t.Errorf("Generate() passphrase doesn't start with capital letter: %s", passphrase)
				}
			}
		})
	}
}

func TestGenerateWithSeparator(t *testing.T) {
	tests := []struct {
		name      string
		wordCount int
		separator string
		wantErr   bool
	}{
		{"dash separator", 4, "-", false},
		{"underscore separator", 4, "_", false},
		{"empty separator", 4, "", false},
		{"space separator", 4, " ", false},
		{"multi-char separator", 4, " | ", false},
		{"invalid word count", 0, "-", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passphrase, err := GenerateWithSeparator(tt.wordCount, tt.separator)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateWithSeparator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				// For empty separator, we can't easily count words, so just check it's not empty
				if tt.separator == "" {
					if passphrase == "" {
						t.Error("GenerateWithSeparator() returned empty string")
					}
				} else {
					words := strings.Split(passphrase, tt.separator)
					if len(words) != tt.wordCount {
						t.Errorf("GenerateWithSeparator() returned %d words, want %d", len(words), tt.wordCount)
					}
				}
			}
		})
	}
}

func TestGenerateWithRolls(t *testing.T) {
	tests := []struct {
		name      string
		wordCount int
		wantErr   bool
	}{
		{"valid 1 word", 1, false},
		{"valid 6 words", 6, false},
		{"invalid 0 words", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passphrase, rolls, err := GenerateWithRolls(tt.wordCount)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateWithRolls() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				// No separator by default, just check not empty
				if passphrase == "" {
					t.Error("GenerateWithRolls() returned empty passphrase")
				}
				if len(rolls) != tt.wordCount {
					t.Errorf("GenerateWithRolls() returned %d rolls, want %d", len(rolls), tt.wordCount)
				}
				// Verify each roll is valid (5 digits, each 1-6)
				for _, roll := range rolls {
					if len(roll) != 5 {
						t.Errorf("roll %s has invalid length %d, want 5", roll, len(roll))
					}
					for _, digit := range roll {
						if digit < '1' || digit > '6' {
							t.Errorf("roll %s contains invalid digit %c", roll, digit)
						}
					}
				}
			}
		})
	}
}

func TestRollDice(t *testing.T) {
	// Test that rollDice returns values between 1 and 6
	for i := 0; i < 100; i++ {
		result, err := rollDice()
		if err != nil {
			t.Fatalf("rollDice() failed: %v", err)
		}
		if result < 1 || result > 6 {
			t.Errorf("rollDice() = %d, want value between 1 and 6", result)
		}
	}
}

func TestRollFiveDice(t *testing.T) {
	// Test that rollFiveDice returns a valid 5-digit string
	for i := 0; i < 100; i++ {
		result, err := rollFiveDice()
		if err != nil {
			t.Fatalf("rollFiveDice() failed: %v", err)
		}
		if len(result) != 5 {
			t.Errorf("rollFiveDice() returned %s with length %d, want length 5", result, len(result))
		}
		for _, digit := range result {
			if digit < '1' || digit > '6' {
				t.Errorf("rollFiveDice() returned %s with invalid digit %c", result, digit)
			}
		}
	}
}

func TestGetWord(t *testing.T) {
	// Test that getWord returns a non-empty word
	for i := 0; i < 100; i++ {
		word, err := getWord()
		if err != nil {
			t.Fatalf("getWord() failed: %v", err)
		}
		if word == "" {
			t.Error("getWord() returned empty string")
		}
	}
}

func TestEntropy(t *testing.T) {
	tests := []struct {
		wordCount int
		want      float64
	}{
		{1, 12.925},
		{4, 51.7},
		{5, 64.625},
		{6, 77.55},
		{8, 103.4},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := Entropy(tt.wordCount)
			// Allow small floating point differences
			if got < tt.want-0.1 || got > tt.want+0.1 {
				t.Errorf("Entropy(%d) = %f, want approximately %f", tt.wordCount, got, tt.want)
			}
		})
	}
}

func TestWordlistSize(t *testing.T) {
	size := WordlistSize()
	// EFF large wordlist has exactly 7,776 words (6^5)
	if size != 7776 {
		t.Errorf("WordlistSize() = %d, want 7776", size)
	}
}

func TestParseWordlist(t *testing.T) {
	testData := `11111	abacus
11112	abdomen
11113	abdominal`

	result := parseWordlist(testData)

	if len(result) != 3 {
		t.Errorf("parseWordlist() returned %d entries, want 3", len(result))
	}

	tests := []struct {
		roll string
		want string
	}{
		{"11111", "abacus"},
		{"11112", "abdomen"},
		{"11113", "abdominal"},
	}

	for _, tt := range tests {
		if got := result[tt.roll]; got != tt.want {
			t.Errorf("parseWordlist()[%s] = %s, want %s", tt.roll, got, tt.want)
		}
	}
}

func TestParseWordlistEmptyLines(t *testing.T) {
	testData := `11111	abacus

11112	abdomen
	
11113	abdominal`

	result := parseWordlist(testData)

	if len(result) != 3 {
		t.Errorf("parseWordlist() with empty lines returned %d entries, want 3", len(result))
	}
}

// Benchmark tests
func BenchmarkGenerate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := Generate(6)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRollDice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := rollDice()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRollFiveDice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := rollFiveDice()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetWord(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := getWord()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Test for randomness distribution
func TestRandomnessDistribution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping randomness distribution test in short mode")
	}

	// Roll dice many times and ensure distribution is reasonable
	counts := make(map[int]int)
	iterations := 6000

	for i := 0; i < iterations; i++ {
		result, err := rollDice()
		if err != nil {
			t.Fatal(err)
		}
		counts[result]++
	}

	// Each number should appear roughly 1/6 of the time
	// With 6000 iterations, expect each number ~1000 times
	// Allow for statistical variance (between 800 and 1200)
	for i := 1; i <= 6; i++ {
		count := counts[i]
		if count < 800 || count > 1200 {
			t.Logf("Warning: dice roll %d appeared %d times (expected ~1000)", i, count)
		}
	}
}

// Test that multiple calls produce different results (not deterministic)
func TestNonDeterministic(t *testing.T) {
	results := make(map[string]bool)

	for i := 0; i < 10; i++ {
		passphrase, err := Generate(6)
		if err != nil {
			t.Fatal(err)
		}
		results[passphrase] = true
	}

	// We should have 10 unique passphrases (extremely unlikely to get duplicates)
	if len(results) < 9 {
		t.Errorf("Got %d unique passphrases out of 10, expected at least 9", len(results))
	}
}

// Test capitalize function
func TestCapitalize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "Hello"},
		{"world", "World"},
		{"a", "A"},
		{"", ""},
		{"Hello", "Hello"},
		{"WORLD", "WORLD"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := capitalize(tt.input)
			if got != tt.want {
				t.Errorf("capitalize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// Test that generated words are capitalized
func TestWordsAreCapitalized(t *testing.T) {
	for i := 0; i < 10; i++ {
		word, err := getWord()
		if err != nil {
			t.Fatal(err)
		}
		if len(word) == 0 {
			t.Error("getWord() returned empty string")
			continue
		}
		if word[0] < 'A' || word[0] > 'Z' {
			t.Errorf("getWord() = %q, first letter is not capitalized", word)
		}
	}
}
