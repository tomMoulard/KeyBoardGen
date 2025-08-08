package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/tommoulard/keyboardgen/pkg/genetic"
)

// TestFullCharsetInOutputs verifies all characters appear in both console output and JSON file.
func TestFullCharsetInOutputs(t *testing.T) {
	// Create test input with all character types
	testContent := `hello world with comprehensive character testing!
programming symbols: () [] {} <> "" '' 
mathematics: + - * / = % ^ & | ~ ` + "`" + `
punctuation: . , ; : ? ! @ # $ _
numbers and digits: 0123456789
special cases: \\ // -- ++ == != <= >=
more text with all letters: abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ
final mixed content with various symbols and characters for complete testing!`

	// Write test file
	testFile := "charset_test_input.txt"

	err := os.WriteFile(testFile, []byte(testContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Build the application
	buildCmd := exec.Command("go", "build", "-o", "keyboardgen_test", "./cmd/keyboardgen")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}

	defer os.Remove("keyboardgen_test")

	// Run the application and capture output
	outputFile := "charset_test_output.json"
	cmd := exec.Command("./keyboardgen_test",
		"-input", testFile,
		"-output", outputFile,
		"-generations", "2",
		"-population", "10")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Application failed: %v\nOutput: %s", err, string(output))
	}

	defer os.Remove(outputFile)

	outputStr := string(output)

	// Test 1: Verify console output mentions full keyboard charset
	if !strings.Contains(outputStr, "full_keyboard") {
		t.Errorf("Console output should mention 'full_keyboard' charset")
	}

	if !strings.Contains(outputStr, "70 characters") {
		t.Errorf("Console output should mention '70 characters'")
	}

	// Test 2: Verify layered display shows optimized characters
	if !strings.Contains(outputStr, "OPTIMIZED KEYBOARD LAYERS") {
		t.Errorf("Console output should show 'OPTIMIZED KEYBOARD LAYERS'")
	}

	// Test 3: Load and verify JSON output
	jsonData, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read JSON output: %v", err)
	}

	var result struct {
		Layout                  string         `json:"layout"`
		OptimizedKeyboardLayers map[string]any `json:"optimized_keyboard_layers"`
		LayerMetadata           map[string]any `json:"layer_metadata"`
		Positions               map[string]int `json:"positions"`
	}

	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Test 4: Verify layout contains 70 characters
	if len(result.Layout) != 70 {
		t.Errorf("Layout should contain 70 characters, got %d", len(result.Layout))
	}

	// Test 5: Verify all characters are unique in layout
	charCount := make(map[rune]int)
	for _, char := range result.Layout {
		charCount[char]++
	}

	if len(charCount) != 70 {
		t.Errorf("Layout should contain 70 unique characters, got %d", len(charCount))
	}

	for char, count := range charCount {
		if count != 1 {
			t.Errorf("Character '%c' appears %d times, should appear exactly once", char, count)
		}
	}

	// Test 6: Verify JSON contains all expected character types
	layoutRunes := []rune(result.Layout)

	hasLetter := false
	hasNumber := false
	hasSymbol := false
	hasSpace := false

	for _, char := range layoutRunes {
		if char >= 'a' && char <= 'z' {
			hasLetter = true
		} else if char >= '0' && char <= '9' {
			hasNumber = true
		} else if char == ' ' {
			hasSpace = true
		} else {
			hasSymbol = true
		}
	}

	if !hasLetter {
		t.Errorf("Layout should contain letters")
	}

	if !hasNumber {
		t.Errorf("Layout should contain numbers")
	}

	if !hasSymbol {
		t.Errorf("Layout should contain symbols")
	}

	if !hasSpace {
		t.Errorf("Layout should contain space character")
	}

	// Test 7: Verify optimized keyboard layers structure exists
	if result.OptimizedKeyboardLayers == nil {
		t.Errorf("JSON should contain 'optimized_keyboard_layers'")
	}

	baseLayer := result.OptimizedKeyboardLayers["base"]
	if baseLayer == nil {
		t.Errorf("Optimized layers should contain 'base' layer")
	}

	shiftLayer := result.OptimizedKeyboardLayers["shift"]
	if shiftLayer == nil {
		t.Errorf("Optimized layers should contain 'shift' layer")
	}

	altgrLayer := result.OptimizedKeyboardLayers["altgr"]
	if altgrLayer == nil {
		t.Errorf("Optimized layers should contain 'altgr' layer")
	}

	// Test 8: Verify characters array exists in base layer
	if baseLayerMap, ok := baseLayer.(map[string]any); ok {
		if chars, exists := baseLayerMap["characters"]; exists {
			if charArray, isArray := chars.([]any); isArray {
				if len(charArray) != 70 {
					t.Errorf("Base layer characters array should contain 70 elements, got %d", len(charArray))
				}
			} else {
				t.Errorf("Base layer 'characters' should be an array")
			}
		} else {
			t.Errorf("Base layer should contain 'characters' array")
		}

		if layoutStr, exists := baseLayerMap["layout_string"]; exists {
			if str, isString := layoutStr.(string); isString {
				if len(str) != 70 {
					t.Errorf("Base layer layout_string should contain 70 characters, got %d", len(str))
				}
			} else {
				t.Errorf("Base layer 'layout_string' should be a string")
			}
		} else {
			t.Errorf("Base layer should contain 'layout_string'")
		}
	}

	// Test 9: Verify positions map contains 70 entries
	if len(result.Positions) != 70 {
		t.Errorf("Positions map should contain 70 entries, got %d", len(result.Positions))
	}

	t.Logf("Layout contains %d characters: %s", len(result.Layout), result.Layout)
	t.Logf("Character types: letters=%t, numbers=%t, symbols=%t, space=%t",
		hasLetter, hasNumber, hasSymbol, hasSpace)
}

