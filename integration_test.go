package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// LayoutResult represents the JSON output from the main program.
type LayoutResult struct {
	Age       int            `json:"age"`
	Fitness   float64        `json:"fitness"`
	Layout    string         `json:"layout"`
	Positions map[string]int `json:"positions"`
	Timestamp string         `json:"timestamp"`
}

// TestMainApplicationNullCharacterRegression tests the complete application.
func TestMainApplicationNullCharacterRegression(t *testing.T) {
	// Build the application
	buildCmd := exec.Command("go", "build", "-o", "keyboardgen_test", "./cmd/keyboardgen")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}

	defer os.Remove("keyboardgen_test")

	// Create test input file with sufficient data
	testInput := "test_input.txt"

	testData := strings.Repeat("the quick brown fox jumps over the lazy dog hello world programming test keyboard layout optimization ", 5)

	if err := os.WriteFile(testInput, []byte(testData), 0o644); err != nil {
		t.Fatalf("Failed to create test input: %v", err)
	}
	defer os.Remove(testInput)

	// Create test output file path
	testOutput := "test_output.json"
	defer os.Remove(testOutput)

	// Run the application with parameters that should produce a valid result
	cmd := exec.Command("./keyboardgen_test",
		"-input", testInput,
		"-output", testOutput,
		"-generations", "5",
		"-population", "20",
		"-verbose")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Application failed: %v\nOutput: %s", err, string(output))
	}

	// Verify the output file was created
	if _, err := os.Stat(testOutput); os.IsNotExist(err) {
		t.Fatalf("Output file was not created")
	}

	// Read and parse the result
	resultBytes, err := os.ReadFile(testOutput)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var result LayoutResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("Failed to parse output JSON: %v", err)
	}

	// CRITICAL REGRESSION TESTS - These would have caught the original bug

	// Test 1: Layout should not contain null characters
	if strings.Contains(result.Layout, "\u0000") {
		t.Errorf("CRITICAL REGRESSION: Layout contains null characters: %q", result.Layout)
	}

	// Test 2: Layout should be exactly 26 characters
	if len(result.Layout) != 26 {
		t.Errorf("REGRESSION: Layout should be 26 characters, got %d", len(result.Layout))
	}

	// Test 3: Layout should contain all letters a-z
	letterCount := make(map[rune]int)
	for _, char := range result.Layout {
		letterCount[char]++
	}

	if len(letterCount) != 26 {
		t.Errorf("REGRESSION: Layout should contain 26 unique letters, got %d", len(letterCount))
	}

	for char := 'a'; char <= 'z'; char++ {
		if letterCount[char] != 1 {
			t.Errorf("REGRESSION: Letter '%c' appears %d times, should appear exactly once", char, letterCount[char])
		}
	}

	// Test 4: Fitness should be positive
	if result.Fitness <= 0.0 {
		t.Errorf("REGRESSION: Fitness should be positive, got %.6f", result.Fitness)
	}

	// Test 5: Positions map should be valid
	if len(result.Positions) != 26 {
		t.Errorf("REGRESSION: Positions map should have 26 entries, got %d", len(result.Positions))
	}

	// Check that positions map doesn't contain null character key
	if _, hasNull := result.Positions["\u0000"]; hasNull {
		t.Errorf("CRITICAL REGRESSION: Positions map contains null character key")
	}

	// Test 6: Verify output contains expected text patterns (not null characters)
	outputStr := string(output)
	if strings.Contains(outputStr, "null") || strings.Contains(outputStr, "\u0000") {
		t.Errorf("REGRESSION: Application output contains null-related content")
	}

	// Test 7: Output should mention QWERTY comparison with actual fitness
	if !strings.Contains(outputStr, "QWERTY") || !strings.Contains(outputStr, "COMPARISON") {
		t.Errorf("REGRESSION: Output should contain QWERTY comparison")
	}

	// Test 8: Should show evolution progress (fitness should be mentioned multiple times)
	fitnessMatches := strings.Count(outputStr, "fitness")
	if fitnessMatches < 3 { // At least in progress updates and final result
		t.Errorf("REGRESSION: Expected multiple fitness mentions, got %d", fitnessMatches)
	}

	t.Logf("✅ Main application regression test passed")
	t.Logf("Final fitness: %.6f", result.Fitness)
	t.Logf("Layout: %s", result.Layout)
}

