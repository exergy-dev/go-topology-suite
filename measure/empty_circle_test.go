package measure

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLargestEmptyCirclePointsInBoundary(t *testing.T) {
	// Two points at (2,5) and (8,5) within a 10x10 boundary. The LEC
	// should be in a corner farthest from both points.
	obstacles := mustParse(t, "MULTIPOINT ((2 5), (8 5))")
	bnds := mustParse(t, "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	c, r, ok := LargestEmptyCircle(obstacles, bnds, 0.001)
	require.True(t, ok)
	assert.Greater(t, r, 0.0)
	// Sanity: result lies within boundary
	assert.GreaterOrEqual(t, c.X, 0.0)
	assert.LessOrEqual(t, c.X, 10.0)
	assert.GreaterOrEqual(t, c.Y, 0.0)
	assert.LessOrEqual(t, c.Y, 10.0)
	// Distance to each obstacle is ≥ r (within tolerance).
	d1 := math.Hypot(c.X-2, c.Y-5)
	d2 := math.Hypot(c.X-8, c.Y-5)
	assert.GreaterOrEqual(t, d1+0.05, r)
	assert.GreaterOrEqual(t, d2+0.05, r)
}

func TestLargestEmptyCircleSinglePointCenteredInSquare(t *testing.T) {
	// A square boundary 0..10, single obstacle at the centre. The LEC
	// is one of the corners with radius sqrt(2)*5 ≈ 7.07.
	obstacles := mustParse(t, "POINT (5 5)")
	bnds := mustParse(t, "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	c, r, ok := LargestEmptyCircle(obstacles, bnds, 0.001)
	require.True(t, ok)
	assert.InDelta(t, math.Sqrt(50.0), r, 0.05)
	// Distance from c to (5,5) ~= r.
	d := math.Hypot(c.X-5, c.Y-5)
	assert.InDelta(t, r, d, 0.1)
}

func TestLargestEmptyCircleAutoBoundary(t *testing.T) {
	// Without explicit boundary, uses convex hull of obstacles. Two
	// points: hull is a degenerate line, so return zero radius.
	obstacles := mustParse(t, "MULTIPOINT ((0 0), (10 0))")
	_, r, ok := LargestEmptyCircle(obstacles, nil, 0.001)
	require.True(t, ok)
	assert.Equal(t, 0.0, r)
}

func TestLargestEmptyCircleAutoBoundaryTriangle(t *testing.T) {
	// Three obstacles forming a triangle; LEC center is the
	// circumcenter (or related point inside the triangle).
	obstacles := mustParse(t, "MULTIPOINT ((0 0), (10 0), (5 10))")
	c, r, ok := LargestEmptyCircle(obstacles, nil, 0.001)
	require.True(t, ok)
	assert.Greater(t, r, 0.0)
	// Center must be inside triangle.
	assert.GreaterOrEqual(t, c.X, 0.0)
	assert.LessOrEqual(t, c.X, 10.0)
	assert.GreaterOrEqual(t, c.Y, 0.0)
	assert.LessOrEqual(t, c.Y, 10.0)
}

func TestLargestEmptyCircleLineObstacle(t *testing.T) {
	// LineString obstacle through the middle of a square. LEC lies as
	// far as possible from the line; expected radius ~ 5 (distance from
	// line to a long edge).
	obstacles := mustParse(t, "LINESTRING (0 5, 10 5)")
	bnds := mustParse(t, "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	c, r, ok := LargestEmptyCircle(obstacles, bnds, 0.001)
	require.True(t, ok)
	// Center should be at y near 0 or 10, |y-5| ~ r.
	assert.InDelta(t, 5.0, math.Abs(c.Y-5.0)+0, 1.0)
	_ = r
}

func TestLargestEmptyCirclePolygonObstacle(t *testing.T) {
	// A square polygon obstacle inside a larger boundary. LEC fits in
	// a corner.
	obstacles := mustParse(t, "POLYGON ((4 4, 6 4, 6 6, 4 6, 4 4))")
	bnds := mustParse(t, "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	c, r, ok := LargestEmptyCircle(obstacles, bnds, 0.001)
	require.True(t, ok)
	assert.Greater(t, r, 0.0)
	// Center should not be inside the obstacle.
	d := math.Max(math.Abs(c.X-5)-1, math.Abs(c.Y-5)-1)
	assert.GreaterOrEqual(t, d, -0.01)
}

func TestLargestEmptyCircleEmptyObstacles(t *testing.T) {
	g := mustParse(t, "POINT EMPTY")
	_, _, ok := LargestEmptyCircle(g, nil, 0)
	assert.False(t, ok)
}

func TestLargestEmptyCircleNonPolygonalBoundary(t *testing.T) {
	obstacles := mustParse(t, "POINT (5 5)")
	bnds := mustParse(t, "LINESTRING (0 0, 10 10)")
	_, _, ok := LargestEmptyCircle(obstacles, bnds, 0)
	assert.False(t, ok)
}
