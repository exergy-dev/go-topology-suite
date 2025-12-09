package algorithm_test

import (
	"math"
	"testing"

	"github.com/go-topology-suite/gts/algorithm"
	"github.com/go-topology-suite/gts/geom"
)

func TestAreaGeometryTypes(t *testing.T) {
	tests := []struct {
		name     string
		geom     geom.Geometry
		expected float64
	}{
		{
			name:     "Point",
			geom:     geom.NewPoint(5, 5),
			expected: 0,
		},
		{
			name:     "LineString",
			geom:     geom.NewLineStringXY(0, 0, 10, 0),
			expected: 0,
		},
		{
			name:     "Polygon",
			geom:     geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
			expected: 100,
		},
		{
			name:     "MultiPoint",
			geom:     geom.NewMultiPoint([]*geom.Point{geom.NewPoint(0, 0), geom.NewPoint(10, 10)}),
			expected: 0,
		},
		{
			name:     "MultiLineString",
			geom:     geom.NewMultiLineString([]*geom.LineString{geom.NewLineStringXY(0, 0, 10, 0)}),
			expected: 0,
		},
		{
			name: "MultiPolygon",
			geom: geom.NewMultiPolygon([]*geom.Polygon{
				geom.NewPolygon(geom.NewLinearRingXY(0, 0, 5, 0, 5, 5, 0, 5, 0, 0), nil),
				geom.NewPolygon(geom.NewLinearRingXY(10, 10, 15, 10, 15, 15, 10, 15, 10, 10), nil),
			}),
			expected: 50,
		},
		{
			name: "GeometryCollection",
			geom: geom.NewGeometryCollection([]geom.Geometry{
				geom.NewPoint(0, 0),
				geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
			}),
			expected: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.Area(tt.geom)
			if math.Abs(result-tt.expected) > 0.001 {
				t.Errorf("Expected area %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMultiPolygonArea(t *testing.T) {
	mp := geom.NewMultiPolygon([]*geom.Polygon{
		geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
		geom.NewPolygon(geom.NewLinearRingXY(20, 20, 30, 20, 30, 30, 20, 30, 20, 20), nil),
	})
	area := algorithm.MultiPolygonArea(mp)
	if area != 200 {
		t.Errorf("Expected area 200, got %v", area)
	}
}

func TestPolygonPerimeter(t *testing.T) {
	tests := []struct {
		name     string
		poly     *geom.Polygon
		expected float64
	}{
		{
			name:     "Empty",
			poly:     geom.NewPolygonEmpty(),
			expected: 0,
		},
		{
			name:     "Square",
			poly:     geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
			expected: 40,
		},
		{
			name: "SquareWithHole",
			poly: geom.NewPolygon(
				geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0),
				[]*geom.LinearRing{geom.NewLinearRingXY(2, 2, 8, 2, 8, 8, 2, 8, 2, 2)},
			),
			expected: 64, // 40 + 24
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.PolygonPerimeter(tt.poly)
			if math.Abs(result-tt.expected) > 0.001 {
				t.Errorf("Expected perimeter %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLength(t *testing.T) {
	tests := []struct {
		name     string
		geom     geom.Geometry
		expected float64
	}{
		{
			name:     "Point",
			geom:     geom.NewPoint(5, 5),
			expected: 0,
		},
		{
			name:     "LineString",
			geom:     geom.NewLineStringXY(0, 0, 3, 4),
			expected: 5,
		},
		{
			name:     "LinearRing",
			geom:     geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0),
			expected: 40,
		},
		{
			name:     "Polygon",
			geom:     geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
			expected: 40,
		},
		{
			name: "MultiLineString",
			geom: geom.NewMultiLineString([]*geom.LineString{
				geom.NewLineStringXY(0, 0, 10, 0),
				geom.NewLineStringXY(0, 0, 0, 10),
			}),
			expected: 20,
		},
		{
			name: "MultiPolygon",
			geom: geom.NewMultiPolygon([]*geom.Polygon{
				geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
				geom.NewPolygon(geom.NewLinearRingXY(20, 20, 30, 20, 30, 30, 20, 30, 20, 20), nil),
			}),
			expected: 80,
		},
		{
			name: "GeometryCollection",
			geom: geom.NewGeometryCollection([]geom.Geometry{
				geom.NewLineStringXY(0, 0, 10, 0),
				geom.NewPoint(5, 5),
			}),
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.Length(tt.geom)
			if math.Abs(result-tt.expected) > 0.001 {
				t.Errorf("Expected length %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLineLength(t *testing.T) {
	tests := []struct {
		name     string
		coords   geom.CoordinateSequence
		expected float64
	}{
		{
			name:     "Empty",
			coords:   geom.CoordinateSequence{},
			expected: 0,
		},
		{
			name:     "SinglePoint",
			coords:   geom.NewCoordinateSequenceXY(5, 5),
			expected: 0,
		},
		{
			name:     "TwoPoints",
			coords:   geom.NewCoordinateSequenceXY(0, 0, 3, 4),
			expected: 5,
		},
		{
			name:     "ThreePoints",
			coords:   geom.NewCoordinateSequenceXY(0, 0, 3, 4, 3, 8),
			expected: 9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.LineLength(tt.coords)
			if math.Abs(result-tt.expected) > 0.001 {
				t.Errorf("Expected length %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCentroidGeometryTypes(t *testing.T) {
	tests := []struct {
		name     string
		geom     geom.Geometry
		expectedX float64
		expectedY float64
	}{
		{
			name:      "Point",
			geom:      geom.NewPoint(5, 10),
			expectedX: 5,
			expectedY: 10,
		},
		{
			name:      "LineString",
			geom:      geom.NewLineStringXY(0, 0, 10, 0),
			expectedX: 5,
			expectedY: 0,
		},
		{
			name:      "LinearRing",
			geom:      geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0),
			expectedX: 5,
			expectedY: 5,
		},
		{
			name:      "Polygon",
			geom:      geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
			expectedX: 5,
			expectedY: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.Centroid(tt.geom)
			if math.Abs(result.X-tt.expectedX) > 0.001 || math.Abs(result.Y-tt.expectedY) > 0.001 {
				t.Errorf("Expected (%v, %v), got (%v, %v)", tt.expectedX, tt.expectedY, result.X, result.Y)
			}
		})
	}
}

func TestMultiPointCentroid(t *testing.T) {
	mp := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(0, 0),
		geom.NewPoint(10, 0),
		geom.NewPoint(10, 10),
		geom.NewPoint(0, 10),
	})
	centroid := algorithm.MultiPointCentroid(mp)
	if math.Abs(centroid.X-5) > 0.001 || math.Abs(centroid.Y-5) > 0.001 {
		t.Errorf("Expected (5, 5), got (%v, %v)", centroid.X, centroid.Y)
	}

	// Test empty
	emptyMp := geom.NewMultiPoint([]*geom.Point{})
	emptyCentroid := algorithm.MultiPointCentroid(emptyMp)
	if !math.IsNaN(emptyCentroid.X) || !math.IsNaN(emptyCentroid.Y) {
		t.Errorf("Expected NaN for empty multipoint")
	}
}

func TestMultiLineStringCentroid(t *testing.T) {
	mls := geom.NewMultiLineString([]*geom.LineString{
		geom.NewLineStringXY(0, 0, 10, 0),
		geom.NewLineStringXY(0, 10, 10, 10),
	})
	centroid := algorithm.MultiLineStringCentroid(mls)
	if math.Abs(centroid.X-5) > 0.001 || math.Abs(centroid.Y-5) > 0.001 {
		t.Errorf("Expected (5, 5), got (%v, %v)", centroid.X, centroid.Y)
	}

	// Test empty
	emptyMls := geom.NewMultiLineString([]*geom.LineString{})
	emptyCentroid := algorithm.MultiLineStringCentroid(emptyMls)
	if !math.IsNaN(emptyCentroid.X) || !math.IsNaN(emptyCentroid.Y) {
		t.Errorf("Expected NaN for empty multilinestring")
	}

	// Test with zero-length lines
	mlsZero := geom.NewMultiLineString([]*geom.LineString{
		geom.NewLineStringXY(5, 5, 5, 5),
	})
	centroidZero := algorithm.MultiLineStringCentroid(mlsZero)
	if math.Abs(centroidZero.X-5) > 0.001 || math.Abs(centroidZero.Y-5) > 0.001 {
		t.Errorf("Expected (5, 5) for zero-length line, got (%v, %v)", centroidZero.X, centroidZero.Y)
	}
}

func TestMultiPolygonCentroid(t *testing.T) {
	mp := geom.NewMultiPolygon([]*geom.Polygon{
		geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
		geom.NewPolygon(geom.NewLinearRingXY(20, 20, 30, 20, 30, 30, 20, 30, 20, 20), nil),
	})
	centroid := algorithm.MultiPolygonCentroid(mp)
	if math.Abs(centroid.X-15) > 0.001 || math.Abs(centroid.Y-15) > 0.001 {
		t.Errorf("Expected (15, 15), got (%v, %v)", centroid.X, centroid.Y)
	}

	// Test empty
	emptyMp := geom.NewMultiPolygon([]*geom.Polygon{})
	emptyCentroid := algorithm.MultiPolygonCentroid(emptyMp)
	if !math.IsNaN(emptyCentroid.X) || !math.IsNaN(emptyCentroid.Y) {
		t.Errorf("Expected NaN for empty multipolygon")
	}
}

func TestGeometryCollectionCentroid(t *testing.T) {
	// Test with polygons (highest dimension)
	gc1 := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
		geom.NewPoint(100, 100), // Should be ignored
	})
	centroid1 := algorithm.GeometryCollectionCentroid(gc1)
	if math.Abs(centroid1.X-5) > 0.001 || math.Abs(centroid1.Y-5) > 0.001 {
		t.Errorf("Expected (5, 5), got (%v, %v)", centroid1.X, centroid1.Y)
	}

	// Test with lines (no polygons)
	gc2 := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewLineStringXY(0, 0, 10, 0),
		geom.NewPoint(100, 100), // Should be ignored
	})
	centroid2 := algorithm.GeometryCollectionCentroid(gc2)
	if math.Abs(centroid2.X-5) > 0.001 || math.Abs(centroid2.Y-0) > 0.001 {
		t.Errorf("Expected (5, 0), got (%v, %v)", centroid2.X, centroid2.Y)
	}

	// Test with points only
	gc3 := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		geom.NewPoint(10, 10),
	})
	centroid3 := algorithm.GeometryCollectionCentroid(gc3)
	if math.Abs(centroid3.X-5) > 0.001 || math.Abs(centroid3.Y-5) > 0.001 {
		t.Errorf("Expected (5, 5), got (%v, %v)", centroid3.X, centroid3.Y)
	}

	// Test empty
	emptyGc := geom.NewGeometryCollection([]geom.Geometry{})
	emptyCentroid := algorithm.GeometryCollectionCentroid(emptyGc)
	if !math.IsNaN(emptyCentroid.X) || !math.IsNaN(emptyCentroid.Y) {
		t.Errorf("Expected NaN for empty collection")
	}

	// Test with MultiPoint
	gc4 := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewMultiPoint([]*geom.Point{geom.NewPoint(0, 0), geom.NewPoint(10, 10)}),
	})
	centroid4 := algorithm.GeometryCollectionCentroid(gc4)
	if math.Abs(centroid4.X-5) > 0.001 || math.Abs(centroid4.Y-5) > 0.001 {
		t.Errorf("Expected (5, 5), got (%v, %v)", centroid4.X, centroid4.Y)
	}
}

func TestLineCentroid(t *testing.T) {
	tests := []struct {
		name      string
		coords    geom.CoordinateSequence
		expectedX float64
		expectedY float64
	}{
		{
			name:      "TwoPoints",
			coords:    geom.NewCoordinateSequenceXY(0, 0, 10, 0),
			expectedX: 5,
			expectedY: 0,
		},
		{
			name:      "SinglePoint",
			coords:    geom.NewCoordinateSequenceXY(5, 5),
			expectedX: 5,
			expectedY: 5,
		},
		{
			name:      "ThreePoints",
			coords:    geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10),
			expectedX: 7.5,
			expectedY: 2.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.LineCentroid(tt.coords)
			if math.Abs(result.X-tt.expectedX) > 0.001 || math.Abs(result.Y-tt.expectedY) > 0.001 {
				t.Errorf("Expected (%v, %v), got (%v, %v)", tt.expectedX, tt.expectedY, result.X, result.Y)
			}
		})
	}

	// Test empty
	empty := geom.CoordinateSequence{}
	emptyCentroid := algorithm.LineCentroid(empty)
	if !math.IsNaN(emptyCentroid.X) || !math.IsNaN(emptyCentroid.Y) {
		t.Errorf("Expected NaN for empty sequence")
	}

	// Test zero-length line
	zeroLen := geom.NewCoordinateSequenceXY(5, 5, 5, 5)
	zeroLenCentroid := algorithm.LineCentroid(zeroLen)
	if math.Abs(zeroLenCentroid.X-5) > 0.001 || math.Abs(zeroLenCentroid.Y-5) > 0.001 {
		t.Errorf("Expected (5, 5) for zero-length, got (%v, %v)", zeroLenCentroid.X, zeroLenCentroid.Y)
	}
}

