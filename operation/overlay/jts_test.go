package overlay

import (
	"fmt"
	"math"
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/io/wkt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// JTS-style test cases for overlay operations
// These tests are ported from Java Topology Suite to verify correctness
// against known input/output pairs

// TestJTS_PolygonIntersection_OverlappingSquares tests intersection of two overlapping squares.
// Expected: 5x5 square at the overlap region
func TestJTS_PolygonIntersection_OverlappingSquares(t *testing.T) {
	// Two 10x10 squares overlapping by 5x5
	poly1, _ := wkt.UnmarshalString("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	poly2, _ := wkt.UnmarshalString("POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))")

	result := Intersection(poly1, poly2)
	require.False(t, result.IsEmpty(), "Intersection should not be empty")

	area := geometryArea(result)
	expectedArea := 25.0 // 5x5 overlap

	assert.InDelta(t, expectedArea, area, 0.1, "Intersection area")
}

// TestJTS_PolygonIntersection_LShapes tests intersection of two L-shaped polygons.
func TestJTS_PolygonIntersection_LShapes(t *testing.T) {
	// L-shape 1: vertical part (0,0)-(4,10), horizontal part (0,0)-(10,4)
	poly1, _ := wkt.UnmarshalString("POLYGON ((0 0, 4 0, 4 6, 10 6, 10 10, 0 10, 0 0))")
	// L-shape 2: rotated 180 degrees
	poly2, _ := wkt.UnmarshalString("POLYGON ((6 0, 10 0, 10 10, 0 10, 0 6, 6 6, 6 0))")

	result := Intersection(poly1, poly2)
	require.False(t, result.IsEmpty(), "Intersection should not be empty")

	// The intersection should be a complex polygon
	area := geometryArea(result)
	assert.Greater(t, area, 0.0, "Intersection should have positive area")
}

// TestJTS_PolygonIntersection_ComplexPolygon tests intersection with a more complex polygon.
func TestJTS_PolygonIntersection_ComplexPolygon(t *testing.T) {
	// A square
	poly1, _ := wkt.UnmarshalString("POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0))")
	// A triangle overlapping the square
	poly2, _ := wkt.UnmarshalString("POLYGON ((10 -5, 25 10, 10 25, 10 -5))")

	result := Intersection(poly1, poly2)
	require.False(t, result.IsEmpty(), "Intersection should not be empty")

	area := geometryArea(result)
	assert.Greater(t, area, 0.0, "Intersection should have positive area")
}

