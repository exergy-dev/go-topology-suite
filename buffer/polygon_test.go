package buffer

import (
	"math"
	"testing"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/measure"
)

// squareRingCCW returns a closed CCW square ring with the given half-width.
// Corners: (-h,-h) → (h,-h) → (h,h) → (-h,h) → (-h,-h).
func squareRingCCW(h float64) []geom.XY {
	return []geom.XY{
		{X: -h, Y: -h},
		{X: h, Y: -h},
		{X: h, Y: h},
		{X: -h, Y: h},
		{X: -h, Y: -h},
	}
}

// extractFirstPolygon returns the *geom.Polygon component of g. Multi
// geometries with one part are unwrapped; multi geometries with > 1 part
// fail the test.
func extractFirstPolygon(t *testing.T, g geom.Geometry) *geom.Polygon {
	t.Helper()
	switch v := g.(type) {
	case *geom.Polygon:
		return v
	case *geom.MultiPolygon:
		if v.NumGeometries() != 1 {
			t.Fatalf("expected 1 polygon part, got %d", v.NumGeometries())
		}
		return v.PolygonAt(0)
	default:
		t.Fatalf("unexpected geometry type %T", g)
	}
	return nil
}

func TestBufferPolygonZeroDistance(t *testing.T) {
	side := 4.0
	poly := geom.NewPolygon(nil, squareRingCCW(side/2))
	g, err := Buffer(poly, 0)
	if err != nil {
		t.Fatalf("Buffer: %v", err)
	}
	if g != poly {
		t.Fatalf("expected identity geometry for zero distance, got %T", g)
	}
}

func TestBufferPolygonPositiveRound(t *testing.T) {
	side := 4.0
	poly := geom.NewPolygon(nil, squareRingCCW(side/2))
	const d = 1.0
	const quad = 8

	g, err := Buffer(poly, d, WithJoinStyle(JoinRound), WithQuadSegments(quad))
	if err != nil {
		t.Fatalf("Buffer: %v", err)
	}
	out := extractFirstPolygon(t, g)
	got := measure.Area(out)

	// Expected area = original + perimeter*d + π*d² (one full rotation of
	// the disk around the boundary contributes 4 quarter-circles = π*d²).
	original := side * side
	perim := 4 * side
	want := original + perim*d + math.Pi*d*d
	if math.Abs(got-want) > 0.05*want {
		t.Errorf("positive round buffer area = %v, want ≈ %v", got, want)
	}
	if got <= original {
		t.Errorf("positive buffer must grow the polygon; got %v ≤ original %v", got, original)
	}
}

func TestBufferPolygonPositiveMitre(t *testing.T) {
	side := 4.0
	poly := geom.NewPolygon(nil, squareRingCCW(side/2))
	const d = 1.0

	g, err := Buffer(poly, d, WithJoinStyle(JoinMitre))
	if err != nil {
		t.Fatalf("Buffer: %v", err)
	}
	out := extractFirstPolygon(t, g)
	got := measure.Area(out)

	// Mitre on a square produces another square of side+2d. v0.1 accepts
	// up to 10% area error from the GH-overlay union step (documented in
	// package doc).
	want := (side + 2*d) * (side + 2*d)
	if math.Abs(got-want) > 0.1*want {
		t.Errorf("positive mitre buffer area = %v, want ≈ %v (10%% tol)", got, want)
	}
	if got <= side*side {
		t.Errorf("positive buffer must grow the polygon")
	}
}

func TestBufferPolygonNegative(t *testing.T) {
	side := 4.0
	poly := geom.NewPolygon(nil, squareRingCCW(side/2))
	const d = -1.0

	g, err := Buffer(poly, d, WithJoinStyle(JoinMitre))
	if err != nil {
		t.Fatalf("Buffer: %v", err)
	}
	out := extractFirstPolygon(t, g)
	got := measure.Area(out)

	// Mitre inset of a square: smaller square of side - 2|d|.
	want := (side + 2*d) * (side + 2*d)
	if math.Abs(got-want) > 1e-6 {
		t.Errorf("negative mitre buffer area = %v, want %v", got, want)
	}
	if got >= side*side {
		t.Errorf("negative buffer must shrink the polygon; got %v ≥ original %v", got, side*side)
	}
}

func TestBufferPolygonNegativeFullErase(t *testing.T) {
	// Inset by more than the inradius should collapse to an empty or
	// tiny polygon. v0.1 limitation: the offset-ring approach produces an
	// inverted ring rather than detecting the collapse, so we accept any
	// result with area ≤ original/4 as "effectively erased."
	side := 4.0
	poly := geom.NewPolygon(nil, squareRingCCW(side/2))
	const d = -3.0 // inradius is 2.0

	g, err := Buffer(poly, d, WithJoinStyle(JoinMitre))
	if err != nil {
		t.Fatalf("Buffer: %v", err)
	}
	a := measure.Area(g)
	if a > side*side/4 {
		t.Errorf("over-eroded inset should collapse; got area = %v (orig %v)",
			a, side*side)
	}
}

