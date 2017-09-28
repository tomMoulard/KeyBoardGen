# -*- coding: utf-8 -*-
"""
Created on tue sept 12 12:00:00 2017

@author: tm
This is some tools for graphs
"""

class Graph():
    """
    Class for an oriented graph:
    [[-1,-1],
    [1,2]]
    The graph have a link from <1> to <0> with a weight of 1
    and one from <1> to <1> with a weight of 2
    """
    def __init__(self, dim):
        self.dim = dim
        self.g   = []
        self.parcour = []
        for x in range(dim):
            self.g.append([-1]*dim)
        self.buildParcour()

    def addLink(self, p1, p2, val):
        """
        This function set the path from <p1>, to <p2> with the value <val> 
        """
        self.g[p1][p2] = val

    def buildParcour(self):
        """
        GOAL: fill in the self.parcour as a list of graph
        list of graph for each letter as a starting point
        """
        self.parcour = [[([], -1) for x in range(self.dim)]\
                                    for y in range(self.dim)]

    def getPath(self, p1, p2):
        """
        <p1> and <p2> lettters ID
        This fuction return (<p>, <l>) as:
            p: is the list of points
            l: is a lenght of the path to take
        if no route is found, return ([], -1)
        """
        return self.parcour[p1][p1][p2]

    def toDot(self, name="my_graph"):
        """
        This function transphorm the graph into DOT format
        It is usefull to print it with Graphviz
        """
        res = "digraph " + name + "{\n"
        for y in range(self.dim):
            for x in range(self.dim):
                if self.g[x][y] != -1:
                    res += "{} -> {};\n".format(x, y)
        return res + "}"