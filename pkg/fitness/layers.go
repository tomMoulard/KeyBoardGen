package fitness

import (
	"github.com/tommoulard/keyboardgen/pkg/genetic"
)

// KeyLayer represents different key layers (normal, shift, etc.).
type KeyLayer int

const (
	BaseLayer KeyLayer = iota
	ShiftLayer
	AltGrLayer
)

// LayeredKey represents a key that can have multiple characters based on modifiers.
type LayeredKey struct {
	BaseChar  rune  // The base character (no modifiers)
	ShiftChar rune  // Character with Shift modifier
	AltGrChar *rune // Optional AltGr character (can be nil)
}

// KeyboardLayout represents a physical keyboard layout with layer support.
type KeyboardLayout struct {
	Name     string             // Layout name (e.g., "QWERTY", "AZERTY")
	Keys     map[int]LayeredKey // Position -> LayeredKey mapping
	Geometry KeyboardGeometry   // Physical key positions
}

// GetCharacterLayer returns which layer and modifiers are needed for a character.
func (kl *KeyboardLayout) GetCharacterLayer(char rune) (position int, layer KeyLayer, cost float64) {
	// Search through all positions and layers
	for pos, layeredKey := range kl.Keys {
		if layeredKey.BaseChar == char {
			return pos, BaseLayer, 1.0 // Base cost for unmodified key
		}

		if layeredKey.ShiftChar == char {
			return pos, ShiftLayer, 1.5 // Higher cost for shift modifier
		}

		if layeredKey.AltGrChar != nil && *layeredKey.AltGrChar == char {
			return pos, AltGrLayer, 2.0 // Highest cost for AltGr modifier
		}
	}

	// Character not found in this layout
	return -1, BaseLayer, 100.0 // Very high penalty for missing characters
}

// LayerPenalty calculates additional fitness penalty for using modifiers.
func (kl *KeyboardLayout) LayerPenalty(char1, char2 rune) float64 {
	_, layer1, cost1 := kl.GetCharacterLayer(char1)
	_, layer2, cost2 := kl.GetCharacterLayer(char2)

	// Base penalty is the sum of individual key costs
	penalty := cost1 + cost2

	// Additional penalties for specific layer combinations
	if layer1 == ShiftLayer && layer2 == ShiftLayer {
		// Two shift keys in a row - requires holding shift
		penalty += 0.3
	} else if layer1 == ShiftLayer || layer2 == ShiftLayer {
		// One shift key - moderate penalty
		penalty += 0.1
	}

	if layer1 == AltGrLayer || layer2 == AltGrLayer {
		// AltGr usage is generally more awkward
		penalty += 0.5
	}

	return penalty
}

// StandardQWERTYLayout creates a QWERTY layout with layer support.
func StandardQWERTYLayout() *KeyboardLayout {
	qwerty := &KeyboardLayout{
		Name:     "QWERTY",
		Keys:     make(map[int]LayeredKey),
		Geometry: StandardGeometry(),
	}

	// Top row (positions 0-9): 1234567890 -> !@#$%^&*()
	// Some keys also have AltGr characters for international layouts
	euro := '€'
	qwerty.Keys[0] = LayeredKey{BaseChar: '1', ShiftChar: '!'}
	qwerty.Keys[1] = LayeredKey{BaseChar: '2', ShiftChar: '@'}
	qwerty.Keys[2] = LayeredKey{BaseChar: '3', ShiftChar: '#'}
	qwerty.Keys[3] = LayeredKey{BaseChar: '4', ShiftChar: '$', AltGrChar: &euro}
	qwerty.Keys[4] = LayeredKey{BaseChar: '5', ShiftChar: '%'}
	qwerty.Keys[5] = LayeredKey{BaseChar: '6', ShiftChar: '^'}
	qwerty.Keys[6] = LayeredKey{BaseChar: '7', ShiftChar: '&'}
	qwerty.Keys[7] = LayeredKey{BaseChar: '8', ShiftChar: '*'}
	qwerty.Keys[8] = LayeredKey{BaseChar: '9', ShiftChar: '('}
	qwerty.Keys[9] = LayeredKey{BaseChar: '0', ShiftChar: ')'}

	// Top letter row (positions 10-19): qwertyuiop
	qwertyTopRow := []rune{'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p'}
	for i, letter := range qwertyTopRow {
		pos := 10 + i
		if pos == 14 { // 't' key gets trademark symbol on AltGr
			tm := '™'
			qwerty.Keys[pos] = LayeredKey{BaseChar: letter, ShiftChar: letter - 32, AltGrChar: &tm}
		} else {
			qwerty.Keys[pos] = LayeredKey{BaseChar: letter, ShiftChar: letter - 32}
		}
	}

	// Home row (positions 20-28): asdfghjkl
	homeRow := []rune{'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l'}
	for i, letter := range homeRow {
		pos := 20 + i
		if pos == 23 { // 'f' key gets function symbol
			degree := '°'
			qwerty.Keys[pos] = LayeredKey{BaseChar: letter, ShiftChar: letter - 32, AltGrChar: &degree}
		} else {
			qwerty.Keys[pos] = LayeredKey{BaseChar: letter, ShiftChar: letter - 32}
		}
	}

	// Bottom row (positions 29-35): zxcvbnm
	bottomRow := []rune{'z', 'x', 'c', 'v', 'b', 'n', 'm'}
	for i, letter := range bottomRow {
		pos := 29 + i
		if pos == 31 { // 'c' key gets copyright symbol
			copyright := '©'
			qwerty.Keys[pos] = LayeredKey{BaseChar: letter, ShiftChar: letter - 32, AltGrChar: &copyright}
		} else {
			qwerty.Keys[pos] = LayeredKey{BaseChar: letter, ShiftChar: letter - 32}
		}
	}

	return qwerty
}

