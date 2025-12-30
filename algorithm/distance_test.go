package algorithm_test

import (
	"math"
	"testing"

	"github.com/go-topology-suite/gts/algorithm"
	"github.com/go-topology-suite/gts/geom"
	"github.com/stretchr/testify/assert"
)

func TestDistancePointToLine(t *testing.T) {
	tests := []struct {
		name     string
		p        geom.Coordinate
		a        geom.Coordinate
		b        geom.Coordinate
		expected float64
	}{
		{
			name:     "PerpendicularDistance",
			p:        geom.NewCoordinate(5, 5),
			a:        geom.NewCoordinate(0, 0),
			b:        geom.NewCoordinate(10, 0),
			expected: 5,
		},
		{
			name:     "OnLine",
			p:        geom.NewCoordinate(5, 0),
			a:        geom.NewCoordinate(0, 0),
			b:        geom.NewCoordinate(10, 0),
			expected: 0,
		},
		{
			name:     "DegenerateLine",
			p:        geom.NewCoordinate(5, 5),
			a:        geom.NewCoordinate(0, 0),
			b:        geom.NewCoordinate(0, 0),
			expected: math.Sqrt(50),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.DistancePointToLine(tt.p, tt.a, tt.b)
			assert.InDelta(t, tt.expected, result, 0.001, "Expected %v", tt.expected)
		})
	}
}

func TestDistancePointToPolygon(t *testing.T) {
	poly := geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)

	tests := []struct {
		name     string
		p        geom.Coordinate
		expected float64
	}{
		{
			name:     "Inside",
			p:        geom.NewCoordinate(5, 5),
			expected: 0,
		},
		{
			name:     "Outside",
			p:        geom.NewCoordinate(15, 5),
			expected: 5,
		},
		{
			name:     "OnBoundary",
			p:        geom.NewCoordinate(0, 5),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.DistancePointToPolygon(tt.p, poly)
			assert.InDelta(t, tt.expected, result, 0.001, "Expected %v", tt.expected)
		})
	}

	// Test empty polygon
	emptyPoly := geom.NewPolygonEmpty()
	emptyDist := algorithm.DistancePointToPolygon(geom.NewCoordinate(5, 5), emptyPoly)
	assert.True(t, math.IsInf(emptyDist, 1), "Expected Inf for empty polygon")

	// Test polygon with hole
	shell := geom.NewLinearRingXY(0, 0, 20, 0, 20, 20, 0, 20, 0, 0)
	hole := geom.NewLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	polyWithHole := geom.NewPolygon(shell, []*geom.LinearRing{hole})

	// Point inside hole
	distInHole := algorithm.DistancePointToPolygon(geom.NewCoordinate(10, 10), polyWithHole)
	assert.NotEqual(t, float64(0), distInHole, "Expected non-zero distance for point in hole")
}

func TestDistancePointToGeometry(t *testing.T) {
	tests := []struct {
		name     string
		p        geom.Coordinate
		g        geom.Geometry
		expected float64
	}{
		{
			name:     "Point",
			p:        geom.NewCoordinate(0, 0),
			g:        geom.NewPoint(3, 4),
			expected: 5,
		},
		{
			name:     "EmptyPoint",
			p:        geom.NewCoordinate(0, 0),
			g:        geom.NewPointEmpty(),
			expected: math.Inf(1),
		},
		{
			name:     "LineString",
			p:        geom.NewCoordinate(5, 5),
			g:        geom.NewLineStringXY(0, 0, 10, 0),
			expected: 5,
		},
		{
			name:     "LinearRing",
			p:        geom.NewCoordinate(15, 5),
			g:        geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0),
			expected: 5,
		},
		{
			name:     "Polygon",
			p:        geom.NewCoordinate(15, 5),
			g:        geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
			expected: 5,
		},
		{
			name:     "MultiPoint",
			p:        geom.NewCoordinate(0, 0),
			g:        geom.NewMultiPoint([]*geom.Point{geom.NewPoint(3, 4), geom.NewPoint(6, 8)}),
			expected: 5,
		},
		{
			name:     "MultiLineString",
			p:        geom.NewCoordinate(5, 5),
			g:        geom.NewMultiLineString([]*geom.LineString{geom.NewLineStringXY(0, 0, 10, 0)}),
			expected: 5,
		},
		{
			name:     "MultiPolygon",
			p:        geom.NewCoordinate(5, 5),
			g:        geom.NewMultiPolygon([]*geom.Polygon{geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)}),
			expected: 0,
		},
		{
			name:     "GeometryCollection",
			p:        geom.NewCoordinate(0, 0),
			g:        geom.NewGeometryCollection([]geom.Geometry{geom.NewPoint(3, 4), geom.NewPoint(6, 8)}),
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.DistancePointToGeometry(tt.p, tt.g)
			if math.IsInf(tt.expected, 1) {
				assert.True(t, math.IsInf(result, 1), "Expected Inf")
			} else {
				assert.InDelta(t, tt.expected, result, 0.001, "Expected %v", tt.expected)
			}
		})
	}
}

