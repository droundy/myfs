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
	echo the following bit is buggy?
	#./simpletest

runshellscript: shellscript.sh
	./shellscript.sh

aminmount:
	echo `pwd` | grep testing-mount

check: aminmount runshellscript runsimpletest
	echo all tests pass!
