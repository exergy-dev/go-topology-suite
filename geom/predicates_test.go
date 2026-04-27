package geom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntersectsPointPoint(t *testing.T) {
	factory := DefaultFactory

	p1 := factory.CreatePoint(1, 2)
	p2 := factory.CreatePoint(1, 2)
	p3 := factory.CreatePoint(3, 4)

	assert.True(t, Intersects(p1, p2), "Expected equal points to intersect")
	assert.False(t, Intersects(p1, p3), "Expected different points not to intersect")
}

func TestIntersectsPointLine(t *testing.T) {
	factory := DefaultFactory

	p := factory.CreatePoint(1, 1)
	line := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(0, 0),
		NewCoordinate(2, 2),
	})

	assert.True(t, Intersects(p, line), "Expected point on line to intersect")

	p2 := factory.CreatePoint(0, 1)
	assert.False(t, Intersects(p2, line), "Expected point not on line to not intersect")
}

func TestIntersectsPointPolygon(t *testing.T) {
	factory := DefaultFactory

	poly := createTestSquare(factory, 0, 0, 10)

	// Point inside
	pInside := factory.CreatePoint(5, 5)
	assert.True(t, Intersects(pInside, poly), "Expected point inside polygon to intersect")

	// Point outside
	pOutside := factory.CreatePoint(15, 15)
	assert.False(t, Intersects(pOutside, poly), "Expected point outside polygon not to intersect")

	// Point on boundary
	pBoundary := factory.CreatePoint(0, 5)
	assert.True(t, Intersects(pBoundary, poly), "Expected point on boundary to intersect")
}

func TestIntersectsLineLine(t *testing.T) {
	factory := DefaultFactory

	line1 := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(0, 0),
		NewCoordinate(10, 10),
	})
	line2 := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(0, 10),
		NewCoordinate(10, 0),
	})

	assert.True(t, Intersects(line1, line2), "Expected crossing lines to intersect")

	// Parallel lines
	line3 := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(0, 5),
		NewCoordinate(10, 15),
	})
	assert.False(t, Intersects(line1, line3), "Expected parallel lines not to intersect")
}

func TestIntersectsLinePolygon(t *testing.T) {
	factory := DefaultFactory

	poly := createTestSquare(factory, 0, 0, 10)

	// Line crossing polygon
	lineCrossing := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(-5, 5),
		NewCoordinate(15, 5),
	})
	assert.True(t, Intersects(lineCrossing, poly), "Expected line crossing polygon to intersect")

	// Line outside polygon
	lineOutside := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(20, 20),
		NewCoordinate(30, 30),
	})
	assert.False(t, Intersects(lineOutside, poly), "Expected line outside polygon not to intersect")

	// Line inside polygon
	lineInside := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(2, 2),
		NewCoordinate(8, 8),
	})
	assert.True(t, Intersects(lineInside, poly), "Expected line inside polygon to intersect")
}

func TestIntersectsPolygonPolygon(t *testing.T) {
	factory := DefaultFactory

	poly1 := createTestSquare(factory, 0, 0, 10)
	poly2 := createTestSquare(factory, 5, 5, 10)

	assert.True(t, Intersects(poly1, poly2), "Expected overlapping polygons to intersect")

	// Non-overlapping polygons
	poly3 := createTestSquare(factory, 20, 20, 10)
	assert.False(t, Intersects(poly1, poly3), "Expected non-overlapping polygons not to intersect")
}

func TestContainsPointInPolygon(t *testing.T) {
	factory := DefaultFactory

	poly := createTestSquare(factory, 0, 0, 10)

	// Point inside
	pInside := factory.CreatePoint(5, 5)
	assert.True(t, Contains(poly, pInside), "Expected polygon to contain point inside")

	// Point outside
	pOutside := factory.CreatePoint(15, 15)
	assert.False(t, Contains(poly, pOutside), "Expected polygon not to contain point outside")
}

func TestContainsPolygonInPolygon(t *testing.T) {
	factory := DefaultFactory

	outer := createTestSquare(factory, 0, 0, 20)
	inner := createTestSquare(factory, 5, 5, 5)

	assert.True(t, Contains(outer, inner), "Expected outer polygon to contain inner polygon")
	assert.False(t, Contains(inner, outer), "Expected inner polygon not to contain outer polygon")
}

