package geom_test

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Helper factories
// ---------------------------------------------------------------------------

func square(x, y, size float64) *geom.Polygon {
	return geom.NewPolygon(
		mustLinearRingXY(x, y, x+size, y, x+size, y+size, x, y+size, x, y),
		nil,
	)
}

func line(x1, y1, x2, y2 float64) *geom.LineString {
	return mustLineStringXY(x1, y1, x2, y2)
}

func point(x, y float64) *geom.Point {
	return geom.NewPoint(x, y)
}

// ---------------------------------------------------------------------------
// Intersects: LinearRing arguments
// ---------------------------------------------------------------------------

func TestIntersects_LinearRingAsArgument(t *testing.T) {
	// NOTE: LinearRing is a distinct Go type from *LineString (it embeds *LineString).
	// The intersectsImpl type switch does NOT have a *LinearRing case, so bare
	// LinearRing arguments fall through and return false. These tests document
	// this current behavior rather than the ideal behavior.

	ring := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)

	t.Run("PointOnRing_unmatched", func(t *testing.T) {
		p := point(5, 0)
		// LinearRing as g2 is not dispatched by pointIntersects, so returns false
		assert.False(t, geom.Intersects(p, ring),
			"LinearRing is not dispatched by intersectsImpl (known gap)")
	})

	t.Run("RingAsG1_unmatched", func(t *testing.T) {
		p := point(5, 0)
		// LinearRing as g1 also falls through the switch
		assert.False(t, geom.Intersects(ring, p),
			"LinearRing as g1 is not dispatched (known gap)")
	})

	t.Run("RingVsDisjointRing_both_unmatched", func(t *testing.T) {
		ring2 := mustLinearRingXY(20, 20, 30, 20, 30, 30, 20, 30, 20, 20)
		// Both are LinearRing, neither dispatched - correctly returns false
		assert.False(t, geom.Intersects(ring, ring2),
			"Two disjoint rings correctly return false")
	})

	t.Run("RingUsedAsPolygonShellWorksViaPolygon", func(t *testing.T) {
		// When the ring is used as a polygon shell, predicates work correctly
		poly := geom.NewPolygon(ring, nil)
		p := point(5, 5)
		assert.True(t, geom.Intersects(p, poly),
			"LinearRing used as polygon shell works through Polygon dispatch")
	})
}

// ---------------------------------------------------------------------------
// Contains: MultiLineString, MultiPolygon, GeometryCollection
// ---------------------------------------------------------------------------

func TestContains_PolygonContainsMultiLineString(t *testing.T) {
	poly := square(0, 0, 20)

	// MultiLineString fully inside
	mls := geom.NewMultiLineString([]*geom.LineString{
		mustLineStringXY(2, 2, 8, 8),
		mustLineStringXY(12, 12, 18, 18),
	})
	assert.True(t, geom.Contains(poly, mls),
		"Polygon should contain MultiLineString fully inside")

	// One component sticks out
	mlsPartOut := geom.NewMultiLineString([]*geom.LineString{
		mustLineStringXY(2, 2, 8, 8),
		mustLineStringXY(12, 12, 25, 25), // extends outside
	})
	assert.False(t, geom.Contains(poly, mlsPartOut),
		"Polygon should NOT contain MultiLineString with a component outside")
}

func TestContains_PolygonContainsMultiPolygon(t *testing.T) {
	outer := square(0, 0, 50)

	mp := geom.NewMultiPolygon([]*geom.Polygon{
		square(5, 5, 10),
		square(25, 25, 10),
	})
	assert.True(t, geom.Contains(outer, mp),
		"Large polygon should contain MultiPolygon wholly inside")

	mpPartOut := geom.NewMultiPolygon([]*geom.Polygon{
		square(5, 5, 10),
		square(45, 45, 10), // extends past outer
	})
	assert.False(t, geom.Contains(outer, mpPartOut),
		"Polygon should NOT contain MultiPolygon partially outside")
}

