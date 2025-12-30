package algorithm_test

import (
	"testing"

	"github.com/go-topology-suite/gts/algorithm"
	"github.com/go-topology-suite/gts/geom"
	"github.com/stretchr/testify/assert"
)

func TestConvexHull(t *testing.T) {
	tests := []struct {
		name         string
		geom         geom.Geometry
		expectedType string
		expectedSize int
	}{
		{
			name:         "Point",
			geom:         geom.NewPoint(5, 5),
			expectedType: "Point",
			expectedSize: 1,
		},
		{
			name:         "TwoPoints",
			geom:         geom.NewLineStringXY(0, 0, 10, 10),
			expectedType: "LineString",
			expectedSize: 2,
		},
		{
			name:         "Triangle",
			geom:         geom.NewLineStringXY(0, 0, 10, 0, 5, 10),
			expectedType: "Polygon",
			expectedSize: 4, // 3 + closing point
		},
		{
			name:         "Square",
			geom:         geom.NewLineStringXY(0, 0, 10, 0, 10, 10, 0, 10),
			expectedType: "Polygon",
			expectedSize: 5, // 4 + closing point
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.ConvexHull(tt.geom)
			assert.Equal(t, tt.expectedType, result.GeometryType(), "Expected type %s", tt.expectedType)
			coords := result.Coordinates()
			assert.Equal(t, tt.expectedSize, len(coords), "Expected %d coordinates", tt.expectedSize)
		})
	}
}

func TestConvexHullFromCoords(t *testing.T) {
	tests := []struct {
		name         string
		coords       geom.CoordinateSequence
		expectedType string
	}{
		{
			name:         "Empty",
			coords:       geom.CoordinateSequence{},
			expectedType: "Point",
		},
		{
			name:         "SinglePoint",
			coords:       geom.NewCoordinateSequenceXY(5, 5),
			expectedType: "Point",
		},
		{
			name:         "TwoPoints",
			coords:       geom.NewCoordinateSequenceXY(0, 0, 10, 10),
			expectedType: "LineString",
		},
		{
			name:         "TwoIdenticalPoints",
			coords:       geom.NewCoordinateSequenceXY(5, 5, 5, 5),
			expectedType: "LineString", // After uniquing, might return LineString with same points
		},
		{
			name:         "ThreeCollinearPoints",
			coords:       geom.NewCoordinateSequenceXY(0, 0, 5, 5, 10, 10),
			expectedType: "LineString",
		},
		{
			name:         "Triangle",
			coords:       geom.NewCoordinateSequenceXY(0, 0, 10, 0, 5, 10),
			expectedType: "Polygon",
		},
		{
			name:         "SquareWithInteriorPoint",
			coords:       geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 10, 5, 5),
			expectedType: "Polygon",
		},
		{
			name:         "DuplicatePoints",
			coords:       geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 0, 5, 10, 5, 10),
			expectedType: "Polygon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.ConvexHullFromCoords(tt.coords)
			assert.Equal(t, tt.expectedType, result.GeometryType(), "Expected type %s", tt.expectedType)
		})
	}
}

func TestMonotoneChain(t *testing.T) {
	tests := []struct {
		name         string
		coords       geom.CoordinateSequence
		expectedType string
	}{
		{
			name:         "Empty",
			coords:       geom.CoordinateSequence{},
			expectedType: "Point",
		},
		{
			name:         "SinglePoint",
			coords:       geom.NewCoordinateSequenceXY(5, 5),
			expectedType: "Point",
		},
		{
			name:         "TwoPoints",
			coords:       geom.NewCoordinateSequenceXY(0, 0, 10, 10),
			expectedType: "LineString",
		},
		{
			name:         "Triangle",
			coords:       geom.NewCoordinateSequenceXY(0, 0, 10, 0, 5, 10),
			expectedType: "Polygon",
		},
		{
			name:         "Square",
			coords:       geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 10),
			expectedType: "Polygon",
		},
		{
			name:         "DuplicatePoints",
			coords:       geom.NewCoordinateSequenceXY(0, 0, 5, 5, 5, 5, 10, 10),
			expectedType: "LineString",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.MonotoneChain(tt.coords)
			assert.Equal(t, tt.expectedType, result.GeometryType(), "Expected type %s", tt.expectedType)
		})
	}
}

func TestIsConvex(t *testing.T) {
	tests := []struct {
		name     string
		poly     *geom.Polygon
		expected bool
	}{
		{
			name:     "Empty",
			poly:     geom.NewPolygonEmpty(),
			expected: false,
		},
		{
			name:     "Square",
			poly:     geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
			expected: true,
		},
		{
			name:     "Triangle",
			poly:     geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 5, 10, 0, 0), nil),
			expected: true,
		},
		{
			name:     "Concave",
			poly:     geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 5, 5, 0, 10, 0, 0), nil),
			expected: false,
		},
		{
			name: "WithHole",
			poly: geom.NewPolygon(
				geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0),
				[]*geom.LinearRing{geom.NewLinearRingXY(2, 2, 8, 2, 8, 8, 2, 8, 2, 2)},
			),
			expected: false,
		},
		{
			name:     "SmallTriangle",
			poly:     geom.NewPolygon(geom.NewLinearRingXY(0, 0, 1, 0, 0, 1, 0, 0), nil),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.IsConvex(tt.poly)
			assert.Equal(t, tt.expected, result, "Expected %v", tt.expected)
		})
	}
}

func TestConvexHullEdgeCases(t *testing.T) {
	// Test with all collinear points
	t.Run("AllCollinear", func(t *testing.T) {
		coords := geom.NewCoordinateSequenceXY(0, 0, 1, 1, 2, 2, 3, 3, 4, 4)
		result := algorithm.ConvexHullFromCoords(coords)
		assert.Equal(t, "LineString", result.GeometryType(), "Expected LineString for collinear points")
	})

	// Test with points forming a star
	t.Run("StarPattern", func(t *testing.T) {
		coords := geom.NewCoordinateSequenceXY(
			0, 10, 5, 5, 10, 10, // Top points
			5, 0, // Center bottom
			0, 0, 10, 0, // Bottom corners
		)
		result := algorithm.ConvexHullFromCoords(coords)
		assert.Equal(t, "Polygon", result.GeometryType(), "Expected Polygon")
	})

	// Test with nearly collinear points
	t.Run("NearlyCollinear", func(t *testing.T) {
		coords := geom.NewCoordinateSequenceXY(
			0, 0, 5, 0.01, 10, 0,
		)
		result := algorithm.ConvexHullFromCoords(coords)
		// Should still form a polygon or line depending on epsilon
		assert.False(t, result.IsEmpty(), "Result should not be empty")
	})
}
