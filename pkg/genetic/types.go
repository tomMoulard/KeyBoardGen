package genetic

import (
	"math/rand"
	"sync"
	"time"
)

func init() {
	// Initialize random seed once at package startup
	rand.Seed(time.Now().UnixNano())
}

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

// NewRandomIndividual creates a random individual with shuffled alphabet (legacy function).
func NewRandomIndividual() Individual {
	return NewRandomIndividualWithCharset(AlphabetOnly())
}

// NewRandomIndividualWithCharset creates a random individual using the specified character set.
func NewRandomIndividualWithCharset(charset *CharacterSet) Individual {
	if charset == nil {
		charset = AlphabetOnly()
	}

	// Copy characters for shuffling
	chars := make([]rune, len(charset.Characters))
	copy(chars, charset.Characters)

	// Fisher-Yates shuffle
	for i := len(chars) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		chars[i], chars[j] = chars[j], chars[i]
	}

	return Individual{
		Layout:  chars,
		Charset: charset,
		Fitness: 0.0,
		Age:     0,
	}
}
