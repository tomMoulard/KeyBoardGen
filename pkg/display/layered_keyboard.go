package display

import (
	"fmt"
	"strings"

	"github.com/tommoulard/keyboardgen/pkg/fitness"
)

// LayeredKeyboardDisplay handles visualization of keyboards with layer support.
type LayeredKeyboardDisplay struct {
	showFrequency bool
	showColors    bool
	compact       bool
	Layout        *fitness.KeyboardLayout // Exported for access
}

// NewLayeredKeyboardDisplay creates a new layered display handler.
func NewLayeredKeyboardDisplay(layout *fitness.KeyboardLayout) *LayeredKeyboardDisplay {
	return &LayeredKeyboardDisplay{
		showFrequency: false,
		showColors:    false,
		compact:       false,
		Layout:        layout,
	}
}

// SetOptions configures display options.
func (lkd *LayeredKeyboardDisplay) SetOptions(showFreq, showColors, compact bool) {
	lkd.showFrequency = showFreq
	lkd.showColors = showColors
	lkd.compact = compact
}

// PrintLayeredLayout displays all keyboard layers (base, shift, AltGr).
func (lkd *LayeredKeyboardDisplay) PrintLayeredLayout(data KeyloggerDataInterface) {
	fmt.Printf("\n╔═══════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                    %s KEYBOARD LAYERS                     ║\n", strings.ToUpper(lkd.Layout.Name))
	fmt.Printf("╚═══════════════════════════════════════════════════════════════════╝\n")

	if lkd.compact {
		lkd.printCompactLayeredLayout(data)
	} else {
		lkd.printFullLayeredLayout(data)
	}
}

// printFullLayeredLayout displays the full graphical keyboard with all layers.
func (lkd *LayeredKeyboardDisplay) printFullLayeredLayout(data KeyloggerDataInterface) {
	// Get the total number of keys in the layout
	maxPos := 0
	for pos := range lkd.Layout.Keys {
		if pos > maxPos {
			maxPos = pos
		}
	}

	// Define keyboard rows for full 70-character layout
	// Arrange in multiple rows to accommodate all characters
	keysPerRow := 10

	rows := make([]struct {
		name      string
		positions []int
	}, 0)

	for startPos := 0; startPos <= maxPos; startPos += keysPerRow {
		endPos := startPos + keysPerRow
		if endPos > maxPos+1 {
			endPos = maxPos + 1
		}

		positions := make([]int, 0)
		for pos := startPos; pos < endPos; pos++ {
			positions = append(positions, pos)
		}

		if len(positions) > 0 {
			rowName := fmt.Sprintf("ROW %d-%d ", startPos, endPos-1)
			if startPos < 10 {
				rowName = "TOP ROW   "
			} else if startPos < 20 {
				rowName = "HOME ROW  "
			} else if startPos < 30 {
				rowName = "BOTTOM ROW"
			}

			rows = append(rows, struct {
				name      string
				positions []int
			}{rowName, positions})
		}
	}

	// Print each layer completely, one after the other
	layers := []struct {
		name  string
		layer fitness.KeyLayer
	}{
		{"BASE LAYER", fitness.BaseLayer},
		{"SHIFT LAYER", fitness.ShiftLayer},
		{"ALTGR LAYER", fitness.AltGrLayer},
	}

	for _, layerInfo := range layers {
		// Check if this layer has any characters (skip empty AltGr if no characters)
		hasChars := false

		if layerInfo.layer == fitness.AltGrLayer {
			for pos := 0; pos <= maxPos; pos++ {
				if lkd.getCharForLayer(pos, layerInfo.layer) != 0 {
					hasChars = true

					break
				}
			}

			if !hasChars {
				continue // Skip empty AltGr layer
			}
		} else {
			hasChars = true // Base and Shift always have characters
		}

		fmt.Printf("\n\033[1;36m%s:\033[0m\n", layerInfo.name)

		for rowIndex, row := range rows {
			// Build dynamic box borders
			boxTop := "┌────────────"
			boxMid := "├────────────"
			boxBot := "└────────────"

			for range len(row.positions) {
				boxTop += "┬─────────"
				boxMid += "┼─────────"
				boxBot += "┴─────────"
			}

			boxTop += "┐"
			boxMid += "┤"
			boxBot += "┘"

			// Print top border (only for first row of each layer)
			if rowIndex == 0 {
				fmt.Println(boxTop)
			} else {
				fmt.Println(boxMid)
			}

			fmt.Printf("│ %-7s │", row.name)

			for _, pos := range row.positions {
				cell := lkd.formatLayerCell(pos, layerInfo.layer, data)
				fmt.Printf("%s│", cell)
			}

			fmt.Println()

			// Print bottom border (only for last row of each layer)
			if rowIndex == len(rows)-1 {
				fmt.Println(boxBot)
			}
		}
	}
}