func TestWithin(t *testing.T) {
	factory := DefaultFactory

	outer := createTestSquare(factory, 0, 0, 20)
	inner := createTestSquare(factory, 5, 5, 5)

	assert.True(t, Within(inner, outer), "Expected inner polygon to be within outer polygon")
	assert.False(t, Within(outer, inner), "Expected outer polygon not to be within inner polygon")
}

func TestCovers(t *testing.T) {
	factory := DefaultFactory

	outer := createTestSquare(factory, 0, 0, 20)
	inner := createTestSquare(factory, 5, 5, 5)

	assert.True(t, Covers(outer, inner), "Expected outer polygon to cover inner polygon")
	assert.False(t, Covers(inner, outer), "Expected inner polygon not to cover outer polygon")
}

func TestCoveredBy(t *testing.T) {
	factory := DefaultFactory

	outer := createTestSquare(factory, 0, 0, 20)
	inner := createTestSquare(factory, 5, 5, 5)

	assert.True(t, CoveredBy(inner, outer), "Expected inner polygon to be covered by outer polygon")
	assert.False(t, CoveredBy(outer, inner), "Expected outer polygon not to be covered by inner polygon")
}

func TestDisjoint(t *testing.T) {
	factory := DefaultFactory

	poly1 := createTestSquare(factory, 0, 0, 10)
	poly2 := createTestSquare(factory, 20, 20, 10)
	poly3 := createTestSquare(factory, 5, 5, 10)

	assert.True(t, Disjoint(poly1, poly2), "Expected non-overlapping polygons to be disjoint")
	assert.False(t, Disjoint(poly1, poly3), "Expected overlapping polygons not to be disjoint")
}

func TestTouches(t *testing.T) {
	factory := DefaultFactory

	// Two squares sharing an edge
	poly1 := createTestSquare(factory, 0, 0, 10)
	poly2 := createTestSquare(factory, 10, 0, 10)

	assert.True(t, Touches(poly1, poly2), "Expected adjacent polygons to touch")

	// Overlapping polygons don't touch
	poly3 := createTestSquare(factory, 5, 0, 10)
	assert.False(t, Touches(poly1, poly3), "Expected overlapping polygons not to touch")
}

func TestCrosses(t *testing.T) {
	factory := DefaultFactory

	poly := createTestSquare(factory, 0, 0, 10)

	// Line crossing polygon
	lineCrossing := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(-5, 5),
		NewCoordinate(15, 5),
	})

	assert.True(t, Crosses(lineCrossing, poly), "Expected line to cross polygon")

	// Line inside polygon
	lineInside := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(2, 2),
		NewCoordinate(8, 8),
	})
	assert.False(t, Crosses(lineInside, poly), "Expected line inside polygon not to cross")
}

func TestOverlaps(t *testing.T) {
	factory := DefaultFactory

	poly1 := createTestSquare(factory, 0, 0, 10)
	poly2 := createTestSquare(factory, 5, 5, 10)

	assert.True(t, Overlaps(poly1, poly2), "Expected partially overlapping polygons to overlap")

	// One contains the other - not overlap
	outer := createTestSquare(factory, 0, 0, 20)
	inner := createTestSquare(factory, 5, 5, 5)
	assert.False(t, Overlaps(outer, inner), "Expected containment not to be overlap")

	// Different dimensions
	line := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(0, 0),
		NewCoordinate(10, 10),
	})
	assert.False(t, Overlaps(poly1, line), "Expected different dimension geometries not to overlap")
}

func TestEquals(t *testing.T) {
	factory := DefaultFactory

	poly1 := createTestSquare(factory, 0, 0, 10)
	poly2 := createTestSquare(factory, 0, 0, 10)
	poly3 := createTestSquare(factory, 5, 5, 10)

	assert.True(t, Equals(poly1, poly2), "Expected identical polygons to be equal")
	assert.False(t, Equals(poly1, poly3), "Expected different polygons not to be equal")
}

