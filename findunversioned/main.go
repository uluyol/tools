package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var isVCS = map[string]bool{
	".git": true,
	".hg":  true,
	".bzr": true,
	".svn": true,
}

func containsVCS(names []string) bool {
	for _, n := range names {
		if isVCS[n] {
			return true
		}
	}
	return false
}

func walk(dir string) ([]string, error) {
	var nover []string
	fi, err := os.Stat(dir)
	if err != nil || !fi.IsDir() {
		return nover, err // err will be nil if fi is not a directory
	}
	d, err := os.Open(dir)
	if err != nil {
		return nover, err
	}
	subfi, err := d.Readdir(-1)
	d.Close()
	if err != nil {
		return nover, err
	}
	var names []string
	for i := 0; i < len(subfi); {
		if !subfi[i].IsDir() {
			if i < len(subfi)-1 {
				subfi[i] = subfi[len(subfi)-1]
			}
			subfi = subfi[:len(subfi)-1]
			continue
		}
		names = append(names, subfi[i].Name())
		i++
	}
	if containsVCS(names) {
		return nover, nil
	}
	for _, n := range names {
		if nv, err := walk(filepath.Join(dir, n)); err != nil {
			return nover, err
		} else {
			nover = append(nover, nv...)
		}
	}
	if len(nover) == len(names) {
		return []string{dir}, nil
	}
	return nover, nil
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [dir1] [dir2] [...]\n", os.Args[0])
		os.Exit(1)
	}
	var nover []string
	for _, d := range os.Args[1:] {
		if nv, err := walk(d); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		} else {
			nover = append(nover, nv...)
		}
	}
	fmt.Println(strings.Join(nover, "\n"))
}
