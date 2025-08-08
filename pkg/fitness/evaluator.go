package fitness

import (
	"math"

	"github.com/tommoulard/keyboardgen/pkg/genetic"
)

// KeyboardGeometry defines the physical layout of keys.
type KeyboardGeometry struct {
	// Position coordinates for each key (x, y)
	KeyPositions map[int][2]float64
	// Finger assignments for each position
	FingerMap map[int]int
}

// Standard QWERTY geometry with finger assignments.
func StandardGeometry() KeyboardGeometry {
	return KeyboardGeometry{
		KeyPositions: map[int][2]float64{
			// Top row (Q-P)
			0: {0, 0}, 1: {1, 0}, 2: {2, 0}, 3: {3, 0}, 4: {4, 0},
			5: {5, 0}, 6: {6, 0}, 7: {7, 0}, 8: {8, 0}, 9: {9, 0},
			// Middle row (A-L)
			10: {0.5, 1}, 11: {1.5, 1}, 12: {2.5, 1}, 13: {3.5, 1}, 14: {4.5, 1},
			15: {5.5, 1}, 16: {6.5, 1}, 17: {7.5, 1}, 18: {8.5, 1},
			// Bottom row (Z-M)
			19: {1, 2}, 20: {2, 2}, 21: {3, 2}, 22: {4, 2}, 23: {5, 2}, 24: {6, 2}, 25: {7, 2},
		},
		FingerMap: map[int]int{
			// Left hand: 0-3 (pinky to index), Right hand: 4-7 (index to pinky)
			0: 0, 1: 1, 2: 2, 3: 3, 4: 3, 5: 4, 6: 4, 7: 5, 8: 6, 9: 7,
			10: 0, 11: 1, 12: 2, 13: 3, 14: 4, 15: 4, 16: 5, 17: 6, 18: 7,
			19: 0, 20: 1, 21: 2, 22: 3, 23: 4, 24: 5, 25: 6,
		},
	}
}

// FitnessEvaluator calculates fitness scores for keyboard layouts.
type FitnessEvaluator struct {
	geometry KeyboardGeometry
	weights  FitnessWeights
}

// FitnessWeights defines importance of different fitness components.
type FitnessWeights struct {
	FingerDistance    float64 `json:"finger_distance"`
	HandAlternation   float64 `json:"hand_alternation"`
	FingerBalance     float64 `json:"finger_balance"`
	RowJumping        float64 `json:"row_jumping"`
	BigramEfficiency  float64 `json:"bigram_efficiency"`
	SameFingerBigrams float64 `json:"same_finger_bigrams"` // New: penalize SFBs heavily
	LateralStretches  float64 `json:"lateral_stretches"`   // New: penalize LSBs
	RollQuality       float64 `json:"roll_quality"`        // New: reward good rolls
	LayerPenalty      float64 `json:"layer_penalty"`       // New: penalize modifier key usage
	// Enhanced KPI-based components
	HomeRowBonus     float64 `json:"home_row_bonus"`    // New: reward home row usage
	RollRatioTarget  float64 `json:"roll_ratio_target"` // New: optimize roll percentage
	ThresholdBonuses float64 `json:"threshold_bonuses"` // New: bonus for crossing thresholds
	PositionMatching float64 `json:"position_matching"` // New: match frequency to ergonomics
}

// DefaultWeights returns balanced fitness weights.
func DefaultWeights() FitnessWeights {
	return FitnessWeights{
		// Core components (reduced to make room for new KPI components)
		FingerDistance:    0.10, // Reduced to make room for new components
		HandAlternation:   0.10, // Reduced
		FingerBalance:     0.10, // Reduced
		RowJumping:        0.06, // Reduced
		BigramEfficiency:  0.06, // Reduced
		SameFingerBigrams: 0.20, // High weight - SFBs are very bad
		LateralStretches:  0.04, // Moderate weight for LSBs
		RollQuality:       0.04, // Moderate weight for rolls
		LayerPenalty:      0.10, // Significant weight for layer penalties
		// Enhanced KPI components
		HomeRowBonus:     0.08, // Reward home row optimization
		RollRatioTarget:  0.04, // Target roll percentage optimization
		ThresholdBonuses: 0.04, // Bonus rewards for crossing key thresholds
		PositionMatching: 0.04, // Match high-frequency chars to ergonomic positions
	}
}