// TestSpecificCharacterPresence tests for specific important characters.
func TestSpecificCharacterPresence(t *testing.T) {
	// Create comprehensive test input
	testContent := `programming test with parentheses (like this) and brackets [arrays] {objects}
mathematical operations: 1 + 2 = 3, 4 * 5 = 20, 6 / 2 = 3, 7 - 1 = 6
special symbols: @ # $ % ^ & * ! ? ~ ` + "`" + ` | \\ " ' : ; , . < > 
quotes and punctuation: "hello world" 'single quotes' and-dashes_underscores
more comprehensive testing with all keyboard characters for genetic optimization!`

	testFile := "specific_chars_test.txt"

	err := os.WriteFile(testFile, []byte(testContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Build and run application
	buildCmd := exec.Command("go", "build", "-o", "keyboardgen_test", "./cmd/keyboardgen")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}

	defer os.Remove("keyboardgen_test")

	outputFile := "specific_chars_output.json"
	cmd := exec.Command("./keyboardgen_test",
		"-input", testFile,
		"-output", outputFile,
		"-generations", "1",
		"-population", "10")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Application failed: %v\nOutput: %s", err, string(output))
	}

	defer os.Remove(outputFile)

	// Load JSON result
	jsonData, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read JSON output: %v", err)
	}

	var result struct {
		Layout string `json:"layout"`
	}

	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Test for specific character categories
	expectedCharsets := genetic.FullKeyboardCharset()

	expectedChars := make(map[rune]bool)
	for _, char := range expectedCharsets.Characters {
		expectedChars[char] = true
	}

	actualChars := make(map[rune]bool)
	for _, char := range result.Layout {
		actualChars[char] = true
	}

	// Verify all expected characters are present
	missingChars := []rune{}

	for expectedChar := range expectedChars {
		if !actualChars[expectedChar] {
			missingChars = append(missingChars, expectedChar)
		}
	}

	if len(missingChars) > 0 {
		t.Errorf("Missing characters from full keyboard charset: %v", string(missingChars))
	}

	// Verify no unexpected characters
	extraChars := []rune{}

	for actualChar := range actualChars {
		if !expectedChars[actualChar] {
			extraChars = append(extraChars, actualChar)
		}
	}

	if len(extraChars) > 0 {
		t.Errorf("Unexpected characters in layout: %v", string(extraChars))
	}

	// Test console output contains character information
	outputStr := string(output)

	// Should show character diversity in the most used keys
	if !strings.Contains(outputStr, "' '") { // Space character
		t.Errorf("Console output should show space character in analysis")
	}

	// Should show full keyboard charset name
	if !strings.Contains(outputStr, "full_keyboard") {
		t.Errorf("Console output should reference full_keyboard charset")
	}

	t.Logf("Layout: %s", result.Layout)
	t.Logf("Total characters: %d", len(result.Layout))
}

// TestLayerCharacterMapping tests that layered display shows all character types.
func TestLayerCharacterMapping(t *testing.T) {
	testContent := `layer mapping test with diverse characters: abcdef 123456 !@#$%^ (){}[] +=*/
programming requires many symbols for proper optimization and genetic algorithm testing
comprehensive character usage analysis with numbers, letters, symbols, and punctuation marks`

	testFile := "layer_mapping_test.txt"

	err := os.WriteFile(testFile, []byte(testContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Build and run with verbose output to see layers
	buildCmd := exec.Command("go", "build", "-o", "keyboardgen_test", "./cmd/keyboardgen")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}

	defer os.Remove("keyboardgen_test")

	outputFile := "layer_mapping_output.json"
	cmd := exec.Command("./keyboardgen_test",
		"-input", testFile,
		"-output", outputFile,
		"-generations", "1",
		"-population", "10",
		"-verbose")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Application failed: %v\nOutput: %s", err, string(output))
	}

	defer os.Remove(outputFile)

	outputStr := string(output)

	// Test console output shows layered keyboard
	if !strings.Contains(outputStr, "BASE") {
		t.Errorf("Console output should show BASE layer")
	}

	if !strings.Contains(outputStr, "SHIFT") {
		t.Errorf("Console output should show SHIFT layer")
	}

	if !strings.Contains(outputStr, "ALTGR") {
		t.Errorf("Console output should show ALTGR layer")
	}

	// Test that different character types appear in output
	hasLetterInOutput := false
	hasNumberInOutput := false
	hasSymbolInOutput := false

	// Look for patterns that indicate different character types in display
	for line := range strings.SplitSeq(outputStr, "\n") {
		if strings.Contains(line, "│") { // Table row with characters
			for _, char := range line {
				if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' {
					hasLetterInOutput = true
				} else if char >= '0' && char <= '9' {
					hasNumberInOutput = true
				} else if strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:'\",.<>?/~`", char) {
					hasSymbolInOutput = true
				}
			}
		}
	}

	if !hasLetterInOutput {
		t.Errorf("Console layered display should show letters")
	}

	if !hasNumberInOutput {
		t.Errorf("Console layered display should show numbers")
	}

	if !hasSymbolInOutput {
		t.Errorf("Console layered display should show symbols")
	}
}
