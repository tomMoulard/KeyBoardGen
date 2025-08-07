package display

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/tommoulard/keyboardgen/pkg/genetic"
)

// KeyboardDisplay handles keyboard layout visualization.
type KeyboardDisplay struct {
	showFrequency bool
	showColors    bool
	compact       bool
}

// NewKeyboardDisplay creates a new display handler.
func NewKeyboardDisplay() *KeyboardDisplay {
	return &KeyboardDisplay{
		showFrequency: false,
		showColors:    false,
		compact:       false,
	}
}

// SetOptions configures display options.
func (kd *KeyboardDisplay) SetOptions(showFreq, showColors, compact bool) {
	kd.showFrequency = showFreq
	kd.showColors = showColors
	kd.compact = compact
}

// Use the KeyloggerDataInterface from genetic package.
type KeyloggerDataInterface = genetic.KeyloggerDataInterface

// FitnessEvaluator interface for calculating fitness.
type FitnessEvaluator interface {
	Evaluate(layout []rune, charset *genetic.CharacterSet, data KeyloggerDataInterface) float64
	// Legacy method for backward compatibility
	EvaluateLegacy(layout [26]rune, data KeyloggerDataInterface) float64
}

// PrintLayout displays the keyboard layout in a visual format.
func (kd *KeyboardDisplay) PrintLayout(individual genetic.Individual, data KeyloggerDataInterface) {
	if kd.compact {
		kd.printCompactLayout(individual, data)
	} else {
		kd.printFullLayout(individual, data)
	}
}

// printFullLayout displays the full graphical keyboard.
func (kd *KeyboardDisplay) printFullLayout(individual genetic.Individual, data KeyloggerDataInterface) {
	fmt.Printf("\nKeyboard Layout (Fitness: %.6f)\n", individual.Fitness)
	fmt.Println("┌─────┬─────┬─────┬─────┬─────┬─────┬─────┬─────┬─────┬─────┐")

	// Top row (positions 0-9)
	fmt.Print("│")

	for i := range 10 {
		if i < len(individual.Layout) {
			char := individual.Layout[i]
			cell := kd.formatCell(char, data)
			fmt.Printf("%s│", cell)
		} else {
			fmt.Print("     │")
		}
	}

	fmt.Println()

	fmt.Println("├─────┼─────┼─────┼─────┼─────┼─────┼─────┼─────┼─────┼─────┤")

	// Middle row (positions 10-18)
	fmt.Print("│")

	for i := 10; i < 19; i++ {
		if i < len(individual.Layout) {
			char := individual.Layout[i]
			cell := kd.formatCell(char, data)
			fmt.Printf("%s│", cell)
		} else {
			fmt.Print("     │")
		}
	}

	fmt.Println()

	fmt.Println("├─────┼─────┼─────┼─────┼─────┼─────┼─────┼─────┼─────┼─────┤")

	// Bottom row (positions 19-25)
	fmt.Print("│")

	for i := 19; i < 26; i++ {
		char := individual.Layout[i]
		cell := kd.formatCell(char, data)
		fmt.Printf("%s│", cell)
	}
	// Fill remaining positions
	for i := 26; i < 28; i++ {
		fmt.Print("     │")
	}

	fmt.Println()

	fmt.Println("└─────┴─────┴─────┴─────┴─────┴─────┴─────┴─────┴─────┴─────┘")
}

// printCompactLayout displays a compact text version.
func (kd *KeyboardDisplay) printCompactLayout(individual genetic.Individual, data KeyloggerDataInterface) {
	fmt.Printf("Layout (Fitness: %.6f): ", individual.Fitness)

	// Group by keyboard rows
	topRow := string(individual.Layout[0:10])
	middleRow := string(individual.Layout[10:19])
	bottomRow := string(individual.Layout[19:26])

	fmt.Printf("%s | %s | %s\n", topRow, middleRow, bottomRow)
}

// formatCell formats a single keyboard cell.
func (kd *KeyboardDisplay) formatCell(char rune, data KeyloggerDataInterface) string {
	// Handle null/empty runes
	if char == 0 {
		return "     "
	}

	if data == nil {
		return fmt.Sprintf("  %c  ", char)
	}

	freq := data.GetCharFreq(char)

	if kd.showFrequency && data.GetTotalChars() > 0 {
		percent := float64(freq) * 100.0 / float64(data.GetTotalChars())
		if percent >= 10 {
			return fmt.Sprintf(" %c%.0f ", char, percent)
		} else if percent >= 1 {
			return fmt.Sprintf(" %c%.1f", char, percent)
		} else {
			return fmt.Sprintf("  %c  ", char)
		}
	}

	if kd.showColors {
		return kd.colorizeCell(char, freq, data.GetTotalChars())
	}

	return fmt.Sprintf("  %c  ", char)
}

