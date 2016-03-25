SHELL:=/bin/bash

test: args 100k long

# 100k short lines
100k: fan
	for i in {1..100000} ; do echo $$i ; done >/tmp/$@.in
	cat /tmp/$@.in | ./fan -n=4 python cat.py | sort --general-numeric-sort >/tmp/$@.out
	diff /tmp/$@.*

# 4 1MiB lines
long: fan
	@# `printf '%s' {1..200000}` generates just over 1MiB, and is 7x faster than `printf '.%.0s' {1..1048576}`
	@# setting the number to 2500000 will yield about 16MiB
	for i in {1..4} ; do echo `printf '%s' {1..200000}` ; done >/tmp/$@.in
	# setting n=16 even though there are only 4 lines to check if this is a problem
	cat /tmp/$@.in | ./fan -n=16 python cat.py >/tmp/$@.out
	diff /tmp/$@.*

args: fan
	# missing command
	echo "ok" | ./fan 2>/dev/null ; test "$$?" == "1"
	# invalid n
	echo "ok" | ./fan -n=666 2>/dev/null ; test "$$?" == "1"
	# invalid n
	echo "ok" | ./fan -n=-99 2>/dev/null ; test "$$?" == "1"
	# invalid n
	echo "ok" | ./fan -n=foo 2>/dev/null ; test "$$?" == "2"
	# correct ; smoke test
	echo "ok" | ./fan cat >/dev/null ; test "$$?" == "0"

.INTERMEDIATE: fan
fan: fan.go
	go build -o fan

.PHONY: install
install: fan
	sudo cp $< /usr/local/bin/fan
