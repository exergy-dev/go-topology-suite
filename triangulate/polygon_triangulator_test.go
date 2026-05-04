package triangulate

import (
	"math"
	"testing"

	"github.com/terra-geo/terra/geom"
)

// triangleArea returns the unsigned area of triangle t.
func triangleArea(t Triangle) float64 {
	return math.Abs(geom.TriangleSignedArea(t.P0, t.P1, t.P2))
}

func sumTriangleArea(tris []Triangle) float64 {
	var s float64
	for _, t := range tris {
		s += triangleArea(t)
	}
	return s
}

func TestTriangulatePolygon_Square(t *testing.T) {
	// Convex square — 2 triangles, total area = 1.
	shell := []geom.XY{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0}}
	p := geom.NewPolygon(nil, shell)
	tris := TriangulatePolygon(p)
	if len(tris) != 2 {
		t.Fatalf("want 2 triangles, got %d", len(tris))
	}
	if math.Abs(sumTriangleArea(tris)-1.0) > 1e-9 {
		t.Fatalf("area: want 1, got %v", sumTriangleArea(tris))
	}
}

func TestTriangulatePolygon_Pentagon(t *testing.T) {
	// Convex regular pentagon; 5 vertices -> 3 triangles.
	var shell []geom.XY
	for i := 0; i < 5; i++ {
		theta := 2 * math.Pi * float64(i) / 5
		shell = append(shell, geom.XY{X: math.Cos(theta), Y: math.Sin(theta)})
	}
	shell = append(shell, shell[0])
	p := geom.NewPolygon(nil, shell)
	tris := TriangulatePolygon(p)
	if len(tris) != 3 {
		t.Fatalf("want 3 triangles, got %d", len(tris))
	}
	// Pentagon area = (5/2) * sin(2π/5) for unit-circle inscribed.
	expected := 2.5 * math.Sin(2*math.Pi/5)
	if math.Abs(sumTriangleArea(tris)-expected) > 1e-9 {
		t.Fatalf("area: want %v, got %v", expected, sumTriangleArea(tris))
	}
}

func TestTriangulatePolygon_Concave(t *testing.T) {
	// Concave "L" shape: 6 vertices -> 4 triangles, area = 3.
	shell := []geom.XY{
		{X: 0, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 1}, {X: 1, Y: 1}, {X: 1, Y: 2}, {X: 0, Y: 2}, {X: 0, Y: 0},
	}
	p := geom.NewPolygon(nil, shell)
	tris := TriangulatePolygon(p)
	if len(tris) != 4 {
		t.Fatalf("want 4 triangles, got %d", len(tris))
	}
	if math.Abs(sumTriangleArea(tris)-3.0) > 1e-9 {
		t.Fatalf("area: want 3, got %v", sumTriangleArea(tris))
	}
}

func TestTriangulatePolygon_OneHole(t *testing.T) {
	// 4x4 square with a 1x1 hole — area should be 16 - 1 = 15.
	shell := []geom.XY{{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 4, Y: 4}, {X: 0, Y: 4}, {X: 0, Y: 0}}
	hole := []geom.XY{{X: 1, Y: 1}, {X: 2, Y: 1}, {X: 2, Y: 2}, {X: 1, Y: 2}, {X: 1, Y: 1}}
	p := geom.NewPolygon(nil, shell, hole)
	tris := TriangulatePolygon(p)
	if len(tris) == 0 {
		t.Fatal("got 0 triangles")
	}
	got := sumTriangleArea(tris)
	if math.Abs(got-15.0) > 1e-7 {
		t.Fatalf("area: want 15, got %v (n=%d)", got, len(tris))
	}
}

func TestTriangulatePolygon_TwoHoles(t *testing.T) {
	// 10x10 square with two 1x1 holes — area = 100 - 2 = 98.
	shell := []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0}}
	h1 := []geom.XY{{X: 1, Y: 1}, {X: 2, Y: 1}, {X: 2, Y: 2}, {X: 1, Y: 2}, {X: 1, Y: 1}}
	h2 := []geom.XY{{X: 6, Y: 6}, {X: 7, Y: 6}, {X: 7, Y: 7}, {X: 6, Y: 7}, {X: 6, Y: 6}}
	p := geom.NewPolygon(nil, shell, h1, h2)
	tris := TriangulatePolygon(p)
	if len(tris) == 0 {
		t.Fatal("got 0 triangles")
	}
	got := sumTriangleArea(tris)
	if math.Abs(got-98.0) > 1e-6 {
		t.Fatalf("area: want 98, got %v", got)
	}
}

func TestTriangulatePolygon_Empty(t *testing.T) {
	p := geom.NewEmptyPolygon(nil, geom.LayoutXY)
	if got := TriangulatePolygon(p); got != nil {
		t.Fatalf("want nil, got %v", got)
	}
	if got := TriangulatePolygon(nil); got != nil {
		t.Fatalf("want nil for nil input")
	}
}

func TestTriangulatePolygons_MultiPolygon(t *testing.T) {
	p1 := geom.NewPolygon(nil, []geom.XY{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0}})
	p2 := geom.NewPolygon(nil, []geom.XY{{X: 2, Y: 2}, {X: 3, Y: 2}, {X: 3, Y: 3}, {X: 2, Y: 3}, {X: 2, Y: 2}})
	mp := geom.NewMultiPolygon(nil, p1, p2)
	tris := TriangulatePolygons(mp)
	if len(tris) != 4 {
		t.Fatalf("want 4 triangles, got %d", len(tris))
	}
	if math.Abs(sumTriangleArea(tris)-2.0) > 1e-9 {
		t.Fatalf("area: want 2, got %v", sumTriangleArea(tris))
	}
}
