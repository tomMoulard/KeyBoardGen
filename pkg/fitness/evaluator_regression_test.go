package fitness

import (
	"strings"
	"testing"

	"github.com/tommoulard/keyboardgen/pkg/genetic"
	"github.com/tommoulard/keyboardgen/pkg/parser"
)

// TestFitnessEvaluatorNullCharacterRegression ensures the critical bug fix stays fixed
// This test specifically addresses the bug where null character layouts got positive fitness scores.
func TestFitnessEvaluatorNullCharacterRegression(t *testing.T) {
	// Parse test data
	sampleText := "the quick brown fox jumps over the lazy dog hello world"
	klparser := parser.NewKeyloggerParser()
	parseConfig := parser.DefaultConfig()

	keyloggerData, err := klparser.Parse(strings.NewReader(sampleText), parseConfig)
	if err != nil {
		t.Fatalf("Failed to parse data: %v", err)
	}

	// Create fitness evaluator
	geometry := StandardGeometry()
	weights := DefaultWeights()
	evaluator := NewFitnessEvaluator(geometry, weights)

	// Test Case 1: The original bug - all null characters should get 0.0 fitness
	nullLayout := make([]rune, 70) // All zero values = null characters
	charset := genetic.FullKeyboardCharset()

	nullFitness := evaluator.Evaluate(nullLayout, charset, keyloggerData)

	if nullFitness != 0.0 {
		t.Errorf("CRITICAL REGRESSION: Null layout got fitness %.10f, expected 0.0", nullFitness)
		t.Error("This was the exact bug that caused keyboard displays to show null characters!")
	}

	// Test Case 2: Partially null layout (another variant of the bug)
	validIndividual := genetic.NewRandomIndividual()
	corruptedLayout := make([]rune, len(validIndividual.Layout))
	copy(corruptedLayout, validIndividual.Layout)
	corruptedLayout[0] = 0 // Introduce null character
	corruptedLayout[5] = 0 // Introduce another null character

	corruptedFitness := evaluator.Evaluate(corruptedLayout, charset, keyloggerData)
	if corruptedFitness != 0.0 {
		t.Errorf("REGRESSION: Partially null layout got fitness %.10f, expected 0.0", corruptedFitness)
	}

	// Test Case 3: Invalid characters should also get 0.0
	invalidLayout := make([]rune, len(validIndividual.Layout))
	copy(invalidLayout, validIndividual.Layout)
	invalidLayout[0] = 'A' // Uppercase is invalid
	invalidLayout[1] = '1' // Number is invalid
	invalidLayout[2] = ' ' // Space is invalid

	invalidFitness := evaluator.Evaluate(invalidLayout, charset, keyloggerData)
	if invalidFitness != 0.0 {
		t.Errorf("REGRESSION: Invalid character layout got fitness %.10f, expected 0.0", invalidFitness)
	}

	// Test Case 4: Duplicate characters should get 0.0
	duplicateLayout := make([]rune, len(validIndividual.Layout))
	copy(duplicateLayout, validIndividual.Layout)
	duplicateLayout[1] = duplicateLayout[0] // Create duplicate

	duplicateFitness := evaluator.Evaluate(duplicateLayout, charset, keyloggerData)
	if duplicateFitness != 0.0 {
		t.Errorf("REGRESSION: Duplicate character layout got fitness %.10f, expected 0.0", duplicateFitness)
	}

	// Test Case 5: Missing characters should get 0.0
	incompleteLayout := make([]rune, len(validIndividual.Layout))
	copy(incompleteLayout, validIndividual.Layout)
	incompleteLayout[25] = incompleteLayout[24] // Remove 'z' by duplicating 'y'

	incompleteFitness := evaluator.Evaluate(incompleteLayout, charset, keyloggerData)
	if incompleteFitness != 0.0 {
		t.Errorf("REGRESSION: Incomplete layout got fitness %.10f, expected 0.0", incompleteFitness)
	}

	// Positive Test: Valid layout should get positive fitness
	validFitness := evaluator.Evaluate(validIndividual.Layout, validIndividual.Charset, keyloggerData)
	if validFitness <= 0.0 {
		t.Errorf("Valid layout should get positive fitness, got %.10f", validFitness)
	}

	t.Logf("All fitness evaluator regression tests passed")
	t.Logf("Valid layout fitness: %.6f", validFitness)
	t.Logf("Null layout fitness: %.6f (correctly rejected)", nullFitness)
}

