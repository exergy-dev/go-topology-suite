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