// NewFitnessEvaluator creates a new fitness evaluator.
func NewFitnessEvaluator(geometry KeyboardGeometry, weights FitnessWeights) *FitnessEvaluator {
	return &FitnessEvaluator{
		geometry: geometry,
		weights:  weights,
	}
}

// Evaluate calculates the fitness score for a keyboard layout.
func (fe *FitnessEvaluator) Evaluate(layout []rune, charset *genetic.CharacterSet, data genetic.KeyloggerDataInterface) float64 {
	// Validate inputs
	if charset == nil {
		return 0.0
	}

	// Validate layout using character set
	if !charset.IsValid(layout) {
		return 0.0 // Invalid layout gets worst possible fitness
	}

	// Create position mapping: char -> position
	charToPos := make(map[rune]int)
	for pos, char := range layout {
		charToPos[char] = pos
	}

	// Calculate individual fitness components
	distanceScore := fe.calculateFingerDistance(layout, charset, data, charToPos)
	alternationScore := fe.calculateHandAlternation(layout, charset, data, charToPos)
	balanceScore := fe.calculateFingerBalance(layout, charset, data, charToPos)
	rowJumpScore := fe.calculateRowJumping(layout, charset, data, charToPos)
	bigramScore := fe.calculateBigramEfficiency(layout, charset, data, charToPos)

	// New modern fitness components
	sfbPenalty := fe.calculateSameFingerBigrams(layout, charset, data, charToPos)
	lsbPenalty := fe.calculateLateralStretches(layout, charset, data, charToPos)
	rollScore := fe.calculateRollQuality(layout, charset, data, charToPos)
	layerPenalty := fe.calculateLayerPenalty(layout, charset, data, charToPos)

	// Enhanced KPI-based components
	homeRowScore := fe.calculateHomeRowBonus(layout, charset, data, charToPos)
	rollRatioScore := fe.calculateRollRatioTarget(layout, charset, data, charToPos)
	thresholdBonus := fe.calculateThresholdBonuses(layout, charset, data, charToPos, alternationScore, rollScore)
	positionScore := fe.calculatePositionMatching(layout, charset, data, charToPos)

	// Weighted sum of all components (higher is better)
	fitness := fe.weights.FingerDistance*distanceScore +
		fe.weights.HandAlternation*alternationScore +
		fe.weights.FingerBalance*balanceScore +
		fe.weights.RowJumping*rowJumpScore +
		fe.weights.BigramEfficiency*bigramScore +
		fe.weights.SameFingerBigrams*sfbPenalty +
		fe.weights.LateralStretches*lsbPenalty +
		fe.weights.RollQuality*rollScore +
		fe.weights.LayerPenalty*layerPenalty +
		// Enhanced KPI components
		fe.weights.HomeRowBonus*homeRowScore +
		fe.weights.RollRatioTarget*rollRatioScore +
		fe.weights.ThresholdBonuses*thresholdBonus +
		fe.weights.PositionMatching*positionScore

	return fitness
}

