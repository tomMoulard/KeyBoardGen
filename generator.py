# -*- coding: utf-8 -*-
"""
Created on tue Jun 8 20:00:00 2017

@author: tm
This uses everithing he can to create a powerfull keyboard
"""

# imports
import usefullFunk
import keyBoard

def main():
    keyboard = keyBoard.KeyBoard("qwerty", seed=42)
    keyboard.toGraph()
    

if __name__ == '__main__':
    main()
    