/*
This tool recursively searches $SRCSEARCHROOT for the directory queried and will
return the path of the most shallow result. Directories are searched in
lexicographic order.

A bash include file is included with this tool that adds a scd command which
uses srcsearch to quickly cd into a workspace.
*/

#include <dirent.h>
#include <iostream>
#include <list>
#include <sstream>
#include <stdlib.h>
#include <string>
#include <vector>

static const std::string kEnvRootVar("SRCSEARCHROOT");
static const std::string kErrNotFound("could not find directory");

static bool IGNORE_HIDDEN = true;
static int MAX_DEPTH = 5;

void split(std::vector<std::string> *out, const std::string &s, char delim) {
  std::stringstream ss;
  ss.str(s);
  std::string item;
  while (std::getline(ss, item, delim)) {
    out->push_back(item);
  }
}

struct Loc {
  std::string path_;
  int depth_;
  Loc(const std::string p, int startdepth);
};

Loc::Loc(const std::string p, int startdepth) {
  path_ = p;
  depth_ = startdepth;
}

struct DirDeleter {
  DirDeleter(){};
  DirDeleter(const DirDeleter &){};
  DirDeleter(DirDeleter &){};
  DirDeleter(DirDeleter &&){};
  void operator()(DIR *d) const { closedir(d); };
};

std::tuple<std::string, bool> search(std::string dir,
                                     const std::vector<std::string> names,
                                     const int ni, const int startdepth) {
  std::list<Loc> q;
  q.push_back(Loc(dir, startdepth));
  while (q.size() > 0) {
    Loc cloc = q.front();
    q.pop_front();

    struct dirent *dp = NULL;
    std::unique_ptr<DIR, DirDeleter> d(opendir(cloc.path_.c_str()),
                                       DirDeleter());
    if (d == nullptr) {
      perror("srcsearch");
      exit(1);
    }
    while ((dp = readdir(d.get())) != NULL) {
      if (dp->d_type != DT_DIR) {
        continue;
      }
      if (IGNORE_HIDDEN && dp->d_name[0] == '.') {
        continue;
      }

      std::string basename = std ::string(dp->d_name);
      std::string abspath = cloc.path_ + "/" + dp->d_name;
      if (basename == names[ni]) {
        if (names.size() - ni == 1) {
          return std::make_tuple(abspath, true);
        }
        std::string subpath;
        bool ok;
        std::tie(subpath, ok) = search(abspath, names, ni + 1, cloc.depth_ + 1);
        if (ok) {
          return std::make_tuple(subpath, ok);
        }
      }
      if (cloc.depth_ < MAX_DEPTH) {
        q.push_back(
            Loc(cloc.path_ + "/" + std::string(dp->d_name), cloc.depth_ + 1));
      }
    }
  }
  return std::make_tuple("", false);
}

void usage(std::string name) {
  std::cerr << "Usage: " << name << " dirname[/subdir/...]\n";
  std::cerr << "\t-ignorehidden=false\tsearch hidden directories\n";
  std::cerr << "\t-maxdepth N\tmaximum search depth (default 5)\n";
  exit(3);
}

int main(const int argc, const char **argv) {
  std::string tofind("");
  for (int i = 1; i < argc; i++) {
    if (strcmp(argv[i], "-ignorehidden=false") == 0) {
      IGNORE_HIDDEN = false;
    } else if (strcmp(argv[i], "-maxdepth") == 0) {
      i++;
      if (i >= argc) {
        usage(argv[0]);
      }
      MAX_DEPTH = std::atoi(argv[i]);
    } else {
      tofind = argv[i];
    }
  }

  if (tofind.empty()) {
    usage(argv[0]);
  }

  std::string root = getenv(kEnvRootVar.c_str());
  if (root == "") {
    std::cerr << kEnvRootVar << " is not set\n";
    return 6;
  }

  std::vector<std::string> names;
  split(&names, tofind, '/');
  std::string p;
  bool ok;
  std::tie(p, ok) = search(root, names, 0, 0);
  if (!ok) {
    std::cerr << kErrNotFound << "\n";
    return 99;
  }
  std::cout << p << "\n";
}