// calculateFingerDistance computes finger travel distance penalty.
func (fe *FitnessEvaluator) calculateFingerDistance(layout []rune, charset *genetic.CharacterSet, data genetic.KeyloggerDataInterface, charToPos map[rune]int) float64 {
	totalDistance := 0.0
	totalFreq := 0

	// Calculate distance for each bigram
	for bigram, freq := range data.GetAllBigrams() {
		if len(bigram) != 2 {
			continue
		}

		char1, char2 := rune(bigram[0]), rune(bigram[1])
		pos1, exists1 := charToPos[char1]
		pos2, exists2 := charToPos[char2]

		// Skip bigrams with characters not in our layout
		if !exists1 || !exists2 {
			continue
		}

		if coord1, ok := fe.geometry.KeyPositions[pos1]; ok {
			if coord2, ok := fe.geometry.KeyPositions[pos2]; ok {
				distance := math.Sqrt((coord1[0]-coord2[0])*(coord1[0]-coord2[0]) + (coord1[1]-coord2[1])*(coord1[1]-coord2[1]))
				totalDistance += distance * float64(freq)
				totalFreq += freq
			}
		}
	}

	if totalFreq == 0 {
		return 0
	}

	avgDistance := totalDistance / float64(totalFreq)
	// Convert to score (lower distance = higher score)
	return 1.0 / (1.0 + avgDistance)
}

// calculateHandAlternation rewards alternating between hands.
func (fe *FitnessEvaluator) calculateHandAlternation(layout []rune, charset *genetic.CharacterSet, data genetic.KeyloggerDataInterface, charToPos map[rune]int) float64 {
	alternations := 0
	total := 0

	for bigram, freq := range data.GetAllBigrams() {
		if len(bigram) != 2 {
			continue
		}

		char1, char2 := rune(bigram[0]), rune(bigram[1])
		pos1, exists1 := charToPos[char1]
		pos2, exists2 := charToPos[char2]

		// Skip bigrams with characters not in our layout
		if !exists1 || !exists2 {
			continue
		}

		finger1, ok1 := fe.geometry.FingerMap[pos1]
		finger2, ok2 := fe.geometry.FingerMap[pos2]

		if !ok1 || !ok2 {
			continue
		}

		// Check if different hands (fingers 0-3 left, 4-7 right)
		if (finger1 < 4 && finger2 >= 4) || (finger1 >= 4 && finger2 < 4) {
			alternations += freq
		}

		total += freq
	}

	if total == 0 {
		return 0
	}

	return float64(alternations) / float64(total)
}

// calculateFingerBalance ensures even distribution across fingers.
func (fe *FitnessEvaluator) calculateFingerBalance(layout []rune, charset *genetic.CharacterSet, data genetic.KeyloggerDataInterface, charToPos map[rune]int) float64 {
	fingerFreq := make([]int, 8) // 8 fingers
	totalFreq := 0

	// Count frequency per finger for all characters in charset
	for _, char := range charset.Characters {
		pos, exists := charToPos[char]
		if !exists {
			continue
		}

		finger, ok := fe.geometry.FingerMap[pos]
		if !ok {
			continue
		}

		freq := data.GetCharFreq(char)
		fingerFreq[finger] += freq
		totalFreq += freq
	}

	if totalFreq == 0 {
		return 0
	}

	// Calculate standard deviation of finger usage
	mean := float64(totalFreq) / 8.0
	variance := 0.0

	for _, freq := range fingerFreq {
		diff := float64(freq) - mean
		variance += diff * diff
	}

	variance /= 8.0

	stdDev := math.Sqrt(variance)

	// Lower standard deviation = better balance = higher score
	return 1.0 / (1.0 + stdDev/mean)
}

// calculateRowJumping penalizes jumping between rows.
func (fe *FitnessEvaluator) calculateRowJumping(layout []rune, charset *genetic.CharacterSet, data genetic.KeyloggerDataInterface, charToPos map[rune]int) float64 {
	rowJumps := 0
	total := 0

	for bigram, freq := range data.GetAllBigrams() {
		if len(bigram) != 2 {
			continue
		}

		char1, char2 := rune(bigram[0]), rune(bigram[1])
		pos1, exists1 := charToPos[char1]
		pos2, exists2 := charToPos[char2]

		// Skip bigrams with characters not in our layout
		if !exists1 || !exists2 {
			continue
		}

		if coord1, ok := fe.geometry.KeyPositions[pos1]; ok {
			if coord2, ok := fe.geometry.KeyPositions[pos2]; ok {
				if math.Abs(coord1[1]-coord2[1]) > 0.5 { // Different rows
					rowJumps += freq
				}

				total += freq
			}
		}
	}

	if total == 0 {
		return 1.0
	}

	jumpRatio := float64(rowJumps) / float64(total)

	return 1.0 - jumpRatio // Lower jumps = higher score
}

