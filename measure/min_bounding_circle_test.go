package measure

import (
	"math"
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
)

func TestMinimumBoundingCircle_Empty(t *testing.T) {
	g := geom.NewEmptyPoint(nil, geom.LayoutXY)
	if _, _, ok := MinimumBoundingCircle(g); ok {
		t.Fatalf("expected empty input → ok=false")
	}
}

func TestMinimumBoundingCircle_SinglePoint(t *testing.T) {
	g := geom.NewPoint(nil, geom.XY{X: 3, Y: 4})
	c, r, ok := MinimumBoundingCircle(g)
	if !ok || r != 0 || c.X != 3 || c.Y != 4 {
		t.Fatalf("got c=%+v r=%v ok=%v", c, r, ok)
	}
}

func TestMinimumBoundingCircle_TwoPoints(t *testing.T) {
	g := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 4, Y: 0}})
	c, r, ok := MinimumBoundingCircle(g)
	if !ok {
		t.Fatalf("ok=false")
	}
	if math.Abs(c.X-2) > 1e-9 || math.Abs(c.Y) > 1e-9 || math.Abs(r-2) > 1e-9 {
		t.Fatalf("got c=%+v r=%v", c, r)
	}
}

func TestMinimumBoundingCircle_Square(t *testing.T) {
	// Unit square; MBC should be circumscribed circle.
	g := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0},
	})
	c, r, ok := MinimumBoundingCircle(g)
	if !ok {
		t.Fatalf("ok=false")
	}
	if math.Abs(c.X-0.5) > 1e-9 || math.Abs(c.Y-0.5) > 1e-9 {
		t.Fatalf("centre=%+v", c)
	}
	if math.Abs(r-math.Sqrt2/2) > 1e-9 {
		t.Fatalf("radius=%v want %v", r, math.Sqrt2/2)
	}
	// All vertices must lie on or inside the circle.
	for _, p := range []geom.XY{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}} {
		d := math.Hypot(p.X-c.X, p.Y-c.Y)
		if d > r+1e-9 {
			t.Fatalf("vertex %+v at distance %v > r=%v", p, d, r)
		}
	}
}

func TestMinimumBoundingCircle_Triangle(t *testing.T) {
	// Acute triangle: circumscribed circle should pass through all three.
	pts := []geom.XY{{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 2, Y: 3}}
	g := geom.NewPolygon(nil, append(append([]geom.XY{}, pts...), pts[0]))
	c, r, ok := MinimumBoundingCircle(g)
	if !ok {
		t.Fatalf("ok=false")
	}
	for _, p := range pts {
		d := math.Hypot(p.X-c.X, p.Y-c.Y)
		if math.Abs(d-r) > 1e-7 {
			t.Fatalf("vertex %+v dist=%v want r=%v", p, d, r)
		}
	}
}

func TestMinimumBoundingCircle_ObtuseTriangle(t *testing.T) {
	// Obtuse triangle: MBC determined by the two endpoints of the longest side.
	pts := []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 1, Y: 1}}
	g := geom.NewPolygon(nil, append(append([]geom.XY{}, pts...), pts[0]))
	c, r, ok := MinimumBoundingCircle(g)
	if !ok {
		t.Fatalf("ok=false")
	}
	// Longest side is (0,0)-(10,0), so MBC is centred at (5,0) with r=5.
	if math.Abs(c.X-5) > 1e-9 || math.Abs(c.Y) > 1e-9 || math.Abs(r-5) > 1e-9 {
		t.Fatalf("c=%+v r=%v", c, r)
	}
	for _, p := range pts {
		d := math.Hypot(p.X-c.X, p.Y-c.Y)
		if d > r+1e-9 {
			t.Fatalf("vertex %+v dist=%v exceeds r=%v", p, d, r)
		}
	}
}
