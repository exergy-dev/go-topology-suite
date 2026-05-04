// Command gts-bench runs the bench package's workloads via testing.Benchmark
// and prints a single comparison table (workload | ns/op | B/op | allocs/op).
//
// Usage:
//
//	go run ./bench/cmd/gts-bench
//
// Benchmark sizing is governed by the constants in bench/fixtures.go; see
// bench/doc.go for the scaling factor and rationale.
package main

import (
	"fmt"
	"os"
	"testing"
	"text/tabwriter"

	"github.com/exergy-dev/go-topology-suite/bench"
)

func main() {
	workloads := bench.Workloads()
	results := make([]struct {
		Name string
		R    testing.BenchmarkResult
	}, 0, len(workloads))

	for _, w := range workloads {
		fmt.Fprintf(os.Stderr, "running %s ...\n", w.Name)
		r := testing.Benchmark(func(b *testing.B) {
			b.ReportAllocs()
			w.Fn(b)
		})
		results = append(results, struct {
			Name string
			R    testing.BenchmarkResult
		}{w.Name, r})
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "workload\tns/op\tB/op\tallocs/op")
	fmt.Fprintln(tw, "--------\t-----\t----\t---------")
	for _, r := range results {
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\n",
			r.Name,
			r.R.NsPerOp(),
			r.R.AllocedBytesPerOp(),
			r.R.AllocsPerOp(),
		)
	}
	if err := tw.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "flush: %v\n", err)
		os.Exit(1)
	}
}
