FILE=test.klf
PICT=tmp.png

all:
	./generator.py $(FILE)

png:
	./generator.py --show-dot $(FILE) | tail -n+3 | dot -Tpng -o $(PICT)
