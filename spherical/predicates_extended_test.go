package spherical

import (
	"testing"

	"github.com/go-topology-suite/gts/geom"
)

// Test GenericWithin with various geometry types
func TestGenericWithin(t *testing.T) {
	tests := []struct {
		name     string
		g1       geom.Geometry
		g2       geom.Geometry
		expected bool
	}{
		{
			name: "point within polygon",
			g1:   geom.NewPoint(0, 0),
			g2: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: -1, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 1}, {X: -1, Y: 1}, {X: -1, Y: -1},
				}),
				nil,
			),
			expected: true,
		},
		{
			name: "point outside polygon",
			g1:   geom.NewPoint(5, 5),
			g2: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: -1, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 1}, {X: -1, Y: 1}, {X: -1, Y: -1},
				}),
				nil,
			),
			expected: false,
		},
		{
			name: "small polygon within larger polygon",
			g1: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: -0.5, Y: -0.5}, {X: 0.5, Y: -0.5}, {X: 0.5, Y: 0.5}, {X: -0.5, Y: 0.5}, {X: -0.5, Y: -0.5},
				}),
				nil,
			),
			g2: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: -1, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 1}, {X: -1, Y: 1}, {X: -1, Y: -1},
				}),
				nil,
			),
			expected: true,
		},
		{
			name: "linestring within polygon",
			g1: geom.NewLineString(geom.CoordinateSequence{
				{X: -0.5, Y: 0}, {X: 0.5, Y: 0},
			}),
			g2: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: -1, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 1}, {X: -1, Y: 1}, {X: -1, Y: -1},
				}),
				nil,
			),
			expected: true,
		},
		{
			name:     "empty geometries",
			g1:       geom.NewPointEmpty(),
			g2:       geom.NewPolygonEmpty(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenericWithin(tt.g1, tt.g2)
			if result != tt.expected {
				t.Errorf("GenericWithin() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test GenericDisjoint with various geometry types
func TestGenericDisjoint(t *testing.T) {
	tests := []struct {
		name     string
		g1       geom.Geometry
		g2       geom.Geometry
		expected bool
	}{
		{
			name:   "disjoint polygons",
			g1: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0},
				}),
				nil,
			),
			g2: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: 2, Y: 2}, {X: 3, Y: 2}, {X: 3, Y: 3}, {X: 2, Y: 3}, {X: 2, Y: 2},
				}),
				nil,
			),
			expected: true,
		},
		{
			name:   "overlapping polygons",
			g1: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: 0, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 2}, {X: 0, Y: 2}, {X: 0, Y: 0},
				}),
				nil,
			),
			g2: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: 1, Y: 1}, {X: 3, Y: 1}, {X: 3, Y: 3}, {X: 1, Y: 3}, {X: 1, Y: 1},
				}),
				nil,
			),
			expected: false,
		},
		{
			name:   "disjoint points",
			g1:     geom.NewPoint(0, 0),
			g2:     geom.NewPoint(5, 5),
			expected: true,
		},
		{
			name:   "disjoint linestrings",
			g1: geom.NewLineString(geom.CoordinateSequence{
				{X: 0, Y: 0}, {X: 1, Y: 0},
			}),
			g2: geom.NewLineString(geom.CoordinateSequence{
				{X: 0, Y: 2}, {X: 1, Y: 2},
			}),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenericDisjoint(tt.g1, tt.g2)
			if result != tt.expected {
				t.Errorf("GenericDisjoint() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test GenericOverlaps with various geometry types
func TestGenericOverlaps(t *testing.T) {
	tests := []struct {
		name     string
		g1       geom.Geometry
		g2       geom.Geometry
		expected bool
	}{
		{
			name:   "overlapping polygons",
			g1: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: 0, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 2}, {X: 0, Y: 2}, {X: 0, Y: 0},
				}),
				nil,
			),
			g2: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: 1, Y: 1}, {X: 3, Y: 1}, {X: 3, Y: 3}, {X: 1, Y: 3}, {X: 1, Y: 1},
				}),
				nil,
			),
			expected: true,
		},
		{
			name:   "disjoint polygons",
			g1: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0},
				}),
				nil,
			),
			g2: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: 2, Y: 2}, {X: 3, Y: 2}, {X: 3, Y: 3}, {X: 2, Y: 3}, {X: 2, Y: 2},
				}),
				nil,
			),
			expected: false,
		},
		{
			name:   "one polygon contains another",
			g1: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: -0.5, Y: -0.5}, {X: 0.5, Y: -0.5}, {X: 0.5, Y: 0.5}, {X: -0.5, Y: 0.5}, {X: -0.5, Y: -0.5},
				}),
				nil,
			),
			g2: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: -1, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 1}, {X: -1, Y: 1}, {X: -1, Y: -1},
				}),
				nil,
			),
			expected: false,
		},
		{
			name:   "different dimensions - point and polygon",
			g1:     geom.NewPoint(0, 0),
			g2: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: -1, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 1}, {X: -1, Y: 1}, {X: -1, Y: -1},
				}),
				nil,
			),
			expected: false,
		},
		{
			name: "overlapping linestrings",
			g1: geom.NewLineString(geom.CoordinateSequence{
				{X: 0, Y: 0}, {X: 2, Y: 0},
			}),
			g2: geom.NewLineString(geom.CoordinateSequence{
				{X: 1, Y: 0}, {X: 3, Y: 0},
			}),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenericOverlaps(tt.g1, tt.g2)
			if result != tt.expected {
				t.Errorf("GenericOverlaps() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test GenericTouches with various geometry types
func TestGenericTouches(t *testing.T) {
	tests := []struct {
		name     string
		g1       geom.Geometry
		g2       geom.Geometry
		expected bool
	}{
		{
			name:   "touching polygons",
			g1: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0},
				}),
				nil,
			),
			g2: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: 1, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 1}, {X: 1, Y: 1}, {X: 1, Y: 0},
				}),
				nil,
			),
			expected: true,
		},
		{
			name:   "overlapping polygons",
			g1: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: 0, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 2}, {X: 0, Y: 2}, {X: 0, Y: 0},
				}),
				nil,
			),
			g2: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: 1, Y: 1}, {X: 3, Y: 1}, {X: 3, Y: 3}, {X: 1, Y: 3}, {X: 1, Y: 1},
				}),
				nil,
			),
			expected: false,
		},
		{
			name:   "disjoint polygons",
			g1: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0},
				}),
				nil,
			),
			g2: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: 2, Y: 2}, {X: 3, Y: 2}, {X: 3, Y: 3}, {X: 2, Y: 3}, {X: 2, Y: 2},
				}),
				nil,
			),
			expected: false,
		},
		{
			name: "point touching polygon boundary",
			g1:   geom.NewPoint(1, 0.5),
			g2: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0},
				}),
				nil,
			),
			expected: true,
		},
		{
			name: "linestring touching polygon boundary",
			g1: geom.NewLineString(geom.CoordinateSequence{
				{X: 1, Y: 0}, {X: 1, Y: 1},
			}),
			g2: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0},
				}),
				nil,
			),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenericTouches(tt.g1, tt.g2)
			if result != tt.expected {
				t.Errorf("GenericTouches() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test locatePointSpherical with various geometry types
func TestLocatePointSpherical(t *testing.T) {
	tests := []struct {
		name     string
		point    geom.Coordinate
		geom     geom.Geometry
		expected geom.Location
	}{
		{
			name:  "point in polygon interior",
			point: geom.NewCoordinate(0, 0),
			geom: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: -1, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 1}, {X: -1, Y: 1}, {X: -1, Y: -1},
				}),
				nil,
			),
			expected: geom.LocationInterior,
		},
		{
			name:  "point on polygon boundary",
			point: geom.NewCoordinate(1, 0),
			geom: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0},
				}),
				nil,
			),
			expected: geom.LocationBoundary,
		},
		{
			name:  "point exterior to polygon",
			point: geom.NewCoordinate(5, 5),
			geom: geom.NewPolygon(
				geom.NewLinearRing(geom.CoordinateSequence{
					{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0},
				}),
				nil,
			),
			expected: geom.LocationExterior,
		},
		{
			name:  "point on linestring",
			point: geom.NewCoordinate(0.5, 0),
			geom: geom.NewLineString(geom.CoordinateSequence{
				{X: 0, Y: 0}, {X: 1, Y: 0},
			}),
			expected: geom.LocationInterior,
		},
		{
			name:  "point at linestring endpoint",
			point: geom.NewCoordinate(0, 0),
			geom: geom.NewLineString(geom.CoordinateSequence{
				{X: 0, Y: 0}, {X: 1, Y: 0},
			}),
			expected: geom.LocationBoundary,
		},
		{
			name:     "point matches point",
			point:    geom.NewCoordinate(0, 0),
			geom:     geom.NewPoint(0, 0),
			expected: geom.LocationInterior,
		},
		{
			name:     "point doesn't match point",
			point:    geom.NewCoordinate(5, 5),
			geom:     geom.NewPoint(0, 0),
			expected: geom.LocationExterior,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := locatePointSpherical(tt.point, tt.geom)
			if result != tt.expected {
				t.Errorf("locatePointSpherical() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test with MultiGeometry types
func TestGenericPredicatesWithMultiGeometries(t *testing.T) {
	multiPoint := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(0, 0),
		geom.NewPoint(1, 1),
		geom.NewPoint(2, 2),
	})

	multiLineString := geom.NewMultiLineString([]*geom.LineString{
		geom.NewLineString(geom.CoordinateSequence{{X: 0, Y: 0}, {X: 1, Y: 0}}),
		geom.NewLineString(geom.CoordinateSequence{{X: 0, Y: 1}, {X: 1, Y: 1}}),
	})

	multiPolygon := geom.NewMultiPolygon([]*geom.Polygon{
		geom.NewPolygon(
			geom.NewLinearRing(geom.CoordinateSequence{
				{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0},
			}),
			nil,
		),
		geom.NewPolygon(
			geom.NewLinearRing(geom.CoordinateSequence{
				{X: 2, Y: 2}, {X: 3, Y: 2}, {X: 3, Y: 3}, {X: 2, Y: 3}, {X: 2, Y: 2},
			}),
			nil,
		),
	})

	largePoly := geom.NewPolygon(
		geom.NewLinearRing(geom.CoordinateSequence{
			{X: -1, Y: -1}, {X: 4, Y: -1}, {X: 4, Y: 4}, {X: -1, Y: 4}, {X: -1, Y: -1},
		}),
		nil,
	)

	t.Run("multipoint within polygon", func(t *testing.T) {
		if !GenericWithin(multiPoint, largePoly) {
			t.Error("Expected multipoint to be within large polygon")
		}
	})

	t.Run("multilinestring intersects polygon", func(t *testing.T) {
		if !Intersects(multiLineString, largePoly) {
			t.Error("Expected multilinestring to intersect large polygon")
		}
	})

	t.Run("multipolygon within large polygon", func(t *testing.T) {
		if !GenericWithin(multiPolygon, largePoly) {
			t.Error("Expected multipolygon to be within large polygon")
		}
	})

	t.Run("multipolygon overlaps itself", func(t *testing.T) {
		if GenericOverlaps(multiPolygon, multiPolygon) {
			t.Error("Expected geometry not to overlap with itself")
		}
	})
}
