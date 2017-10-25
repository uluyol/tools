#include <fcntl.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/resource.h>
#include <sys/time.h>
#include <unistd.h>

void usage(void) {
  fprintf(stderr, "usage: runbg [-l logpath] -- command args...\n");
  fprintf(stderr,
          "    -l logpath  path to log stderr/stdout (default bgproc.log)\n");
  fprintf(stderr, "    -v          enable verbose logging\n");
  exit(2);
}

int main(const int argc, char **argv) {
  char *logfile = "bgproc.log";
  bool verbose = false;
  int i, c;

  while ((c = getopt(argc, argv, "hvl:")) != -1) {
    switch (c) {
    case 'l':
      logfile = optarg;
      break;
    case 'v':
      verbose = true;
      break;
    case 'h':
    case '?':
      printf("unkown %c\n", optopt);
      usage();
      break;
    }
  }

  if (optind >= argc) {
    usage();
  }

  char *cmd = argv[optind];
  char **args = calloc(argc - optind + 1, sizeof(char *));
  for (i = optind; i < argc; i++) {
    args[i - optind] = argv[i];
  }

  if (verbose) {
    fprintf(stderr, "command: %s\n", cmd);
    fprintf(stderr, "narg: %d\n", argc - optind);
    for (i = 0; i < argc - optind; i++)
      fprintf(stderr, "arg %d: %s\n", i, args[i]);
  }

  int fdout = creat(logfile, 0644);
  if (fdout < 0) {
    perror("runbg: unable to create log file");
    exit(3);
  }

  pid_t pid;
  struct rlimit lim;

  getrlimit(RLIMIT_NOFILE, &lim);

  /* close all open files--NR_OPEN is overkill, but works */
  for (i = 3; i < lim.rlim_cur; i++) {
    close(i);
  }

  pid = fork();
  if (pid == -1) {
    return -1;
  } else if (pid != 0) {
    close(fdout);
    exit(0);
  }

  /* create new session and process group */
  if (setsid() == -1) {
    return -1;
  }

  pid = fork();
  if (pid == -1) {
    return -1;
  } else if (pid != 0) {
    exit(0);
  }

  int fd = open("/dev/null", O_RDONLY);
  fdout = creat(logfile, 0644);

  dup2(fd, 0);
  dup2(fdout, 1);
  dup2(fdout, 2);

  execvp(cmd, args);
}