func TestContains_PolygonContainsGeometryCollection(t *testing.T) {
	poly := square(0, 0, 20)

	gcInside := geom.NewGeometryCollection([]geom.Geometry{
		point(5, 5),
		line(2, 2, 8, 8),
		square(3, 3, 4),
	})
	assert.True(t, geom.Contains(poly, gcInside),
		"Polygon should contain GeometryCollection fully inside")

	gcPartOut := geom.NewGeometryCollection([]geom.Geometry{
		point(25, 25), // outside
		line(2, 2, 8, 8),
	})
	assert.False(t, geom.Contains(poly, gcPartOut),
		"Polygon should NOT contain GeometryCollection with component outside")
}

// ---------------------------------------------------------------------------
// Crosses: Line vs Polygon, Line vs Line
// ---------------------------------------------------------------------------

func TestCrosses_LineCrossesPolygon(t *testing.T) {
	poly := square(0, 0, 10)

	t.Run("LineCrossingThroughPolygon", func(t *testing.T) {
		ls := line(-5, 5, 15, 5)
		assert.True(t, geom.Crosses(ls, poly),
			"Line passing through polygon should cross it")
	})

	t.Run("LineEntirelyInside", func(t *testing.T) {
		ls := line(2, 2, 8, 8)
		assert.False(t, geom.Crosses(ls, poly),
			"Line entirely inside polygon should NOT cross it")
	})

	t.Run("LineEntirelyOutside", func(t *testing.T) {
		ls := line(20, 20, 30, 30)
		assert.False(t, geom.Crosses(ls, poly),
			"Line entirely outside polygon should NOT cross it")
	})

	t.Run("LineTouchingBoundary", func(t *testing.T) {
		// Line along the bottom edge
		ls := line(-5, 0, 15, 0)
		// This line is on the boundary - does not enter interior AND exterior properly
		// in the way Crosses requires (line must be partly interior, partly exterior)
		// Actually, this line enters the polygon boundary and the exterior,
		// but per OGC, line crossing area means some point of line is in interior
		// and some in exterior. Here the line overlaps the boundary only.
		// So Crosses should be false.
		assert.False(t, geom.Crosses(ls, poly),
			"Line along polygon boundary should NOT cross the polygon")
	})
}

func TestCrosses_PolygonVsLine(t *testing.T) {
	// Crosses(area, line) should give the same result as Crosses(line, area)
	poly := square(0, 0, 10)
	ls := line(-5, 5, 15, 5)
	assert.True(t, geom.Crosses(poly, ls),
		"Crosses(polygon, line) should be symmetric with Crosses(line, polygon)")
}

func TestCrosses_LineVsLine(t *testing.T) {
	t.Run("ProperCrossing", func(t *testing.T) {
		l1 := line(0, 0, 10, 10)
		l2 := line(0, 10, 10, 0)
		assert.True(t, geom.Crosses(l1, l2),
			"Two properly crossing lines should cross")
	})

	t.Run("SharedEndpointNoProperCross", func(t *testing.T) {
		l1 := line(0, 0, 10, 10)
		l2 := line(10, 10, 20, 0)
		assert.False(t, geom.Crosses(l1, l2),
			"Lines sharing only an endpoint should NOT cross")
	})

	t.Run("ParallelLines", func(t *testing.T) {
		l1 := line(0, 0, 10, 0)
		l2 := line(0, 5, 10, 5)
		assert.False(t, geom.Crosses(l1, l2),
			"Parallel lines should NOT cross")
	})
}

func TestCrosses_PointVsPoint_AreaVsArea_ReturnFalse(t *testing.T) {
	// Per OGC: Point/Point and Area/Area cannot cross
	assert.False(t, geom.Crosses(point(0, 0), point(1, 1)))
	assert.False(t, geom.Crosses(square(0, 0, 10), square(5, 5, 10)))
}

// ---------------------------------------------------------------------------
// Touches: various geometry type pairs
// ---------------------------------------------------------------------------

