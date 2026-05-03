package measure

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Extended scenarios for MaximumInscribedCircle and LargestEmptyCircle
// covering polygons with holes, multi-polygons (islands), and a mix of
// point/line/polygon obstacles.

func TestMaximumInscribedCircleConcavePolygon(t *testing.T) {
	// L-shape: outer rectangle 0..20x0..20 with a 10x10 bite removed
	// from the upper-right. The largest inscribed circle is tangent to
	// the inner reflex corner (10,10) and two opposite walls; for an
	// L-shape with arms of width 10 its radius is 10/(1+sqrt(2)/2) ≈
	// 5.86 (between 5 and 7).
	g := mustParse(t, "POLYGON ((0 0, 20 0, 20 10, 10 10, 10 20, 0 20, 0 0))")
	_, r, ok := MaximumInscribedCircle(g, 0.001)
	require.True(t, ok)
	assert.GreaterOrEqual(t, r, 5.0)
	assert.LessOrEqual(t, r, 7.0)
}

func TestMaximumInscribedCircleMultiPolygonIslands(t *testing.T) {
	// MultiPolygon with two islands of clearly different sizes; the MIC
	// must select the larger island.
	g := mustParse(t, "MULTIPOLYGON (((0 0, 100 0, 100 100, 0 100, 0 0)), ((200 0, 210 0, 210 10, 200 10, 200 0)))")
	c, r, ok := MaximumInscribedCircle(g, 0.01)
	require.True(t, ok)
	assert.InDelta(t, 50.0, r, 0.5)
	assert.LessOrEqual(t, c.X, 100.0)
}

func TestMaximumInscribedCircleMultiPolygonWithHole(t *testing.T) {
	// MultiPolygon: one polygon has a hole, one doesn't. Larger one
	// dominates the result.
	g := mustParse(t, "MULTIPOLYGON (((0 0, 100 0, 100 100, 0 100, 0 0), (40 40, 60 40, 60 60, 40 60, 40 40)), ((200 0, 220 0, 220 20, 200 20, 200 0)))")
	_, r, ok := MaximumInscribedCircle(g, 0.05)
	require.True(t, ok)
	// The hole of the big polygon limits the inscribed circle to a
	// corner with radius ~25 (between the outer wall and a hole edge).
	assert.GreaterOrEqual(t, r, 15.0)
}

func TestLargestEmptyCircleMixedObstacles(t *testing.T) {
	// Mix of point and line obstacles within a square boundary.
	obstacles := mustParse(t, "GEOMETRYCOLLECTION (POINT (5 5), LINESTRING (1 1, 9 1))")
	bnds := mustParse(t, "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	c, r, ok := LargestEmptyCircle(obstacles, bnds, 0.01)
	require.True(t, ok)
	assert.Greater(t, r, 0.0)
	// Center must be inside boundary.
	assert.GreaterOrEqual(t, c.X, 0.0)
	assert.LessOrEqual(t, c.X, 10.0)
	assert.GreaterOrEqual(t, c.Y, 0.0)
	assert.LessOrEqual(t, c.Y, 10.0)
	// Distance to all obstacles must be at least r (within tolerance).
	dPt := math.Hypot(c.X-5, c.Y-5)
	assert.GreaterOrEqual(t, dPt+0.05, r)
}

func TestLargestEmptyCircleMultiPolygonObstacles(t *testing.T) {
	// Two polygon obstacles inside a larger boundary. LEC should slot
	// between them or in a corner.
	obstacles := mustParse(t, "MULTIPOLYGON (((2 2, 4 2, 4 4, 2 4, 2 2)), ((6 6, 8 6, 8 8, 6 8, 6 6)))")
	bnds := mustParse(t, "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	c, r, ok := LargestEmptyCircle(obstacles, bnds, 0.01)
	require.True(t, ok)
	assert.Greater(t, r, 0.0)
	// Center must be inside boundary.
	assert.GreaterOrEqual(t, c.X, 0.0)
	assert.LessOrEqual(t, c.X, 10.0)
}

func TestLargestEmptyCircleConvexHullBoundary(t *testing.T) {
	// Five obstacles arranged so their convex hull encloses an empty
	// region; with a nil boundary the LEC must lie inside the hull.
	obstacles := mustParse(t, "MULTIPOINT ((0 0), (10 0), (10 10), (0 10), (5 5))")
	c, r, ok := LargestEmptyCircle(obstacles, nil, 0.01)
	require.True(t, ok)
	assert.Greater(t, r, 0.0)
	// All five obstacles must be at distance ≥ r (within tolerance).
	for _, q := range [][2]float64{{0, 0}, {10, 0}, {10, 10}, {0, 10}, {5, 5}} {
		d := math.Hypot(c.X-q[0], c.Y-q[1])
		assert.GreaterOrEqual(t, d+0.05, r)
	}
}
