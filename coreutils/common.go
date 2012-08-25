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
	"bufio"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
)

func ckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}

func ckSums(in io.Reader, new func() hash.Hash) {
	var (
		err error
		fails int
	)
	fData := make([]byte, 1024)
	isPrefix := false
	bufIn := bufio.NewReader(in)
	for {
		fData, isPrefix, err = bufIn.ReadLine()
		if err == io.EOF {
			break
		} else { ckError(err) }
		if isPrefix {
			fmt.Fprintf(os.Stderr, "Buffer not large enough\n")
			os.Exit(1)
		}
		hashLen := 0
		for i, v := range fData {
			if v == byte(' ') { hashLen = i; break }
		}
		hash := string(fData[:hashLen])
		fName := string(fData[hashLen+2:])
		file, err := os.Open(fName)
		ckError(err)
		if ckSum(file, new) == hash {
			fmt.Printf("%s: OK\n", fName)
		} else {
			fmt.Printf("%s: FAILED\n", fName)
			fails++
		}
	}
	if fails > 0 {
		grammarFix := "checksums"
		if fails == 1 {
			grammarFix = "checksum"
		}
		fmt.Printf("WARNING: %d computed %s did NOT match\n", fails, grammarFix)
		os.Exit(1)
	}
	os.Exit(0)
}

func ckSum(in io.Reader, new func() hash.Hash) string {
	data, err := ioutil.ReadAll(in)
	ckError(err)
	hash := new()
	hash.Write(data)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func cat(in *os.File, singleByte bool) {
	if singleByte {
		buf := make([]byte, 1)
		for {
			n, _ := in.Read(buf)
			if n != 1 { break }
			_, err := os.Stdout.Write(buf)
			ckError(err)
		}
	} else {
		buf, err := ioutil.ReadAll(in)
		ckError(err)
		_, err = os.Stdout.Write(buf)
		ckError(err)
	}
}