// printCompactLayeredLayout displays a compact version with all layers.
func (lkd *LayeredKeyboardDisplay) printCompactLayeredLayout(data KeyloggerDataInterface) {
	// Get the total number of keys in the layout
	maxPos := 0
	for pos := range lkd.Layout.Keys {
		if pos > maxPos {
			maxPos = pos
		}
	}

	// Define keyboard rows for full 70-character layout
	keysPerRow := 15 // More keys per row for compact display

	for startPos := 0; startPos <= maxPos; startPos += keysPerRow {
		endPos := startPos + keysPerRow
		if endPos > maxPos+1 {
			endPos = maxPos + 1
		}

		positions := make([]int, 0)
		for pos := startPos; pos < endPos; pos++ {
			positions = append(positions, pos)
		}

		if len(positions) > 0 {
			rowName := fmt.Sprintf("ROW %d-%d", startPos, endPos-1)
			if startPos < 15 {
				rowName = "TOP"
			} else if startPos < 30 {
				rowName = "HOME"
			} else if startPos < 45 {
				rowName = "BOTTOM"
			}

			fmt.Printf("\033[1;36m%s:\033[0m\n", rowName)

			// Base layer
			fmt.Print("Base:  ")

			for _, pos := range positions {
				char := lkd.getCharForLayer(pos, fitness.BaseLayer)
				if char != 0 {
					fmt.Printf("%c ", char)
				} else {
					fmt.Print("  ")
				}
			}

			fmt.Println()

			// Shift layer
			fmt.Print("Shift: ")

			for _, pos := range positions {
				char := lkd.getCharForLayer(pos, fitness.ShiftLayer)
				if char != 0 {
					fmt.Printf("%c ", char)
				} else {
					fmt.Print("  ")
				}
			}

			fmt.Println()

			// AltGr layer (if exists)
			if lkd.hasAltGrInRow(positions) {
				fmt.Print("AltGr: ")

				for _, pos := range positions {
					char := lkd.getCharForLayer(pos, fitness.AltGrLayer)
					if char != 0 {
						fmt.Printf("%c ", char)
					} else {
						fmt.Print("  ")
					}
				}

				fmt.Println()
			}

			fmt.Println()
		}
	}
}

// formatLayerCell formats a single keyboard cell for a specific layer.
func (lkd *LayeredKeyboardDisplay) formatLayerCell(pos int, layer fitness.KeyLayer, data KeyloggerDataInterface) string {
	char := lkd.getCharForLayer(pos, layer)

	// Handle empty/null characters
	if char == 0 {
		return "         "
	}

	if data == nil {
		return fmt.Sprintf("    %c    ", char)
	}

	freq := data.GetCharFreq(char)

	if lkd.showFrequency && data.GetTotalChars() > 0 {
		percent := float64(freq) * 100.0 / float64(data.GetTotalChars())
		if percent >= 10 {
			return fmt.Sprintf("  %c %.0f%%  ", char, percent)
		} else if percent >= 1 {
			return fmt.Sprintf(" %c %.1f%%  ", char, percent)
		} else if percent > 0 {
			return fmt.Sprintf("  %c %.2f ", char, percent)
		} else {
			return fmt.Sprintf("    %c    ", char)
		}
	}

	if lkd.showColors {
		return lkd.colorizeLayerCell(char, freq, data.GetTotalChars())
	}

	return fmt.Sprintf("    %c    ", char)
}

// colorizeLayerCell applies color coding based on frequency.
func (lkd *LayeredKeyboardDisplay) colorizeLayerCell(char rune, freq, total int) string {
	if total == 0 {
		return fmt.Sprintf("    %c    ", char)
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
		return fmt.Sprintf("%s   %c   %s", colorCode, char, reset)
	}

	return fmt.Sprintf("    %c    ", char)
}

