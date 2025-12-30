package algorithm_test

import (
	"math"
	"testing"

	"github.com/go-topology-suite/gts/algorithm"
	"github.com/go-topology-suite/gts/geom"
	"github.com/stretchr/testify/assert"
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
			assert.InDelta(t, tt.expected, result, 0.001, "Expected area %v", tt.expected)
		})
	}
}

func TestMultiPolygonArea(t *testing.T) {
	mp := geom.NewMultiPolygon([]*geom.Polygon{
		geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
		geom.NewPolygon(geom.NewLinearRingXY(20, 20, 30, 20, 30, 30, 20, 30, 20, 20), nil),
	})
	area := algorithm.MultiPolygonArea(mp)
	assert.Equal(t, 200.0, area, "Expected area 200")
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
			assert.InDelta(t, tt.expected, result, 0.001, "Expected perimeter %v", tt.expected)
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
			assert.InDelta(t, tt.expected, result, 0.001, "Expected length %v", tt.expected)
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
			assert.InDelta(t, tt.expected, result, 0.001, "Expected length %v", tt.expected)
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
			assert.InDelta(t, tt.expectedX, result.X, 0.001, "Expected X %v", tt.expectedX)
			assert.InDelta(t, tt.expectedY, result.Y, 0.001, "Expected Y %v", tt.expectedY)
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
	assert.InDelta(t, 5.0, centroid.X, 0.001, "Expected X 5")
	assert.InDelta(t, 5.0, centroid.Y, 0.001, "Expected Y 5")

	// Test empty
	emptyMp := geom.NewMultiPoint([]*geom.Point{})
	emptyCentroid := algorithm.MultiPointCentroid(emptyMp)
	assert.True(t, math.IsNaN(emptyCentroid.X), "Expected NaN for empty multipoint X")
	assert.True(t, math.IsNaN(emptyCentroid.Y), "Expected NaN for empty multipoint Y")
}

func TestMultiLineStringCentroid(t *testing.T) {
	mls := geom.NewMultiLineString([]*geom.LineString{
		geom.NewLineStringXY(0, 0, 10, 0),
		geom.NewLineStringXY(0, 10, 10, 10),
	})
	centroid := algorithm.MultiLineStringCentroid(mls)
	assert.InDelta(t, 5.0, centroid.X, 0.001, "Expected X 5")
	assert.InDelta(t, 5.0, centroid.Y, 0.001, "Expected Y 5")

	// Test empty
	emptyMls := geom.NewMultiLineString([]*geom.LineString{})
	emptyCentroid := algorithm.MultiLineStringCentroid(emptyMls)
	assert.True(t, math.IsNaN(emptyCentroid.X), "Expected NaN for empty multilinestring X")
	assert.True(t, math.IsNaN(emptyCentroid.Y), "Expected NaN for empty multilinestring Y")

	// Test with zero-length lines
	mlsZero := geom.NewMultiLineString([]*geom.LineString{
		geom.NewLineStringXY(5, 5, 5, 5),
	})
	centroidZero := algorithm.MultiLineStringCentroid(mlsZero)
	assert.InDelta(t, 5.0, centroidZero.X, 0.001, "Expected X 5 for zero-length line")
	assert.InDelta(t, 5.0, centroidZero.Y, 0.001, "Expected Y 5 for zero-length line")
}

func TestMultiPolygonCentroid(t *testing.T) {
	mp := geom.NewMultiPolygon([]*geom.Polygon{
		geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
		geom.NewPolygon(geom.NewLinearRingXY(20, 20, 30, 20, 30, 30, 20, 30, 20, 20), nil),
	})
	centroid := algorithm.MultiPolygonCentroid(mp)
	assert.InDelta(t, 15.0, centroid.X, 0.001, "Expected X 15")
	assert.InDelta(t, 15.0, centroid.Y, 0.001, "Expected Y 15")

	// Test empty
	emptyMp := geom.NewMultiPolygon([]*geom.Polygon{})
	emptyCentroid := algorithm.MultiPolygonCentroid(emptyMp)
	assert.True(t, math.IsNaN(emptyCentroid.X), "Expected NaN for empty multipolygon X")
	assert.True(t, math.IsNaN(emptyCentroid.Y), "Expected NaN for empty multipolygon Y")
}

