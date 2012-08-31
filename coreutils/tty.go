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
	"fmt"
	"os"
)

// #include <unistd.h>
import "C"

func main() {
	ttyFd := C.int(os.Stdout.Fd())
	ttyname := C.GoString(C.ttyname(ttyFd))
	if len(ttyname) < 1 {
		ttyname = "not a tty"
	}
	fmt.Printf("%s\n", ttyname)
	if int(C.isatty(ttyFd)) != 1 {
		os.Exit(1)
	}
}