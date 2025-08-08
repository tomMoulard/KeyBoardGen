package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/tommoulard/keyboardgen/pkg/display"
	"github.com/tommoulard/keyboardgen/pkg/fitness"
	"github.com/tommoulard/keyboardgen/pkg/genetic"
	"github.com/tommoulard/keyboardgen/pkg/parser"
)

// Config holds application configuration.
type Config struct {
	InputFile      string  `json:"input_file"`
	OutputFile     string  `json:"output_file"`
	ConfigFile     string  `json:"config_file"`
	CharacterSet   string  `json:"character_set"` // "alphabet", "alphanumeric", "programming", "full"
	PopulationSize int     `json:"population_size"`
	MaxGeneration  int     `json:"max_generation"`
	MutationRate   float64 `json:"mutation_rate"`
	CrossoverRate  float64 `json:"crossover_rate"`
	ElitismCount   int     `json:"elitism_count"`
	WorkerCount    int     `json:"worker_count"`
	Verbose        bool    `json:"verbose"`
	ShowProgress   bool    `json:"show_progress"`
	SaveInterval   int     `json:"save_interval"`
	DiverseInit    bool    `json:"diverse_init"` // Use diverse initialization strategies
}

// DefaultAppConfig returns default application configuration.
func DefaultAppConfig() Config {
	return Config{
		InputFile:      "",
		OutputFile:     "best_layout.json",
		ConfigFile:     "",
		CharacterSet:   "alphabet",
		PopulationSize: 100,
		MaxGeneration:  1000,
		MutationRate:   0.1,
		CrossoverRate:  0.8,
		ElitismCount:   5,
		WorkerCount:    0, // Auto-detect
		Verbose:        false,
		ShowProgress:   true,
		SaveInterval:   50,
		DiverseInit:    true, // Enable by default for better evolution
	}
}

func main() {
	// Parse command line flags
	config := parseFlags()

	// Load configuration file if specified
	if config.ConfigFile != "" {
		err := loadConfig(config.ConfigFile, &config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
	}

	// Validate configuration
	err := validateConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Set up context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal, shutting down gracefully...")
		cancel()
	}()

	// Run the genetic algorithm
	err = runGA(ctx, config)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			fmt.Println("Operation canceled by user")
			os.Exit(130) // Standard exit code for SIGINT
		}

		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// parseFlags parses command line arguments.
func parseFlags() Config {
	config := DefaultAppConfig()

	flag.StringVar(&config.InputFile, "input", "", "Input keylogger file (required)")
	flag.StringVar(&config.OutputFile, "output", config.OutputFile, "Output file for best layout")
	flag.StringVar(&config.ConfigFile, "config", "", "Configuration file (JSON)")
	flag.StringVar(&config.CharacterSet, "charset", config.CharacterSet, "Character set: alphabet, alphanumeric, programming, full")
	flag.IntVar(&config.PopulationSize, "population", config.PopulationSize, "Population size")
	flag.IntVar(&config.MaxGeneration, "generations", config.MaxGeneration, "Maximum generations")
	flag.Float64Var(&config.MutationRate, "mutation", config.MutationRate, "Mutation rate")
	flag.Float64Var(&config.CrossoverRate, "crossover", config.CrossoverRate, "Crossover rate")
	flag.IntVar(&config.ElitismCount, "elitism", config.ElitismCount, "Number of elite individuals")
	flag.IntVar(&config.WorkerCount, "workers", config.WorkerCount, "Number of parallel workers (0=auto)")
	flag.BoolVar(&config.Verbose, "verbose", config.Verbose, "Verbose output")
	flag.BoolVar(&config.ShowProgress, "progress", config.ShowProgress, "Show progress")
	flag.IntVar(&config.SaveInterval, "save-interval", config.SaveInterval, "Save best layout every N generations")
	flag.BoolVar(&config.DiverseInit, "diverse-init", config.DiverseInit, "Use diverse initialization strategies for broader starting population")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "KeyBoardGen - Genetic Algorithm Keyboard Layout Optimizer\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s -input keylog.txt -generations 500 -population 200\n", os.Args[0])
	}

	flag.Parse()

	return config
}

// loadConfig loads configuration from JSON file.
func loadConfig(filename string, config *Config) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	return json.Unmarshal(data, config)
}

