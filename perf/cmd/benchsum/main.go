package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/gonum/stat"
)

var noPercent = flag.Bool("nopercent", false, "don't use percentages, show the true value")

type benchResults struct {
	vals  []float64
	niter []float64
}

var benchVals = make(map[string]map[string]benchResults)

type statPair struct {
	stat    string
	results benchResults
}

type byStat []statPair

func (s byStat) Len() int           { return len(s) }
func (s byStat) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byStat) Less(i, j int) bool { return s[i].stat < s[j].stat }

type benchValsPair struct {
	name  string
	stats []statPair
}

type byName []benchValsPair

func (s byName) Len() int           { return len(s) }
func (s byName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byName) Less(i, j int) bool { return s[i].name < s[j].name }

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
	var allVals []benchValsPair
	for name, statmap := range benchVals {
		var stats []statPair
		for st, results := range statmap {
			stats = append(stats, statPair{stat: st, results: results})
		}
		sort.Sort(byStat(stats))
		allVals = append(allVals, benchValsPair{name: name, stats: stats})
	}
	sort.Sort(byName(allVals))
	for _, p := range allVals {
		name := p.name
		for _, p2 := range p.stats {
			st := p2.stat
			results := p2.results
			μ, σ := stat.MeanStdDev(results.vals, results.niter)
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

func addVal(key, stat string, val, niter float64) {
	if benchVals[key] == nil {
		benchVals[key] = make(map[string]benchResults)
	}
	results := benchVals[key][stat]
	results.vals = append(results.vals, val)
	results.niter = append(results.niter, niter)
	benchVals[key][stat] = results
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
			addVal(name, f[i+1], v, float64(niter))
		}
	}
}
