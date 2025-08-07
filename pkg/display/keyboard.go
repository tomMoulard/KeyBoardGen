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

// PrintStatistics shows comprehensive keyboard layout statistics.
func (kd *KeyboardDisplay) PrintStatistics(individual genetic.Individual, data KeyloggerDataInterface) {
	if data == nil {
		fmt.Printf("Fitness: %.6f, Age: %d generations\n", individual.Fitness, individual.Age)
		return
	}

	fmt.Printf("\n╔═══════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                          LAYOUT ANALYSIS                         ║\n")
	fmt.Printf("╚═══════════════════════════════════════════════════════════════════╝\n")

	// Basic layout information
	fmt.Printf("\n\033[1;36mBASIC METRICS:\033[0m\n")
	fmt.Printf("   * Character set: %s (%d characters)\n", individual.Charset.Name, individual.Charset.Size)
	fmt.Printf("   * Total keystrokes analyzed: %d\n", data.GetTotalChars())
	fmt.Printf("   * Unique bigrams: %d\n", len(data.GetAllBigrams()))
	fmt.Printf("   * Fitness score: %.6f\n", individual.Fitness)
	fmt.Printf("   * Generation age: %d\n", individual.Age)

	// Hand alternation analysis
	alternationCount := kd.calculateHandAlternation(individual, data)
	totalBigrams := data.GetTotalChars() - 1
	alternationPercent := float64(alternationCount) * 100.0 / float64(totalBigrams)

	fmt.Printf("\n\033[1;33mHAND ALTERNATION:\033[0m\n")
	fmt.Printf("   * Hand alternation rate: %.1f%% (%d/%d bigrams)\n",
		alternationPercent, alternationCount, totalBigrams)

	if alternationPercent >= 60 {
		fmt.Printf("   * \033[1;32m[EXCELLENT]\033[0m Hand balance >60%%\n")
	} else if alternationPercent >= 45 {
		fmt.Printf("   * \033[1;32m[GOOD]\033[0m Hand balance 45-60%%\n")
	} else if alternationPercent >= 30 {
		fmt.Printf("   * \033[1;33m[MODERATE]\033[0m Hand balance 30-45%%\n")
	} else {
		fmt.Printf("   * \033[1;31m[POOR]\033[0m Hand balance <30%%\n")
	}

	// Finger usage distribution with visual bars
	fingerUsage := kd.calculateFingerUsage(individual, data)
	maxUsage := 0
	for _, usage := range fingerUsage {
		if usage > maxUsage {
			maxUsage = usage
		}
	}

	fmt.Printf("\n\033[1;35mFINGER WORKLOAD DISTRIBUTION:\033[0m\n")
	fingers := []string{"L.Pinky", "L.Ring ", "L.Mid  ", "L.Index", "R.Index", "R.Mid  ", "R.Ring ", "R.Pinky"}

	for i, usage := range fingerUsage {
		percent := float64(usage) * 100.0 / float64(data.GetTotalChars())
		barLength := int(float64(usage) * 20.0 / float64(maxUsage))
		bar := strings.Repeat("█", barLength) + strings.Repeat("░", 20-barLength)
		fmt.Printf("   %s: %s %5.1f%% (%d)\n", fingers[i], bar, percent, usage)
	}

	// Check for balanced finger usage
	leftTotal := fingerUsage[0] + fingerUsage[1] + fingerUsage[2] + fingerUsage[3]
	rightTotal := fingerUsage[4] + fingerUsage[5] + fingerUsage[6] + fingerUsage[7]
	lrBalance := float64(leftTotal) * 100.0 / float64(leftTotal+rightTotal)

	fmt.Printf("\n   * Left/Right balance: %.1f%%/%.1f%%", lrBalance, 100-lrBalance)
	if lrBalance >= 45 && lrBalance <= 55 {
		fmt.Printf(" \033[1;32m[BALANCED]\033[0m\n")
	} else {
		fmt.Printf(" \033[1;33m[IMBALANCED]\033[0m\n")
	}

	// Row usage analysis
	rowUsage := kd.calculateRowUsage(individual, data)
	fmt.Printf("\n\033[1;34mROW USAGE DISTRIBUTION:\033[0m\n")
	rows := []string{"Top   ", "Home  ", "Bottom"}
	for i, usage := range rowUsage {
		percent := float64(usage) * 100.0 / float64(data.GetTotalChars())
		barLength := int(percent / 5) // Scale for 0-100%
		bar := strings.Repeat("█", barLength) + strings.Repeat("░", 20-barLength)
		fmt.Printf("   %s row: %s %5.1f%% (%d)\n", rows[i], bar, percent, usage)
	}

	if rowUsage[1] > rowUsage[0] && rowUsage[1] > rowUsage[2] {
		fmt.Printf("   * \033[1;32m[OPTIMIZED]\033[0m Home row usage maximized\n")
	} else {
		fmt.Printf("   * \033[1;33m[SUBOPTIMAL]\033[0m Home row could be better utilized\n")
	}

	// Most and least used keys with position analysis
	keyFreqs := kd.getKeyFrequencies(individual, data)

	fmt.Printf("\n\033[1;32mTOP 10 MOST USED KEYS:\033[0m\n")
	for i := 0; i < 10 && i < len(keyFreqs); i++ {
		kf := keyFreqs[i]
		percent := float64(kf.freq) * 100.0 / float64(data.GetTotalChars())
		row, finger := kd.getKeyInfo(kf.pos)
		fmt.Printf("   %2d. '%c' at pos %2d (%s, %s): %5.1f%% (%d)\n",
			i+1, kf.char, kf.pos, row, finger, percent, kf.freq)
	}

	fmt.Printf("\n\033[1;31mBOTTOM 5 LEAST USED KEYS:\033[0m\n")
	startIdx := len(keyFreqs) - 5
	if startIdx < 0 {
		startIdx = 0
	}
	for i := startIdx; i < len(keyFreqs); i++ {
		kf := keyFreqs[i]
		percent := float64(kf.freq) * 100.0 / float64(data.GetTotalChars())
		row, finger := kd.getKeyInfo(kf.pos)
		fmt.Printf("   %2d. '%c' at pos %2d (%s, %s): %5.1f%% (%d)\n",
			len(keyFreqs)-i, kf.char, kf.pos, row, finger, percent, kf.freq)
	}

	// Bigram efficiency analysis
	kd.analyzeBigramEfficiency(individual, data)
	
	// Modern ergonomic analysis
	kd.analyzeErgonomicMetrics(individual, data)
}

