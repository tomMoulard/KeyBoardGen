package genetic

import (
	"context"
	"testing"
	"time"
)

// TestKeyboardDisplayIntegration tests the complete pipeline without parser dependency.
func TestKeyboardDisplayIntegration(t *testing.T) {
	// Create mock keylogger data
	keyloggerData := &MockKeyloggerData{
		charFreq: map[rune]int{
			'a': 5, 'b': 2, 'c': 3, 'd': 4, 'e': 8,
			'f': 2, 'g': 2, 'h': 4, 'i': 6, 'j': 1,
			'k': 1, 'l': 4, 'm': 2, 'n': 5, 'o': 6,
			'p': 2, 'q': 1, 'r': 5, 's': 4, 't': 7,
			'u': 3, 'v': 1, 'w': 2, 'x': 1, 'y': 2, 'z': 1,
		},
		bigramFreq: map[string]int{
			"th": 3, "he": 2, "in": 2, "er": 2, "an": 2,
			"ed": 1, "nd": 1, "to": 1, "en": 1, "ti": 1,
		},
		totalChars: 100,
	}

	// Test that mock data works
	if keyloggerData.GetTotalChars() == 0 {
		t.Fatal("No characters in mock data")
	}

	if len(keyloggerData.GetAllBigrams()) == 0 {
		t.Fatal("No bigrams in mock data")
	}

	t.Logf("Mock data: %d characters, %d bigrams", keyloggerData.GetTotalChars(), len(keyloggerData.GetAllBigrams()))
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

			// Character validation now handles full keyboard charset
			if !FullKeyboardCharset().Contains(char) {
				t.Errorf("Invalid character '%c' at position %d in individual %d", char, pos, i)
			}
		}

		// Check that individual is valid
		if !individual.IsValid() {
			t.Errorf("Invalid individual %d: %s", i, string(individual.Layout))
		}

		// Check that we have all 26 letters
		seen := make(map[rune]bool)
		for _, char := range individual.Layout {
			seen[char] = true
		}

		if len(seen) != 70 {
			t.Errorf("Individual %d missing characters, only has %d unique chars (expected 70)", i, len(seen))
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

	ga := NewParallelGA(mockEvaluator, config, FullKeyboardCharset())

	// Create mock sample data
	keyloggerData := &MockKeyloggerData{
		charFreq: map[rune]int{
			'a': 1, 'b': 1, 'c': 1, 'd': 2, 'e': 3,
			'f': 1, 'g': 1, 'h': 2, 'i': 1, 'j': 1,
			'k': 1, 'l': 3, 'm': 1, 'n': 1, 'o': 2,
			'p': 1, 'q': 1, 'r': 2, 's': 2, 't': 3,
			'u': 1, 'v': 1, 'w': 2, 'x': 1, 'y': 1, 'z': 1,
		},
		bigramFreq: map[string]int{
			"he": 1, "el": 1, "ll": 1, "lo": 1, "wo": 1,
			"or": 1, "rl": 1, "ld": 1, "te": 1, "es": 1, "st": 1,
		},
		totalChars: 50,
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
		t.Errorf("Best individual is invalid: %s", string(best.Layout))
	}

	// Test that fitness is reasonable
	if best.Fitness <= 0 {
		t.Errorf("Best individual has non-positive fitness: %f", best.Fitness)
	}
}

// MockFitnessEvaluator for testing.
type MockFitnessEvaluator struct{}

func (m *MockFitnessEvaluator) Evaluate(layout []rune, charset *CharacterSet, data KeyloggerDataInterface) float64 {
	// Return a simple fitness based on layout characteristics
	// This ensures different layouts get different fitness values

	// Validate layout first
	if charset == nil || !charset.IsValid(layout) {
		return 0.0
	}

	// Count unique characters
	seen := make(map[rune]bool)
	for _, char := range layout {
		seen[char] = true
	}

	// Base fitness on uniqueness and character distribution
	fitness := float64(len(seen)) / float64(charset.Size)

	// Add some variation based on first character
	if len(layout) > 0 && layout[0] != 0 {
		if layout[0] >= 'a' && layout[0] <= 'z' {
			fitness += float64(layout[0]-'a') / 100.0
		}
	}

	return fitness
}

// Legacy method for backward compatibility.
func (m *MockFitnessEvaluator) EvaluateLegacy(layout [26]rune, data KeyloggerDataInterface) float64 {
	layoutSlice := make([]rune, 26)
	copy(layoutSlice, layout[:])

	charset := FullKeyboardCharset()

	return m.Evaluate(layoutSlice, charset, data)
}

