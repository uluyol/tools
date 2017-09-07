package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"9fans.net/go/acme"
)

const doc = `
Fmt is a source code formatting harness for Acme.
It is intended to replace Edit ,|myformatter for goimports and other formatters.
Fmt must be used from within an Acme buffer or its tag.
It takes a single argument: the formatting command to run over the buffer contents.
Fmt provides two benefits over Edit ,|myformatter:
1) After formatting it doesn't leave you looking at the top of the buffer,
but tries to show you where you were when you clicked Fmt.
2) If the formatter returns in error the buffer contents are left unchanged.
`

type bodyReader struct{ *acme.Win }

func (r bodyReader) Read(data []byte) (int, error) {
	return r.Win.Read("body", data)
}

type countReader struct {
	count int
	r     io.Reader
}

func (r *countReader) Read(data []byte) (int, error) {
	n, err := r.r.Read(data)
	r.count += n
	return n, err
}

type countWriter struct {
	count int
	w     io.Writer
}

func (w *countWriter) Write(data []byte) (int, error) {
	n, err := w.w.Write(data)
	w.count += n
	return n, err
}

type dataWriter struct{ *acme.Win }

func (w dataWriter) Write(data []byte) (int, error) {
	return w.Win.Write("data", data)
}

func usage() {
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, doc)
}

type formatter struct {
	cmd    []string
	stderr io.WriteCloser
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

type textReplacer struct {
	text, replacement []byte
	buf               bytes.Buffer
	w                 io.Writer
}

func (r *textReplacer) Write(b []byte) (int, error) {
	return r.buf.Write(b)
}

func (r *textReplacer) Close() error {
	b := r.buf.Bytes()
	for i := bytes.Index(b, r.text); i >= 0; i = bytes.Index(b, r.text) {
		if _, err := r.w.Write(b[:i]); err != nil {
			return fmt.Errorf("error flushing buffer: %v", err)
		}
		if _, err := r.w.Write(r.replacement); err != nil {
			return fmt.Errorf("error writing replacement: %v", err)
		}
		b = b[len(r.text):]
	}
	if len(b) > 0 {
		if _, err := r.w.Write(b); err != nil {
			return fmt.Errorf("error flushing remainder of buffer: %v", err)
		}
	}
	return nil
}

func replacer(w io.Writer, text, replacement string) io.WriteCloser {
	return &textReplacer{
		text:        []byte(text),
		replacement: []byte(replacement),
		w:           w,
	}
}

func main() {
	flag.Usage = usage
	flag.Parse()
	var cfmt formatter
	if flag.NArg() == 0 {
		p := os.Getenv("samfile")
		switch {
		case strings.HasSuffix(p, ".go"):
			cfmt = formatter{
				cmd:    []string{"goimports"},
				stderr: nopCloser{os.Stderr},
			}
		case suffixOneOf(p, ".c", ".cc", ".cpp", ".cxx", ".h", ".hpp"):
			cfmt = formatter{
				cmd:    []string{"clang-format"},
				stderr: nopCloser{os.Stderr},
			}
		case strings.HasSuffix(p, ".java"):
			cfmt = formatter{
				cmd:    []string{"google-java-format", "-"},
				stderr: replacer(os.Stderr, "<stdin>", filepath.Base(p)),
			}
		case strings.HasSuffix(p, ".scala"):
			cfmt = formatter{
				cmd:    []string{"scalafmt", "--stdin"},
				stderr: nopCloser{os.Stderr},
			}
		case strings.HasSuffix(p, ".json"):
			cfmt = formatter{
				cmd:    []string{"jq", "-M", "."},
				stderr: nopCloser{os.Stderr},
			}
		case suffixOneOf(p, ".sh", ".bash"):
			cfmt = formatter{
				cmd:    []string{"shfmt"},
				stderr: nopCloser{os.Stderr},
			}
		default:
			fmt.Fprintf(os.Stderr, "no default formatter for %s", p)
			os.Exit(3)
		}
	} else {
		cfmt = formatter{
			cmd:    flag.Args(),
			stderr: nopCloser{os.Stderr},
		}
	}
	win, err := openWin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open win: %s\n", err)
		os.Exit(1)
	}
	q0, q1, err := readAddr(win)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get the current selection: %s\n", err)
		os.Exit(1)
	}
	status := 0
	ffile, sameSize, err := format(win, cfmt)
	diff := !sameSize
	if err != nil {
		fmt.Fprintf(os.Stderr, "format failed: %s\n", err)
		status = 1
		goto out
	}
	if !diff {
		diff, err = bodyDiff(win, ffile)
		if err != nil {
			// Not fatal. Re-write the body anyway.
			fmt.Fprintf(os.Stderr, "failed to diff the body: ", err)
			diff = true
		}
	}
	if diff {
		if err := writeBody(win, ffile); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write the body: %s\n", err)
			status = 1
			goto out
		}
		if err := showAddr(win, q0, q1); err != nil {
			fmt.Fprintf(os.Stderr, "failed to restore the selection: %s\n", err)
			status = 1
			goto out
		}
	}

