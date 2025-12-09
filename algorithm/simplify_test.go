package algorithm_test

import (
	"testing"

	"github.com/go-topology-suite/gts/algorithm"
	"github.com/go-topology-suite/gts/geom"
)

func TestDouglasPeucker(t *testing.T) {
	tests := []struct {
		name       string
		geom       geom.Geometry
		tolerance  float64
		maxPoints  int
	}{
		{
			name:      "Point",
			geom:      geom.NewPoint(5, 5),
			tolerance: 1.0,
			maxPoints: 1,
		},
		{
			name:      "LineString",
			geom:      geom.NewLineStringXY(0, 0, 1, 0.1, 2, -0.1, 3, 0.1, 4, -0.1, 5, 0),
			tolerance: 0.5,
			maxPoints: 6, // Should simplify to fewer points
		},
		{
			name:      "LinearRing",
			geom:      geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0),
			tolerance: 1.0,
			maxPoints: 5,
		},
		{
			name:      "Polygon",
			geom:      geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
			tolerance: 1.0,
			maxPoints: 10,
		},
		{
			name:      "MultiPoint",
			geom:      geom.NewMultiPoint([]*geom.Point{geom.NewPoint(0, 0), geom.NewPoint(10, 10)}),
			tolerance: 1.0,
			maxPoints: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.DouglasPeucker(tt.geom, tt.tolerance)
			if result.IsEmpty() && !tt.geom.IsEmpty() {
				t.Error("Result should not be empty")
			}
			coords := result.Coordinates()
			if len(coords) > tt.maxPoints {
				t.Errorf("Expected at most %d points, got %d", tt.maxPoints, len(coords))
			}
		})
	}
}

func TestDouglasPeuckerLineString(t *testing.T) {
	// Create a zigzag line
	ls := geom.NewLineStringXY(0, 0, 1, 0.1, 2, -0.1, 3, 0.1, 4, -0.1, 5, 0)

	tests := []struct {
		name      string
		tolerance float64
		minPoints int
		maxPoints int
	}{
		{
			name:      "HighTolerance",
			tolerance: 1.0,
			minPoints: 2,
			maxPoints: 6,
		},
		{
			name:      "LowTolerance",
			tolerance: 0.01,
			minPoints: 2,
			maxPoints: 7,
		},
		{
			name:      "ZeroTolerance",
			tolerance: 0.0,
			minPoints: 6,
			maxPoints: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.DouglasPeucker(ls, tt.tolerance)
			coords := result.Coordinates()
			if len(coords) < tt.minPoints {
				t.Errorf("Expected at least %d points, got %d", tt.minPoints, len(coords))
			}
			if len(coords) > tt.maxPoints {
				t.Errorf("Expected at most %d points, got %d", tt.maxPoints, len(coords))
			}
		})
	}

	// Test very short line
	t.Run("ShortLine", func(t *testing.T) {
		shortLine := geom.NewLineStringXY(0, 0, 1, 0)
		result := algorithm.DouglasPeucker(shortLine, 1.0)
		coords := result.Coordinates()
		if len(coords) != 2 {
			t.Errorf("Expected 2 points for short line, got %d", len(coords))
		}
	})
}

func TestDouglasPeuckerPolygon(t *testing.T) {
	// Create polygon with extra vertices
	shell := geom.NewLinearRingXY(0, 0, 5, 0.1, 10, 0, 10, 5, 10.1, 10, 10, 10, 5, 10, 0, 10, 0, 5, 0, 0)
	poly := geom.NewPolygon(shell, nil)

	result := algorithm.DouglasPeucker(poly, 0.5)
	if result.GeometryType() != "Polygon" {
		t.Errorf("Expected Polygon, got %s", result.GeometryType())
	}

	resultPoly := result.(*geom.Polygon)
	coords := resultPoly.ExteriorRing().Coordinates()
	if len(coords) < 4 { // Minimum for a valid ring
		t.Errorf("Expected at least 4 points (including closure), got %d", len(coords))
	}

	// Test polygon with hole
	t.Run("PolygonWithHole", func(t *testing.T) {
		shellWithHole := geom.NewLinearRingXY(0, 0, 20, 0, 20, 20, 0, 20, 0, 0)
		hole := geom.NewLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
		polyWithHole := geom.NewPolygon(shellWithHole, []*geom.LinearRing{hole})

		resultWithHole := algorithm.DouglasPeucker(polyWithHole, 1.0)
		resultPolyWithHole := resultWithHole.(*geom.Polygon)

		if resultPolyWithHole.NumInteriorRings() != 1 {
			t.Errorf("Expected 1 hole, got %d", resultPolyWithHole.NumInteriorRings())
		}
	})
}

