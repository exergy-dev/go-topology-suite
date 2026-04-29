package overlay

import (
	"testing"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/measure"
	"github.com/terra-geo/terra/wkt"
)

func mp(t *testing.T, s string) geom.Geometry {
	t.Helper()
	g, err := wkt.Unmarshal(s)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

// Two overlapping squares — Greiner-Hormann path is exercised because
// neither side is convex-only-with-respect-to-the-other-as-clipper in
// the routing sense. Actually Intersection still picks convex fast-path
// here; verify the general path explicitly.
func TestGHIntersectionTwoSquares(t *testing.T) {
	a := mp(t, "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	b := mp(t, "POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))")
	got, err := IntersectionGeneral(a, b)
	if err != nil {
		t.Fatal(err)
	}
	want := 25.0 // 5×5
	if a := measure.Area(got); a < 24.99 || a > 25.01 {
		t.Errorf("area = %v, want %v", a, want)
	}
}

func TestGHUnionTwoSquares(t *testing.T) {
	a := mp(t, "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	b := mp(t, "POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))")
	got, err := Union(a, b)
	if err != nil {
		t.Fatal(err)
	}
	// Areas: A=100, B=100, A∩B=25 → A∪B = 175.
	want := 175.0
	if a := measure.Area(got); a < 174.5 || a > 175.5 {
		t.Errorf("area = %v, want %v", a, want)
	}
}

func TestGHDifferenceTwoSquares(t *testing.T) {
	a := mp(t, "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	b := mp(t, "POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))")
	got, err := Difference(a, b)
	if err != nil {
		t.Fatal(err)
	}
	// A \ B = 100 - 25 = 75.
	want := 75.0
	if ar := measure.Area(got); ar < 74.5 || ar > 75.5 {
		t.Errorf("area = %v, want %v", ar, want)
	}
}

func TestGHSymmetricDifference(t *testing.T) {
	a := mp(t, "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	b := mp(t, "POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))")
	got, err := SymmetricDifference(a, b)
	if err != nil {
		t.Fatal(err)
	}
	// A∪B - A∩B = 175 - 25 = 150.
	want := 150.0
	if ar := measure.Area(got); ar < 149.5 || ar > 150.5 {
		t.Errorf("area = %v, want %v", ar, want)
	}
}

func TestGHContainmentNoIntersection(t *testing.T) {
	outer := mp(t, "POLYGON ((0 0, 100 0, 100 100, 0 100, 0 0))")
	inner := mp(t, "POLYGON ((40 40, 60 40, 60 60, 40 60, 40 40))")

	// Intersection: smaller one.
	ix, _ := IntersectionGeneral(outer, inner)
	if a := measure.Area(ix); a < 399 || a > 401 {
		t.Errorf("contained intersection area = %v, want 400", a)
	}

	// Union: larger one.
	un, _ := Union(outer, inner)
	if a := measure.Area(un); a < 9999 || a > 10001 {
		t.Errorf("contained union area = %v, want 10000", a)
	}

	// Difference: outer with inner as hole = 10000 - 400 = 9600.
	d, _ := Difference(outer, inner)
	if a := measure.Area(d); a < 9599 || a > 9601 {
		t.Errorf("contained difference area = %v, want 9600", a)
	}
}

func TestGHDisjointNoIntersection(t *testing.T) {
	a := mp(t, "POLYGON ((0 0, 1 0, 1 1, 0 1, 0 0))")
	b := mp(t, "POLYGON ((5 5, 6 5, 6 6, 5 6, 5 5))")

	ix, _ := IntersectionGeneral(a, b)
	if !ix.IsEmpty() {
		t.Errorf("disjoint intersection should be empty")
	}
	un, _ := Union(a, b)
	if un.Type() != geom.MultiPolygonType {
		t.Errorf("disjoint union should be MultiPolygon, got %v", un.Type())
	}
	d, _ := Difference(a, b)
	if a := measure.Area(d); a < 0.99 || a > 1.01 {
		t.Errorf("disjoint difference = a, area %v want 1", a)
	}
}

// Two crossed L-shaped polygons producing a multi-piece intersection.
// Uses integer coords so alpha sorting and intersection arithmetic are
// exact-representable.
func TestGHCrossingRectangles(t *testing.T) {
	// Two rectangles crossing at right angles forming a "+" shape.
	// Their intersection is a 2×2 square at the centre.
	a := mp(t, "POLYGON ((-5 -1, 5 -1, 5 1, -5 1, -5 -1))")
	b := mp(t, "POLYGON ((-1 -5, 1 -5, 1 5, -1 5, -1 -5))")
	got, err := IntersectionGeneral(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if got.IsEmpty() {
		t.Fatalf("intersection of crossing rectangles should not be empty")
	}
	a1 := measure.Area(got)
	if a1 < 3.99 || a1 > 4.01 {
		t.Errorf("crossing-rect intersection area = %v, want 4", a1)
	}
}
