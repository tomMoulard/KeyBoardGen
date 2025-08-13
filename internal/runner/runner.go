package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tommoulard/keyboardgen/pkg/config"
	"github.com/tommoulard/keyboardgen/pkg/display"
	"github.com/tommoulard/keyboardgen/pkg/fitness"
	"github.com/tommoulard/keyboardgen/pkg/genetic"
	"github.com/tommoulard/keyboardgen/pkg/parser"
)

// ProgressCallback is called during optimization to report progress.
type ProgressCallback func(generation int, best genetic.Individual)

// Runner handles the execution of the genetic algorithm optimization.
type Runner struct {
	config     config.Config
	evaluator  *fitness.FitnessEvaluator
	ga         *genetic.ParallelGA
	keylogData *parser.KeyloggerData
	kbDisplay  *display.KeyboardDisplay
}

// New creates a new Runner with the given configuration.
func New(cfg config.Config) (*Runner, error) {
	return &Runner{
		config:    cfg,
		kbDisplay: display.NewKeyboardDisplay(),
	}, nil
}

// RunFromFile runs the optimization using input from a file.
func (r *Runner) RunFromFile(ctx context.Context, progressCallback ProgressCallback) (*genetic.Individual, []float64, error) {
	if r.config.InputFile == "" {
		return nil, nil, errors.New("input file is required")
	}

	file, err := os.Open(r.config.InputFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open input file: %w", err)
	}
	defer file.Close()

	return r.runFromReader(ctx, file, r.config.InputFile, progressCallback)
}

// RunFromText runs the optimization using input text.
func (r *Runner) RunFromText(ctx context.Context, inputText string, progressCallback ProgressCallback) (*genetic.Individual, []float64, error) {
	if inputText == "" {
		return nil, nil, errors.New("input text is required")
	}

	reader := strings.NewReader(inputText)

	return r.runFromReader(ctx, reader, "text", progressCallback)
}

// runFromReader performs the optimization using the provided reader.
func (r *Runner) runFromReader(ctx context.Context, reader io.Reader, sourceName string, progressCallback ProgressCallback) (*genetic.Individual, []float64, error) {
	// Parse keylogger data
	if r.config.Verbose {
		fmt.Printf("Parsing keylogger data from %s...\n", sourceName)
	}

	klparser := parser.NewKeyloggerParser()
	parseConfig := parser.FullKeyboardConfig()

	// Detect format based on file extension or content
	if strings.HasSuffix(strings.ToLower(sourceName), ".json") {
		parseConfig.Format = parser.JSONFormat
	} else if strings.Contains(strings.ToLower(filepath.Base(sourceName)), "vim") {
		parseConfig.Format = parser.VimCommandFormat
	}

	keyloggerData, err := klparser.Parse(reader, parseConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse keylogger data: %w", err)
	}

	if err := keyloggerData.Validate(); err != nil {
		return nil, nil, fmt.Errorf("insufficient keylogger data: %w", err)
	}

	r.keylogData = keyloggerData

	if r.config.Verbose {
		fmt.Printf("Parsed %d characters, %d unique bigrams\n",
			keyloggerData.TotalChars, len(keyloggerData.BigramFreq))
		r.printFrequencyInfo()
	}

	// Set up fitness evaluation
	charset := genetic.FullKeyboardCharset()
	geometry := fitness.GetGeometryForCharset(charset)
	weights := fitness.DefaultWeights()
	r.evaluator = fitness.NewFitnessEvaluator(geometry, weights)

	// Set up genetic algorithm
	gaConfig := r.createGAConfig(keyloggerData.TotalChars)
	r.ga = genetic.NewParallelGA(r.evaluator, gaConfig, charset)

	if r.config.Verbose {
		r.printGAConfig(gaConfig)
	}

	// Run genetic algorithm
	startTime := time.Now()

	var (
		fitnessHistory      []float64
		lastSavedGeneration int
	)

	bestIndividual, err := r.ga.Run(ctx, keyloggerData, func(generation int, best genetic.Individual) {
		fitnessHistory = append(fitnessHistory, best.Fitness)

		if r.config.ShowProgress {
			r.printProgress(generation, best, fitnessHistory, startTime, gaConfig)
		}

		// Save intermediate results
		if r.config.Verbose &&
			r.config.SaveInterval > 0 &&
			generation-lastSavedGeneration >= r.config.SaveInterval &&
			r.config.OutputFile != "" {
			r.saveLayout(best, fmt.Sprintf("%s.gen%d", r.config.OutputFile, generation))
			lastSavedGeneration = generation
		}

		// Call external progress callback
		if progressCallback != nil {
			progressCallback(generation, best)
		}
	})
	if err != nil {
		return nil, nil, fmt.Errorf("genetic algorithm failed: %w", err)
	}

	return &bestIndividual, fitnessHistory, nil
}

// PrintResults prints the optimization results.
func (r *Runner) PrintResults(bestIndividual *genetic.Individual, fitnessHistory []float64, totalTime time.Duration) {
	if r.keylogData == nil || r.evaluator == nil {
		return
	}

	fmt.Printf("\nOptimization complete!\n")
	fmt.Printf("Best fitness: %.6f\n", bestIndividual.Fitness)
	fmt.Printf("Total time: %v\n", totalTime.Round(time.Second))

	// Display fitness convergence chart
	r.printFitnessConvergenceChart(fitnessHistory)

	// Show enhanced summary first
	r.kbDisplay.PrintSummary(*bestIndividual, r.keylogData, r.evaluator)

	// Use layered layout display
	fmt.Printf("\n\033[1;34mOPTIMIZED KEYBOARD LAYOUT (ALL LAYERS):\033[0m\n")
	r.kbDisplay.PrintLayeredLayout(*bestIndividual, r.keylogData, "qwerty")

	// Print comprehensive statistics
	r.kbDisplay.PrintStatistics(*bestIndividual, r.keylogData)

	// Always show comparison with QWERTY
	r.kbDisplay.PrintComparisonWithEvaluator(*bestIndividual, r.keylogData, r.evaluator)

	// Show heatmap if verbose
	if r.config.Verbose {
		r.kbDisplay.PrintHeatmap(*bestIndividual, r.keylogData)
	}
}