// colorizeCell applies color coding based on frequency.
func (kd *KeyboardDisplay) colorizeCell(char rune, freq, total int) string {
	if total == 0 {
		return fmt.Sprintf("  %c  ", char)
	}

	percent := float64(freq) * 100.0 / float64(total)

	// Color coding based on frequency
	var colorCode string

	switch {
	case percent >= 8: // Very high frequency - red
		colorCode = "\033[41m"
	case percent >= 6: // High frequency - yellow
		colorCode = "\033[43m"
	case percent >= 4: // Medium frequency - green
		colorCode = "\033[42m"
	case percent >= 2: // Low frequency - blue
		colorCode = "\033[44m"
	default: // Very low frequency - default
		colorCode = ""
	}

	reset := "\033[0m"
	if colorCode != "" {
		return fmt.Sprintf("%s %c %s", colorCode, char, reset)
	}

	return fmt.Sprintf("  %c  ", char)
}

// PrintStatistics shows keyboard layout statistics.
func (kd *KeyboardDisplay) PrintStatistics(individual genetic.Individual, data KeyloggerDataInterface) {
	if data == nil {
		fmt.Printf("Fitness: %.6f, Age: %d generations\n", individual.Fitness, individual.Age)

		return
	}

	fmt.Printf("Layout Statistics:\n")
	fmt.Printf("- Fitness: %.6f\n", individual.Fitness)
	fmt.Printf("- Age: %d generations\n", individual.Age)
	fmt.Printf("- Total characters analyzed: %d\n", data.GetTotalChars())

	// Calculate finger usage distribution
	fingerUsage := kd.calculateFingerUsage(individual, data)

	fmt.Println("\nFinger usage distribution:")

	fingers := []string{"L.Pinky", "L.Ring", "L.Middle", "L.Index", "R.Index", "R.Middle", "R.Ring", "R.Pinky"}

	for i, usage := range fingerUsage {
		percent := float64(usage) * 100.0 / float64(data.GetTotalChars())
		fmt.Printf("  %s: %6d chars (%.1f%%)\n", fingers[i], usage, percent)
	}

	// Most used keys
	fmt.Println("\nMost frequently used keys:")

	keyFreqs := make([]struct {
		char rune
		freq int
		pos  int
	}, 26)

	for i, char := range individual.Layout {
		keyFreqs[i] = struct {
			char rune
			freq int
			pos  int
		}{char, data.GetCharFreq(char), i}
	}

	sort.Slice(keyFreqs, func(i, j int) bool {
		return keyFreqs[i].freq > keyFreqs[j].freq
	})

	for i := 0; i < 10 && i < len(keyFreqs); i++ {
		kf := keyFreqs[i]
		percent := float64(kf.freq) * 100.0 / float64(data.GetTotalChars())
		fmt.Printf("  %d. '%c' at position %2d: %6d (%.1f%%)\n",
			i+1, kf.char, kf.pos, kf.freq, percent)
	}
}

// calculateFingerUsage computes usage per finger.
func (kd *KeyboardDisplay) calculateFingerUsage(individual genetic.Individual, data KeyloggerDataInterface) []int {
	// Standard finger mapping for QWERTY-style layout
	fingerMap := map[int]int{
		// Left hand: 0-3 (pinky to index), Right hand: 4-7 (index to pinky)
		0: 0, 1: 1, 2: 2, 3: 3, 4: 3, 5: 4, 6: 4, 7: 5, 8: 6, 9: 7, // Top row
		10: 0, 11: 1, 12: 2, 13: 3, 14: 4, 15: 4, 16: 5, 17: 6, 18: 7, // Middle row
		19: 0, 20: 1, 21: 2, 22: 3, 23: 4, 24: 5, 25: 6, // Bottom row
	}

	fingerUsage := make([]int, 8)

	for pos, char := range individual.Layout {
		if finger, ok := fingerMap[pos]; ok {
			fingerUsage[finger] += data.GetCharFreq(char)
		}
	}

	return fingerUsage
}

// PrintComparison shows a side-by-side comparison with QWERTY.
func (kd *KeyboardDisplay) PrintComparison(individual genetic.Individual, data KeyloggerDataInterface) {
	kd.PrintComparisonWithEvaluator(individual, data, nil)
}