// calculateHandAlternation counts bigrams that alternate between hands.
func (kd *KeyboardDisplay) calculateHandAlternation(individual genetic.Individual, data KeyloggerDataInterface) int {
	alternationCount := 0

	// Get all bigrams and their frequencies
	bigrams := data.GetAllBigrams()

	for bigram, freq := range bigrams {
		if len(bigram) == 2 {
			char1, char2 := rune(bigram[0]), rune(bigram[1])
			pos1, pos2 := -1, -1

			// Find positions
			for i, char := range individual.Layout {
				if char == char1 {
					pos1 = i
				}
				if char == char2 {
					pos2 = i
				}
			}

			// If both characters found and on different hands, count as alternation
			if pos1 != -1 && pos2 != -1 && kd.isDifferentHands(pos1, pos2) {
				alternationCount += freq
			}
		}
	}

	return alternationCount
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
		if pos < len(fingerMap) {
			if finger, ok := fingerMap[pos]; ok {
				fingerUsage[finger] += data.GetCharFreq(char)
			}
		}
	}

	return fingerUsage
}

// calculateRowUsage computes usage per keyboard row.
func (kd *KeyboardDisplay) calculateRowUsage(individual genetic.Individual, data KeyloggerDataInterface) []int {
	rowUsage := make([]int, 3) // top, home, bottom

	for pos, char := range individual.Layout {
		var row int
		if pos <= 9 {
			row = 0 // top row
		} else if pos <= 18 {
			row = 1 // home row
		} else {
			row = 2 // bottom row
		}
		rowUsage[row] += data.GetCharFreq(char)
	}

	return rowUsage
}

// getKeyFrequencies returns sorted key frequencies.
func (kd *KeyboardDisplay) getKeyFrequencies(individual genetic.Individual, data KeyloggerDataInterface) []struct {
	char rune
	freq int
	pos  int
} {
	keyFreqs := make([]struct {
		char rune
		freq int
		pos  int
	}, len(individual.Layout))

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

	return keyFreqs
}

