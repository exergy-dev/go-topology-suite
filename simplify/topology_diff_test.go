package simplify

import (
	"math"
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/internal/benchfix"
	"github.com/exergy-dev/go-topology-suite/wkt"
	"pgregory.net/rapid"
)

// The oracle in topology_old_test.go is a frozen copy of the pre-index
// brute-force implementation; the indexed version must match it exactly.

func TestTopologyPreservingMatchesBruteForce(t *testing.T) {
	for _, n := range []int{16, 64, 300, 1024, 3000} {
		g := benchfix.Star(n, 1000, 600)
		for _, tol := range []float64{0.5, 5, 20, 100} {
			a, b := TopologyPreserving(g, tol), TopologyPreservingOld(g, tol)
			if ws(a) != ws(b) {
				t.Fatalf("star n=%d tol=%v differ", n, tol)
			}
		}
	}
	grid := benchfix.Grid(6, 6, 0.1)
	for _, tol := range []float64{0.01, 0.1, 0.5} {
		a, b := TopologyPreserving(grid, tol), TopologyPreservingOld(grid, tol)
		if ws(a) != ws(b) {
			t.Fatalf("grid tol=%v differ", tol)
		}
	}
}

func TestTopologyPreservingMatchesBruteForceRandom(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		nl := rapid.IntRange(1, 4).Draw(t, "lines")
		var lines []*geom.LineString
		for i := 0; i < nl; i++ {
			n := rapid.IntRange(2, 40).Draw(t, "n")
			pts := make([]geom.XY, n)
			for k := range pts {
				pts[k] = geom.XY{X: float64(rapid.IntRange(0, 60).Draw(t, "x")), Y: float64(rapid.IntRange(0, 60).Draw(t, "y"))}
			}
			lines = append(lines, geom.NewLineString(nil, pts))
		}
		g := geom.NewMultiLineString(nil, lines...)
		tol := rapid.Float64Range(0.1, 30).Draw(t, "tol")
		a, b := TopologyPreserving(g, tol), TopologyPreservingOld(g, tol)
		if ws(a) != ws(b) {
			t.Fatalf("differ tol=%v\nnew %s\nold %s", tol, ws(a), ws(b))
		}
	})
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(4, 60).Draw(t, "n")
		pts := make([]geom.XY, n+1)
		for k := 0; k < n; k++ {
			ang := 2 * math.Pi * float64(k) / float64(n)
			r := float64(rapid.IntRange(20, 100).Draw(t, "r"))
			pts[k] = geom.XY{X: r * math.Cos(ang), Y: r * math.Sin(ang)}
		}
		pts[n] = pts[0]
		g := geom.NewPolygon(nil, pts)
		tol := rapid.Float64Range(0.1, 40).Draw(t, "tol")
		a, b := TopologyPreserving(g, tol), TopologyPreservingOld(g, tol)
		if ws(a) != ws(b) {
			t.Fatalf("polygon differ tol=%v", tol)
		}
	})
}

func ws(g geom.Geometry) string {
	s, err := wkt.Marshal(g)
	if err != nil {
		return err.Error()
	}
	return s
}

// TestTopologyPreservingMatchesBruteForceIndexed forces the indexed
// path on small inputs by dropping the size gate to zero.
func TestTopologyPreservingMatchesBruteForceIndexed(t *testing.T) {
	old := indexThreshold
	indexThreshold = 0
	defer func() { indexThreshold = old }()
	TestTopologyPreservingMatchesBruteForce(t)
	TestTopologyPreservingMatchesBruteForceRandom(t)
}