func TestRingCentroid(t *testing.T) {
	tests := []struct {
		name      string
		coords    geom.CoordinateSequence
		expectedX float64
		expectedY float64
	}{
		{
			name:      "Square",
			coords:    geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0),
			expectedX: 5,
			expectedY: 5,
		},
		{
			name:      "Triangle",
			coords:    geom.NewCoordinateSequenceXY(0, 0, 10, 0, 5, 10, 0, 0),
			expectedX: 5,
			expectedY: 10.0 / 3,
		},
		{
			name:      "TwoPoints",
			coords:    geom.NewCoordinateSequenceXY(0, 0, 10, 0),
			expectedX: 5,
			expectedY: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.RingCentroid(tt.coords)
			if math.Abs(result.X-tt.expectedX) > 0.001 || math.Abs(result.Y-tt.expectedY) > 0.001 {
				t.Errorf("Expected (%v, %v), got (%v, %v)", tt.expectedX, tt.expectedY, result.X, result.Y)
			}
		})
	}

	// Test degenerate ring (zero area)
	degenerateRing := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 0, 0, 0, 0)
	degenerateCentroid := algorithm.RingCentroid(degenerateRing)
	// Should fall back to line centroid
	if math.IsNaN(degenerateCentroid.X) || math.IsNaN(degenerateCentroid.Y) {
		t.Errorf("Expected valid centroid for degenerate ring, got NaN")
	}
}

