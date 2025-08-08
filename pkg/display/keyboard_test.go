package display

import (
	"testing"

	"github.com/tommoulard/keyboardgen/pkg/genetic"
)

// TestCharacterUniquenessConstraint verifies that each character appears exactly once across all keyboard layers.
func TestCharacterUniquenessConstraint(t *testing.T) {
	// Create a test individual with the full keyboard charset
	charset := genetic.FullKeyboardCharset()
	testLayout := make([]rune, len(charset.Characters))
	copy(testLayout, charset.Characters)

	individual := genetic.Individual{
		Layout:  testLayout,
		Charset: charset,
		Fitness: 0.6,
		Age:     1,
	}

	// Create keyboard display and generate layered layout
	kd := NewKeyboardDisplay()
	optimizedLayout := kd.CreateOptimizedLayeredLayout(individual)

	// Collect all characters from all layers
	baseChars := make([]rune, 0)
	shiftChars := make([]rune, 0)
	altgrChars := make([]rune, 0)

	for pos := range len(individual.Layout) {
		if layeredKey, exists := optimizedLayout.Keys[pos]; exists {
			// Base layer
			baseChars = append(baseChars, layeredKey.BaseChar)

			// Shift layer (only if character is assigned)
			if layeredKey.ShiftChar != 0 {
				shiftChars = append(shiftChars, layeredKey.ShiftChar)
			}

			// AltGr layer (only if character is assigned)
			if layeredKey.AltGrChar != nil {
				altgrChars = append(altgrChars, *layeredKey.AltGrChar)
			}
		}
	}

	t.Run("BaseLayerNoDuplicates", func(t *testing.T) {
		charCount := make(map[rune]int)
		for _, char := range baseChars {
			charCount[char]++
		}

		for char, count := range charCount {
			if count > 1 {
				t.Errorf("Character '%c' appears %d times in base layer, expected exactly 1", char, count)
			}
		}
	})

	t.Run("ShiftLayerNoDuplicates", func(t *testing.T) {
		charCount := make(map[rune]int)
		for _, char := range shiftChars {
			charCount[char]++
		}

		for char, count := range charCount {
			if count > 1 {
				t.Errorf("Character '%c' appears %d times in shift layer, expected exactly 1", char, count)
			}
		}
	})

	t.Run("LetterMappingCorrectness", func(t *testing.T) {
		// Verify that every lowercase letter in base layer has corresponding uppercase in shift layer
		for pos := range len(individual.Layout) {
			if layeredKey, exists := optimizedLayout.Keys[pos]; exists {
				baseChar := layeredKey.BaseChar
				shiftChar := layeredKey.ShiftChar

				if baseChar >= 'a' && baseChar <= 'z' {
					expectedUppercase := baseChar - 'a' + 'A'
					if shiftChar != expectedUppercase {
						t.Errorf("Position %d: base char '%c' should map to '%c' in shift layer, got '%c'",
							pos, baseChar, expectedUppercase, shiftChar)
					}
				}
			}
		}
	})

	t.Run("CrossLayerUniqueness", func(t *testing.T) {
		// Count all characters across all layers
		allChars := make([]rune, 0)
		allChars = append(allChars, baseChars...)
		allChars = append(allChars, shiftChars...)
		allChars = append(allChars, altgrChars...)

		charCount := make(map[rune]int)
		for _, char := range allChars {
			charCount[char]++
		}

		// Check for unexpected duplicates
		// Letters are expected to appear twice (lowercase + uppercase)
		for char, count := range charCount {
			if char >= 'a' && char <= 'z' {
				// Lowercase letters should appear exactly once
				if count != 1 {
					t.Errorf("Lowercase letter '%c' appears %d times, expected exactly 1", char, count)
				}
			} else if char >= 'A' && char <= 'Z' {
				// Uppercase letters should appear exactly once (in shift layer)
				if count != 1 {
					t.Errorf("Uppercase letter '%c' appears %d times, expected exactly 1", char, count)
				}
			} else {
				// Non-letters should appear exactly once
				if count != 1 {
					t.Errorf("Non-letter character '%c' appears %d times, expected exactly 1", char, count)
				}
			}
		}
	})

	t.Run("CharsetCoverage", func(t *testing.T) {
		// Verify that all characters from the original charset appear exactly once in base layer
		baseCharSet := make(map[rune]bool)
		for _, char := range baseChars {
			baseCharSet[char] = true
		}

		for _, expectedChar := range charset.Characters {
			if !baseCharSet[expectedChar] {
				t.Errorf("Character '%c' from charset is missing in base layer", expectedChar)
			}
		}

		if len(baseChars) != len(charset.Characters) {
			t.Errorf("Base layer has %d characters, expected %d from charset",
				len(baseChars), len(charset.Characters))
		}
	})

	t.Run("ShiftLayerOptimalUsage", func(t *testing.T) {
		// Verify that shift layer contains only uppercase letters and no base layer duplicates
		baseCharSet := make(map[rune]bool)
		for _, char := range baseChars {
			baseCharSet[char] = true
		}

		for _, shiftChar := range shiftChars {
			// If it's an uppercase letter, verify corresponding lowercase exists in base
			if shiftChar >= 'A' && shiftChar <= 'Z' {
				expectedLowercase := shiftChar - 'A' + 'a'
				if !baseCharSet[expectedLowercase] {
					t.Errorf("Uppercase '%c' in shift layer has no corresponding '%c' in base layer",
						shiftChar, expectedLowercase)
				}
			} else {
				// Non-letters in shift layer should not duplicate base layer
				if baseCharSet[shiftChar] {
					t.Errorf("Non-letter character '%c' appears in both base and shift layers", shiftChar)
				}
			}
		}
	})
}

