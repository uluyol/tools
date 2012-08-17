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