func TestPolygonCentroid(t *testing.T) {
	// Test with hole
	shell := geom.NewLinearRingXY(0, 0, 20, 0, 20, 20, 0, 20, 0, 0)
	hole := geom.NewLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	poly := geom.NewPolygon(shell, []*geom.LinearRing{hole})
	centroid := algorithm.PolygonCentroid(poly)
	// Centroid should be at (10, 10) for this symmetric case
	if math.Abs(centroid.X-10) > 0.001 || math.Abs(centroid.Y-10) > 0.001 {
		t.Errorf("Expected (10, 10), got (%v, %v)", centroid.X, centroid.Y)
	}

	// Test empty polygon
	emptyPoly := geom.NewPolygonEmpty()
	emptyCentroid := algorithm.PolygonCentroid(emptyPoly)
	if !math.IsNaN(emptyCentroid.X) || !math.IsNaN(emptyCentroid.Y) {
		t.Errorf("Expected NaN for empty polygon")
	}

	// Test polygon with zero total area (hole same size as shell)
	shell2 := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	hole2 := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly2 := geom.NewPolygon(shell2, []*geom.LinearRing{hole2})
	centroid2 := algorithm.PolygonCentroid(poly2)
	// Should return shell centroid
	if math.Abs(centroid2.X-5) > 0.001 || math.Abs(centroid2.Y-5) > 0.001 {
		t.Errorf("Expected (5, 5) for zero-area polygon, got (%v, %v)", centroid2.X, centroid2.Y)
	}
}
