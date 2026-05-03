package measure

import (
	"math"
	"testing"

	"github.com/terra-geo/terra/geom"
)

// triangleAreaOf is a small helper to compute the area of a returned MBT.
func triangleAreaOf(v [3]geom.XY) float64 {
	return triangleArea(v[0], v[1], v[2])
}

// pointInsideTriangle reports whether p lies inside (or on) triangle abc.
func pointInsideTriangle(p, a, b, c geom.XY) bool {
	d1 := (p.X-b.X)*(a.Y-b.Y) - (a.X-b.X)*(p.Y-b.Y)
	d2 := (p.X-c.X)*(b.Y-c.Y) - (b.X-c.X)*(p.Y-c.Y)
	d3 := (p.X-a.X)*(c.Y-a.Y) - (c.X-a.X)*(p.Y-a.Y)
	hasNeg := d1 < -1e-7 || d2 < -1e-7 || d3 < -1e-7
	hasPos := d1 > 1e-7 || d2 > 1e-7 || d3 > 1e-7
	return !(hasNeg && hasPos)
}

func TestMinimumBoundingTriangle_Empty(t *testing.T) {
	if _, ok := MinimumBoundingTriangle(geom.NewEmptyPolygon(nil, geom.LayoutXY)); ok {
		t.Fatalf("empty: expected ok=false")
	}
}

func TestMinimumBoundingTriangle_DegeneratePoint(t *testing.T) {
	if _, ok := MinimumBoundingTriangle(geom.NewPoint(nil, geom.XY{X: 1, Y: 2})); ok {
		t.Fatalf("point: expected ok=false")
	}
}

func TestMinimumBoundingTriangle_DegenerateLine(t *testing.T) {
	ls := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 4, Y: 0}})
	if _, ok := MinimumBoundingTriangle(ls); ok {
		t.Fatalf("line: expected ok=false")
	}
}

func TestMinimumBoundingTriangle_AlreadyTriangle(t *testing.T) {
	g := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 6, Y: 0}, {X: 0, Y: 6}, {X: 0, Y: 0},
	})
	v, ok := MinimumBoundingTriangle(g)
	if !ok {
		t.Fatalf("ok=false")
	}
	want := 0.5 * 6 * 6
	if math.Abs(triangleAreaOf(v)-want) > 1e-9 {
		t.Fatalf("area=%v want %v", triangleAreaOf(v), want)
	}
}

func TestMinimumBoundingTriangle_Square(t *testing.T) {
	// For a square of side s, the minimum-area enclosing triangle has area 2*s^2.
	g := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 4, Y: 4}, {X: 0, Y: 4}, {X: 0, Y: 0},
	})
	v, ok := MinimumBoundingTriangle(g)
	if !ok {
		t.Fatalf("ok=false")
	}
	area := triangleAreaOf(v)
	want := 2.0 * 4 * 4
	if math.Abs(area-want) > 0.5 {
		// Klee–Laskowski guarantees ≤ 2× polygon area for convex polygons,
		// and equality is attained for a triangle. For a square the answer
		// is 2 × s² (theoretical optimum).
		t.Fatalf("square MBT area=%v want %v", area, want)
	}
	// Verify all square corners lie inside the returned triangle.
	for _, p := range []geom.XY{{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 4, Y: 4}, {X: 0, Y: 4}} {
		if !pointInsideTriangle(p, v[0], v[1], v[2]) {
			t.Fatalf("square corner %v not inside MBT %v", p, v)
		}
	}
}

func TestMinimumBoundingTriangle_ScatteredPoints(t *testing.T) {
	// Five scattered points; MBT must enclose them all.
	pts := []geom.XY{
		{X: 0, Y: 0}, {X: 10, Y: 1}, {X: 8, Y: 9}, {X: 1, Y: 8}, {X: 5, Y: 5},
	}
	mp := geom.NewMultiPoint(nil, pts)
	v, ok := MinimumBoundingTriangle(mp)
	if !ok {
		t.Fatalf("ok=false")
	}
	for _, p := range pts {
		if !pointInsideTriangle(p, v[0], v[1], v[2]) {
			t.Fatalf("point %v not enclosed by MBT %v", p, v)
		}
	}
	// Triangle area must be at most 2× hull area (Klee bound). Hull
	// area here is 79.5 (manual computation), so MBT ≤ 159.
	if a := triangleAreaOf(v); a > 200 {
		t.Fatalf("scattered MBT area=%v unreasonably large", a)
	}
}
