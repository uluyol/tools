CFLAGS ?= -O2 -Wall
BINDIR ?= $(HOME)/bin

$(BINDIR)/humannum: humannum/main.c
	$(CC) -o $(BINDIR)/humannum humannum/main.c
