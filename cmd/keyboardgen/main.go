package main

import (
	"bytes"
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
	PopulationSize int     `json:"population_size"`
	MaxGeneration  int     `json:"max_generation"`
	MutationRate   float64 `json:"mutation_rate"`
	CrossoverRate  float64 `json:"crossover_rate"`
	ElitismCount   int     `json:"elitism_count"`
	WorkerCount    int     `json:"worker_count"`
	Verbose        bool    `json:"verbose"`
	ShowProgress   bool    `json:"show_progress"`
	SaveInterval   int     `json:"save_interval"`
	// Convergence-based stopping
	ConvergenceStops     int     `json:"convergence_stops"`
	ConvergenceTolerance float64 `json:"convergence_tolerance"`
}

// DefaultAppConfig returns default application configuration.
func DefaultAppConfig() Config {
	return Config{
		InputFile:      "",
		OutputFile:     "best_layout.json",
		ConfigFile:     "",
		PopulationSize: 100,
		MaxGeneration:  1000,
		MutationRate:   0.1,
		CrossoverRate:  0.8,
		ElitismCount:   5,
		WorkerCount:    0, // Auto-detect
		Verbose:        false,
		ShowProgress:   true,
		SaveInterval:   50,
		// Convergence defaults
		ConvergenceStops:     0,        // Disabled by default
		ConvergenceTolerance: 0.000001, // Very small tolerance
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
	flag.IntVar(&config.PopulationSize, "population", config.PopulationSize, "Population size")
	flag.IntVar(&config.MaxGeneration, "generations", config.MaxGeneration, "Maximum generations")
	flag.Float64Var(&config.MutationRate, "mutation", config.MutationRate, "Mutation rate")
	flag.Float64Var(&config.CrossoverRate, "crossover", config.CrossoverRate, "Crossover rate")
	flag.IntVar(&config.ElitismCount, "elitism", config.ElitismCount, "Number of elite individuals")
	flag.IntVar(&config.WorkerCount, "workers", config.WorkerCount, "Number of parallel workers (0=auto)")
	flag.BoolVar(&config.Verbose, "verbose", config.Verbose, "Verbose output")
	flag.BoolVar(&config.ShowProgress, "progress", config.ShowProgress, "Show progress")
	flag.IntVar(&config.SaveInterval, "save-interval", config.SaveInterval, "Save best layout every N generations")
	flag.IntVar(&config.ConvergenceStops, "convergence-stops", config.ConvergenceStops, "Stop after N generations with same fitness (0=disabled, overrides max generations)")
	flag.Float64Var(&config.ConvergenceTolerance, "convergence-tolerance", config.ConvergenceTolerance, "Fitness difference tolerance for convergence detection")

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

	if config.MaxGeneration < 0 {
		return errors.New("max generations must be non-negative (0 = unlimited with convergence)")
	}

	if config.MaxGeneration == 0 && config.ConvergenceStops == 0 {
		return errors.New("either max generations or convergence stops must be set (not both zero)")
	}

	if config.ConvergenceStops < 0 {
		return errors.New("convergence stops must be non-negative")
	}

	if config.ConvergenceTolerance < 0 {
		return errors.New("convergence tolerance must be non-negative")
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

	// Always use full keyboard configuration
	parseConfig := parser.FullKeyboardConfig()

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

	// Set up fitness evaluation with full keyboard character set
	charset := genetic.FullKeyboardCharset()
	geometry := fitness.GetGeometryForCharset(charset)
	weights := fitness.DefaultWeights()
	fitnessEvaluator := fitness.NewFitnessEvaluator(geometry, weights)

	// Set up genetic algorithm with adaptive configuration
	var gaConfig genetic.Config

	// Use adaptive configuration if user hasn't specified custom genetic algorithm parameters
	// Allow adaptive config when only population, generation, or convergence parameters are customized
	isUsingDefaults := (appConfig.MutationRate == 0.1 && appConfig.CrossoverRate == 0.8 && appConfig.ElitismCount == 5)

	if isUsingDefaults {
		// User is using defaults for genetic parameters, apply adaptive configuration
		gaConfig = genetic.AdaptiveConfig(keyloggerData.TotalChars)

		// Override with user's specific settings if provided
		if appConfig.PopulationSize != 100 { // User specified custom population
			gaConfig.PopulationSize = appConfig.PopulationSize
		}
		if appConfig.ConvergenceStops > 0 {
			gaConfig.ConvergenceStops = appConfig.ConvergenceStops
		}
		if appConfig.ConvergenceTolerance != 0.000001 { // Non-default tolerance
			gaConfig.ConvergenceTolerance = appConfig.ConvergenceTolerance
		}
		if appConfig.MaxGeneration != 1000 { // User specified non-default generations
			gaConfig.MaxGenerations = appConfig.MaxGeneration
		}

		fmt.Printf("Using adaptive configuration for dataset size: %d characters\n", keyloggerData.TotalChars)
		if appConfig.PopulationSize != 100 {
			fmt.Printf("with custom population size: %d\n", appConfig.PopulationSize)
		}
		if appConfig.ConvergenceStops > 0 {
			fmt.Printf("with convergence stopping after %d stagnant generations\n", appConfig.ConvergenceStops)
		}
	} else {
		// User has customized parameters, respect their choices
		gaConfig = genetic.Config{
			PopulationSize:       appConfig.PopulationSize,
			MaxGenerations:       appConfig.MaxGeneration,
			MutationRate:         appConfig.MutationRate,
			CrossoverRate:        appConfig.CrossoverRate,
			ElitismCount:         appConfig.ElitismCount,
			TournamentSize:       3,
			ParallelWorkers:      appConfig.WorkerCount,
			ConvergenceStops:     appConfig.ConvergenceStops,
			ConvergenceTolerance: appConfig.ConvergenceTolerance,
		}

		fmt.Println("Using custom configuration")
	}

	ga := genetic.NewParallelGA(fitnessEvaluator, gaConfig, charset)

	fmt.Printf("Starting genetic algorithm with:\n")
	fmt.Printf("- Population size: %d\n", gaConfig.PopulationSize)
	if gaConfig.MaxGenerations > 0 {
		fmt.Printf("- Max generations: %d\n", gaConfig.MaxGenerations)
	} else {
		fmt.Printf("- Max generations: unlimited (convergence-based)\n")
	}
	if gaConfig.ConvergenceStops > 0 {
		fmt.Printf("- Convergence stops: %d (after %d generations with same fitness)\n", gaConfig.ConvergenceStops, gaConfig.ConvergenceStops)
		fmt.Printf("- Convergence tolerance: %.6f\n", gaConfig.ConvergenceTolerance)
	}
	fmt.Printf("- Mutation rate: %.2f\n", gaConfig.MutationRate)
	fmt.Printf("- Crossover rate: %.2f\n", gaConfig.CrossoverRate)
	fmt.Printf("- Elite count: %d\n", gaConfig.ElitismCount)
	fmt.Printf("- Parallel workers: %d\n", gaConfig.ParallelWorkers)
	fmt.Println()

	// Progress tracking
	startTime := time.Now()

	var lastSavedGeneration int

	// Fitness history for convergence chart
	var fitnessHistory []float64

	// Run genetic algorithm
	bestIndividual, err := ga.Run(ctx, keyloggerData, func(generation int, best genetic.Individual) {
		// Track fitness history for convergence chart
		fitnessHistory = append(fitnessHistory, best.Fitness)

		if appConfig.ShowProgress {
			elapsed := time.Since(startTime)
			if gaConfig.MaxGenerations > 0 {
				avgTime := elapsed / time.Duration(generation+1)
				remaining := avgTime * time.Duration(gaConfig.MaxGenerations-generation-1)
				fmt.Printf("Generation %4d: Best fitness = %.6f (ETA: %v)\n",
					generation, best.Fitness, remaining.Round(time.Second))
			} else {
				// Convergence mode - show more detail for debugging
				fitnessChange := 0.0
				if len(fitnessHistory) > 1 {
					fitnessChange = best.Fitness - fitnessHistory[len(fitnessHistory)-2]
				}
				fmt.Printf("Generation %4d: Best fitness = %.6f (Δ%.6f, elapsed: %v)\n",
					generation, best.Fitness, fitnessChange, elapsed.Round(time.Second))
			}
		}

		// Save intermediate results
		if appConfig.Verbose &&
			appConfig.SaveInterval > 0 &&
			generation-lastSavedGeneration >= appConfig.SaveInterval {
			saveLayout(best, fmt.Sprintf("%s.gen%d", appConfig.OutputFile, generation))
			lastSavedGeneration = generation
		}
	})
	if err != nil {
		return fmt.Errorf("genetic algorithm failed: %w", err)
	}

	// Save final result
	fmt.Printf("\nOptimization complete!\n")
	fmt.Printf("Best fitness: %.6f\n", bestIndividual.Fitness)
	fmt.Printf("Total time: %v\n", time.Since(startTime).Round(time.Second))

	// Display fitness convergence chart
	printFitnessConvergenceChart(fitnessHistory, gaConfig.ConvergenceStops > 0)

	// Create display handler
	kbDisplay := display.NewKeyboardDisplay()

	// Show enhanced summary first
	kbDisplay.PrintSummary(bestIndividual, keyloggerData, fitnessEvaluator)

	// Use layered layout display
	fmt.Printf("\n\033[1;34mOPTIMIZED KEYBOARD LAYOUT (ALL LAYERS):\033[0m\n")
	kbDisplay.PrintLayeredLayout(bestIndividual, keyloggerData, "qwerty")

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

	// Add optimized keyboard layers information using the improved display logic
	kbDisplay := display.NewKeyboardDisplay()
	optimizedLayout := kbDisplay.CreateOptimizedLayeredLayout(individual)

	optimizedLayers := make(map[string]map[string]any)
	optimizedLayers["base"] = make(map[string]any)
	optimizedLayers["shift"] = make(map[string]any)
	optimizedLayers["altgr"] = make(map[string]any)

	// Show all characters in the optimized layout
	optimizedChars := make([]string, len(individual.Layout))
	for i, char := range individual.Layout {
		optimizedChars[i] = string(char)
	}

	optimizedLayers["base"]["characters"] = optimizedChars
	optimizedLayers["base"]["layout_string"] = string(individual.Layout)

	// Use the properly constructed layered layout from display package
	for pos := 0; pos < len(individual.Layout) && pos < 70; pos++ {
		if layeredKey, exists := optimizedLayout.Keys[pos]; exists {
			// Base layer
			optimizedLayers["base"][fmt.Sprintf("pos_%d", pos)] = string(layeredKey.BaseChar)

			// Shift layer (only if character is assigned, i.e., not 0)
			if layeredKey.ShiftChar != 0 {
				optimizedLayers["shift"][fmt.Sprintf("pos_%d", pos)] = string(layeredKey.ShiftChar)
			}

			// AltGr layer (if exists)
			if layeredKey.AltGrChar != nil {
				optimizedLayers["altgr"][fmt.Sprintf("pos_%d", pos)] = string(*layeredKey.AltGrChar)
			}
		}
	}

	// Add character set information
	optimizedLayers["base"]["charset_name"] = individual.Charset.Name
	optimizedLayers["base"]["charset_size"] = individual.Charset.Size
	optimizedLayers["base"]["total_positions"] = len(individual.Layout)

	layout["optimized_keyboard_layers"] = optimizedLayers

	// Add metadata about layers
	layerMetadata := make(map[string]any)
	layerMetadata["layer_costs"] = map[string]float64{
		"base":  1.0,
		"shift": 1.5,
		"altgr": 2.0,
	}
	layerMetadata["description"] = "Keyboard layers show character access patterns. Base layer requires no modifiers, Shift layer requires Shift key, AltGr layer requires AltGr (Right Alt) key."

	layout["layer_metadata"] = layerMetadata

	// Marshal to JSON with HTML escaping disabled for better readability
	var buf bytes.Buffer

	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false) // Don't escape <, >, &, etc.

	err := encoder.Encode(layout)
	if err != nil {
		return fmt.Errorf("failed to marshal layout: %w", err)
	}

	// Write to file
	return os.WriteFile(filename, buf.Bytes(), 0o644)
}

// printFitnessConvergenceChart displays an ASCII chart showing fitness convergence over generations.
func printFitnessConvergenceChart(fitnessHistory []float64, isConvergenceMode bool) {
	if len(fitnessHistory) < 2 {
		return
	}

	fmt.Printf("\n╔═══════════════════════════════════════════════════════════════════╗\n")
	if isConvergenceMode {
		fmt.Printf("║                    FITNESS CONVERGENCE CHART                     ║\n")
	} else {
		fmt.Printf("║                      FITNESS EVOLUTION CHART                     ║\n")
	}
	fmt.Printf("╚═══════════════════════════════════════════════════════════════════╝\n\n")

	// Find min and max fitness for scaling
	minFitness := fitnessHistory[0]
	maxFitness := fitnessHistory[0]
	for _, fitness := range fitnessHistory {
		if fitness < minFitness {
			minFitness = fitness
		}
		if fitness > maxFitness {
			maxFitness = fitness
		}
	}

	// Chart dimensions
	const chartHeight = 20
	const chartWidth = 60

	// Calculate improvement percentage
	totalImprovement := 0.0
	if maxFitness > minFitness {
		totalImprovement = ((maxFitness - minFitness) / minFitness) * 100
	}

	// Scale fitness values to chart height
	scaleFitness := func(fitness float64) int {
		if maxFitness == minFitness {
			return chartHeight / 2
		}
		scaled := int((fitness - minFitness) / (maxFitness - minFitness) * float64(chartHeight-1))
		return chartHeight - 1 - scaled // Flip for display (higher values at top)
	}

	// Group generations for display (subsample if too many generations)
	step := 1
	if len(fitnessHistory) > chartWidth {
		step = len(fitnessHistory) / chartWidth
	}

	// Create the chart
	chart := make([][]rune, chartHeight)
	for i := range chart {
		chart[i] = make([]rune, chartWidth)
		for j := range chart[i] {
			chart[i][j] = ' '
		}
	}

	// Plot the fitness line
	for i := 0; i < chartWidth && i*step < len(fitnessHistory); i++ {
		fitnessIdx := i * step
		if fitnessIdx >= len(fitnessHistory) {
			fitnessIdx = len(fitnessHistory) - 1
		}

		row := scaleFitness(fitnessHistory[fitnessIdx])
		if row >= 0 && row < chartHeight {
			chart[row][i] = '█'

			// Add gradient effect with dots for smoother visualization
			if i > 0 && i-1 < chartWidth {
				prevIdx := (i - 1) * step
				if prevIdx < len(fitnessHistory) {
					prevRow := scaleFitness(fitnessHistory[prevIdx])
					// Fill between current and previous points
					startRow, endRow := prevRow, row
					if startRow > endRow {
						startRow, endRow = endRow, startRow
					}
					for r := startRow; r <= endRow; r++ {
						if r >= 0 && r < chartHeight && chart[r][i] == ' ' {
							chart[r][i] = '▓'
						}
					}
				}
			}
		}
	}

	// Display the chart with Y-axis labels
	fmt.Printf("Fitness\n")
	for i := 0; i < chartHeight; i++ {
		// Calculate the fitness value for this row
		rowFitness := minFitness + (maxFitness-minFitness)*float64(chartHeight-1-i)/float64(chartHeight-1)
		fmt.Printf("%7.4f │", rowFitness)

		for j := 0; j < chartWidth; j++ {
			fmt.Printf("%c", chart[i][j])
		}
		fmt.Printf("│\n")
	}

	// X-axis
	fmt.Printf("        └")
	for i := 0; i < chartWidth; i++ {
		fmt.Printf("─")
	}
	fmt.Printf("┘\n")

	// Generation markers
	fmt.Printf("         ")
	for i := 0; i < chartWidth; i += 10 {
		gen := i * step
		if gen < len(fitnessHistory) {
			genStr := fmt.Sprintf("%d", gen)
			for j, r := range genStr {
				if i+j < chartWidth {
					fmt.Printf("%c", r)
				}
			}
			// Fill remaining space to next marker
			for j := len(genStr); j < 10 && i+j < chartWidth; j++ {
				fmt.Printf(" ")
			}
		}
	}
	fmt.Printf("\n         Generation\n\n")

	// Summary statistics
	fmt.Printf("\033[1;36mCONVERGENCE SUMMARY:\033[0m\n")
	fmt.Printf("   * Total generations: %d\n", len(fitnessHistory))
	fmt.Printf("   * Starting fitness: %.6f\n", fitnessHistory[0])
	fmt.Printf("   * Final fitness: %.6f\n", fitnessHistory[len(fitnessHistory)-1])
	fmt.Printf("   * Best fitness: %.6f\n", maxFitness)
	fmt.Printf("   * Total improvement: %.2f%%\n", totalImprovement)

	// Find convergence point if in convergence mode
	if isConvergenceMode && len(fitnessHistory) > 10 {
		// Look for when fitness stopped improving significantly
		const convergenceWindow = 5
		convergenceGen := -1

		for i := convergenceWindow; i < len(fitnessHistory); i++ {
			isConverged := true
			for j := 1; j <= convergenceWindow; j++ {
				if abs(fitnessHistory[i]-fitnessHistory[i-j]) > 0.000001 {
					isConverged = false
					break
				}
			}
			if isConverged {
				convergenceGen = i
				break
			}
		}

		if convergenceGen > 0 {
			fmt.Printf("   * Converged at generation: %d\n", convergenceGen)
			fmt.Printf("   * Convergence fitness: %.6f\n", fitnessHistory[convergenceGen])
		}
	}

	fmt.Println()
}

// abs returns the absolute value of a float64.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
