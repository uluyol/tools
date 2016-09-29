#include <stdlib.h>

char __latencybench_readbuf[4096];

size_t reqbufalign(char *devpath) {
	return 0;
}

unsigned long bufalign() {
	return __alignof__(__latencybench_readbuf);
}
