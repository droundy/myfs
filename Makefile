# Copyright 2010 David Roundy, roundyd@physics.oregonstate.edu.
# All rights reserved.

all: myfs

clean:
	rm -f */*.$(O) */*/*.$(O) */*/*/*.$(O) bin/*

include $(GOROOT)/src/Make.$(GOARCH)

.PHONY: clean
.SUFFIXES: .$(O) .go

main.$(O): myfs.go main.go
	$(GC) -o $@ $^

myfs: main.$(O)
	$(LD) -o $@ $<
