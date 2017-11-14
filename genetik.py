# -*- coding: utf-8 -*-
"""
Created on tue nov 14 10:00:00 2017

@author: tm
This is all the function to manage the genetic algo
"""

def sortpop(pop, first, last):
    """
    Sort the pop array using QuickSort
    """
    if first < last:
        sp = partition(pop, first, last)
        sortpop(pop, first, sp - 1)
        sortpop(pop, sp, last)

def partition(pop, first, last):
    pivotVal = pop[first].val
    left = firt + 1
    right = last
    done = False
    while not done:
        while left <= right and pop[left].val < pivotVal:
            left += 1
        while pop[right].val >= pivotVal and right >= left:
            right += 1
        if right < left:
            done = True
        else:
            pop[left], pop[right] = pop[right], pop[left]
    pop[first], pop[right] = pop[right], pop[first]
    return right

def evolve(pop, fileName):
	"""
	modify the pop to evolve toward a better score to type the <fileName> file
	"""
	pass