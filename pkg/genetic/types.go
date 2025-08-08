package genetic

import (
	"math/rand/v2"
	"sync"
)

// KeyPosition represents a key position on the keyboard.
type KeyPosition int

// Individual represents a keyboard layout individual in the GA.
type Individual struct {
	Layout  []rune        // Maps positions to characters (dynamic size)
	Charset *CharacterSet // Character set this layout uses
	Fitness float64       // Cached fitness score
	Age     int           // Generation age
}

// Population represents a collection of individuals.
type Population []Individual

// Config holds genetic algorithm configuration.
type Config struct {
	PopulationSize  int     `json:"population_size"`
	MaxGenerations  int     `json:"max_generations"`
	MutationRate    float64 `json:"mutation_rate"`
	CrossoverRate   float64 `json:"crossover_rate"`
	ElitismCount    int     `json:"elitism_count"`
	TournamentSize  int     `json:"tournament_size"`
	ParallelWorkers int     `json:"parallel_workers"`
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig() Config {
	return Config{
		PopulationSize:  100,
		MaxGenerations:  1000,
		MutationRate:    0.1,
		CrossoverRate:   0.8,
		ElitismCount:    5,
		TournamentSize:  3,
		ParallelWorkers: 4,
	}
}

// LargeDatasetConfig returns configuration optimized for large datasets.
func LargeDatasetConfig() Config {
	return Config{
		PopulationSize:  500, // Much larger population for diversity
		MaxGenerations:  100, // Fewer generations but larger population
		MutationRate:    0.3, // Higher mutation for exploration
		CrossoverRate:   0.9, // High crossover for mixing
		ElitismCount:    2,   // Very low elitism to prevent dominance
		TournamentSize:  7,   // Larger tournaments for better selection pressure
		ParallelWorkers: 8,   // More workers for larger population
	}
}

// AdaptiveConfig returns configuration that adapts based on data size.
func AdaptiveConfig(dataSize int) Config {
	if dataSize > 100000 { // Large dataset like Harry Potter
		return LargeDatasetConfig()
	} else if dataSize > 10000 { // Medium dataset
		return Config{
			PopulationSize:  200,
			MaxGenerations:  200,
			MutationRate:    0.2,
			CrossoverRate:   0.85,
			ElitismCount:    3,
			TournamentSize:  5,
			ParallelWorkers: 6,
		}
	} else { // Small dataset
		return DefaultConfig()
	}
}

// GeneticAlgorithm defines the interface for GA operations.
type GeneticAlgorithm interface {
	Initialize(config Config) Population
	Evaluate(individual *Individual, data *KeyloggerData) float64
	Select(population Population, count int) []Individual
	Crossover(parent1, parent2 Individual) Individual
	Mutate(individual Individual) Individual
	Evolve(population Population, config Config, data *KeyloggerData) Population
}

// KeyloggerData holds parsed keylogger information.
type KeyloggerData struct {
	CharFrequency map[rune]int   `json:"char_frequency"`
	BigramFreq    map[string]int `json:"bigram_frequency"`
	TrigramFreq   map[string]int `json:"trigram_frequency"`
	TotalChars    int            `json:"total_chars"`
	mutex         sync.RWMutex
}

// NewKeyloggerData creates a new KeyloggerData instance.
func NewKeyloggerData() *KeyloggerData {
	return &KeyloggerData{
		CharFrequency: make(map[rune]int),
		BigramFreq:    make(map[string]int),
		TrigramFreq:   make(map[string]int),
	}
}

// AddChar safely adds a character frequency.
func (kd *KeyloggerData) AddChar(char rune) {
	kd.mutex.Lock()
	defer kd.mutex.Unlock()

	kd.CharFrequency[char]++
	kd.TotalChars++
}

// AddBigram safely adds a bigram frequency.
func (kd *KeyloggerData) AddBigram(bigram string) {
	kd.mutex.Lock()
	defer kd.mutex.Unlock()

	kd.BigramFreq[bigram]++
}

// AddTrigram safely adds a trigram frequency.
func (kd *KeyloggerData) AddTrigram(trigram string) {
	kd.mutex.Lock()
	defer kd.mutex.Unlock()

	kd.TrigramFreq[trigram]++
}

// GetCharFreq safely gets character frequency.
func (kd *KeyloggerData) GetCharFreq(char rune) int {
	kd.mutex.RLock()
	defer kd.mutex.RUnlock()

	return kd.CharFrequency[char]
}

// GetBigramFreq safely gets bigram frequency.
func (kd *KeyloggerData) GetBigramFreq(bigram string) int {
	kd.mutex.RLock()
	defer kd.mutex.RUnlock()

	return kd.BigramFreq[bigram]
}

// GetTrigramFreq safely gets trigram frequency.
func (kd *KeyloggerData) GetTrigramFreq(trigram string) int {
	kd.mutex.RLock()
	defer kd.mutex.RUnlock()

	return kd.TrigramFreq[trigram]
}

// GetTotalChars safely gets total character count.
func (kd *KeyloggerData) GetTotalChars() int {
	kd.mutex.RLock()
	defer kd.mutex.RUnlock()

	return kd.TotalChars
}

// GetAllBigrams safely gets all bigrams.
func (kd *KeyloggerData) GetAllBigrams() map[string]int {
	kd.mutex.RLock()
	defer kd.mutex.RUnlock()

	// Create a copy to avoid race conditions
	result := make(map[string]int)
	for k, v := range kd.BigramFreq {
		result[k] = v
	}

	return result
}

// Clone creates a deep copy of an individual.
func (ind Individual) Clone() Individual {
	clone := Individual{
		Layout:  make([]rune, len(ind.Layout)),
		Charset: ind.Charset,
		Fitness: ind.Fitness,
		Age:     ind.Age,
	}
	copy(clone.Layout, ind.Layout)

	return clone
}

// IsValid checks if the individual has a valid layout using its character set.
func (ind Individual) IsValid() bool {
	if ind.Charset == nil {
		return false
	}

	return ind.Charset.IsValid(ind.Layout)
}

// NewRandomIndividual creates a random individual with full keyboard charset.
func NewRandomIndividual() Individual {
	return NewRandomIndividualWithCharset(FullKeyboardCharset())
}

// InitializationStrategy represents different ways to create initial individuals.
type InitializationStrategy int

const (
	RandomShuffle InitializationStrategy = iota
	FrequencyBased
	HandBalance
	RowBalance
	CommonPatternsFirst
	AntiQWERTY
)

// NewRandomIndividualWithCharset creates a random individual using the specified character set.
func NewRandomIndividualWithCharset(charset *CharacterSet) Individual {
	if charset == nil {
		charset = FullKeyboardCharset()
	}

	// Copy characters for shuffling
	chars := make([]rune, len(charset.Characters))
	copy(chars, charset.Characters)

	// Fisher-Yates shuffle
	for i := len(chars) - 1; i > 0; i-- {
		j := rand.IntN(i + 1)
		chars[i], chars[j] = chars[j], chars[i]
	}

	return Individual{
		Layout:  chars,
		Charset: charset,
		Fitness: 0.0,
		Age:     0,
	}
}

// NewRandomIndividualWithStrategy creates a random individual using a specific initialization strategy.
func NewRandomIndividualWithStrategy(charset *CharacterSet, strategy InitializationStrategy, data KeyloggerDataInterface) Individual {
	if charset == nil {
		charset = FullKeyboardCharset()
	}

	var chars []rune

	switch strategy {
	case FrequencyBased:
		chars = createFrequencyBasedLayout(charset, data)
	case HandBalance:
		chars = createHandBalancedLayout(charset, data)
	case RowBalance:
		chars = createRowBalancedLayout(charset, data)
	case CommonPatternsFirst:
		chars = createPatternBasedLayout(charset, data)
	case AntiQWERTY:
		chars = createAntiQWERTYLayout(charset)
	default:
		// Default to random shuffle
		chars = make([]rune, len(charset.Characters))
		copy(chars, charset.Characters)

		for i := len(chars) - 1; i > 0; i-- {
			j := rand.IntN(i + 1)
			chars[i], chars[j] = chars[j], chars[i]
		}
	}

	return Individual{
		Layout:  chars,
		Charset: charset,
		Fitness: 0.0,
		Age:     0,
	}
}

// createFrequencyBasedLayout places frequent characters in home row positions.
func createFrequencyBasedLayout(charset *CharacterSet, data KeyloggerDataInterface) []rune {
	chars := make([]rune, len(charset.Characters))
	copy(chars, charset.Characters)

	if data == nil {
		// Fallback to random if no data
		for i := len(chars) - 1; i > 0; i-- {
			j := rand.IntN(i + 1)
			chars[i], chars[j] = chars[j], chars[i]
		}

		return chars
	}

	// Sort characters by frequency
	type charFreq struct {
		char rune
		freq int
	}

	frequencies := make([]charFreq, 0, len(chars))
	for _, char := range chars {
		freq := data.GetCharFreq(char)
		frequencies = append(frequencies, charFreq{char, freq})
	}

	// Sort by frequency (descending)
	for i := range len(frequencies) - 1 {
		for j := i + 1; j < len(frequencies); j++ {
			if frequencies[j].freq > frequencies[i].freq {
				frequencies[i], frequencies[j] = frequencies[j], frequencies[i]
			}
		}
	}

	// Place most frequent characters in home row (positions 9-17 for alphabet)
	result := make([]rune, len(chars))
	homeRowPositions := []int{9, 10, 11, 12, 13, 14, 15, 16, 17} // Home row

	// Fill home row with most frequent characters
	freqIndex := 0

	for i, pos := range homeRowPositions {
		if i < len(frequencies) && pos < len(result) {
			result[pos] = frequencies[freqIndex].char
			freqIndex++
		}
	}

	// Fill remaining positions with remaining characters (shuffled)
	remainingChars := make([]rune, 0)
	for i := freqIndex; i < len(frequencies); i++ {
		remainingChars = append(remainingChars, frequencies[i].char)
	}

	// Shuffle remaining characters
	for i := len(remainingChars) - 1; i > 0; i-- {
		j := rand.IntN(i + 1)
		remainingChars[i], remainingChars[j] = remainingChars[j], remainingChars[i]
	}

	// Fill non-home-row positions
	remainingIndex := 0

	for i := range result {
		if result[i] == 0 { // Empty position
			if remainingIndex < len(remainingChars) {
				result[i] = remainingChars[remainingIndex]
				remainingIndex++
			}
		}
	}

	return result
}

// createHandBalancedLayout tries to balance frequent characters between hands.
func createHandBalancedLayout(charset *CharacterSet, data KeyloggerDataInterface) []rune {
	chars := make([]rune, len(charset.Characters))
	copy(chars, charset.Characters)

	if data == nil {
		// Fallback to random
		for i := len(chars) - 1; i > 0; i-- {
			j := rand.IntN(i + 1)
			chars[i], chars[j] = chars[j], chars[i]
		}

		return chars
	}

	// Get character frequencies
	type charFreq struct {
		char rune
		freq int
	}

	frequencies := make([]charFreq, 0, len(chars))
	for _, char := range chars {
		freq := data.GetCharFreq(char)
		frequencies = append(frequencies, charFreq{char, freq})
	}

	// Sort by frequency (descending)
	for i := range len(frequencies) - 1 {
		for j := i + 1; j < len(frequencies); j++ {
			if frequencies[j].freq > frequencies[i].freq {
				frequencies[i], frequencies[j] = frequencies[j], frequencies[i]
			}
		}
	}

	result := make([]rune, len(chars))

	// Define left and right hand positions (for standard keyboard)
	leftHand := []int{0, 1, 2, 3, 4, 9, 10, 11, 12, 13, 18, 19, 20, 21, 22} // Left side
	rightHand := []int{5, 6, 7, 8, 14, 15, 16, 17, 23, 24, 25}              // Right side

	// Alternate placing frequent characters between hands
	leftIndex, rightIndex := 0, 0
	for i, cf := range frequencies {
		if i%2 == 0 && leftIndex < len(leftHand) {
			// Place on left hand
			result[leftHand[leftIndex]] = cf.char
			leftIndex++
		} else if rightIndex < len(rightHand) {
			// Place on right hand
			result[rightHand[rightIndex]] = cf.char
			rightIndex++
		} else if leftIndex < len(leftHand) {
			// Fallback to left if right is full
			result[leftHand[leftIndex]] = cf.char
			leftIndex++
		}
	}

	return result
}

// createRowBalancedLayout distributes characters across rows based on frequency.
func createRowBalancedLayout(charset *CharacterSet, data KeyloggerDataInterface) []rune {
	chars := make([]rune, len(charset.Characters))
	copy(chars, charset.Characters)

	// Simple row-based distribution: home row gets priority
	topRow := []int{0, 1, 2, 3, 4, 5, 6, 7, 8}
	homeRow := []int{9, 10, 11, 12, 13, 14, 15, 16, 17}
	bottomRow := []int{18, 19, 20, 21, 22, 23, 24, 25}

	result := make([]rune, len(chars))

	// Shuffle and distribute
	for i := len(chars) - 1; i > 0; i-- {
		j := rand.IntN(i + 1)
		chars[i], chars[j] = chars[j], chars[i]
	}

	// Fill home row first, then top, then bottom
	charIndex := 0

	// Fill home row
	for _, pos := range homeRow {
		if charIndex < len(chars) && pos < len(result) {
			result[pos] = chars[charIndex]
			charIndex++
		}
	}

	// Fill top row
	for _, pos := range topRow {
		if charIndex < len(chars) && pos < len(result) {
			result[pos] = chars[charIndex]
			charIndex++
		}
	}

	// Fill bottom row
	for _, pos := range bottomRow {
		if charIndex < len(chars) && pos < len(result) {
			result[pos] = chars[charIndex]
			charIndex++
		}
	}

	return result
}

// createPatternBasedLayout tries to optimize for common bigram patterns.
func createPatternBasedLayout(charset *CharacterSet, data KeyloggerDataInterface) []rune {
	chars := make([]rune, len(charset.Characters))
	copy(chars, charset.Characters)

	if data == nil {
		// Fallback to random
		for i := len(chars) - 1; i > 0; i-- {
			j := rand.IntN(i + 1)
			chars[i], chars[j] = chars[j], chars[i]
		}

		return chars
	}

	// Start with a base shuffle
	for i := len(chars) - 1; i > 0; i-- {
		j := rand.IntN(i + 1)
		chars[i], chars[j] = chars[j], chars[i]
	}

	// Apply some heuristics for common patterns
	result := make([]rune, len(chars))
	copy(result, chars)

	// Try to separate vowels (basic heuristic)
	vowels := []rune{'a', 'e', 'i', 'o', 'u'}
	for i, char := range result {
		for _, vowel := range vowels {
			if char == vowel {
				// Try to move vowel to a different hand position
				swapWith := (i + len(result)/2) % len(result)
				result[i], result[swapWith] = result[swapWith], result[i]

				break
			}
		}
	}

	return result
}

// createAntiQWERTYLayout creates a layout that's maximally different from QWERTY.
func createAntiQWERTYLayout(charset *CharacterSet) []rune {
	chars := make([]rune, len(charset.Characters))
	copy(chars, charset.Characters)

	// QWERTY layout for reference
	qwerty := []rune{
		'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p',
		'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l',
		'z', 'x', 'c', 'v', 'b', 'n', 'm',
	}

	result := make([]rune, len(chars))

	// Place characters as far as possible from their QWERTY positions
	used := make(map[rune]bool)

	for i, qwertyChar := range qwerty {
		if i >= len(result) {
			break
		}

		// Find the character in our charset
		charIndex := -1

		for j, char := range chars {
			if char == qwertyChar {
				charIndex = j

				break
			}
		}

		if charIndex >= 0 && !used[qwertyChar] {
			// Place it as far as possible from its QWERTY position
			targetPos := (i + len(result)/2) % len(result)

			// If target position is taken, find next available
			for result[targetPos] != 0 {
				targetPos = (targetPos + 1) % len(result)
			}

			result[targetPos] = qwertyChar
			used[qwertyChar] = true
		}
	}

	// Fill remaining positions with remaining characters
	charIndex := 0

	for i, char := range result {
		if char == 0 { // Empty position
			// Find next unused character
			for charIndex < len(chars) && used[chars[charIndex]] {
				charIndex++
			}

			if charIndex < len(chars) {
				result[i] = chars[charIndex]
				used[chars[charIndex]] = true
				charIndex++
			}
		}
	}

	return result
}