// TestCreateOptimizedLayeredLayoutEdgeCases tests edge cases for the layout creation.
func TestCreateOptimizedLayeredLayoutEdgeCases(t *testing.T) {
	kd := NewKeyboardDisplay()

	t.Run("EmptyIndividual", func(t *testing.T) {
		individual := genetic.Individual{
			Layout:  []rune{},
			Charset: genetic.FullKeyboardCharset(),
			Fitness: 0.0,
			Age:     0,
		}

		layout := kd.CreateOptimizedLayeredLayout(individual)
		if layout == nil {
			t.Error("CreateOptimizedLayeredLayout returned nil for empty individual")
		}

		if len(layout.Keys) != 0 {
			t.Errorf("Expected 0 keys for empty layout, got %d", len(layout.Keys))
		}
	})

	t.Run("NilCharset", func(t *testing.T) {
		individual := genetic.Individual{
			Layout:  []rune{'a', 'b', 'c'},
			Charset: nil,
			Fitness: 0.0,
			Age:     0,
		}

		layout := kd.CreateOptimizedLayeredLayout(individual)
		if layout == nil {
			t.Error("CreateOptimizedLayeredLayout returned nil for nil charset")
		}

		// Verify that shift layer has empty characters for non-letters when charset is nil
		for pos := range len(individual.Layout) {
			if layeredKey, exists := layout.Keys[pos]; exists {
				baseChar := layeredKey.BaseChar
				if baseChar < 'a' || baseChar > 'z' {
					if layeredKey.ShiftChar != 0 {
						t.Errorf("Position %d: non-letter should have empty shift char when charset is nil, got '%c'",
							pos, layeredKey.ShiftChar)
					}
				}
			}
		}
	})

	t.Run("OnlyLetters", func(t *testing.T) {
		letters := []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}
		individual := genetic.Individual{
			Layout:  letters,
			Charset: genetic.FullKeyboardCharset(),
			Fitness: 0.0,
			Age:     0,
		}

		layout := kd.CreateOptimizedLayeredLayout(individual)

		// All shift characters should be uppercase versions
		for pos := range letters {
			if layeredKey, exists := layout.Keys[pos]; exists {
				baseChar := layeredKey.BaseChar

				expectedShift := baseChar - 'a' + 'A'
				if layeredKey.ShiftChar != expectedShift {
					t.Errorf("Position %d: expected shift char '%c', got '%c'",
						pos, expectedShift, layeredKey.ShiftChar)
				}
			}
		}
	})

	t.Run("OnlyNonLetters", func(t *testing.T) {
		symbols := []rune{'1', '2', '3', '!', '@', '#', '$', '%', '^', '&'}
		individual := genetic.Individual{
			Layout:  symbols,
			Charset: genetic.FullKeyboardCharset(),
			Fitness: 0.0,
			Age:     0,
		}

		layout := kd.CreateOptimizedLayeredLayout(individual)

		// Shift characters should be different from base characters (no duplicates)
		baseCharSet := make(map[rune]bool)

		for pos := range symbols {
			if layeredKey, exists := layout.Keys[pos]; exists {
				baseCharSet[layeredKey.BaseChar] = true
			}
		}

		for pos := range symbols {
			if layeredKey, exists := layout.Keys[pos]; exists {
				if layeredKey.ShiftChar != 0 && baseCharSet[layeredKey.ShiftChar] {
					t.Errorf("Position %d: shift char '%c' duplicates a base layer character",
						pos, layeredKey.ShiftChar)
				}
			}
		}
	})
}

// BenchmarkCreateOptimizedLayeredLayout benchmarks the layout creation performance.
func BenchmarkCreateOptimizedLayeredLayout(b *testing.B) {
	charset := genetic.FullKeyboardCharset()
	testLayout := make([]rune, len(charset.Characters))
	copy(testLayout, charset.Characters)

	individual := genetic.Individual{
		Layout:  testLayout,
		Charset: charset,
		Fitness: 0.6,
		Age:     1,
	}

	kd := NewKeyboardDisplay()

	b.ResetTimer()

	for range b.N {
		_ = kd.CreateOptimizedLayeredLayout(individual)
	}
}