// TestMainApplicationLargeDataset tests with larger dataset (like Harry Potter scenario).
func TestMainApplicationLargeDataset(t *testing.T) {
	// This test may take longer, so allow for timeout
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}

	// Build the application
	buildCmd := exec.Command("go", "build", "-o", "keyboardgen_test", "./cmd/keyboardgen")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}

	defer os.Remove("keyboardgen_test")

	// Create large test input (simulating Harry Potter-like dataset)
	testInput := "large_test_input.txt"
	// Create manageable chunks to avoid scanner token limit
	baseText := "the quick brown fox jumps over the lazy dog hello world programming test keyboard layout optimization genetic algorithms "

	largeText := ""
	for i := range 500 {
		largeText += baseText
		if i%10 == 0 {
			largeText += "\n" // Add line breaks to avoid token size issues
		}
	}

	if err := os.WriteFile(testInput, []byte(largeText), 0o644); err != nil {
		t.Fatalf("Failed to create large test input: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Remove(testInput)
	})

	testOutput := "large_test_output.json"

	t.Cleanup(func() {
		_ = os.Remove(testOutput)
	})

	// Run with default settings (should trigger adaptive configuration)
	cmd := exec.Command("./keyboardgen_test",
		"-input", testInput,
		"-output", testOutput)

	// Set timeout for large dataset processing - handled by test timeout

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Large dataset test failed: %v\nOutput: %s", err, string(output))
	}

	// Read result
	resultBytes, err := os.ReadFile(testOutput)
	if err != nil {
		t.Fatalf("Failed to read large dataset output: %v", err)
	}

	var result LayoutResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("Failed to parse large dataset output: %v", err)
	}

	// Same critical tests as above
	if strings.Contains(result.Layout, "\u0000") {
		t.Errorf("CRITICAL REGRESSION (large dataset): Layout contains null characters")
	}

	if result.Fitness <= 0.0 {
		t.Errorf("REGRESSION (large dataset): Fitness should be positive, got %.6f", result.Fitness)
	}

	// Test adaptive configuration was used
	outputStr := string(output)
	if !strings.Contains(outputStr, "adaptive configuration") {
		t.Errorf("Large dataset should trigger adaptive configuration")
	}

	// Test for evolution (large dataset should show fitness improvement)
	if strings.Count(outputStr, "Generation") < 10 {
		t.Errorf("Large dataset should show multiple generations")
	}

	t.Logf("✅ Large dataset regression test passed")
	t.Logf("Final fitness: %.6f", result.Fitness)
	t.Logf("Layout length: %d characters", len(result.Layout))
}

// TestMainApplicationErrorHandling tests error conditions.
func TestMainApplicationErrorHandling(t *testing.T) {
	// Build the application
	buildCmd := exec.Command("go", "build", "-o", "keyboardgen_test", "./cmd/keyboardgen")

	err := buildCmd.Run()
	if err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}

	defer os.Remove("keyboardgen_test")

	// Test 1: Missing input file
	cmd1 := exec.Command("./keyboardgen_test", "-input", "nonexistent.txt")

	output1, err1 := cmd1.CombinedOutput()
	if err1 == nil {
		t.Errorf("Expected error for missing input file, but command succeeded")
	}

	// Should not crash with null character error
	if strings.Contains(string(output1), "null character") {
		t.Errorf("Error message mentions null characters, suggesting regression")
	}

	// Test 2: Empty input file
	emptyFile := "empty_test.txt"

	err = os.WriteFile(emptyFile, []byte(""), 0o644)
	if err != nil {
		t.Fatalf("Failed to create empty test file: %v", err)
	}
	defer os.Remove(emptyFile)

	cmd2 := exec.Command("./keyboardgen_test", "-input", emptyFile)

	output2, err2 := cmd2.CombinedOutput()
	if err2 == nil {
		t.Logf("Empty file handling: %s", string(output2))
		// Empty file might be handled gracefully, that's OK
	}

	// Test 3: Invalid parameters
	cmd3 := exec.Command("./keyboardgen_test", "-population", "-1")

	_, err3 := cmd3.CombinedOutput()
	if err3 == nil {
		t.Errorf("Expected error for invalid population size")
	}

	t.Logf("✅ Error handling tests completed")
}

// TestMainApplicationMakefile tests the Makefile targets.
func TestMainApplicationMakefile(t *testing.T) {
	// Test basic example
	cmd1 := exec.Command("make", "example")

	output1, err1 := cmd1.CombinedOutput()
	if err1 != nil {
		t.Fatalf("Make example failed: %v\nOutput: %s", err1, string(output1))
	}

	// Verify the result file
	resultFile := "examples/result.json"
	if _, err := os.Stat(resultFile); os.IsNotExist(err) {
		t.Errorf("Make example didn't create result file")
	} else {
		// Check the result file for regressions
		resultBytes, err := os.ReadFile(resultFile)
		if err != nil {
			t.Fatalf("Failed to read result file: %v", err)
		}

		var result LayoutResult
		if err := json.Unmarshal(resultBytes, &result); err != nil {
			t.Fatalf("Failed to parse make example result: %v", err)
		}

		if strings.Contains(result.Layout, "\u0000") {
			t.Errorf("CRITICAL REGRESSION in make example: Layout contains null characters")
		}

		if result.Fitness <= 0.0 {
			t.Errorf("REGRESSION in make example: Invalid fitness %.6f", result.Fitness)
		}
	}

	t.Logf("✅ Makefile integration test passed")
}