func TestDouglasPeuckerMultiGeometries(t *testing.T) {
	// Test MultiLineString
	t.Run("MultiLineString", func(t *testing.T) {
		mls := geom.NewMultiLineString([]*geom.LineString{
			geom.NewLineStringXY(0, 0, 5, 0, 10, 0),
			geom.NewLineStringXY(0, 10, 5, 10, 10, 10),
		})
		result := algorithm.DouglasPeucker(mls, 1.0)
		if result.GeometryType() != "MultiLineString" {
			t.Errorf("Expected MultiLineString, got %s", result.GeometryType())
		}
	})

	// Test MultiPolygon
	t.Run("MultiPolygon", func(t *testing.T) {
		mp := geom.NewMultiPolygon([]*geom.Polygon{
			geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
			geom.NewPolygon(geom.NewLinearRingXY(20, 20, 30, 20, 30, 30, 20, 30, 20, 20), nil),
		})
		result := algorithm.DouglasPeucker(mp, 1.0)
		if result.GeometryType() != "MultiPolygon" {
			t.Errorf("Expected MultiPolygon, got %s", result.GeometryType())
		}
	})

	// Test GeometryCollection
	t.Run("GeometryCollection", func(t *testing.T) {
		gc := geom.NewGeometryCollection([]geom.Geometry{
			geom.NewPoint(5, 5),
			geom.NewLineStringXY(0, 0, 10, 0),
		})
		result := algorithm.DouglasPeucker(gc, 1.0)
		if result.GeometryType() != "GeometryCollection" {
			t.Errorf("Expected GeometryCollection, got %s", result.GeometryType())
		}
	})
}

func TestVisvalingamWhyatt(t *testing.T) {
	tests := []struct {
		name      string
		geom      geom.Geometry
		threshold float64
	}{
		{
			name:      "LineString",
			geom:      geom.NewLineStringXY(0, 0, 1, 1, 2, 0, 3, 1, 4, 0),
			threshold: 0.5,
		},
		{
			name:      "Polygon",
			geom:      geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 5, 12, 0, 10, 0, 0), nil),
			threshold: 1.0,
		},
		{
			name:      "Point",
			geom:      geom.NewPoint(5, 5),
			threshold: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.VisvalingamWhyatt(tt.geom, tt.threshold)
			if result.IsEmpty() && !tt.geom.IsEmpty() {
				t.Error("Result should not be empty")
			}
		})
	}

	// Test with very small threshold
	t.Run("SmallThreshold", func(t *testing.T) {
		ls := geom.NewLineStringXY(0, 0, 1, 1, 2, 0, 3, 1, 4, 0)
		result := algorithm.VisvalingamWhyatt(ls, 0.01)
		coords := result.Coordinates()
		// Should keep most points with small threshold
		if len(coords) < 3 {
			t.Errorf("Expected at least 3 points, got %d", len(coords))
		}
	})

	// Test polygon with hole
	t.Run("PolygonWithHole", func(t *testing.T) {
		shell := geom.NewLinearRingXY(0, 0, 20, 0, 20, 20, 0, 20, 0, 0)
		hole := geom.NewLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
		poly := geom.NewPolygon(shell, []*geom.LinearRing{hole})

		result := algorithm.VisvalingamWhyatt(poly, 10.0)
		resultPoly := result.(*geom.Polygon)

		if resultPoly.IsEmpty() {
			t.Error("Result should not be empty")
		}
	})
}

