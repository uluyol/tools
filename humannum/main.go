package main

import (
	"flag"
	"fmt"
	"os"
)

func usage() {
	fmt.Fprintln(os.Stderr, "usage: humannum number")
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		usage()
		os.Exit(1)
	}
	num := flag.Arg(0)
	var nz int
	for i := len(num) - 1; i >= 0; i-- {
		if num[i] == '0' {
			nz++
		} else {
			break
		}
	}
	var pretty string
	switch {
	case 3 <= nz && nz < 6:
		pretty = num[:len(num)-3] + "K"
	case 6 <= nz && nz < 9:
		pretty = num[:len(num)-6] + "M"
	case 9 <= nz && nz < 12:
		pretty = num[:len(num)-9] + "G"
	case 12 <= nz:
		pretty = num[:len(num)-12] + "T"
	default:
		pretty = num
	}
	fmt.Println(pretty)
}