// calculateBigramEfficiency rewards common bigrams on easy finger combinations.
func (fe *FitnessEvaluator) calculateBigramEfficiency(layout []rune, charset *genetic.CharacterSet, data genetic.KeyloggerDataInterface, charToPos map[rune]int) float64 {
	efficiencySum := 0.0
	totalFreq := 0

	for bigram, freq := range data.GetAllBigrams() {
		if len(bigram) != 2 {
			continue
		}

		char1, char2 := rune(bigram[0]), rune(bigram[1])
		pos1, exists1 := charToPos[char1]
		pos2, exists2 := charToPos[char2]

		// Skip bigrams with characters not in our layout
		if !exists1 || !exists2 {
			continue
		}

		finger1, ok1 := fe.geometry.FingerMap[pos1]
		finger2, ok2 := fe.geometry.FingerMap[pos2]

		if !ok1 || !ok2 {
			continue
		}

		// Rate bigram efficiency based on finger combination
		efficiency := rateBigramEfficiency(finger1, finger2)
		efficiencySum += efficiency * float64(freq)
		totalFreq += freq
	}

	if totalFreq == 0 {
		return 0
	}

	return efficiencySum / float64(totalFreq)
}

// Legacy method for backward compatibility with [26]rune layouts.
func (fe *FitnessEvaluator) EvaluateLegacy(layout [26]rune, data genetic.KeyloggerDataInterface) float64 {
	// Create a full keyboard layout by placing the 26 letters in first 26 positions
	// and filling the rest with remaining characters from FullKeyboardCharset
	fullCharset := genetic.FullKeyboardCharset()
	fullLayout := make([]rune, 70)

	// Copy the 26-character layout to the first positions
	copy(fullLayout, layout[:])

	// Fill remaining positions with non-letter characters from full charset
	pos := 26
	for _, char := range fullCharset.Characters {
		if pos >= 70 {
			break
		}
		// Check if this character is already in the layout (i.e., is a letter)
		isLetter := false

		for _, letter := range layout {
			if char == letter {
				isLetter = true

				break
			}
		}

		if !isLetter {
			fullLayout[pos] = char
			pos++
		}
	}

	return fe.Evaluate(fullLayout, fullCharset, data)
}

// calculateSameFingerBigrams heavily penalizes same finger bigrams (SFBs).
func (fe *FitnessEvaluator) calculateSameFingerBigrams(layout []rune, charset *genetic.CharacterSet, data genetic.KeyloggerDataInterface, charToPos map[rune]int) float64 {
	sfbCount := 0
	totalBigrams := 0

	for bigram, freq := range data.GetAllBigrams() {
		if len(bigram) != 2 {
			continue
		}

		char1, char2 := rune(bigram[0]), rune(bigram[1])
		pos1, exists1 := charToPos[char1]
		pos2, exists2 := charToPos[char2]

		// Skip bigrams with characters not in our layout
		if !exists1 || !exists2 {
			continue
		}

		finger1, ok1 := fe.geometry.FingerMap[pos1]
		finger2, ok2 := fe.geometry.FingerMap[pos2]

		if !ok1 || !ok2 {
			continue
		}

		totalBigrams += freq

		// Same finger bigram (SFB) detection
		if finger1 == finger2 {
			sfbCount += freq
		}
	}

	if totalBigrams == 0 {
		return 1.0 // Perfect score if no data
	}

	sfbRate := float64(sfbCount) / float64(totalBigrams)

	// Return inverted score (fewer SFBs = higher score)
	return 1.0 - sfbRate
}

