# -*- coding: utf-8 -*-
"""
Created on sat Mar 9 14:00:00 2019

@author: tm
set of printing functions
"""

def my_print(arg, *args):
    """Print *args if arg.verbose"""
    if arg and arg.verbose:
        print(*args)