func TestTouches_PointTouchesLineBoundary(t *testing.T) {
	ls := line(0, 0, 10, 0)
	// Endpoint of line is boundary
	p := point(0, 0)
	assert.True(t, geom.Touches(p, ls),
		"Point at line endpoint should touch the line")
}

func TestTouches_PointOnLineInteriorDoesNotTouch(t *testing.T) {
	ls := line(0, 0, 10, 0)
	p := point(5, 0) // interior of line, not endpoint
	// Touches requires boundary intersection but no interior-interior.
	// Point's "interior" is the point itself; line's interior includes (5,0).
	// So there IS an interior-interior intersection -> Touches is false.
	assert.False(t, geom.Touches(p, ls),
		"Point in line interior should NOT be a touch")
}

func TestTouches_PointTouchesPolygonBoundary(t *testing.T) {
	poly := square(0, 0, 10)
	p := point(5, 0) // on boundary
	assert.True(t, geom.Touches(p, poly),
		"Point on polygon boundary should touch")
}

func TestTouches_PointInsidePolygonDoesNotTouch(t *testing.T) {
	poly := square(0, 0, 10)
	p := point(5, 5) // interior
	assert.False(t, geom.Touches(p, poly),
		"Point in polygon interior should NOT be a touch")
}

func TestTouches_LineTouchesPolygonBoundary(t *testing.T) {
	poly := square(0, 0, 10)
	// Line sits entirely on top boundary
	ls := line(2, 10, 8, 10)
	assert.True(t, geom.Touches(ls, poly),
		"Line along polygon boundary should touch")
}

func TestTouches_LineCrossesPolygonDoesNotTouch(t *testing.T) {
	poly := square(0, 0, 10)
	ls := line(-5, 5, 15, 5)
	assert.False(t, geom.Touches(ls, poly),
		"Line crossing polygon interior should NOT be a touch")
}

func TestTouches_PolygonsShareEdge(t *testing.T) {
	poly1 := square(0, 0, 10)
	poly2 := square(10, 0, 10) // shares right edge of poly1
	assert.True(t, geom.Touches(poly1, poly2),
		"Polygons sharing an edge should touch")
}

func TestTouches_PolygonsShareCorner(t *testing.T) {
	poly1 := square(0, 0, 10)
	poly2 := square(10, 10, 10) // shares only the corner (10,10)
	assert.True(t, geom.Touches(poly1, poly2),
		"Polygons sharing only a corner should touch")
}

func TestTouches_OverlappingPolygonsDoNotTouch(t *testing.T) {
	poly1 := square(0, 0, 10)
	poly2 := square(5, 5, 10) // overlaps
	assert.False(t, geom.Touches(poly1, poly2),
		"Overlapping polygons should NOT touch")
}

// ---------------------------------------------------------------------------
// Within as complement of Contains
// ---------------------------------------------------------------------------

func TestWithin_PointWithinPolygon(t *testing.T) {
	poly := square(0, 0, 10)
	p := point(5, 5)
	assert.True(t, geom.Within(p, poly),
		"Point inside polygon should be within it")
	assert.False(t, geom.Within(poly, p),
		"Polygon should NOT be within a point")
}

func TestWithin_LineWithinPolygon(t *testing.T) {
	poly := square(0, 0, 20)
	ls := line(2, 2, 18, 18)
	assert.True(t, geom.Within(ls, poly),
		"Line inside polygon should be within it")
}

func TestWithin_PolygonWithinLargerPolygon(t *testing.T) {
	outer := square(0, 0, 50)
	inner := square(10, 10, 10)
	assert.True(t, geom.Within(inner, outer))
	assert.False(t, geom.Within(outer, inner))
}

func TestWithin_Symmetry_with_Contains(t *testing.T) {
	outer := square(0, 0, 50)
	inner := square(10, 10, 10)
	assert.Equal(t, geom.Contains(outer, inner), geom.Within(inner, outer),
		"Within(A,B) must equal Contains(B,A)")
}

// ---------------------------------------------------------------------------
// CoveredBy predicate
// ---------------------------------------------------------------------------

