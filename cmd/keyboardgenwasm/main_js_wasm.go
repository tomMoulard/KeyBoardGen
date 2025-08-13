package main

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"

	"github.com/tommoulard/keyboardgen/internal/runner"
	"github.com/tommoulard/keyboardgen/pkg/config"
	"github.com/tommoulard/keyboardgen/pkg/genetic"
)

// Result holds the optimization result for WASM.
type Result struct {
	Success         bool           `json:"success"`
	Error           string         `json:"error,omitempty"`
	BestFitness     float64        `json:"best_fitness"`
	Generation      int            `json:"generation"`
	Layout          string         `json:"layout"`
	Positions       map[string]int `json:"positions"`
	OptimizedLayers map[string]any `json:"optimized_layers"`
	FitnessHistory  []float64      `json:"fitness_history"`
	Statistics      map[string]any `json:"statistics"`
	TotalTime       string         `json:"total_time"`
}

// Progress holds progress information for WASM.
type Progress struct {
	Generation    int     `json:"generation"`
	BestFitness   float64 `json:"best_fitness"`
	FitnessChange float64 `json:"fitness_change"`
	ElapsedTime   string  `json:"elapsed_time"`
	ETA           string  `json:"eta,omitempty"`
}

var (
	progressCallback js.Value
	cancelCtx        context.CancelFunc
)

func main() {
	c := make(chan struct{}, 0)

	// Register WASM functions
	js.Global().Set("optimizeKeyboard", js.FuncOf(optimizeKeyboard))
	js.Global().Set("stopOptimization", js.FuncOf(stopOptimization))
	js.Global().Set("getDefaultConfig", js.FuncOf(getDefaultConfig))
	js.Global().Set("validateConfig", js.FuncOf(validateConfig))
	js.Global().Set("getParameterInfo", js.FuncOf(getParameterInfo))

	fmt.Println("KeyboardGen WASM module loaded")

	<-c
}

// getDefaultConfig returns the default configuration as JSON.
func getDefaultConfig(this js.Value, args []js.Value) any {
	cfg := config.DefaultForWASM()
	jsonBytes, _ := json.Marshal(cfg)
	return string(jsonBytes)
}

// getParameterInfo returns information about all configuration parameters.
func getParameterInfo(this js.Value, args []js.Value) any {
	paramInfo := config.GetParameterInfo()
	jsonBytes, _ := json.Marshal(paramInfo)
	return string(jsonBytes)
}

// validateConfig validates the provided configuration.
func validateConfig(this js.Value, args []js.Value) any {
	if len(args) != 1 {
		return map[string]any{
			"valid": false,
			"error": "Invalid number of arguments",
		}
	}

	configJSON := args[0].String()
	var cfg config.Config
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return map[string]any{
			"valid": false,
			"error": "Invalid JSON configuration: " + err.Error(),
		}
	}

	if err := cfg.ValidateForWASM(); err != nil {
		return map[string]any{
			"valid": false,
			"error": err.Error(),
		}
	}

	return map[string]any{
		"valid": true,
	}
}

// optimizeKeyboard runs the keyboard optimization algorithm.
func optimizeKeyboard(this js.Value, args []js.Value) any {
	if len(args) < 1 || len(args) > 2 {
		return createErrorResult("Invalid number of arguments")
	}

	configJSON := args[0].String()
	if len(args) > 1 {
		progressCallback = args[1]
	}

	var cfg config.Config
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return createErrorResult("Invalid JSON configuration: " + err.Error())
	}

	if err := cfg.ValidateForWASM(); err != nil {
		return createErrorResult("Configuration error: " + err.Error())
	}

	// Create context with cancellation
	ctx := context.Background()
	ctx, cancelCtx = context.WithCancel(ctx)

	go func() {
		result := runOptimization(ctx, cfg)
		resultJSON, _ := json.Marshal(result)

		// Return result via callback or global variable
		if progressCallback.IsUndefined() {
			js.Global().Set("keyboardGenResult", string(resultJSON))
		} else {
			progressCallback.Invoke("result", string(resultJSON))
		}
	}()

	return "Optimization started"
}

// stopOptimization cancels the running optimization.
func stopOptimization(this js.Value, args []js.Value) any {
	if cancelCtx != nil {
		cancelCtx()
		return "Optimization stopped"
	}
	return "No optimization running"
}

// createErrorResult creates an error result.
func createErrorResult(errorMsg string) Result {
	return Result{
		Success: false,
		Error:   errorMsg,
	}
}

// runOptimization runs the genetic algorithm optimization.
func runOptimization(ctx context.Context, cfg config.Config) Result {
	startTime := time.Now()

	// Create runner
	r, err := runner.New(cfg)
	if err != nil {
		return createErrorResult("Failed to create runner: " + err.Error())
	}

	var fitnessHistory []float64

	// Create progress callback for runner
	progressCB := func(generation int, best genetic.Individual) {
		fitnessHistory = append(fitnessHistory, best.Fitness)

		if cfg.ShowProgress {
			elapsed := time.Since(startTime)

			progress := Progress{
				Generation:  generation,
				BestFitness: best.Fitness,
				ElapsedTime: elapsed.Round(time.Second).String(),
			}

			if len(fitnessHistory) > 1 {
				progress.FitnessChange = best.Fitness - fitnessHistory[len(fitnessHistory)-2]
			}

			if cfg.MaxGeneration > 0 {
				avgTime := elapsed / time.Duration(generation+1)
				remaining := avgTime * time.Duration(cfg.MaxGeneration-generation-1)
				progress.ETA = remaining.Round(time.Second).String()
			}

			// Send progress update
			if !progressCallback.IsUndefined() {
				progressJSON, _ := json.Marshal(progress)
				progressCallback.Invoke("progress", string(progressJSON))
			}
		}
	}

	// Run optimization
	bestIndividual, history, err := r.RunFromText(ctx, cfg.InputText, progressCB)
	if err != nil {
		if ctx.Err() == context.Canceled {
			return createErrorResult("Optimization was canceled")
		}
		return createErrorResult("Optimization failed: " + err.Error())
	}

	// Use the history from runner if available, otherwise use our local copy
	if len(history) > 0 {
		fitnessHistory = history
	}

	// Create result
	result := Result{
		Success:        true,
		BestFitness:    bestIndividual.Fitness,
		Layout:         string(bestIndividual.Layout),
		FitnessHistory: fitnessHistory,
		TotalTime:      time.Since(startTime).Round(time.Second).String(),
	}

	// Create position mapping
	positions := make(map[string]int)
	for i, char := range bestIndividual.Layout {
		positions[string(char)] = i
	}
	result.Positions = positions

	// Create statistics
	statistics := make(map[string]any)
	statistics["charset_name"] = bestIndividual.Charset.Name
	statistics["charset_size"] = bestIndividual.Charset.Size
	statistics["generation_count"] = len(fitnessHistory)
	statistics["total_positions"] = len(bestIndividual.Layout)

	if len(fitnessHistory) > 0 {
		statistics["initial_fitness"] = fitnessHistory[0]
		statistics["final_fitness"] = fitnessHistory[len(fitnessHistory)-1]

		if fitnessHistory[0] > 0 {
			improvement := ((result.BestFitness - fitnessHistory[0]) / fitnessHistory[0]) * 100
			statistics["improvement_percentage"] = improvement
		}
	}

	result.Statistics = statistics

	return result
}