func TestDistanceSegmentToSegment(t *testing.T) {
	tests := []struct {
		name     string
		a1, a2   geom.Coordinate
		b1, b2   geom.Coordinate
		expected float64
	}{
		{
			name:     "Intersecting",
			a1:       geom.NewCoordinate(0, 0),
			a2:       geom.NewCoordinate(10, 10),
			b1:       geom.NewCoordinate(0, 10),
			b2:       geom.NewCoordinate(10, 0),
			expected: 0,
		},
		{
			name:     "Parallel",
			a1:       geom.NewCoordinate(0, 0),
			a2:       geom.NewCoordinate(10, 0),
			b1:       geom.NewCoordinate(0, 5),
			b2:       geom.NewCoordinate(10, 5),
			expected: 5,
		},
		{
			name:     "EndToEnd",
			a1:       geom.NewCoordinate(0, 0),
			a2:       geom.NewCoordinate(5, 0),
			b1:       geom.NewCoordinate(10, 0),
			b2:       geom.NewCoordinate(15, 0),
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.DistanceSegmentToSegment(tt.a1, tt.a2, tt.b1, tt.b2)
			assert.InDelta(t, tt.expected, result, 0.001, "Expected %v", tt.expected)
		})
	}
}

func TestDistanceGeometryToGeometry(t *testing.T) {
	tests := []struct {
		name     string
		g1       geom.Geometry
		g2       geom.Geometry
		expected float64
	}{
		{
			name:     "TwoPoints",
			g1:       geom.NewPoint(0, 0),
			g2:       geom.NewPoint(3, 4),
			expected: 5,
		},
		{
			name:     "PointAndLine",
			g1:       geom.NewPoint(5, 5),
			g2:       geom.NewLineStringXY(0, 0, 10, 0),
			expected: 5,
		},
		{
			name:     "TwoLines",
			g1:       geom.NewLineStringXY(0, 0, 10, 0),
			g2:       geom.NewLineStringXY(0, 5, 10, 5),
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.DistanceGeometryToGeometry(tt.g1, tt.g2)
			assert.InDelta(t, tt.expected, result, 0.001, "Expected %v", tt.expected)
		})
	}

	// Test empty geometries
	empty1 := geom.NewPointEmpty()
	empty2 := geom.NewLineStringEmpty()
	emptyDist := algorithm.DistanceGeometryToGeometry(empty1, empty2)
	assert.True(t, math.IsInf(emptyDist, 1), "Expected Inf for empty geometries")
}

func TestIsWithinDistance(t *testing.T) {
	g1 := geom.NewPoint(0, 0)
	g2 := geom.NewPoint(3, 4)

	tests := []struct {
		name     string
		distance float64
		expected bool
	}{
		{
			name:     "Within",
			distance: 10,
			expected: true,
		},
		{
			name:     "NotWithin",
			distance: 3,
			expected: false,
		},
		{
			name:     "Exact",
			distance: 5,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.IsWithinDistance(g1, g2, tt.distance)
			assert.Equal(t, tt.expected, result, "Expected %v", tt.expected)
		})
	}

	// Test with envelope quick rejection
	farAway1 := geom.NewPoint(0, 0)
	farAway2 := geom.NewPoint(100, 100)
	assert.False(t, algorithm.IsWithinDistance(farAway1, farAway2, 1), "Expected false for far away geometries")
}