func TestCoveredBy_PointOnBoundaryIsCoveredBy(t *testing.T) {
	poly := square(0, 0, 10)
	p := point(0, 5)
	assert.True(t, geom.CoveredBy(p, poly),
		"Point on polygon boundary should be covered by the polygon")
}

func TestCoveredBy_PointInsideIsCoveredBy(t *testing.T) {
	poly := square(0, 0, 10)
	p := point(5, 5)
	assert.True(t, geom.CoveredBy(p, poly),
		"Point inside polygon should be covered by it")
}

func TestCoveredBy_PointOutsideIsNotCoveredBy(t *testing.T) {
	poly := square(0, 0, 10)
	p := point(20, 20)
	assert.False(t, geom.CoveredBy(p, poly),
		"Point outside polygon should NOT be covered by it")
}

func TestCoveredBy_LineOnBoundary(t *testing.T) {
	poly := square(0, 0, 10)
	ls := line(0, 0, 10, 0) // along bottom edge
	assert.True(t, geom.CoveredBy(ls, poly),
		"Line along polygon boundary should be covered by the polygon")
}

func TestCoveredBy_Symmetry_with_Covers(t *testing.T) {
	outer := square(0, 0, 50)
	inner := square(10, 10, 10)
	assert.Equal(t, geom.Covers(outer, inner), geom.CoveredBy(inner, outer),
		"CoveredBy(A,B) must equal Covers(B,A)")
}

// ---------------------------------------------------------------------------
// Empty geometry handling for all predicates
// ---------------------------------------------------------------------------

func TestPredicates_EmptyGeometry(t *testing.T) {
	poly := square(0, 0, 10)
	ls := line(0, 0, 10, 10)
	ep := geom.NewPointEmpty()
	els := geom.NewLineStringEmpty()
	ePoly := geom.NewPolygonEmpty()

	empties := []geom.Geometry{ep, els, ePoly}
	nonEmpties := []geom.Geometry{poly, ls, point(5, 5)}

	for _, empty := range empties {
		for _, g := range nonEmpties {
			t.Run("Intersects_empty_"+empty.GeometryType()+"_vs_"+g.GeometryType(), func(t *testing.T) {
				assert.False(t, geom.Intersects(empty, g),
					"Empty geometry should not intersect anything")
				assert.False(t, geom.Intersects(g, empty),
					"Nothing should intersect an empty geometry")
			})

			t.Run("Contains_empty_"+empty.GeometryType()+"_vs_"+g.GeometryType(), func(t *testing.T) {
				// The implementation rejects empty geometries via envelope
				// (ContainsEnvelope returns false when either envelope is null).
				assert.False(t, geom.Contains(g, empty),
					"Contains returns false for empty arg due to envelope rejection")
				assert.False(t, geom.Contains(empty, g),
					"Empty geometry should not contain anything")
			})

			t.Run("Covers_empty_"+empty.GeometryType()+"_vs_"+g.GeometryType(), func(t *testing.T) {
				// Same envelope rejection as Contains
				assert.False(t, geom.Covers(g, empty),
					"Covers returns false for empty arg due to envelope rejection")
				assert.False(t, geom.Covers(empty, g),
					"Empty geometry should not cover anything")
			})

			t.Run("Touches_empty_"+empty.GeometryType()+"_vs_"+g.GeometryType(), func(t *testing.T) {
				assert.False(t, geom.Touches(empty, g),
					"Empty geometry should not touch anything")
				assert.False(t, geom.Touches(g, empty),
					"Nothing should touch an empty geometry")
			})

			t.Run("Crosses_empty_"+empty.GeometryType()+"_vs_"+g.GeometryType(), func(t *testing.T) {
				assert.False(t, geom.Crosses(empty, g),
					"Empty geometry should not cross anything")
				assert.False(t, geom.Crosses(g, empty),
					"Nothing should cross an empty geometry")
			})
		}
	}
}

