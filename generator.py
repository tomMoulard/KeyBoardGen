#!/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on tue Jun 8 20:00:00 2017

@author: tm
This uses everithing he can to create a powerfull keyboard
"""

# imports
import usefullFunk
import keyBoard
import genetik
import argparse
import sys
from my_print import my_print

def init_kb(args):
    keyboard = keyBoard.KeyBoard("qwerty", seed=args.seed)
    keyboard.toGraph()
    if (args.show_dot):
        print(keyboard.graph.toDot())
        raise SystemExit(0)
    for x in range(len(keyboard.keys)):
        my_print(args, x, keyboard.keys[x])
    my_print(args, keyboard)
    keyboard.randomize()
    my_print(args, keyboard)
    return keyboard

def main(args):
    keyboard = init_kb(args)
    play(args.generation_number, args.population_size, args.klf)

def play(genNumber, popSize, fileName):
    pop = [keyBoard.KeyBoard("random", seed=42) for x in range(popSize)]
    genetik.sortPop(pop, 0, popSize - 1)
    for gen in range(genNumber):
        for kb in pop:
            kb.valuate()
        genetik.evolve(pop, fileName)
        print(pop[0], gen)# Print the best keyboard for each generation

def parse_args():
    parser = argparse.ArgumentParser(str(sys.argv[0]))
    parser.add_argument("klf",
                        metavar="KLF", type=argparse.FileType("r+"),
                        default="./keys.log",
                        help="keylogger file (KLF) to train the programm")
    parser.add_argument("--len", "-l",
                        metavar="nb", type=int, default=10,
                        help="represent the lenght of the learning")
    parser.add_argument("--show-dot",
                        help="Display the dot version of the graph",
                        action = "store_true")
    parser.add_argument("--generation-number", "--gn", "-g",
                        metavar="nb", type=int, default=10,
                        help="represent the number of generations to have")
    parser.add_argument("--population-size", "--ps", "-p",
                        metavar="nb", type=int, default=10,
                        help="represent the population size to evolve")
    parser.add_argument("--seed", "-s",
                        metavar="seed", type=int, default=None,
                        help="to set the seed")
    parser.add_argument("--verbose", "-v", action="count", default=0,
                        help="verbose level(default=0, max=100)")
    args = parser.parse_args()
    my_print(args, args.klf.readlines())
    my_print(args, "len: {}, popsize:{}, seed: {}, verbose: {}"
            .format(args.len, args.population_size, args.seed, args.verbose))
    return args

if __name__ == "__main__":
    main(parse_args())
