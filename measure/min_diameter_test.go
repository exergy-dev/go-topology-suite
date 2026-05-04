package measure

import (
	"math"
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/kernel/planar"
)

func TestMinimumDiameter_Empty(t *testing.T) {
	g := geom.NewEmptyPolygon(nil, geom.LayoutXY)
	if _, _, ok := MinimumDiameter(g); ok {
		t.Fatalf("expected ok=false for empty input")
	}
}

func TestMinimumDiameter_Square(t *testing.T) {
	g := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 4, Y: 4}, {X: 0, Y: 4}, {X: 0, Y: 0},
	})
	_, length, ok := MinimumDiameter(g)
	if !ok {
		t.Fatalf("ok=false")
	}
	if math.Abs(length-4) > 1e-9 {
		t.Fatalf("length=%v want 4", length)
	}
}

func TestMinimumDiameter_Rectangle(t *testing.T) {
	// 6x2 axis-aligned: minimum diameter = 2.
	g := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 6, Y: 0}, {X: 6, Y: 2}, {X: 0, Y: 2}, {X: 0, Y: 0},
	})
	_, length, _ := MinimumDiameter(g)
	if math.Abs(length-2) > 1e-9 {
		t.Fatalf("length=%v want 2", length)
	}
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
	if math.Abs(length-2) > 1e-7 {
		t.Fatalf("length=%v want 2", length)
	}
}

func TestMinimumDiameter_Point(t *testing.T) {
	g := geom.NewPoint(nil, geom.XY{X: 1, Y: 1})
	_, length, ok := MinimumDiameter(g)
	if !ok || length != 0 {
		t.Fatalf("got length=%v ok=%v", length, ok)
	}
}

func TestMinimumDiameterRectangle_Square(t *testing.T) {
	g := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0},
	})
	rect, ok := MinimumDiameterRectangle(g)
	if !ok {
		t.Fatalf("ok=false")
	}
	// Rectangle should have area approximately 1.
	a := math.Abs((planar.Kernel{}).RingArea(rect.Ring(0)))
	if math.Abs(a-1) > 1e-9 {
		t.Fatalf("area=%v want 1", a)
	}
}
