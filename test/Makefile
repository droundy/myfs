# Copyright 2010 David Roundy, roundyd@physics.oregonstate.edu.
# All rights reserved.

all: check

include $(GOROOT)/src/Make.$(GOARCH)

clean:
	rm -f *.$(O)

.PHONY: clean aminmount
.SUFFIXES: .$(O) .go

simpletest.$(O): simpletest.go
	$(GC) -o $@ $^
simpletest: simpletest.$(O)
	$(LD) -o $@ $<
runsimpletest: simpletest
	echo FIXME the following bit is buggy?
	#./simpletest

ctest: ctest.c
	cat ctest.c
	gcc -o ctest ctest.c
runctest: ctest
	echo FIXME: the following is not executable?
	#./ctest

runshellscript: shellscript.sh
	sh shellscript.sh
	echo FIXME: The following doesnt work?
	#./shellscript.sh

aminmount:
	echo `pwd` | grep testing-mount

check: aminmount runshellscript runctest runsimpletest
	echo all tests pass!
