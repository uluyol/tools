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
	"flag"
	"os"
)

var singleByte *bool = flag.Bool("u",
                                   false,
                                   "Write bytes from the input file to stout without delay")

func main() {
	flag.Parse()
	fNames := flag.Args()
	if len(fNames) < 1 {
		cat(os.Stdin, *singleByte)
	} else {
		for _, fName := range fNames {
			file, err := os.Open(fName)
			ckError(err)
			cat(file, *singleByte)
		}
	}
}