#include <fstream>
#include <iostream>
#include <vector>

int main(int argc, char **argv) {
  std::vector<std::string> args(argv, argv + argc);
  std::ofstream tty;
  tty.open("/dev/tty");

  if (args.size() <= 1 || (args.size() & 2) == 1) {
    std::cerr << "usage: maplabel [devlocal remote]... remotedir\n";
    return 1;
  }

  std::string curdir = args[args.size() - 1];
  for (size_t i = 1; i + 1 < args.size(); i += 2) {
    if (curdir.find(args[i + 1]) == 0) {
      if ((curdir.size() > args[i + 1].size() &&
           curdir[args[i + 1].size()] == '/') ||
          curdir.size() == args[i + 1].size()) {
        tty << "\033];" << args[i] + curdir.substr(args[i + 1].size())
            << "\007\n";
        return 0;
      }
    }
  }

  tty << "\033];" << curdir << "\007\n";
  return 0;
}
