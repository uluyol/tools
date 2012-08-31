/*
 *  Copyright (c) 2012, Muhammed Uluyol <uluyol0@gmail.com>
 *
 *  Permission to use, copy, modify, and/or distribute this software for any
 *  purpose with or without fee is hereby granted, provided that the above
 *  copyright notice and this permission notice appear in all copies.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 *  WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 *  MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY
 *  SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 *  WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION
 *  OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF OR IN
 *  CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
)

var doCheck *bool = flag.Bool("c",
	false,
	"read SHA224 sums from the FILEs and check them")

func main() {
	var (
		file *os.File
		err  error
	)
	flag.Parse()
	fName := flag.Arg(0)
	if fName == "" || fName == "-" {
		file = os.Stdin
		fName = "-"
	} else {
		file, err = os.Open(fName)
		ckError(err)
	}
	if *doCheck {
		ckSums(file, sha256.New224)
	}
	fmt.Printf("%s  %s\n", ckSum(file, sha256.New224), fName)
}