// SaveLayout saves the layout to a JSON file.
func (r *Runner) SaveLayout(individual *genetic.Individual, filename string) error {
	if individual == nil {
		return errors.New("individual is nil")
	}

	return r.saveLayout(*individual, filename)
}

// createGAConfig creates genetic algorithm configuration.
func (r *Runner) createGAConfig(totalChars int) genetic.Config {
	var gaConfig genetic.Config

	if r.config.IsUsingDefaultGeneticParams() {
		gaConfig = genetic.AdaptiveConfig(totalChars)

		// Override with user's specific settings if provided
		if r.config.PopulationSize != 100 {
			gaConfig.PopulationSize = r.config.PopulationSize
		}

		if r.config.ConvergenceStops > 0 {
			gaConfig.ConvergenceStops = r.config.ConvergenceStops
		}

		if r.config.ConvergenceTolerance != 0.000001 {
			gaConfig.ConvergenceTolerance = r.config.ConvergenceTolerance
		}

		if r.config.MaxGeneration != 1000 {
			gaConfig.MaxGenerations = r.config.MaxGeneration
		}

		if r.config.Verbose {
			fmt.Printf("Using adaptive configuration for dataset size: %d characters\n", totalChars)
		}
	} else {
		gaConfig = genetic.Config{
			PopulationSize:       r.config.PopulationSize,
			MaxGenerations:       r.config.MaxGeneration,
			MutationRate:         r.config.MutationRate,
			CrossoverRate:        r.config.CrossoverRate,
			ElitismCount:         r.config.ElitismCount,
			TournamentSize:       3,
			ParallelWorkers:      r.config.WorkerCount,
			ConvergenceStops:     r.config.ConvergenceStops,
			ConvergenceTolerance: r.config.ConvergenceTolerance,
		}

		if r.config.Verbose {
			fmt.Println("Using custom configuration")
		}
	}

	return gaConfig
}

// printFrequencyInfo prints character and bigram frequency information.
func (r *Runner) printFrequencyInfo() {
	fmt.Println("\nMost frequent characters:")

	for i, cf := range r.keylogData.GetMostFrequentChars(10) {
		fmt.Printf("%2d. %c: %d (%.2f%%)\n", i+1, cf.Char, cf.Freq,
			float64(cf.Freq)*100/float64(r.keylogData.TotalChars))
	}

	fmt.Println("\nMost frequent bigrams:")

	for i, bf := range r.keylogData.GetMostFrequentBigrams(10) {
		fmt.Printf("%2d. %s: %d\n", i+1, bf.Bigram, bf.Freq)
	}

	fmt.Println()
}

// printGAConfig prints genetic algorithm configuration.
func (r *Runner) printGAConfig(gaConfig genetic.Config) {
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
}

// printProgress prints optimization progress.
func (r *Runner) printProgress(generation int, best genetic.Individual, fitnessHistory []float64, startTime time.Time, gaConfig genetic.Config) {
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

// saveLayout saves the layout to a JSON file.
func (r *Runner) saveLayout(individual genetic.Individual, filename string) error {
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

	// Add optimized keyboard layers information
	optimizedLayout := r.kbDisplay.CreateOptimizedLayeredLayout(individual)

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
			optimizedLayers["base"][fmt.Sprintf("pos_%d", pos)] = string(layeredKey.BaseChar)

			if layeredKey.ShiftChar != 0 {
				optimizedLayers["shift"][fmt.Sprintf("pos_%d", pos)] = string(layeredKey.ShiftChar)
			}

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
	encoder.SetEscapeHTML(false)

	err := encoder.Encode(layout)
	if err != nil {
		return fmt.Errorf("failed to marshal layout: %w", err)
	}

	// Write to file
	return os.WriteFile(filename, buf.Bytes(), 0o644)
}

// printFitnessConvergenceChart displays an ASCII chart showing fitness convergence over generations.
func (r *Runner) printFitnessConvergenceChart(fitnessHistory []float64) {
	if len(fitnessHistory) < 2 {
		return
	}

	isConvergenceMode := r.config.ConvergenceStops > 0

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
	const (
		chartHeight = 20
		chartWidth  = 60
	)

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

	for i := range chartHeight {
		// Calculate the fitness value for this row
		rowFitness := minFitness + (maxFitness-minFitness)*float64(chartHeight-1-i)/float64(chartHeight-1)
		fmt.Printf("%7.4f │", rowFitness)

		for j := range chartWidth {
			fmt.Printf("%c", chart[i][j])
		}

		fmt.Printf("│\n")
	}

	// X-axis
	fmt.Printf("        └")

	for range chartWidth {
		fmt.Printf("─")
	}

	fmt.Printf("┘\n")

	// Generation markers
	fmt.Printf("         ")

	for i := 0; i < chartWidth; i += 10 {
		gen := i * step
		if gen < len(fitnessHistory) {
			genStr := strconv.Itoa(gen)
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
