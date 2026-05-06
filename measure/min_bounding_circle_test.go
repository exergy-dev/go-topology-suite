package measure

import (
	"math"
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMinimumBoundingCircle_Empty(t *testing.T) {
	g := geom.NewEmptyPoint(nil, geom.LayoutXY)
	_, _, ok := MinimumBoundingCircle(g)
	require.False(t, ok, "expected empty input → ok=false")
}

func TestMinimumBoundingCircle_SinglePoint(t *testing.T) {
	g := geom.NewPoint(nil, geom.XY{X: 3, Y: 4})
	c, r, ok := MinimumBoundingCircle(g)
	require.True(t, ok)
	assert.Equal(t, 0.0, r)
	assert.Equal(t, 3.0, c.X)
	assert.Equal(t, 4.0, c.Y)
}

func TestMinimumBoundingCircle_TwoPoints(t *testing.T) {
	g := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 4, Y: 0}})
	c, r, ok := MinimumBoundingCircle(g)
	require.True(t, ok)
	assert.InDelta(t, 2.0, c.X, 1e-9)
	assert.InDelta(t, 0.0, c.Y, 1e-9)
	assert.InDelta(t, 2.0, r, 1e-9)
}

func TestMinimumBoundingCircle_Square(t *testing.T) {
	// Unit square; MBC should be circumscribed circle.
	g := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0},
	})
	c, r, ok := MinimumBoundingCircle(g)
	require.True(t, ok)
	assert.InDelta(t, 0.5, c.X, 1e-9)
	assert.InDelta(t, 0.5, c.Y, 1e-9)
	assert.InDelta(t, math.Sqrt2/2, r, 1e-9)
	// All vertices must lie on or inside the circle.
	for _, p := range []geom.XY{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}} {
		d := math.Hypot(p.X-c.X, p.Y-c.Y)
		assert.LessOrEqual(t, d, r+1e-9, "vertex %+v at distance %v > r=%v", p, d, r)
	}
}

func TestMinimumBoundingCircle_Triangle(t *testing.T) {
	// Acute triangle: circumscribed circle should pass through all three.
	pts := []geom.XY{{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 2, Y: 3}}
	g := geom.NewPolygon(nil, append(append([]geom.XY{}, pts...), pts[0]))
	c, r, ok := MinimumBoundingCircle(g)
	require.True(t, ok)
	for _, p := range pts {
		d := math.Hypot(p.X-c.X, p.Y-c.Y)
		assert.InDelta(t, r, d, 1e-7, "vertex %+v dist=%v want r=%v", p, d, r)
	}
}

func TestMinimumBoundingCircle_ObtuseTriangle(t *testing.T) {
	// Obtuse triangle: MBC determined by the two endpoints of the longest side.
	pts := []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 1, Y: 1}}
	g := geom.NewPolygon(nil, append(append([]geom.XY{}, pts...), pts[0]))
	c, r, ok := MinimumBoundingCircle(g)
	require.True(t, ok)
	// Longest side is (0,0)-(10,0), so MBC is centred at (5,0) with r=5.
	assert.InDelta(t, 5.0, c.X, 1e-9)
	assert.InDelta(t, 0.0, c.Y, 1e-9)
	assert.InDelta(t, 5.0, r, 1e-9)
	for _, p := range pts {
		d := math.Hypot(p.X-c.X, p.Y-c.Y)
		assert.LessOrEqual(t, d, r+1e-9, "vertex %+v dist=%v exceeds r=%v", p, d, r)
	}
}