func TestDisjoint_EmptyGeometries(t *testing.T) {
	ep := geom.NewPointEmpty()
	poly := square(0, 0, 10)
	// Empty geometries have no points in common with anything
	assert.True(t, geom.Disjoint(ep, poly),
		"Empty geometry should be disjoint from everything")
	assert.True(t, geom.Disjoint(poly, ep),
		"Everything should be disjoint from an empty geometry")
}

func TestEquals_EmptyGeometriesSameType(t *testing.T) {
	ep1 := geom.NewPointEmpty()
	ep2 := geom.NewPointEmpty()
	assert.True(t, geom.Equals(ep1, ep2),
		"Two empty points should be equal")

	els1 := geom.NewLineStringEmpty()
	els2 := geom.NewLineStringEmpty()
	assert.True(t, geom.Equals(els1, els2),
		"Two empty linestrings should be equal")
}

func TestEquals_EmptyGeometriesDifferentType(t *testing.T) {
	ep := geom.NewPointEmpty()
	els := geom.NewLineStringEmpty()
	assert.False(t, geom.Equals(ep, els),
		"Empty geometries of different types should NOT be equal")
}

// ---------------------------------------------------------------------------
// Intersects: multi-geometry dispatches
// ---------------------------------------------------------------------------

func TestIntersects_MultiPointVsPolygon(t *testing.T) {
	poly := square(0, 0, 10)
	mp := geom.NewMultiPoint([]*geom.Point{
		point(5, 5),   // inside
		point(20, 20), // outside
	})
	assert.True(t, geom.Intersects(mp, poly),
		"MultiPoint with one point inside polygon should intersect")

	mpOut := geom.NewMultiPoint([]*geom.Point{
		point(20, 20),
		point(30, 30),
	})
	assert.False(t, geom.Intersects(mpOut, poly),
		"MultiPoint entirely outside polygon should not intersect")
}

func TestIntersects_MultiLineStringVsPolygon(t *testing.T) {
	poly := square(0, 0, 10)
	mls := geom.NewMultiLineString([]*geom.LineString{
		line(20, 20, 30, 30), // outside
		line(2, 2, 8, 8),    // inside
	})
	assert.True(t, geom.Intersects(mls, poly),
		"MultiLineString with one line inside polygon should intersect")
}

func TestIntersects_MultiPolygonVsPoint(t *testing.T) {
	mp := geom.NewMultiPolygon([]*geom.Polygon{
		square(0, 0, 10),
		square(20, 20, 10),
	})
	assert.True(t, geom.Intersects(mp, point(5, 5)))
	assert.True(t, geom.Intersects(mp, point(25, 25)))
	assert.False(t, geom.Intersects(mp, point(15, 15)))
}

func TestIntersects_GeometryCollectionVsPoint(t *testing.T) {
	gc := geom.NewGeometryCollection([]geom.Geometry{
		square(0, 0, 10),
		line(20, 20, 30, 30),
	})
	assert.True(t, geom.Intersects(gc, point(5, 5)))
	assert.False(t, geom.Intersects(gc, point(50, 50)))
}

// ---------------------------------------------------------------------------
// Intersects: Point vs MultiPoint, MultiLineString, MultiPolygon, GC
// ---------------------------------------------------------------------------

func TestIntersects_PointVsMultiPoint(t *testing.T) {
	p := point(3, 4)
	mp := geom.NewMultiPoint([]*geom.Point{
		point(1, 2),
		point(3, 4),
	})
	assert.True(t, geom.Intersects(p, mp))
}

func TestIntersects_PointVsMultiLineString(t *testing.T) {
	p := point(5, 0)
	mls := geom.NewMultiLineString([]*geom.LineString{
		line(0, 0, 10, 0), // p is on this line
		line(0, 5, 10, 5),
	})
	assert.True(t, geom.Intersects(p, mls))
}

func TestIntersects_PointVsMultiPolygon(t *testing.T) {
	p := point(5, 5)
	mp := geom.NewMultiPolygon([]*geom.Polygon{
		square(0, 0, 10),
		square(20, 20, 10),
	})
	assert.True(t, geom.Intersects(p, mp))

	pOut := point(15, 15)
	assert.False(t, geom.Intersects(pOut, mp))
}

