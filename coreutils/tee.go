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
	"os/signal"
)

var doAppend *bool = flag.Bool("a",
	false,
	"Append the output to the files.")

var ignoreInt *bool = flag.Bool("i",
	false,
	"Ignore the SIGINT signal.")

func main() {
	var outs []*os.File
	mode := os.O_WRONLY | os.O_CREATE
	flag.Parse()
	paths := flag.Args()
	if *doAppend {
		mode = os.O_WRONLY | os.O_APPEND | os.O_CREATE
	}
	if *ignoreInt {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
	}
	for _, path := range paths {
		out, err := os.OpenFile(path, mode, 0666)
		ckError(err)
		outs = append(outs, out)
	}
	buf := make([]byte, 4096)
	for {
		n, err := os.Stdin.Read(buf)
		ckError(err)
		_, err = os.Stdout.Write(buf[:n])
		ckError(err)
		for _, outFile := range outs {
			_, err = outFile.Write(buf[:n])
			ckError(err)
		}
		if n < 4096 {
			break
		}
	}
}