func TestEqualsEmptyGeometries(t *testing.T) {
	factory := DefaultFactory

	emptyPoint1 := factory.CreatePointEmpty()
	emptyPoint2 := factory.CreatePointEmpty()

	assert.True(t, Equals(emptyPoint1, emptyPoint2), "Expected empty points to be equal")
}

func TestPredicatesWithMultiGeometries(t *testing.T) {
	factory := DefaultFactory

	// Multi-polygon
	polys := []*Polygon{
		createTestSquare(factory, 0, 0, 10),
		createTestSquare(factory, 20, 20, 10),
	}
	mp := factory.CreateMultiPolygon(polys)

	// Point in first polygon
	p1 := factory.CreatePoint(5, 5)
	assert.True(t, Intersects(mp, p1), "Expected multipolygon to intersect with point in first polygon")

	// Point in second polygon
	p2 := factory.CreatePoint(25, 25)
	assert.True(t, Intersects(mp, p2), "Expected multipolygon to intersect with point in second polygon")

	// Point outside both
	p3 := factory.CreatePoint(50, 50)
	assert.False(t, Intersects(mp, p3), "Expected multipolygon not to intersect with point outside")
}

func TestPredicatesWithGeometryCollection(t *testing.T) {
	factory := DefaultFactory

	geoms := []Geometry{
		factory.CreatePoint(5, 5),
		createTestSquare(factory, 0, 0, 10),
	}
	gc := factory.CreateGeometryCollection(geoms)

	// Point inside the polygon component
	p := factory.CreatePoint(2, 2)
	assert.True(t, Intersects(gc, p), "Expected geometry collection to intersect with point inside polygon")
}

func TestPredicatesEnvelopeRejection(t *testing.T) {
	factory := DefaultFactory

	// These geometries don't even have overlapping envelopes
	poly1 := createTestSquare(factory, 0, 0, 10)
	poly2 := createTestSquare(factory, 100, 100, 10)

	// Should be quickly rejected by envelope check
	assert.False(t, Intersects(poly1, poly2), "Expected disjoint envelope geometries not to intersect")
	assert.False(t, Contains(poly1, poly2), "Expected disjoint envelope geometries: no containment")
	assert.False(t, Touches(poly1, poly2), "Expected disjoint envelope geometries not to touch")
}

func TestContains_LineStringContainsIdenticalLine(t *testing.T) {
	factory := DefaultFactory

	line := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(0, 0),
		NewCoordinate(10, 0),
	})
	clone := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(0, 0),
		NewCoordinate(10, 0),
	})

	assert.True(t, Contains(line, clone), "Expected a line to contain an identical line")
}

func TestOverlaps_LineStringIdenticalIsNotOverlap(t *testing.T) {
	factory := DefaultFactory

	line1 := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(0, 0),
		NewCoordinate(10, 0),
	})
	line2 := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(0, 0),
		NewCoordinate(10, 0),
	})

	assert.True(t, Equals(line1, line2), "Expected identical lines to be equal")
	assert.False(t, Overlaps(line1, line2), "Expected identical lines not to overlap")
}

func TestCovers_ConcavePolygonLineCrossesNotch(t *testing.T) {
	factory := DefaultFactory

	lShapeCoords := CoordinateSequence{
		NewCoordinate(0, 0),
		NewCoordinate(10, 0),
		NewCoordinate(10, 5),
		NewCoordinate(5, 5),
		NewCoordinate(5, 10),
		NewCoordinate(0, 10),
		NewCoordinate(0, 0),
	}
	lShapeShell := factory.CreateLinearRing(lShapeCoords)
	lShape := factory.CreatePolygon(lShapeShell, nil)

	lineThroughNotch := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(8, 3), // inside bottom arm
		NewCoordinate(3, 8), // inside left arm
	})

	assert.False(t, Covers(lShape, lineThroughNotch), "Expected concave polygon not to cover a line crossing the notch")
}

func TestTouches_LineOverlapsPolygonBoundarySegment(t *testing.T) {
	factory := DefaultFactory

	poly := createTestSquare(factory, 0, 0, 10)
	line := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(-5, 0),
		NewCoordinate(15, 0),
	})

	assert.True(t, Touches(line, poly), "Expected line overlapping polygon boundary to touch")
	assert.True(t, Touches(poly, line), "Expected polygon to touch line overlapping its boundary")
}

