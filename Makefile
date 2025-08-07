.PHONY: help build test clean run benchmark lint deps

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

example: build
	mkdir -p examples
	echo "the quick brown fox jumps over the lazy dog" > examples/sample.txt
	echo "hello world programming test keyboard layout" >> examples/sample.txt
	echo "genetic algorithms optimize keyboard layouts for typing efficiency" >> examples/sample.txt
	$(BUILD_DIR)/$(BINARY_NAME) -input examples/sample.txt -output examples/result.json -generations 10 -population 20

example-verbose: build
	mkdir -p examples
	echo "the quick brown fox jumps over the lazy dog" > examples/sample.txt
	echo "hello world programming test keyboard layout optimization" >> examples/sample.txt
	echo "genetic algorithms can create efficient typing patterns" >> examples/sample.txt
	echo "abcdefghijklmnopqrstuvwxyz" >> examples/sample.txt
	$(BUILD_DIR)/$(BINARY_NAME) -input examples/sample.txt -output examples/result.json -generations 15 -population 30 -verbose

example-hp: build
	$(BUILD_DIR)/$(BINARY_NAME) -input harrypotter.txt -output examples/result.json -verbose

# install: build
	# go install $(CMD_DIR)