// getCharForLayer returns the character for a specific position and layer.
func (lkd *LayeredKeyboardDisplay) getCharForLayer(pos int, layer fitness.KeyLayer) rune {
	if layeredKey, exists := lkd.Layout.Keys[pos]; exists {
		switch layer {
		case fitness.BaseLayer:
			return layeredKey.BaseChar
		case fitness.ShiftLayer:
			return layeredKey.ShiftChar
		case fitness.AltGrLayer:
			if layeredKey.AltGrChar != nil {
				return *layeredKey.AltGrChar
			}

			return 0
		default:
			return 0
		}
	}

	return 0
}

// hasAltGrInRow checks if any keys in the row have AltGr characters.
func (lkd *LayeredKeyboardDisplay) hasAltGrInRow(positions []int) bool {
	for _, pos := range positions {
		if layeredKey, exists := lkd.Layout.Keys[pos]; exists {
			if layeredKey.AltGrChar != nil {
				return true
			}
		}
	}

	return false
}

// PrintLayerStatistics shows detailed statistics about layer usage.
func (lkd *LayeredKeyboardDisplay) PrintLayerStatistics(data KeyloggerDataInterface) {
	fmt.Printf("\n╔═══════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                       LAYER USAGE ANALYSIS                       ║\n")
	fmt.Printf("╚═══════════════════════════════════════════════════════════════════╝\n")

	if data == nil {
		fmt.Println("No data available for layer analysis")

		return
	}

	// Count usage by layer
	layerUsage := make(map[fitness.KeyLayer]int)
	layerUsage[fitness.BaseLayer] = 0
	layerUsage[fitness.ShiftLayer] = 0
	layerUsage[fitness.AltGrLayer] = 0

	totalChars := 0

	// Iterate through all characters in the layout
	for _, layeredKey := range lkd.Layout.Keys {
		// Base layer
		freq := data.GetCharFreq(layeredKey.BaseChar)
		layerUsage[fitness.BaseLayer] += freq
		totalChars += freq

		// Shift layer
		freq = data.GetCharFreq(layeredKey.ShiftChar)
		layerUsage[fitness.ShiftLayer] += freq
		totalChars += freq

		// AltGr layer
		if layeredKey.AltGrChar != nil {
			freq = data.GetCharFreq(*layeredKey.AltGrChar)
			layerUsage[fitness.AltGrLayer] += freq
			totalChars += freq
		}
	}

	if totalChars == 0 {
		fmt.Println("No character usage data found")

		return
	}

	fmt.Printf("\n\033[1;32mLAYER USAGE DISTRIBUTION:\033[0m\n")

	layers := []struct {
		layer fitness.KeyLayer
		name  string
	}{
		{fitness.BaseLayer, "Base Layer (no modifiers)"},
		{fitness.ShiftLayer, "Shift Layer"},
		{fitness.AltGrLayer, "AltGr Layer"},
	}

	for _, l := range layers {
		usage := layerUsage[l.layer]
		percent := float64(usage) * 100.0 / float64(totalChars)
		barLength := int(percent / 5) // Scale for 0-100%
		bar := strings.Repeat("█", barLength) + strings.Repeat("░", 20-barLength)

		fmt.Printf("   %s: %s %5.1f%% (%d keystrokes)\n",
			l.name, bar, percent, usage)
	}

	// Calculate efficiency metrics
	basePercent := float64(layerUsage[fitness.BaseLayer]) * 100.0 / float64(totalChars)
	modifierPercent := 100.0 - basePercent

	fmt.Printf("\n\033[1;33mEFFICIENCY METRICS:\033[0m\n")
	fmt.Printf("   * Base layer usage: %.1f%% (no modifier keys needed)\n", basePercent)
	fmt.Printf("   * Modifier usage: %.1f%% (requires Shift/AltGr)\n", modifierPercent)

	if basePercent >= 70 {
		fmt.Printf("   * \033[1;32m[EXCELLENT]\033[0m High base layer usage - efficient typing\n")
	} else if basePercent >= 50 {
		fmt.Printf("   * \033[1;33m[GOOD]\033[0m Moderate base layer usage\n")
	} else {
		fmt.Printf("   * \033[1;31m[POOR]\033[0m Low base layer usage - many modifier keys needed\n")
	}
}

