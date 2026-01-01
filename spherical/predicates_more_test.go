package spherical

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

// TestCrosses tests the Crosses predicate.
func TestCrosses(t *testing.T) {
	tests := []struct {
		name     string
		g1       geom.Geometry
		g2       geom.Geometry
		expected bool
	}{
		{
			name:     "Point/Point cannot cross",
			g1:       geom.NewPoint(0.0, 0.0),
			g2:       geom.NewPoint(1.0, 1.0),
			expected: false,
		},
		{
			name: "Polygon/Polygon cannot cross (same dimension)",
			g1: geom.NewPolygon(geom.NewLinearRingXY(
				0.0, 0.0,
				1.0, 0.0,
				1.0, 1.0,
				0.0, 1.0,
				0.0, 0.0,
			), nil),
			g2: geom.NewPolygon(geom.NewLinearRingXY(
				0.5, 0.5,
				1.5, 0.5,
				1.5, 1.5,
				0.5, 1.5,
				0.5, 0.5,
			), nil),
			expected: false,
		},
		{
			name: "Line/Line crossing",
			g1: geom.NewLineStringXY(
				0.0, -1.0,
				0.0, 1.0,
			),
			g2: geom.NewLineStringXY(
				-1.0, 0.0,
				1.0, 0.0,
			),
			expected: true,
		},
		{
			name: "Line/Line parallel (not crossing)",
			g1: geom.NewLineStringXY(
				0.0, 0.0,
				1.0, 0.0,
			),
			g2: geom.NewLineStringXY(
				0.0, 1.0,
				1.0, 1.0,
			),
			expected: false,
		},
		{
			name: "Line/Polygon crossing",
			g1: geom.NewLineStringXY(
				-0.5, 0.0,
				1.5, 0.0,
			),
			g2: geom.NewPolygon(geom.NewLinearRingXY(
				0.0, -1.0,
				1.0, -1.0,
				1.0, 1.0,
				0.0, 1.0,
				0.0, -1.0,
			), nil),
			expected: true,
		},
		{
			name: "Line completely inside polygon (not crossing)",
			g1: geom.NewLineStringXY(
				0.2, 0.0,
				0.8, 0.0,
			),
			g2: geom.NewPolygon(geom.NewLinearRingXY(
				0.0, -1.0,
				1.0, -1.0,
				1.0, 1.0,
				0.0, 1.0,
				0.0, -1.0,
			), nil),
			expected: false,
		},
		{
			name:     "Empty geometries",
			g1:       geom.NewLineStringEmpty(),
			g2:       geom.NewLineStringEmpty(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Crosses(tt.g1, tt.g2)
			if result != tt.expected {
				t.Errorf("Crosses() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestCovers tests the Covers predicate.
func TestCovers(t *testing.T) {
	tests := []struct {
		name     string
		g1       geom.Geometry
		g2       geom.Geometry
		expected bool
	}{
		{
			name: "Polygon covers point inside",
			g1: geom.NewPolygon(geom.NewLinearRingXY(
				0.0, 0.0,
				2.0, 0.0,
				2.0, 2.0,
				0.0, 2.0,
				0.0, 0.0,
			), nil),
			g2:       geom.NewPoint(1.0, 1.0),
			expected: true,
		},
		{
			name: "Polygon does not cover point outside",
			g1: geom.NewPolygon(geom.NewLinearRingXY(
				0.0, 0.0,
				1.0, 0.0,
				1.0, 1.0,
				0.0, 1.0,
				0.0, 0.0,
			), nil),
			g2:       geom.NewPoint(2.0, 2.0),
			expected: false,
		},
		{
			name: "Polygon covers smaller polygon inside",
			g1: geom.NewPolygon(geom.NewLinearRingXY(
				0.0, 0.0,
				2.0, 0.0,
				2.0, 2.0,
				0.0, 2.0,
				0.0, 0.0,
			), nil),
			g2: geom.NewPolygon(geom.NewLinearRingXY(
				0.5, 0.5,
				1.5, 0.5,
				1.5, 1.5,
				0.5, 1.5,
				0.5, 0.5,
			), nil),
			expected: true,
		},
		{
			name: "Polygon does not cover overlapping polygon",
			g1: geom.NewPolygon(geom.NewLinearRingXY(
				0.0, 0.0,
				1.0, 0.0,
				1.0, 1.0,
				0.0, 1.0,
				0.0, 0.0,
			), nil),
			g2: geom.NewPolygon(geom.NewLinearRingXY(
				0.5, 0.5,
				1.5, 0.5,
				1.5, 1.5,
				0.5, 1.5,
				0.5, 0.5,
			), nil),
			expected: false,
		},
		{
			name:     "Empty geometries",
			g1:       geom.NewPolygonEmpty(),
			g2:       geom.NewPointEmpty(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Covers(tt.g1, tt.g2)
			if result != tt.expected {
				t.Errorf("Covers() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestCoveredBy tests the CoveredBy predicate.
func TestCoveredBy(t *testing.T) {
	tests := []struct {
		name     string
		g1       geom.Geometry
		g2       geom.Geometry
		expected bool
	}{
		{
			name: "Point inside is covered by polygon",
			g1:   geom.NewPoint(1.0, 1.0),
			g2: geom.NewPolygon(geom.NewLinearRingXY(
				0.0, 0.0,
				2.0, 0.0,
				2.0, 2.0,
				0.0, 2.0,
				0.0, 0.0,
			), nil),
			expected: true,
		},
		{
			name: "Point outside is not covered by polygon",
			g1:   geom.NewPoint(3.0, 3.0),
			g2: geom.NewPolygon(geom.NewLinearRingXY(
				0.0, 0.0,
				2.0, 0.0,
				2.0, 2.0,
				0.0, 2.0,
				0.0, 0.0,
			), nil),
			expected: false,
		},
		{
			name: "CoveredBy is inverse of Covers",
			g1: geom.NewPolygon(geom.NewLinearRingXY(
				0.5, 0.5,
				1.5, 0.5,
				1.5, 1.5,
				0.5, 1.5,
				0.5, 0.5,
			), nil),
			g2: geom.NewPolygon(geom.NewLinearRingXY(
				0.0, 0.0,
				2.0, 0.0,
				2.0, 2.0,
				0.0, 2.0,
				0.0, 0.0,
			), nil),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CoveredBy(tt.g1, tt.g2)
			if result != tt.expected {
				t.Errorf("CoveredBy() = %v, want %v", result, tt.expected)
			}

			// Verify CoveredBy(a,b) == Covers(b,a)
			inverse := Covers(tt.g2, tt.g1)
			if result != inverse {
				t.Errorf("CoveredBy() and Covers() not inverse: %v vs %v", result, inverse)
			}
		})
	}
}

// TestEquals tests the Equals predicate.
func TestEquals(t *testing.T) {
	tests := []struct {
		name     string
		g1       geom.Geometry
		g2       geom.Geometry
		expected bool
	}{
		{
			name:     "Same point",
			g1:       geom.NewPoint(1.0, 1.0),
			g2:       geom.NewPoint(1.0, 1.0),
			expected: true,
		},
		{
			name:     "Different points",
			g1:       geom.NewPoint(1.0, 1.0),
			g2:       geom.NewPoint(2.0, 2.0),
			expected: false,
		},
		{
			name: "Same polygon",
			g1: geom.NewPolygon(geom.NewLinearRingXY(
				0.0, 0.0,
				1.0, 0.0,
				1.0, 1.0,
				0.0, 1.0,
				0.0, 0.0,
			), nil),
			g2: geom.NewPolygon(geom.NewLinearRingXY(
				0.0, 0.0,
				1.0, 0.0,
				1.0, 1.0,
				0.0, 1.0,
				0.0, 0.0,
			), nil),
			expected: true,
		},
		{
			name: "Different polygons",
			g1: geom.NewPolygon(geom.NewLinearRingXY(
				0.0, 0.0,
				1.0, 0.0,
				1.0, 1.0,
				0.0, 1.0,
				0.0, 0.0,
			), nil),
			g2: geom.NewPolygon(geom.NewLinearRingXY(
				0.0, 0.0,
				2.0, 0.0,
				2.0, 2.0,
				0.0, 2.0,
				0.0, 0.0,
			), nil),
			expected: false,
		},
		{
			name:     "Different geometry types",
			g1:       geom.NewPoint(0.0, 0.0),
			g2:       geom.NewLineStringXY(0.0, 0.0, 1.0, 1.0),
			expected: false,
		},
		{
			name:     "Both empty (same type)",
			g1:       geom.NewPointEmpty(),
			g2:       geom.NewPointEmpty(),
			expected: true,
		},
		{
			name:     "One empty, one not",
			g1:       geom.NewPoint(0.0, 0.0),
			g2:       geom.NewPointEmpty(),
			expected: false,
		},
		{
			name:     "Both nil",
			g1:       nil,
			g2:       nil,
			expected: true,
		},
		{
			name:     "One nil",
			g1:       geom.NewPoint(0.0, 0.0),
			g2:       nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Equals(tt.g1, tt.g2)
			if result != tt.expected {
				t.Errorf("Equals() = %v, want %v", result, tt.expected)
			}

			// Test commutativity
			if tt.g1 != nil && tt.g2 != nil {
				result2 := Equals(tt.g2, tt.g1)
				if result != result2 {
					t.Errorf("Equals() not commutative: %v vs %v", result, result2)
				}
			}
		})
	}
}

// TestHelperFunctions tests the helper functions.
func TestHelperFunctions(t *testing.T) {
	t.Run("getLineStringsFromGeometry", func(t *testing.T) {
		ls := geom.NewLineStringXY(0.0, 0.0, 1.0, 1.0)
		result := getLineStringsFromGeometry(ls)
		if len(result) != 1 {
			t.Errorf("Expected 1 linestring, got %d", len(result))
		}

		// Test with MultiLineString
		mls := geom.NewMultiLineString([]*geom.LineString{
			geom.NewLineStringXY(0.0, 0.0, 1.0, 1.0),
			geom.NewLineStringXY(2.0, 2.0, 3.0, 3.0),
		})
		result = getLineStringsFromGeometry(mls)
		if len(result) != 2 {
			t.Errorf("Expected 2 linestrings, got %d", len(result))
		}
	})

	t.Run("getPolygonsFromGeometry", func(t *testing.T) {
		poly := geom.NewPolygon(geom.NewLinearRingXY(
			0.0, 0.0,
			1.0, 0.0,
			1.0, 1.0,
			0.0, 1.0,
			0.0, 0.0,
		), nil)
		result := getPolygonsFromGeometry(poly)
		if len(result) != 1 {
			t.Errorf("Expected 1 polygon, got %d", len(result))
		}

		// Test with MultiPolygon
		mp := geom.NewMultiPolygon([]*geom.Polygon{poly, poly})
		result = getPolygonsFromGeometry(mp)
		if len(result) != 2 {
			t.Errorf("Expected 2 polygons, got %d", len(result))
		}
	})
}
