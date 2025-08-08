package genetic

import (
	"testing"
)

func TestFullKeyboardCharset(t *testing.T) {
	t.Parallel()

	charset := FullKeyboardCharset()
	if charset == nil {
		t.Fatal("FullKeyboardCharset returned nil")
	}

	// Should contain programming-specific characters (no need to test all since it's now the full keyboard)
	if !charset.Contains('$') || !charset.Contains('{') || !charset.Contains('}') {
		t.Error("FullKeyboardCharset should contain programming characters")
	}

	// Should contain space
	if !charset.Contains(' ') {
		t.Error("FullKeyboardCharset should contain space")
	}
}

func TestCharacterSetValidation(t *testing.T) {
	t.Parallel()

	charset := FullKeyboardCharset()

	// Test valid layout
	validLayout := make([]rune, charset.Size)
	copy(validLayout, charset.Characters)

	if !charset.IsValid(validLayout) {
		t.Error("Valid layout should pass validation")
	}

	// Test invalid layout - wrong size
	invalidLayout := make([]rune, charset.Size-1)
	if charset.IsValid(invalidLayout) {
		t.Error("Layout with wrong size should fail validation")
	}

	// Test invalid layout - null character
	corruptedLayout := make([]rune, charset.Size)
	copy(corruptedLayout, charset.Characters)
	corruptedLayout[0] = 0 // null character

	if charset.IsValid(corruptedLayout) {
		t.Error("Layout with null character should fail validation")
	}

	// Test invalid layout - duplicate character
	duplicateLayout := make([]rune, charset.Size)
	copy(duplicateLayout, charset.Characters)
	duplicateLayout[1] = duplicateLayout[0] // create duplicate

	if charset.IsValid(duplicateLayout) {
		t.Error("Layout with duplicate character should fail validation")
	}
}

func TestCustomCharset(t *testing.T) {
	t.Parallel()

	customChars := "abc123!@#"
	charset := CustomCharset("test", customChars)

	if charset.Size != 9 {
		t.Errorf("Custom charset size should be 9, got %d", charset.Size)
	}

	if !charset.Contains('a') || !charset.Contains('1') || !charset.Contains('!') {
		t.Error("Custom charset should contain specified characters")
	}

	if charset.Contains('z') {
		t.Error("Custom charset should not contain unspecified characters")
	}
}

func TestNewRandomIndividualWithCharset(t *testing.T) {
	t.Parallel()

	charset := FullKeyboardCharset()
	individual := NewRandomIndividualWithCharset(charset)

	if individual.Charset != charset {
		t.Error("Individual should reference the correct charset")
	}

	if len(individual.Layout) != charset.Size {
		t.Errorf("Individual layout size should be %d, got %d", charset.Size, len(individual.Layout))
	}

	if !individual.IsValid() {
		t.Error("Random individual should be valid")
	}

	// Test that all characters from charset are present
	seen := make(map[rune]bool)
	for _, char := range individual.Layout {
		seen[char] = true
	}

	if len(seen) != charset.Size {
		t.Errorf("Individual should contain all %d unique characters, got %d", charset.Size, len(seen))
	}
}
