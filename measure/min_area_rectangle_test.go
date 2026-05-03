package measure

import (
	"math"
	"testing"

	"github.com/terra-geo/terra/geom"
)

func TestMinimumAreaRectangle_Empty(t *testing.T) {
	g := geom.NewEmptyPolygon(nil, geom.LayoutXY)
	if _, ok := MinimumAreaRectangle(g); ok {
		t.Fatalf("expected ok=false")
	}
}

func TestMinimumAreaRectangle_Square(t *testing.T) {
	g := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 4, Y: 4}, {X: 0, Y: 4}, {X: 0, Y: 0},
	})
	rect, ok := MinimumAreaRectangle(g)
	if !ok {
		t.Fatalf("ok=false")
	}
	a := math.Abs(ringSignedArea(rect.Ring(0)))
	if math.Abs(a-16) > 1e-9 {
		t.Fatalf("area=%v want 16", a)
	}
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
	if !ok {
		t.Fatalf("ok=false")
	}
	a := math.Abs(ringSignedArea(rect.Ring(0)))
	if math.Abs(a-16) > 1e-7 {
		t.Fatalf("area=%v want 16", a)
	}
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
	if !ok {
		t.Fatalf("ok=false")
	}
	a := math.Abs(ringSignedArea(rect.Ring(0)))
	if math.Abs(a-16) > 1e-6 {
		t.Fatalf("area=%v want 16", a)
	}
}

func TestMinimumAreaRectangle_PointDegenerate(t *testing.T) {
	g := geom.NewPoint(nil, geom.XY{X: 1, Y: 2})
	if _, ok := MinimumAreaRectangle(g); ok {
		t.Fatalf("expected ok=false for point")
	}
}
