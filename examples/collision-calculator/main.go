package main

import (
	"fmt"
	"math"
	"math/big"
)

func main() {
	// Parameters
	students := 70
	words := 4
	wordlistSize := 7776 // 6^5 for Diceware

	// Calculate total possible passphrases
	// For 4 words: 7776^4
	totalPassphrases := new(big.Int).Exp(
		big.NewInt(int64(wordlistSize)),
		big.NewInt(int64(words)),
		nil,
	)

	fmt.Println("=== Diceware Collision Probability Analysis ===")
	fmt.Println()
	fmt.Printf("Number of students: %d\n", students)
	fmt.Printf("Words per passphrase: %d\n", words)
	fmt.Printf("Wordlist size: %d\n", wordlistSize)
	fmt.Println()
	fmt.Printf("Total possible passphrases: %s\n", totalPassphrases.String())
	fmt.Printf("Total possible passphrases: %.2e\n", new(big.Float).SetInt(totalPassphrases))
	fmt.Println()

	// Calculate entropy
	entropy := float64(words) * math.Log2(float64(wordlistSize))
	fmt.Printf("Entropy: %.1f bits\n", entropy)
	fmt.Println()

	// Birthday paradox calculation
	// Probability of NO collision = (N/N) * ((N-1)/N) * ((N-2)/N) * ... * ((N-k+1)/N)
	// where N = total passphrases, k = number of students

	// Use logarithms to avoid overflow
	// P(no collision) = exp(sum(log((N-i)/N))) for i from 0 to k-1

	N := new(big.Float).SetInt(totalPassphrases)
	logProbNoCollision := 0.0

	for i := 0; i < students; i++ {
		// Calculate (N - i) / N
		numerator := new(big.Float).Sub(N, big.NewFloat(float64(i)))
		ratio, _ := new(big.Float).Quo(numerator, N).Float64()

		if ratio > 0 {
			logProbNoCollision += math.Log(ratio)
		}
	}

	probNoCollision := math.Exp(logProbNoCollision)
	probCollision := 1.0 - probNoCollision

	fmt.Println("=== Results ===")
	fmt.Printf("Probability of NO collision: %.10f (%.2e)\n", probNoCollision, probNoCollision)
	fmt.Printf("Probability of AT LEAST ONE collision: %.10f (%.2e)\n", probCollision, probCollision)
	fmt.Println()

	// Express as percentage
	fmt.Printf("Chance of collision: %.8f%%\n", probCollision*100)
	fmt.Println()

	// For comparison, calculate for different word counts
	fmt.Println("=== Comparison with different word counts ===")
	for w := 3; w <= 8; w++ {
		totalPass := new(big.Int).Exp(
			big.NewInt(int64(wordlistSize)),
			big.NewInt(int64(w)),
			nil,
		)

		NComp := new(big.Float).SetInt(totalPass)
		logProbNo := 0.0

		for i := 0; i < students; i++ {
			numerator := new(big.Float).Sub(NComp, big.NewFloat(float64(i)))
			ratio, _ := new(big.Float).Quo(numerator, NComp).Float64()

			if ratio > 0 {
				logProbNo += math.Log(ratio)
			}
		}

		probNo := math.Exp(logProbNo)
		probCol := 1.0 - probNo
		ent := float64(w) * math.Log2(float64(wordlistSize))

		fmt.Printf("%d words (%.1f bits): %.2e (%.8f%%)\n",
			w, ent, probCol, probCol*100)
	}

	fmt.Println()
	fmt.Println("=== Interpretation ===")
	if probCollision < 0.0001 {
		fmt.Printf("With 4-word passphrases, the collision risk is EXTREMELY LOW.\n")
		fmt.Printf("You would need approximately %.0f students before reaching 1%% collision probability.\n",
			math.Sqrt(float64(totalPassphrases.Int64()))*0.12)
	} else if probCollision < 0.01 {
		fmt.Printf("With 4-word passphrases, the collision risk is VERY LOW.\n")
	} else if probCollision < 0.5 {
		fmt.Printf("With 4-word passphrases, there is a LOW but measurable collision risk.\n")
	} else {
		fmt.Printf("With 4-word passphrases, collision is LIKELY.\n")
	}

	fmt.Println()
	fmt.Println("Recommendation: For 70 students, use at least 5-6 words to ensure negligible collision probability.")
}
