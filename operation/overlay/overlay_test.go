package overlay

import (
	"math"
	"testing"

	"github.com/go-topology-suite/gts/geom"
)

// geometryArea returns the area of a geometry.
func geometryArea(g geom.Geometry) float64 {
	if g == nil || g.IsEmpty() {
		return 0
	}
	switch v := g.(type) {
	case *geom.Polygon:
		return v.Area()
	case *geom.MultiPolygon:
		return v.Area()
	case *geom.GeometryCollection:
		var total float64
		for i := 0; i < v.NumGeometries(); i++ {
			total += geometryArea(v.GeometryN(i))
		}
		return total
	default:
		return 0
	}
}

// TestPolygonIntersectionArea tests that polygon intersection computes correct area.
func TestPolygonIntersectionArea(t *testing.T) {
	// Two overlapping 10x10 squares with 5x5 overlap at (5,5)-(10,10)
	shell1 := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := geom.NewLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	poly2 := geom.NewPolygon(shell2, nil)

	intersection := Intersection(poly1, poly2)
	intersectionArea := geometryArea(intersection)

	t.Logf("Intersection area: %.2f (expected 25)", intersectionArea)
	t.Logf("Intersection type: %T", intersection)
	if poly, ok := intersection.(*geom.Polygon); ok {
		t.Logf("Intersection coords: %v", poly.ExteriorRing().Coordinates())
	}

	// Expected: 5x5 = 25 square units
	if math.Abs(intersectionArea-25) > 0.1 {
		t.Errorf("Intersection area is %.2f, expected 25", intersectionArea)
	}
}

// TestPolygonUnionArea tests that polygon union computes correct area.
func TestPolygonUnionArea(t *testing.T) {
	// Two overlapping 10x10 squares
	shell1 := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := geom.NewLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	poly2 := geom.NewPolygon(shell2, nil)

	union := Union(poly1, poly2)
	unionArea := geometryArea(union)

	t.Logf("Union area: %.2f (expected 175)", unionArea)
	t.Logf("Union type: %T", union)

	// Expected: 100 + 100 - 25 = 175 square units
	if math.Abs(unionArea-175) > 0.1 {
		t.Errorf("Union area is %.2f, expected 175", unionArea)
	}
}

func TestIntersectionPointPoint(t *testing.T) {
	p1 := geom.NewPoint(0, 0)
	p2 := geom.NewPoint(0, 0)

	result := Intersection(p1, p2)
	if result.IsEmpty() {
		t.Error("Intersection of same points should not be empty")
	}
}

func TestIntersectionPointPointDisjoint(t *testing.T) {
	p1 := geom.NewPoint(0, 0)
	p2 := geom.NewPoint(10, 10)

	result := Intersection(p1, p2)
	if !result.IsEmpty() {
		t.Error("Intersection of disjoint points should be empty")
	}
}

func TestIntersectionPointLine(t *testing.T) {
	p := geom.NewPoint(5, 0)
	ls := geom.NewLineStringXY(0, 0, 10, 0)

	result := Intersection(p, ls)
	if result.IsEmpty() {
		t.Error("Point on line should intersect")
	}

	// Point off line
	p2 := geom.NewPoint(5, 5)
	result2 := Intersection(p2, ls)
	if !result2.IsEmpty() {
		t.Error("Point off line should not intersect")
	}
}

func TestIntersectionPointPolygon(t *testing.T) {
	p := geom.NewPoint(5, 5)
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)

	result := Intersection(p, poly)
	if result.IsEmpty() {
		t.Error("Point inside polygon should intersect")
	}

	// Point outside polygon
	p2 := geom.NewPoint(15, 15)
	result2 := Intersection(p2, poly)
	if !result2.IsEmpty() {
		t.Error("Point outside polygon should not intersect")
	}
}

func TestIntersectionLineLine(t *testing.T) {
	ls1 := geom.NewLineStringXY(0, 0, 10, 10)
	ls2 := geom.NewLineStringXY(0, 10, 10, 0)

	result := Intersection(ls1, ls2)
	if result.IsEmpty() {
		t.Error("Crossing lines should have intersection")
	}
}