out:
	if err := os.Remove(ffile); err != nil {
		fmt.Fprintf(os.Stderr, "failed to remove tempfile %s: %s\n", ffile, err)
	}
	os.Exit(status)
}

func openWin() (*acme.Win, error) {
	id, err := strconv.Atoi(os.Getenv("winid"))
	if err != nil {
		return nil, err
	}
	return acme.Open(id, nil)
}

func readAddr(win *acme.Win) (q0, q1 int, err error) {
	// This first read is bogus.
	// Acme zeroes the win's address the first time addr is opened.
	// So, we need to open it before setting addr=dot,
	// lest we just read back a zero address.
	if _, _, err := win.ReadAddr(); err != nil {
		return 0, 0, err
	}
	if err := win.Ctl("addr=dot\n"); err != nil {
		return 0, 0, err
	}
	return win.ReadAddr()
}

func showAddr(win *acme.Win, q0, q1 int) error {
	if err := win.Addr("#%d,#%d", q0, q1); err != nil {
		return err
	}
	return win.Ctl("dot=addr\nshow\n")
}

// If tmpFile is non-empty, it is created and must be removed by the caller.
func format(win *acme.Win, formatter formatter) (tmpFile string, sameSize bool, err error) {
	defer formatter.stderr.Close()
	tf, err := ioutil.TempFile(os.TempDir(), "Fmt")
	if err != nil {
		return "", false, err
	}
	tmpFile = tf.Name()
	br := &countReader{0, bodyReader{win}}
	fw := &countWriter{0, tf}
	cmd := exec.Command(formatter.cmd[0], formatter.cmd[1:]...)
	cmd.Stdin = br
	cmd.Stdout = fw
	cmd.Stderr = formatter.stderr
	if err = cmd.Run(); err != nil {
		tf.Close()
	} else {
		err = tf.Close()
	}
	sameSize = fw.count == br.count
	return
}

func writeBody(win *acme.Win, ffile string) error {
	if err := win.Ctl("nomark"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set nomark: %s", err)
	}
	defer func() {
		if err := win.Ctl("mark"); err != nil {
			fmt.Fprintf(os.Stderr, "failed to set mark: %s", err)
		}
	}()
	tf, err := os.Open(ffile)
	if err != nil {
		return err
	}
	defer tf.Close()
	if err := win.Addr("0,$"); err != nil {
		return err
	}
	_, err = io.Copy(dataWriter{win}, tf)
	return err
}

func bodyDiff(win *acme.Win, ffile string) (bool, error) {
	tf, err := os.Open(ffile)
	if err != nil {
		return false, err
	}
	defer tf.Close()
	win.Seek("body", 0, 0)
	fr := bufio.NewReader(tf)
	br := bufio.NewReader(&bodyReader{win})
	for {
		fb, errf := fr.ReadByte()
		if errf != nil && errf != io.EOF {
			return false, errf
		}
		bb, errb := br.ReadByte()
		if errb != nil && errb != io.EOF {
			return false, errb
		}
		if fb != bb {
			return true, nil
		}
		if errf == io.EOF && errb == io.EOF {
			return false, nil
		}
	}
}

func suffixOneOf(s string, suffixes ...string) bool {
	for _, suf := range suffixes {
		if strings.HasSuffix(s, suf) {
			return true
		}
	}
	return false
}
