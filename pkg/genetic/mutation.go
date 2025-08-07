package genetic

import (
	"math/rand"
)

// MutationMethod defines different mutation strategies.
type MutationMethod int

const (
	SwapMutation MutationMethod = iota
	InsertionMutation
	InversionMutation
	ScrambleMutation
	DisplacementMutation
)

// Mutator handles genetic mutation operations.
type Mutator struct {
	method MutationMethod
	rate   float64
}

// NewMutator creates a new mutation operator.
func NewMutator(method MutationMethod, rate float64) *Mutator {
	return &Mutator{
		method: method,
		rate:   rate,
	}
}

// Apply performs mutation on an individual.
func (m *Mutator) Apply(individual Individual) Individual {
	// Check if mutation should occur
	if rand.Float64() > m.rate {
		return individual // No mutation
	}

	mutated := individual.Clone()

	switch m.method {
	case SwapMutation:
		m.swapMutation(&mutated)
	case InsertionMutation:
		m.insertionMutation(&mutated)
	case InversionMutation:
		m.inversionMutation(&mutated)
	case ScrambleMutation:
		m.scrambleMutation(&mutated)
	case DisplacementMutation:
		m.displacementMutation(&mutated)
	default:
		m.swapMutation(&mutated)
	}

	// Reset fitness as layout has changed
	mutated.Fitness = 0.0

	return mutated
}

// swapMutation swaps two random positions.
func (m *Mutator) swapMutation(individual *Individual) {
	length := len(individual.Layout)

	// Choose two random positions
	pos1 := rand.Intn(length)
	pos2 := rand.Intn(length)

	// Ensure different positions
	for pos1 == pos2 && length > 1 {
		pos2 = rand.Intn(length)
	}

	// Swap
	individual.Layout[pos1], individual.Layout[pos2] = individual.Layout[pos2], individual.Layout[pos1]
}

// insertionMutation removes element and inserts it elsewhere.
func (m *Mutator) insertionMutation(individual *Individual) {
	length := len(individual.Layout)
	if length < 2 {
		return
	}

	// Choose source and destination positions
	sourcePos := rand.Intn(length)
	destPos := rand.Intn(length)

	// Ensure different positions
	for sourcePos == destPos {
		destPos = rand.Intn(length)
	}

	// Remove element from source
	element := individual.Layout[sourcePos]

	if sourcePos < destPos {
		// Shift elements left
		for i := sourcePos; i < destPos; i++ {
			individual.Layout[i] = individual.Layout[i+1]
		}
	} else {
		// Shift elements right
		for i := sourcePos; i > destPos; i-- {
			individual.Layout[i] = individual.Layout[i-1]
		}
	}

	// Insert element at destination
	individual.Layout[destPos] = element
}

// inversionMutation reverses a random subsequence.
func (m *Mutator) inversionMutation(individual *Individual) {
	length := len(individual.Layout)
	if length < 2 {
		return
	}

	// Choose two random positions
	pos1 := rand.Intn(length)
	pos2 := rand.Intn(length)

	// Ensure pos1 < pos2
	if pos1 > pos2 {
		pos1, pos2 = pos2, pos1
	}

	// Reverse subsequence
	for pos1 < pos2 {
		individual.Layout[pos1], individual.Layout[pos2] = individual.Layout[pos2], individual.Layout[pos1]
		pos1++
		pos2--
	}
}

// scrambleMutation randomly shuffles a subsequence.
func (m *Mutator) scrambleMutation(individual *Individual) {
	length := len(individual.Layout)
	if length < 2 {
		return
	}

	// Choose two random positions
	pos1 := rand.Intn(length)
	pos2 := rand.Intn(length)

	// Ensure pos1 <= pos2
	if pos1 > pos2 {
		pos1, pos2 = pos2, pos1
	}

	// Extract subsequence
	subLen := pos2 - pos1 + 1

	subsequence := make([]rune, subLen)
	for i := range subLen {
		subsequence[i] = individual.Layout[pos1+i]
	}

	// Shuffle subsequence using Fisher-Yates
	for i := len(subsequence) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		subsequence[i], subsequence[j] = subsequence[j], subsequence[i]
	}

	// Put shuffled subsequence back
	for i := range subLen {
		individual.Layout[pos1+i] = subsequence[i]
	}
}

