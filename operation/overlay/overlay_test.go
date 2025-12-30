package overlay

import (
	"testing"

	"github.com/go-topology-suite/gts/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	// Expected: 5x5 = 25 square units
	expectedArea := 25.0
	// JTS-compatible 1% tolerance for overlay operations
	tolerance := expectedArea * 0.01

	assert.InDelta(t, expectedArea, intersectionArea, tolerance, "Intersection area")
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

	// Expected: 100 + 100 - 25 = 175 square units
	expectedArea := 175.0
	// JTS-compatible 1% tolerance for overlay operations
	tolerance := expectedArea * 0.01

	assert.InDelta(t, expectedArea, unionArea, tolerance, "Union area")
}

func TestIntersectionPointPoint(t *testing.T) {
	p1 := geom.NewPoint(0, 0)
	p2 := geom.NewPoint(0, 0)

	result := Intersection(p1, p2)
	assert.False(t, result.IsEmpty(), "Intersection of same points should not be empty")
}

func TestIntersectionPointPointDisjoint(t *testing.T) {
	p1 := geom.NewPoint(0, 0)
	p2 := geom.NewPoint(10, 10)

	result := Intersection(p1, p2)
	assert.True(t, result.IsEmpty(), "Intersection of disjoint points should be empty")
}

func TestIntersectionPointLine(t *testing.T) {
	p := geom.NewPoint(5, 0)
	ls := geom.NewLineStringXY(0, 0, 10, 0)

	result := Intersection(p, ls)
	assert.False(t, result.IsEmpty(), "Point on line should intersect")

	// Point off line
	p2 := geom.NewPoint(5, 5)
	result2 := Intersection(p2, ls)
	assert.True(t, result2.IsEmpty(), "Point off line should not intersect")
}

func TestIntersectionPointPolygon(t *testing.T) {
	p := geom.NewPoint(5, 5)
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)

	result := Intersection(p, poly)
	assert.False(t, result.IsEmpty(), "Point inside polygon should intersect")

	// Point outside polygon
	p2 := geom.NewPoint(15, 15)
	result2 := Intersection(p2, poly)
	assert.True(t, result2.IsEmpty(), "Point outside polygon should not intersect")
}

func TestIntersectionLineLine(t *testing.T) {
	ls1 := geom.NewLineStringXY(0, 0, 10, 10)
	ls2 := geom.NewLineStringXY(0, 10, 10, 0)

	result := Intersection(ls1, ls2)
	assert.False(t, result.IsEmpty(), "Crossing lines should have intersection")
}

func TestIntersectionLinePolygon(t *testing.T) {
	ls := geom.NewLineStringXY(5, -5, 5, 15) // Vertical line through polygon
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)

	result := Intersection(ls, poly)

	// A vertical line at x=5 through a 10x10 polygon should produce a line segment
	// from (5, 0) to (5, 10), with length 10
	require.False(t, result.IsEmpty(), "Line through polygon should have non-empty intersection")

	// The result should be a LineString
	if resultLine, ok := result.(*geom.LineString); ok {
		expectedLength := 10.0
		// Calculate actual length
		coords := resultLine.Coordinates()
		if len(coords) >= 2 {
			actualLength := coords[0].Distance(coords[len(coords)-1])
			tolerance := expectedLength * 0.01 // 1% tolerance
			assert.InDelta(t, expectedLength, actualLength, tolerance, "Line-polygon intersection length")
		}
	}
}

func TestIntersectionPolygonPolygon(t *testing.T) {
	shell1 := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := geom.NewLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	poly2 := geom.NewPolygon(shell2, nil)

	result := Intersection(poly1, poly2)
	assert.False(t, result.IsEmpty(), "Overlapping polygons should have intersection")
}

func TestIntersectionDisjointPolygons(t *testing.T) {
	shell1 := geom.NewLinearRingXY(0, 0, 5, 0, 5, 5, 0, 5, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := geom.NewLinearRingXY(10, 10, 15, 10, 15, 15, 10, 15, 10, 10)
	poly2 := geom.NewPolygon(shell2, nil)

	result := Intersection(poly1, poly2)
	assert.True(t, result.IsEmpty(), "Disjoint polygons should have empty intersection")
}

func TestUnionPointPoint(t *testing.T) {
	p1 := geom.NewPoint(0, 0)
	p2 := geom.NewPoint(10, 10)

	result := Union(p1, p2)
	assert.False(t, result.IsEmpty(), "Union of points should not be empty")
}

func TestUnionEmptyGeometries(t *testing.T) {
	p := geom.NewPoint(5, 5)
	empty := geom.NewPointEmpty()

	result := Union(p, empty)
	assert.False(t, result.IsEmpty(), "Union with empty should return non-empty")

	result2 := Union(empty, p)
	assert.False(t, result2.IsEmpty(), "Union with empty should return non-empty")

	result3 := Union(empty, empty)
	assert.True(t, result3.IsEmpty(), "Union of empty geometries should be empty")
}

func TestUnionPolygonPolygon(t *testing.T) {
	shell1 := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := geom.NewLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	poly2 := geom.NewPolygon(shell2, nil)

	result := Union(poly1, poly2)
	assert.False(t, result.IsEmpty(), "Union of polygons should not be empty")
}

func TestDifferencePointPolygon(t *testing.T) {
	// Point inside polygon
	p := geom.NewPoint(5, 5)
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)

	result := Difference(p, poly)
	assert.True(t, result.IsEmpty(), "Point inside polygon minus polygon should be empty")

	// Point outside polygon
	p2 := geom.NewPoint(15, 15)
	result2 := Difference(p2, poly)
	assert.False(t, result2.IsEmpty(), "Point outside polygon minus polygon should be the point")
}

