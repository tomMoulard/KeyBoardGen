package genetic

import (
	"math/rand"
	"sort"
)

// SelectionMethod defines different selection strategies.
type SelectionMethod int

const (
	TournamentSelection SelectionMethod = iota
	RouletteWheelSelection
	RankSelection
	EliteSelection
)

// Selector handles individual selection for reproduction.
type Selector struct {
	method SelectionMethod
	config Config
}

// NewSelector creates a new selector with the specified method.
func NewSelector(method SelectionMethod, config Config) *Selector {
	return &Selector{
		method: method,
		config: config,
	}
}

// Select chooses individuals from population for reproduction.
func (s *Selector) Select(population Population, count int) []Individual {
	switch s.method {
	case TournamentSelection:
		return s.tournamentSelect(population, count)
	case RouletteWheelSelection:
		return s.rouletteWheelSelect(population, count)
	case RankSelection:
		return s.rankSelect(population, count)
	case EliteSelection:
		return s.eliteSelect(population, count)
	default:
		return s.tournamentSelect(population, count)
	}
}

// tournamentSelect implements tournament selection.
func (s *Selector) tournamentSelect(population Population, count int) []Individual {
	selected := make([]Individual, 0, count)

	for range count {
		// Create tournament
		tournament := make([]Individual, s.config.TournamentSize)
		for j := range s.config.TournamentSize {
			idx := rand.Intn(len(population))
			tournament[j] = population[idx]
		}

		// Find best in tournament
		best := tournament[0]
		for _, individual := range tournament[1:] {
			if individual.Fitness > best.Fitness {
				best = individual
			}
		}

		selected = append(selected, best.Clone())
	}

	return selected
}

// rouletteWheelSelect implements fitness-proportionate selection.
func (s *Selector) rouletteWheelSelect(population Population, count int) []Individual {
	if len(population) == 0 {
		return []Individual{}
	}

	// Calculate fitness sum and handle negative fitness
	minFitness := population[0].Fitness
	for _, individual := range population {
		if individual.Fitness < minFitness {
			minFitness = individual.Fitness
		}
	}

	// Shift fitness values to be positive
	offset := 0.0
	if minFitness < 0 {
		offset = -minFitness + 1.0
	}

	// Calculate total adjusted fitness
	totalFitness := 0.0
	for _, individual := range population {
		totalFitness += individual.Fitness + offset
	}

	if totalFitness == 0 {
		// Fallback to random selection
		return s.randomSelect(population, count)
	}

	selected := make([]Individual, 0, count)

	for range count {
		r := rand.Float64() * totalFitness
		runningSum := 0.0

		for _, individual := range population {
			runningSum += individual.Fitness + offset
			if runningSum >= r {
				selected = append(selected, individual.Clone())

				break
			}
		}
	}

	return selected
}

// rankSelect implements rank-based selection.
func (s *Selector) rankSelect(population Population, count int) []Individual {
	if len(population) == 0 {
		return []Individual{}
	}

	// Sort population by fitness (descending)
	sorted := make(Population, len(population))
	copy(sorted, population)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Fitness > sorted[j].Fitness
	})

	// Calculate rank-based probabilities
	n := float64(len(sorted))
	totalRank := n * (n + 1) / 2 // Sum of ranks 1 to n

	selected := make([]Individual, 0, count)

	for range count {
		r := rand.Float64() * totalRank
		runningSum := 0.0

		for rank, individual := range sorted {
			// Higher rank (better fitness) gets higher probability
			rankWeight := n - float64(rank)
			runningSum += rankWeight

			if runningSum >= r {
				selected = append(selected, individual.Clone())

				break
			}
		}
	}

	return selected
}

// eliteSelect selects the best individuals.
func (s *Selector) eliteSelect(population Population, count int) []Individual {
	if len(population) == 0 {
		return []Individual{}
	}

	// Sort population by fitness (descending)
	sorted := make(Population, len(population))
	copy(sorted, population)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Fitness > sorted[j].Fitness
	})

	// Select top individuals
	selected := make([]Individual, 0, count)
	for i := 0; i < count && i < len(sorted); i++ {
		selected = append(selected, sorted[i].Clone())
	}

	// If we need more individuals than available, duplicate the best ones
	for len(selected) < count {
		idx := len(selected) % len(sorted)
		selected = append(selected, sorted[idx].Clone())
	}

	return selected
}

// randomSelect provides fallback random selection.
func (s *Selector) randomSelect(population Population, count int) []Individual {
	selected := make([]Individual, 0, count)

	for range count {
		idx := rand.Intn(len(population))
		selected = append(selected, population[idx].Clone())
	}

	return selected
}

// SelectParents selects two parents for crossover.
func (s *Selector) SelectParents(population Population) (Individual, Individual) {
	parents := s.Select(population, 2)
	if len(parents) < 2 {
		// Fallback: duplicate parent if population is too small
		parent := parents[0]

		return parent, parent
	}

	return parents[0], parents[1]
}

// SelectSurvivors implements generational replacement strategies.
func SelectSurvivors(oldPop, newPop Population, config Config) Population {
	combined := make(Population, 0, len(oldPop)+len(newPop))
	combined = append(combined, oldPop...)
	combined = append(combined, newPop...)

	// Sort by fitness (descending)
	sort.Slice(combined, func(i, j int) bool {
		return combined[i].Fitness > combined[j].Fitness
	})

	// Select top individuals for next generation
	survivors := make(Population, 0, config.PopulationSize)

	// Always include elite individuals
	eliteCount := config.ElitismCount
	if eliteCount > len(combined) {
		eliteCount = len(combined)
	}

	for i := range eliteCount {
		survivors = append(survivors, combined[i].Clone())
	}

	// Fill remaining slots with diverse selection
	remaining := config.PopulationSize - eliteCount
	if remaining > 0 {
		selector := NewSelector(TournamentSelection, config)

		remainingPop := combined[eliteCount:]
		if len(remainingPop) > 0 {
			selected := selector.Select(remainingPop, remaining)
			survivors = append(survivors, selected...)
		}
	}

	// Ensure we have exactly the right population size
	if len(survivors) > config.PopulationSize {
		survivors = survivors[:config.PopulationSize]
	}

	return survivors
}
