SHELL:=/bin/bash

.PHONY: test
test: test.in test.out
	diff $^

.INTERMEDIATE: test.out
test.out: test.in np test.py 
	cat $< | ./np -n 4 python test.py | sort --general-numeric-sort >$@

.INTERMEDIATE: test.in
test.in:
	for i in {1..10} ; do echo $$i ; done >$@

.INTERMEDIATE: np
np: main.go
	go build -o np

.PHONY: install
install: np
	sudo cp $< /usr/local/bin/np
