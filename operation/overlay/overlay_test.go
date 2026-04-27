package overlay

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
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
	shell1 := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := mustLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
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
	shell1 := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := mustLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
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
	ls := mustLineStringXY(0, 0, 10, 0)

	result := Intersection(p, ls)
	assert.False(t, result.IsEmpty(), "Point on line should intersect")

	// Point off line
	p2 := geom.NewPoint(5, 5)
	result2 := Intersection(p2, ls)
	assert.True(t, result2.IsEmpty(), "Point off line should not intersect")
}

func TestIntersectionPointPolygon(t *testing.T) {
	p := geom.NewPoint(5, 5)
	shell := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)

	result := Intersection(p, poly)
	assert.False(t, result.IsEmpty(), "Point inside polygon should intersect")

	// Point outside polygon
	p2 := geom.NewPoint(15, 15)
	result2 := Intersection(p2, poly)
	assert.True(t, result2.IsEmpty(), "Point outside polygon should not intersect")
}

func TestIntersectionLineLine(t *testing.T) {
	ls1 := mustLineStringXY(0, 0, 10, 10)
	ls2 := mustLineStringXY(0, 10, 10, 0)

	result := Intersection(ls1, ls2)
	assert.False(t, result.IsEmpty(), "Crossing lines should have intersection")
}

func TestIntersectionLineLineOverlapReturnsNodedLine(t *testing.T) {
	ls1 := mustLineStringXY(0, 0, 10, 0)
	ls2 := mustLineStringXY(5, 0, 15, 0)

	result := Intersection(ls1, ls2)
	line, ok := result.(*geom.LineString)
	require.True(t, ok, "Expected overlap intersection to return LineString, got %T", result)

	coords := line.Coordinates()
	require.Len(t, coords, 2)
	assert.True(t, coords[0].Equals2D(geom.NewCoordinate(5, 0), geom.DefaultEpsilon), "overlap start")
	assert.True(t, coords[1].Equals2D(geom.NewCoordinate(10, 0), geom.DefaultEpsilon), "overlap end")
}

func TestIntersectionLineLineCollinearTouchReturnsPoint(t *testing.T) {
	ls1 := mustLineStringXY(0, 0, 10, 0)
	ls2 := mustLineStringXY(10, 0, 20, 0)

	result := Intersection(ls1, ls2)
	point, ok := result.(*geom.Point)
	require.True(t, ok, "Expected endpoint touch intersection to return Point, got %T", result)
	assert.True(t, point.Coordinate().Equals2D(geom.NewCoordinate(10, 0), geom.DefaultEpsilon))
}

func TestIntersectionLinePolygon(t *testing.T) {
	ls := mustLineStringXY(5, -5, 5, 15) // Vertical line through polygon
	shell := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
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
	shell1 := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := mustLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	poly2 := geom.NewPolygon(shell2, nil)

	result := Intersection(poly1, poly2)
	assert.False(t, result.IsEmpty(), "Overlapping polygons should have intersection")
}

func TestIntersectionDisjointPolygons(t *testing.T) {
	shell1 := mustLinearRingXY(0, 0, 5, 0, 5, 5, 0, 5, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := mustLinearRingXY(10, 10, 15, 10, 15, 15, 10, 15, 10, 10)
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

func TestUnionLineLineNodedAtCrossing(t *testing.T) {
	ls1 := mustLineStringXY(0, 0, 10, 0)
	ls2 := mustLineStringXY(5, -5, 5, 5)

	result := Union(ls1, ls2)
	mls, ok := result.(*geom.MultiLineString)
	require.True(t, ok, "Expected noded line union to return MultiLineString, got %T", result)
	assert.Equal(t, 4, mls.NumGeometries(), "Crossing line union should be split at the crossing node")
}

func TestUnionLineLineDissolvesDuplicateOverlap(t *testing.T) {
	ls1 := mustLineStringXY(0, 0, 10, 0)
	ls2 := mustLineStringXY(5, 0, 15, 0)

	result := Union(ls1, ls2)
	mls, ok := result.(*geom.MultiLineString)
	require.True(t, ok, "Expected overlapping line union to return MultiLineString, got %T", result)
	assert.Equal(t, 3, mls.NumGeometries(), "Overlapping line union should node overlap endpoints and remove duplicate overlap")
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
	shell1 := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := mustLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	poly2 := geom.NewPolygon(shell2, nil)

	result := Union(poly1, poly2)
	assert.False(t, result.IsEmpty(), "Union of polygons should not be empty")
}

func TestDifferencePointPolygon(t *testing.T) {
	// Point inside polygon
	p := geom.NewPoint(5, 5)
	shell := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)

	result := Difference(p, poly)
	assert.True(t, result.IsEmpty(), "Point inside polygon minus polygon should be empty")

	// Point outside polygon
	p2 := geom.NewPoint(15, 15)
	result2 := Difference(p2, poly)
	assert.False(t, result2.IsEmpty(), "Point outside polygon minus polygon should be the point")
}

func TestDifferencePolygonPoint(t *testing.T) {
	shell := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)
	p := geom.NewPoint(5, 5)

	result := Difference(poly, p)
	assert.False(t, result.IsEmpty(), "Polygon minus point should be polygon")
}

