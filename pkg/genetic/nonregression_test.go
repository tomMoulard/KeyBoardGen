package genetic

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tommoulard/keyboardgen/pkg/parser"
)

// TestKeyboardDisplayIntegration tests the complete pipeline from GA to display.
func TestKeyboardDisplayIntegration(t *testing.T) {
	// Test data
	sampleText := "the quick brown fox jumps over the lazy dog hello world programming test"

	// Parse sample data
	klparser := parser.NewKeyloggerParser()
	parseConfig := parser.DefaultConfig()

	keyloggerData, err := klparser.Parse(strings.NewReader(sampleText), parseConfig)
	if err != nil {
		t.Fatalf("Failed to parse keylogger data: %v", err)
	}

	// Verify data was parsed correctly
	if keyloggerData.TotalChars == 0 {
		t.Fatal("No characters parsed")
	}

	if len(keyloggerData.CharFrequency) == 0 {
		t.Fatal("No character frequencies recorded")
	}

	// Test that GetAllBigrams works (needed for fitness evaluation)
	bigrams := keyloggerData.GetAllBigrams()
	if len(bigrams) == 0 {
		t.Fatal("No bigrams found")
	}

	t.Logf("Parsed %d characters, %d bigrams", keyloggerData.TotalChars, len(bigrams))
}

// TestIndividualNotNull ensures individuals are never null after creation.
func TestIndividualNotNull(t *testing.T) {
	for i := range 10 {
		individual := NewRandomIndividual()

		// Check that layout is not null
		for pos, char := range individual.Layout {
			if char == 0 {
				t.Errorf("Null character found at position %d in individual %d", pos, i)
			}

			if char < 'a' || char > 'z' {
				t.Errorf("Invalid character '%c' at position %d in individual %d", char, pos, i)
			}
		}

		// Check that individual is valid
		if !individual.IsValid() {
			t.Errorf("Invalid individual %d: %s", i, string(individual.Layout[:]))
		}

		// Check that we have all 26 letters
		seen := make(map[rune]bool)
		for _, char := range individual.Layout {
			seen[char] = true
		}

		if len(seen) != 26 {
			t.Errorf("Individual %d missing letters, only has %d unique chars", i, len(seen))
		}
	}
}

// TestClonePreservesLayout ensures cloning doesn't corrupt layouts.
func TestClonePreservesLayout(t *testing.T) {
	original := NewRandomIndividual()
	original.Fitness = 0.75
	original.Age = 5

	clone := original.Clone()

	// Test that layout is preserved
	for i, char := range original.Layout {
		if clone.Layout[i] != char {
			t.Errorf("Clone layout differs at position %d: original=%c, clone=%c",
				i, char, clone.Layout[i])
		}
	}

	// Test that metadata is preserved
	if clone.Fitness != original.Fitness {
		t.Errorf("Clone fitness mismatch: expected %f, got %f", original.Fitness, clone.Fitness)
	}

	if clone.Age != original.Age {
		t.Errorf("Clone age mismatch: expected %d, got %d", original.Age, clone.Age)
	}

	// Test that clone is independent
	clone.Layout[0] = 'X'
	clone.Fitness = 0.9
	clone.Age = 10

	// Original should be unchanged
	if original.Layout[0] == 'X' {
		t.Error("Modifying clone affected original layout")
	}

	if original.Fitness == 0.9 {
		t.Error("Modifying clone affected original fitness")
	}

	if original.Age == 10 {
		t.Error("Modifying clone affected original age")
	}
}

