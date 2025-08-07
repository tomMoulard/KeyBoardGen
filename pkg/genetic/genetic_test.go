package genetic

import (
	"testing"
)

func TestNewRandomIndividual(t *testing.T) {
	ind := NewRandomIndividual()

	// Test that individual is valid
	if !ind.IsValid() {
		t.Error("NewRandomIndividual created invalid individual")
	}

	// Test that all 26 letters are present
	seen := make(map[rune]bool)
	for _, char := range ind.Layout {
		seen[char] = true
	}

	if len(seen) != 26 {
		t.Errorf("Expected 26 unique characters, got %d", len(seen))
	}

	// Test that all characters are lowercase letters
	for _, char := range ind.Layout {
		if char < 'a' || char > 'z' {
			t.Errorf("Invalid character in layout: %c", char)
		}
	}
}

func TestIndividualClone(t *testing.T) {
	original := NewRandomIndividual()
	original.Fitness = 0.5
	original.Age = 10

	clone := original.Clone()

	// Test that clone has same values
	if clone.Fitness != original.Fitness {
		t.Errorf("Clone fitness mismatch: expected %f, got %f", original.Fitness, clone.Fitness)
	}

	if clone.Age != original.Age {
		t.Errorf("Clone age mismatch: expected %d, got %d", original.Age, clone.Age)
	}

	// Test that layouts are equal
	for i := range original.Layout {
		if clone.Layout[i] != original.Layout[i] {
			t.Errorf("Clone layout mismatch at position %d: expected %c, got %c",
				i, original.Layout[i], clone.Layout[i])
		}
	}

	// Test that modifying clone doesn't affect original
	clone.Fitness = 0.8
	clone.Layout[0] = 'x'

	if original.Fitness == 0.8 {
		t.Error("Modifying clone affected original fitness")
	}

	if original.Layout[0] == 'x' {
		t.Error("Modifying clone layout affected original")
	}
}

func TestIndividualIsValid(t *testing.T) {
	// Test valid individual
	valid := NewRandomIndividual()
	if !valid.IsValid() {
		t.Error("Valid individual reported as invalid")
	}

	// Test invalid individual with duplicate
	invalid := NewRandomIndividual()

	invalid.Layout[0] = invalid.Layout[1] // Create duplicate
	if invalid.IsValid() {
		t.Error("Invalid individual with duplicate reported as valid")
	}

	// Test invalid individual with invalid character
	invalid2 := NewRandomIndividual()

	invalid2.Layout[0] = '1' // Not a lowercase letter
	if invalid2.IsValid() {
		t.Error("Invalid individual with number reported as valid")
	}
}

func TestSwapMutation(t *testing.T) {
	mutator := NewMutator(SwapMutation, 1.0) // 100% mutation rate
	original := NewRandomIndividual()
	original.Fitness = 0.5

	mutated := mutator.Apply(original)

	// Test that mutated individual is valid
	if !mutated.IsValid() {
		t.Error("Mutation produced invalid individual")
	}

	// Test that fitness was reset
	if mutated.Fitness != 0.0 {
		t.Errorf("Expected fitness to be reset to 0, got %f", mutated.Fitness)
	}

	// Test that original is unchanged
	if original.Fitness != 0.5 {
		t.Error("Mutation modified original individual")
	}

	// Test that layout changed (with very high probability)
	changes := 0

	for i := range original.Layout {
		if original.Layout[i] != mutated.Layout[i] {
			changes++
		}
	}

	// Should have exactly 2 changes (one swap)
	if changes != 2 {
		t.Errorf("Expected 2 position changes from swap, got %d", changes)
	}
}

func TestOrderCrossover(t *testing.T) {
	crossover := NewCrossover(OrderCrossover)

	// Create test parents with proper initialization
	charset := AlphabetOnly()
	parent1 := Individual{
		Layout:  make([]rune, 26),
		Charset: charset,
	}
	parent2 := Individual{
		Layout:  make([]rune, 26),
		Charset: charset,
	}

	// Set up known layouts
	for i := range 26 {
		parent1.Layout[i] = rune('a' + i) // abcd...z
		parent2.Layout[i] = rune('z' - i) // zyxw...a
	}

	child := crossover.Apply(parent1, parent2)

	// Test that child is valid
	if !child.IsValid() {
		t.Error("Crossover produced invalid child")
	}

	// Test that child has elements from both parents
	// (This is a basic test - more sophisticated tests could verify crossover mechanics)
	seen := make(map[rune]bool)
	for _, char := range child.Layout {
		if seen[char] {
			t.Error("Crossover produced duplicate characters")
		}

		seen[char] = true
	}
}