// StandardAZERTYLayout creates an AZERTY layout with layer support.
func StandardAZERTYLayout() *KeyboardLayout {
	azerty := &KeyboardLayout{
		Name:     "AZERTY",
		Keys:     make(map[int]LayeredKey),
		Geometry: StandardGeometry(),
	}

	// AZERTY top row: &é"'(-è_çà -> 1234567890
	// Many keys have AltGr characters for European symbols
	euro := '€'
	pound := '£'
	yen := '¥'
	micro := 'µ'

	azerty.Keys[0] = LayeredKey{BaseChar: '&', ShiftChar: '1'}
	azerty.Keys[1] = LayeredKey{BaseChar: 'é', ShiftChar: '2', AltGrChar: &euro}
	azerty.Keys[2] = LayeredKey{BaseChar: '"', ShiftChar: '3', AltGrChar: &pound}
	azerty.Keys[3] = LayeredKey{BaseChar: '\'', ShiftChar: '4'}
	azerty.Keys[4] = LayeredKey{BaseChar: '(', ShiftChar: '5'}
	azerty.Keys[5] = LayeredKey{BaseChar: '-', ShiftChar: '6'}
	azerty.Keys[6] = LayeredKey{BaseChar: 'è', ShiftChar: '7', AltGrChar: &yen}
	azerty.Keys[7] = LayeredKey{BaseChar: '_', ShiftChar: '8'}
	azerty.Keys[8] = LayeredKey{BaseChar: 'ç', ShiftChar: '9'}
	azerty.Keys[9] = LayeredKey{BaseChar: 'à', ShiftChar: '0'}

	// AZERTY top letter row (positions 10-19): azertyuiop
	azertyTopLetters := []struct {
		pos   int
		base  rune
		shift rune
		altgr *rune
	}{
		{10, 'a', 'A', nil},
		{11, 'z', 'Z', nil},
		{12, 'e', 'E', &euro},
		{13, 'r', 'R', nil},
		{14, 't', 'T', nil},
		{15, 'y', 'Y', nil},
		{16, 'u', 'U', &micro},
		{17, 'i', 'I', nil},
		{18, 'o', 'O', nil},
		{19, 'p', 'P', nil},
	}

	for _, letter := range azertyTopLetters {
		azerty.Keys[letter.pos] = LayeredKey{
			BaseChar:  letter.base,
			ShiftChar: letter.shift,
			AltGrChar: letter.altgr,
		}
	}

	// AZERTY home row (positions 20-28): qsdfghjklm
	qsdfghjklm := []rune{'q', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l'}
	for i, letter := range qsdfghjklm {
		pos := 20 + i

		degree := '°'
		if pos == 23 { // 'f' key gets degree symbol
			azerty.Keys[pos] = LayeredKey{BaseChar: letter, ShiftChar: letter - 32, AltGrChar: &degree}
		} else {
			azerty.Keys[pos] = LayeredKey{BaseChar: letter, ShiftChar: letter - 32}
		}
	}

	// AZERTY bottom row (positions 29-35): wxcvbnm
	wxcvbnm := []rune{'w', 'x', 'c', 'v', 'b', 'n', 'm'}
	for i, letter := range wxcvbnm {
		pos := 29 + i
		copyright := '©'
		registered := '®'

		switch pos {
		case 31: // 'c' key gets copyright symbol
			azerty.Keys[pos] = LayeredKey{BaseChar: letter, ShiftChar: letter - 32, AltGrChar: &copyright}
		case 34: // 'b' key gets registered symbol
			azerty.Keys[pos] = LayeredKey{BaseChar: letter, ShiftChar: letter - 32, AltGrChar: &registered}
		default:
			azerty.Keys[pos] = LayeredKey{BaseChar: letter, ShiftChar: letter - 32}
		}
	}

	// Additional punctuation keys for AZERTY
	azerty.Keys[36] = LayeredKey{BaseChar: '°', ShiftChar: ')'}
	azerty.Keys[37] = LayeredKey{BaseChar: '!', ShiftChar: '§'}
	azerty.Keys[38] = LayeredKey{BaseChar: ';', ShiftChar: '.'}
	azerty.Keys[39] = LayeredKey{BaseChar: ':', ShiftChar: '/'}

	return azerty
}

// LayerAwareFitnessEvaluator extends the standard fitness evaluator with layer awareness.
type LayerAwareFitnessEvaluator struct {
	*FitnessEvaluator

	Layout *KeyboardLayout
}

// NewLayerAwareFitnessEvaluator creates a new layer-aware fitness evaluator.
func NewLayerAwareFitnessEvaluator(layout *KeyboardLayout, weights FitnessWeights) *LayerAwareFitnessEvaluator {
	baseEvaluator := NewFitnessEvaluator(layout.Geometry, weights)

	return &LayerAwareFitnessEvaluator{
		FitnessEvaluator: baseEvaluator,
		Layout:           layout,
	}
}

// EvaluateWithLayers calculates fitness with layer penalties.
func (lafe *LayerAwareFitnessEvaluator) EvaluateWithLayers(individual genetic.Individual, data genetic.KeyloggerDataInterface) float64 {
	// Start with the base fitness calculation
	baseFitness := lafe.Evaluate(individual.Layout, individual.Charset, data)

	// Calculate layer penalty
	layerPenalty := lafe.calculateLayerPenalty(data)

	// Combine base fitness with layer penalty
	// Layer penalty is subtracted to reduce fitness for layouts requiring many modifiers
	adjustedFitness := baseFitness - (layerPenalty * 0.2) // 20% weight for layer penalties

	return adjustedFitness
}

// calculateLayerPenalty computes the total penalty for using modifier keys.
func (lafe *LayerAwareFitnessEvaluator) calculateLayerPenalty(data genetic.KeyloggerDataInterface) float64 {
	totalPenalty := 0.0
	totalFreq := 0

	// Analyze character frequencies by checking all possible characters
	// Since we don't have GetCharFrequencies, we'll iterate through common characters
	commonChars := []rune{
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
		'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
		'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'!', '@', '#', '$', '%', '^', '&', '*', '(', ')', '-', '_', '=', '+',
		'[', ']', '{', '}', '\\', '|', ';', ':', '\'', '"', ',', '.', '<', '>', '/', '?',
	}

	for _, char := range commonChars {
		freq := data.GetCharFreq(char)
		if freq > 0 {
			_, _, cost := lafe.Layout.GetCharacterLayer(char)
			if cost > 1.0 {
				// Apply penalty for characters requiring modifiers
				penalty := (cost - 1.0) * float64(freq)
				totalPenalty += penalty
			}

			totalFreq += freq
		}
	}

	// Analyze bigram penalties
	for bigram, freq := range data.GetAllBigrams() {
		if len(bigram) == 2 {
			char1, char2 := rune(bigram[0]), rune(bigram[1])

			bigramPenalty := lafe.Layout.LayerPenalty(char1, char2)
			if bigramPenalty > 2.0 { // Only apply penalty if above baseline cost
				totalPenalty += (bigramPenalty - 2.0) * float64(freq)
			}
		}

		totalFreq += freq
	}

	if totalFreq == 0 {
		return 0.0
	}

	// Normalize penalty by total frequency
	return totalPenalty / float64(totalFreq)
}