func TestIntersects_PointVsGeometryCollection(t *testing.T) {
	p := point(5, 5)
	gc := geom.NewGeometryCollection([]geom.Geometry{
		square(0, 0, 10),
		point(20, 20),
	})
	assert.True(t, geom.Intersects(p, gc))
}

// ---------------------------------------------------------------------------
// Intersects: LineString vs MultiPoint, MultiLineString, MultiPolygon, GC
// ---------------------------------------------------------------------------

func TestIntersects_LineVsMultiPoint(t *testing.T) {
	ls := line(0, 0, 10, 0)
	mp := geom.NewMultiPoint([]*geom.Point{
		point(5, 0),   // on the line
		point(20, 20), // off
	})
	assert.True(t, geom.Intersects(ls, mp))

	mpOff := geom.NewMultiPoint([]*geom.Point{
		point(0, 5),
		point(20, 20),
	})
	assert.False(t, geom.Intersects(ls, mpOff))
}

func TestIntersects_LineVsMultiLineString(t *testing.T) {
	ls := line(0, 0, 10, 10)
	mls := geom.NewMultiLineString([]*geom.LineString{
		line(0, 10, 10, 0), // crosses
		line(20, 20, 30, 30),
	})
	assert.True(t, geom.Intersects(ls, mls))
}

func TestIntersects_LineVsMultiPolygon(t *testing.T) {
	ls := line(5, 5, 25, 25)
	mp := geom.NewMultiPolygon([]*geom.Polygon{
		square(0, 0, 10),
		square(20, 20, 10),
	})
	assert.True(t, geom.Intersects(ls, mp))
}

func TestIntersects_LineVsGeometryCollection(t *testing.T) {
	ls := line(0, 0, 10, 10)
	gc := geom.NewGeometryCollection([]geom.Geometry{
		point(5, 5), // on the line
	})
	assert.True(t, geom.Intersects(ls, gc))
}

// ---------------------------------------------------------------------------
// Intersects: Polygon vs MultiPoint, MultiLineString, MultiPolygon, GC
// ---------------------------------------------------------------------------

func TestIntersects_PolygonVsMultiPoint(t *testing.T) {
	poly := square(0, 0, 10)
	mp := geom.NewMultiPoint([]*geom.Point{
		point(5, 5), // inside
		point(20, 20),
	})
	assert.True(t, geom.Intersects(poly, mp))
}

func TestIntersects_PolygonVsMultiLineString(t *testing.T) {
	poly := square(0, 0, 10)
	mls := geom.NewMultiLineString([]*geom.LineString{
		line(2, 2, 8, 8), // inside
		line(20, 20, 30, 30),
	})
	assert.True(t, geom.Intersects(poly, mls))
}

func TestIntersects_PolygonVsMultiPolygon(t *testing.T) {
	poly := square(0, 0, 10)
	mp := geom.NewMultiPolygon([]*geom.Polygon{
		square(5, 5, 10),
		square(30, 30, 10),
	})
	assert.True(t, geom.Intersects(poly, mp))
}

func TestIntersects_PolygonVsGeometryCollection(t *testing.T) {
	poly := square(0, 0, 10)
	gc := geom.NewGeometryCollection([]geom.Geometry{
		point(5, 5),
	})
	assert.True(t, geom.Intersects(poly, gc))
}

// ---------------------------------------------------------------------------
// Contains: empty set handling
// ---------------------------------------------------------------------------

func TestContains_EmptySecondArgument(t *testing.T) {
	// The implementation performs an early envelope rejection:
	// ContainsEnvelope returns false when either envelope is null.
	// Empty geometries have null envelopes, so Contains returns false.
	poly := square(0, 0, 10)
	ep := geom.NewPointEmpty()
	els := geom.NewLineStringEmpty()
	ePoly := geom.NewPolygonEmpty()

	assert.False(t, geom.Contains(poly, ep),
		"Contains returns false for empty point (envelope rejection)")
	assert.False(t, geom.Contains(poly, els),
		"Contains returns false for empty linestring (envelope rejection)")
	assert.False(t, geom.Contains(poly, ePoly),
		"Contains returns false for empty polygon (envelope rejection)")
}