// calculateLateralStretches penalizes lateral stretch bigrams on index fingers.
func (fe *FitnessEvaluator) calculateLateralStretches(layout []rune, charset *genetic.CharacterSet, data genetic.KeyloggerDataInterface, charToPos map[rune]int) float64 {
	lsbCount := 0
	totalBigrams := 0

	for bigram, freq := range data.GetAllBigrams() {
		if len(bigram) != 2 {
			continue
		}

		char1, char2 := rune(bigram[0]), rune(bigram[1])
		pos1, exists1 := charToPos[char1]
		pos2, exists2 := charToPos[char2]

		// Skip bigrams with characters not in our layout
		if !exists1 || !exists2 {
			continue
		}

		finger1, ok1 := fe.geometry.FingerMap[pos1]
		finger2, ok2 := fe.geometry.FingerMap[pos2]
		coord1, ok3 := fe.geometry.KeyPositions[pos1]
		coord2, ok4 := fe.geometry.KeyPositions[pos2]

		if !ok1 || !ok2 || !ok3 || !ok4 {
			continue
		}

		totalBigrams += freq

		// Check for lateral stretch bigrams (index fingers stretching outward)
		// Left index (finger 3) stretching left, or right index (finger 4) stretching right
		isLSB := false

		if finger1 == 3 && finger2 == 3 { // Both on left index
			// Check if positions are far apart horizontally on same row
			if coord1[1] == coord2[1] && math.Abs(coord1[0]-coord2[0]) > 2.0 {
				isLSB = true
			}
		} else if finger1 == 4 && finger2 == 4 { // Both on right index
			// Check if positions are far apart horizontally on same row
			if coord1[1] == coord2[1] && math.Abs(coord1[0]-coord2[0]) > 2.0 {
				isLSB = true
			}
		}

		if isLSB {
			lsbCount += freq
		}
	}

	if totalBigrams == 0 {
		return 1.0 // Perfect score if no data
	}

	lsbRate := float64(lsbCount) / float64(totalBigrams)

	// Return inverted score (fewer LSBs = higher score)
	return 1.0 - lsbRate
}

// calculateRollQuality rewards smooth rolling motions using total roll rate.
func (fe *FitnessEvaluator) calculateRollQuality(layout []rune, charset *genetic.CharacterSet, data genetic.KeyloggerDataInterface, charToPos map[rune]int) float64 {
	rollCount := 0
	totalBigrams := 0

	for bigram, freq := range data.GetAllBigrams() {
		if len(bigram) != 2 {
			continue
		}

		char1, char2 := rune(bigram[0]), rune(bigram[1])
		pos1, exists1 := charToPos[char1]
		pos2, exists2 := charToPos[char2]

		// Skip bigrams with characters not in our layout
		if !exists1 || !exists2 {
			continue
		}

		finger1, ok1 := fe.geometry.FingerMap[pos1]
		finger2, ok2 := fe.geometry.FingerMap[pos2]
		coord1, ok3 := fe.geometry.KeyPositions[pos1]
		coord2, ok4 := fe.geometry.KeyPositions[pos2]

		if !ok1 || !ok2 || !ok3 || !ok4 {
			continue
		}

		totalBigrams += freq

		// Check for rolling motion (same hand, adjacent fingers, same row)
		sameHand := (finger1 < 4 && finger2 < 4) || (finger1 >= 4 && finger2 >= 4)
		adjacentFingers := math.Abs(float64(finger1-finger2)) == 1
		sameRow := math.Abs(coord1[1]-coord2[1]) < 0.3

		if sameHand && adjacentFingers && sameRow {
			rollCount += freq // Count all rolls (inward + outward)
		}
	}

	if totalBigrams == 0 {
		return 0.0
	}

	// Calculate total roll rate like the analysis does
	totalRollRate := float64(rollCount) / float64(totalBigrams)

	// Score based on analysis thresholds: >30% = EXCELLENT, >15% = GOOD, <15% = POOR
	if totalRollRate > 0.3 {
		return 1.0 // EXCELLENT
	} else if totalRollRate > 0.15 {
		return 0.7 // GOOD
	} else {
		// Scale POOR linearly: 0% = 0.0, 15% = 0.4
		return totalRollRate / 0.15 * 0.4
	}
}

