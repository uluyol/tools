#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main(const int argc, char **argv) {
  char prettybuf[1024];

  char *num = NULL;
  char *pretty = prettybuf;
  ssize_t i = 0;
  ssize_t numlen = 0;
  int nz = 0;
  if (argc != 2) {
    fprintf(stderr, "humannum number\n");
    return 1;
  }

  num = argv[1];

  numlen = strlen(num);
  for (i = numlen - 1; i >= 0; i--) {
    if (num[i] == '0') {
      nz++;
    } else {
      break;
    }
  }

  if (numlen > sizeof prettybuf) {
    pretty = calloc(numlen, 1);
    if (pretty == NULL) {
      puts(num);
      return 0;
    }
  }

  if (3 <= nz && nz < 6) {
    strncpy(pretty, num, numlen - 3);
    pretty[numlen - 3] = 'K';
    pretty[numlen - 2] = '\0';
  } else if (6 <= nz && nz < 9) {
    strncpy(pretty, num, numlen - 6);
    pretty[numlen - 6] = 'M';
    pretty[numlen - 5] = '\0';
  } else if (9 <= nz && nz < 12) {
    strncpy(pretty, num, numlen - 9);
    pretty[numlen - 9] = 'G';
    pretty[numlen - 8] = '\0';
  } else if (12 <= nz) {
    strncpy(pretty, num, numlen - 12);
    pretty[numlen - 12] = 'T';
    pretty[numlen - 11] = '\0';
  } else {
    pretty = num;
  }
  puts(pretty);
}
