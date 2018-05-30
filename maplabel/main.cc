#include <iostream>
#include <vector>

int main(int argc, char **argv) {
  std::vector<std::string> args(argv, argv + argc);

  if (args.size() <= 1 || (args.size() & 2) == 1) {
    std::cerr << "usage: maplabel [devlocal remote]... remotedir\n";
    return 1;
  }

  std::string curdir = args[args.size() - 1];
  for (size_t i = 1; i + 1 < args.size(); i += 2) {
    if (curdir.find(args[i + 1]) == 0 &&
        (curdir[curdir.size() - 1] == '/' ||
         curdir.size() == args[i + 1].size())) {
      std::cout << args[i] + curdir.substr(args[i + 1].size()) << "\n";
      return 0;
    }
  }

  std::cout << curdir << "\n";
  return 0;
}
