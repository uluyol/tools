#include <fcntl.h>
#include <linux/fs.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/ioctl.h>
#include <unistd.h>

char __latencybench_readbuf[4096] __attribute__ ((aligned(4096)));

size_t reqbufalign(char *devpath) {
	size_t align;
	int fd = open(devpath, 0);
	if (fd == -1) {
		perror("unable to open device");
		return 0;
	}
	if (ioctl(fd, BLKBSZGET, &align) != 0) {
		perror("unable to get correct alignment");
	}
	close(fd);
	return align;
}

unsigned long bufalign() {
	return __alignof__(__latencybench_readbuf);
}