func TestDifferencePolygonPoint(t *testing.T) {
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)
	p := geom.NewPoint(5, 5)

	result := Difference(poly, p)
	assert.False(t, result.IsEmpty(), "Polygon minus point should be polygon")
}

func TestDifferencePolygonPolygon(t *testing.T) {
	shell1 := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := geom.NewLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	poly2 := geom.NewPolygon(shell2, nil)

	result := Difference(poly1, poly2)
	assert.False(t, result.IsEmpty(), "Partial overlap difference should not be empty")
}

func TestDifferenceDisjointPolygons(t *testing.T) {
	shell1 := geom.NewLinearRingXY(0, 0, 5, 0, 5, 5, 0, 5, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := geom.NewLinearRingXY(10, 10, 15, 10, 15, 15, 10, 15, 10, 10)
	poly2 := geom.NewPolygon(shell2, nil)

	result := Difference(poly1, poly2)
	assert.False(t, result.IsEmpty(), "Difference of disjoint polygons should be first polygon")
}

func TestSymDifferencePointPoint(t *testing.T) {
	p1 := geom.NewPoint(0, 0)
	p2 := geom.NewPoint(10, 10)

	result := SymDifference(p1, p2)
	assert.False(t, result.IsEmpty(), "SymDifference of different points should not be empty")
}

func TestSymDifferencePolygonPolygon(t *testing.T) {
	shell1 := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := geom.NewLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	poly2 := geom.NewPolygon(shell2, nil)

	result := SymDifference(poly1, poly2)
	assert.False(t, result.IsEmpty(), "SymDifference of overlapping polygons should not be empty")
}

func TestOverlayNilGeometries(t *testing.T) {
	p := geom.NewPoint(0, 0)

	result := Intersection(nil, p)
	assert.True(t, result.IsEmpty(), "Intersection with nil should be empty")

	result = Intersection(p, nil)
	assert.True(t, result.IsEmpty(), "Intersection with nil should be empty")

	result = Union(nil, p)
	assert.False(t, result.IsEmpty(), "Union with nil should return other geometry")

	result = Union(p, nil)
	assert.False(t, result.IsEmpty(), "Union with nil should return other geometry")
}

func TestExtractPoints(t *testing.T) {
	p := geom.NewPoint(1, 2)
	points := extractPoints(p)
	assert.Len(t, points, 1, "Expected 1 point")

	mp := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(0, 0),
		geom.NewPoint(1, 1),
	})
	points = extractPoints(mp)
	assert.Len(t, points, 2, "Expected 2 points")
}

func TestExtractLineStrings(t *testing.T) {
	ls := geom.NewLineStringXY(0, 0, 10, 10)
	lines := extractLineStrings(ls)
	assert.Len(t, lines, 1, "Expected 1 line")

	mls := geom.NewMultiLineString([]*geom.LineString{
		geom.NewLineStringXY(0, 0, 10, 0),
		geom.NewLineStringXY(0, 10, 10, 10),
	})
	lines = extractLineStrings(mls)
	assert.Len(t, lines, 2, "Expected 2 lines")
}

func TestExtractPolygons(t *testing.T) {
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)
	polygons := extractPolygons(poly)
	assert.Len(t, polygons, 1, "Expected 1 polygon")
}

func TestCollectGeometries(t *testing.T) {
	p1 := geom.NewPoint(0, 0)
	p2 := geom.NewPoint(10, 10)

	result := collectGeometries(p1, p2)
	gc, ok := result.(*geom.GeometryCollection)
	require.True(t, ok, "Expected GeometryCollection, got %T", result)
	assert.Equal(t, 2, gc.NumGeometries(), "Expected 2 geometries")
}

func TestCollectGeometriesSingle(t *testing.T) {
	p := geom.NewPoint(0, 0)
	empty := geom.NewPointEmpty()

	result := collectGeometries(p, empty)
	// Should return just p, not a collection
	_, ok := result.(*geom.Point)
	assert.True(t, ok, "Expected Point, got %T", result)
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
	assert.False(t, result.IsEmpty(), "Expected non-empty intersection")
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