// displacementMutation moves a subsequence to a new position.
func (m *Mutator) displacementMutation(individual *Individual) {
	length := len(individual.Layout)
	if length < 3 {
		return
	}

	// Choose subsequence to move
	subStart := rand.Intn(length - 1)
	subEnd := subStart + rand.Intn(length-subStart)
	subLen := subEnd - subStart + 1

	// Choose new position (not overlapping with current)
	newPos := rand.Intn(length - subLen + 1)

	// Avoid overlapping positions
	if newPos >= subStart && newPos <= subEnd {
		return // Skip this mutation
	}

	// Extract subsequence
	subsequence := make([]rune, subLen)
	for i := range subLen {
		subsequence[i] = individual.Layout[subStart+i]
	}

	// Create new layout with same size as current layout
	newLayout := make([]rune, length)
	newIdx := 0

	// Add elements before new position (excluding moved subsequence)
	for i := range length {
		if i >= subStart && i <= subEnd {
			continue // Skip moved subsequence
		}

		if newIdx == newPos {
			// Insert moved subsequence
			for j := range subLen {
				newLayout[newIdx] = subsequence[j]
				newIdx++
			}
		}

		if newIdx < length {
			newLayout[newIdx] = individual.Layout[i]
			newIdx++
		}
	}

	// If subsequence wasn't inserted yet, add it at the end
	if newPos >= newIdx {
		for j := 0; j < subLen && newIdx < length; j++ {
			newLayout[newIdx] = subsequence[j]
			newIdx++
		}
	}

	individual.Layout = newLayout
}

// AdaptiveMutator adjusts mutation rate based on population diversity.
type AdaptiveMutator struct {
	baseMutator     *Mutator
	minRate         float64
	maxRate         float64
	diversityTarget float64
}

// NewAdaptiveMutator creates an adaptive mutation operator.
func NewAdaptiveMutator(method MutationMethod, minRate, maxRate, diversityTarget float64) *AdaptiveMutator {
	return &AdaptiveMutator{
		baseMutator:     NewMutator(method, minRate),
		minRate:         minRate,
		maxRate:         maxRate,
		diversityTarget: diversityTarget,
	}
}

// Apply performs adaptive mutation.
func (am *AdaptiveMutator) Apply(individual Individual, populationDiversity float64) Individual {
	// Adjust mutation rate based on diversity
	diversityRatio := populationDiversity / am.diversityTarget

	var adjustedRate float64
	if diversityRatio < 1.0 {
		// Low diversity - increase mutation rate
		adjustedRate = am.minRate + (am.maxRate-am.minRate)*(1.0-diversityRatio)
	} else {
		// High diversity - use minimum rate
		adjustedRate = am.minRate
	}

	// Apply mutation with adjusted rate
	if rand.Float64() <= adjustedRate {
		am.baseMutator.rate = adjustedRate

		return am.baseMutator.Apply(individual)
	}

	return individual
}

// CalculatePopulationDiversity measures genetic diversity in population.
func CalculatePopulationDiversity(population Population) float64 {
	if len(population) < 2 {
		return 0.0
	}

	totalDistance := 0.0
	comparisons := 0

	// Calculate average pairwise distance
	for i := range population {
		for j := i + 1; j < len(population); j++ {
			distance := calculateLayoutDistance(population[i], population[j])
			totalDistance += distance
			comparisons++
		}
	}

	if comparisons == 0 {
		return 0.0
	}

	return totalDistance / float64(comparisons)
}

// calculateLayoutDistance computes distance between two keyboard layouts.
func calculateLayoutDistance(ind1, ind2 Individual) float64 {
	differences := 0

	for i := range len(ind1.Layout) {
		if ind1.Layout[i] != ind2.Layout[i] {
			differences++
		}
	}

	return float64(differences) / float64(len(ind1.Layout))
}

// MultiMutator applies multiple mutation operators with different probabilities.
type MultiMutator struct {
	mutators []struct {
		mutator *Mutator
		weight  float64
	}
	totalWeight float64
}

// NewMultiMutator creates a multi-mutation operator.
func NewMultiMutator() *MultiMutator {
	return &MultiMutator{
		mutators: make([]struct {
			mutator *Mutator
			weight  float64
		}, 0),
	}
}

// AddMutator adds a mutation operator with weight.
func (mm *MultiMutator) AddMutator(method MutationMethod, rate, weight float64) {
	mutator := NewMutator(method, rate)
	mm.mutators = append(mm.mutators, struct {
		mutator *Mutator
		weight  float64
	}{mutator, weight})
	mm.totalWeight += weight
}

// Apply randomly selects and applies one of the mutation operators.
func (mm *MultiMutator) Apply(individual Individual) Individual {
	if len(mm.mutators) == 0 || mm.totalWeight == 0 {
		return individual
	}

	// Select mutator based on weights
	r := rand.Float64() * mm.totalWeight
	runningWeight := 0.0

	for _, m := range mm.mutators {
		runningWeight += m.weight
		if r <= runningWeight {
			return m.mutator.Apply(individual)
		}
	}

	// Fallback to last mutator
	return mm.mutators[len(mm.mutators)-1].mutator.Apply(individual)
}