func TestNearestPoints(t *testing.T) {
	tests := []struct {
		name      string
		g1        geom.Geometry
		g2        geom.Geometry
		expectedX1 float64
		expectedY1 float64
		expectedX2 float64
		expectedY2 float64
	}{
		{
			name:      "TwoPoints",
			g1:        geom.NewPoint(0, 0),
			g2:        geom.NewPoint(10, 10),
			expectedX1: 0,
			expectedY1: 0,
			expectedX2: 10,
			expectedY2: 10,
		},
		{
			name:       "PointAndLine",
			g1:         geom.NewPoint(5, 5),
			g2:         geom.NewLineStringXY(0, 0, 10, 0),
			expectedX1: 5,
			expectedY1: 5,
			expectedX2: 5, // NearestPoints now correctly finds the projection point on segment
			expectedY2: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p1, p2 := algorithm.NearestPoints(tt.g1, tt.g2)
			assert.InDelta(t, tt.expectedX1, p1.X, 0.001, "Expected p1.X %v", tt.expectedX1)
			assert.InDelta(t, tt.expectedY1, p1.Y, 0.001, "Expected p1.Y %v", tt.expectedY1)
			assert.InDelta(t, tt.expectedX2, p2.X, 0.001, "Expected p2.X %v", tt.expectedX2)
			assert.InDelta(t, tt.expectedY2, p2.Y, 0.001, "Expected p2.Y %v", tt.expectedY2)
		})
	}

	// Test empty geometries
	empty1 := geom.NewPointEmpty()
	empty2 := geom.NewPointEmpty()
	p1, p2 := algorithm.NearestPoints(empty1, empty2)
	assert.True(t, math.IsNaN(p1.X), "Expected NaN for p1.X")
	assert.True(t, math.IsNaN(p1.Y), "Expected NaN for p1.Y")
	assert.True(t, math.IsNaN(p2.X), "Expected NaN for p2.X")
	assert.True(t, math.IsNaN(p2.Y), "Expected NaN for p2.Y")
}

func TestDistancePointToLineString(t *testing.T) {
	tests := []struct {
		name     string
		p        geom.Coordinate
		ls       *geom.LineString
		expected float64
	}{
		{
			name:     "EmptyLineString",
			p:        geom.NewCoordinate(5, 5),
			ls:       geom.NewLineStringEmpty(),
			expected: math.Inf(1),
		},
		{
			name:     "SinglePointLineString",
			p:        geom.NewCoordinate(0, 0),
			ls:       geom.NewLineString(geom.NewCoordinateSequenceXY(3, 4)),
			expected: 5,
		},
		{
			name:     "MultiSegmentLineString",
			p:        geom.NewCoordinate(5, 5),
			ls:       geom.NewLineStringXY(0, 0, 10, 0, 10, 10),
			expected: 5, // Closest to segment (0,0)-(10,0) is (5,0) at distance 5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.DistancePointToLineString(tt.p, tt.ls)
			if math.IsInf(tt.expected, 1) {
				assert.True(t, math.IsInf(result, 1), "Expected Inf")
			} else {
				assert.InDelta(t, tt.expected, result, 0.001, "Expected %v", tt.expected)
			}
		})
	}
}

func TestDistanceToMultiGeometries(t *testing.T) {
	p := geom.NewCoordinate(5, 5)

	// Test MultiPoint - empty
	emptyMp := geom.NewMultiPoint([]*geom.Point{})
	distToEmptyMp := algorithm.DistancePointToGeometry(p, emptyMp)
	assert.True(t, math.IsInf(distToEmptyMp, 1), "Expected Inf for empty MultiPoint")

	// Test MultiLineString - empty
	emptyMls := geom.NewMultiLineString([]*geom.LineString{})
	distToEmptyMls := algorithm.DistancePointToGeometry(p, emptyMls)
	assert.True(t, math.IsInf(distToEmptyMls, 1), "Expected Inf for empty MultiLineString")

	// Test MultiPolygon - empty
	emptyMpoly := geom.NewMultiPolygon([]*geom.Polygon{})
	distToEmptyMpoly := algorithm.DistancePointToGeometry(p, emptyMpoly)
	assert.True(t, math.IsInf(distToEmptyMpoly, 1), "Expected Inf for empty MultiPolygon")

	// Test MultiPolygon with point inside one polygon
	mp := geom.NewMultiPolygon([]*geom.Polygon{
		geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
		geom.NewPolygon(geom.NewLinearRingXY(20, 20, 30, 20, 30, 30, 20, 30, 20, 20), nil),
	})
	distToMp := algorithm.DistancePointToGeometry(p, mp)
	assert.Equal(t, float64(0), distToMp, "Expected 0 for point inside MultiPolygon")

	// Test GeometryCollection - empty
	emptyGc := geom.NewGeometryCollection([]geom.Geometry{})
	distToEmptyGc := algorithm.DistancePointToGeometry(p, emptyGc)
	assert.True(t, math.IsInf(distToEmptyGc, 1), "Expected Inf for empty GeometryCollection")

	// Test GeometryCollection with point inside polygon
	gc := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
	})
	distToGc := algorithm.DistancePointToGeometry(p, gc)
	assert.Equal(t, float64(0), distToGc, "Expected 0 for point inside GeometryCollection")
}

