package algorithm_test

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/algorithm"
	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
)

func TestLineIntersection(t *testing.T) {
	tests := []struct {
		name               string
		p1, p2, p3, p4     geom.Coordinate
		hasIntersection    bool
		isProper           bool
		isCollinear        bool
		expectedX          float64
		expectedY          float64
	}{
		{
			name:            "ProperIntersection",
			p1:              geom.NewCoordinate(0, 0),
			p2:              geom.NewCoordinate(10, 10),
			p3:              geom.NewCoordinate(0, 10),
			p4:              geom.NewCoordinate(10, 0),
			hasIntersection: true,
			isProper:        true,
			isCollinear:     false,
			expectedX:       5,
			expectedY:       5,
		},
		{
			name:            "EndpointIntersection",
			p1:              geom.NewCoordinate(0, 0),
			p2:              geom.NewCoordinate(10, 0),
			p3:              geom.NewCoordinate(10, 0),
			p4:              geom.NewCoordinate(10, 10),
			hasIntersection: true,
			isProper:        false,
			isCollinear:     false,
			expectedX:       10,
			expectedY:       0,
		},
		{
			name:            "NoIntersection",
			p1:              geom.NewCoordinate(0, 0),
			p2:              geom.NewCoordinate(5, 0),
			p3:              geom.NewCoordinate(10, 0),
			p4:              geom.NewCoordinate(15, 0),
			hasIntersection: false,
			isProper:        false,
			isCollinear:     true, // Collinear but no overlap
		},
		{
			name:            "ParallelNoIntersection",
			p1:              geom.NewCoordinate(0, 0),
			p2:              geom.NewCoordinate(10, 0),
			p3:              geom.NewCoordinate(0, 5),
			p4:              geom.NewCoordinate(10, 5),
			hasIntersection: false,
			isProper:        false,
			isCollinear:     false,
		},
		{
			name:            "CollinearOverlapping",
			p1:              geom.NewCoordinate(0, 0),
			p2:              geom.NewCoordinate(10, 0),
			p3:              geom.NewCoordinate(5, 0),
			p4:              geom.NewCoordinate(15, 0),
			hasIntersection: true,
			isProper:        false,
			isCollinear:     true,
			expectedX:       5,
			expectedY:       0,
		},
		{
			name:            "CollinearNoOverlap",
			p1:              geom.NewCoordinate(0, 0),
			p2:              geom.NewCoordinate(5, 0),
			p3:              geom.NewCoordinate(10, 0),
			p4:              geom.NewCoordinate(15, 0),
			hasIntersection: false,
			isProper:        false,
			isCollinear:     true,
		},
		{
			name:            "CollinearSamePoint",
			p1:              geom.NewCoordinate(5, 5),
			p2:              geom.NewCoordinate(10, 10),
			p3:              geom.NewCoordinate(5, 5),
			p4:              geom.NewCoordinate(15, 15),
			hasIntersection: true,
			isProper:        false,
			isCollinear:     true,
			expectedX:       5,
			expectedY:       5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.LineIntersection(tt.p1, tt.p2, tt.p3, tt.p4)
			assert.Equal(t, tt.hasIntersection, result.HasIntersection, "HasIntersection")
			assert.Equal(t, tt.isProper, result.IsProper, "IsProper")
			assert.Equal(t, tt.isCollinear, result.IsCollinear, "IsCollinear")
			if tt.hasIntersection && !tt.isCollinear {
				assert.InDelta(t, tt.expectedX, result.Intersection.X, 0.001, "Intersection X")
				assert.InDelta(t, tt.expectedY, result.Intersection.Y, 0.001, "Intersection Y")
			}
		})
	}
}

func TestLineIntersectionCollinear(t *testing.T) {
	// Test collinear segments with overlap
	t.Run("CollinearOverlap", func(t *testing.T) {
		p1 := geom.NewCoordinate(0, 0)
		p2 := geom.NewCoordinate(10, 0)
		p3 := geom.NewCoordinate(5, 0)
		p4 := geom.NewCoordinate(15, 0)

		result := algorithm.LineIntersection(p1, p2, p3, p4)
		assert.True(t, result.HasIntersection && result.IsCollinear, "Expected collinear intersection")
	})

	// Test collinear degenerate case (both segments are points)
	t.Run("BothPoints", func(t *testing.T) {
		p1 := geom.NewCoordinate(5, 5)
		p2 := geom.NewCoordinate(5, 5)
		p3 := geom.NewCoordinate(5, 5)
		p4 := geom.NewCoordinate(5, 5)

		result := algorithm.LineIntersection(p1, p2, p3, p4)
		assert.True(t, result.HasIntersection, "Expected intersection for identical points")
	})

	// Test collinear with vertical line
	t.Run("VerticalCollinear", func(t *testing.T) {
		p1 := geom.NewCoordinate(5, 0)
		p2 := geom.NewCoordinate(5, 10)
		p3 := geom.NewCoordinate(5, 5)
		p4 := geom.NewCoordinate(5, 15)

		result := algorithm.LineIntersection(p1, p2, p3, p4)
		assert.True(t, result.HasIntersection && result.IsCollinear, "Expected collinear intersection for vertical segments")
	})
}

