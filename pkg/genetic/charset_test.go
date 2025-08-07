package genetic

import (
	"testing"
)

func TestAlphabetOnly(t *testing.T) {
	charset := AlphabetOnly()
	if charset == nil {
		t.Fatal("AlphabetOnly returned nil")
	}

	if charset.Size != 26 {
		t.Errorf("AlphabetOnly size should be 26, got %d", charset.Size)
	}

	if !charset.Contains('a') || !charset.Contains('z') {
		t.Error("AlphabetOnly should contain 'a' and 'z'")
	}

	if charset.Contains('1') || charset.Contains('!') {
		t.Error("AlphabetOnly should not contain numbers or special characters")
	}
}

func TestProgrammingCharset(t *testing.T) {
	charset := ProgrammingCharset()
	if charset == nil {
		t.Fatal("ProgrammingCharset returned nil")
	}

	// Should contain letters
	if !charset.Contains('a') || !charset.Contains('z') {
		t.Error("ProgrammingCharset should contain letters")
	}

	// Should contain numbers
	if !charset.Contains('0') || !charset.Contains('9') {
		t.Error("ProgrammingCharset should contain numbers")
	}

	// Should contain special characters
	if !charset.Contains('$') || !charset.Contains('(') || !charset.Contains(')') {
		t.Error("ProgrammingCharset should contain special characters like $, (, )")
	}

	if !charset.Contains('{') || !charset.Contains('}') {
		t.Error("ProgrammingCharset should contain braces")
	}

	if !charset.Contains(';') || !charset.Contains('/') {
		t.Error("ProgrammingCharset should contain programming symbols")
	}
}

func TestFullKeyboardCharset(t *testing.T) {
	charset := FullKeyboardCharset()
	if charset == nil {
		t.Fatal("FullKeyboardCharset returned nil")
	}

	// Should contain all programming charset characters
	progCharset := ProgrammingCharset()
	for _, char := range progCharset.Characters {
		if !charset.Contains(char) {
			t.Errorf("FullKeyboardCharset should contain programming character '%c'", char)
		}
	}

	// Should contain space
	if !charset.Contains(' ') {
		t.Error("FullKeyboardCharset should contain space")
	}
}

func TestCharacterSetValidation(t *testing.T) {
	charset := ProgrammingCharset()

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

func TestCharsetByName(t *testing.T) {
	testCases := []struct {
		name     string
		expected int
	}{
		{"alphabet", 26},
		{"alphanumeric", 36},
		{"programming", len(ProgrammingCharset().Characters)},
		{"full", len(FullKeyboardCharset().Characters)},
		{"unknown", 26}, // Should default to alphabet
	}

	for _, tc := range testCases {
		charset := GetCharsetByName(tc.name)
		if charset.Size != tc.expected {
			t.Errorf("Charset %s should have size %d, got %d", tc.name, tc.expected, charset.Size)
		}
	}
}

func TestNewRandomIndividualWithCharset(t *testing.T) {
	charset := ProgrammingCharset()
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
