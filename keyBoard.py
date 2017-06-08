# -*- coding: utf-8 -*-
"""
Created on tue Jun 8 20:00:00 2017

@author: tm
This contain everything about keyboard
To use:
k = keyboard("random|qwerty|azerty")
k.note()
k.randomize()
"""

# imports
import usefullFunk


class KeyBoard():
    """
    This is the keyboard class to manage a keyboard and give him a grade
    """
    def __init__(self, mode, keys=[]):
        self.mode = mode
        self.keyLayout = keys
        if self.mode == "azerty":
            self.keyLayout = [] #TODO: fill me
        else:
            self.keyLayout = [] #TODO: fill me
        self.keyLayout = keys
        if self.mode == "random":
            self.randomize()
    
    def __str__(self):
        usefullFunk.printList(self.keyLayout)

    def randomize(self):
        # self.keyLayout.usefullFunk.randomize()
        pass