// getKeyInfo returns row and finger information for a position.
func (kd *KeyboardDisplay) getKeyInfo(pos int) (string, string) {
	var row string
	var finger string

	if pos <= 9 {
		row = "Top"
	} else if pos <= 18 {
		row = "Home"
	} else {
		row = "Bottom"
	}

	// Standard finger mapping for QWERTY-style layout
	fingerMap := map[int]string{
		// Top row: 0-9
		0: "L.Pinky", 1: "L.Ring", 2: "L.Middle", 3: "L.Index", 4: "L.Index",
		5: "R.Index", 6: "R.Index", 7: "R.Middle", 8: "R.Ring", 9: "R.Pinky",
		// Home row: 10-18
		10: "L.Pinky", 11: "L.Ring", 12: "L.Middle", 13: "L.Index", 14: "R.Index",
		15: "R.Index", 16: "R.Middle", 17: "R.Ring", 18: "R.Pinky",
		// Bottom row: 19-25
		19: "L.Pinky", 20: "L.Ring", 21: "L.Middle", 22: "L.Index", 23: "R.Index",
		24: "R.Middle", 25: "R.Ring",
	}

	if f, ok := fingerMap[pos]; ok {
		finger = f
	} else {
		finger = "Unknown"
	}

	return row, finger
}

// analyzeBigramEfficiency analyzes the most common bigrams and their efficiency.
func (kd *KeyboardDisplay) analyzeBigramEfficiency(individual genetic.Individual, data KeyloggerDataInterface) {
	fmt.Printf("\n\033[1;36mBIGRAM ANALYSIS:\033[0m\n")

	// Get all bigrams and sort by frequency
	bigrams := data.GetAllBigrams()
	type bigramInfo struct {
		bigram string
		freq   int
	}

	bigramList := make([]bigramInfo, 0, len(bigrams))
	for bigram, freq := range bigrams {
		bigramList = append(bigramList, bigramInfo{bigram, freq})
	}

	sort.Slice(bigramList, func(i, j int) bool {
		return bigramList[i].freq > bigramList[j].freq
	})

	// Analyze top 10 bigrams
	fmt.Printf("   Top 10 bigrams and their efficiency:\n")
	for i := 0; i < 10 && i < len(bigramList); i++ {
		bigram := bigramList[i]
		if len(bigram.bigram) == 2 {
			efficiency := kd.calculateBigramEfficiency(bigram.bigram, individual)
			percent := float64(bigram.freq) * 100.0 / float64(data.GetTotalChars())
			fmt.Printf("   %2d. '%s': %s (%.1f%%, %d times)\n",
				i+1, bigram.bigram, efficiency, percent, bigram.freq)
		}
	}
}

// calculateBigramEfficiency determines if a bigram is efficiently placed.
func (kd *KeyboardDisplay) calculateBigramEfficiency(bigram string, individual genetic.Individual) string {
	if len(bigram) != 2 {
		return "N/A"
	}

	char1, char2 := rune(bigram[0]), rune(bigram[1])
	pos1, pos2 := -1, -1

	// Find positions
	for i, char := range individual.Layout {
		if char == char1 {
			pos1 = i
		}
		if char == char2 {
			pos2 = i
		}
	}

	if pos1 == -1 || pos2 == -1 {
		return "\033[1;37m[UNKNOWN]\033[0m"
	}

	// Check if it's hand alternation (good)
	if kd.isDifferentHands(pos1, pos2) {
		return "\033[1;32m[ALTERNATION]\033[0m"
	}

	// Check if it's same finger (bad)
	if kd.isSameFinger(pos1, pos2) {
		return "\033[1;31m[SAME FINGER]\033[0m"
	}

	// Same hand, different fingers (okay)
	return "\033[1;33m[SAME HAND]\033[0m"
}