func TestLineLineIntersection(t *testing.T) {
	tests := []struct {
		name      string
		p1, p2    geom.Coordinate
		p3, p4    geom.Coordinate
		parallel  bool
		expectedX float64
		expectedY float64
	}{
		{
			name:      "Intersecting",
			p1:        geom.NewCoordinate(0, 0),
			p2:        geom.NewCoordinate(10, 10),
			p3:        geom.NewCoordinate(0, 10),
			p4:        geom.NewCoordinate(10, 0),
			parallel:  false,
			expectedX: 5,
			expectedY: 5,
		},
		{
			name:     "Parallel",
			p1:       geom.NewCoordinate(0, 0),
			p2:       geom.NewCoordinate(10, 0),
			p3:       geom.NewCoordinate(0, 5),
			p4:       geom.NewCoordinate(10, 5),
			parallel: true,
		},
		{
			name:      "IntersectingOutsideSegments",
			p1:        geom.NewCoordinate(0, 0),
			p2:        geom.NewCoordinate(1, 1),
			p3:        geom.NewCoordinate(0, 10),
			p4:        geom.NewCoordinate(10, 0),
			parallel:  false,
			expectedX: 5,
			expectedY: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, hasIntersection := algorithm.LineLineIntersection(tt.p1, tt.p2, tt.p3, tt.p4)
			if tt.parallel {
				assert.False(t, hasIntersection, "Expected no intersection for parallel lines")
			} else {
				assert.True(t, hasIntersection, "Expected intersection")
				assert.InDelta(t, tt.expectedX, result.X, 0.001, "Expected X %v", tt.expectedX)
				assert.InDelta(t, tt.expectedY, result.Y, 0.001, "Expected Y %v", tt.expectedY)
			}
		})
	}
}

func TestRaySegmentIntersection(t *testing.T) {
	tests := []struct {
		name            string
		origin          geom.Coordinate
		dir             geom.Coordinate
		segStart        geom.Coordinate
		segEnd          geom.Coordinate
		hasIntersection bool
		expectedX       float64
		expectedY       float64
	}{
		{
			name:            "Intersecting",
			origin:          geom.NewCoordinate(0, 0),
			dir:             geom.NewCoordinate(1, 1),
			segStart:        geom.NewCoordinate(0, 10),
			segEnd:          geom.NewCoordinate(10, 0),
			hasIntersection: true,
			expectedX:       5,
			expectedY:       5,
		},
		{
			name:            "NoIntersection",
			origin:          geom.NewCoordinate(0, 0),
			dir:             geom.NewCoordinate(1, 0),
			segStart:        geom.NewCoordinate(0, 5),
			segEnd:          geom.NewCoordinate(10, 5),
			hasIntersection: false,
		},
		{
			name:            "ParallelRay",
			origin:          geom.NewCoordinate(0, 0),
			dir:             geom.NewCoordinate(1, 0),
			segStart:        geom.NewCoordinate(5, 0),
			segEnd:          geom.NewCoordinate(10, 0),
			hasIntersection: false,
		},
		{
			name:            "RayBehindOrigin",
			origin:          geom.NewCoordinate(10, 10),
			dir:             geom.NewCoordinate(1, 0),
			segStart:        geom.NewCoordinate(0, 0),
			segEnd:          geom.NewCoordinate(5, 0),
			hasIntersection: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, hasIntersection := algorithm.RaySegmentIntersection(tt.origin, tt.dir, tt.segStart, tt.segEnd)
			assert.Equal(t, tt.hasIntersection, hasIntersection, "HasIntersection")
			if tt.hasIntersection {
				assert.InDelta(t, tt.expectedX, result.X, 0.001, "Expected X %v", tt.expectedX)
				assert.InDelta(t, tt.expectedY, result.Y, 0.001, "Expected Y %v", tt.expectedY)
			}
		})
	}
}

