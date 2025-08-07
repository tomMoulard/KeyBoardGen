# KeyBoardGen

A genetic algorithm-based keyboard layout optimizer that analyzes your typing patterns to generate optimal keyboard layouts for maximum typing efficiency.

## Features

- **Multiple Character Sets**: Support for alphabet-only, alphanumeric, programming symbols, and full keyboard layouts
- **Genetic Algorithm Optimization**: Uses tournament selection, order crossover, and swap mutation
- **Adaptive Configuration**: Automatically adjusts parameters based on dataset size
- **Parallel Processing**: Multi-threaded evaluation for improved performance
- **Progress Tracking**: Real-time fitness monitoring and periodic saves
- **Comprehensive Testing**: Full regression test suite with null character bug protection

## Installation

```bash
git clone https://github.com/tommoulard/KeyBoardGen.git
cd KeyBoardGen
go build ./cmd/keyboardgen
```

## Usage

### Basic Usage

```bash
./keyboardgen -input keylog.txt -output optimized_layout.json
```

### Advanced Usage

```bash
./keyboardgen \
  -input keylog.txt \
  -output layout.json \
  -charset programming \
  -generations 500 \
  -population 200 \
  -mutation 0.15 \
  -crossover 0.9 \
  -verbose
```

### Input Format

The input file should contain your typing data as plain text:

```
the quick brown fox jumps over the lazy dog
hello world programming test keyboard layout
function calculateScore(x, y) {
    return Math.sqrt(x * x + y * y);
}
```

For programming layouts, include code samples:

```javascript
function calculateScore(x, y) {
    if (x > 0 && y > 0) {
        return Math.sqrt(x * x + y * y);
    }
    return 0.0;
}

const config = {
    "name": "test",
    "values": [1, 2, 3],
    "active": true
};
```

### Character Sets

- **`alphabet`**: 26 lowercase letters (a-z)
- **`alphanumeric`**: Letters + numbers (a-z, 0-9)
- **`programming`**: Letters + numbers + common symbols ({}[]();,.)
- **`full`**: Complete keyboard including special characters ($*+%/=)

### Configuration Options

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-input` | string | required | Input keylogger file |
| `-output` | string | `best_layout.json` | Output file for optimized layout |
| `-charset` | string | `alphabet` | Character set to optimize |
| `-generations` | int | 1000 | Maximum number of generations |
| `-population` | int | 100 | Population size |
| `-mutation` | float | 0.1 | Mutation rate (0.0-1.0) |
| `-crossover` | float | 0.8 | Crossover rate (0.0-1.0) |
| `-elitism` | int | 5 | Number of elite individuals preserved |
| `-workers` | int | auto | Parallel worker threads |
| `-verbose` | bool | false | Enable verbose output |
| `-progress` | bool | true | Show progress updates |
| `-save-interval` | int | 50 | Save interval (generations) |

### Configuration File

Use a JSON configuration file for complex setups:

```json
{
  "input_file": "keylog.txt",
  "output_file": "layout.json",
  "character_set": "programming",
  "population_size": 200,
  "max_generation": 500,
  "mutation_rate": 0.15,
  "crossover_rate": 0.9,
  "elitism_count": 10,
  "worker_count": 8,
  "verbose": true,
  "save_interval": 25
}
```

Run with: `./keyboardgen -config config.json`

### Output Format

The optimizer generates a JSON file with the optimal layout:

```json
{
  "age": 450,
  "fitness": 0.8234,
  "layout": "qwertyuiopasdfghjklzxcvbnm",
  "positions": {
    "a": 10,
    "b": 23,
    "c": 20,
    ...
  },
  "timestamp": "2025-08-07T02:11:58+02:00"
}
```

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
make lint
```

## Algorithm Details

### Fitness Evaluation

The fitness function considers:
- **Character frequency**: More frequent characters get better positions
- **Bigram frequency**: Common letter pairs minimize finger travel
- **Hand alternation**: Encourages alternating between hands
- **Finger load balancing**: Distributes work across all fingers
- **Row preferences**: Prioritizes home row placement

### Genetic Operations

- **Selection**: Tournament selection with configurable size
- **Crossover**: Order crossover (OX) preserves permutation validity
- **Mutation**: Swap mutation maintains character uniqueness
- **Elitism**: Best individuals preserved across generations

### Adaptive Configuration

The system automatically adjusts parameters based on dataset size:
- Small datasets (< 10K chars): Conservative settings
- Medium datasets (10K-100K chars): Balanced exploration
- Large datasets (> 100K chars): High diversity settings

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/new-feature`)
3. Make changes and add tests
4. Run `make test` and `make lint`
5. Commit your changes (`git commit -am 'Add new feature'`)
6. Push to the branch (`git push origin feature/new-feature`)
7. Create a Pull Request

## License

This project is available under the MIT License.