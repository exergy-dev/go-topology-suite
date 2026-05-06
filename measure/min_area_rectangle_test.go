package measure

import (
	"math"
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/kernel/planar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMinimumAreaRectangle_Empty(t *testing.T) {
	g := geom.NewEmptyPolygon(nil, geom.LayoutXY)
	_, ok := MinimumAreaRectangle(g)
	require.False(t, ok, "expected ok=false")
}

func TestMinimumAreaRectangle_Square(t *testing.T) {
	g := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 4, Y: 4}, {X: 0, Y: 4}, {X: 0, Y: 0},
	})
	rect, ok := MinimumAreaRectangle(g)
	require.True(t, ok)
	a := math.Abs((planar.Kernel{}).RingArea(rect.Ring(0)))
	assert.InDelta(t, 16.0, a, 1e-9)
}

func TestMinimumAreaRectangle_RotatedSquare(t *testing.T) {
	// 4x4 square rotated 45° → area 16, MAR should still find area 16.
	pts := []geom.XY{{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 4, Y: 4}, {X: 0, Y: 4}}
	theta := math.Pi / 4
	cos, sin := math.Cos(theta), math.Sin(theta)
	for i := range pts {
		x, y := pts[i].X, pts[i].Y
		pts[i] = geom.XY{X: x*cos - y*sin, Y: x*sin + y*cos}
	}
	g := geom.NewPolygon(nil, append(append([]geom.XY{}, pts...), pts[0]))
	rect, ok := MinimumAreaRectangle(g)
	require.True(t, ok)
	a := math.Abs((planar.Kernel{}).RingArea(rect.Ring(0)))
	assert.InDelta(t, 16.0, a, 1e-7)
}

func TestMinimumAreaRectangle_LongDiagonalRectangle(t *testing.T) {
	// 8x2 rectangle rotated 30° — MAR area should be 16.
	pts := []geom.XY{{X: 0, Y: 0}, {X: 8, Y: 0}, {X: 8, Y: 2}, {X: 0, Y: 2}}
	theta := math.Pi / 6
	cos, sin := math.Cos(theta), math.Sin(theta)
	for i := range pts {
		x, y := pts[i].X, pts[i].Y
		pts[i] = geom.XY{X: x*cos - y*sin, Y: x*sin + y*cos}
	}
	g := geom.NewPolygon(nil, append(append([]geom.XY{}, pts...), pts[0]))
	rect, ok := MinimumAreaRectangle(g)
	require.True(t, ok)
	a := math.Abs((planar.Kernel{}).RingArea(rect.Ring(0)))
	assert.InDelta(t, 16.0, a, 1e-6)
}

func TestMinimumAreaRectangle_PointDegenerate(t *testing.T) {
	g := geom.NewPoint(nil, geom.XY{X: 1, Y: 2})
	_, ok := MinimumAreaRectangle(g)
	require.False(t, ok, "expected ok=false for point")
}