// analyzeErgonomicMetrics displays detailed analysis of modern typing comfort metrics.
func (kd *KeyboardDisplay) analyzeErgonomicMetrics(individual genetic.Individual, data KeyloggerDataInterface) {
	fmt.Printf("\n╔═══════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                     ERGONOMIC ANALYSIS                           ║\n")
	fmt.Printf("╚═══════════════════════════════════════════════════════════════════╝\n")

	// Calculate SFBs
	sfbData := kd.calculateSameFingerBigrams(individual, data)
	fmt.Printf("\n\033[1;31mSAME FINGER BIGRAMS (SFBs):\033[0m\n")
	fmt.Printf("   * Total SFBs: %d/%d bigrams (%.1f%%)\n", sfbData.count, sfbData.total, sfbData.rate*100)
	if sfbData.rate < 0.02 {
		fmt.Printf("   * \033[1;32m[EXCELLENT]\033[0m Very low SFB rate\n")
	} else if sfbData.rate < 0.05 {
		fmt.Printf("   * \033[1;33m[GOOD]\033[0m Acceptable SFB rate\n")
	} else {
		fmt.Printf("   * \033[1;31m[POOR]\033[0m High SFB rate - uncomfortable typing\n")
	}
	if len(sfbData.examples) > 0 {
		fmt.Printf("   * Worst SFBs: %v\n", sfbData.examples[:min(5, len(sfbData.examples))])
	}

	// Calculate LSBs  
	lsbData := kd.calculateLateralStretches(individual, data)
	fmt.Printf("\n\033[1;35mLATERAL STRETCH BIGRAMS (LSBs):\033[0m\n")
	fmt.Printf("   * Total LSBs: %d/%d bigrams (%.1f%%)\n", lsbData.count, lsbData.total, lsbData.rate*100)
	if lsbData.rate < 0.01 {
		fmt.Printf("   * \033[1;32m[EXCELLENT]\033[0m Minimal lateral stretching\n")
	} else if lsbData.rate < 0.03 {
		fmt.Printf("   * \033[1;33m[GOOD]\033[0m Low lateral stretching\n")
	} else {
		fmt.Printf("   * \033[1;31m[POOR]\033[0m High lateral stretching - strain on index fingers\n")
	}
	if len(lsbData.examples) > 0 {
		fmt.Printf("   * LSB examples: %v\n", lsbData.examples[:min(3, len(lsbData.examples))])
	}

	// Calculate rolls
	rollData := kd.calculateRollQuality(individual, data)
	fmt.Printf("\n\033[1;34mROLL QUALITY:\033[0m\n")
	fmt.Printf("   * Inward rolls: %d bigrams (%.1f%%) \033[1;32m[SMOOTH]\033[0m\n", rollData.inwardCount, rollData.inwardRate*100)
	fmt.Printf("   * Outward rolls: %d bigrams (%.1f%%) \033[1;33m[OKAY]\033[0m\n", rollData.outwardCount, rollData.outwardRate*100)
	fmt.Printf("   * Roll ratio: %.1f%% of same-hand bigrams are rolls\n", rollData.totalRollRate*100)
	if rollData.totalRollRate > 0.3 {
		fmt.Printf("   * \033[1;32m[EXCELLENT]\033[0m High roll frequency - smooth typing feel\n")
	} else if rollData.totalRollRate > 0.15 {
		fmt.Printf("   * \033[1;33m[GOOD]\033[0m Moderate roll frequency\n")
	} else {
		fmt.Printf("   * \033[1;31m[POOR]\033[0m Low roll frequency - choppy typing\n")
	}
}

// Ergonomic data structures
type SFBData struct {
	count    int
	total    int
	rate     float64
	examples []string
}

type LSBData struct {
	count    int
	total    int
	rate     float64
	examples []string
}

type RollData struct {
	inwardCount    int
	outwardCount   int
	inwardRate     float64
	outwardRate    float64
	totalRollRate  float64
}

// calculateSameFingerBigrams analyzes SFB frequency and examples.
func (kd *KeyboardDisplay) calculateSameFingerBigrams(individual genetic.Individual, data KeyloggerDataInterface) SFBData {
	sfbCount := 0
	totalBigrams := 0
	examples := make([]string, 0)
	
	bigrams := data.GetAllBigrams()
	for bigram, freq := range bigrams {
		if len(bigram) == 2 {
			totalBigrams += freq
			if kd.isSameFingerBigram(bigram, individual) {
				sfbCount += freq
				if len(examples) < 10 { // Collect examples
					examples = append(examples, bigram)
				}
			}
		}
	}
	
	rate := 0.0
	if totalBigrams > 0 {
		rate = float64(sfbCount) / float64(totalBigrams)
	}
	
	return SFBData{
		count:    sfbCount,
		total:    totalBigrams,
		rate:     rate,
		examples: examples,
	}
}

// calculateLateralStretches analyzes LSB frequency and examples.
func (kd *KeyboardDisplay) calculateLateralStretches(individual genetic.Individual, data KeyloggerDataInterface) LSBData {
	lsbCount := 0
	totalBigrams := 0
	examples := make([]string, 0)
	
	bigrams := data.GetAllBigrams()
	for bigram, freq := range bigrams {
		if len(bigram) == 2 {
			totalBigrams += freq
			if kd.isLateralStretchBigram(bigram, individual) {
				lsbCount += freq
				if len(examples) < 10 {
					examples = append(examples, bigram)
				}
			}
		}
	}
	
	rate := 0.0
	if totalBigrams > 0 {
		rate = float64(lsbCount) / float64(totalBigrams)
	}
	
	return LSBData{
		count:    lsbCount,
		total:    totalBigrams,
		rate:     rate,
		examples: examples,
	}
}