func TestRadialDistance(t *testing.T) {
	tests := []struct {
		name      string
		geom      geom.Geometry
		threshold float64
		maxPoints int
	}{
		{
			name:      "LineString",
			geom:      geom.NewLineStringXY(0, 0, 0.1, 0, 0.2, 0, 5, 0, 5.1, 0, 10, 0),
			threshold: 1.0,
			maxPoints: 4,
		},
		{
			name:      "ShortLine",
			geom:      geom.NewLineStringXY(0, 0, 1, 0),
			threshold: 1.0,
			maxPoints: 2,
		},
		{
			name:      "Point",
			geom:      geom.NewPoint(5, 5),
			threshold: 1.0,
			maxPoints: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.RadialDistance(tt.geom, tt.threshold)
			coords := result.Coordinates()
			if len(coords) > tt.maxPoints {
				t.Errorf("Expected at most %d points, got %d", tt.maxPoints, len(coords))
			}
		})
	}

	// Test that endpoints are always kept
	t.Run("EndpointsKept", func(t *testing.T) {
		ls := geom.NewLineStringXY(0, 0, 0.1, 0, 0.2, 0, 10, 0)
		result := algorithm.RadialDistance(ls, 1.0)
		coords := result.Coordinates()

		if len(coords) < 2 {
			t.Error("Expected at least 2 points (start and end)")
		}

		firstCoord := ls.Coordinates()[0]
		lastCoord := ls.Coordinates()[len(ls.Coordinates())-1]
		resultFirst := coords[0]
		resultLast := coords[len(coords)-1]

		if !firstCoord.Equals2D(resultFirst, 0.001) {
			t.Error("First point should be preserved")
		}
		if !lastCoord.Equals2D(resultLast, 0.001) {
			t.Error("Last point should be preserved")
		}
	})
}

func TestSimplifyEdgeCases(t *testing.T) {
	// Test empty geometries
	t.Run("EmptyLineString", func(t *testing.T) {
		empty := geom.NewLineStringEmpty()
		result := algorithm.DouglasPeucker(empty, 1.0)
		if !result.IsEmpty() {
			t.Error("Result should be empty")
		}
	})

	t.Run("EmptyPolygon", func(t *testing.T) {
		empty := geom.NewPolygonEmpty()
		result := algorithm.DouglasPeucker(empty, 1.0)
		if !result.IsEmpty() {
			t.Error("Result should be empty")
		}
	})

	// Test degenerate cases
	t.Run("TwoPointLineString", func(t *testing.T) {
		ls := geom.NewLineStringXY(0, 0, 10, 0)
		result := algorithm.DouglasPeucker(ls, 1.0)
		coords := result.Coordinates()
		if len(coords) != 2 {
			t.Errorf("Expected 2 points, got %d", len(coords))
		}
	})

	t.Run("TriangleRing", func(t *testing.T) {
		ring := geom.NewLinearRingXY(0, 0, 10, 0, 5, 10, 0, 0)
		result := algorithm.DouglasPeucker(ring, 1.0)
		coords := result.Coordinates()
		if len(coords) < 4 {
			t.Errorf("Expected at least 4 points (triangle + closure), got %d", len(coords))
		}
	})

	// Test with very high tolerance
	t.Run("HighTolerance", func(t *testing.T) {
		ls := geom.NewLineStringXY(0, 0, 1, 0.1, 2, -0.1, 3, 0.1, 4, 0)
		result := algorithm.DouglasPeucker(ls, 10.0)
		coords := result.Coordinates()
		// Should simplify to just endpoints
		if len(coords) != 2 {
			t.Errorf("Expected 2 points with high tolerance, got %d", len(coords))
		}
	})

	// Test linear ring that simplifies too much
	t.Run("LinearRingMinimumPoints", func(t *testing.T) {
		ring := geom.NewLinearRingXY(0, 0, 1, 0, 2, 0, 3, 0, 0, 0)
		result := algorithm.DouglasPeucker(ring, 10.0)
		coords := result.Coordinates()
		// Should maintain at least 4 points for a valid ring
		if len(coords) < 4 {
			t.Errorf("Expected at least 4 points for ring, got %d", len(coords))
		}
	})
}

func TestSimplifyPreservesGeometryType(t *testing.T) {
	tests := []struct {
		name         string
		geom         geom.Geometry
		expectedType string
	}{
		{
			name:         "LineString",
			geom:         geom.NewLineStringXY(0, 0, 5, 0, 10, 0),
			expectedType: "LineString",
		},
		{
			name:         "Polygon",
			geom:         geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil),
			expectedType: "Polygon",
		},
		{
			name:         "MultiLineString",
			geom:         geom.NewMultiLineString([]*geom.LineString{geom.NewLineStringXY(0, 0, 10, 0)}),
			expectedType: "MultiLineString",
		},
		{
			name:         "MultiPolygon",
			geom:         geom.NewMultiPolygon([]*geom.Polygon{geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)}),
			expectedType: "MultiPolygon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.DouglasPeucker(tt.geom, 1.0)
			if result.GeometryType() != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, result.GeometryType())
			}
		})
	}
}
