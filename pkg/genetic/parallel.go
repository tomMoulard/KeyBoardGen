package genetic

import (
	"context"
	"math"
	"math/rand"
	"runtime"
	"sync"
)

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

// ParallelEvaluator handles concurrent fitness evaluation.
type ParallelEvaluator struct {
	workerCount int
	evaluator   FitnessEvaluator
}

// FitnessEvaluator interface for fitness evaluation.
type FitnessEvaluator interface {
	Evaluate(layout []rune, charset *CharacterSet, data KeyloggerDataInterface) float64
	// Legacy method for backward compatibility
	EvaluateLegacy(layout [26]rune, data KeyloggerDataInterface) float64
}

// KeyloggerDataInterface provides thread-safe access to keylogger data.
type KeyloggerDataInterface interface {
	GetCharFreq(char rune) int
	GetBigramFreq(bigram string) int
	GetTrigramFreq(trigram string) int
	GetTotalChars() int
	GetAllBigrams() map[string]int
}

// NewParallelEvaluator creates a new parallel fitness evaluator.
func NewParallelEvaluator(evaluator FitnessEvaluator, workerCount int) *ParallelEvaluator {
	if workerCount <= 0 {
		workerCount = runtime.NumCPU()
	}

	return &ParallelEvaluator{
		workerCount: workerCount,
		evaluator:   evaluator,
	}
}