// calculateLayerPenalty penalizes characters that require modifier keys.
func (fe *FitnessEvaluator) calculateLayerPenalty(layout []rune, charset *genetic.CharacterSet, data genetic.KeyloggerDataInterface, charToPos map[rune]int) float64 {
	// For the basic evaluator, we'll use a simple heuristic:
	// Penalize uppercase letters and special characters that typically require Shift
	totalPenalty := 0.0
	totalFreq := 0

	// Characters that typically require Shift key on most layouts
	shiftChars := map[rune]bool{
		'A': true, 'B': true, 'C': true, 'D': true, 'E': true, 'F': true, 'G': true, 'H': true, 'I': true, 'J': true,
		'K': true, 'L': true, 'M': true, 'N': true, 'O': true, 'P': true, 'Q': true, 'R': true, 'S': true, 'T': true,
		'U': true, 'V': true, 'W': true, 'X': true, 'Y': true, 'Z': true,
		'!': true, '@': true, '#': true, '$': true, '%': true, '^': true, '&': true, '*': true, '(': true, ')': true,
		'_': true, '+': true, '{': true, '}': true, '|': true, ':': true, '"': true, '<': true, '>': true, '?': true,
	}

	// Characters that typically require AltGr or other special modifiers
	altGrChars := map[rune]bool{
		'€': true, '£': true, '¥': true, '©': true, '®': true, '°': true,
	}

	// Check character frequencies
	for _, char := range charset.Characters {
		freq := data.GetCharFreq(char)
		if freq > 0 {
			totalFreq += freq

			if shiftChars[char] {
				// Apply moderate penalty for Shift characters
				totalPenalty += float64(freq) * 0.5
			} else if altGrChars[char] {
				// Apply higher penalty for AltGr characters
				totalPenalty += float64(freq) * 1.0
			}
		}
	}

	// Check bigram penalties for consecutive modifier usage
	for bigram, freq := range data.GetAllBigrams() {
		if len(bigram) == 2 {
			char1, char2 := rune(bigram[0]), rune(bigram[1])
			totalFreq += freq

			// Penalty for consecutive Shift usage (requires holding Shift)
			if shiftChars[char1] && shiftChars[char2] {
				totalPenalty += float64(freq) * 0.3
			}

			// Penalty for mixing modifiers
			if (shiftChars[char1] && altGrChars[char2]) || (altGrChars[char1] && shiftChars[char2]) {
				totalPenalty += float64(freq) * 0.7
			}
		}
	}

	if totalFreq == 0 {
		return 1.0 // Perfect score if no data
	}

	// Normalize penalty and invert (lower penalty = higher score)
	normalizedPenalty := totalPenalty / float64(totalFreq)

	return 1.0 / (1.0 + normalizedPenalty)
}

// rateBigramEfficiency rates how efficient a finger combination is.
func rateBigramEfficiency(finger1, finger2 int) float64 {
	// Same finger = worst
	if finger1 == finger2 {
		return 0.1
	}

	// Adjacent fingers on same hand = poor
	if math.Abs(float64(finger1-finger2)) == 1 &&
		((finger1 < 4 && finger2 < 4) || (finger1 >= 4 && finger2 >= 4)) {
		return 0.3
	}

	// Different hands = best
	if (finger1 < 4 && finger2 >= 4) || (finger1 >= 4 && finger2 < 4) {
		return 1.0
	}

	// Other combinations = moderate
	return 0.6
}

