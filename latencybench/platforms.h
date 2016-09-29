#ifdef __linux__
	#define IS_ACCURATE true
	#define _GNU_SOURCE
#else
	#define IS_ACCURATE false
	#define O_DIRECT 0
#endif

#include <stdlib.h>

extern char __latencybench_readbuf[4096];
int reqbufalign(char *devpath);
unsigned long bufalign();
