package measure

import (
	"math"
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/kernel/planar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMinimumDiameter_Empty(t *testing.T) {
	g := geom.NewEmptyPolygon(nil, geom.LayoutXY)
	_, _, ok := MinimumDiameter(g)
	require.False(t, ok, "expected ok=false for empty input")
}

func TestMinimumDiameter_Square(t *testing.T) {
	g := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 4, Y: 4}, {X: 0, Y: 4}, {X: 0, Y: 0},
	})
	_, length, ok := MinimumDiameter(g)
	require.True(t, ok)
	assert.InDelta(t, 4.0, length, 1e-9)
}

func TestMinimumDiameter_Rectangle(t *testing.T) {
	// 6x2 axis-aligned: minimum diameter = 2.
	g := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 6, Y: 0}, {X: 6, Y: 2}, {X: 0, Y: 2}, {X: 0, Y: 0},
	})
	_, length, _ := MinimumDiameter(g)
	assert.InDelta(t, 2.0, length, 1e-9)
}

func TestMinimumDiameter_RotatedRectangle(t *testing.T) {
	// 6x2 rotated 30°: same minimum diameter of 2.
	pts := []geom.XY{{X: 0, Y: 0}, {X: 6, Y: 0}, {X: 6, Y: 2}, {X: 0, Y: 2}}
	theta := math.Pi / 6
	cos, sin := math.Cos(theta), math.Sin(theta)
	for i := range pts {
		x, y := pts[i].X, pts[i].Y
		pts[i] = geom.XY{X: x*cos - y*sin, Y: x*sin + y*cos}
	}
	g := geom.NewPolygon(nil, append(append([]geom.XY{}, pts...), pts[0]))
	_, length, _ := MinimumDiameter(g)
	assert.InDelta(t, 2.0, length, 1e-7)
}

func TestMinimumDiameter_Point(t *testing.T) {
	g := geom.NewPoint(nil, geom.XY{X: 1, Y: 1})
	_, length, ok := MinimumDiameter(g)
	require.True(t, ok)
	assert.Equal(t, 0.0, length)
}

func TestMinimumDiameterRectangle_Square(t *testing.T) {
	g := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0},
	})
	rect, ok := MinimumDiameterRectangle(g)
	require.True(t, ok)
	// Rectangle should have area approximately 1.
	a := math.Abs((planar.Kernel{}).RingArea(rect.Ring(0)))
	assert.InDelta(t, 1.0, a, 1e-9)
}