// validateConfig validates the configuration.
func validateConfig(config Config) error {
	if config.InputFile == "" {
		return errors.New("input file is required")
	}

	if _, err := os.Stat(config.InputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", config.InputFile)
	}

	if config.PopulationSize < 10 {
		return errors.New("population size must be at least 10")
	}

	if config.MaxGeneration < 1 {
		return errors.New("max generations must be at least 1")
	}

	if config.MutationRate < 0 || config.MutationRate > 1 {
		return errors.New("mutation rate must be between 0 and 1")
	}

	if config.CrossoverRate < 0 || config.CrossoverRate > 1 {
		return errors.New("crossover rate must be between 0 and 1")
	}

	if config.ElitismCount < 0 || config.ElitismCount >= config.PopulationSize {
		return errors.New("elitism count must be between 0 and population size")
	}

	return nil
}

// runGA executes the genetic algorithm.
func runGA(ctx context.Context, appConfig Config) error {
	// Parse keylogger data
	fmt.Printf("Parsing keylogger data from %s...\n", appConfig.InputFile)

	file, err := os.Open(appConfig.InputFile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer file.Close()

	klparser := parser.NewKeyloggerParser()

	// Choose parser configuration based on character set
	var parseConfig parser.ParseConfig

	switch appConfig.CharacterSet {
	case "programming":
		parseConfig = parser.ProgrammingConfig()
	case "full", "full_keyboard":
		parseConfig = parser.FullKeyboardConfig()
	default:
		parseConfig = parser.DefaultConfig()
	}

	// Set character set in config
	parseConfig.CharacterSet = appConfig.CharacterSet

	// Detect format based on file extension or content
	if strings.HasSuffix(strings.ToLower(appConfig.InputFile), ".json") {
		parseConfig.Format = parser.JSONFormat
	} else if strings.Contains(strings.ToLower(filepath.Base(appConfig.InputFile)), "vim") {
		parseConfig.Format = parser.VimCommandFormat
	}

	keyloggerData, err := klparser.Parse(file, parseConfig)
	if err != nil {
		return fmt.Errorf("failed to parse keylogger data: %w", err)
	}

	if err := keyloggerData.Validate(); err != nil {
		return fmt.Errorf("insufficient keylogger data: %w", err)
	}

	fmt.Printf("Parsed %d characters, %d unique bigrams\n",
		keyloggerData.TotalChars, len(keyloggerData.BigramFreq))

	// Show most frequent characters and bigrams if verbose
	if appConfig.Verbose {
		fmt.Println("\nMost frequent characters:")

		for i, cf := range keyloggerData.GetMostFrequentChars(10) {
			fmt.Printf("%2d. %c: %d (%.2f%%)\n", i+1, cf.Char, cf.Freq,
				float64(cf.Freq)*100/float64(keyloggerData.TotalChars))
		}

		fmt.Println("\nMost frequent bigrams:")

		for i, bf := range keyloggerData.GetMostFrequentBigrams(10) {
			fmt.Printf("%2d. %s: %d\n", i+1, bf.Bigram, bf.Freq)
		}

		fmt.Println()
	}

	// Set up fitness evaluation with appropriate geometry for character set
	charset := genetic.GetCharsetByName(appConfig.CharacterSet)
	geometry := fitness.GetGeometryForCharset(charset)
	weights := fitness.DefaultWeights()
	fitnessEvaluator := fitness.NewFitnessEvaluator(geometry, weights)

	// Set up genetic algorithm with adaptive configuration
	var gaConfig genetic.Config

	// Use adaptive configuration if user hasn't specified custom parameters
	if appConfig.PopulationSize == 100 && appConfig.MaxGeneration == 1000 &&
		appConfig.MutationRate == 0.1 && appConfig.CrossoverRate == 0.8 && appConfig.ElitismCount == 5 {
		// User is using defaults, apply adaptive configuration
		gaConfig = genetic.AdaptiveConfig(keyloggerData.TotalChars)
		fmt.Printf("Using adaptive configuration for dataset size: %d characters\n", keyloggerData.TotalChars)
	} else {
		// User has customized parameters, respect their choices
		gaConfig = genetic.Config{
			PopulationSize:  appConfig.PopulationSize,
			MaxGenerations:  appConfig.MaxGeneration,
			MutationRate:    appConfig.MutationRate,
			CrossoverRate:   appConfig.CrossoverRate,
			ElitismCount:    appConfig.ElitismCount,
			TournamentSize:  3,
			ParallelWorkers: appConfig.WorkerCount,
		}

		fmt.Println("Using custom configuration")
	}

	ga := genetic.NewParallelGA(fitnessEvaluator, gaConfig)

	fmt.Printf("Starting genetic algorithm with:\n")
	fmt.Printf("- Population size: %d\n", gaConfig.PopulationSize)
	fmt.Printf("- Max generations: %d\n", gaConfig.MaxGenerations)
	fmt.Printf("- Mutation rate: %.2f\n", gaConfig.MutationRate)
	fmt.Printf("- Crossover rate: %.2f\n", gaConfig.CrossoverRate)
	fmt.Printf("- Elite count: %d\n", gaConfig.ElitismCount)
	fmt.Printf("- Parallel workers: %d\n", gaConfig.ParallelWorkers)
	fmt.Println()

	// Progress tracking
	startTime := time.Now()

	var lastSavedGeneration int

	// Run genetic algorithm
	bestIndividual, err := ga.RunWithDiverseInit(ctx, keyloggerData, func(generation int, best genetic.Individual) {
		if appConfig.ShowProgress {
			elapsed := time.Since(startTime)
			avgTime := elapsed / time.Duration(generation+1)
			remaining := avgTime * time.Duration(gaConfig.MaxGenerations-generation-1)

			fmt.Printf("Generation %4d: Best fitness = %.6f (ETA: %v)\n",
				generation, best.Fitness, remaining.Round(time.Second))
		}

		// Save intermediate results
		if appConfig.Verbose &&
			appConfig.SaveInterval > 0 &&
			generation-lastSavedGeneration >= appConfig.SaveInterval {
			saveLayout(best, fmt.Sprintf("%s.gen%d", appConfig.OutputFile, generation))
			lastSavedGeneration = generation
		}
	}, appConfig.DiverseInit)
	if err != nil {
		return fmt.Errorf("genetic algorithm failed: %w", err)
	}

	// Save final result
	fmt.Printf("\nOptimization complete!\n")
	fmt.Printf("Best fitness: %.6f\n", bestIndividual.Fitness)
	fmt.Printf("Total time: %v\n", time.Since(startTime).Round(time.Second))

	// Create display handler
	kbDisplay := display.NewKeyboardDisplay()

	// Show enhanced summary first
	kbDisplay.PrintSummary(bestIndividual, keyloggerData, fitnessEvaluator)

	// Basic layout display
	fmt.Printf("\n\033[1;34mOPTIMIZED KEYBOARD LAYOUT:\033[0m\n")
	kbDisplay.SetOptions(false, false, false)
	kbDisplay.PrintLayout(bestIndividual, keyloggerData)

	// Print comprehensive statistics
	kbDisplay.PrintStatistics(bestIndividual, keyloggerData)

	// Always show comparison with QWERTY
	kbDisplay.PrintComparisonWithEvaluator(bestIndividual, keyloggerData, fitnessEvaluator)

	// Show heatmap if verbose
	if appConfig.Verbose {
		kbDisplay.PrintHeatmap(bestIndividual, keyloggerData)
	}

	// Save to file
	if err := saveLayout(bestIndividual, appConfig.OutputFile); err != nil {
		return fmt.Errorf("failed to save layout: %w", err)
	}

	fmt.Printf("\nLayout saved to: %s\n", appConfig.OutputFile)

	return nil
}

// saveLayout saves the layout to a JSON file.
func saveLayout(individual genetic.Individual, filename string) error {
	// Create layout mapping
	layout := make(map[string]any)
	layout["fitness"] = individual.Fitness
	layout["age"] = individual.Age
	layout["layout"] = string(individual.Layout)
	layout["timestamp"] = time.Now().Format(time.RFC3339)

	// Create position mapping
	positions := make(map[string]int)
	for i, char := range individual.Layout {
		positions[string(char)] = i
	}

	layout["positions"] = positions

	// Marshal to JSON
	data, err := json.MarshalIndent(layout, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal layout: %w", err)
	}

	// Write to file
	return os.WriteFile(filename, data, 0o644)
}