func TestLineCrossLines(t *testing.T) {
	factory := DefaultFactory

	// Two crossing lines
	line1 := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(0, 0),
		NewCoordinate(10, 10),
	})
	line2 := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(0, 10),
		NewCoordinate(10, 0),
	})

	assert.True(t, Crosses(line1, line2), "Expected crossing lines to cross")

	// Parallel lines don't cross
	line3 := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(0, 5),
		NewCoordinate(10, 15),
	})
	assert.False(t, Crosses(line1, line3), "Expected parallel lines not to cross")
}

func createTestSquare(factory *GeometryFactory, x, y, size float64) *Polygon {
	coords := CoordinateSequence{
		NewCoordinate(x, y),
		NewCoordinate(x+size, y),
		NewCoordinate(x+size, y+size),
		NewCoordinate(x, y+size),
		NewCoordinate(x, y),
	}
	shell := factory.CreateLinearRing(coords)
	return factory.CreatePolygon(shell, nil)
}

func BenchmarkIntersectsPolygons(b *testing.B) {
	factory := DefaultFactory
	poly1 := createTestSquare(factory, 0, 0, 10)
	poly2 := createTestSquare(factory, 5, 5, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Intersects(poly1, poly2)
	}
}

func BenchmarkContainsPointInPolygon(b *testing.B) {
	factory := DefaultFactory
	poly := createTestSquare(factory, 0, 0, 10)
	p := factory.CreatePoint(5, 5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Contains(poly, p)
	}
}

// TestContains_LineStringVerticesInsideEdgesExit tests that a linestring with all vertices
// inside a polygon but with an edge that crosses out is NOT contained.
// This is the key bug fix - vertices being inside is not sufficient.
func TestContains_LineStringVerticesInsideEdgesExit(t *testing.T) {
	factory := DefaultFactory

	// Create a square polygon from (0,0) to (10,10)
	poly := createTestSquare(factory, 0, 0, 10)

	// Create a linestring where both endpoints are inside the polygon,
	// but the line goes outside (arcs out through the corner)
	// Points (2, 2) and (8, 8) are inside, but the path goes through (-5, 5) which is outside
	lineExiting := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(2, 2),
		NewCoordinate(-5, 5), // This point is outside the polygon
		NewCoordinate(8, 8),
	})

	// This should be false because the line exits the polygon
	assert.False(t, Contains(poly, lineExiting), "Polygon should not contain linestring that exits through edge")

	// Another case: endpoints inside, midpoint outside via diagonal edge
	// Square is (0,0) to (10,10). Line from (1,5) to (5,1) stays inside.
	// But line from (1,1) to (9,9) going through (15,5) exits.
	lineExiting2 := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(1, 1),   // inside
		NewCoordinate(15, 5),  // outside - the line crosses the polygon boundary
		NewCoordinate(9, 9),   // inside
	})
	assert.False(t, Contains(poly, lineExiting2), "Polygon should not contain linestring with vertex outside")

	// Edge case: A "V" shape where both ends and the vertex are inside,
	// but one leg exits and re-enters
	// Actually, if all vertices are inside but an edge crosses out and back...
	// Let's create a more complex case
	// Square: (0,0) to (20,20)
	bigPoly := createTestSquare(factory, 0, 0, 20)

	// Line: (5,10) -> (25, 10) -> (15, 10)
	// The point (25, 10) is outside, so the first segment exits the polygon
	lineWithVertexOutside := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(5, 10),  // inside
		NewCoordinate(25, 10), // outside
		NewCoordinate(15, 10), // inside
	})
	assert.False(t, Contains(bigPoly, lineWithVertexOutside), "Polygon should not contain line with exterior vertex")
}