// PrintComparisonWithEvaluator shows comparison with actual QWERTY fitness calculation.
func (kd *KeyboardDisplay) PrintComparisonWithEvaluator(individual genetic.Individual, data KeyloggerDataInterface, evaluator FitnessEvaluator) {
	qwerty := [26]rune{
		'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', // Top row
		'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', // Middle row
		'z', 'x', 'c', 'v', 'b', 'n', 'm', // Bottom row
	}

	fmt.Println("\nComparison: Optimized vs QWERTY")
	fmt.Println("┌─────────────────────────────────┬─────────────────────────────────┐")
	fmt.Println("│           OPTIMIZED             │             QWERTY              │")
	fmt.Println("├─────────────────────────────────┼─────────────────────────────────┤")

	// Print each row comparison
	rows := [][]int{
		{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},       // Top
		{10, 11, 12, 13, 14, 15, 16, 17, 18}, // Middle
		{19, 20, 21, 22, 23, 24, 25},         // Bottom
	}

	for _, row := range rows {
		// Optimized layout row
		fmt.Print("│ ")

		for _, pos := range row {
			if pos < len(individual.Layout) {
				fmt.Printf("%c ", individual.Layout[pos])
			} else {
				fmt.Print("  ")
			}
		}

		fmt.Print(strings.Repeat(" ", 33-len(row)*2))

		fmt.Print("│ ")

		// QWERTY row
		for _, pos := range row {
			if pos < len(qwerty) {
				fmt.Printf("%c ", qwerty[pos])
			} else {
				fmt.Print("  ")
			}
		}

		fmt.Print(strings.Repeat(" ", 33-len(row)*2))
		fmt.Println("│")
	}

	fmt.Println("└─────────────────────────────────┴─────────────────────────────────┘")

	// Calculate QWERTY fitness for comparison
	if data != nil {
		var qwertyFitness float64
		if evaluator != nil {
			qwertyFitness = evaluator.EvaluateLegacy(qwerty, data)
		}

		if evaluator != nil {
			fmt.Printf("Fitness comparison: Optimized=%.6f vs QWERTY=%.6f (Improvement: %.1f%%)\n",
				individual.Fitness, qwertyFitness, (individual.Fitness/qwertyFitness-1.0)*100)
		} else {
			fmt.Printf("Fitness comparison: Optimized=%.6f vs QWERTY=%.6f\n",
				individual.Fitness, 0.0)
		}
	}
}

// PrintHeatmap shows a visual heatmap of key usage.
func (kd *KeyboardDisplay) PrintHeatmap(individual genetic.Individual, data KeyloggerDataInterface) {
	if data == nil || data.GetTotalChars() == 0 {
		fmt.Println("No usage data available for heatmap")

		return
	}

	fmt.Println("\nUsage Heatmap (darker = more frequent):")
	fmt.Println("┌─────┬─────┬─────┬─────┬─────┬─────┬─────┬─────┬─────┬─────┐")

	// Top row
	fmt.Print("│")

	for i := range 10 {
		if i < len(individual.Layout) {
			char := individual.Layout[i]
			freq := data.GetCharFreq(char)
			percent := float64(freq) * 100.0 / float64(data.GetTotalChars())
			symbol := kd.getHeatmapSymbol(percent)
			fmt.Printf(" %c%s  │", char, symbol)
		} else {
			fmt.Print("     │")
		}
	}

	fmt.Println()

	fmt.Println("├─────┼─────┼─────┼─────┼─────┼─────┼─────┼─────┼─────┼─────┤")

	// Middle row
	fmt.Print("│")

	for i := 10; i < 19; i++ {
		if i < len(individual.Layout) {
			char := individual.Layout[i]
			freq := data.GetCharFreq(char)
			percent := float64(freq) * 100.0 / float64(data.GetTotalChars())
			symbol := kd.getHeatmapSymbol(percent)
			fmt.Printf(" %c%s  │", char, symbol)
		} else {
			fmt.Print("     │")
		}
	}

	fmt.Println()

	fmt.Println("├─────┼─────┼─────┼─────┼─────┼─────┼─────┼─────┼─────┼─────┤")

	// Bottom row
	fmt.Print("│")

	for i := 19; i < 26; i++ {
		char := individual.Layout[i]
		freq := data.GetCharFreq(char)
		percent := float64(freq) * 100.0 / float64(data.GetTotalChars())
		symbol := kd.getHeatmapSymbol(percent)
		fmt.Printf(" %c%s  │", char, symbol)
	}

	fmt.Print("     │     │")
	fmt.Println()

	fmt.Println("└─────┴─────┴─────┴─────┴─────┴─────┴─────┴─────┴─────┴─────┘")

	// Legend
	fmt.Println("\nHeatmap Legend:")
	fmt.Println("  · = 0-1%    ▪ = 1-3%    ▫ = 3-6%    ■ = 6%+")
}

// getHeatmapSymbol returns appropriate symbol for frequency percentage.
func (kd *KeyboardDisplay) getHeatmapSymbol(percent float64) string {
	switch {
	case percent >= 6:
		return "■"
	case percent >= 3:
		return "▫"
	case percent >= 1:
		return "▪"
	default:
		return "·"
	}
}

// PrintLayoutString returns the layout as a simple string.
func (kd *KeyboardDisplay) PrintLayoutString(individual genetic.Individual) string {
	return string(individual.Layout)
}

// SaveLayoutImage would save the layout as an image (placeholder).
func (kd *KeyboardDisplay) SaveLayoutImage(individual genetic.Individual, filename string) error {
	// This would require an image library like "image" package
	// For now, just save as text
	return errors.New("image export not implemented yet")
}
