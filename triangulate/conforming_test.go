package triangulate

import (
	"math"
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
)

func TestConformingDelaunayOf_SquareWithDiagonal(t *testing.T) {
	// Unit square with a diagonal constraint forces exactly two triangles
	// joined along that diagonal.
	corners := []geom.XY{
		{X: 0, Y: 0}, {X: 1, Y: 0},
		{X: 1, Y: 1}, {X: 0, Y: 1},
	}
	diag := [2]geom.XY{{X: 0, Y: 0}, {X: 1, Y: 1}}
	tris, err := ConformingDelaunayOf(corners, [][2]geom.XY{diag})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(tris) != 2 {
		t.Fatalf("want 2 triangles, got %d", len(tris))
	}
	// Both triangles must share the diagonal as one of their edges.
	for _, tri := range tris {
		if !triangleHasEdge(tri, diag[0], diag[1]) {
			t.Fatalf("triangle %v does not contain the constraint diagonal", tri)
		}
	}
}

func TestConformingDelaunayOf_NoConstraints(t *testing.T) {
	// Without any constraints the result is just the Delaunay triangulation
	// of the points (here, two triangles for a unit square).
	pts := []geom.XY{
		{X: 0, Y: 0}, {X: 1, Y: 0},
		{X: 1, Y: 1}, {X: 0, Y: 1},
	}
	tris, err := ConformingDelaunayOf(pts, nil)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(tris) != 2 {
		t.Fatalf("want 2 triangles, got %d", len(tris))
	}
}

func TestConformingDelaunayOf_SplitsEncroachingSegment(t *testing.T) {
	// Constrain a segment that gets encroached by an interior point. The
	// algorithm must split the segment at its midpoint and produce a
	// triangulation that includes both halves of the original segment as
	// triangle edges.
	pts := []geom.XY{
		// Site exactly at the midpoint will encroach the diameter circle
		// of the long horizontal segment below.
		{X: 0.5, Y: 0.1},
	}
	seg := [2]geom.XY{{X: 0, Y: 0}, {X: 1, Y: 0}}
	tris, err := ConformingDelaunayOf(pts, [][2]geom.XY{seg})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	// The encroached segment should be split. The resulting triangulation
	// must include both sub-segments as edges.
	mid := geom.XY{X: 0.5, Y: 0}
	if !anyTriangleHasEdge(tris, geom.XY{X: 0, Y: 0}, mid) {
		t.Fatal("expected sub-segment (0,0)-(0.5,0) to appear as an edge")
	}
	if !anyTriangleHasEdge(tris, mid, geom.XY{X: 1, Y: 0}) {
		t.Fatal("expected sub-segment (0.5,0)-(1,0) to appear as an edge")
	}
}

// triangleHasEdge reports whether triangle tri has p-q (in either
// direction) as one of its three edges.
func triangleHasEdge(tri Triangle, p, q geom.XY) bool {
	type edge struct{ a, b geom.XY }
	edges := []edge{
		{tri.P0, tri.P1}, {tri.P1, tri.P2}, {tri.P2, tri.P0},
	}
	for _, e := range edges {
		if (xyClose(e.a, p) && xyClose(e.b, q)) ||
			(xyClose(e.a, q) && xyClose(e.b, p)) {
			return true
		}
	}
	return false
}

func anyTriangleHasEdge(tris []Triangle, p, q geom.XY) bool {
	for _, t := range tris {
		if triangleHasEdge(t, p, q) {
			return true
		}
	}
	return false
}

func xyClose(a, b geom.XY) bool {
	return math.Abs(a.X-b.X) < 1e-9 && math.Abs(a.Y-b.Y) < 1e-9
}
