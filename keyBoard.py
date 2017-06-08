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

defaultfLayout=[
    [ # default mode
        ["`", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "-", "=", "{backspace}"],
        ["{tab}", "q", "w", "e", "r", "t", "y", "u", "i", "o", "p", "[", "]", "\\", ],
        ["{caps}", "a", "s", "d", "f", "g", "h", "j", "k", "l", ";", "'", "{enter}", ],
        ["{shiftl}", "z", "x", "c", "v", "b", "n", "m", ",", ".", "/", "{shiftr}", ],
        ["{next}", "{space}", "{accept}" ],
        ],[ # shifted mode
        ["~", "!", "@", "#", "$", "%", "^", "&", "*", "(", ")", "_", "+", "{backspace}", ],
        ["{tab}" ,"Q" ,"W" ,"E" ,"R" ,"T" ,"Y" ,"U" ,"I" ,"O" ,"P" ,"{" ,"}" ,"|", ],
        ["{caps}", "A", "S", "D", "F", "G", "H", "J", "K", "L", ":", "\"", "{enter}", ],
        ["{shiftl}", "Z", "X", "C", "V", "B", "N", "M", "<", ">", "?", "{shiftr}", ],
        ["{next}", "{space}", "{accept}" ],
    ],[ # capsed mode
        ["`", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "-", "=", "{backspace}", ],
        ["{tab}", "Q", "W", "E", "R", "T", "Y", "U", "I", "O", "P", "[", "]", "\\", ],
        ["{caps}", "A", "S", "D", "F", "G", "H", "J", "K", "L", ";", "'", "{enter}", ],
        ["{shiftl}", "Z", "X", "C", "V", "B", "N", "M", ",", ".", "/", "{shiftr}", ],
        ["{next}", "{space}", "{accept}" ],
    ]
]

class KeyBoard():
    """
    This is the keyboard class to manage a keyboard and give him a grade
    """
    def __init__(self, mode, keys=defaultfLayout):
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
