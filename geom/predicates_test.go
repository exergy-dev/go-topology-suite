package geom

import (
	"testing"
)

func TestIntersectsPointPoint(t *testing.T) {
	factory := DefaultFactory

	p1 := factory.CreatePoint(1, 2)
	p2 := factory.CreatePoint(1, 2)
	p3 := factory.CreatePoint(3, 4)

	if !Intersects(p1, p2) {
		t.Error("Expected equal points to intersect")
	}

	if Intersects(p1, p3) {
		t.Error("Expected different points not to intersect")
	}
}

func TestIntersectsPointLine(t *testing.T) {
	factory := DefaultFactory

	p := factory.CreatePoint(1, 1)
	line := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(0, 0),
		NewCoordinate(2, 2),
	})

	if !Intersects(p, line) {
		t.Error("Expected point on line to intersect")
	}

	p2 := factory.CreatePoint(0, 1)
	if Intersects(p2, line) {
		t.Error("Expected point not on line to not intersect")
	}
}

func TestIntersectsPointPolygon(t *testing.T) {
	factory := DefaultFactory

	poly := createTestSquare(factory, 0, 0, 10)

	// Point inside
	pInside := factory.CreatePoint(5, 5)
	if !Intersects(pInside, poly) {
		t.Error("Expected point inside polygon to intersect")
	}

	// Point outside
	pOutside := factory.CreatePoint(15, 15)
	if Intersects(pOutside, poly) {
		t.Error("Expected point outside polygon not to intersect")
	}

	// Point on boundary
	pBoundary := factory.CreatePoint(0, 5)
	if !Intersects(pBoundary, poly) {
		t.Error("Expected point on boundary to intersect")
	}
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

	if !Intersects(line1, line2) {
		t.Error("Expected crossing lines to intersect")
	}

	// Parallel lines
	line3 := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(0, 5),
		NewCoordinate(10, 15),
	})
	if Intersects(line1, line3) {
		t.Error("Expected parallel lines not to intersect")
	}
}

func TestIntersectsLinePolygon(t *testing.T) {
	factory := DefaultFactory

	poly := createTestSquare(factory, 0, 0, 10)

	// Line crossing polygon
	lineCrossing := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(-5, 5),
		NewCoordinate(15, 5),
	})
	if !Intersects(lineCrossing, poly) {
		t.Error("Expected line crossing polygon to intersect")
	}

	// Line outside polygon
	lineOutside := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(20, 20),
		NewCoordinate(30, 30),
	})
	if Intersects(lineOutside, poly) {
		t.Error("Expected line outside polygon not to intersect")
	}

	// Line inside polygon
	lineInside := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(2, 2),
		NewCoordinate(8, 8),
	})
	if !Intersects(lineInside, poly) {
		t.Error("Expected line inside polygon to intersect")
	}
}

func TestIntersectsPolygonPolygon(t *testing.T) {
	factory := DefaultFactory

	poly1 := createTestSquare(factory, 0, 0, 10)
	poly2 := createTestSquare(factory, 5, 5, 10)

	if !Intersects(poly1, poly2) {
		t.Error("Expected overlapping polygons to intersect")
	}

	// Non-overlapping polygons
	poly3 := createTestSquare(factory, 20, 20, 10)
	if Intersects(poly1, poly3) {
		t.Error("Expected non-overlapping polygons not to intersect")
	}
}

func TestContainsPointInPolygon(t *testing.T) {
	factory := DefaultFactory

	poly := createTestSquare(factory, 0, 0, 10)

	// Point inside
	pInside := factory.CreatePoint(5, 5)
	if !Contains(poly, pInside) {
		t.Error("Expected polygon to contain point inside")
	}

	// Point outside
	pOutside := factory.CreatePoint(15, 15)
	if Contains(poly, pOutside) {
		t.Error("Expected polygon not to contain point outside")
	}
}

func TestContainsPolygonInPolygon(t *testing.T) {
	factory := DefaultFactory

	outer := createTestSquare(factory, 0, 0, 20)
	inner := createTestSquare(factory, 5, 5, 5)

	if !Contains(outer, inner) {
		t.Error("Expected outer polygon to contain inner polygon")
	}

	if Contains(inner, outer) {
		t.Error("Expected inner polygon not to contain outer polygon")
	}
}

func TestWithin(t *testing.T) {
	factory := DefaultFactory

	outer := createTestSquare(factory, 0, 0, 20)
	inner := createTestSquare(factory, 5, 5, 5)

	if !Within(inner, outer) {
		t.Error("Expected inner polygon to be within outer polygon")
	}

	if Within(outer, inner) {
		t.Error("Expected outer polygon not to be within inner polygon")
	}
}

func TestCovers(t *testing.T) {
	factory := DefaultFactory

	outer := createTestSquare(factory, 0, 0, 20)
	inner := createTestSquare(factory, 5, 5, 5)

	if !Covers(outer, inner) {
		t.Error("Expected outer polygon to cover inner polygon")
	}

	if Covers(inner, outer) {
		t.Error("Expected inner polygon not to cover outer polygon")
	}
}

