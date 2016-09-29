#include "platforms.h"

#include <fcntl.h>
#include <inttypes.h>
#include <unistd.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

static char *workpath = "latencybench.data.bin";

void usage() {
	fprintf(stderr, "usage: latencybench fsize nops devpath\n");
	fprintf(stderr, "\nfsize must be in MB\n");
	fprintf(stderr,
		"\nlatencybench will issue random 1K reads and "
		"output latency measurements to stdout.\n");
	fprintf(stderr,
		"latencybench will create a temporary file named %s "
		"in the current directory.\n", workpath);
	fprintf(stderr, "Measurements will only be accurate on Linux.\n");
	exit(123);
}

int64_t diff_time_us(struct timespec start, struct timespec end) {
	int64_t sec_diff = end.tv_sec - start.tv_sec;
	int64_t ns_diff = end.tv_nsec - start.tv_nsec;
	return (int64_t)(sec_diff*1000000L + ns_diff/1000L);
}

char *getreadbuf() {
	return __latencybench_readbuf;
}

int main(int argc, char **argv) {
	if (argc != 4) {
		usage();
	}

	ssize_t fsize_mb = atol(argv[1]);
	int32_t nops = atoi(argv[2]);
	char *devpath = argv[3];

	if (fsize_mb == 0 || nops == 0) {
		usage();
	}

	if (!IS_ACCURATE) {
		fprintf(stderr, "warn: latencies will be inaccurate, see help for details\n");
	} else {
		fprintf(stderr, "require buffer alignment of %d, assuming %lu\n", reqbufalign(devpath), bufalign());
	}

	char wbuf[1024*1024];
	int fdrand = open("/dev/urandom", O_RDONLY);

	fprintf(stderr, "writing random data to %s\n", workpath);
	int wfd = creat(workpath, 0666);

	ssize_t wrote = 0;
	while (wrote < fsize_mb*1024*1024) {
		ssize_t got = read(fdrand, wbuf, 1024*1024);
		wrote += write(wfd, wbuf, got);
	}

	fprintf(stderr, "wrote %ld bytes\n", (long)wrote);

	close(fdrand);
	close(wfd);

	fprintf(stderr, "collecting latency measurements\n");
	char *rbuf = getreadbuf();
	int rfd = open(workpath, O_RDONLY | O_DIRECT);
	if (rfd == -1) {
		perror("unable to open data file");
		return 3;
	}
	int64_t *latencies = calloc(nops, sizeof(int64_t));

	for (int32_t i=0; i < nops; i++) {
		struct timespec start, end;
		off_t offset = rand() % (1024*1024*fsize_mb);
		offset /= 4096;
		offset *= 4096;
		clock_gettime(CLOCK_MONOTONIC, &start);
		if (lseek(rfd, offset, SEEK_SET) != offset) {
			perror("unable to seek");
		}
		for (int sofar=0; sofar < 4096; ) {
			int got = read(rfd, rbuf, 4096-sofar);
			if (got == -1) {
				perror("unable to read data");
			}
			sofar += got;
		}
		clock_gettime(CLOCK_MONOTONIC, &end);
		latencies[i] = diff_time_us(start, end);
	}

	close(rfd);
	remove(workpath);

	for (int32_t i=0; i < nops; i++) {
		printf("%" PRId64 "\n", latencies[i]);
	}
}