// calculateRollQuality analyzes roll frequency and quality.
func (kd *KeyboardDisplay) calculateRollQuality(individual genetic.Individual, data KeyloggerDataInterface) RollData {
	inwardCount := 0
	outwardCount := 0
	totalBigrams := 0
	
	bigrams := data.GetAllBigrams()
	for bigram, freq := range bigrams {
		if len(bigram) == 2 {
			totalBigrams += freq
			rollType := kd.getRollType(bigram, individual)
			if rollType == "inward" {
				inwardCount += freq
			} else if rollType == "outward" {
				outwardCount += freq
			}
		}
	}
	
	inwardRate := 0.0
	outwardRate := 0.0
	totalRollRate := 0.0
	
	if totalBigrams > 0 {
		inwardRate = float64(inwardCount) / float64(totalBigrams)
		outwardRate = float64(outwardCount) / float64(totalBigrams)
		totalRollRate = float64(inwardCount+outwardCount) / float64(totalBigrams)
	}
	
	return RollData{
		inwardCount:   inwardCount,
		outwardCount:  outwardCount,
		inwardRate:    inwardRate,
		outwardRate:   outwardRate,
		totalRollRate: totalRollRate,
	}
}

// Helper functions for ergonomic analysis
func (kd *KeyboardDisplay) isSameFingerBigram(bigram string, individual genetic.Individual) bool {
	if len(bigram) != 2 {
		return false
	}
	
	char1, char2 := rune(bigram[0]), rune(bigram[1])
	pos1, pos2 := -1, -1
	
	for i, char := range individual.Layout {
		if char == char1 {
			pos1 = i
		}
		if char == char2 {
			pos2 = i
		}
	}
	
	return pos1 != -1 && pos2 != -1 && kd.isSameFinger(pos1, pos2)
}

func (kd *KeyboardDisplay) isLateralStretchBigram(bigram string, individual genetic.Individual) bool {
	if len(bigram) != 2 {
		return false
	}
	
	char1, char2 := rune(bigram[0]), rune(bigram[1])
	pos1, pos2 := -1, -1
	
	for i, char := range individual.Layout {
		if char == char1 {
			pos1 = i
		}
		if char == char2 {
			pos2 = i
		}
	}
	
	if pos1 == -1 || pos2 == -1 {
		return false
	}
	
	// Check if both are on index fingers and far apart horizontally
	finger1 := kd.getFingerForPos(pos1)
	finger2 := kd.getFingerForPos(pos2)
	
	// Index fingers are 3 (left) and 4 (right)
	if (finger1 == 3 && finger2 == 3) || (finger1 == 4 && finger2 == 4) {
		// Check if on same row but far apart
		row1 := kd.getRowForPos(pos1)
		row2 := kd.getRowForPos(pos2)
		if row1 == row2 {
			col1 := kd.getColumnForPos(pos1)
			col2 := kd.getColumnForPos(pos2)
			return abs(col1-col2) > 2
		}
	}
	
	return false
}

func (kd *KeyboardDisplay) getRollType(bigram string, individual genetic.Individual) string {
	if len(bigram) != 2 {
		return "none"
	}
	
	char1, char2 := rune(bigram[0]), rune(bigram[1])
	pos1, pos2 := -1, -1
	
	for i, char := range individual.Layout {
		if char == char1 {
			pos1 = i
		}
		if char == char2 {
			pos2 = i
		}
	}
	
	if pos1 == -1 || pos2 == -1 {
		return "none"
	}
	
	finger1 := kd.getFingerForPos(pos1)
	finger2 := kd.getFingerForPos(pos2)
	row1 := kd.getRowForPos(pos1)
	row2 := kd.getRowForPos(pos2)
	
	// Check for roll: same hand, adjacent fingers, same row
	sameHand := (finger1 < 4 && finger2 < 4) || (finger1 >= 4 && finger2 >= 4)
	adjacentFingers := abs(finger1-finger2) == 1
	sameRow := row1 == row2
	
	if sameHand && adjacentFingers && sameRow {
		if finger1 < 4 && finger2 < 4 { // Left hand
			if finger1 < finger2 {
				return "inward" // Pinky to index
			}
			return "outward" // Index to pinky
		} else if finger1 >= 4 && finger2 >= 4 { // Right hand
			if finger1 > finger2 {
				return "inward" // Index to pinky
			}
			return "outward" // Pinky to index
		}
	}
	
	return "none"
}

