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
        for x in range(dim):
            self.g.append([-1]*dim)

    def addLink(self, p1, p2, val):
        """
        This function set the path from <p1>, to <p2> with the value <val> 
        """
        self.g[p1][p2] = val

    def getPath(self, p1, p2, prev=None):
        """
        This fuction return (<l>, <p>) as:
            l: is a lenght of the path to take
            p: is the list of points
        if no route is found, return (-1, [])
        """
        pass

    def getPathRec(self, pathL, path, cur, end):
        """
        pathL: path lenght
        path: list of points for this path
        curr: current point'd ID
        end: ID of the end point
        """
        if cur == end:
            return (pathL, path)
        else:
            res = []
            for tries in range(len(self.g[cur])):
                if self.g[cur][tries] != -1\
                    and tries != cur  #don't take care of loop back

                    res.append(self.getPathRec(pathL+self.g[cur][tries],
                        path.append(tries), tries, end))