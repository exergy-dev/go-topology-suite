package algorithm_test

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/algorithm"
	"github.com/robert-malhotra/go-topology-suite/geom"
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

func TestConvexHull_LargeCoordinates(t *testing.T) {
	// Test with coordinates having large magnitudes (1e10+)
	coords := geom.NewCoordinateSequenceXY(
		1e10, 1e10,
		1e10+100, 1e10,
		1e10+100, 1e10+100,
		1e10, 1e10+100,
	)
	result := algorithm.ConvexHullFromCoords(coords)
	assert.Equal(t, "Polygon", result.GeometryType(), "Expected Polygon for large coordinates")
	// Should have 5 coordinates (4 corners + closing point)
	assert.Equal(t, 5, len(result.Coordinates()), "Expected 5 coordinates for square hull")
}

func TestConvexHull_NegativeCoordinates(t *testing.T) {
	// Test with all negative coordinates
	coords := geom.NewCoordinateSequenceXY(
		-100, -100,
		-50, -100,
		-50, -50,
		-100, -50,
	)
	result := algorithm.ConvexHullFromCoords(coords)
	assert.Equal(t, "Polygon", result.GeometryType(), "Expected Polygon for negative coordinates")
	assert.Equal(t, 5, len(result.Coordinates()), "Expected 5 coordinates for square hull")

	// Test with very negative coordinates (potential integer underflow in old code)
	coordsLarge := geom.NewCoordinateSequenceXY(
		-1e10, -1e10,
		-1e10+100, -1e10,
		-1e10+100, -1e10+100,
		-1e10, -1e10+100,
	)
	resultLarge := algorithm.ConvexHullFromCoords(coordsLarge)
	assert.Equal(t, "Polygon", resultLarge.GeometryType(), "Expected Polygon for large negative coordinates")
}

func TestConvexHull_CloseCoordinates(t *testing.T) {
	// Test with points differing by less than 1e-9
	// These should be treated as distinct points
	coords := geom.NewCoordinateSequenceXY(
		0, 0,
		1e-10, 0,
		1e-10, 1e-10,
		0, 1e-10,
	)
	result := algorithm.ConvexHullFromCoords(coords)
	// These are distinct points and should form a valid hull
	assert.False(t, result.IsEmpty(), "Result should not be empty for close coordinates")

	// Test with slightly different coordinates that old code would collide
	coords2 := geom.NewCoordinateSequenceXY(
		0.0000000001, 0.0000000001,
		0.0000000002, 0.0000000001,
		0.0000000002, 0.0000000002,
		0.0000000001, 0.0000000002,
	)
	result2 := algorithm.ConvexHullFromCoords(coords2)
	assert.False(t, result2.IsEmpty(), "Result should not be empty for very close coordinates")
}

func TestConvexHull_MixedMagnitudes(t *testing.T) {
	// Test with a mix of very large and very small coordinates
	coords := geom.NewCoordinateSequenceXY(
		0, 0,
		1e-15, 0,
		1e10, 0,
		1e10, 1e10,
		0, 1e10,
	)
	result := algorithm.ConvexHullFromCoords(coords)
	assert.Equal(t, "Polygon", result.GeometryType(), "Expected Polygon for mixed magnitude coordinates")

	// Ensure all distinct points are preserved in uniquing
	coordsDistinct := geom.NewCoordinateSequenceXY(
		0, 0,
		1e-20, 1e-20,
		1e15, 1e15,
	)
	resultDistinct := algorithm.ConvexHullFromCoords(coordsDistinct)
	assert.False(t, resultDistinct.IsEmpty(), "Result should not be empty")
}

func TestConvexHull_NoPointsDropped(t *testing.T) {
	// Verify all distinct points contribute to hull computation
	// Create a set of points where each is on the convex hull
	coords := geom.NewCoordinateSequenceXY(
		0, 0,
		10, 0,
		10, 10,
		0, 10,
	)

	result := algorithm.ConvexHullFromCoords(coords)
	assert.Equal(t, "Polygon", result.GeometryType())
	resultCoords := result.Coordinates()
	// Should have 5 points (4 corners + closing)
	assert.Equal(t, 5, len(resultCoords), "All 4 distinct hull points plus closing should be present")

	// Test with many distinct points - ensure none are lost due to key collisions
	var manyCoords geom.CoordinateSequence
	for i := 0; i < 100; i++ {
		manyCoords = append(manyCoords, geom.Coordinate{X: float64(i), Y: 0})
	}
	// Add points to form a triangle hull
	manyCoords = append(manyCoords, geom.Coordinate{X: 50, Y: 100})

	resultMany := algorithm.ConvexHullFromCoords(manyCoords)
	assert.Equal(t, "Polygon", resultMany.GeometryType())
	// Hull should be a triangle with 3 vertices: (0,0), (99,0), (50,100)
	manyResultCoords := resultMany.Coordinates()
	assert.Equal(t, 4, len(manyResultCoords), "Expected triangle hull with 4 coords (3 + closing)")
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