// calculateHomeRowBonus rewards placing high-frequency characters on the home row.
func (fe *FitnessEvaluator) calculateHomeRowBonus(layout []rune, charset *genetic.CharacterSet, data genetic.KeyloggerDataInterface, charToPos map[rune]int) float64 {
	homeRowPositions := map[int]bool{
		10: true, 11: true, 12: true, 13: true, 14: true, // Left home row
		15: true, 16: true, 17: true, 18: true, // Right home row
	}

	totalHomeRowFreq := 0
	totalFreq := 0

	// Calculate frequency of characters on home row
	for _, char := range charset.Characters {
		freq := data.GetCharFreq(char)
		if freq > 0 {
			totalFreq += freq

			if pos, exists := charToPos[char]; exists {
				if homeRowPositions[pos] {
					totalHomeRowFreq += freq
				}
			}
		}
	}

	if totalFreq == 0 {
		return 0.0
	}

	homeRowUsage := float64(totalHomeRowFreq) / float64(totalFreq)

	// Bonus scoring based on home row usage thresholds (matching analysis display)
	if homeRowUsage > 0.4 {
		// OPTIMIZED: >40% home row usage gets full bonus plus extra
		return 1.0 + (homeRowUsage-0.4)*2.0 // Max 2.2 score for 50%+ usage
	} else if homeRowUsage > 0.3 {
		// Good usage gets proportional bonus
		return homeRowUsage / 0.4 // Scale 0.75-1.0
	} else {
		// SUBOPTIMAL: <30% gets penalized
		return homeRowUsage / 0.4 * 0.5 // Scale 0-0.375
	}
}

// calculateRollRatioTarget optimizes for ideal roll percentage of same-hand bigrams.
func (fe *FitnessEvaluator) calculateRollRatioTarget(layout []rune, charset *genetic.CharacterSet, data genetic.KeyloggerDataInterface, charToPos map[rune]int) float64 {
	sameHandBigrams := 0
	rollBigrams := 0

	for bigram, freq := range data.GetAllBigrams() {
		if len(bigram) != 2 {
			continue
		}

		char1, char2 := rune(bigram[0]), rune(bigram[1])
		pos1, exists1 := charToPos[char1]
		pos2, exists2 := charToPos[char2]

		if !exists1 || !exists2 {
			continue
		}

		finger1, ok1 := fe.geometry.FingerMap[pos1]
		finger2, ok2 := fe.geometry.FingerMap[pos2]
		coord1, ok3 := fe.geometry.KeyPositions[pos1]
		coord2, ok4 := fe.geometry.KeyPositions[pos2]

		if !ok1 || !ok2 || !ok3 || !ok4 {
			continue
		}

		// Check if same hand
		sameHand := (finger1 < 4 && finger2 < 4) || (finger1 >= 4 && finger2 >= 4)
		if !sameHand {
			continue
		}

		sameHandBigrams += freq

		// Check for rolling motion
		adjacentFingers := math.Abs(float64(finger1-finger2)) == 1
		sameRow := math.Abs(coord1[1]-coord2[1]) < 0.3

		if adjacentFingers && sameRow {
			rollBigrams += freq
		}
	}

	if sameHandBigrams == 0 {
		return 0.0
	}

	rollRatio := float64(rollBigrams) / float64(sameHandBigrams)

	// Target roll ratio of 35% (based on ergonomic research)
	targetRatio := 0.35
	deviation := math.Abs(rollRatio - targetRatio)

	// Score decreases as we deviate from target
	return math.Max(0.0, 1.0-deviation*2.0)
}