func TestDifferenceLineLineOverlapReturnsUncoveredNodedLine(t *testing.T) {
	ls1 := mustLineStringXY(0, 0, 10, 0)
	ls2 := mustLineStringXY(5, 0, 15, 0)

	result := Difference(ls1, ls2)
	line, ok := result.(*geom.LineString)
	require.True(t, ok, "Expected overlap difference to return LineString, got %T", result)

	coords := line.Coordinates()
	require.Len(t, coords, 2)
	assert.True(t, coords[0].Equals2D(geom.NewCoordinate(0, 0), geom.DefaultEpsilon), "difference start")
	assert.True(t, coords[1].Equals2D(geom.NewCoordinate(5, 0), geom.DefaultEpsilon), "difference end")
}

func TestDifferenceLineLineCrossingIsNoded(t *testing.T) {
	ls1 := mustLineStringXY(0, 0, 10, 0)
	ls2 := mustLineStringXY(5, -5, 5, 5)

	result := Difference(ls1, ls2)
	mls, ok := result.(*geom.MultiLineString)
	require.True(t, ok, "Expected crossing difference to return noded MultiLineString, got %T", result)
	assert.Equal(t, 2, mls.NumGeometries())
}

func TestDifferencePolygonPolygon(t *testing.T) {
	shell1 := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := mustLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	poly2 := geom.NewPolygon(shell2, nil)

	result := Difference(poly1, poly2)
	assert.False(t, result.IsEmpty(), "Partial overlap difference should not be empty")
}

func TestDifferenceDisjointPolygons(t *testing.T) {
	shell1 := mustLinearRingXY(0, 0, 5, 0, 5, 5, 0, 5, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := mustLinearRingXY(10, 10, 15, 10, 15, 15, 10, 15, 10, 10)
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

func TestSymDifferenceLineLineOverlapExcludesSharedSegment(t *testing.T) {
	ls1 := mustLineStringXY(0, 0, 10, 0)
	ls2 := mustLineStringXY(5, 0, 15, 0)

	result := SymDifference(ls1, ls2)
	mls, ok := result.(*geom.MultiLineString)
	require.True(t, ok, "Expected overlapping line symmetric difference to return MultiLineString, got %T", result)
	assert.Equal(t, 2, mls.NumGeometries())
}

func TestSymDifferencePolygonPolygon(t *testing.T) {
	shell1 := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := mustLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
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
	points := geom.ExtractPoints(p)
	assert.Len(t, points, 1, "Expected 1 point")

	mp := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(0, 0),
		geom.NewPoint(1, 1),
	})
	points = geom.ExtractPoints(mp)
	assert.Len(t, points, 2, "Expected 2 points")
}

func TestExtractLineStrings(t *testing.T) {
	ls := mustLineStringXY(0, 0, 10, 10)
	lines := geom.ExtractLineStringsWithRings(ls)
	assert.Len(t, lines, 1, "Expected 1 line")

	mls := geom.NewMultiLineString([]*geom.LineString{
		mustLineStringXY(0, 0, 10, 0),
		mustLineStringXY(0, 10, 10, 10),
	})
	lines = geom.ExtractLineStringsWithRings(mls)
	assert.Len(t, lines, 2, "Expected 2 lines")
}

func TestExtractPolygons(t *testing.T) {
	shell := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)
	polygons := geom.ExtractPolygons(poly)
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
	shell := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)

	result := Intersection(mp, poly)
	// Two points should be inside the polygon
	assert.False(t, result.IsEmpty(), "Expected non-empty intersection")
}

func BenchmarkIntersectionPolygonPolygon(b *testing.B) {
	shell1 := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := mustLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	poly2 := geom.NewPolygon(shell2, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Intersection(poly1, poly2)
	}
}

func BenchmarkUnionPolygonPolygon(b *testing.B) {
	shell1 := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := mustLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	poly2 := geom.NewPolygon(shell2, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Union(poly1, poly2)
	}
}

// --- Line/Line overlay tests ---

