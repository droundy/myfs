# Copyright 2010 David Roundy, roundyd@physics.oregonstate.edu.
# All rights reserved.

all: myfs

clean:
	rm -f */*.$(O) */*/*.$(O) */*/*/*.$(O) bin/* myfs

include $(GOROOT)/src/Make.$(GOARCH)

.PHONY: clean
.SUFFIXES: .$(O) .go

main.$(O): main.go sillyfiles.go
	$(GC) -o $@ $^

myfs: main.$(O)
	$(LD) -o $@ $<

check: all
	./test.sh