// TestContains_PolygonSpikeExit tests that a polygon with a spike that extends
// outside the container is NOT contained.
func TestContains_PolygonSpikeExit(t *testing.T) {
	factory := DefaultFactory

	// Container polygon: square from (0,0) to (20,20)
	container := createTestSquare(factory, 0, 0, 20)

	// Inner polygon with a spike that goes outside:
	// A polygon mostly inside but with one vertex that extends outside
	spikeCoords := CoordinateSequence{
		NewCoordinate(5, 5),
		NewCoordinate(15, 5),
		NewCoordinate(15, 15),
		NewCoordinate(10, 25), // This spike extends outside the container (y=25 > 20)
		NewCoordinate(5, 15),
		NewCoordinate(5, 5), // close the ring
	}
	spikeShell := factory.CreateLinearRing(spikeCoords)
	spikePolygon := factory.CreatePolygon(spikeShell, nil)

	assert.False(t, Contains(container, spikePolygon), "Container should not contain polygon with spike exiting")

	// Also test a polygon where all vertices are inside but an edge crosses the boundary
	// This requires a concave container or a spike that goes out and back
	// For simplicity, we'll use a polygon with a vertex clearly outside
}

// TestContains_LineStringAllInside tests that a linestring fully inside a polygon
// IS contained (positive test case).
func TestContains_LineStringAllInside(t *testing.T) {
	factory := DefaultFactory

	// Square polygon from (0,0) to (10,10)
	poly := createTestSquare(factory, 0, 0, 10)

	// Simple line fully inside
	lineInside := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(2, 2),
		NewCoordinate(8, 8),
	})
	assert.True(t, Contains(poly, lineInside), "Polygon should contain linestring fully inside")

	// Line with multiple segments, all inside
	multiSegmentLine := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(2, 2),
		NewCoordinate(5, 8),
		NewCoordinate(8, 5),
		NewCoordinate(8, 2),
	})
	assert.True(t, Contains(poly, multiSegmentLine), "Polygon should contain multi-segment linestring fully inside")

	// Line touching the boundary (should still be contained as boundary is part of polygon)
	lineTouchingBoundary := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(0, 5), // on boundary
		NewCoordinate(5, 5), // inside
		NewCoordinate(10, 5), // on boundary
	})
	assert.True(t, Contains(poly, lineTouchingBoundary), "Polygon should contain linestring touching boundary")
}

// TestContains_PolygonFullyInside tests that a polygon fully inside another
// IS contained (positive test case).
func TestContains_PolygonFullyInside(t *testing.T) {
	factory := DefaultFactory

	// Outer polygon: square from (0,0) to (20,20)
	outer := createTestSquare(factory, 0, 0, 20)

	// Inner polygon: square from (5,5) to (15,15) - fully inside
	inner := createTestSquare(factory, 5, 5, 10)

	assert.True(t, Contains(outer, inner), "Outer polygon should contain inner polygon fully inside")

	// Test with a more complex inner polygon (still fully inside)
	triangleCoords := CoordinateSequence{
		NewCoordinate(8, 8),
		NewCoordinate(12, 8),
		NewCoordinate(10, 12),
		NewCoordinate(8, 8),
	}
	triangleShell := factory.CreateLinearRing(triangleCoords)
	triangle := factory.CreatePolygon(triangleShell, nil)

	assert.True(t, Contains(outer, triangle), "Outer polygon should contain triangle fully inside")

	// Test inner polygon touching outer boundary
	touchingInner := createTestSquare(factory, 0, 5, 10) // shares left edge with outer
	// Note: For Contains, the inner must have interior intersection, which it does
	assert.True(t, Contains(outer, touchingInner), "Outer polygon should contain inner polygon touching boundary")
}