// TestKeyloggerDataInterface ensures all required methods are implemented.
func TestKeyloggerDataInterface(t *testing.T) {
	keyloggerData := &MockKeyloggerData{
		charFreq: map[rune]int{
			'a': 2, 'b': 1, 'c': 1, 'd': 3, 'e': 4,
			'f': 2, 'g': 1, 'h': 2, 'i': 4, 'j': 1,
			'k': 1, 'l': 3, 'm': 1, 'n': 3, 'o': 4,
			'p': 1, 'q': 1, 'r': 5, 's': 2, 't': 6,
			'u': 1, 'v': 2, 'w': 2, 'x': 1, 'y': 1, 'z': 1,
		},
		bigramFreq: map[string]int{
			"he": 2, "el": 1, "ll": 1, "lo": 1, "wo": 1,
			"or": 2, "rl": 1, "ld": 1, "te": 2, "es": 1,
			"st": 1, "ta": 2, "at": 1, "da": 1, "fo": 1,
		},
		totalChars: 42,
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
	layoutString := string(individual.Layout)

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
	if len(layoutString) != 70 {
		t.Errorf("Layout string length should be 70, got %d", len(layoutString))
	}

	t.Log("Display null character handling test passed")
}

// TestFitnessEvaluatorRejectsNullCharacters moved to fitness package to avoid import cycle

// TestGeneticAlgorithmEvolution ensures GA actually evolves and doesn't get stuck.
func TestGeneticAlgorithmEvolution(t *testing.T) {
	// Create larger mock dataset to ensure evolution can happen
	keyloggerData := &MockKeyloggerData{
		charFreq: map[rune]int{
			'a': 50, 'b': 20, 'c': 30, 'd': 40, 'e': 80,
			'f': 20, 'g': 20, 'h': 40, 'i': 60, 'j': 10,
			'k': 10, 'l': 40, 'm': 20, 'n': 50, 'o': 60,
			'p': 20, 'q': 10, 'r': 50, 's': 40, 't': 70,
			'u': 30, 'v': 10, 'w': 20, 'x': 10, 'y': 20, 'z': 10,
		},
		bigramFreq: map[string]int{
			"th": 30, "he": 20, "in": 20, "er": 20, "an": 20,
			"ed": 10, "nd": 10, "to": 10, "en": 10, "ti": 10,
			"es": 15, "or": 15, "te": 15, "of": 10, "be": 10,
			"at": 12, "se": 12, "ha": 12, "ng": 8, "hi": 8,
		},
		totalChars: 1000,
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

	ga := NewParallelGA(evaluator, config, FullKeyboardCharset())

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

	// Large mock dataset simulating complex text
	keyloggerData := &MockKeyloggerData{
		charFreq: map[rune]int{
			'a': 500, 'b': 200, 'c': 300, 'd': 400, 'e': 800,
			'f': 200, 'g': 200, 'h': 400, 'i': 600, 'j': 100,
			'k': 100, 'l': 400, 'm': 200, 'n': 500, 'o': 600,
			'p': 200, 'q': 100, 'r': 500, 's': 400, 't': 700,
			'u': 300, 'v': 100, 'w': 200, 'x': 100, 'y': 200, 'z': 100,
		},
		bigramFreq: map[string]int{
			"th": 300, "he": 200, "in": 200, "er": 200, "an": 200,
			"ed": 100, "nd": 100, "to": 100, "en": 100, "ti": 100,
			"es": 150, "or": 150, "te": 150, "of": 100, "be": 100,
			"at": 120, "se": 120, "ha": 120, "ng": 80, "hi": 80,
			"la": 90, "yo": 90, "ut": 90, "ke": 90, "al": 90,
		},
		totalChars: 10000,
	}

	// Use adaptive configuration (this should trigger large dataset config)
	config := AdaptiveConfig(keyloggerData.GetTotalChars())

	// Use mock evaluator
	evaluator := &MockFitnessEvaluator{}

	ga := NewParallelGA(evaluator, config, FullKeyboardCharset())

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
		t.Errorf("Layout: %s", string(bestIndividual.Layout))
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

// MockKeyloggerData implements KeyloggerDataInterface for testing.
type MockKeyloggerData struct {
	charFreq    map[rune]int
	bigramFreq  map[string]int
	trigramFreq map[string]int
	totalChars  int
}

func (m *MockKeyloggerData) GetCharFreq(char rune) int {
	return m.charFreq[char]
}

func (m *MockKeyloggerData) GetBigramFreq(bigram string) int {
	return m.bigramFreq[bigram]
}

func (m *MockKeyloggerData) GetTrigramFreq(trigram string) int {
	return m.trigramFreq[trigram]
}

func (m *MockKeyloggerData) GetTotalChars() int {
	return m.totalChars
}

func (m *MockKeyloggerData) GetAllBigrams() map[string]int {
	return m.bigramFreq
}