func TestCoveredBy(t *testing.T) {
	factory := DefaultFactory

	outer := createTestSquare(factory, 0, 0, 20)
	inner := createTestSquare(factory, 5, 5, 5)

	if !CoveredBy(inner, outer) {
		t.Error("Expected inner polygon to be covered by outer polygon")
	}

	if CoveredBy(outer, inner) {
		t.Error("Expected outer polygon not to be covered by inner polygon")
	}
}

func TestDisjoint(t *testing.T) {
	factory := DefaultFactory

	poly1 := createTestSquare(factory, 0, 0, 10)
	poly2 := createTestSquare(factory, 20, 20, 10)
	poly3 := createTestSquare(factory, 5, 5, 10)

	if !Disjoint(poly1, poly2) {
		t.Error("Expected non-overlapping polygons to be disjoint")
	}

	if Disjoint(poly1, poly3) {
		t.Error("Expected overlapping polygons not to be disjoint")
	}
}

func TestTouches(t *testing.T) {
	factory := DefaultFactory

	// Two squares sharing an edge
	poly1 := createTestSquare(factory, 0, 0, 10)
	poly2 := createTestSquare(factory, 10, 0, 10)

	if !Touches(poly1, poly2) {
		t.Error("Expected adjacent polygons to touch")
	}

	// Overlapping polygons don't touch
	poly3 := createTestSquare(factory, 5, 0, 10)
	if Touches(poly1, poly3) {
		t.Error("Expected overlapping polygons not to touch")
	}
}

func TestCrosses(t *testing.T) {
	factory := DefaultFactory

	poly := createTestSquare(factory, 0, 0, 10)

	// Line crossing polygon
	lineCrossing := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(-5, 5),
		NewCoordinate(15, 5),
	})

	if !Crosses(lineCrossing, poly) {
		t.Error("Expected line to cross polygon")
	}

	// Line inside polygon
	lineInside := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(2, 2),
		NewCoordinate(8, 8),
	})
	if Crosses(lineInside, poly) {
		t.Error("Expected line inside polygon not to cross")
	}
}

func TestOverlaps(t *testing.T) {
	factory := DefaultFactory

	poly1 := createTestSquare(factory, 0, 0, 10)
	poly2 := createTestSquare(factory, 5, 5, 10)

	if !Overlaps(poly1, poly2) {
		t.Error("Expected partially overlapping polygons to overlap")
	}

	// One contains the other - not overlap
	outer := createTestSquare(factory, 0, 0, 20)
	inner := createTestSquare(factory, 5, 5, 5)
	if Overlaps(outer, inner) {
		t.Error("Expected containment not to be overlap")
	}

	// Different dimensions
	line := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(0, 0),
		NewCoordinate(10, 10),
	})
	if Overlaps(poly1, line) {
		t.Error("Expected different dimension geometries not to overlap")
	}
}

func TestEquals(t *testing.T) {
	factory := DefaultFactory

	poly1 := createTestSquare(factory, 0, 0, 10)
	poly2 := createTestSquare(factory, 0, 0, 10)
	poly3 := createTestSquare(factory, 5, 5, 10)

	if !Equals(poly1, poly2) {
		t.Error("Expected identical polygons to be equal")
	}

	if Equals(poly1, poly3) {
		t.Error("Expected different polygons not to be equal")
	}
}

func TestEqualsEmptyGeometries(t *testing.T) {
	factory := DefaultFactory

	emptyPoint1 := factory.CreatePointEmpty()
	emptyPoint2 := factory.CreatePointEmpty()

	if !Equals(emptyPoint1, emptyPoint2) {
		t.Error("Expected empty points to be equal")
	}
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
	if !Intersects(mp, p1) {
		t.Error("Expected multipolygon to intersect with point in first polygon")
	}

	// Point in second polygon
	p2 := factory.CreatePoint(25, 25)
	if !Intersects(mp, p2) {
		t.Error("Expected multipolygon to intersect with point in second polygon")
	}

	// Point outside both
	p3 := factory.CreatePoint(50, 50)
	if Intersects(mp, p3) {
		t.Error("Expected multipolygon not to intersect with point outside")
	}
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
	if !Intersects(gc, p) {
		t.Error("Expected geometry collection to intersect with point inside polygon")
	}
}

func TestPredicatesEnvelopeRejection(t *testing.T) {
	factory := DefaultFactory

	// These geometries don't even have overlapping envelopes
	poly1 := createTestSquare(factory, 0, 0, 10)
	poly2 := createTestSquare(factory, 100, 100, 10)

	// Should be quickly rejected by envelope check
	if Intersects(poly1, poly2) {
		t.Error("Expected disjoint envelope geometries not to intersect")
	}

	if Contains(poly1, poly2) {
		t.Error("Expected disjoint envelope geometries: no containment")
	}

	if Touches(poly1, poly2) {
		t.Error("Expected disjoint envelope geometries not to touch")
	}
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

	if !Crosses(line1, line2) {
		t.Error("Expected crossing lines to cross")
	}

	// Parallel lines don't cross
	line3 := factory.CreateLineString(CoordinateSequence{
		NewCoordinate(0, 5),
		NewCoordinate(10, 15),
	})
	if Crosses(line1, line3) {
		t.Error("Expected parallel lines not to cross")
	}
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