func TestLineLineIntersectionCollinearOverlap(t *testing.T) {
	// Two collinear lines that partially overlap:
	// lineA: (0,0)-(10,0), lineB: (5,0)-(15,0)
	// Expected intersection: LineString (5,0)-(10,0)
	lineA := mustLineStringXY(0, 0, 10, 0)
	lineB := mustLineStringXY(5, 0, 15, 0)

	result := Intersection(lineA, lineB)
	require.False(t, result.IsEmpty(), "Collinear overlapping lines should have non-empty intersection")

	ls, ok := result.(*geom.LineString)
	require.True(t, ok, "Expected LineString for collinear overlap, got %T", result)

	coords := ls.Coordinates()
	require.Equal(t, 2, len(coords), "Expected 2 coordinates in intersection LineString")

	// The overlap is from x=5 to x=10
	assert.InDelta(t, 5.0, coords[0].X, 0.01)
	assert.InDelta(t, 0.0, coords[0].Y, 0.01)
	assert.InDelta(t, 10.0, coords[1].X, 0.01)
	assert.InDelta(t, 0.0, coords[1].Y, 0.01)
}

func TestLineLineIntersectionIdentical(t *testing.T) {
	// Identical lines: intersection should return the line itself
	lineA := mustLineStringXY(0, 0, 10, 0)
	lineB := mustLineStringXY(0, 0, 10, 0)

	result := Intersection(lineA, lineB)
	require.False(t, result.IsEmpty(), "Identical lines should have non-empty intersection")

	ls, ok := result.(*geom.LineString)
	require.True(t, ok, "Expected LineString for identical lines, got %T", result)

	coords := ls.Coordinates()
	require.Equal(t, 2, len(coords))
	assert.InDelta(t, 0.0, coords[0].X, 0.01)
	assert.InDelta(t, 10.0, coords[1].X, 0.01)
}

func TestLineLineIntersectionDisjoint(t *testing.T) {
	// Disjoint lines: no common points
	lineA := mustLineStringXY(0, 0, 5, 0)
	lineB := mustLineStringXY(0, 5, 5, 5)

	result := Intersection(lineA, lineB)
	assert.True(t, result.IsEmpty(), "Disjoint lines should have empty intersection")
}

func TestLineLineIntersectionCrossing(t *testing.T) {
	// Two lines crossing at a point: (0,0)-(10,10) and (0,10)-(10,0) cross at (5,5)
	lineA := mustLineStringXY(0, 0, 10, 10)
	lineB := mustLineStringXY(0, 10, 10, 0)

	result := Intersection(lineA, lineB)
	require.False(t, result.IsEmpty(), "Crossing lines should have non-empty intersection")

	pt, ok := result.(*geom.Point)
	require.True(t, ok, "Expected Point for crossing lines, got %T", result)

	assert.InDelta(t, 5.0, pt.X(), 0.01)
	assert.InDelta(t, 5.0, pt.Y(), 0.01)
}

func TestLineLineDifferenceCollinearOverlap(t *testing.T) {
	// lineA: (0,0)-(10,0), lineB: (5,0)-(15,0)
	// Difference A-B should return (0,0)-(5,0)
	lineA := mustLineStringXY(0, 0, 10, 0)
	lineB := mustLineStringXY(5, 0, 15, 0)

	result := Difference(lineA, lineB)
	require.False(t, result.IsEmpty(), "Difference of partially overlapping lines should not be empty")

	ls, ok := result.(*geom.LineString)
	require.True(t, ok, "Expected LineString, got %T", result)

	coords := ls.Coordinates()
	require.Equal(t, 2, len(coords))
	assert.InDelta(t, 0.0, coords[0].X, 0.01)
	assert.InDelta(t, 5.0, coords[1].X, 0.01)
}

func TestLineLineDifferenceIdentical(t *testing.T) {
	// Identical lines: difference should be empty
	lineA := mustLineStringXY(0, 0, 10, 0)
	lineB := mustLineStringXY(0, 0, 10, 0)

	result := Difference(lineA, lineB)
	assert.True(t, result.IsEmpty(), "Difference of identical lines should be empty")
}

func TestLineLineDifferenceDisjoint(t *testing.T) {
	// Disjoint lines: difference should return the original line
	lineA := mustLineStringXY(0, 0, 10, 0)
	lineB := mustLineStringXY(0, 5, 10, 5)

	result := Difference(lineA, lineB)
	require.False(t, result.IsEmpty(), "Difference of disjoint lines should not be empty")

	ls, ok := result.(*geom.LineString)
	require.True(t, ok, "Expected LineString, got %T", result)

	coords := ls.Coordinates()
	require.Equal(t, 2, len(coords))
	assert.InDelta(t, 0.0, coords[0].X, 0.01)
	assert.InDelta(t, 10.0, coords[1].X, 0.01)
}

func TestLineLineSymDifferenceCollinearOverlap(t *testing.T) {
	// lineA: (0,0)-(10,0), lineB: (5,0)-(15,0)
	// SymDifference should return (0,0)-(5,0) and (10,0)-(15,0)
	lineA := mustLineStringXY(0, 0, 10, 0)
	lineB := mustLineStringXY(5, 0, 15, 0)

	result := SymDifference(lineA, lineB)
	require.False(t, result.IsEmpty(), "SymDifference of partially overlapping lines should not be empty")
}