func TestProjectPointOntoLine(t *testing.T) {
	tests := []struct {
		name      string
		p         geom.Coordinate
		lineStart geom.Coordinate
		lineEnd   geom.Coordinate
		expectedX float64
		expectedY float64
	}{
		{
			name:      "ProjectOntoHorizontalLine",
			p:         geom.NewCoordinate(5, 5),
			lineStart: geom.NewCoordinate(0, 0),
			lineEnd:   geom.NewCoordinate(10, 0),
			expectedX: 5,
			expectedY: 0,
		},
		{
			name:      "ProjectOntoVerticalLine",
			p:         geom.NewCoordinate(5, 5),
			lineStart: geom.NewCoordinate(0, 0),
			lineEnd:   geom.NewCoordinate(0, 10),
			expectedX: 0,
			expectedY: 5,
		},
		{
			name:      "ProjectOntoDiagonalLine",
			p:         geom.NewCoordinate(0, 10),
			lineStart: geom.NewCoordinate(0, 0),
			lineEnd:   geom.NewCoordinate(10, 10),
			expectedX: 5,
			expectedY: 5,
		},
		{
			name:      "DegenerateLine",
			p:         geom.NewCoordinate(5, 5),
			lineStart: geom.NewCoordinate(3, 3),
			lineEnd:   geom.NewCoordinate(3, 3),
			expectedX: 3,
			expectedY: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.ProjectPointOntoLine(tt.p, tt.lineStart, tt.lineEnd)
			assert.InDelta(t, tt.expectedX, result.X, 0.001, "Expected X %v", tt.expectedX)
			assert.InDelta(t, tt.expectedY, result.Y, 0.001, "Expected Y %v", tt.expectedY)
		})
	}
}

func TestProjectPointOntoSegment(t *testing.T) {
	tests := []struct {
		name      string
		p         geom.Coordinate
		segStart  geom.Coordinate
		segEnd    geom.Coordinate
		expectedX float64
		expectedY float64
	}{
		{
			name:      "ProjectWithinSegment",
			p:         geom.NewCoordinate(5, 5),
			segStart:  geom.NewCoordinate(0, 0),
			segEnd:    geom.NewCoordinate(10, 0),
			expectedX: 5,
			expectedY: 0,
		},
		{
			name:      "ProjectBeforeStart",
			p:         geom.NewCoordinate(-5, 5),
			segStart:  geom.NewCoordinate(0, 0),
			segEnd:    geom.NewCoordinate(10, 0),
			expectedX: 0,
			expectedY: 0,
		},
		{
			name:      "ProjectAfterEnd",
			p:         geom.NewCoordinate(15, 5),
			segStart:  geom.NewCoordinate(0, 0),
			segEnd:    geom.NewCoordinate(10, 0),
			expectedX: 10,
			expectedY: 0,
		},
		{
			name:      "DegenerateSegment",
			p:         geom.NewCoordinate(5, 5),
			segStart:  geom.NewCoordinate(3, 3),
			segEnd:    geom.NewCoordinate(3, 3),
			expectedX: 3,
			expectedY: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.ProjectPointOntoSegment(tt.p, tt.segStart, tt.segEnd)
			assert.InDelta(t, tt.expectedX, result.X, 0.001, "Expected X %v", tt.expectedX)
			assert.InDelta(t, tt.expectedY, result.Y, 0.001, "Expected Y %v", tt.expectedY)
		})
	}
}

func TestReflectPointOverLine(t *testing.T) {
	tests := []struct {
		name      string
		p         geom.Coordinate
		lineStart geom.Coordinate
		lineEnd   geom.Coordinate
		expectedX float64
		expectedY float64
	}{
		{
			name:      "ReflectOverHorizontalLine",
			p:         geom.NewCoordinate(5, 5),
			lineStart: geom.NewCoordinate(0, 0),
			lineEnd:   geom.NewCoordinate(10, 0),
			expectedX: 5,
			expectedY: -5,
		},
		{
			name:      "ReflectOverVerticalLine",
			p:         geom.NewCoordinate(5, 5),
			lineStart: geom.NewCoordinate(0, 0),
			lineEnd:   geom.NewCoordinate(0, 10),
			expectedX: -5,
			expectedY: 5,
		},
		{
			name:      "ReflectOverDiagonalLine",
			p:         geom.NewCoordinate(0, 10),
			lineStart: geom.NewCoordinate(0, 0),
			lineEnd:   geom.NewCoordinate(10, 10),
			expectedX: 10,
			expectedY: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.ReflectPointOverLine(tt.p, tt.lineStart, tt.lineEnd)
			assert.InDelta(t, tt.expectedX, result.X, 0.001, "Expected X %v", tt.expectedX)
			assert.InDelta(t, tt.expectedY, result.Y, 0.001, "Expected Y %v", tt.expectedY)
		})
	}
}
