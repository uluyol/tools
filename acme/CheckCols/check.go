package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"

	"9fans.net/go/acme"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("CheckCols: ")

	maxCols := 80

	if len(os.Args) > 1 {
		var err error
		maxCols, err = strconv.Atoi(os.Args[1])
		if err != nil {
			log.Fatalf("invalid number of columns: %v", err)
		}
	}
	wid, err := strconv.Atoi(os.Getenv("winid"))
	if err != nil {
		log.Fatal("unable to find window")
	}
	win, err := acme.Open(wid, nil)
	if err != nil {
		log.Fatalf("unable to open window: %v", err)
	}
	wis, _ := acme.Windows()
	var name string
	for _, wi := range wis {
		if wi.ID == wid {
			name = wi.Name
			break
		}
	}
	s := bufio.NewScanner(bodyReader{win})
	lineno := 1
	for s.Scan() {
		if len(s.Text()) > maxCols {
			fmt.Printf("%s:%d: have %d cols, want %d\n", name, lineno, len(s.Text()), maxCols)
		}
		lineno++
	}

	if s.Err() != nil {
		log.Fatalf("error scanning text: %v", s.Err())
	}
}

type bodyReader struct{ *acme.Win }

func (r bodyReader) Read(data []byte) (int, error) {
	return r.Win.Read("body", data)
}
