package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

func getPercentile(vals []float64, pct float64) float64 {
	if len(vals) == 0 {
		return math.NaN()
	}
	k := float64(len(vals)-1) * pct
	f := math.Floor(k)
	c := math.Ceil(k)
	if f == c {
		return vals[int(k)]
	}
	d0 := vals[int(f)] * (c - k)
	d1 := vals[int(c)] * (k - f)
	return d0 + d1
}

func main() {
	log.SetPrefix("qstats: ")
	log.SetFlags(0)
	var all []float64
	min := math.Inf(1)
	max := math.Inf(-1)
	prod := float64(1)
	var count, total float64

	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		count++
		v, err := strconv.ParseFloat(strings.TrimSpace(s.Text()), 64)
		if err != nil {
			log.Fatalf("stdin:%d: bad value: %v", err)
		}
		total += v
		prod *= v
		all = append(all, v)
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}

	sort.Float64s(all)
	mean := total / count
	var stddev float64
	for _, v := range all {
		x := v - mean
		stddev += x * x
	}
	stddev /= float64(len(all)) - 1
	stddev = math.Sqrt(stddev)

	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	emit := func(name string, val float64) {
		fmt.Fprintf(tw, "%s\t%v\n", name, val)
	}

	emit("sum", total)
	emit("min", min)
	emit("pct25", getPercentile(all, 0.25))
	emit("pct50", getPercentile(all, 0.5))
	emit("pct75", getPercentile(all, 0.75))
	emit("pct90", getPercentile(all, 0.9))
	emit("pct95", getPercentile(all, 0.95))
	emit("pct99", getPercentile(all, 0.99))
	emit("max", max)
	emit("geomean", math.Pow(prod, 1/count))
	emit("mean", mean)
	emit("stddev", stddev)
	emit("count", count)

	tw.Flush()
}
