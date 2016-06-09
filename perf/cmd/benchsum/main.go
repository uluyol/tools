package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/gonum/stat"
)

var noPercent = flag.Bool("nopercent", false, "don't use percentages, show the true value")

var benchVals = make(map[string]map[string][]float64)

func main() {
	log.SetPrefix("benchsum: ")
	log.SetFlags(0)
	flag.Parse()
	if flag.NArg() == 0 {
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		read(b)
	} else {
		for _, p := range flag.Args() {
			b, err := ioutil.ReadFile(p)
			if err != nil {
				log.Fatal(err)
			}
			read(b)
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 8, 8, 1, ' ', 0)
	fmt.Fprint(w, "name\tmean\tstdev\tstat ...\n")
	for name, statmap := range benchVals {
		for st, vals := range statmap {
			μ, σ := stat.MeanStdDev(vals, nil)
			if *noPercent {
				fmt.Fprintf(w, "%s\t%.2e\t±%.2e\t%s\n", name, μ, σ, st)
			} else {
				σPercent := 100 * σ / μ
				fmt.Fprintf(w, "%s\t%.2e\t±%.0f%%\t%s\n", name, μ, σPercent, st)
			}
		}
	}
	w.Flush()
}

func addVal(key, stat string, val float64) {
	if benchVals[key] == nil {
		benchVals[key] = make(map[string][]float64)
	}
	benchVals[key][stat] = append(benchVals[key][stat], val)
}

func read(data []byte) {
	for _, line := range strings.Split(string(data), "\n") {
		f := strings.Fields(line)
		if len(f) < 4 {
			continue
		}
		if !strings.HasPrefix(f[0], "Benchmark") {
			continue
		}
		name := strings.TrimPrefix(f[0], "Benchmark")
		niter, err := strconv.Atoi(f[1])
		if err != nil {
			continue
		}
		for i := 2; i+2 <= len(f); i += 2 {
			v, err := strconv.ParseFloat(f[i], 64)
			if err != nil {
				continue
			}
			addVal(name, f[i+1], v/float64(niter))
		}
	}
}
