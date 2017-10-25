CFLAGS ?= -O2 -Wall
BINDIR ?= $(HOME)/bin

install: $(BINDIR)/humannum $(BINDIR)/srcsearch $(BINDIR)/runbg $(BINDIR)/runbg-linux

$(BINDIR)/humannum: humannum/main.c
	$(CC) $(CFLAGS) -o $@ $<
	strip $@

$(BINDIR)/srcsearch: srcsearch/main.cc
	$(CXX) -std=c++11 $(CFLAGS) -o $@ $<
	strip $@

$(BINDIR)/runbg: runbg/runbg.c
	$(CC) $(CFLAGS) -o $@ $<
	strip $@

$(BINDIR)/runbg-linux: runbg/runbg.c
	docker run --rm -v $(CURDIR)/runbg:/w gcc:7 gcc $(CFLAGS) -o /w/runbg-linux.bin /w/runbg.c
	docker run --rm -v $(CURDIR)/runbg:/w gcc:7 strip /w/runbg-linux.bin
	mv runbg/runbg-linux.bin $@