// calculateThresholdBonuses provides bonus rewards for crossing key ergonomic thresholds.
func (fe *FitnessEvaluator) calculateThresholdBonuses(
	layout []rune,
	charset *genetic.CharacterSet,
	data genetic.KeyloggerDataInterface,
	charToPos map[rune]int,
	alternationScore, rollScore float64,
) float64 {
	bonusScore := 0.0

	// Hand alternation threshold bonuses (based on analysis categories)
	if alternationScore > 0.6 {
		bonusScore += 1.0 // EXCELLENT bonus
	} else if alternationScore > 0.45 {
		bonusScore += 0.6 // GOOD bonus
	} else if alternationScore > 0.3 {
		bonusScore += 0.3 // MODERATE bonus
	}

	// Roll quality threshold bonus
	if rollScore > 0.4 {
		bonusScore += 0.8 // High roll frequency bonus
	} else if rollScore > 0.2 {
		bonusScore += 0.4 // Moderate roll frequency bonus
	}

	// Calculate SFB rate for threshold bonus
	sfbCount := 0
	totalBigrams := 0

	for bigram, freq := range data.GetAllBigrams() {
		if len(bigram) != 2 {
			continue
		}

		char1, char2 := rune(bigram[0]), rune(bigram[1])
		pos1, exists1 := charToPos[char1]
		pos2, exists2 := charToPos[char2]

		if !exists1 || !exists2 {
			continue
		}

		finger1, ok1 := fe.geometry.FingerMap[pos1]
		finger2, ok2 := fe.geometry.FingerMap[pos2]

		if !ok1 || !ok2 {
			continue
		}

		totalBigrams += freq
		if finger1 == finger2 {
			sfbCount += freq
		}
	}

	if totalBigrams > 0 {
		sfbRate := float64(sfbCount) / float64(totalBigrams)
		// SFB threshold bonuses
		if sfbRate < 0.02 {
			bonusScore += 1.0 // EXCELLENT SFB rate
		} else if sfbRate < 0.05 {
			bonusScore += 0.5 // GOOD SFB rate
		}
	}

	// Normalize bonus score (max possible: ~3.0)
	return bonusScore / 3.0
}

// calculatePositionMatching rewards placing high-frequency characters in ergonomic positions.
func (fe *FitnessEvaluator) calculatePositionMatching(layout []rune, charset *genetic.CharacterSet, data genetic.KeyloggerDataInterface, charToPos map[rune]int) float64 {
	// Define ergonomic scores for each position (higher = more ergonomic)
	ergonomicScores := map[int]float64{
		// Home row (most ergonomic)
		13: 1.0, 14: 1.0, 15: 1.0, 16: 1.0, // Index and middle fingers
		12: 0.9, 17: 0.9, // Ring fingers
		11: 0.7, 18: 0.7, // Pinky fingers
		10: 0.6, // Left pinky home

		// Top row (moderately ergonomic)
		3: 0.8, 4: 0.8, 5: 0.8, 6: 0.8, // Index and middle
		2: 0.7, 7: 0.7, // Ring
		1: 0.5, 8: 0.5, 9: 0.5, // Pinky and edges
		0: 0.4, // Far left

		// Bottom row (less ergonomic)
		22: 0.6, 23: 0.6, 24: 0.6, // Index and middle
		21: 0.5, 25: 0.5, // Ring and right pinky
		20: 0.4, // Left ring
		19: 0.3, // Left pinky
	}

	// Get character frequencies and sort them
	type charFreq struct {
		char rune
		freq int
	}

	var frequencies []charFreq

	totalScore := 0.0
	totalWeight := 0.0

	for _, char := range charset.Characters {
		freq := data.GetCharFreq(char)
		if freq > 0 {
			frequencies = append(frequencies, charFreq{char, freq})
		}
	}

	// Calculate weighted ergonomic score based on frequency-position matching
	for _, cf := range frequencies {
		if pos, exists := charToPos[cf.char]; exists {
			if ergScore, hasErg := ergonomicScores[pos]; hasErg {
				weight := float64(cf.freq)
				totalScore += ergScore * weight
				totalWeight += weight
			}
		}
	}

	if totalWeight == 0 {
		return 0.0
	}

	// Return average ergonomic score weighted by character frequency
	return totalScore / totalWeight
}
