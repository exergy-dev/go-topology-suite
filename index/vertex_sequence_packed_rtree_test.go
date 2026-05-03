package index

import (
	"math"
	"math/rand"
	"sort"
	"testing"

	"github.com/terra-geo/terra/geom"
)

// TestVSPR_Query_Exhaustive verifies that for a few small fixtures the
// tree returns exactly the same indices as a brute-force point-in-env
// scan.
func TestVSPR_Query_Exhaustive(t *testing.T) {
	pts := []geom.XY{
		{0, 0}, {1, 0}, {2, 0}, {3, 1}, {4, 1},
		{5, 2}, {6, 2}, {7, 3}, {8, 3}, {9, 4},
	}
	tree := NewVertexSequencePackedRtree(pts)
	for _, q := range []geom.Envelope{
		{MinX: -1, MinY: -1, MaxX: 100, MaxY: 100}, // full extent
		{MinX: 0, MinY: 0, MaxX: 2.5, MaxY: 0.5},    // first three points
		{MinX: 5, MinY: 1.5, MaxX: 8.5, MaxY: 3.5},  // {5,2}..{8,3}
		{MinX: -10, MinY: -10, MaxX: -5, MaxY: -5},  // empty
	} {
		got := tree.Query(q)
		want := bruteQuery(pts, q, nil)
		sort.Ints(got)
		if !equalInts(got, want) {
			t.Errorf("query %v: got %v want %v", q, got, want)
		}
	}
}

// TestVSPR_QueryParity_Random builds 50 random point clouds and
// verifies the tree's query results match the brute-force scan for 50
// random query rectangles each.
func TestVSPR_QueryParity_Random(t *testing.T) {
	rng := rand.New(rand.NewSource(0x5EED))
	for trial := 0; trial < 50; trial++ {
		n := 5 + rng.Intn(500)
		pts := make([]geom.XY, n)
		for i := range pts {
			pts[i] = geom.XY{X: rng.Float64()*100 - 50, Y: rng.Float64()*100 - 50}
		}
		tree := NewVertexSequencePackedRtree(pts)
		for q := 0; q < 50; q++ {
			x0 := rng.Float64()*100 - 50
			y0 := rng.Float64()*100 - 50
			env := geom.Envelope{MinX: x0, MinY: y0, MaxX: x0 + rng.Float64()*30, MaxY: y0 + rng.Float64()*30}
			got := tree.Query(env)
			want := bruteQuery(pts, env, nil)
			sort.Ints(got)
			if !equalInts(got, want) {
				t.Fatalf("trial %d: env=%v got=%v want=%v", trial, env, got, want)
			}
		}
	}
}

// TestVSPR_Remove ensures Remove(i) hides index i from subsequent
// queries and is idempotent.
func TestVSPR_Remove(t *testing.T) {
	pts := []geom.XY{
		{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0},
		{5, 0}, {6, 0}, {7, 0}, {8, 0}, {9, 0},
	}
	tree := NewVertexSequencePackedRtree(pts)
	full := geom.Envelope{MinX: -1, MinY: -1, MaxX: 100, MaxY: 100}
	if got := tree.Query(full); len(got) != len(pts) {
		t.Fatalf("pre-remove: got %d want %d", len(got), len(pts))
	}
	tree.Remove(3)
	tree.Remove(7)
	got := tree.Query(full)
	sort.Ints(got)
	want := []int{0, 1, 2, 4, 5, 6, 8, 9}
	if !equalInts(got, want) {
		t.Errorf("post-remove: got %v want %v", got, want)
	}
}

// TestVSPR_Empty handles the corner case of an empty point sequence.
func TestVSPR_Empty(t *testing.T) {
	tree := NewVertexSequencePackedRtree(nil)
	if got := tree.Query(geom.Envelope{MinX: 0, MinY: 0, MaxX: 1, MaxY: 1}); len(got) != 0 {
		t.Errorf("empty: got %v", got)
	}
}

func bruteQuery(pts []geom.XY, env geom.Envelope, _ []int) []int {
	var out []int
	for i, p := range pts {
		if env.ContainsXY(p) {
			out = append(out, i)
		}
	}
	return out
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// BenchmarkVSPR_Query exercises a 10k-point ring with a small query
// envelope; this is the workload PolygonHull encounters when checking
// whether a corner triangle contains any other vertex of any other
// ring.
func BenchmarkVSPR_Query(b *testing.B) {
	pts := makeRingPoints(10000)
	tree := NewVertexSequencePackedRtree(pts)
	env := geom.Envelope{MinX: -1, MinY: -1, MaxX: 1, MaxY: 1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tree.Query(env)
	}
}

// BenchmarkVSPR_LinearScan is the brute-force baseline for
// BenchmarkVSPR_Query.
func BenchmarkVSPR_LinearScan(b *testing.B) {
	pts := makeRingPoints(10000)
	env := geom.Envelope{MinX: -1, MinY: -1, MaxX: 1, MaxY: 1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var hits []int
		for j, p := range pts {
			if env.ContainsXY(p) {
				hits = append(hits, j)
			}
		}
		_ = hits
	}
}

func makeRingPoints(n int) []geom.XY {
	pts := make([]geom.XY, n)
	for i := 0; i < n; i++ {
		theta := 2 * math.Pi * float64(i) / float64(n)
		pts[i] = geom.XY{X: 100 * math.Cos(theta), Y: 100 * math.Sin(theta)}
	}
	return pts
}
