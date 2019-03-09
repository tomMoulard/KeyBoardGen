FILE=test.klf
PICT=tmp.png
BIN=./generator.py

all:
	$(BIN) $(FILE) $(ARGS)

help:
	$(BIN) --help

png:
	$(BIN) --show-dot $(FILE) | tail -n+3 | dot -Tpng -o $(PICT)
