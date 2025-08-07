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
	go test $(PKG_DIR)

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
	go run ./cmd/keyboardgen/ -input examples/sample.txt -output examples/result.json -generations 10 -population 20

example-verbose:
	mkdir -p examples
	echo "the quick brown fox jumps over the lazy dog" > examples/sample.txt
	echo "hello world programming test keyboard layout optimization" >> examples/sample.txt
	echo "genetic algorithms can create efficient typing patterns" >> examples/sample.txt
	echo "abcdefghijklmnopqrstuvwxyz" >> examples/sample.txt
	go run ./cmd/keyboardgen/ -input examples/sample.txt -output examples/result.json -generations 15 -population 30 -verbose

example-hp:
	go run ./cmd/keyboardgen/ -input harrypotter.txt -output examples/result.json -verbose

example-programming:
	mkdir -p examples
	find . -type f -name '*.go' -exec cat {} \; > examples/programming.txt
	find . -type f -name '*.md' -exec cat {} \; >> examples/programming.txt
	go run ./cmd/keyboardgen/ -input examples/programming.txt -output examples/programming_result.json -generations 20 -population 50

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
	go run ./cmd/keyboardgen/ -input examples/fullset.txt -output examples/fullset_result.json -generations 25 -population 60