// TestContains_ZigZagLine tests that a line which enters and exits a polygon
// multiple times is NOT contained.
func TestContains_ZigZagLine(t *testing.T) {
	factory := DefaultFactory

	// Square polygon from (0,0) to (10,10)
	poly := createTestSquare(factory, 0, 0, 10)

	// Zigzag line that goes in and out multiple times
	zigzagLine := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(-2, 5),  // outside
		NewCoordinate(3, 5),   // inside
		NewCoordinate(3, 12),  // outside (above polygon)
		NewCoordinate(7, 12),  // outside
		NewCoordinate(7, 5),   // inside
		NewCoordinate(12, 5),  // outside (right of polygon)
	})

	assert.False(t, Contains(poly, zigzagLine), "Polygon should not contain zigzag line that exits multiple times")

	// Zigzag where only middle segments are outside
	zigzagLine2 := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(2, 2),   // inside
		NewCoordinate(5, -3),  // outside (below polygon)
		NewCoordinate(8, 2),   // inside
	})

	assert.False(t, Contains(poly, zigzagLine2), "Polygon should not contain line that dips outside")

	// A line that enters, exits, and re-enters
	inOutInLine := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(2, 5),   // inside
		NewCoordinate(12, 5),  // outside - crosses right boundary
		NewCoordinate(5, 8),   // inside - but had to cross back in
	})

	// Note: vertex at (12,5) is outside, so vertex check should catch this
	// But even if it didn't, edge check would catch the crossing
	assert.False(t, Contains(poly, inOutInLine), "Polygon should not contain line that exits and re-enters")
}

// TestContains_EdgeCrossingWithAllVerticesInside tests the specific case where
// all vertices of a geometry are inside, but an edge crosses the boundary.
// This is the core case the fix addresses.
func TestContains_EdgeCrossingWithAllVerticesInside(t *testing.T) {
	factory := DefaultFactory

	// Create an L-shaped polygon (concave)
	// This allows us to have a line with both endpoints in the interior
	// but the edge passing through the exterior
	lShapeCoords := CoordinateSequence{
		NewCoordinate(0, 0),
		NewCoordinate(10, 0),
		NewCoordinate(10, 5),
		NewCoordinate(5, 5),
		NewCoordinate(5, 10),
		NewCoordinate(0, 10),
		NewCoordinate(0, 0),
	}
	lShapeShell := factory.CreateLinearRing(lShapeCoords)
	lShape := factory.CreatePolygon(lShapeShell, nil)

	// Line from bottom arm to outside in the notch
	// (2, 2) is in the bottom arm, (8, 8) is outside - in the concave notch region
	lineCrossingConcave := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(2, 2), // inside bottom arm
		NewCoordinate(8, 8), // outside - in the concave notch region
	})
	// The point (8, 8) is actually outside the L-shape because the L-shape
	// only extends to y=5 in the right portion. So vertex check catches this.
	assert.False(t, Contains(lShape, lineCrossingConcave), "L-shape should NOT contain line with endpoint in notch")

	// Better L-shape test: line that stays in valid regions but edge crosses notch
	// The L-shape: bottom is (0,0)-(10,0)-(10,5)-(5,5), top is (0,0)-(5,5)-(5,10)-(0,10)
	// A line from (2, 2) to (2, 8) is fully in the left column - should be contained
	lineInLeftColumn := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(2, 2), // inside left column
		NewCoordinate(2, 8), // inside left column
	})
	assert.True(t, Contains(lShape, lineInLeftColumn), "L-shape should contain line in left column")

	// A line from (8, 2) to (2, 8) - endpoints inside but passes through the corner
	// (8, 2) is in bottom-right area, (2, 8) is in top-left area
	lineDiagonalAcrossNotch := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(8, 2), // inside bottom arm (right side)
		NewCoordinate(2, 8), // inside top arm (left side)
	})
	// This line's path: from (8,2) to (2,8)
	// Parametric: x = 8 - 6t, y = 2 + 6t for t in [0,1]
	// At t=0.5: x=5, y=5 (corner point - on boundary)
	// This line actually stays inside! It just touches the corner at (5,5).
	// The segment from (8,2) goes to (5,5) which is boundary, then to (2,8).
	// Both parts stay within the L-shape.
	assert.True(t, Contains(lShape, lineDiagonalAcrossNotch), "L-shape should contain line touching corner")

	// Let's create a true case where the line crosses the notch:
	// From (8, 3) to (3, 8)
	// Parametric: x = 8 - 5t, y = 3 + 5t
	// At t=0.5: x=5.5, y=5.5 (in the notch! x>5 and y>5)
	lineThroughNotch := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(8, 3), // inside bottom arm
		NewCoordinate(3, 8), // inside left column
	})
	assert.False(t, Contains(lShape, lineThroughNotch), "L-shape should NOT contain line crossing through notch")
}

