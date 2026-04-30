package overlay

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/measure"
	"github.com/terra-geo/terra/wkt"
)

func mp(t *testing.T, s string) geom.Geometry {
	t.Helper()
	g, err := wkt.Unmarshal(s)
	require.NoError(t, err)
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
	require.NoError(t, err)
	want := 25.0 // 5×5
	assert.InDelta(t, want, measure.Area(got), 0.01, "area")
}

func TestGHUnionTwoSquares(t *testing.T) {
	a := mp(t, "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	b := mp(t, "POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))")
	got, err := Union(a, b)
	require.NoError(t, err)
	// Areas: A=100, B=100, A∩B=25 → A∪B = 175.
	want := 175.0
	assert.InDelta(t, want, measure.Area(got), 0.5, "area")
}

func TestGHDifferenceTwoSquares(t *testing.T) {
	a := mp(t, "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	b := mp(t, "POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))")
	got, err := Difference(a, b)
	require.NoError(t, err)
	// A \ B = 100 - 25 = 75.
	want := 75.0
	assert.InDelta(t, want, measure.Area(got), 0.5, "area")
}

func TestGHSymmetricDifference(t *testing.T) {
	a := mp(t, "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	b := mp(t, "POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))")
	got, err := SymmetricDifference(a, b)
	require.NoError(t, err)
	// A∪B - A∩B = 175 - 25 = 150.
	want := 150.0
	assert.InDelta(t, want, measure.Area(got), 0.5, "area")
}

func TestGHContainmentNoIntersection(t *testing.T) {
	outer := mp(t, "POLYGON ((0 0, 100 0, 100 100, 0 100, 0 0))")
	inner := mp(t, "POLYGON ((40 40, 60 40, 60 60, 40 60, 40 40))")

	// Intersection: smaller one.
	ix, _ := IntersectionGeneral(outer, inner)
	assert.InDelta(t, 400.0, measure.Area(ix), 1.0, "contained intersection area")

	// Union: larger one.
	un, _ := Union(outer, inner)
	assert.InDelta(t, 10000.0, measure.Area(un), 1.0, "contained union area")

	// Difference: outer with inner as hole = 10000 - 400 = 9600.
	d, _ := Difference(outer, inner)
	assert.InDelta(t, 9600.0, measure.Area(d), 1.0, "contained difference area")
}

func TestGHDisjointNoIntersection(t *testing.T) {
	a := mp(t, "POLYGON ((0 0, 1 0, 1 1, 0 1, 0 0))")
	b := mp(t, "POLYGON ((5 5, 6 5, 6 6, 5 6, 5 5))")

	ix, _ := IntersectionGeneral(a, b)
	assert.True(t, ix.IsEmpty(), "disjoint intersection should be empty")
	un, _ := Union(a, b)
	assert.Equal(t, geom.MultiPolygonType, un.Type(), "disjoint union should be MultiPolygon")
	d, _ := Difference(a, b)
	assert.InDelta(t, 1.0, measure.Area(d), 0.01, "disjoint difference area")
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
	require.NoError(t, err)
	require.False(t, got.IsEmpty(), "intersection of crossing rectangles should not be empty")
	assert.InDelta(t, 4.0, measure.Area(got), 0.01, "crossing-rect intersection area")
}
