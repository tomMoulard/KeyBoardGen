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

# random
import random

DEFAULTLAYOUT = [# some fixing TODO
    [
        [" # default mode"],
        ["`", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "-", "=", "{backspace}"],
        ["{tab}", "q", "w", "e", "r", "t", "y", "u", "i", "o", "p", "[", "]", "\\"],
        ["{caps}", "a", "s", "d", "f", "g", "h", "j", "k", "l", ";", "'", "{enter}"],
        ["{shiftl}", "z", "x", "c", "v", "b", "n", "m", ",", ".", "/", "{shiftr}"],
        ["{next}", "{space}", "{accept}" ],
    ],[
        [" # shifted mode"],
        ["~", "!", "@", "#", "$", "%", "^", "&", "*", "(", ")", "_", "+", "{backspace}"],
        ["{tab}" ,"Q" ,"W" ,"E" ,"R" ,"T" ,"Y" ,"U" ,"I" ,"O" ,"P" ,"{" ,"}" ,"|"],
        ["{caps}", "A", "S", "D", "F", "G", "H", "J", "K", "L", ":", "\"", "{enter}"],
        ["{shiftl}", "Z", "X", "C", "V", "B", "N", "M", "<", ">", "?", "{shiftr}"],
        ["{next}", "{space}", "{accept}" ],
    ],[
        [" # capsed mode"],
        ["`", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "-", "=", "{backspace}"],
        ["{tab}", "Q", "W", "E", "R", "T", "Y", "U", "I", "O", "P", "[", "]", "\\"],
        ["{caps}", "A", "S", "D", "F", "G", "H", "J", "K", "L", ";", "'", "{enter}"],
        ["{shiftl}", "Z", "X", "C", "V", "B", "N", "M", ",", ".", "/", "{shiftr}"],
        ["{next}", "{space}", "{accept}" ],

    ]
]
QWERTY = DEFAULTLAYOUT

AZERTY = [
    [
        [" # default mode"],
        ["²", "&", "é", "\"", "'", "(", "-", "è", "_", "ç", "à", ")", "=", "{backspace}"],
        ["{tab}", "a", "z", "e", "r", "t", "y", "u", "i", "o", "p", "^", "$", "{enter}"],
        ["{caps}", "q", "s", "d", "f", "g", "h", "j", "k", "l", "m", "ù", "*", "{enter}"],
        ["{shiftl}", "<", "w", "x", "c", "v", "b", "n", ",", ";", ":", "!", "{shiftr}"],
        ["{next}", "{space}", "{alt-gr}" ],
    ],[
        [" # shifted mode"],
        ["~", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "°", "+", "{backspace}"],
        ["{tab}" ,"A" ,"Z" ,"E" ,"R" ,"T" ,"Y" ,"U" ,"I" ,"O" ,"P" ,"¨" ,"£", "{enter}"],
        ["{caps}", "Q", "S", "D", "F", "G", "H", "J", "K", "L", "M", "%", "µ", "{enter}"],
        ["{shiftl}", ">", "W", "X", "C", "V", "B", "N", "?", ".", "/", "§", "{shiftr}"],
        ["{next}", "{space}", "{alt-gr}" ],
    ],[
        [" # alt-gr mode"],
        ["¬", "¹", "~", "#", "{", "[", "|", "`", "\\", "^", "@", "]", "}", "{backspace}"],
        ["{tab}", "æ", "«", "€", "¶", "ŧ", "←", "↓", "→", "ø", "þ", "¨", "¤", "{enter}"],
        ["{caps}", "@", "ß", "ð", "đ", "ŋ", "ħ", "̉", "ĸ", "ł", "µ", "^", "`", "{enter}"],
        ["{shiftl}", "ł", "»", "¢", "“", "”", "n", "´", "", "·", "̣", "{shiftr}"],
        ["{next}", "{space}", "{alt-gr}" ],
    ]
]

