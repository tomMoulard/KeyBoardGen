package fitness

import (
	"fmt"

	"github.com/tommoulard/keyboardgen/pkg/genetic"
)

// DemonstrateLayerPenalty shows how different keyboard layouts handle modifier keys.
func DemonstrateLayerPenalty() {
	fmt.Println("=== KEYBOARD LAYER PENALTY DEMONSTRATION ===")
	fmt.Println()

	// Create sample typing data with parentheses and numbers (common in programming)
	data := genetic.NewKeyloggerData()

	// Add parentheses usage
	for range 100 {
		data.AddChar('(')
		data.AddChar(')')
		data.AddBigram("()")
	}

	// Add numbers
	for range 50 {
		data.AddChar('1')
		data.AddChar('2')
		data.AddChar('3')
	}

	// Add letters
	for range 200 {
		data.AddChar('a')
		data.AddChar('b')
		data.AddChar('c')
	}

	// Create keyboard layouts
	qwertyLayout := StandardQWERTYLayout()
	azertyLayout := StandardAZERTYLayout()

	fmt.Printf("Sample data: %d parentheses, %d numbers, %d letters\n", 200, 150, 600)
	fmt.Println()

	// Analyze character access costs
	fmt.Println("CHARACTER ACCESS COMPARISON:")
	fmt.Println("Char | QWERTY Layer | Cost | AZERTY Layer | Cost | Advantage")
	fmt.Println("-----|--------------|------|--------------|------|----------")

	testChars := []rune{'(', ')', '1', '2', '3', 'a', '!'}
	for _, char := range testChars {
		_, qLayer, qCost := qwertyLayout.GetCharacterLayer(char)
		_, aLayer, aCost := azertyLayout.GetCharacterLayer(char)

		advantage := "Equal"
		if qCost > aCost {
			advantage = "AZERTY"
		} else if aCost > qCost {
			advantage = "QWERTY"
		}

		fmt.Printf("  %c  |      %d       | %.1f  |      %d       | %.1f  | %s\n",
			char, qLayer, qCost, aLayer, aCost, advantage)
	}

	fmt.Println()

	// Demonstrate bigram penalties
	fmt.Println("BIGRAM PENALTY COMPARISON:")
	fmt.Println("Bigram | QWERTY Penalty | AZERTY Penalty | Advantage")
	fmt.Println("-------|----------------|----------------|----------")

	testBigrams := []string{"()", "(1", "1)", "ab"}
	for _, bigram := range testBigrams {
		if len(bigram) == 2 {
			char1, char2 := rune(bigram[0]), rune(bigram[1])
			qPenalty := qwertyLayout.LayerPenalty(char1, char2)
			aPenalty := azertyLayout.LayerPenalty(char1, char2)

			advantage := "Equal"
			if qPenalty > aPenalty {
				advantage = "AZERTY"
			} else if aPenalty > qPenalty {
				advantage = "QWERTY"
			}

			fmt.Printf("  %s   |     %.2f       |     %.2f       | %s\n",
				bigram, qPenalty, aPenalty, advantage)
		}
	}

	fmt.Println()
	fmt.Println("KEY INSIGHTS:")
	fmt.Println("• In QWERTY: ( and ) require Shift key (higher cost)")
	fmt.Println("• In AZERTY: ( is on base layer, numbers require Shift")
	fmt.Println("• Programming with lots of parentheses benefits from AZERTY")
	fmt.Println("• Text with numbers benefits from QWERTY")
	fmt.Println("• The layer penalty system captures these trade-offs!")
}

// CompareLayoutEfficiency compares two layouts for specific use cases.
func CompareLayoutEfficiency(layout1, layout2 *KeyboardLayout, data genetic.KeyloggerDataInterface, useCase string) {
	fmt.Printf("\n=== %s EFFICIENCY COMPARISON ===\n", useCase)

	// Create evaluators
	weights := DefaultWeights()
	weights.LayerPenalty = 0.25 // Emphasize layer penalty for this demo

	eval1 := NewLayerAwareFitnessEvaluator(layout1, weights)
	eval2 := NewLayerAwareFitnessEvaluator(layout2, weights)

	// Create dummy individual for testing
	individual := genetic.Individual{
		Layout:  []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'},
		Charset: genetic.FullKeyboardCharset(),
	}

	fitness1 := eval1.EvaluateWithLayers(individual, data)
	fitness2 := eval2.EvaluateWithLayers(individual, data)

	penalty1 := eval1.calculateLayerPenalty(data)
	penalty2 := eval2.calculateLayerPenalty(data)

	fmt.Printf("%s fitness: %.6f (penalty: %.4f)\n", layout1.Name, fitness1, penalty1)
	fmt.Printf("%s fitness: %.6f (penalty: %.4f)\n", layout2.Name, fitness2, penalty2)

	if fitness1 > fitness2 {
		improvement := ((fitness1 - fitness2) / fitness2) * 100
		fmt.Printf("Winner: %s (%.1f%% better)\n", layout1.Name, improvement)
	} else if fitness2 > fitness1 {
		improvement := ((fitness2 - fitness1) / fitness1) * 100
		fmt.Printf("Winner: %s (%.1f%% better)\n", layout2.Name, improvement)
	} else {
		fmt.Println("Result: Equal performance")
	}
}
