package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// Config holds application configuration.
type Config struct {
	InputFile      string  `json:"input_file"`
	InputText      string  `json:"input_text"`
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

// Default returns default application configuration.
func Default() Config {
	return Config{
		InputFile:      "",
		InputText:      "",
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

// DefaultForWASM returns default configuration optimized for WASM usage.
func DefaultForWASM() Config {
	config := Default()
	config.WorkerCount = 1 // Single-threaded for WASM
	config.InputFile = ""  // Not used in WASM

	return config
}

// LoadFromFile loads configuration from a JSON file.
func LoadFromFile(filename string) (Config, error) {
	config := Default()

	data, err := os.ReadFile(filename)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// LoadFromJSON loads configuration from a JSON string.
func LoadFromJSON(jsonStr string) (Config, error) {
	config := Default()

	err := json.Unmarshal([]byte(jsonStr), &config)
	if err != nil {
		return config, fmt.Errorf("failed to parse JSON config: %w", err)
	}

	return config, nil
}

// SaveToFile saves configuration to a JSON file.
func (c Config) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(filename, data, 0o644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ToJSON returns the configuration as a JSON string.
func (c Config) ToJSON() (string, error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	return string(data), nil
}

// Validate validates the configuration for CLI usage.
func (c Config) Validate() error {
	if c.InputFile == "" && c.InputText == "" {
		return errors.New("either input file or input text is required")
	}

	if c.InputFile != "" {
		if _, err := os.Stat(c.InputFile); os.IsNotExist(err) {
			return fmt.Errorf("input file does not exist: %s", c.InputFile)
		}
	}

	return c.validateCommon()
}

// ValidateForWASM validates the configuration for WASM usage.
func (c Config) ValidateForWASM() error {
	if c.InputText == "" {
		return errors.New("input text is required for WASM")
	}

	return c.validateCommon()
}

// validateCommon performs common validation shared between CLI and WASM.
func (c Config) validateCommon() error {
	if c.PopulationSize < 10 {
		return errors.New("population size must be at least 10")
	}

	if c.MaxGeneration < 0 {
		return errors.New("max generations must be non-negative (0 = unlimited with convergence)")
	}

	if c.MaxGeneration == 0 && c.ConvergenceStops == 0 {
		return errors.New("either max generations or convergence stops must be set (not both zero)")
	}

	if c.ConvergenceStops < 0 {
		return errors.New("convergence stops must be non-negative")
	}

	if c.ConvergenceTolerance < 0 {
		return errors.New("convergence tolerance must be non-negative")
	}

	if c.MutationRate < 0 || c.MutationRate > 1 {
		return errors.New("mutation rate must be between 0 and 1")
	}

	if c.CrossoverRate < 0 || c.CrossoverRate > 1 {
		return errors.New("crossover rate must be between 0 and 1")
	}

	if c.ElitismCount < 0 || c.ElitismCount >= c.PopulationSize {
		return errors.New("elitism count must be between 0 and population size")
	}

	if c.WorkerCount < 0 {
		return errors.New("worker count must be non-negative (0 = auto-detect)")
	}

	if c.SaveInterval < 0 {
		return errors.New("save interval must be non-negative")
	}

	return nil
}

// IsUsingDefaultGeneticParams returns true if the config uses default genetic algorithm parameters.
func (c Config) IsUsingDefaultGeneticParams() bool {
	return c.MutationRate == 0.1 && c.CrossoverRate == 0.8 && c.ElitismCount == 5
}

// GetParameterInfo returns information about all configuration parameters.
func GetParameterInfo() []ParameterInfo {
	return []ParameterInfo{
		{
			Name:        "input_file",
			Type:        "string",
			Description: "Path to keylogger input file (CLI only)",
			Default:     "",
			Required:    false,
		},
		{
			Name:        "input_text",
			Type:        "string",
			Description: "Input text for analysis (WASM only)",
			Default:     "",
			Required:    false,
		},
		{
			Name:        "output_file",
			Type:        "string",
			Description: "Output file for best layout (CLI only)",
			Default:     "best_layout.json",
			Required:    false,
		},
		{
			Name:        "population_size",
			Type:        "integer",
			Description: "Size of the genetic algorithm population",
			Default:     100,
			Required:    false,
			Min:         10,
		},
		{
			Name:        "max_generation",
			Type:        "integer",
			Description: "Maximum number of generations (0 = unlimited with convergence)",
			Default:     1000,
			Required:    false,
			Min:         0,
		},
		{
			Name:        "mutation_rate",
			Type:        "float",
			Description: "Probability of mutation per individual",
			Default:     0.1,
			Required:    false,
			Min:         0.0,
			Max:         1.0,
		},
		{
			Name:        "crossover_rate",
			Type:        "float",
			Description: "Probability of crossover between individuals",
			Default:     0.8,
			Required:    false,
			Min:         0.0,
			Max:         1.0,
		},
		{
			Name:        "elitism_count",
			Type:        "integer",
			Description: "Number of best individuals to preserve each generation",
			Default:     5,
			Required:    false,
			Min:         0,
		},
		{
			Name:        "worker_count",
			Type:        "integer",
			Description: "Number of parallel workers (0 = auto-detect, 1 for WASM)",
			Default:     0,
			Required:    false,
			Min:         0,
		},
		{
			Name:        "verbose",
			Type:        "boolean",
			Description: "Enable verbose output and detailed logging",
			Default:     false,
			Required:    false,
		},
		{
			Name:        "show_progress",
			Type:        "boolean",
			Description: "Show optimization progress updates",
			Default:     true,
			Required:    false,
		},
		{
			Name:        "save_interval",
			Type:        "integer",
			Description: "Save intermediate results every N generations",
			Default:     50,
			Required:    false,
			Min:         0,
		},
		{
			Name:        "convergence_stops",
			Type:        "integer",
			Description: "Stop after N generations with same fitness (0 = disabled)",
			Default:     0,
			Required:    false,
			Min:         0,
		},
		{
			Name:        "convergence_tolerance",
			Type:        "float",
			Description: "Fitness difference tolerance for convergence detection",
			Default:     0.000001,
			Required:    false,
			Min:         0.0,
		},
	}
}

// ParameterInfo describes a configuration parameter.
type ParameterInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Default     any    `json:"default"`
	Required    bool   `json:"required"`
	Min         any    `json:"min,omitempty"`
	Max         any    `json:"max,omitempty"`
}