// TestFitnessConsistency ensures fitness calculation is deterministic.
func TestFitnessConsistency(t *testing.T) {
	sampleText := "hello world test data"
	klparser := parser.NewKeyloggerParser()
	parseConfig := parser.DefaultConfig()

	keyloggerData, err := klparser.Parse(strings.NewReader(sampleText), parseConfig)
	if err != nil {
		t.Fatalf("Failed to parse data: %v", err)
	}

	geometry := StandardGeometry()
	weights := DefaultWeights()
	evaluator := NewFitnessEvaluator(geometry, weights)

	// Test same layout multiple times
	individual := genetic.NewRandomIndividual()

	var fitnessValues []float64

	for range 10 {
		fitness := evaluator.Evaluate(individual.Layout, individual.Charset, keyloggerData)
		fitnessValues = append(fitnessValues, fitness)
	}

	// All fitness values should be identical (within floating point precision)
	tolerance := 1e-10

	for i := 1; i < len(fitnessValues); i++ {
		diff := fitnessValues[i] - fitnessValues[0]
		if diff < 0 {
			diff = -diff
		}

		if diff > tolerance {
			t.Errorf("Fitness inconsistency: evaluation %d gave %.10f, expected %.10f",
				i, fitnessValues[i], fitnessValues[0])
		}
	}

	t.Logf("Fitness consistency test passed: %.6f", fitnessValues[0])
}

// TestFitnessDiversity ensures different layouts get different fitness scores.
func TestFitnessDiversity(t *testing.T) {
	sampleText := "the quick brown fox jumps over the lazy dog"
	klparser := parser.NewKeyloggerParser()
	parseConfig := parser.DefaultConfig()

	keyloggerData, err := klparser.Parse(strings.NewReader(sampleText), parseConfig)
	if err != nil {
		t.Fatalf("Failed to parse data: %v", err)
	}

	geometry := StandardGeometry()
	weights := DefaultWeights()
	evaluator := NewFitnessEvaluator(geometry, weights)

	// Generate multiple random layouts
	const numLayouts = 20

	fitnessScores := make([]float64, numLayouts)

	for i := range numLayouts {
		individual := genetic.NewRandomIndividual()
		fitnessScores[i] = evaluator.Evaluate(individual.Layout, individual.Charset, keyloggerData)
	}

	// Check that we have diversity in fitness scores
	uniqueScores := make(map[float64]bool)
	for _, score := range fitnessScores {
		uniqueScores[score] = true
	}

	if len(uniqueScores) < numLayouts/2 {
		t.Errorf("Low fitness diversity: only %d unique scores out of %d layouts",
			len(uniqueScores), numLayouts)
	}

	// Check that all scores are positive (no invalid layouts)
	for i, score := range fitnessScores {
		if score <= 0.0 {
			t.Errorf("Layout %d got non-positive fitness: %.6f", i, score)
		}
	}

	t.Logf("Fitness diversity test passed: %d unique scores from %d layouts",
		len(uniqueScores), numLayouts)
}

// TestQWERTYFitnessBaseline ensures QWERTY gets a reasonable fitness score.
func TestQWERTYFitnessBaseline(t *testing.T) {
	// Use English text for QWERTY baseline
	englishText := "the quick brown fox jumps over the lazy dog hello world programming test"
	klparser := parser.NewKeyloggerParser()
	parseConfig := parser.DefaultConfig()

	keyloggerData, err := klparser.Parse(strings.NewReader(englishText), parseConfig)
	if err != nil {
		t.Fatalf("Failed to parse data: %v", err)
	}

	geometry := StandardGeometry()
	weights := DefaultWeights()
	evaluator := NewFitnessEvaluator(geometry, weights)

	// QWERTY layout as defined in standard keyboard
	qwerty := [26]rune{
		'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', // Top row
		'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', // Middle row
		'z', 'x', 'c', 'v', 'b', 'n', 'm', // Bottom row
	}

	qwertyFitness := evaluator.EvaluateLegacy(qwerty, keyloggerData)

	// QWERTY should get a reasonable positive fitness
	if qwertyFitness <= 0.0 {
		t.Errorf("QWERTY should get positive fitness, got %.6f", qwertyFitness)
	}

	// QWERTY fitness should be in a reasonable range (0.3-0.7 typically)
	if qwertyFitness < 0.2 || qwertyFitness > 0.8 {
		t.Logf("QWERTY fitness %.6f seems unusual (expected 0.2-0.8)", qwertyFitness)
	}

	// Compare with a random layout
	randomIndividual := genetic.NewRandomIndividual()
	randomFitness := evaluator.Evaluate(randomIndividual.Layout, randomIndividual.Charset, keyloggerData)

	t.Logf("QWERTY fitness: %.6f", qwertyFitness)
	t.Logf("Random layout fitness: %.6f", randomFitness)

	if randomFitness > qwertyFitness {
		t.Logf("Random layout outperformed QWERTY by %.1f%%",
			(randomFitness/qwertyFitness-1.0)*100)
	} else {
		t.Logf("QWERTY outperformed random layout by %.1f%%",
			(qwertyFitness/randomFitness-1.0)*100)
	}
}
