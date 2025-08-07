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
	FingerDistance   float64 `json:"finger_distance"`
	HandAlternation  float64 `json:"hand_alternation"`
	FingerBalance    float64 `json:"finger_balance"`
	RowJumping       float64 `json:"row_jumping"`
	BigramEfficiency float64 `json:"bigram_efficiency"`
}

// DefaultWeights returns balanced fitness weights.
func DefaultWeights() FitnessWeights {
	return FitnessWeights{
		FingerDistance:   0.3,
		HandAlternation:  0.2,
		FingerBalance:    0.2,
		RowJumping:       0.15,
		BigramEfficiency: 0.15,
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

	// Weighted sum of all components (higher is better)
	fitness := fe.weights.FingerDistance*distanceScore +
		fe.weights.HandAlternation*alternationScore +
		fe.weights.FingerBalance*balanceScore +
		fe.weights.RowJumping*rowJumpScore +
		fe.weights.BigramEfficiency*bigramScore

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
	// Convert [26]rune to []rune with alphabet charset
	layoutSlice := make([]rune, 26)
	copy(layoutSlice, layout[:])

	return fe.Evaluate(layoutSlice, genetic.AlphabetOnly(), data)
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
