CFLAGS ?= -O2 -Wall
BINDIR ?= $(HOME)/bin

install: $(BINDIR)/humannum $(BINDIR)/srcsearch

$(BINDIR)/humannum: humannum/main.c
	$(CC) $(CFLAGS) -o $(BINDIR)/humannum $<

$(BINDIR)/srcsearch: srcsearch/main.cc
	$(CXX) -std=c++11 $(CFLAGS) -o $(BINDIR)/srcsearch $<
