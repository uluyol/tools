package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	flag "github.com/spf13/pflag"
)

var (
	fieldDelim  = flag.StringP("field-delim", "F", "", "field delimeter for tabular input (default whitespace)")
	colSelector = flag.StringP("cols", "c", "", "columns to un-flatten (1-based indexing)")
	targetCol   = flag.IntP("target", "t", 0, "target data column (1-based indexing)")
	noHead      = flag.BoolP("no-header", "H", false, "data has header row")
)

const desc = `
unflat can be used to convert data of the form key1,key2,key3,kind,val into a
format like key1,key2,key3,kind1val,kind2val,kind3val (assuming that 3 kinds
exist).

To do this, run unflat -F, -c 4 -t 5.

To select multiple columns, separate the values by commas. You can select
ranges using -. For example, 3-4,6 would select columns 3, 4, and 6.`

func parseSelector(s string) (indiv []int, err error) {
	defer func() {
		if e := recover(); e != nil {
			if ee, ok := e.(error); ok {
				err = ee
			} else {
				panic(e)
			}
		}
	}()

	groups := strings.Split(s, ",")
	for _, g := range groups {
		fs := strings.Split(g, "-")
		if len(fs) == 1 {
			indiv = append(indiv, indSingle(fs[0]))
		} else if len(fs) == 2 {
			indiv = append(indiv, indRange(fs[0], fs[1])...)
		} else {
			panic(fmt.Errorf("too many fields in range: %s", g))
		}
	}

	indiv = indMulti(indiv)
	return
}

func indSingle(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	if v < 1 {
		panic(errors.New("bad column index"))
	}
	return v - 1
}

func indRange(start, end string) []int {
	vs, err := strconv.Atoi(start)
	if err != nil {
		panic(err)
	}
	if vs < 1 {
		panic(errors.New("bad column start"))
	}
	ve, err := strconv.Atoi(end)
	if err != nil {
		panic(err)
	}
	if ve < 1 {
		panic(errors.New("bad column end"))
	}
	var inds []int
	for i := vs - 1; i < ve; i++ {
		inds = append(inds, i)
	}
	return inds
}

func indMulti(inds []int) []int {
	sort.Ints(inds)
	prev := -1
	for i := 0; i < len(inds); {
		if inds[i] == prev {
			inds = append(inds[:i], inds[i+1:]...)
		} else {
			prev = inds[i]
			i++
		}
	}

	return inds
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: unflat -c COL_SELECTOR -t TARGET_COL")
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, desc)
	os.Exit(2)
}

func main() {
	log.SetPrefix("unflat: ")
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	if *colSelector == "" || *targetCol <= 0 {
		usage()
	}

	colSel, err := parseSelector(*colSelector)
	if err != nil {
		log.Fatal(err)
	}

	res, err := slurp(os.Stdin, !*noHead, *fieldDelim, colSel, *targetCol-1)
	if err != nil {
		log.Fatal(err)
	}

	print(os.Stderr, res, colSel, *targetCol-1, *fieldDelim)
}

func print(w io.Writer, res slurpResult, colSet []int, targetCol int, delim string) error {
	if delim == "" {
		delim = "\t"
	}

	var buf bytes.Buffer
	for i, h := range res.Header {
		if i == targetCol || contains(colSet, i) {
			continue
		}
		buf.WriteString(h)
		buf.WriteString(delim)
	}
	for _, k := range res.KindOrder {
		buf.WriteString(k)
		buf.WriteString(delim)
	}
	if _, err := w.Write(bufDone(&buf, len(delim))); err != nil {
		return err
	}

	for _, static := range res.StaticOrder {
		buf.Reset()
		buf.WriteString(static)

		for _, kind := range res.KindOrder {
			buf.WriteString(delim)
			buf.WriteString(res.Vals[staticKind{static, kind}])
		}
		buf.WriteString("\n")

		if _, err := w.Write(buf.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

func bufDone(buf *bytes.Buffer, dlen int) []byte {
	b := buf.Bytes()
	b = b[:len(b)-dlen]
	return append(b, '\n')
}

type staticKind struct {
	Static string
	Kind   string
}

type slurpResult struct {
	Header      []string
	StaticOrder []string
	KindOrder   []string
	Vals        map[staticKind]string
}

func slurp(r io.Reader, hasHead bool, delim string, colSel []int, targetCol int) (slurpResult, error) {
	var (
		staticOrder StringSet
		kindOrder   StringSet
		header      []string
		vals        = make(map[staticKind]string)

		first = true

		// reused every iteration
		static []string
		kind   []string
	)

	writeDelim := delim
	if writeDelim == "" {
		writeDelim = "\t"
	}

	s := bufio.NewScanner(os.Stdin)

	for s.Scan() {
		var fs []string
		if delim == "" {
			fs = strings.Fields(s.Text())
		} else {
			fs = strings.Split(s.Text(), delim)
		}
		if !first && len(header) != len(fs) {
			return slurpResult{}, fmt.Errorf("malformed data: varying number of cols: %d != %d", len(header), len(fs))
		}
		if first {
			first = false
			if hasHead {
				header = fs
				continue
			} else {
				header = make([]string, len(fs))
				for i := range header {
					header[i] = "Col." + strconv.Itoa(i)
				}
			}
		}

		var val maybeString
		static = static[:0]
		kind = kind[:0]

		for i, f := range fs {
			switch {
			case i == targetCol:
				val.Set(f)
			case contains(colSel, i):
				kind = append(kind, f)
			default:
				static = append(static, f)
			}
		}

		if len(colSel) != len(kind) {
			return slurpResult{}, errors.New("selected columns do not exist")
		}

		if !val.OK {
			return slurpResult{}, errors.New("target column does not exist")
		}

		staticKey := strings.Join(static, writeDelim)
		kindKey := strings.Join(kind, ".")

		staticOrder.Add(staticKey)
		kindOrder.Add(kindKey)

		sk := staticKind{staticKey, kindKey}
		if _, ok := vals[sk]; ok {
			return slurpResult{}, fmt.Errorf("duplicate value for %v", sk)
		}
		vals[sk] = val.Val
	}

	res := slurpResult{
		Header:      header,
		StaticOrder: staticOrder.Order(),
		KindOrder:   kindOrder.Order(),
		Vals:        vals,
	}

	return res, nil
}

type maybeString struct {
	Val string
	OK  bool
}

func (ms *maybeString) Set(s string) {
	ms.OK = true
	ms.Val = s
}

func contains(vs []int, v int) bool {
	for _, v2 := range vs {
		if v2 == v {
			return true
		}
	}
	return false
}

type StringSet struct {
	order []string
	has   map[string]bool
}

func (ss *StringSet) Add(s string) {
	if ss.has == nil {
		ss.has = make(map[string]bool)
	}
	if ss.has[s] {
		return
	}
	ss.has[s] = true
	ss.order = append(ss.order, s)
}

func (ss *StringSet) Order() []string {
	return ss.order
}
