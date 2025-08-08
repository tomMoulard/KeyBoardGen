.PHONY: help build test clean run benchmark lint deps example example-verbose example-hp example-programming example-full-charset

BINARY_NAME := keyboardgen
BUILD_DIR := ./build
CMD_DIR := ./cmd/keyboardgen
PKG_DIR := ./...

.DEFAULT_GOAL := all

all: deps lint test build

build: deps
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

test:
	go test -race -cover $(PKG_DIR)

test-coverage:
	go test -v -coverprofile=coverage.out $(PKG_DIR)
	go tool cover -html=coverage.out -o coverage.html

benchmark:
	go test -bench=. -benchmem $(PKG_DIR)

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.3.1 run

deps:
	go mod tidy

clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	rm -rf examples

example:
	mkdir -p examples
	echo "the quick brown fox jumps over the lazy dog" > examples/sample.txt
	echo "hello world programming test keyboard layout" >> examples/sample.txt
	echo "genetic algorithms optimize keyboard layouts for typing efficiency" >> examples/sample.txt
	go run ./cmd/keyboardgen/ -input examples/sample.txt -output examples/result.json -generations 30 -population 80 -mutation 0.15 -elitism 2

example-verbose:
	mkdir -p examples
	echo "the quick brown fox jumps over the lazy dog" > examples/sample.txt
	echo "hello world programming test keyboard layout optimization" >> examples/sample.txt
	echo "genetic algorithms can create efficient typing patterns" >> examples/sample.txt
	echo "abcdefghijklmnopqrstuvwxyz" >> examples/sample.txt
	go run ./cmd/keyboardgen/ -input examples/sample.txt -output examples/result.json -generations 50 -population 100 -mutation 0.15 -elitism 2 -verbose

example-hp:
	go run ./cmd/keyboardgen/ -input harrypotter.txt -output examples/result.json -verbose

example-programming:
	mkdir -p examples
	find . -type f -name '*.go' -exec cat {} \; > examples/programming.txt
	find . -type f -name '*.md' -exec cat {} \; >> examples/programming.txt
	go run ./cmd/keyboardgen/ -input examples/programming.txt -output examples/programming_result.json -generations 100 -population 200 -mutation 0.2 -elitism 3

example-diverse:
	mkdir -p examples
	echo "the quick brown fox jumps over the lazy dog" > examples/diverse_test.txt
	echo "hello world programming test keyboard layout optimization" >> examples/diverse_test.txt
	echo "genetic algorithms can create efficient typing patterns" >> examples/diverse_test.txt
	echo "abcdefghijklmnopqrstuvwxyz" >> examples/diverse_test.txt
	go run ./cmd/keyboardgen/ -input examples/diverse_test.txt -output examples/diverse_result.json -generations 30 -population 60 -mutation 0.25 -elitism 2 -diverse-init=true -verbose

example-comparison:
	mkdir -p examples
	echo "testing keyboard layout optimization with different initialization strategies for genetic algorithms" > examples/comparison.txt
	echo "the quick brown fox jumps over the lazy dog" >> examples/comparison.txt
	echo "hello world programming genetic algorithm keyboard layout optimization testing" >> examples/comparison.txt
	echo "Random initialization test:" >> examples/comparison_random.txt
	go run ./cmd/keyboardgen/ -input examples/comparison.txt -output examples/random_init.json -generations 20 -population 30 -mutation 0.3 -elitism 1 -diverse-init=false
	echo "Diverse initialization test:" >> examples/comparison_diverse.txt  
	go run ./cmd/keyboardgen/ -input examples/comparison.txt -output examples/diverse_init.json -generations 20 -population 30 -mutation 0.3 -elitism 1 -diverse-init=true

example-full-charset:
	mkdir -p examples
	echo "Programming with full character set support:" > examples/fullset.txt
	echo "const obj = { key: 'value', count: 42, active: true };" >> examples/fullset.txt
	echo "function compute() { return (x + y) * z / 2.5; }" >> examples/fullset.txt
	echo "if (condition && flag || !disabled) { process(); }" >> examples/fullset.txt
	echo "array[index] = \`template \$${variable} string\`;" >> examples/fullset.txt
	echo "regex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$$/" >> examples/fullset.txt
	echo "Special chars: !@#$$%^&*()_+-=[]{}|;':\",./<>?" >> examples/fullset.txt
	echo "Numbers: 0123456789" >> examples/fullset.txt
	echo "Mixed: Hello, World! Testing 123... Done." >> examples/fullset.txt
	go run ./cmd/keyboardgen/ -input examples/fullset.txt -output examples/fullset_result.json -generations 75 -population 150 -mutation 0.18 -elitism 2