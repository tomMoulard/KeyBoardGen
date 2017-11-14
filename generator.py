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

def main():
    keyboard = keyBoard.KeyBoard("qwerty", seed=42)
    keyboard.toGraph()
    
def play(genNumber, popSize, fileName):
    pop = [keyBoard.KeyBoard("random", seed=42) for x in range(popSize)]
    genetik.sortPop(pop, 0, popSize - 1)
    for gen in range(genNumber):
        genetik.evolve(pop, fileName)
        print(pop[0], gen)# Print the best keyboard for each generation

if __name__ == '__main__':
    main()
    