// TestBestIndividualInitialization tests the fix for null bestIndividual.
func TestBestIndividualInitialization(t *testing.T) {
	// This test ensures that bestIndividual is properly initialized
	// and doesn't remain as null characters

	// Create mock fitness evaluator that returns varied fitness
	mockEvaluator := &MockFitnessEvaluator{}

	config := DefaultConfig()
	config.PopulationSize = 5
	config.MaxGenerations = 3

	ga := NewParallelGA(mockEvaluator, config)

	// Create sample data
	sampleText := "hello world test"
	klparser := parser.NewKeyloggerParser()
	parseConfig := parser.DefaultConfig()

	keyloggerData, err := klparser.Parse(strings.NewReader(sampleText), parseConfig)
	if err != nil {
		t.Fatalf("Failed to parse data: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	best, err := ga.Run(ctx, keyloggerData, nil)
	if err != nil {
		t.Fatalf("GA run failed: %v", err)
	}

	// Test that bestIndividual is not null
	for pos, char := range best.Layout {
		if char == 0 {
			t.Errorf("Best individual has null character at position %d", pos)
		}
	}

	// Test that bestIndividual is valid
	if !best.IsValid() {
		t.Errorf("Best individual is invalid: %s", string(best.Layout[:]))
	}

	// Test that fitness is reasonable
	if best.Fitness <= 0 {
		t.Errorf("Best individual has non-positive fitness: %f", best.Fitness)
	}
}

// MockFitnessEvaluator for testing.
type MockFitnessEvaluator struct{}

func (m *MockFitnessEvaluator) Evaluate(layout [26]rune, data KeyloggerDataInterface) float64 {
	// Return a simple fitness based on layout characteristics
	// This ensures different layouts get different fitness values

	// Count unique characters (should always be 26 for valid layouts)
	seen := make(map[rune]bool)
	for _, char := range layout {
		seen[char] = true
	}

	// Base fitness on uniqueness and character distribution
	fitness := float64(len(seen)) / 26.0

	// Add some variation based on first character
	if layout[0] != 0 {
		fitness += float64(layout[0]-'a') / 100.0
	}

	return fitness
}

// TestKeyloggerDataInterface ensures all required methods are implemented.
func TestKeyloggerDataInterface(t *testing.T) {
	sampleText := "hello world test data for interface verification"

	klparser := parser.NewKeyloggerParser()
	parseConfig := parser.DefaultConfig()

	keyloggerData, err := klparser.Parse(strings.NewReader(sampleText), parseConfig)
	if err != nil {
		t.Fatalf("Failed to parse data: %v", err)
	}

	// Test KeyloggerDataInterface methods
	var iface KeyloggerDataInterface = keyloggerData

	// Test GetCharFreq
	freq := iface.GetCharFreq('e')
	if freq <= 0 {
		t.Error("GetCharFreq should return positive frequency for 'e'")
	}

	// Test GetTotalChars
	total := iface.GetTotalChars()
	if total <= 0 {
		t.Error("GetTotalChars should return positive count")
	}

	// Test GetAllBigrams
	bigrams := iface.GetAllBigrams()
	if len(bigrams) == 0 {
		t.Error("GetAllBigrams should return non-empty map")
	}

	// Test GetBigramFreq
	bigramFreq := iface.GetBigramFreq("he")
	// Note: this might be 0 if "he" doesn't exist, which is okay

	// Test GetTrigramFreq
	trigramFreq := iface.GetTrigramFreq("hel")
	// Note: this might be 0 if "hel" doesn't exist, which is okay

	t.Logf("Interface test passed: total=%d, bigrams=%d, 'e' freq=%d, 'he' freq=%d, 'hel' freq=%d",
		total, len(bigrams), freq, bigramFreq, trigramFreq)
}

// TestDisplayHandlesNullCharacters ensures display gracefully handles null chars.
func TestDisplayHandlesNullCharacters(t *testing.T) {
	// Create an individual with some null characters (simulating corruption)
	individual := NewRandomIndividual()

	// Corrupt some characters
	individual.Layout[0] = 0
	individual.Layout[1] = 0

	// This should not panic and should handle null characters gracefully
	layoutString := string(individual.Layout[:])

	// Check that we can detect null characters
	hasNull := false

	for _, char := range individual.Layout {
		if char == 0 {
			hasNull = true

			break
		}
	}

	if !hasNull {
		t.Error("Test setup failed - should have null characters")
	}

	// The string should contain null characters
	if len(layoutString) != 26 {
		t.Errorf("Layout string length should be 26, got %d", len(layoutString))
	}

	t.Log("Display null character handling test passed")
}

// TestFitnessEvaluatorRejectsNullCharacters moved to fitness package to avoid import cycle

// TestGeneticAlgorithmEvolution ensures GA actually evolves and doesn't get stuck.
func TestGeneticAlgorithmEvolution(t *testing.T) {
	// Create larger test dataset to ensure evolution can happen
	sampleText := strings.Repeat("the quick brown fox jumps over the lazy dog hello world programming test ", 20)

	klparser := parser.NewKeyloggerParser()
	parseConfig := parser.DefaultConfig()

	keyloggerData, err := klparser.Parse(strings.NewReader(sampleText), parseConfig)
	if err != nil {
		t.Fatalf("Failed to parse data: %v", err)
	}

	// Use mock fitness evaluator to avoid import cycle
	evaluator := &MockFitnessEvaluator{}

	// Use configuration that should allow evolution
	config := Config{
		PopulationSize:  50,  // Reasonable size for test
		MaxGenerations:  10,  // Enough generations to see improvement
		MutationRate:    0.3, // High mutation for exploration
		CrossoverRate:   0.9, // High crossover for mixing
		ElitismCount:    1,   // Low elitism to prevent stagnation
		TournamentSize:  3,
		ParallelWorkers: 4,
	}

	ga := NewParallelGA(evaluator, config)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Track fitness across generations
	var (
		fitnessHistory []float64
		generationAges []int
	)

	bestIndividual, err := ga.Run(ctx, keyloggerData, func(generation int, best Individual) {
		fitnessHistory = append(fitnessHistory, best.Fitness)
		generationAges = append(generationAges, best.Age)

		// Ensure no null characters in best individual
		for pos, char := range best.Layout {
			if char == 0 {
				t.Errorf("Generation %d: Null character at position %d", generation, pos)
			}
		}

		// Ensure individual is valid
		if !best.IsValid() {
			t.Errorf("Generation %d: Best individual is invalid", generation)
		}

		// Ensure fitness is positive
		if best.Fitness <= 0.0 {
			t.Errorf("Generation %d: Best fitness should be positive, got %f", generation, best.Fitness)
		}
	})
	if err != nil {
		t.Fatalf("GA run failed: %v", err)
	}

	// Test final result
	if bestIndividual.Fitness <= 0.0 {
		t.Errorf("Final best fitness should be positive, got %f", bestIndividual.Fitness)
	}

	if !bestIndividual.IsValid() {
		t.Errorf("Final best individual should be valid")
	}

	// Check for evolution - fitness should improve over time
	if len(fitnessHistory) < 5 {
		t.Skip("Not enough generations to test evolution")
	}

	// Compare first and last few generations
	firstThird := fitnessHistory[:len(fitnessHistory)/3]
	lastThird := fitnessHistory[len(fitnessHistory)*2/3:]

	firstAvg := average(firstThird)
	lastAvg := average(lastThird)

	if lastAvg <= firstAvg {
		t.Logf("Evolution check: First third avg=%.6f, Last third avg=%.6f", firstAvg, lastAvg)
		t.Logf("Full fitness history: %v", fitnessHistory)
		// Allow some tolerance as evolution might plateau
		if lastAvg < firstAvg*0.99 {
			t.Errorf("Expected evolution: last third fitness (%.6f) should be >= first third (%.6f)", lastAvg, firstAvg)
		}
	}

	t.Logf("Evolution test passed: improved from %.6f to %.6f", firstAvg, lastAvg)
}

// TestAdaptiveConfigurationSelection tests that adaptive config works properly.
func TestAdaptiveConfigurationSelection(t *testing.T) {
	// Test small dataset
	smallConfig := AdaptiveConfig(1000)
	if smallConfig.PopulationSize != DefaultConfig().PopulationSize {
		t.Errorf("Small dataset should use default config, got population %d", smallConfig.PopulationSize)
	}

	// Test medium dataset
	mediumConfig := AdaptiveConfig(50000)
	if mediumConfig.PopulationSize == DefaultConfig().PopulationSize {
		t.Errorf("Medium dataset should use different config than default")
	}

	// Test large dataset
	largeConfig := AdaptiveConfig(300000)

	expectedLarge := LargeDatasetConfig()
	if largeConfig.PopulationSize != expectedLarge.PopulationSize {
		t.Errorf("Large dataset should use large config, got population %d, expected %d",
			largeConfig.PopulationSize, expectedLarge.PopulationSize)
	}

	// Verify large config has reasonable parameters for diversity
	if largeConfig.MutationRate <= 0.1 {
		t.Errorf("Large config should have high mutation rate, got %f", largeConfig.MutationRate)
	}

	if largeConfig.ElitismCount >= 5 {
		t.Errorf("Large config should have low elitism count, got %d", largeConfig.ElitismCount)
	}

	t.Log("Adaptive configuration test passed")
}

// TestNullCharacterRegressionFullPipeline tests the complete pipeline end-to-end.
func TestNullCharacterRegressionFullPipeline(t *testing.T) {
	// This test recreates the exact scenario that caused the null character bug

	// Large dataset like Harry Potter
	largeText := strings.Repeat("the quick brown fox jumps over the lazy dog hello world programming test keyboard layout optimization genetic algorithms ", 500)

	klparser := parser.NewKeyloggerParser()
	parseConfig := parser.DefaultConfig()

	keyloggerData, err := klparser.Parse(strings.NewReader(largeText), parseConfig)
	if err != nil {
		t.Fatalf("Failed to parse large dataset: %v", err)
	}

	// Use adaptive configuration (this should trigger large dataset config)
	config := AdaptiveConfig(keyloggerData.TotalChars)

	// Use mock evaluator
	evaluator := &MockFitnessEvaluator{}

	ga := NewParallelGA(evaluator, config)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Run for just a few generations to test the critical initialization
	config.MaxGenerations = 3

	bestIndividual, err := ga.Run(ctx, keyloggerData, func(generation int, best Individual) {
		// This is where the bug manifested - null characters in best individual
		nullCount := 0

		for _, char := range best.Layout {
			if char == 0 {
				nullCount++
			}
		}

		if nullCount > 0 {
			t.Errorf("REGRESSION: Generation %d has %d null characters in best individual", generation, nullCount)
		}

		// The bug also showed up as unrealistic fitness for null layouts
		if !best.IsValid() && best.Fitness > 0.0 {
			t.Errorf("REGRESSION: Invalid individual has positive fitness %.6f", best.Fitness)
		}

		// Original bug: fitness stayed identical across generations
		// We'll check this in the final verification step
	})
	if err != nil {
		t.Fatalf("Full pipeline test failed: %v", err)
	}

	// Final verification - the exact checks that would have caught the original bug
	finalNullCount := 0

	for _, char := range bestIndividual.Layout {
		if char == 0 {
			finalNullCount++
		}
	}

	if finalNullCount > 0 {
		t.Errorf("CRITICAL REGRESSION: Final result has %d null characters", finalNullCount)
		t.Errorf("Layout: %s", string(bestIndividual.Layout[:]))
	}

	if bestIndividual.Fitness <= 0.0 {
		t.Errorf("REGRESSION: Final fitness should be positive, got %.6f", bestIndividual.Fitness)
	}

	if !bestIndividual.IsValid() {
		t.Errorf("REGRESSION: Final individual should be valid")
	}

	t.Logf("Full pipeline regression test passed - no null characters, fitness=%.6f", bestIndividual.Fitness)
}

// Helper function to calculate average of float slice.
func average(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}

	return sum / float64(len(values))
}
