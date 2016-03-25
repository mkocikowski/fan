#!/usr/bin/python

import sys

# using buffered reads on purpose; they exposed a bug with prematurely
# closing stdout; info on buffering:
# http://stackoverflow.com/a/3670470/469997

for line in sys.stdin:
    print(line[:-1])