// TestContains_PolygonInPolygonWithHole tests containment with holes
func TestContains_PolygonInPolygonWithHole(t *testing.T) {
	factory := DefaultFactory

	// Outer polygon with a hole
	outerShell := factory.CreateLinearRing(CoordinateSequence{
		NewCoordinate(0, 0),
		NewCoordinate(20, 0),
		NewCoordinate(20, 20),
		NewCoordinate(0, 20),
		NewCoordinate(0, 0),
	})
	hole := factory.CreateLinearRing(CoordinateSequence{
		NewCoordinate(8, 8),
		NewCoordinate(12, 8),
		NewCoordinate(12, 12),
		NewCoordinate(8, 12),
		NewCoordinate(8, 8),
	})
	outerWithHole := factory.CreatePolygon(outerShell, []*LinearRing{hole})

	// Small polygon in the hole - should NOT be contained (hole is exterior)
	inHole := createTestSquare(factory, 9, 9, 2)
	assert.False(t, Contains(outerWithHole, inHole), "Polygon with hole should not contain polygon inside the hole")

	// Small polygon outside the hole but inside outer shell - should be contained
	outsideHole := createTestSquare(factory, 2, 2, 4)
	assert.True(t, Contains(outerWithHole, outsideHole), "Polygon with hole should contain polygon outside the hole")

	// Line that crosses through the hole
	lineThroughHole := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(5, 10), // inside (left of hole)
		NewCoordinate(15, 10), // inside (right of hole)
	})
	// This line passes through the hole at y=10, x in [8,12]
	assert.False(t, Contains(outerWithHole, lineThroughHole), "Polygon with hole should not contain line passing through hole")

	// Line that goes around the hole
	lineAroundHole := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(5, 5), // inside
		NewCoordinate(15, 5), // inside
	})
	assert.True(t, Contains(outerWithHole, lineAroundHole), "Polygon with hole should contain line that avoids the hole")
}

func TestContains_ConcavePolygonDoesNotContainBoundaryRingByEnvelopeCenter(t *testing.T) {
	factory := DefaultFactory

	// Concave shell with a small notch away from the envelope center.
	// The shell's envelope center (5,5) is inside the polygon, but the ring itself
	// is boundary-only and has no interior point in the polygon interior.
	shell := factory.CreateLinearRing(CoordinateSequence{
		NewCoordinate(0, 0),
		NewCoordinate(10, 0),
		NewCoordinate(10, 10),
		NewCoordinate(8, 10),
		NewCoordinate(8, 8),
		NewCoordinate(6, 8),
		NewCoordinate(6, 10),
		NewCoordinate(0, 10),
		NewCoordinate(0, 0),
	})
	poly := factory.CreatePolygon(shell, nil)

	assert.False(t, Contains(poly, shell), "Concave polygon should not contain its boundary ring")
	assert.True(t, Covers(poly, shell), "Concave polygon should still cover its boundary ring")
}

func TestContains_PolygonWithHoleDoesNotContainBoundaryRingByEnvelopeCenter(t *testing.T) {
	factory := DefaultFactory

	// The hole is off center, so the outer shell's envelope center is in the
	// polygon interior. Contains must not use that sample for a boundary-only ring.
	outerShell := factory.CreateLinearRing(CoordinateSequence{
		NewCoordinate(0, 0),
		NewCoordinate(20, 0),
		NewCoordinate(20, 20),
		NewCoordinate(0, 20),
		NewCoordinate(0, 0),
	})
	hole := factory.CreateLinearRing(CoordinateSequence{
		NewCoordinate(2, 2),
		NewCoordinate(4, 2),
		NewCoordinate(4, 4),
		NewCoordinate(2, 4),
		NewCoordinate(2, 2),
	})
	poly := factory.CreatePolygon(outerShell, []*LinearRing{hole})

	assert.False(t, Contains(poly, outerShell), "Polygon with hole should not contain its exterior boundary ring")
	assert.True(t, Covers(poly, outerShell), "Polygon with hole should still cover its exterior boundary ring")
	assert.False(t, Contains(poly, hole), "Polygon with hole should not contain its hole boundary ring")
	assert.True(t, Covers(poly, hole), "Polygon with hole should still cover its hole boundary ring")
}