func TestTournamentSelection(t *testing.T) {
	selector := NewSelector(TournamentSelection, DefaultConfig())

	// Create test population with known fitness values
	population := make(Population, 10)
	for i := range population {
		population[i] = NewRandomIndividual()
		population[i].Fitness = float64(i) // Fitness 0, 1, 2, ..., 9
	}

	// Select individuals
	selected := selector.Select(population, 5)

	if len(selected) != 5 {
		t.Errorf("Expected 5 selected individuals, got %d", len(selected))
	}

	// Test that selected individuals are valid
	for _, ind := range selected {
		if !ind.IsValid() {
			t.Error("Selection returned invalid individual")
		}
	}
}

func TestCalculatePopulationDiversity(t *testing.T) {
	// Test with identical population (zero diversity)
	identical := make(Population, 5)

	template := NewRandomIndividual()
	for i := range identical {
		identical[i] = template.Clone()
	}

	diversity := CalculatePopulationDiversity(identical)
	if diversity != 0.0 {
		t.Errorf("Expected zero diversity for identical population, got %f", diversity)
	}

	// Test with completely different population (high diversity)
	different := make(Population, 2)
	different[0] = NewRandomIndividual()
	different[1] = NewRandomIndividual()

	// Ensure they're different
	different[1].Layout[0] = different[0].Layout[1]
	different[1].Layout[1] = different[0].Layout[0]

	diversity2 := CalculatePopulationDiversity(different)
	if diversity2 == 0.0 {
		t.Error("Expected non-zero diversity for different individuals")
	}
}

func TestValidateChild(t *testing.T) {
	// Create invalid child with duplicates
	invalid := Individual{
		Layout:  make([]rune, 26),
		Charset: AlphabetOnly(),
	}
	invalid.Layout[0] = 'a'
	invalid.Layout[1] = 'a' // Duplicate
	invalid.Layout[2] = 'b'
	// Leave rest empty (invalid)

	validated := ValidateChild(invalid)

	if !validated.IsValid() {
		t.Error("ValidateChild failed to produce valid individual")
	}
}

func TestKeyloggerData(t *testing.T) {
	data := NewKeyloggerData()

	// Test adding characters
	data.AddChar('a')
	data.AddChar('b')
	data.AddChar('a') // Duplicate

	if data.GetCharFreq('a') != 2 {
		t.Errorf("Expected frequency 2 for 'a', got %d", data.GetCharFreq('a'))
	}

	if data.GetCharFreq('b') != 1 {
		t.Errorf("Expected frequency 1 for 'b', got %d", data.GetCharFreq('b'))
	}

	if data.TotalChars != 3 {
		t.Errorf("Expected total chars 3, got %d", data.TotalChars)
	}

	// Test adding bigrams
	data.AddBigram("ab")
	data.AddBigram("ab") // Duplicate

	// Note: We need to implement GetBigramFreq method in KeyloggerData
	// This test assumes the method exists
}

// Benchmark tests.
func BenchmarkNewRandomIndividual(b *testing.B) {
	for range b.N {
		NewRandomIndividual()
	}
}

func BenchmarkSwapMutation(b *testing.B) {
	mutator := NewMutator(SwapMutation, 1.0)
	individual := NewRandomIndividual()

	b.ResetTimer()

	for range b.N {
		mutator.Apply(individual)
	}
}

func BenchmarkOrderCrossover(b *testing.B) {
	crossover := NewCrossover(OrderCrossover)
	parent1 := NewRandomIndividual()
	parent2 := NewRandomIndividual()

	b.ResetTimer()

	for range b.N {
		crossover.Apply(parent1, parent2)
	}
}

func BenchmarkTournamentSelection(b *testing.B) {
	selector := NewSelector(TournamentSelection, DefaultConfig())

	population := make(Population, 100)
	for i := range population {
		population[i] = NewRandomIndividual()
		population[i].Fitness = float64(i)
	}

	b.ResetTimer()

	for range b.N {
		selector.Select(population, 10)
	}
}
