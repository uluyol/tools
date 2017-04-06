package main

import (
	"io"
	"os"
	"text/tabwriter"
)

func main() {
	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	io.Copy(tw, os.Stdin)
	tw.Flush()
}
