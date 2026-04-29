package overlay

import (
	"testing"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/measure"
	"github.com/terra-geo/terra/wkt"
)

var _ = geom.PointType // keep geom import even if unused after edits

func mustParse(t *testing.T, s string) geom.Geometry {
	t.Helper()
	g, err := wkt.Unmarshal(s)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

// Sutherland-Hodgman: clipping a 10×10 square with a centred 5×5 square
// yields the centre 5×5 square.
func TestIntersectionSquareSquare(t *testing.T) {
	subj := mustParse(t, "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	clip := mustParse(t, "POLYGON ((2 2, 7 2, 7 7, 2 7, 2 2))")
	got, err := Intersection(subj, clip)
	if err != nil {
		t.Fatal(err)
	}
	a := measure.Area(got)
	if a != 25 {
		t.Errorf("area = %v, want 25", a)
	}
}

// L-shaped subject through a convex clipper: still fine because the
// subject (which is non-convex) is the one being clipped.
func TestIntersectionLShapeByBox(t *testing.T) {
	subj := mustParse(t, "POLYGON ((0 0, 4 0, 4 2, 2 2, 2 4, 0 4, 0 0))")
	clip := mustParse(t, "POLYGON ((0 0, 3 0, 3 3, 0 3, 0 0))")
	got, err := Intersection(subj, clip)
	if err != nil {
		t.Fatal(err)
	}
	a := measure.Area(got)
	// Expected intersection: L ∩ box = (3×2 left arm + 2×1 bottom-right strip)
	// = subject ∩ clipper. By visual inspection: the bottom 3×2 strip (6) plus
	// the left 2×1 strip from y=2..3 = 6 + 2 = 8.
	if a < 7.9 || a > 8.1 {
		t.Errorf("area = %v, want ≈ 8", a)
	}
}

func TestIntersectionDisjointEmpty(t *testing.T) {
	subj := mustParse(t, "POLYGON ((0 0, 1 0, 1 1, 0 1, 0 0))")
	clip := mustParse(t, "POLYGON ((10 10, 11 10, 11 11, 10 11, 10 10))")
	got, _ := Intersection(subj, clip)
	if !got.IsEmpty() {
		t.Errorf("disjoint intersection should be empty, got %+v", got)
	}
}

func TestNonConvexClipperFallsBackToGH(t *testing.T) {
	// Now that GH is wired up, a non-convex clipper goes through the
	// general path and returns a real polygon (used to be ErrUnsupportedKernel).
	subj := mustParse(t, "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	clip := mustParse(t, "POLYGON ((0 0, 4 0, 4 2, 2 2, 2 4, 0 4, 0 0))") // L-shape
	got, err := Intersection(subj, clip)
	if err != nil {
		t.Fatal(err)
	}
	if got.IsEmpty() {
		t.Errorf("L-shape ∩ square should not be empty")
	}
}