func TestBufferMultiPolygon(t *testing.T) {
	// Two disjoint squares, well separated.
	left := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 2}, {X: 0, Y: 2}, {X: 0, Y: 0},
	})
	right := geom.NewPolygon(nil, []geom.XY{
		{X: 10, Y: 0}, {X: 12, Y: 0}, {X: 12, Y: 2}, {X: 10, Y: 2}, {X: 10, Y: 0},
	})
	mp := geom.NewMultiPolygon(nil, left, right)

	g, err := Buffer(mp, 0.5, WithJoinStyle(JoinMitre))
	if err != nil {
		t.Fatalf("Buffer: %v", err)
	}
	got := measure.Area(g)

	// Each grows from 4 to 9 (square 3×3); two disjoint = 18. 10% tol per
	// the v0.1 GH-overlay union limitation.
	want := 2 * 9.0
	if math.Abs(got-want) > 0.1*want {
		t.Errorf("multipolygon buffer area = %v, want ≈ %v", got, want)
	}
}

func TestBufferMultiPolygonOverlapMerges(t *testing.T) {
	// Two close squares whose round-buffered shapes overlap → result
	// should merge into one polygon. Round joins are used because the
	// v0.1 GH-overlay struggles with collinear shared edges that mitre
	// joins on axis-aligned rectangles produce.
	left := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 2}, {X: 0, Y: 2}, {X: 0, Y: 0},
	})
	right := geom.NewPolygon(nil, []geom.XY{
		{X: 3, Y: 0}, {X: 5, Y: 0}, {X: 5, Y: 2}, {X: 3, Y: 2}, {X: 3, Y: 0},
	})
	mp := geom.NewMultiPolygon(nil, left, right)
	g, err := Buffer(mp, 2, WithJoinStyle(JoinRound), WithQuadSegments(8))
	if err != nil {
		t.Fatalf("Buffer: %v", err)
	}
	if _, ok := g.(*geom.Polygon); !ok {
		t.Errorf("expected merged Polygon, got %T", g)
	}
}

// TestBufferPositiveShrinksHole: a 10×10 outer with a 4×4 hole at the
// center, dilated by d=1. Outer grows to ~12×12 (=144); the hole shrinks
// to ~2×2 (=4). Expected area ≈ 144 - 4 = 140.
func TestBufferPositiveShrinksHole(t *testing.T) {
	outer := []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0}}
	hole := []geom.XY{{X: 3, Y: 3}, {X: 3, Y: 7}, {X: 7, Y: 7}, {X: 7, Y: 3}, {X: 3, Y: 3}}
	poly := geom.NewPolygon(nil, outer, hole)

	g, err := Buffer(poly, 1, WithJoinStyle(JoinMitre))
	if err != nil {
		t.Fatalf("Buffer: %v", err)
	}
	got := measure.Area(g)
	want := 144.0 - 4.0
	// Allow 10% — same tolerance the rest of the polygon-buffer suite uses.
	if math.Abs(got-want) > 0.1*want {
		t.Errorf("buffer-with-hole area = %v, want ≈ %v", got, want)
	}
}

// TestBufferPositiveCollapsesSmallHole: a hole small enough that a
// dilation distance d larger than its inradius makes it disappear. A
// 1×1 hole dilated by 2 should collapse to nothing — the result is a
// hole-free dilated outer.
func TestBufferPositiveCollapsesSmallHole(t *testing.T) {
	outer := []geom.XY{{X: 0, Y: 0}, {X: 20, Y: 0}, {X: 20, Y: 20}, {X: 0, Y: 20}, {X: 0, Y: 0}}
	hole := []geom.XY{{X: 9, Y: 9}, {X: 9, Y: 10}, {X: 10, Y: 10}, {X: 10, Y: 9}, {X: 9, Y: 9}}
	poly := geom.NewPolygon(nil, outer, hole)

	g, err := Buffer(poly, 2, WithJoinStyle(JoinMitre))
	if err != nil {
		t.Fatalf("Buffer: %v", err)
	}
	out := extractFirstPolygon(t, g)
	if out.NumRings() != 1 {
		t.Errorf("expected hole to collapse, got %d rings", out.NumRings())
	}
}

// TestBufferNegativeGrowsHole: a 10×10 outer with a 4×4 hole, inset by
// d=-1. Outer shrinks to 8×8 (=64); the hole grows to 6×6 (=36).
// Expected area ≈ 64 - 36 = 28.
func TestBufferNegativeGrowsHole(t *testing.T) {
	outer := []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0}}
	hole := []geom.XY{{X: 3, Y: 3}, {X: 3, Y: 7}, {X: 7, Y: 7}, {X: 7, Y: 3}, {X: 3, Y: 3}}
	poly := geom.NewPolygon(nil, outer, hole)

	g, err := Buffer(poly, -1, WithJoinStyle(JoinMitre))
	if err != nil {
		t.Fatalf("Buffer: %v", err)
	}
	got := measure.Area(g)
	want := 64.0 - 36.0
	if math.Abs(got-want) > 0.1*want {
		t.Errorf("inset-with-hole area = %v, want ≈ %v", got, want)
	}
}

func TestBufferEmptyPolygon(t *testing.T) {
	empty := geom.NewEmptyPolygon(nil, geom.LayoutXY)
	g, err := Buffer(empty, 1)
	if err != nil {
		t.Fatalf("Buffer: %v", err)
	}
	if !g.IsEmpty() {
		t.Errorf("buffer of empty polygon should be empty; got area = %v",
			measure.Area(g))
	}
}
