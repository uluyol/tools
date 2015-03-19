/*
This tool recursively searches $SRCSEARCHROOT for the directory queried and will return the path
of the most shallow result. Directories are searched in lexicographic order.

A bash include file is included with this tool that adds a scd command which uses srcsearch to
quickly cd into a workspace.
*/

package main

import (
	"container/list"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const envRootVar = "SRCSEARCHROOT"

var (
	errNotFound = errors.New("could not find directory")

	ignoreHidden = flag.Bool("ignorehidden", true, "ignore hidden directories")
	maxDepth     = flag.Int("maxdepth", 5, "maximum search depth")
)

type location struct {
	path  string
	depth int
}

func search(dir string, names []string, startdepth int) (path string, err error) {
	q := list.New()
	q.PushBack(location{dir, startdepth})
	for q.Len() > 0 {
		front := q.Front()
		cloc := q.Remove(front).(location)
		entries, err := ioutil.ReadDir(cloc.path)
		if err != nil {
			return "", err
		}
		for _, e := range entries {
			if !e.IsDir() || (*ignoreHidden && strings.HasPrefix(e.Name(), ".")) {
				continue
			}
			absPath := cloc.path + "/" + e.Name()
			if e.Name() == names[0] {
				if len(names) == 1 {
					return absPath, nil
				} else if subPath, err := search(absPath, names[1:], cloc.depth+1); err == nil {
					return subPath, nil
				}
			}
			if cloc.depth < *maxDepth {
				q.PushBack(location{absPath, cloc.depth + 1})
			}
		}
	}
	return "", errNotFound
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s dirname[/subdir/...]\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	searchpath := flag.Arg(0)
	root := os.Getenv(envRootVar)
	if root == "" {
		fmt.Fprintln(os.Stderr, envRootVar+" must be set.")
		os.Exit(2)
	}
	names := strings.Split(searchpath, string(os.PathSeparator))
	p, err := search(root, names, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
	fmt.Println(p)
}
