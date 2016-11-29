package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/goware/prefixer"

	"9fans.net/go/acme"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("")
	wid, err := strconv.Atoi(os.Getenv("winid"))
	if err != nil {
		log.Fatal("unable to find window")
	}
	win, err := acme.Open(wid, nil)
	if err != nil {
		log.Fatalf("unable to open window: %v", err)
	}
	wis, _ := acme.Windows()
	var name string
	for _, wi := range wis {
		if wi.ID == wid {
			name = wi.Name
			break
		}
	}
	cmd := exec.Command("aspell", "-a")
	cmd.Stdin = prefixer.New(bodyReader{win}, "^")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("unable to create pipe")
	}
	cmd.Start()
	p := printer{name}
	go p.scan(stdout)
	if err := cmd.Wait(); err != nil && err != io.EOF {
		log.Fatalf("error running \"aspell -a\": %v", err)
	}
}

type bodyReader struct{ *acme.Win }

func (r bodyReader) Read(data []byte) (int, error) {
	return r.Win.Read("body", data)
}

type printer struct {
	name string
}

func (p printer) scan(r io.ReadCloser) {
	defer r.Close()
	s := bufio.NewScanner(r)
	lineno := 1
	for s.Scan() {
		l := s.Text()
		switch {
		case strings.HasPrefix(l, "&"):
			fields := strings.Fields(l)
			if len(fields) < 4 {
				log.Fatalf("malformed input: %s", l)
			}
			off, _ := strconv.Atoi(strings.TrimRight(fields[3], ":"))
			word := fields[1]
			top3 := fields[4:]
			if len(top3) > 3 {
				top3 = top3[:4]
				top3[3] = "..."
			}
			log.Printf("%s: %s", getAddr(p.name, lineno, off, len(word)), strings.Join(top3, " "))
		case strings.HasPrefix(l, "#"):
			fields := strings.Fields(l)
			if len(fields) < 3 {
				log.Fatalf("malformed input: %s", l)
			}
			off, _ := strconv.Atoi(strings.TrimRight(fields[2], ":"))
			word := fields[1]
			log.Printf("%s: [no suggestions]", getAddr(p.name, lineno, off, len(word)), off+len(word))
		case l == "":
			lineno++ // EOL
		}
	}
	if s.Err() != nil {
		log.Fatalf("error reading aspell output: %v", s.Err())
	}
}

func getAddr(name string, lineno int, offset int, wordlen int) string {
	if lineno == 1 {
		// hack to word around weird 1:M,1:N issues
		return fmt.Sprintf("%s:#%d,#%d", name, offset-1, offset-1+wordlen)
	}
	return fmt.Sprintf("%s:%d:%d,%d:%d", name, lineno, offset, lineno, offset+wordlen)
}
