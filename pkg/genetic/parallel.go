package genetic

import (
	"context"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
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
		// Create adaptive mutator for diversity maintenance
		adaptiveMutator = NewAdaptiveMutator(SwapMutation, config.MutationRate, config.MutationRate*3, 0.3)
	}

	return &ParallelEvolver{
		evaluator:               NewParallelEvaluator(evaluator, config.ParallelWorkers),
		selector:                NewSelector(TournamentSelection, config),
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

	// Preserve elite individuals (but ensure some diversity)
	elites := pe.selector.Select(population, config.ElitismCount)
	if pe.useDiversityMaintenance && diversity < 0.2 {
		// Low diversity: reduce elitism and add some random individuals
		eliteCount := max(1, config.ElitismCount/2)
		elites = elites[:eliteCount]

		// Add some random individuals to boost diversity
		for range config.ElitismCount - eliteCount {
			randomInd := NewRandomIndividual()
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
}

// NewParallelGA creates a new parallel genetic algorithm.
func NewParallelGA(evaluator FitnessEvaluator, config Config) *ParallelGA {
	return &ParallelGA{
		evolver: NewParallelEvolver(evaluator, config),
		config:  config,
	}
}

// Run executes the genetic algorithm.
func (pga *ParallelGA) Run(ctx context.Context, data KeyloggerDataInterface, callback func(generation int, best Individual)) (Individual, error) {
	// Initialize population
	population := make(Population, pga.config.PopulationSize)
	for i := range population {
		population[i] = NewRandomIndividual()
	}

	// Evaluate initial population fitness
	err := pga.evolver.evaluator.EvaluatePopulation(ctx, population, data)
	if err != nil {
		return Individual{}, err
	}

	var bestIndividual Individual

	bestFitness := -1.0
	bestInitialized := false

	for generation := range pga.config.MaxGenerations {
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

// initRandom initializes random seed for parallel GA
func init() {
	rand.Seed(time.Now().UnixNano())
}

// WorkerPool manages a pool of workers for various GA tasks.
type WorkerPool struct {
	workerCount int
	jobs        chan func()
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewWorkerPool creates a new worker pool.
func NewWorkerPool(workerCount int) *WorkerPool {
	if workerCount <= 0 {
		workerCount = runtime.NumCPU()
	}

	ctx, cancel := context.WithCancel(context.Background())

	wp := &WorkerPool{
		workerCount: workerCount,
		jobs:        make(chan func(), workerCount*2),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Start workers
	for range workerCount {
		wp.wg.Add(1)

		go wp.worker()
	}

	return wp
}

// Submit adds a job to the worker pool.
func (wp *WorkerPool) Submit(job func()) {
	select {
	case wp.jobs <- job:
	case <-wp.ctx.Done():
	}
}

// Close shuts down the worker pool.
func (wp *WorkerPool) Close() {
	wp.cancel()
	close(wp.jobs)
	wp.wg.Wait()
}

// worker processes jobs from the queue.
func (wp *WorkerPool) worker() {
	defer wp.wg.Done()

	for {
		select {
		case job, ok := <-wp.jobs:
			if !ok {
				return // Channel closed
			}

			job()
		case <-wp.ctx.Done():
			return
		}
	}
}
