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

def main(args):
    keyboard = keyBoard.KeyBoard("qwerty", seed=args.seed)
    keyboard.toGraph()
    print(keyboard.graph.toDot())
    # print(keyboard.getAllLetters())
    for x in range(len(keyboard.keys)):
        print(x, keyboard.keys[x])
    print(keyboard)
    keyboard.randomize()
    print(keyboard)
    
def play(genNumber, popSize, fileName):
    pop = [keyBoard.KeyBoard("random", seed=42) for x in range(popSize)]
    genetik.sortPop(pop, 0, popSize - 1)
    for gen in range(genNumber):
        genetik.evolve(pop, fileName)
        print(pop[0], gen)# Print the best keyboard for each generation

if __name__ == "__main__":
    import argparse
    import sys
    parser = argparse.ArgumentParser("python3 " + str(sys.argv[0]))
    parser.add_argument("klf",
                        metavar="KLF", type=argparse.FileType("r+"), default="./keys.log",
                        help="to add the keylogger file (KLF) to train the programm to")
    parser.add_argument("--len",
                        metavar="nb", type=int, default=10,
                        help="represent the lenght of the learning")
    parser.add_argument("--ps",
                        metavar="nb", type=int, default=10,
                        help="represent the population size to evolve")
    parser.add_argument("--seed",
                        metavar="seed", type=int, default=42,
                        help="to set the seed")
    parser.add_argument("--verbose", "-v", action="count", default=0,
                        help="add as many as you want to set the verbose level")
    args = parser.parse_args()
    print(args.klf.readlines())
    print("len: {}, popsize:{}, seed: {}, verbose: {}".format(args.len, args.ps, args.seed, args.verbose))
    main(args)