// ---------------------------------------------------------------------------
// Overlaps: additional dispatch paths
// ---------------------------------------------------------------------------

func TestOverlaps_PartiallyOverlappingLines(t *testing.T) {
	l1 := line(0, 0, 10, 0)
	l2 := line(5, 0, 15, 0) // overlapping segment on x-axis
	// Both have DimensionLine, they're not equal, they intersect, neither contains the other
	// This should be an overlap.
	assert.True(t, geom.Overlaps(l1, l2),
		"Partially overlapping lines should overlap")
}

func TestOverlaps_DisjointLines(t *testing.T) {
	l1 := line(0, 0, 10, 0)
	l2 := line(0, 5, 10, 5)
	assert.False(t, geom.Overlaps(l1, l2),
		"Disjoint lines should not overlap")
}

// ---------------------------------------------------------------------------
// Covers: additional tests
// ---------------------------------------------------------------------------

func TestCovers_PolygonCoversBoundaryPoint(t *testing.T) {
	poly := square(0, 0, 10)
	p := point(0, 5) // on boundary
	assert.True(t, geom.Covers(poly, p),
		"Polygon covers point on its boundary")
}

func TestCovers_PolygonCoversInteriorPoint(t *testing.T) {
	poly := square(0, 0, 10)
	p := point(5, 5)
	assert.True(t, geom.Covers(poly, p))
}

func TestCovers_PolygonDoesNotCoverExteriorPoint(t *testing.T) {
	poly := square(0, 0, 10)
	p := point(20, 20)
	assert.False(t, geom.Covers(poly, p))
}

// ---------------------------------------------------------------------------
// Equals: additional coverage
// ---------------------------------------------------------------------------

func TestEquals_IdenticalLines(t *testing.T) {
	l1 := line(0, 0, 10, 10)
	l2 := line(0, 0, 10, 10)
	assert.True(t, geom.Equals(l1, l2))
}

func TestEquals_DifferentTypesFalse(t *testing.T) {
	p := point(0, 0)
	ls := line(0, 0, 10, 10)
	assert.False(t, geom.Equals(p, ls),
		"Different geometry types should not be equal")
}

func TestEquals_NonEmptyVsEmptyFalse(t *testing.T) {
	p := point(0, 0)
	ep := geom.NewPointEmpty()
	assert.False(t, geom.Equals(p, ep))
	assert.False(t, geom.Equals(ep, p))
}

// ---------------------------------------------------------------------------
// MultiPoint empty edge cases
// ---------------------------------------------------------------------------

func TestIntersects_EmptyMultiPoint(t *testing.T) {
	mp := geom.NewMultiPointEmpty()
	poly := square(0, 0, 10)
	assert.False(t, geom.Intersects(mp, poly))
	assert.False(t, geom.Intersects(poly, mp))
}

func TestIntersects_EmptyMultiLineString(t *testing.T) {
	mls := geom.NewMultiLineStringEmpty()
	poly := square(0, 0, 10)
	assert.False(t, geom.Intersects(mls, poly))
	assert.False(t, geom.Intersects(poly, mls))
}

func TestIntersects_EmptyMultiPolygon(t *testing.T) {
	mp := geom.NewMultiPolygonEmpty()
	p := point(5, 5)
	assert.False(t, geom.Intersects(mp, p))
	assert.False(t, geom.Intersects(p, mp))
}

func TestIntersects_EmptyGeometryCollection(t *testing.T) {
	gc := geom.NewGeometryCollectionEmpty()
	p := point(5, 5)
	assert.False(t, geom.Intersects(gc, p))
	assert.False(t, geom.Intersects(p, gc))
}
