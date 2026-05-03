package measure

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaximumInscribedCircleSquare(t *testing.T) {
	g := mustParse(t, "POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	c, r, ok := MaximumInscribedCircle(g, 0.001)
	require.True(t, ok)
	assert.InDelta(t, 5.0, c.X, 0.05)
	assert.InDelta(t, 5.0, c.Y, 0.05)
	assert.InDelta(t, 5.0, r, 0.05)
}

func TestMaximumInscribedCircleRectangle(t *testing.T) {
	// 20×4 rectangle: MIC radius = 2 (limited by short axis), centered along midline.
	g := mustParse(t, "POLYGON ((0 0, 20 0, 20 4, 0 4, 0 0))")
	c, r, ok := MaximumInscribedCircle(g, 0.001)
	require.True(t, ok)
	assert.InDelta(t, 2.0, c.Y, 0.05)
	assert.InDelta(t, 2.0, r, 0.05)
}

func TestMaximumInscribedCirclePolygonWithHole(t *testing.T) {
	// 20x20 outer, with a centred 4x4 hole. The MIC must avoid the hole;
	// the largest inscribed circle is roughly tangent to the outer ring
	// and the hole boundary, with radius about (20-4)/4 = 4 (each
	// quadrant). It is positioned in one of the four quadrants.
	g := mustParse(t, "POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0), (8 8, 12 8, 12 12, 8 12, 8 8))")
	_, r, ok := MaximumInscribedCircle(g, 0.01)
	require.True(t, ok)
	// Quadrant fits a circle of radius ~4 (inscribed between two
	// bounding rectangles of width 4 and 8 — limited by the narrower
	// gap). Radius should be at least ~3.5 and at most ~6.
	assert.GreaterOrEqual(t, r, 3.0)
	assert.LessOrEqual(t, r, 6.5)
}

func TestMaximumInscribedCircleMultiPolygon(t *testing.T) {
	// Two disjoint squares of side 10 and 4. MIC is in the larger one,
	// radius ~5.
	g := mustParse(t, "MULTIPOLYGON (((0 0, 10 0, 10 10, 0 10, 0 0)), ((20 0, 24 0, 24 4, 20 4, 20 0)))")
	c, r, ok := MaximumInscribedCircle(g, 0.001)
	require.True(t, ok)
	assert.InDelta(t, 5.0, r, 0.05)
	// Center should be in the large square.
	assert.GreaterOrEqual(t, c.X, 0.0)
	assert.LessOrEqual(t, c.X, 10.0)
}

func TestMaximumInscribedCircleNonPolygon(t *testing.T) {
	g := mustParse(t, "POINT (1 1)")
	_, _, ok := MaximumInscribedCircle(g, 0)
	assert.False(t, ok)
}

func TestMaximumInscribedCircleEmpty(t *testing.T) {
	g := mustParse(t, "POLYGON EMPTY")
	_, _, ok := MaximumInscribedCircle(g, 0)
	assert.False(t, ok)
}

func TestMaximumInscribedCircleAutoTolerance(t *testing.T) {
	g := mustParse(t, "POLYGON ((0 0, 100 0, 100 100, 0 100, 0 0))")
	_, r, ok := MaximumInscribedCircle(g, 0)
	require.True(t, ok)
	assert.InDelta(t, 50.0, r, 0.5)
}

func TestMaximumInscribedCircleTriangle(t *testing.T) {
	// Equilateral-ish triangle. Inscribed circle radius = area / s,
	// where s is semi-perimeter. For (0,0)-(10,0)-(5,10): area=50,
	// sides = 10, sqrt(125), sqrt(125), s = 5+sqrt(125), r = 50/s.
	g := mustParse(t, "POLYGON ((0 0, 10 0, 5 10, 0 0))")
	_, r, ok := MaximumInscribedCircle(g, 0.001)
	require.True(t, ok)
	expected := 50.0 / (5.0 + math.Sqrt(125.0))
	assert.InDelta(t, expected, r, 0.05)
}
