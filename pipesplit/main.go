package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"hash"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/dustin/go-humanize"
)

var (
	verbose      *bool   = flag.Bool("v", false, "Verbose output")
	checksumKind *string = flag.String("checksum", "sha256", "Checksum to create, valid values are none, sha256")
	chunksizeRaw *string = flag.String("size", "512MB", "Chunksize")
	chunksize    uint64
)

const (
	bufSize     = 1 << 14
	usageString = `Usage: %s template
where template is the command to run and the final argument is the destination
filename. The chunk number will be appended to the filename so that file1.ext
becomes file1.ext.0, file1.ext.1, etc. Similarly if checksumming is enabled, you
will have file1.ext.0.sha256, file1.ext.1.sha256, and so on.

Besides splitting, files are not modified in any way. Use 'cat' to recombine
them.
`
)

func usage() {
	fmt.Fprintf(os.Stderr, usageString, os.Args[0])
	flag.PrintDefaults()
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

type NullHash struct{}

func (NullHash) Write(b []byte) (int, error) { return len(b), nil }
func (NullHash) Sum([]byte) []byte           { return nil }
func (NullHash) Reset()                      {}
func (NullHash) Size() int                   { return 0 }
func (NullHash) BlockSize() int              { return 0 }

type SimpleCommand []string

func (sc SimpleCommand) String() string {
	return strings.Join(sc, " ")
}

func main() {
	flag.Usage = usage
	flag.Parse()
	runcmd := flag.Args()
	if len(runcmd) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	uchunksize, err := humanize.ParseBytes(*chunksizeRaw)
	chunksize := int64(uchunksize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid chunksize %s: %v\n", *chunksizeRaw, err)
		os.Exit(2)
	}
	if chunksize < 0 {
		fmt.Fprintf(os.Stderr, "got negative filesize %d\n", chunksize)
	}

	currentChunk := 0
	lastPiece := runcmd[len(runcmd)-1]
	var (
		buf        [bufSize]byte
		hashBytes  []byte
		hash       hash.Hash
		cmdOutBuf  bytes.Buffer
		hashOutBuf bytes.Buffer
		noMoreData bool
	)
	switch *checksumKind {
	case "none":
		hash = NullHash{}
	case "sha256":
		hash = sha256.New()
	default:
		fmt.Fprintf(os.Stderr, "invalid checksum %s\n", *checksumKind)
		os.Exit(2)
	}
	for !noMoreData {
		runcmd[len(runcmd)-1] = lastPiece + "." + fmt.Sprint(currentChunk)
		cmd := exec.Command(runcmd[0], runcmd[1:]...)
		siw, err := cmd.StdinPipe()
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to open pipe: %v\n", err)
			os.Exit(1)
		}
		if *verbose {
			cmd.Stdout = &cmdOutBuf
			cmd.Stderr = &cmdOutBuf
		}
		if err := cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "unable to execute %s: %v\n", SimpleCommand(runcmd), err)
			os.Exit(1)
		}
		w := io.MultiWriter(siw, hash)
		for c := int64(0); c < chunksize && err == nil; {
			n := int(min(bufSize, chunksize-c))
			n, err = os.Stdin.Read(buf[:n])
			if err != nil && err != io.EOF {
				fmt.Fprintf(os.Stderr, "unrecoverable error on read: %v\n", err)
				os.Exit(1)
			}
			if err == io.EOF {
				if c == 0 && n == 0 {
					os.Exit(0)
				}
				noMoreData = true
			}
			if _, err := w.Write(buf[:n]); err != nil {
				fmt.Fprintf(os.Stderr, "unrecoverable error on write: %v\n", err)
				os.Exit(1)
			}
			c += int64(n)
		}
		siw.Close()
		if err := cmd.Wait(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to correctly execute '%s'\n", SimpleCommand(runcmd))
		}
		if *verbose {
			fmt.Printf("%s:\n", SimpleCommand(runcmd))
			io.Copy(os.Stdout, &cmdOutBuf)
		}
		if _, ok := hash.(NullHash); !ok {
			hashBytes = hash.Sum(hashBytes[:0])
			runcmd[len(runcmd)-1] += "." + *checksumKind
			sumCmd := exec.Command(runcmd[0], runcmd[1:]...)
			siw, err := sumCmd.StdinPipe()
			if err != nil {
				fmt.Fprintf(os.Stderr, "unable to open pipe: %v\n", err)
				os.Exit(1)
			}
			if *verbose {
				sumCmd.Stdout = &hashOutBuf
				sumCmd.Stderr = &hashOutBuf
			}
			if err := sumCmd.Start(); err != nil {
				fmt.Fprintln(os.Stderr, "unable to execute command:", SimpleCommand(runcmd))
				os.Exit(1)
			}
			if _, err := siw.Write([]byte(hex.EncodeToString(hashBytes))); err != nil {
				fmt.Fprintf(os.Stderr, "unrecoverable error on write: %v\n", err)
				os.Exit(1)
			}
			siw.Close()
			if err := sumCmd.Wait(); err != nil {
				fmt.Fprintf(os.Stderr, "failed to correctly execute '%s'\n", SimpleCommand(runcmd))
				os.Exit(1)
			}
			if *verbose {
				fmt.Printf("%s:\n", SimpleCommand(runcmd))
				io.Copy(os.Stdout, &hashOutBuf)
			}
		}
		hash.Reset()
		cmdOutBuf.Reset()
		hashOutBuf.Reset()
		currentChunk += 1
	}
}