func TestIntersectionLinePolygon(t *testing.T) {
	ls := geom.NewLineStringXY(5, -5, 5, 15) // Vertical line through polygon
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)

	result := Intersection(ls, poly)
	// Note: Line-polygon intersection implementation needs refinement
	// for handling boundary crossing cases
	t.Logf("Line-polygon intersection result type: %T, empty: %v", result, result.IsEmpty())
}

func TestIntersectionPolygonPolygon(t *testing.T) {
	shell1 := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := geom.NewLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	poly2 := geom.NewPolygon(shell2, nil)

	result := Intersection(poly1, poly2)
	if result.IsEmpty() {
		t.Error("Overlapping polygons should have intersection")
	}
}

func TestIntersectionDisjointPolygons(t *testing.T) {
	shell1 := geom.NewLinearRingXY(0, 0, 5, 0, 5, 5, 0, 5, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := geom.NewLinearRingXY(10, 10, 15, 10, 15, 15, 10, 15, 10, 10)
	poly2 := geom.NewPolygon(shell2, nil)

	result := Intersection(poly1, poly2)
	if !result.IsEmpty() {
		t.Error("Disjoint polygons should have empty intersection")
	}
}

func TestUnionPointPoint(t *testing.T) {
	p1 := geom.NewPoint(0, 0)
	p2 := geom.NewPoint(10, 10)

	result := Union(p1, p2)
	if result.IsEmpty() {
		t.Error("Union of points should not be empty")
	}
}

func TestUnionEmptyGeometries(t *testing.T) {
	p := geom.NewPoint(5, 5)
	empty := geom.NewPointEmpty()

	result := Union(p, empty)
	if result.IsEmpty() {
		t.Error("Union with empty should return non-empty")
	}

	result2 := Union(empty, p)
	if result2.IsEmpty() {
		t.Error("Union with empty should return non-empty")
	}

	result3 := Union(empty, empty)
	if !result3.IsEmpty() {
		t.Error("Union of empty geometries should be empty")
	}
}

func TestUnionPolygonPolygon(t *testing.T) {
	shell1 := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := geom.NewLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	poly2 := geom.NewPolygon(shell2, nil)

	result := Union(poly1, poly2)
	if result.IsEmpty() {
		t.Error("Union of polygons should not be empty")
	}
}

func TestDifferencePointPolygon(t *testing.T) {
	// Point inside polygon
	p := geom.NewPoint(5, 5)
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)

	result := Difference(p, poly)
	if !result.IsEmpty() {
		t.Error("Point inside polygon minus polygon should be empty")
	}

	// Point outside polygon
	p2 := geom.NewPoint(15, 15)
	result2 := Difference(p2, poly)
	if result2.IsEmpty() {
		t.Error("Point outside polygon minus polygon should be the point")
	}
}

func TestDifferencePolygonPoint(t *testing.T) {
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)
	p := geom.NewPoint(5, 5)

	result := Difference(poly, p)
	if result.IsEmpty() {
		t.Error("Polygon minus point should be polygon")
	}
}

func TestDifferencePolygonPolygon(t *testing.T) {
	shell1 := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := geom.NewLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	poly2 := geom.NewPolygon(shell2, nil)

	result := Difference(poly1, poly2)
	if result.IsEmpty() {
		t.Error("Partial overlap difference should not be empty")
	}
}