// TestDistanceGeometryToGeometry_PolygonWithHole tests that distance calculations
// correctly handle polygons with holes, without creating phantom segments
// between the shell and holes.
func TestDistanceGeometryToGeometry_PolygonWithHole(t *testing.T) {
	// Create a polygon with a hole
	// Shell: 0,0 to 20,20
	// Hole: 5,5 to 15,15
	shell := geom.NewLinearRingXY(0, 0, 20, 0, 20, 20, 0, 20, 0, 0)
	hole := geom.NewLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	polyWithHole := geom.NewPolygon(shell, []*geom.LinearRing{hole})

	// A point in the center of the hole (10, 10)
	pointInHole := geom.NewPoint(10, 10)

	// Distance from point in hole to the polygon boundary
	// The closest point should be on the hole ring at (10, 5) or (10, 15), distance = 5
	dist := algorithm.DistanceGeometryToGeometry(pointInHole, polyWithHole)
	assert.InDelta(t, 5.0, dist, 0.001, "Distance from point in hole to polygon should be 5")

	// Test symmetric case
	distSymmetric := algorithm.DistanceGeometryToGeometry(polyWithHole, pointInHole)
	assert.InDelta(t, 5.0, distSymmetric, 0.001, "Distance should be symmetric")
}

// TestDistanceGeometryToGeometry_MultiPolygon tests distance calculations
// between disjoint polygons in a MultiPolygon.
func TestDistanceGeometryToGeometry_MultiPolygon(t *testing.T) {
	// Two disjoint polygons
	poly1 := geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)
	poly2 := geom.NewPolygon(geom.NewLinearRingXY(20, 0, 30, 0, 30, 10, 20, 10, 20, 0), nil)
	mp := geom.NewMultiPolygon([]*geom.Polygon{poly1, poly2})

	// A point between the two polygons
	pointBetween := geom.NewPoint(15, 5)

	// Distance should be 5 (to nearest edge of either polygon)
	dist := algorithm.DistanceGeometryToGeometry(pointBetween, mp)
	assert.InDelta(t, 5.0, dist, 0.001, "Distance from point between polygons should be 5")

	// Distance between two disjoint polygons
	anotherPoly := geom.NewPolygon(geom.NewLinearRingXY(14, 4, 16, 4, 16, 6, 14, 6, 14, 4), nil)
	distBetweenPolys := algorithm.DistanceGeometryToGeometry(mp, anotherPoly)
	assert.InDelta(t, 4.0, distBetweenPolys, 0.001, "Distance between MultiPolygon and another polygon")
}

// TestNearestPoints_PolygonWithHole tests that nearest points are correctly
// calculated for polygons with holes.
func TestNearestPoints_PolygonWithHole(t *testing.T) {
	// Polygon with hole
	shell := geom.NewLinearRingXY(0, 0, 20, 0, 20, 20, 0, 20, 0, 0)
	hole := geom.NewLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	polyWithHole := geom.NewPolygon(shell, []*geom.LinearRing{hole})

	// Point at center of hole
	pointInHole := geom.NewPoint(10, 10)

	p1, p2 := algorithm.NearestPoints(pointInHole, polyWithHole)

	// p1 should be the point (10, 10)
	assert.InDelta(t, 10.0, p1.X, 0.001, "Nearest point on g1 X")
	assert.InDelta(t, 10.0, p1.Y, 0.001, "Nearest point on g1 Y")

	// p2 should be on the hole ring, at (10, 5) or (10, 15) - closest is (10, 5) or (10, 15) at distance 5
	dist := p1.Distance(p2)
	assert.InDelta(t, 5.0, dist, 0.001, "Distance between nearest points should be 5")

	// p2 should be on the hole boundary (y = 5 or y = 15)
	assert.True(t, math.Abs(p2.Y-5.0) < 0.001 || math.Abs(p2.Y-15.0) < 0.001,
		"Nearest point on polygon should be on hole boundary, got Y=%v", p2.Y)
}