// EvaluatePopulation evaluates fitness for entire population in parallel.
func (pe *ParallelEvaluator) EvaluatePopulation(ctx context.Context, population Population, data KeyloggerDataInterface) error {
	if len(population) == 0 {
		return nil
	}

	// Create job queue
	jobs := make(chan int, len(population))
	results := make(chan struct {
		index   int
		fitness float64
		err     error
	}, len(population))

	// Start workers
	var wg sync.WaitGroup
	for range pe.workerCount {
		wg.Add(1)

		go pe.worker(ctx, &wg, jobs, results, population, data)
	}

	// Send jobs
	go func() {
		defer close(jobs)

		for i := range population {
			select {
			case jobs <- i:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	for result := range results {
		if result.err != nil {
			return result.err
		}

		population[result.index].Fitness = result.fitness
	}

	return ctx.Err()
}

// worker processes fitness evaluation jobs.
func (pe *ParallelEvaluator) worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan int, results chan<- struct {
	index   int
	fitness float64
	err     error
}, population Population, data KeyloggerDataInterface,
) {
	defer wg.Done()

	for {
		select {
		case index, ok := <-jobs:
			if !ok {
				return // Channel closed
			}

			// Evaluate fitness
			fitness := pe.evaluator.Evaluate(population[index].Layout, population[index].Charset, data)

			// Send result
			select {
			case results <- struct {
				index   int
				fitness float64
				err     error
			}{index, fitness, nil}:
			case <-ctx.Done():
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

// ParallelEvolver handles parallel genetic algorithm operations.
type ParallelEvolver struct {
	evaluator               *ParallelEvaluator
	selector                *Selector
	crossover               *Crossover
	mutator                 *Mutator
	adaptiveMutator         *AdaptiveMutator
	workerCount             int
	useDiversityMaintenance bool
}

// NewParallelEvolver creates a new parallel genetic algorithm evolver.
func NewParallelEvolver(evaluator FitnessEvaluator, config Config) *ParallelEvolver {
	// Enable diversity maintenance for large populations
	useDiversityMaintenance := config.PopulationSize >= 200

	var adaptiveMutator *AdaptiveMutator

	if useDiversityMaintenance {
		// Create adaptive mutator for diversity maintenance with higher rates for large populations
		baseRate := config.MutationRate
		maxRate := config.MutationRate * 4 // More aggressive mutation for large populations
		threshold := 0.25                  // Higher threshold for large populations
		adaptiveMutator = NewAdaptiveMutator(SwapMutation, baseRate, maxRate, threshold)
	}

	// Adjust tournament size for large populations to reduce selection pressure
	adaptiveConfig := config
	if useDiversityMaintenance && config.PopulationSize >= 200 {
		// Use smaller tournament size for large populations to increase diversity
		adaptiveConfig.TournamentSize = max(2, min(4, config.TournamentSize))
	}

	return &ParallelEvolver{
		evaluator:               NewParallelEvaluator(evaluator, config.ParallelWorkers),
		selector:                NewSelector(TournamentSelection, adaptiveConfig),
		crossover:               NewCrossover(OrderCrossover),
		mutator:                 NewMutator(SwapMutation, config.MutationRate),
		adaptiveMutator:         adaptiveMutator,
		workerCount:             config.ParallelWorkers,
		useDiversityMaintenance: useDiversityMaintenance,
	}
}

// Evolve performs one generation of evolution in parallel.
func (pe *ParallelEvolver) Evolve(ctx context.Context, population Population, config Config, data KeyloggerDataInterface) (Population, error) {
	// Evaluate current population fitness
	err := pe.evaluator.EvaluatePopulation(ctx, population, data)
	if err != nil {
		return nil, err
	}

	// Calculate population diversity if using diversity maintenance
	var diversity float64
	if pe.useDiversityMaintenance {
		diversity = CalculatePopulationDiversity(population)
	}

	// Create next generation
	newPopulation := make(Population, 0, config.PopulationSize)

	// Adaptive elitism based on population size and diversity
	adaptiveElitismCount := config.ElitismCount
	if pe.useDiversityMaintenance {
		// For large populations, use percentage-based elitism (1-3% of population)
		minElitism := max(1, config.PopulationSize/100) // 1% minimum
		maxElitism := max(3, config.PopulationSize/33)  // 3% maximum
		adaptiveElitismCount = max(minElitism, min(maxElitism, config.ElitismCount))
	}

	elites := pe.selector.Select(population, adaptiveElitismCount)

	// Only inject random individuals if diversity is critically low
	if pe.useDiversityMaintenance && diversity < 0.1 {
		// Very low diversity: replace some elites with evaluated random individuals
		eliteCount := max(1, adaptiveElitismCount*2/3) // Keep 2/3 of elites
		elites = elites[:eliteCount]

		// Add some random individuals to boost diversity (but evaluate them first)
		for range adaptiveElitismCount - eliteCount {
			randomInd := NewRandomIndividual()
			// Evaluate the random individual before adding to elites
			randomInd.Fitness = pe.evaluator.evaluator.Evaluate(randomInd.Layout, randomInd.Charset, data)
			elites = append(elites, randomInd)
		}
	}

	newPopulation = append(newPopulation, elites...)

	// Generate offspring in parallel
	remaining := config.PopulationSize - len(elites)
	if remaining > 0 {
		offspring, err := pe.generateOffspring(ctx, population, remaining, config, data, diversity)
		if err != nil {
			return nil, err
		}

		newPopulation = append(newPopulation, offspring...)
	}

	return newPopulation, nil
}

// generateOffspring creates new individuals through crossover and mutation.
func (pe *ParallelEvolver) generateOffspring(ctx context.Context, population Population, count int, config Config, data KeyloggerDataInterface, diversity float64) (Population, error) {
	// Create job queue for offspring generation
	jobs := make(chan int, count)
	results := make(chan struct {
		index      int
		individual Individual
		err        error
	}, count)

	// Start workers
	var wg sync.WaitGroup
	for range pe.workerCount {
		wg.Add(1)

		go pe.offspringWorker(ctx, &wg, jobs, results, population, config, diversity)
	}

	// Send jobs
	go func() {
		defer close(jobs)

		for i := range count {
			select {
			case jobs <- i:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	offspring := make(Population, count)

	for result := range results {
		if result.err != nil {
			return nil, result.err
		}

		offspring[result.index] = result.individual
	}

	// Evaluate offspring fitness in parallel
	err := pe.evaluator.EvaluatePopulation(ctx, offspring, data)
	if err != nil {
		return nil, err
	}

	return offspring, nil
}

// offspringWorker generates individual offspring.
func (pe *ParallelEvolver) offspringWorker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan int, results chan<- struct {
	index      int
	individual Individual
	err        error
}, population Population, config Config, diversity float64,
) {
	defer wg.Done()

	for {
		select {
		case index, ok := <-jobs:
			if !ok {
				return // Channel closed
			}

			// Select parents
			parent1, parent2 := pe.selector.SelectParents(population)

			// Apply crossover
			var child Individual
			if rand.Float64() < config.CrossoverRate {
				child = pe.crossover.Apply(parent1, parent2)
			} else {
				// No crossover - clone a parent
				if rand.Float64() < 0.5 {
					child = parent1.Clone()
				} else {
					child = parent2.Clone()
				}
			}

			// Apply mutation (adaptive if enabled)
			if pe.useDiversityMaintenance && pe.adaptiveMutator != nil {
				child = pe.adaptiveMutator.Apply(child, diversity)
			} else {
				child = pe.mutator.Apply(child)
			}

			// Validate child
			child = ValidateChild(child)

			// Send result
			select {
			case results <- struct {
				index      int
				individual Individual
				err        error
			}{index, child, nil}:
			case <-ctx.Done():
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

// ParallelGA implements a complete parallel genetic algorithm.
type ParallelGA struct {
	evolver *ParallelEvolver
	config  Config
	charset *CharacterSet
}

// NewParallelGA creates a new parallel genetic algorithm.
func NewParallelGA(evaluator FitnessEvaluator, config Config, charset *CharacterSet) *ParallelGA {
	return &ParallelGA{
		evolver: NewParallelEvolver(evaluator, config),
		config:  config,
		charset: charset,
	}
}

// CreateDiversePopulation creates a population using multiple initialization strategies.
func CreateDiversePopulation(size int, data KeyloggerDataInterface, charset *CharacterSet) Population {
	population := make(Population, size)

	strategies := []InitializationStrategy{
		FrequencyBased,
		HandBalance,
		RowBalance,
		CommonPatternsFirst,
		AntiQWERTY,
	}

	// For large populations, use fewer diverse individuals to avoid local optima trapping
	diverseRatio := 0.5 // 50% diverse individuals
	if size >= 500 {
		diverseRatio = 0.3 // Only 30% for very large populations
	} else if size >= 200 {
		diverseRatio = 0.4 // 40% for medium-large populations
	}

	diverseCount := int(float64(size) * diverseRatio)

	// Create diverse individuals
	for i := range diverseCount {
		strategyIndex := i % len(strategies)
		strategy := strategies[strategyIndex]

		// Use data-aware strategies when data is available
		if data != nil && (strategy == FrequencyBased || strategy == HandBalance || strategy == CommonPatternsFirst) {
			population[i] = NewRandomIndividualWithStrategy(charset, strategy, data)
		} else {
			population[i] = NewRandomIndividualWithStrategy(charset, strategy, nil)
		}
	}

	// Fill the rest with random individuals for exploration
	for i := diverseCount; i < size; i++ {
		population[i] = NewRandomIndividualWithCharset(charset)
	}

	return population
}

// Run executes the genetic algorithm.
func (pga *ParallelGA) Run(ctx context.Context, data KeyloggerDataInterface, callback func(generation int, best Individual)) (Individual, error) {
	// Initialize population
	population := CreateDiversePopulation(pga.config.PopulationSize, data, pga.charset)

	// Evaluate initial population fitness
	err := pga.evolver.evaluator.EvaluatePopulation(ctx, population, data)
	if err != nil {
		return Individual{}, err
	}

	var bestIndividual Individual

	bestFitness := -1.0
	bestInitialized := false

	// Convergence tracking
	lastBestFitness := -1.0

	convergenceCount := 0

	// Determine loop limit: use MaxGenerations if > 0, otherwise unlimited (but with convergence)
	maxGens := pga.config.MaxGenerations
	if maxGens == 0 && pga.config.ConvergenceStops > 0 {
		maxGens = int(^uint(0) >> 1) // Max int value for unlimited generations
	}

	for generation := range maxGens {
		select {
		case <-ctx.Done():
			return bestIndividual, ctx.Err()
		default:
		}

		// Evolve population
		var err error

		population, err = pga.evolver.Evolve(ctx, population, pga.config, data)
		if err != nil {
			return bestIndividual, err
		}

		// Find best individual in current generation
		for _, individual := range population {
			// Initialize bestIndividual with first evaluated individual
			if !bestInitialized {
				bestIndividual = individual.Clone()
				bestFitness = individual.Fitness
				bestInitialized = true
			} else if individual.Fitness > bestFitness {
				bestFitness = individual.Fitness
				bestIndividual = individual.Clone()
				bestIndividual.Age = generation // Update age when we find a better individual
			}
		}

		// Check for convergence based on fitness stagnation
		if pga.config.ConvergenceStops > 0 {
			if lastBestFitness >= 0 { // Not the first generation
				fitnessChange := math.Abs(bestFitness - lastBestFitness)
				if fitnessChange <= pga.config.ConvergenceTolerance {
					convergenceCount++
					if convergenceCount >= pga.config.ConvergenceStops {
						// Fitness has not improved for ConvergenceStops generations
						if callback != nil {
							callback(generation, bestIndividual) // Final callback
						}

						break
					}
				} else {
					// Fitness improved significantly, reset counter
					convergenceCount = 0
				}
			}

			lastBestFitness = bestFitness
		}

		// Call callback if provided
		if callback != nil {
			callback(generation, bestIndividual)
		}

		// Check for convergence (optional early stopping)
		if bestFitness >= 0.99 { // Near-perfect fitness
			break
		}
	}

	return bestIndividual, nil
}
