package relateng

import (
	"fmt"
	"math"
	"testing"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/wkt"
)

// countingProcessor implements SegmentPairProcessor by counting
// chain-pair dispatches. Used to observe the require-self-noding
// guard's effect on EdgeSetIntersector.Process.
type countingProcessor struct{ count int }

func (c *countingProcessor) ProcessIntersections(_ *RelateSegmentString, _ int, _ *RelateSegmentString, _ int) {
	c.count++
}
func (c *countingProcessor) IsDone() bool { return false }

// TestEdgeSetIntersector_RequireSelfNodingGuard verifies that the
// require-self-noding flag controls whether intra-input (A-vs-A,
// B-vs-B) chain pairs are dispatched. With four single-segment
// edges (two on each side, all in the same envelope box), there are
// 6 unordered pairs total: 1 A/A, 1 B/B, 4 A/B. The guard should
// elide the two same-side pairs.
func TestEdgeSetIntersector_RequireSelfNodingGuard(t *testing.T) {
	mk := func(isA bool, pts ...geom.XY) *RelateSegmentString {
		return NewRelateLineString(pts, isA, 0)
	}
	// Four edges all sharing the same envelope box so the R-tree
	// returns every pair as a candidate.
	edgesA := []*RelateSegmentString{
		mk(true, geom.XY{X: 0, Y: 0}, geom.XY{X: 10, Y: 10}),
		mk(true, geom.XY{X: 0, Y: 10}, geom.XY{X: 10, Y: 0}),
	}
	edgesB := []*RelateSegmentString{
		mk(false, geom.XY{X: 1, Y: 1}, geom.XY{X: 9, Y: 9}),
		mk(false, geom.XY{X: 1, Y: 9}, geom.XY{X: 9, Y: 1}),
	}
	env := geom.Envelope{MinX: -100, MinY: -100, MaxX: 100, MaxY: 100}

	cases := []struct {
		name              string
		requireSelfNoding bool
		wantPairs         int
	}{
		{"guard off (no self-noding)", false, 4},
		{"guard on (self-noding)", true, 6},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			es := NewEdgeSetIntersector(edgesA, edgesB, env)
			counter := &countingProcessor{}
			es.Process(counter, tc.requireSelfNoding)
			if counter.count != tc.wantPairs {
				t.Errorf("dispatched pairs = %d, want %d", counter.count, tc.wantPairs)
			}
		})
	}
}

// TestEdgeSetIntersector_GuardDoesNotChangeAnswer verifies that for
// inputs that don't require self-noding (Intersects / Disjoint), the
// guarded fast path still produces the correct answer. We compare
// against the matrix predicate, which always self-nodes.
func TestEdgeSetIntersector_GuardDoesNotChangeAnswer(t *testing.T) {
	cases := []struct {
		name string
		a, b string
	}{
		{"two crossing lines",
			"LINESTRING (0 0, 10 10)",
			"LINESTRING (0 10, 10 0)"},
		{"polygon edge crosses line",
			"POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			"LINESTRING (-1 5, 11 5)"},
		{"adjacent polygons",
			"POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			"POLYGON ((10 0, 20 0, 20 10, 10 10, 10 0))"},
		{"overlapping polygons",
			"POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
			"POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a, err := wkt.Unmarshal(c.a)
			if err != nil {
				t.Fatal(err)
			}
			b, err := wkt.Unmarshal(c.b)
			if err != nil {
				t.Fatal(err)
			}
			rng := NewRelateNG(a, OGCSFSBoundaryRule)
			matrix := rng.EvaluateMatrix(b)
			intersects := rng.Evaluate(b, NewIntersectsPredicate())
			disjoint := rng.Evaluate(b, NewDisjointPredicate())
			// Sanity: the matrix-derived intersection answer must
			// agree with the bool-predicate fast path that uses the
			// require-self-noding=false guard.
			matrixIntersects := matrix.Matches("T********") ||
				matrix.Matches("*T*******") ||
				matrix.Matches("***T*****") ||
				matrix.Matches("****T****")
			if intersects != matrixIntersects {
				t.Errorf("Intersects=%v, but matrix=%s implies %v",
					intersects, matrix.String(), matrixIntersects)
			}
			if disjoint == intersects {
				t.Errorf("Disjoint=%v, Intersects=%v: should be opposite", disjoint, intersects)
			}
		})
	}
}

// BenchmarkEdgeSetIntersector_GuardSavesWork measures the chain-pair
// reduction when require-self-noding is false. The fixture is a
// many-segment ring against a many-segment line in the same box; with
// the guard off only A-vs-B pairs run, with it on the A-vs-A pairs
// dominate and the benchmark should be visibly slower.
func BenchmarkEdgeSetIntersector_GuardSavesWork(b *testing.B) {
	const n = 64
	// Build a closed n-gon as A and a horizontal scan-line as B.
	ring := make([]geom.XY, 0, n+1)
	for i := 0; i < n; i++ {
		theta := 2 * math.Pi * float64(i) / float64(n)
		ring = append(ring, geom.XY{X: 1000 + 100*math.Cos(theta), Y: 1000 + 100*math.Sin(theta)})
	}
	ring = append(ring, ring[0])
	a := NewRelateRing(ring, true, 0, 0, nil)

	line := make([]geom.XY, 0, n+1)
	for i := 0; i <= n; i++ {
		line = append(line, geom.XY{X: 800 + 4*float64(i), Y: 1000})
	}
	bb := NewRelateLineString(line, false, 0)

	env := geom.Envelope{MinX: 0, MinY: 0, MaxX: 2000, MaxY: 2000}

	for _, requireSelfNoding := range []bool{false, true} {
		name := fmt.Sprintf("requireSelfNoding=%v", requireSelfNoding)
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				es := NewEdgeSetIntersector(
					[]*RelateSegmentString{a},
					[]*RelateSegmentString{bb},
					env,
				)
				counter := &countingProcessor{}
				es.Process(counter, requireSelfNoding)
				_ = counter.count
			}
		})
	}
}