// TestNearestPoints_MultiLineString tests that no artificial segments are created
// between lines in a MultiLineString.
func TestNearestPoints_MultiLineString(t *testing.T) {
	// Two separate horizontal line segments with a gap
	line1 := geom.NewLineStringXY(0, 0, 10, 0)
	line2 := geom.NewLineStringXY(20, 0, 30, 0)
	mls := geom.NewMultiLineString([]*geom.LineString{line1, line2})

	// A point at (15, 0) - between the two lines
	point := geom.NewPoint(15, 0)

	p1, p2 := algorithm.NearestPoints(point, mls)

	// p1 should be (15, 0)
	assert.InDelta(t, 15.0, p1.X, 0.001, "Point X")
	assert.InDelta(t, 0.0, p1.Y, 0.001, "Point Y")

	// p2 should be either (10, 0) or (20, 0), the closest endpoints
	// Distance should be 5
	dist := p1.Distance(p2)
	assert.InDelta(t, 5.0, dist, 0.001, "Distance to nearest point on MultiLineString")

	// The nearest point should NOT be at (15, 0) on a phantom segment
	// It should be at (10, 0) or (20, 0)
	assert.True(t, (math.Abs(p2.X-10.0) < 0.001 || math.Abs(p2.X-20.0) < 0.001),
		"Nearest point should be at line endpoint, not on phantom segment, got X=%v", p2.X)
}

// TestDistance_PolygonHoleInterior tests the Distance function for a point
// inside a polygon hole.
func TestDistance_PolygonHoleInterior(t *testing.T) {
	// Large polygon with a small hole
	shell := geom.NewLinearRingXY(0, 0, 100, 0, 100, 100, 0, 100, 0, 0)
	hole := geom.NewLinearRingXY(40, 40, 60, 40, 60, 60, 40, 60, 40, 40)
	polyWithHole := geom.NewPolygon(shell, []*geom.LinearRing{hole})

	// Point at center of hole (50, 50)
	pointInHole := geom.NewPoint(50, 50)

	// Using the high-level Distance function
	dist := algorithm.Distance(pointInHole, polyWithHole)

	// Distance should be 10 (to nearest edge of hole at y=40 or y=60)
	assert.InDelta(t, 10.0, dist, 0.001,
		"Point in hole should have distance 10 to hole boundary")

	// Verify it's not 0 (which would be wrong - that would mean the point is "inside")
	assert.NotEqual(t, 0.0, dist, "Point in hole should NOT have distance 0")
}

// TestDistanceGeometryToGeometry_NoPhantomSegments verifies that no phantom
// segments are created between consecutive rings in a polygon.
func TestDistanceGeometryToGeometry_NoPhantomSegments(t *testing.T) {
	// Create a polygon where a phantom segment would cause incorrect results
	// Shell: square from (0,0) to (100,100)
	// Hole: square from (10,10) to (20,20)
	// If there's a phantom segment from (0,0) to (10,10), it would be closer
	// to a point at (5,5) than the actual hole boundary
	shell := geom.NewLinearRingXY(0, 0, 100, 0, 100, 100, 0, 100, 0, 0)
	hole := geom.NewLinearRingXY(10, 10, 20, 10, 20, 20, 10, 20, 10, 10)
	poly := geom.NewPolygon(shell, []*geom.LinearRing{hole})

	// Point at (5, 5) - inside the polygon but outside the hole
	point := geom.NewPoint(5, 5)

	dist := algorithm.DistanceGeometryToGeometry(point, poly)

	// The point (5,5) is inside the polygon (outside the hole)
	// Closest boundary is the shell at x=0 or y=0, distance = 5
	// If there were a phantom segment from shell's last point to hole's first point,
	// it might incorrectly calculate a smaller distance
	assert.InDelta(t, 5.0, dist, 0.001,
		"Distance should be to shell boundary, not to phantom segment")
}