// TestJTS_PolygonIntersection_TouchingPolygons tests polygons that only touch at a point.
func TestJTS_PolygonIntersection_TouchingPolygons(t *testing.T) {
	// Two squares touching at corner (10, 10)
	poly1, _ := wkt.UnmarshalString("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	poly2, _ := wkt.UnmarshalString("POLYGON ((10 10, 20 10, 20 20, 10 20, 10 10))")

	result := Intersection(poly1, poly2)
	point, ok := result.(*geom.Point)
	require.True(t, ok, "JTS-compatible point-touch intersection should be Point, got %T", result)
	assert.True(t, point.Coordinate().Equals2D(geom.NewCoordinate(10, 10), geom.DefaultEpsilon))
}

// TestJTS_PolygonIntersection_SharedEdge tests polygons sharing an edge.
func TestJTS_PolygonIntersection_SharedEdge(t *testing.T) {
	// Two rectangles sharing a vertical edge
	poly1, _ := wkt.UnmarshalString("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	poly2, _ := wkt.UnmarshalString("POLYGON ((10 0, 20 0, 20 10, 10 10, 10 0))")

	result := Intersection(poly1, poly2)
	line, ok := result.(*geom.LineString)
	require.True(t, ok, "JTS-compatible shared-edge intersection should be LineString, got %T", result)
	assert.InDelta(t, 10.0, line.Length(), geom.DefaultEpsilon)
}

// TestJTS_PolygonIntersection_PolygonContained tests one polygon completely inside another.
func TestJTS_PolygonIntersection_PolygonContained(t *testing.T) {
	// Large square containing small square
	poly1, _ := wkt.UnmarshalString("POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0))")
	poly2, _ := wkt.UnmarshalString("POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))")

	result := Intersection(poly1, poly2)
	require.False(t, result.IsEmpty(), "Intersection should not be empty")

	area := geometryArea(result)
	expectedArea := 100.0 // Area of inner square

	assert.InDelta(t, expectedArea, area, 0.1, "Contained polygon intersection area")
}

// TestJTS_PolygonUnion_AdjacentSquares tests union of adjacent squares.
func TestJTS_PolygonUnion_AdjacentSquares(t *testing.T) {
	// Two adjacent 10x10 squares
	poly1, _ := wkt.UnmarshalString("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	poly2, _ := wkt.UnmarshalString("POLYGON ((10 0, 20 0, 20 10, 10 10, 10 0))")

	result := Union(poly1, poly2)
	require.False(t, result.IsEmpty(), "Union should not be empty")

	area := geometryArea(result)
	expectedArea := 200.0 // Combined area

	assert.InDelta(t, expectedArea, area, 0.1, "Union area")
}

// TestJTS_PolygonUnion_OverlappingSquares tests union of overlapping squares.
func TestJTS_PolygonUnion_OverlappingSquares(t *testing.T) {
	// Two 10x10 squares with 5x5 overlap
	poly1, _ := wkt.UnmarshalString("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	poly2, _ := wkt.UnmarshalString("POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))")

	result := Union(poly1, poly2)
	require.False(t, result.IsEmpty(), "Union should not be empty")

	area := geometryArea(result)
	expectedArea := 175.0 // 100 + 100 - 25 (overlap)

	assert.InDelta(t, expectedArea, area, 0.1, "Union area")
}

// TestJTS_PolygonUnion_ContainedPolygon tests union where one polygon contains another.
func TestJTS_PolygonUnion_ContainedPolygon(t *testing.T) {
	// Large square containing small square
	poly1, _ := wkt.UnmarshalString("POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0))")
	poly2, _ := wkt.UnmarshalString("POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))")

	result := Union(poly1, poly2)
	require.False(t, result.IsEmpty(), "Union should not be empty")

	area := geometryArea(result)
	expectedArea := 400.0 // Area of outer polygon

	assert.InDelta(t, expectedArea, area, 0.1, "Contained union area")
}

// TestJTS_PolygonDifference_OverlappingSquares tests difference of overlapping squares.
func TestJTS_PolygonDifference_OverlappingSquares(t *testing.T) {
	// A - B where A and B are overlapping squares
	poly1, _ := wkt.UnmarshalString("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	poly2, _ := wkt.UnmarshalString("POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))")

	result := Difference(poly1, poly2)
	require.False(t, result.IsEmpty(), "Difference should not be empty")

	area := geometryArea(result)
	expectedArea := 75.0 // 100 - 25 (overlap)

	assert.InDelta(t, expectedArea, area, 0.1, "Difference area")
}

// TestJTS_PolygonDifference_ContainedPolygon tests difference where B is inside A.
func TestJTS_PolygonDifference_ContainedPolygon(t *testing.T) {
	// Large square minus inner square (creates a polygon with hole)
	poly1, _ := wkt.UnmarshalString("POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0))")
	poly2, _ := wkt.UnmarshalString("POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))")

	result := Difference(poly1, poly2)
	require.False(t, result.IsEmpty(), "Difference should not be empty")

	area := geometryArea(result)
	expectedArea := 300.0 // 400 - 100

	assert.InDelta(t, expectedArea, area, 0.1, "Difference with hole area")

	// Check if result is a polygon with hole
	if poly, ok := result.(*geom.Polygon); ok {
		assert.Equal(t, 1, poly.NumInteriorRings(), "Expected polygon with 1 hole")
	}
}

// TestJTS_PolygonDifference_DisjointPolygons tests difference of disjoint polygons.
func TestJTS_PolygonDifference_DisjointPolygons(t *testing.T) {
	// A - B where A and B don't intersect should return A
	poly1, _ := wkt.UnmarshalString("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	poly2, _ := wkt.UnmarshalString("POLYGON ((20 20, 30 20, 30 30, 20 30, 20 20))")

	result := Difference(poly1, poly2)
	require.False(t, result.IsEmpty(), "Difference should not be empty")

	area := geometryArea(result)
	expectedArea := 100.0 // Area of poly1

	assert.InDelta(t, expectedArea, area, 0.1, "Disjoint difference area")
}

// TestJTS_SymmetricDifference_OverlappingSquares tests symmetric difference.
func TestJTS_SymmetricDifference_OverlappingSquares(t *testing.T) {
	// (A - B) union (B - A)
	poly1, _ := wkt.UnmarshalString("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	poly2, _ := wkt.UnmarshalString("POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))")

	result := SymDifference(poly1, poly2)
	require.False(t, result.IsEmpty(), "SymDifference should not be empty")

	area := geometryArea(result)
	expectedArea := 150.0 // (100 - 25) + (100 - 25)

	assert.InDelta(t, expectedArea, area, 0.1, "SymDifference area")
}

// TestJTS_SymmetricDifference_DisjointPolygons tests symmetric difference of disjoint polygons.
func TestJTS_SymmetricDifference_DisjointPolygons(t *testing.T) {
	// Should return both polygons as MultiPolygon or GeometryCollection
	poly1, _ := wkt.UnmarshalString("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	poly2, _ := wkt.UnmarshalString("POLYGON ((20 20, 30 20, 30 30, 20 30, 20 20))")

	result := SymDifference(poly1, poly2)
	require.False(t, result.IsEmpty(), "SymDifference should not be empty")

	area := geometryArea(result)
	expectedArea := 200.0 // Both polygons

	assert.InDelta(t, expectedArea, area, 0.1, "Disjoint SymDifference area")
}

// TestJTS_LineIntersection_CrossingLines tests intersection of crossing lines.
func TestJTS_LineIntersection_CrossingLines(t *testing.T) {
	// Two lines crossing at (5, 5)
	line1, _ := wkt.UnmarshalString("LINESTRING (0 0, 10 10)")
	line2, _ := wkt.UnmarshalString("LINESTRING (0 10, 10 0)")

	result := Intersection(line1, line2)
	require.False(t, result.IsEmpty(), "Intersection should not be empty")

	// Should return a Point at (5, 5)
	point, ok := result.(*geom.Point)
	require.True(t, ok, "Expected Point result, got %T", result)
	assert.InDelta(t, 5.0, point.X(), 0.01, "Intersection point X")
	assert.InDelta(t, 5.0, point.Y(), 0.01, "Intersection point Y")
}

// TestJTS_LineIntersection_OverlappingSegments tests overlapping line segments.
func TestJTS_LineIntersection_OverlappingSegments(t *testing.T) {
	// Two lines with overlapping segment from (5,5) to (10,10)
	line1, _ := wkt.UnmarshalString("LINESTRING (0 0, 10 10)")
	line2, _ := wkt.UnmarshalString("LINESTRING (5 5, 15 15)")

	result := Intersection(line1, line2)
	assert.False(t, result.IsEmpty(), "Intersection should not be empty")
}

// TestJTS_PolygonWithHole_Intersection tests intersection with polygons containing holes.
func TestJTS_PolygonWithHole_Intersection(t *testing.T) {
	// Polygon with hole
	poly1, _ := wkt.UnmarshalString("POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0), (5 5, 15 5, 15 15, 5 15, 5 5))")
	// Simple square overlapping
	poly2, _ := wkt.UnmarshalString("POLYGON ((10 10, 30 10, 30 30, 10 30, 10 10))")

	result := Intersection(poly1, poly2)
	require.False(t, result.IsEmpty(), "Intersection should not be empty")

	area := geometryArea(result)
	assert.Greater(t, area, 0.0, "Intersection should have positive area")
}

// TestJTS_MultiPolygon_Intersection tests intersection with MultiPolygon.
func TestJTS_MultiPolygon_Intersection(t *testing.T) {
	// MultiPolygon with two separate squares
	multi1, _ := wkt.UnmarshalString("MULTIPOLYGON (((0 0, 10 0, 10 10, 0 10, 0 0)), ((20 0, 30 0, 30 10, 20 10, 20 0)))")
	// Square overlapping first polygon in multi
	poly2, _ := wkt.UnmarshalString("POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))")

	result := Intersection(multi1, poly2)
	require.False(t, result.IsEmpty(), "Intersection should not be empty")

	area := geometryArea(result)
	expectedArea := 25.0 // 5x5 overlap with first polygon only

	assert.InDelta(t, expectedArea, area, 1.0, "MultiPolygon intersection area")
}

// TestJTS_EdgeCases_EmptyGeometries tests overlay with empty geometries.
func TestJTS_EdgeCases_EmptyGeometries(t *testing.T) {
	poly, _ := wkt.UnmarshalString("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	empty, _ := wkt.UnmarshalString("POLYGON EMPTY")

	// Intersection with empty should be empty
	result := Intersection(poly, empty)
	assert.True(t, result.IsEmpty(), "Intersection with empty should be empty")

	// Union with empty should be the non-empty geometry
	result = Union(poly, empty)
	assert.False(t, result.IsEmpty(), "Union with empty should return non-empty geometry")

	area := geometryArea(result)
	assert.InDelta(t, 100.0, area, 0.1, "Union with empty area")
}

// TestJTS_Precision_TinyPolygons tests overlay with very small polygons.
func TestJTS_Precision_TinyPolygons(t *testing.T) {
	// Very small polygons to test precision handling
	poly1, _ := wkt.UnmarshalString("POLYGON ((0 0, 0.001 0, 0.001 0.001, 0 0.001, 0 0))")
	poly2, _ := wkt.UnmarshalString("POLYGON ((0.0005 0.0005, 0.0015 0.0005, 0.0015 0.0015, 0.0005 0.0015, 0.0005 0.0005))")

	result := Intersection(poly1, poly2)
	// Should handle tiny coordinates without panicking
	_ = result.IsEmpty()
}

// TestJTS_ComplexGeometry_ManyVertices tests overlay with complex geometries.
func TestJTS_ComplexGeometry_ManyVertices(t *testing.T) {
	// Create a polygon with many vertices (approximating a circle)
	var coords []string
	coords = append(coords, "0 0")
	for i := 0; i <= 360; i += 10 {
		angle := float64(i) * math.Pi / 180.0
		x := 10 + 5*math.Cos(angle)
		y := 10 + 5*math.Sin(angle)
		coords = append(coords, fmt.Sprintf("%.2f %.2f", x, y))
	}
	coords = append(coords, "0 0")

	wktStr := "POLYGON ((" + coords[0]
	for i := 1; i < len(coords); i++ {
		wktStr += ", " + coords[i]
	}
	wktStr += "))"

	poly1, err := wkt.UnmarshalString(wktStr)
	require.NoError(t, err)

	// Simple square
	poly2, _ := wkt.UnmarshalString("POLYGON ((8 8, 12 8, 12 12, 8 12, 8 8))")

	result := Intersection(poly1, poly2)
	assert.False(t, result.IsEmpty(), "Intersection should not be empty")
}