// PrintLayerComparison compares layer usage between different keyboard layouts.
func (lkd *LayeredKeyboardDisplay) PrintLayerComparison(otherLayout *fitness.KeyboardLayout, data KeyloggerDataInterface) {
	fmt.Printf("\n╔═══════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║              LAYER USAGE COMPARISON: %s vs %s              ║\n",
		strings.ToUpper(lkd.Layout.Name), strings.ToUpper(otherLayout.Name))
	fmt.Printf("╚═══════════════════════════════════════════════════════════════════╝\n")

	if data == nil {
		fmt.Println("No data available for comparison")

		return
	}

	// Compare specific characters that differ between layouts
	fmt.Printf("\n\033[1;36mCHARACTER ACCESS COMPARISON:\033[0m\n")
	fmt.Println("Character | This Layout | Other Layout | Advantage")
	fmt.Println("----------|-------------|--------------|----------")

	// Test common characters that often differ between layouts
	testChars := []rune{
		'(', ')', '[', ']', '{', '}', // Brackets
		'1', '2', '3', '4', '5', // Numbers
		'!', '@', '#', '$', '%', // Symbols
		'&', '*', '+', '=', '-', // Operators
	}

	for _, char := range testChars {
		if data.GetCharFreq(char) > 0 { // Only show characters that are actually used
			_, thisLayer, thisCost := lkd.Layout.GetCharacterLayer(char)
			_, otherLayer, otherCost := otherLayout.GetCharacterLayer(char)

			thisLayerName := lkd.getLayerName(thisLayer)
			otherLayerName := lkd.getLayerName(otherLayer)

			advantage := "Equal"
			if thisCost < otherCost {
				advantage = lkd.Layout.Name
			} else if otherCost < thisCost {
				advantage = otherLayout.Name
			}

			freq := data.GetCharFreq(char)
			percent := float64(freq) * 100.0 / float64(data.GetTotalChars())

			fmt.Printf("    %c     |   %s    |   %s    | %s (%.2f%%)\n",
				char, thisLayerName, otherLayerName, advantage, percent)
		}
	}
}

// getLayerName returns a human-readable layer name.
func (lkd *LayeredKeyboardDisplay) getLayerName(layer fitness.KeyLayer) string {
	switch layer {
	case fitness.BaseLayer:
		return "Base "
	case fitness.ShiftLayer:
		return "Shift"
	case fitness.AltGrLayer:
		return "AltGr"
	default:
		return "???"
	}
}

// PrintLayoutLegend shows a legend explaining the layer system.
func (lkd *LayeredKeyboardDisplay) PrintLayoutLegend() {
	fmt.Printf("\n╔═══════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                        KEYBOARD LAYERS LEGEND                    ║\n")
	fmt.Printf("╚═══════════════════════════════════════════════════════════════════╝\n")

	fmt.Printf("\n\033[1;37mLayer Types:\033[0m\n")
	fmt.Printf("   🔹 \033[1;32mBase Layer\033[0m:   Characters typed without modifier keys\n")
	fmt.Printf("   🔹 \033[1;33mShift Layer\033[0m:  Characters typed while holding Shift\n")
	fmt.Printf("   🔹 \033[1;35mAltGr Layer\033[0m:  Characters typed while holding AltGr (Right Alt)\n")

	fmt.Printf("\n\033[1;37mTyping Cost:\033[0m\n")
	fmt.Printf("   • Base Layer:  Cost 1.0 (most efficient)\n")
	fmt.Printf("   • Shift Layer: Cost 1.5 (moderate effort)\n")
	fmt.Printf("   • AltGr Layer: Cost 2.0 (highest effort)\n")

	fmt.Printf("\n\033[1;37mLayout Differences:\033[0m\n")
	fmt.Printf("   • QWERTY: Numbers on base layer, symbols on shift\n")
	fmt.Printf("   • AZERTY: Some symbols on base layer, numbers on shift\n")
	fmt.Printf("   • Different layouts optimize for different character usage patterns\n")
}
