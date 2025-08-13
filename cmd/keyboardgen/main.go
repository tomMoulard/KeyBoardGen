package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tommoulard/keyboardgen/internal/runner"
	"github.com/tommoulard/keyboardgen/pkg/config"
)

func main() {
	// Parse command line flags
	cfg := parseFlags()

	// Load configuration file if specified
	if cfg.ConfigFile != "" {
		loadedCfg, err := config.LoadFromFile(cfg.ConfigFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		cfg = loadedCfg
	}

	// Validate configuration
	err := cfg.Validate()
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
	err = runOptimization(ctx, cfg)
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
func parseFlags() config.Config {
	cfg := config.Default()

	flag.StringVar(&cfg.InputFile, "input", "", "Input keylogger file (required)")
	flag.StringVar(&cfg.OutputFile, "output", cfg.OutputFile, "Output file for best layout")
	flag.StringVar(&cfg.ConfigFile, "config", "", "Configuration file (JSON)")
	flag.IntVar(&cfg.PopulationSize, "population", cfg.PopulationSize, "Population size")
	flag.IntVar(&cfg.MaxGeneration, "generations", cfg.MaxGeneration, "Maximum generations")
	flag.Float64Var(&cfg.MutationRate, "mutation", cfg.MutationRate, "Mutation rate")
	flag.Float64Var(&cfg.CrossoverRate, "crossover", cfg.CrossoverRate, "Crossover rate")
	flag.IntVar(&cfg.ElitismCount, "elitism", cfg.ElitismCount, "Number of elite individuals")
	flag.IntVar(&cfg.WorkerCount, "workers", cfg.WorkerCount, "Number of parallel workers (0=auto)")
	flag.BoolVar(&cfg.Verbose, "verbose", cfg.Verbose, "Verbose output")
	flag.BoolVar(&cfg.ShowProgress, "progress", cfg.ShowProgress, "Show progress")
	flag.IntVar(&cfg.SaveInterval, "save-interval", cfg.SaveInterval, "Save best layout every N generations")
	flag.IntVar(&cfg.ConvergenceStops, "convergence-stops", cfg.ConvergenceStops, "Stop after N generations with same fitness (0=disabled, overrides max generations)")
	flag.Float64Var(&cfg.ConvergenceTolerance, "convergence-tolerance", cfg.ConvergenceTolerance, "Fitness difference tolerance for convergence detection")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "KeyBoardGen - Genetic Algorithm Keyboard Layout Optimizer\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s -input keylog.txt -generations 500 -population 200\n", os.Args[0])
	}

	flag.Parse()

	return cfg
}

// runOptimization executes the genetic algorithm using the runner.
func runOptimization(ctx context.Context, cfg config.Config) error {
	// Create runner
	r, err := runner.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create runner: %w", err)
	}

	// Start timing
	startTime := time.Now()

	// Run optimization
	bestIndividual, fitnessHistory, err := r.RunFromFile(ctx, nil)
	if err != nil {
		return err
	}

	// Calculate total time
	totalTime := time.Since(startTime)

	// Print results
	r.PrintResults(bestIndividual, fitnessHistory, totalTime)

	// Save to file
	if err := r.SaveLayout(bestIndividual, cfg.OutputFile); err != nil {
		return fmt.Errorf("failed to save layout: %w", err)
	}

	fmt.Printf("\nLayout saved to: %s\n", cfg.OutputFile)

	return nil
}