// Helper functions for position analysis
func (kd *KeyboardDisplay) getFingerForPos(pos int) int {
	fingerMap := map[int]int{
		0: 0, 1: 1, 2: 2, 3: 3, 4: 3, 5: 4, 6: 4, 7: 5, 8: 6, 9: 7,
		10: 0, 11: 1, 12: 2, 13: 3, 14: 4, 15: 4, 16: 5, 17: 6, 18: 7,
		19: 0, 20: 1, 21: 2, 22: 3, 23: 4, 24: 5, 25: 6,
	}
	if finger, ok := fingerMap[pos]; ok {
		return finger
	}
	return -1
}

func (kd *KeyboardDisplay) getRowForPos(pos int) int {
	if pos <= 9 {
		return 0 // top row
	} else if pos <= 18 {
		return 1 // home row
	} else {
		return 2 // bottom row
	}
}

func (kd *KeyboardDisplay) getColumnForPos(pos int) int {
	if pos <= 9 {
		return pos
	} else if pos <= 18 {
		return pos - 10
	} else {
		return pos - 19
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// isDifferentHands checks if two positions are on different hands.
func (kd *KeyboardDisplay) isDifferentHands(pos1, pos2 int) bool {
	// Positions 0-4, 10-14, 19-22 are left hand
	// Positions 5-9, 15-18, 23-25 are right hand
	leftHand1 := (pos1 <= 4) || (pos1 >= 10 && pos1 <= 14) || (pos1 >= 19 && pos1 <= 22)
	leftHand2 := (pos2 <= 4) || (pos2 >= 10 && pos2 <= 14) || (pos2 >= 19 && pos2 <= 22)
	return leftHand1 != leftHand2
}

// isSameFinger checks if two positions use the same finger.
func (kd *KeyboardDisplay) isSameFinger(pos1, pos2 int) bool {
	fingerMap := map[int]int{
		0: 0, 1: 1, 2: 2, 3: 3, 4: 3, 5: 4, 6: 4, 7: 5, 8: 6, 9: 7,
		10: 0, 11: 1, 12: 2, 13: 3, 14: 4, 15: 4, 16: 5, 17: 6, 18: 7,
		19: 0, 20: 1, 21: 2, 22: 3, 23: 4, 24: 5, 25: 6,
	}

	finger1, ok1 := fingerMap[pos1]
	finger2, ok2 := fingerMap[pos2]
	return ok1 && ok2 && finger1 == finger2
}

// PrintComparison shows a side-by-side comparison with QWERTY.
func (kd *KeyboardDisplay) PrintComparison(individual genetic.Individual, data KeyloggerDataInterface) {
	kd.PrintComparisonWithEvaluator(individual, data, nil)
}

// PrintComparisonWithEvaluator shows enhanced comparison with actual QWERTY fitness calculation.
func (kd *KeyboardDisplay) PrintComparisonWithEvaluator(individual genetic.Individual, data KeyloggerDataInterface, evaluator FitnessEvaluator) {
	qwerty := [26]rune{
		'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', // Top row
		'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', // Middle row
		'z', 'x', 'c', 'v', 'b', 'n', 'm', // Bottom row
	}

	fmt.Printf("\n╔═══════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                      LAYOUT COMPARISON                           ║\n")
	fmt.Printf("╚═══════════════════════════════════════════════════════════════════╝\n")

	// Calculate QWERTY fitness if evaluator is provided
	var qwertyFitness float64
	if evaluator != nil {
		qwertyFitness = evaluator.EvaluateLegacy(qwerty, data)
	} else {
		qwertyFitness = 0.4 // Estimated baseline
	}

	improvementPercent := ((individual.Fitness - qwertyFitness) / qwertyFitness) * 100

	fmt.Printf("\n\033[1;36mFITNESS COMPARISON:\033[0m\n")
	fmt.Printf("   * Optimized layout: %.6f\n", individual.Fitness)
	fmt.Printf("   * QWERTY layout:    %.6f\n", qwertyFitness)
	if improvementPercent > 0 {
		fmt.Printf("   * Improvement:      \033[1;32m+%.1f%%\033[0m\n", improvementPercent)
	} else {
		fmt.Printf("   * Difference:       \033[1;33m%.1f%%\033[0m\n", improvementPercent)
	}

	fmt.Printf("\n\033[1;34mVISUAL LAYOUT COMPARISON:\033[0m\n")
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

	// Detailed comparison metrics
	if evaluator != nil {
		kd.printDetailedComparison(individual, qwerty, data, evaluator)
	}

	fmt.Printf("\n\033[1;37mOPTIMIZATION INSIGHTS:\033[0m\n")
	if improvementPercent > 20 {
		fmt.Printf("   * \033[1;32mEXCELLENT!\033[0m Your typing efficiency improved significantly.\n")
	} else if improvementPercent > 10 {
		fmt.Printf("   * \033[1;32mGOOD!\033[0m Noticeable improvement in typing efficiency.\n")
	} else if improvementPercent > 0 {
		fmt.Printf("   * \033[1;33mMODEST\033[0m improvement over QWERTY.\n")
	} else {
		fmt.Printf("   * \033[1;33mNEEDS WORK\033[0m - May need more data or different parameters.\n")
	}

	fmt.Printf("   * Consider using this layout for %s typing tasks.\n", individual.Charset.Name)
	if individual.Charset.Name != "alphabet" {
		fmt.Printf("   * This layout is optimized for programming/special characters.\n")
	}
}

// printDetailedComparison shows detailed metrics comparison.
func (kd *KeyboardDisplay) printDetailedComparison(individual genetic.Individual, qwerty [26]rune, data KeyloggerDataInterface, evaluator FitnessEvaluator) {
	// Create QWERTY individual for analysis
	qwertyIndividual := genetic.Individual{
		Layout:  make([]rune, 26),
		Charset: individual.Charset,
		Fitness: evaluator.EvaluateLegacy(qwerty, data),
	}
	copy(qwertyIndividual.Layout, qwerty[:])

	fmt.Printf("\n\033[1;35mDETAILED METRICS COMPARISON:\033[0m\n")

	// Hand alternation comparison
	optAlternation := kd.calculateHandAlternation(individual, data)
	qwertyAlternation := kd.calculateHandAlternation(qwertyIndividual, data)
	totalBigrams := data.GetTotalChars() - 1

	optAltPercent := float64(optAlternation) * 100.0 / float64(totalBigrams)
	qwertyAltPercent := float64(qwertyAlternation) * 100.0 / float64(totalBigrams)

	fmt.Printf("   Hand alternation:  %.1f%% vs %.1f%% (QWERTY)", optAltPercent, qwertyAltPercent)
	if optAltPercent > qwertyAltPercent {
		fmt.Printf(" \033[1;32m[BETTER]\033[0m\n")
	} else {
		fmt.Printf(" \033[1;33m[WORSE]\033[0m\n")
	}

	// Row usage comparison
	optRows := kd.calculateRowUsage(individual, data)
	qwertyRows := kd.calculateRowUsage(qwertyIndividual, data)

	optHomePercent := float64(optRows[1]) * 100.0 / float64(data.GetTotalChars())
	qwertyHomePercent := float64(qwertyRows[1]) * 100.0 / float64(data.GetTotalChars())

	fmt.Printf("   Home row usage:    %.1f%% vs %.1f%% (QWERTY)", optHomePercent, qwertyHomePercent)
	if optHomePercent > qwertyHomePercent {
		fmt.Printf(" \033[1;32m[BETTER]\033[0m\n")
	} else {
		fmt.Printf(" \033[1;33m[WORSE]\033[0m\n")
	}

	// Finger balance comparison
	optFingers := kd.calculateFingerUsage(individual, data)
	qwertyFingers := kd.calculateFingerUsage(qwertyIndividual, data)

	// Calculate finger load variance (lower is better)
	optVariance := kd.calculateFingerVariance(optFingers, data.GetTotalChars())
	qwertyVariance := kd.calculateFingerVariance(qwertyFingers, data.GetTotalChars())

	fmt.Printf("   Finger balance:    %.3f vs %.3f variance (QWERTY)", optVariance, qwertyVariance)
	if optVariance < qwertyVariance {
		fmt.Printf(" \033[1;32m[BETTER]\033[0m\n")
	} else {
		fmt.Printf(" \033[1;33m[WORSE]\033[0m\n")
	}

	// Add ergonomic comparison
	kd.printErgonomicComparison(individual, qwertyIndividual, data)
}

// printErgonomicComparison shows modern ergonomic metrics comparison.
func (kd *KeyboardDisplay) printErgonomicComparison(individual genetic.Individual, qwertyIndividual genetic.Individual, data KeyloggerDataInterface) {
	fmt.Printf("\n\033[1;36mMODERN ERGONOMIC COMPARISON:\033[0m\n")

	// SFB comparison
	optSFB := kd.calculateSameFingerBigrams(individual, data)
	qwertySFB := kd.calculateSameFingerBigrams(qwertyIndividual, data)

	fmt.Printf("   Same Finger Bigrams: %.1f%% vs %.1f%% (QWERTY)", optSFB.rate*100, qwertySFB.rate*100)
	if optSFB.rate < qwertySFB.rate {
		improvement := (qwertySFB.rate - optSFB.rate) / qwertySFB.rate * 100
		fmt.Printf(" \033[1;32m[%.1f%% BETTER]\033[0m\n", improvement)
	} else {
		fmt.Printf(" \033[1;33m[WORSE]\033[0m\n")
	}

	// LSB comparison  
	optLSB := kd.calculateLateralStretches(individual, data)
	qwertyLSB := kd.calculateLateralStretches(qwertyIndividual, data)

	fmt.Printf("   Lateral Stretches:   %.1f%% vs %.1f%% (QWERTY)", optLSB.rate*100, qwertyLSB.rate*100)
	if optLSB.rate < qwertyLSB.rate {
		improvement := (qwertyLSB.rate - optLSB.rate) / qwertyLSB.rate * 100
		fmt.Printf(" \033[1;32m[%.1f%% BETTER]\033[0m\n", improvement)
	} else {
		fmt.Printf(" \033[1;33m[WORSE]\033[0m\n")
	}

	// Roll comparison
	optRoll := kd.calculateRollQuality(individual, data)
	qwertyRoll := kd.calculateRollQuality(qwertyIndividual, data)

	fmt.Printf("   Roll Quality:        %.1f%% vs %.1f%% (QWERTY)", optRoll.totalRollRate*100, qwertyRoll.totalRollRate*100)
	if optRoll.totalRollRate > qwertyRoll.totalRollRate {
		improvement := (optRoll.totalRollRate - qwertyRoll.totalRollRate) / qwertyRoll.totalRollRate * 100
		fmt.Printf(" \033[1;32m[%.1f%% BETTER]\033[0m\n", improvement)
	} else {
		fmt.Printf(" \033[1;33m[WORSE]\033[0m\n")
	}
}

// calculateFingerVariance calculates the variance in finger usage (lower is better).
func (kd *KeyboardDisplay) calculateFingerVariance(fingerUsage []int, totalChars int) float64 {
	if len(fingerUsage) == 0 {
		return 0
	}

	// Calculate mean
	mean := float64(totalChars) / float64(len(fingerUsage))

	// Calculate variance
	variance := 0.0
	for _, usage := range fingerUsage {
		diff := float64(usage) - mean
		variance += diff * diff
	}
	variance /= float64(len(fingerUsage))

	return variance
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

// PrintSummary prints a concise summary of the optimization results.
func (kd *KeyboardDisplay) PrintSummary(individual genetic.Individual, data KeyloggerDataInterface, evaluator FitnessEvaluator) {
	fmt.Printf("\n╔═══════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                        OPTIMIZATION SUMMARY                      ║\n")
	fmt.Printf("╚═══════════════════════════════════════════════════════════════════╝\n")

	alternationCount := kd.calculateHandAlternation(individual, data)
	totalBigrams := data.GetTotalChars() - 1
	alternationPercent := float64(alternationCount) * 100.0 / float64(totalBigrams)

	rowUsage := kd.calculateRowUsage(individual, data)
	homeRowPercent := float64(rowUsage[1]) * 100.0 / float64(data.GetTotalChars())

	fmt.Printf("\n\033[1;36mYour optimized %s keyboard layout:\033[0m\n", individual.Charset.Name)
	fmt.Printf("   * Fitness Score:     %.6f\n", individual.Fitness)
	fmt.Printf("   * Hand Alternation:  %.1f%%\n", alternationPercent)
	fmt.Printf("   * Home Row Usage:    %.1f%%\n", homeRowPercent)
	fmt.Printf("   * Training Data:     %d keystrokes\n", data.GetTotalChars())
	fmt.Printf("   * Evolution Age:     %d generations\n", individual.Age)

	if evaluator != nil {
		qwerty := [26]rune{
			'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p',
			'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l',
			'z', 'x', 'c', 'v', 'b', 'n', 'm',
		}
		qwertyFitness := evaluator.EvaluateLegacy(qwerty, data)
		improvementPercent := ((individual.Fitness - qwertyFitness) / qwertyFitness) * 100
		fmt.Printf("   * QWERTY Improvement: %.1f%%\n", improvementPercent)
	}

	fmt.Printf("\n\033[1;37mNext steps:\033[0m\n")
	fmt.Printf("   1. Try this layout for your typical typing tasks\n")
	fmt.Printf("   2. Consider a gradual transition if improvement is significant\n")
	fmt.Printf("   3. Collect more typing data for further refinement\n")
}
