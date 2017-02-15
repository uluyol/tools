/*

Timeleft prints the amount of time left until some date.
It can be used as a countdown timer for deadlines.

Usage:

	timeleft datestamp

where datestamp is of the time.UnixDate form
as seen below:

	Mon Jan 2 15:04:05 MST 2006

*/
package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: timeleft event datestamp\n")
	os.Exit(2)
}

func main() {
	log.SetPrefix("timeleft: ")
	log.SetFlags(0)
	if len(os.Args) < 3 {
		usage()
	}
	event := os.Args[1]
	datestamp := strings.Join(os.Args[2:], " ")
	t, err := time.Parse(time.UnixDate, datestamp)
	if err != nil {
		log.Fatal("datestamp must be of the form: Mon Jan 2 15:04:05 MST 2006")
	}
	d := t.Sub(time.Now())
	if d < 0 {
		fmt.Println(event, "already happened")
		return
	}
	if d/time.Minute == 0 {
		fmt.Println(event, "is now")
		return
	}
	days := int(d/time.Hour) / 24
	weeks := days / 7
	hours := int(d/time.Hour) - 24*days
	days = days - 7*weeks
	min := int(d/time.Minute) - 60*int(d/time.Hour)

	var buf bytes.Buffer
	buf.WriteString(event)
	buf.WriteString(" is in ")
	var fieldCount int
	printIfPresent(&buf, &fieldCount, "weeks", weeks)
	printIfPresent(&buf, &fieldCount, "days", days)
	printIfPresent(&buf, &fieldCount, "hours", hours)
	printIfPresent(&buf, &fieldCount, "min", min)
	buf.Truncate(buf.Len() - 1)
	buf.WriteByte('\n')

	io.Copy(os.Stdout, &buf)
}

func printIfPresent(buf *bytes.Buffer, numFields *int, name string, val int) {
	if val != 0 && *numFields < 2 {
		buf.WriteString(strconv.Itoa(val))
		buf.WriteByte(' ')
		buf.WriteString(name)
		buf.WriteByte(' ')
		*numFields++
	}
}
