#!/usr/bin/python

"""Like /bin/cat output stdin to stdout line by line.

The difference is that this uses python's input buffering, which makes
it more tricky for fan (because multiple lines of input, possibly all,
can be consumed before any output is produced). Info on python input
buffering: http://stackoverflow.com/a/3670470/469997
"""

import sys

for line in sys.stdin:
    print(line[:-1])