func TestGeometryCollectionCentroid(t *testing.T) {
	// Test with polygons (highest dimension)
	gc1 := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
		geom.NewPoint(100, 100), // Should be ignored
	})
	centroid1 := algorithm.GeometryCollectionCentroid(gc1)
	assert.InDelta(t, 5.0, centroid1.X, 0.001, "Expected X 5")
	assert.InDelta(t, 5.0, centroid1.Y, 0.001, "Expected Y 5")

	// Test with lines (no polygons)
	gc2 := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewLineStringXY(0, 0, 10, 0),
		geom.NewPoint(100, 100), // Should be ignored
	})
	centroid2 := algorithm.GeometryCollectionCentroid(gc2)
	assert.InDelta(t, 5.0, centroid2.X, 0.001, "Expected X 5")
	assert.InDelta(t, 0.0, centroid2.Y, 0.001, "Expected Y 0")

	// Test with points only
	gc3 := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(0, 0),
		geom.NewPoint(10, 10),
	})
	centroid3 := algorithm.GeometryCollectionCentroid(gc3)
	assert.InDelta(t, 5.0, centroid3.X, 0.001, "Expected X 5")
	assert.InDelta(t, 5.0, centroid3.Y, 0.001, "Expected Y 5")

	// Test empty
	emptyGc := geom.NewGeometryCollection([]geom.Geometry{})
	emptyCentroid := algorithm.GeometryCollectionCentroid(emptyGc)
	assert.True(t, math.IsNaN(emptyCentroid.X), "Expected NaN for empty collection X")
	assert.True(t, math.IsNaN(emptyCentroid.Y), "Expected NaN for empty collection Y")

	// Test with MultiPoint
	gc4 := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewMultiPoint([]*geom.Point{geom.NewPoint(0, 0), geom.NewPoint(10, 10)}),
	})
	centroid4 := algorithm.GeometryCollectionCentroid(gc4)
	assert.InDelta(t, 5.0, centroid4.X, 0.001, "Expected X 5")
	assert.InDelta(t, 5.0, centroid4.Y, 0.001, "Expected Y 5")
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
			assert.InDelta(t, tt.expectedX, result.X, 0.001, "Expected X %v", tt.expectedX)
			assert.InDelta(t, tt.expectedY, result.Y, 0.001, "Expected Y %v", tt.expectedY)
		})
	}

	// Test empty
	empty := geom.CoordinateSequence{}
	emptyCentroid := algorithm.LineCentroid(empty)
	assert.True(t, math.IsNaN(emptyCentroid.X), "Expected NaN for empty sequence X")
	assert.True(t, math.IsNaN(emptyCentroid.Y), "Expected NaN for empty sequence Y")

	// Test zero-length line
	zeroLen := geom.NewCoordinateSequenceXY(5, 5, 5, 5)
	zeroLenCentroid := algorithm.LineCentroid(zeroLen)
	assert.InDelta(t, 5.0, zeroLenCentroid.X, 0.001, "Expected X 5 for zero-length")
	assert.InDelta(t, 5.0, zeroLenCentroid.Y, 0.001, "Expected Y 5 for zero-length")
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
			assert.InDelta(t, tt.expectedX, result.X, 0.001, "Expected X %v", tt.expectedX)
			assert.InDelta(t, tt.expectedY, result.Y, 0.001, "Expected Y %v", tt.expectedY)
		})
	}

	// Test degenerate ring (zero area)
	degenerateRing := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 0, 0, 0, 0)
	degenerateCentroid := algorithm.RingCentroid(degenerateRing)
	// Should fall back to line centroid
	assert.False(t, math.IsNaN(degenerateCentroid.X), "Expected valid centroid X for degenerate ring")
	assert.False(t, math.IsNaN(degenerateCentroid.Y), "Expected valid centroid Y for degenerate ring")
}

func TestPolygonCentroid(t *testing.T) {
	// Test with hole
	shell := geom.NewLinearRingXY(0, 0, 20, 0, 20, 20, 0, 20, 0, 0)
	hole := geom.NewLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	poly := geom.NewPolygon(shell, []*geom.LinearRing{hole})
	centroid := algorithm.PolygonCentroid(poly)
	// Centroid should be at (10, 10) for this symmetric case
	assert.InDelta(t, 10.0, centroid.X, 0.001, "Expected X 10")
	assert.InDelta(t, 10.0, centroid.Y, 0.001, "Expected Y 10")

	// Test empty polygon
	emptyPoly := geom.NewPolygonEmpty()
	emptyCentroid := algorithm.PolygonCentroid(emptyPoly)
	assert.True(t, math.IsNaN(emptyCentroid.X), "Expected NaN for empty polygon X")
	assert.True(t, math.IsNaN(emptyCentroid.Y), "Expected NaN for empty polygon Y")

	// Test polygon with zero total area (hole same size as shell)
	shell2 := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	hole2 := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly2 := geom.NewPolygon(shell2, []*geom.LinearRing{hole2})
	centroid2 := algorithm.PolygonCentroid(poly2)
	// Should return shell centroid
	assert.InDelta(t, 5.0, centroid2.X, 0.001, "Expected X 5 for zero-area polygon")
	assert.InDelta(t, 5.0, centroid2.Y, 0.001, "Expected Y 5 for zero-area polygon")
}
