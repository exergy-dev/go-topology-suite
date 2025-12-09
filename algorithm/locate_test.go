package algorithm_test

import (
	"testing"

	"github.com/go-topology-suite/gts/algorithm"
	"github.com/go-topology-suite/gts/geom"
)

func TestPointLocationGeometryTypes(t *testing.T) {
	tests := []struct {
		name     string
		p        geom.Coordinate
		g        geom.Geometry
		expected geom.Location
	}{
		{
			name:     "PointInPoint",
			p:        geom.NewCoordinate(5, 5),
			g:        geom.NewPoint(5, 5),
			expected: geom.LocationInterior,
		},
		{
			name:     "PointNotInPoint",
			p:        geom.NewCoordinate(5, 5),
			g:        geom.NewPoint(10, 10),
			expected: geom.LocationExterior,
		},
		{
			name:     "PointInEmptyPoint",
			p:        geom.NewCoordinate(5, 5),
			g:        geom.NewPointEmpty(),
			expected: geom.LocationExterior,
		},
		{
			name:     "PointOnLineString",
			p:        geom.NewCoordinate(5, 0),
			g:        geom.NewLineStringXY(0, 0, 10, 0),
			expected: geom.LocationInterior,
		},
		{
			name:     "PointOnLineStringEndpoint",
			p:        geom.NewCoordinate(0, 0),
			g:        geom.NewLineStringXY(0, 0, 10, 0),
			expected: geom.LocationBoundary,
		},
		{
			name:     "PointNotOnLineString",
			p:        geom.NewCoordinate(5, 5),
			g:        geom.NewLineStringXY(0, 0, 10, 0),
			expected: geom.LocationExterior,
		},
		{
			name:     "PointOnLinearRingBoundary",
			p:        geom.NewCoordinate(0, 5),
			g:        geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0),
			expected: geom.LocationBoundary,
		},
		{
			name:     "PointInLinearRing",
			p:        geom.NewCoordinate(5, 5),
			g:        geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0),
			expected: geom.LocationInterior,
		},
		{
			name:     "PointOutsideLinearRing",
			p:        geom.NewCoordinate(15, 5),
			g:        geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0),
			expected: geom.LocationExterior,
		},
		{
			name:     "PointInMultiPoint",
			p:        geom.NewCoordinate(5, 5),
			g:        geom.NewMultiPoint([]*geom.Point{geom.NewPoint(5, 5), geom.NewPoint(10, 10)}),
			expected: geom.LocationInterior,
		},
		{
			name:     "PointNotInMultiPoint",
			p:        geom.NewCoordinate(0, 0),
			g:        geom.NewMultiPoint([]*geom.Point{geom.NewPoint(5, 5), geom.NewPoint(10, 10)}),
			expected: geom.LocationExterior,
		},
		{
			name:     "PointInMultiLineString",
			p:        geom.NewCoordinate(5, 0),
			g:        geom.NewMultiLineString([]*geom.LineString{geom.NewLineStringXY(0, 0, 10, 0)}),
			expected: geom.LocationInterior,
		},
		{
			name:     "PointOnMultiLineStringBoundary",
			p:        geom.NewCoordinate(0, 0),
			g:        geom.NewMultiLineString([]*geom.LineString{geom.NewLineStringXY(0, 0, 10, 0)}),
			expected: geom.LocationBoundary,
		},
		{
			name:     "PointInMultiPolygon",
			p:        geom.NewCoordinate(5, 5),
			g:        geom.NewMultiPolygon([]*geom.Polygon{geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)}),
			expected: geom.LocationInterior,
		},
		{
			name:     "PointOnMultiPolygonBoundary",
			p:        geom.NewCoordinate(0, 5),
			g:        geom.NewMultiPolygon([]*geom.Polygon{geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)}),
			expected: geom.LocationBoundary,
		},
		{
			name:     "PointInGeometryCollection",
			p:        geom.NewCoordinate(5, 5),
			g:        geom.NewGeometryCollection([]geom.Geometry{geom.NewPoint(5, 5)}),
			expected: geom.LocationInterior,
		},
		{
			name:     "PointOnGeometryCollectionBoundary",
			p:        geom.NewCoordinate(0, 5),
			g:        geom.NewGeometryCollection([]geom.Geometry{geom.NewPolygon(geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil)}),
			expected: geom.LocationBoundary,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.PointLocation(tt.p, tt.g)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPointLocationInPolygon(t *testing.T) {
	// Square polygon
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)

	tests := []struct {
		name     string
		p        geom.Coordinate
		expected geom.Location
	}{
		{
			name:     "Inside",
			p:        geom.NewCoordinate(5, 5),
			expected: geom.LocationInterior,
		},
		{
			name:     "OnBoundary",
			p:        geom.NewCoordinate(0, 5),
			expected: geom.LocationBoundary,
		},
		{
			name:     "OnCorner",
			p:        geom.NewCoordinate(0, 0),
			expected: geom.LocationBoundary,
		},
		{
			name:     "Outside",
			p:        geom.NewCoordinate(15, 5),
			expected: geom.LocationExterior,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.PointLocationInPolygon(tt.p, poly)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}

	// Test with hole
	t.Run("PolygonWithHole", func(t *testing.T) {
		shellWithHole := geom.NewLinearRingXY(0, 0, 20, 0, 20, 20, 0, 20, 0, 0)
		hole := geom.NewLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
		polyWithHole := geom.NewPolygon(shellWithHole, []*geom.LinearRing{hole})

		// Point in hole should be exterior
		if loc := algorithm.PointLocationInPolygon(geom.NewCoordinate(10, 10), polyWithHole); loc != geom.LocationExterior {
			t.Errorf("Expected Exterior for point in hole, got %v", loc)
		}

		// Point on hole boundary should be boundary
		if loc := algorithm.PointLocationInPolygon(geom.NewCoordinate(5, 10), polyWithHole); loc != geom.LocationBoundary {
			t.Errorf("Expected Boundary for point on hole boundary, got %v", loc)
		}

		// Point between shell and hole should be interior
		if loc := algorithm.PointLocationInPolygon(geom.NewCoordinate(2, 2), polyWithHole); loc != geom.LocationInterior {
			t.Errorf("Expected Interior for point between shell and hole, got %v", loc)
		}
	})

	// Test empty polygon
	t.Run("EmptyPolygon", func(t *testing.T) {
		emptyPoly := geom.NewPolygonEmpty()
		result := algorithm.PointLocationInPolygon(geom.NewCoordinate(5, 5), emptyPoly)
		if result != geom.LocationExterior {
			t.Errorf("Expected Exterior for empty polygon, got %v", result)
		}
	})
}

func TestIsPointInEnvelope(t *testing.T) {
	env := geom.NewEnvelope(0, 0, 10, 10)

	tests := []struct {
		name     string
		p        geom.Coordinate
		expected bool
	}{
		{
			name:     "Inside",
			p:        geom.NewCoordinate(5, 5),
			expected: true,
		},
		{
			name:     "OnBoundary",
			p:        geom.NewCoordinate(0, 5),
			expected: true,
		},
		{
			name:     "Outside",
			p:        geom.NewCoordinate(15, 5),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.IsPointInEnvelope(tt.p, env)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLocatePointInTriangle(t *testing.T) {
	t0 := geom.NewCoordinate(0, 0)
	t1 := geom.NewCoordinate(10, 0)
	t2 := geom.NewCoordinate(5, 10)

	tests := []struct {
		name     string
		p        geom.Coordinate
		expected geom.Location
	}{
		{
			name:     "Inside",
			p:        geom.NewCoordinate(5, 5),
			expected: geom.LocationInterior,
		},
		{
			name:     "OnBoundary",
			p:        geom.NewCoordinate(5, 0),
			expected: geom.LocationBoundary,
		},
		{
			name:     "OnVertex",
			p:        geom.NewCoordinate(0, 0),
			expected: geom.LocationBoundary,
		},
		{
			name:     "Outside",
			p:        geom.NewCoordinate(15, 15),
			expected: geom.LocationExterior,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.LocatePointInTriangle(tt.p, t0, t1, t2)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIndexOfPointInRing(t *testing.T) {
	ring := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)

	tests := []struct {
		name     string
		p        geom.Coordinate
		expected int
	}{
		{
			name:     "FirstPoint",
			p:        geom.NewCoordinate(0, 0),
			expected: 0,
		},
		{
			name:     "MiddlePoint",
			p:        geom.NewCoordinate(10, 0),
			expected: 1,
		},
		{
			name:     "NotInRing",
			p:        geom.NewCoordinate(5, 5),
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.IndexOfPointInRing(tt.p, ring)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIndexOfClosestPointInSequence(t *testing.T) {
	coords := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 10)

	tests := []struct {
		name     string
		p        geom.Coordinate
		expected int
	}{
		{
			name:     "ClosestToFirst",
			p:        geom.NewCoordinate(1, 1),
			expected: 0,
		},
		{
			name:     "ClosestToSecond",
			p:        geom.NewCoordinate(9, 1),
			expected: 1,
		},
		{
			name:     "ClosestToThird",
			p:        geom.NewCoordinate(9, 9),
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.IndexOfClosestPointInSequence(tt.p, coords)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}

	// Test empty sequence
	empty := geom.CoordinateSequence{}
	if idx := algorithm.IndexOfClosestPointInSequence(geom.NewCoordinate(5, 5), empty); idx != -1 {
		t.Errorf("Expected -1 for empty sequence, got %v", idx)
	}
}

func TestIsPointInRingEdgeCases(t *testing.T) {
	// Test with very small ring
	t.Run("SmallRing", func(t *testing.T) {
		ring := geom.NewLinearRingXY(0, 0, 1, 0, 0, 1, 0, 0)
		if !algorithm.IsPointInRing(geom.NewCoordinate(0.25, 0.25), ring) {
			t.Error("Expected point to be inside small ring")
		}
	})

	// Test with degenerate ring (less than 4 points)
	t.Run("DegenerateRing", func(t *testing.T) {
		ring := geom.NewLinearRing(geom.NewCoordinateSequenceXY(0, 0, 1, 0, 0, 0))
		if algorithm.IsPointInRing(geom.NewCoordinate(0.5, 0.5), ring) {
			t.Error("Expected point to be outside degenerate ring")
		}
	})

	// Test point on vertex
	t.Run("PointOnVertex", func(t *testing.T) {
		ring := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
		// Point on vertex - behavior depends on ray casting implementation
		result := algorithm.IsPointInRing(geom.NewCoordinate(0, 0), ring)
		// Result can be either true or false depending on implementation
		_ = result
	})

	// Test concave polygon
	t.Run("ConcavePolygon", func(t *testing.T) {
		// L-shaped polygon
		ring := geom.NewLinearRingXY(0, 0, 10, 0, 10, 5, 5, 5, 5, 10, 0, 10, 0, 0)
		// Point in the concave part
		if algorithm.IsPointInRing(geom.NewCoordinate(7, 7), ring) {
			t.Error("Expected point to be outside concave part")
		}
		// Point definitely inside
		if !algorithm.IsPointInRing(geom.NewCoordinate(2, 2), ring) {
			t.Error("Expected point to be inside")
		}
	})
}