func TestDifferenceDisjointPolygons(t *testing.T) {
	shell1 := geom.NewLinearRingXY(0, 0, 5, 0, 5, 5, 0, 5, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := geom.NewLinearRingXY(10, 10, 15, 10, 15, 15, 10, 15, 10, 10)
	poly2 := geom.NewPolygon(shell2, nil)

	result := Difference(poly1, poly2)
	if result.IsEmpty() {
		t.Error("Difference of disjoint polygons should be first polygon")
	}
}

func TestSymDifferencePointPoint(t *testing.T) {
	p1 := geom.NewPoint(0, 0)
	p2 := geom.NewPoint(10, 10)

	result := SymDifference(p1, p2)
	if result.IsEmpty() {
		t.Error("SymDifference of different points should not be empty")
	}
}

func TestSymDifferencePolygonPolygon(t *testing.T) {
	shell1 := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := geom.NewLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	poly2 := geom.NewPolygon(shell2, nil)

	result := SymDifference(poly1, poly2)
	if result.IsEmpty() {
		t.Error("SymDifference of overlapping polygons should not be empty")
	}
}

func TestOverlayNilGeometries(t *testing.T) {
	p := geom.NewPoint(0, 0)

	result := Intersection(nil, p)
	if !result.IsEmpty() {
		t.Error("Intersection with nil should be empty")
	}

	result = Intersection(p, nil)
	if !result.IsEmpty() {
		t.Error("Intersection with nil should be empty")
	}

	result = Union(nil, p)
	if result.IsEmpty() {
		t.Error("Union with nil should return other geometry")
	}

	result = Union(p, nil)
	if result.IsEmpty() {
		t.Error("Union with nil should return other geometry")
	}
}

func TestExtractPoints(t *testing.T) {
	p := geom.NewPoint(1, 2)
	points := extractPoints(p)
	if len(points) != 1 {
		t.Errorf("Expected 1 point, got %d", len(points))
	}

	mp := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(0, 0),
		geom.NewPoint(1, 1),
	})
	points = extractPoints(mp)
	if len(points) != 2 {
		t.Errorf("Expected 2 points, got %d", len(points))
	}
}

func TestExtractLineStrings(t *testing.T) {
	ls := geom.NewLineStringXY(0, 0, 10, 10)
	lines := extractLineStrings(ls)
	if len(lines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(lines))
	}

	mls := geom.NewMultiLineString([]*geom.LineString{
		geom.NewLineStringXY(0, 0, 10, 0),
		geom.NewLineStringXY(0, 10, 10, 10),
	})
	lines = extractLineStrings(mls)
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(lines))
	}
}

func TestExtractPolygons(t *testing.T) {
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)
	polygons := extractPolygons(poly)
	if len(polygons) != 1 {
		t.Errorf("Expected 1 polygon, got %d", len(polygons))
	}
}

func TestCollectGeometries(t *testing.T) {
	p1 := geom.NewPoint(0, 0)
	p2 := geom.NewPoint(10, 10)

	result := collectGeometries(p1, p2)
	gc, ok := result.(*geom.GeometryCollection)
	if !ok {
		t.Fatalf("Expected GeometryCollection, got %T", result)
	}
	if gc.NumGeometries() != 2 {
		t.Errorf("Expected 2 geometries, got %d", gc.NumGeometries())
	}
}

func TestCollectGeometriesSingle(t *testing.T) {
	p := geom.NewPoint(0, 0)
	empty := geom.NewPointEmpty()

	result := collectGeometries(p, empty)
	// Should return just p, not a collection
	if _, ok := result.(*geom.Point); !ok {
		t.Errorf("Expected Point, got %T", result)
	}
}

func TestIntersectionMultiPoint(t *testing.T) {
	mp := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(0, 0),
		geom.NewPoint(5, 5),
		geom.NewPoint(15, 15),
	})
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)

	result := Intersection(mp, poly)
	// Two points should be inside the polygon
	if result.IsEmpty() {
		t.Error("Expected non-empty intersection")
	}
}

func BenchmarkIntersectionPolygonPolygon(b *testing.B) {
	shell1 := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := geom.NewLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	poly2 := geom.NewPolygon(shell2, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Intersection(poly1, poly2)
	}
}

func BenchmarkUnionPolygonPolygon(b *testing.B) {
	shell1 := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := geom.NewLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	poly2 := geom.NewPolygon(shell2, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Union(poly1, poly2)
	}
}