class KeyBoard():
    """
    This is the keyboard class to manage a keyboard and give him a grade
    if you want to create a keyboard with other char you can make a new basic
    layout like :
    [
        [
            [" # default mode"],
            ["`", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "-", "=", "{backspace}"],
            ["{tab}", "q", "w", "e", "r", "t", "y", "u", "i", "o", "p", "[", "]", "\\", ],
            ["{caps}", "a", "s", "d", "f", "g", "h", "j", "k", "l", ";", "'", "{enter}", ],
            ["{shiftl}", "z", "x", "c", "v", "b", "n", "m", ",", ".", "/", "{shiftr}", ],
            ["{next}", "{space}", "{accept}" ],
        ],[
            [" # shifted mode"],
            ["~", "!", "@", "#", "$", "%", "^", "&", "*", "(", ")", "_", "+", "{backspace}", ],
            ["{tab}" ,"Q" ,"W" ,"E" ,"R" ,"T" ,"Y" ,"U" ,"I" ,"O" ,"P" ,"{" ,"}" ,"|", ],
            ["{caps}", "A", "S", "D", "F", "G", "H", "J", "K", "L", ":", "\"", "{enter}", ],
            ["{shiftl}", "Z", "X", "C", "V", "B", "N", "M", "<", ">", "?", "{shiftr}", ],
            ["{next}", "{space}", "{accept}" ],
        ],[
            [" # capsed mode"],
            ["`", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "-", "=", "{backspace}", ],
            ["{tab}", "Q", "W", "E", "R", "T", "Y", "U", "I", "O", "P", "[", "]", "\\", ],
            ["{caps}", "A", "S", "D", "F", "G", "H", "J", "K", "L", ";", "'", "{enter}", ],
            ["{shiftl}", "Z", "X", "C", "V", "B", "N", "M", ",", ".", "/", "{shiftr}", ],
            ["{next}", "{space}", "{accept}" ],

        ]
    ]
    """
    def __init__(self, mode, keys=[], seed=None):
        self.mode = mode
        self.keyLayout = keys
        self.seed = seed
        self.graph = None
        self.keys = []# This is a key list, and a id list
        random.seed(self.seed)
        if self.mode == "azerty":  
            self.keyLayout = AZERTY
        elif self.mode == "qwerty":
            self.keyLayout = QWERTY
        elif self.mode == "random":
            self.keyLayout = DEFAULTLAYOUT
            self.randomize()
        elif self.mode == "set":
            self.keyLayout = keys
        else:
            print("No default keyboard provided, taking a generic one")
            self.keyLayout = DEFAULTLAYOUT
        self.keys = self.getAllLetters()
    
    def __str__(self):
        """
        Is the function that return a str value of the keyboard
        """
        res = "[\n"
        for mode in self.keyLayout:
            for line in mode[:-1]:
                res += "  [ "
                for letter in line:
                    res += letter + " "
                res += "],\n"
            res += "  [ "
            for letter in mode[-1]:
                res += letter + " "
            res += "]\n\n"
        res += "]\n"
        return res

    def getAllLetters(self):
        """
        This return an array of all letters on the keyboard
        """
        letters = []
        for mode in self.keyLayout:
            for letter in mode[1:]:
                letters += letter
        return letters

    def randomize(self):
        """
        This randomize a keyboard entirelly
        """
        # self.keyLayout.usefullFunk.randomize()
        letters = []
        # layout = []
        # for mode in self.keyLayout:
        #     # extract all letters to randomize but keeping the format
        #     l = []
        #     for letter in mode[1:]:
        #         letters += letter
        #         l.append(len(letter))
        #     layout += l
        # usefullFunk.randomizeList(letters, seed=self.seed)
        # print(layout)

    def getIdLetter(self, l):
        """
        Give back the ID of the letter <l> -1 otherwise
        """
        if l in self.keys:
            return self.keys.index(l)
        else:
            return -1


    def toGraph(self):
        """
        This function has the purpose to full in the self.graph property
        to allow a good valurization of the keyboard
        """
        import graphTools
        self.graph = graphTools.Graph(len(self.keys))

        # now adding links
        # in line first
        # [[["a"]]]
        # modes letters
        for modes in self.keyLayout:
            for lines in modes[1:]:
                for letters in range(len(lines)-1):
                    idL1 = self.getIdLetter(lines[letters])
                    idL2 = self.getIdLetter(lines[letters])
                    if not (idL1 == -1 or idL2 == -1):
                        self.graph.addLink(idL1, idL2, 1)
                        #also adding the reverser cause it's an oriented graph
                        self.graph.addLink(idL2, idL1, 1)
        # then bottoms up and diaonals
        for modes in self.keyLayout:
            for lines in range(len(modes)-1):
                for letters in range(min(len(modes[lines]), len(modes[lines+1]))):
                    idL1 = self.getIdLetter(modes[lines][letters])
                    idL2 = self.getIdLetter(modes[lines+1][letters])
                    if not (idL1 == -1 or idL2 == -1):
                        self.graph.addLink(idL1, idL2, 1)
                        #also adding the reverser cause it's an oriented graph
                        self.graph.addLink(idL2, idL1, 1)