# -*- coding: utf-8 -*-
"""
Created on tue Jun 8 20:00:00 2017

@author: tm
This is a set of usefull functions for debuging or simple stuff
"""

def printList(l, debug=False):
    print("[", end="")
    if debug == True:
        for x in l[:-1]:
            print([x], ",")
    else:
        for x in l[:-1]:
            print(x, ",")
    print(l[-1], "]")