package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

var (
	shouldDelete = flag.Bool("d", false, "Delete this bookmark")
	shouldList   = flag.Bool("l", false, "List all bookmarks")
)

type bookmark struct {
	name string
	path string
}

func (b bookmark) String() string {
	return b.name + "," + b.path
}

func parseBookmark(txt string) (bookmark, error) {
	var b bookmark
	split := strings.SplitN(txt, ",", 2)
	if len(split) != 2 {
		return b, errors.New("incorrect bookmark: " + txt)
	}
	b.name = split[0]
	b.path = split[1]
	return b, nil
}

type bookmarkList struct {
	bmks []bookmark
}

func (bl *bookmarkList) Len() int           { return len(bl.bmks) }
func (bl *bookmarkList) Swap(i, j int)      { bl.bmks[i], bl.bmks[j] = bl.bmks[j], bl.bmks[i] }
func (bl *bookmarkList) Less(i, j int) bool { return bl.bmks[i].name < bl.bmks[j].name }

func (bl *bookmarkList) String() string {
	s := ""
	for _, b := range bl.bmks {
		s += b.String()
	}
	return s
}

func (bl *bookmarkList) Add(name, path string) {
	bl.bmks = append(bl.bmks, bookmark{name, path})
}

func (bl *bookmarkList) Compact() {
	seen := make(map[string]bool)
	compact := make([]bookmark, 0, len(bl.bmks))
	for i := len(bl.bmks) - 1; i >= 0; i-- {
		if seen[bl.bmks[i].name] {
			continue
		}
		compact = append(compact, bl.bmks[i])
		seen[bl.bmks[i].name] = true
	}
	bl.bmks = compact
}

func (bl *bookmarkList) Remove(name string) {
	for i, b := range bl.bmks {
		if b.name == name {
			bl.bmks[i], bl.bmks = bl.bmks[len(bl.bmks)-1], bl.bmks[:len(bl.bmks)-1]
			break
		}
	}
}

func (bl *bookmarkList) dump(w io.Writer) {
	for _, b := range bl.bmks {
		fmt.Fprintf(w, "%s,%s\n", b.name, b.path)
	}
}

func (bl *bookmarkList) Write(p string) error {
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	sort.Sort(bl)
	bl.dump(w)
	return w.Flush()
}

func parseBookmarks(p string) (*bookmarkList, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	bl := make([]bookmark, 0, 10)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		b, err := parseBookmark(scanner.Text())
		if err != nil {
			return &bookmarkList{bl}, err
		}
		bl = append(bl, b)
	}
	return &bookmarkList{bl}, scanner.Err()
}

func fixPath(p string) string {
	if !strings.HasPrefix(p, "/") {
		cwd, err := os.Getwd()
		if err != nil {
			return p
		}
		return cwd + "/" + p
	}
	return p
}

func usage() {
	fmt.Fprintf(os.Stderr, "%s: [-dl] name [path]\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	bmkDir := os.Getenv("HOME") + "/.config/bmk"
	os.MkdirAll(bmkDir, 0755)
	bmkPath := bmkDir + "/bookmarks"
	bmks, err := parseBookmarks(bmkPath)
	if err != nil {
		if _, ok := err.(*os.PathError); !ok {
			fmt.Println("HERE")
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		} else {
			bmks = &bookmarkList{}
		}
	}
	if *shouldList {
		bmks.dump(os.Stdout)
		return
	}
	if *shouldDelete {
		if len(flag.Args()) != 1 {
			flag.Usage()
			os.Exit(1)
		}
		bmks.Remove(flag.Arg(0))
	} else {
		if len(flag.Args()) != 2 {
			flag.Usage()
			os.Exit(1)
		}
		bmks.Add(flag.Arg(0), fixPath(flag.Arg(1)))
	}
	bmks.Compact()
	if err := bmks.Write(bmkPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
