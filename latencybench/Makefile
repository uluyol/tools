#CFLAGS += -fno-omit-frame-pointer -fsanitize=address -g -O0
CFLAGS += -Wall -O2

TARGETS = latencybench latencybench-linux-cross

SRCS_LINUX = main.c main_linux.c
SRCS = main.c

ifeq ($(shell uname), Linux)
	SRCS += main_linux.c
else
	SRCS += main_unix.c
endif

all: $(TARGETS)

latencybench: $(SRCS) *.h
	$(CC) $(CFLAGS) $(SRCS) -o $@

latencybench-linux-cross: $(SRCS_LINUX) *.h
	docker run --rm -v $(CURDIR):/w gcc:6 \
		gcc $(CFLAGS) $(addprefix /w/,$(SRCS_LINUX)) -o /w/$@

clean:
	rm -f $(TARGETS)
