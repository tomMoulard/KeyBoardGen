# -*- coding: utf-8 -*-
"""
Created on tue Jun 8 20:00:00 2017

@author: tm
This is a set of usefull functions for debuging or simple stuff
"""

import random

def printList(l, debug=False):
    """
    This is supposed to print a list with a pretty layout
    like :
    >>> l = [1, 2, 3,]
    >>> l
    [1, 2, 3]
    >>> printList(l)
    [1,
    2,
    3]
    >>> l = ["This\n", "is\n", "a\n", "string\n",]
    >>> l
    ['This\n', 'is\n', 'a\n', 'string\n']
    >>> printList(l)
    [This
    ,
    is
    ,
    a
    ,
    string
    ]
     >>> printList(l, debug=True)
    [['This\n'] ,
    ['is\n'] ,
    ['a\n'] ,
    ['string\n']]
    """
    print("[", end="")
    if debug == True:
        for x in l[:-1]:
            print([x], ",")
        print("{}]".format([l[-1]]))
    else:
        for x in l[:-1]:
            print("{},".format(x))
        print("{}]".format(l[-1]))

def swap(l, posX, posY):
    """
    This swap the posX eme element and the posY eme in l
    """
    l[posX], l[posY] = l[posY], l[posX]

def randomizeList(l, seed=random.random()):
    """
    This randomize the list
    """
    random.seed(seed)
    ll = len(l) - 1
    for x in range(ll):
        swap(l, x, random.randint(x+1, ll))