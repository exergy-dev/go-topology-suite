package algorithm_test

import (
	"math"
	"testing"

	"github.com/robert-malhotra/go-topology-suite/algorithm"
	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
)

func TestSignedArea(t *testing.T) {
	tests := []struct {
		name     string
		ring     geom.CoordinateSequence
		positive bool
	}{
		{
			name:     "CounterClockwise",
			ring:     geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0),
			positive: true,
		},
		{
			name:     "Clockwise",
			ring:     geom.NewCoordinateSequenceXY(0, 0, 0, 10, 10, 10, 10, 0, 0, 0),
			positive: false,
		},
		{
			name:     "TwoPoints",
			ring:     geom.NewCoordinateSequenceXY(0, 0, 10, 0),
			positive: false, // Zero area
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.SignedArea(tt.ring)
			if tt.positive {
				assert.Greater(t, result, 0.0, "Expected positive area")
			} else {
				assert.LessOrEqual(t, result, 0.0, "Expected non-positive area")
			}
		})
	}

	// Test non-closed ring
	t.Run("NonClosedRing", func(t *testing.T) {
		ring := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 10)
		area := algorithm.SignedArea(ring)
		// Should still compute area by closing implicitly
		assert.Greater(t, math.Abs(area), 1.0, "Expected non-zero area for non-closed ring")
	})
}

func TestAngleBetween(t *testing.T) {
	tests := []struct {
		name      string
		p0, p1, p2 geom.Coordinate
		expected   float64
	}{
		{
			name:     "RightAngle",
			p0:       geom.NewCoordinate(0, 0),
			p1:       geom.NewCoordinate(1, 0),
			p2:       geom.NewCoordinate(1, 1),
			expected: math.Pi / 2,
		},
		{
			name:     "StraightAngle",
			p0:       geom.NewCoordinate(0, 0),
			p1:       geom.NewCoordinate(1, 0),
			p2:       geom.NewCoordinate(2, 0),
			expected: math.Pi,
		},
		{
			name:     "AcuteAngle",
			p0:       geom.NewCoordinate(0, 0),
			p1:       geom.NewCoordinate(1, 0),
			p2:       geom.NewCoordinate(1, 1),
			expected: math.Pi / 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.AngleBetween(tt.p0, tt.p1, tt.p2)
			assert.InDelta(t, tt.expected, result, 0.1, "Expected %v", tt.expected)
		})
	}
}

func TestAngleBetweenOriented(t *testing.T) {
	tests := []struct {
		name       string
		p0, p1, p2 geom.Coordinate
		positive   bool
	}{
		{
			name:     "PositiveAngle",
			p0:       geom.NewCoordinate(1, 0),
			p1:       geom.NewCoordinate(0, 0),
			p2:       geom.NewCoordinate(0, 1),
			positive: true,
		},
		{
			name:     "NegativeAngle",
			p0:       geom.NewCoordinate(0, 1),
			p1:       geom.NewCoordinate(0, 0),
			p2:       geom.NewCoordinate(1, 0),
			positive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.AngleBetweenOriented(tt.p0, tt.p1, tt.p2)
			if tt.positive {
				assert.Greater(t, result, 0.0, "Expected positive angle")
			} else {
				assert.Less(t, result, 0.0, "Expected negative angle")
			}
		})
	}
}

func TestNormalizeAngle(t *testing.T) {
	tests := []struct {
		name     string
		angle    float64
		inRange  bool
	}{
		{
			name:    "AlreadyNormalized",
			angle:   math.Pi / 2,
			inRange: true,
		},
		{
			name:    "TooLarge",
			angle:   3 * math.Pi,
			inRange: true,
		},
		{
			name:    "TooSmall",
			angle:   -3 * math.Pi,
			inRange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.NormalizeAngle(tt.angle)
			assert.True(t, result > -math.Pi && result <= math.Pi, "Angle %v not in range (-Pi, Pi]", result)
		})
	}
}

func TestNormalizePositiveAngle(t *testing.T) {
	tests := []struct {
		name  string
		angle float64
	}{
		{
			name:  "AlreadyNormalized",
			angle: math.Pi / 2,
		},
		{
			name:  "TooLarge",
			angle: 3 * math.Pi,
		},
		{
			name:  "Negative",
			angle: -math.Pi / 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.NormalizePositiveAngle(tt.angle)
			assert.True(t, result >= 0 && result < 2*math.Pi, "Angle %v not in range [0, 2*Pi)", result)
		})
	}
}

func TestIsAcute(t *testing.T) {
	tests := []struct {
		name       string
		p0, p1, p2 geom.Coordinate
		expected   bool
	}{
		{
			name:     "Acute",
			p0:       geom.NewCoordinate(1, 0),
			p1:       geom.NewCoordinate(0, 0),
			p2:       geom.NewCoordinate(1, 1),
			expected: true,
		},
		{
			name:     "RightAngle",
			p0:       geom.NewCoordinate(1, 0),
			p1:       geom.NewCoordinate(0, 0),
			p2:       geom.NewCoordinate(0, 1),
			expected: false,
		},
		{
			name:     "Obtuse",
			p0:       geom.NewCoordinate(1, 0),
			p1:       geom.NewCoordinate(0, 0),
			p2:       geom.NewCoordinate(-1, 1),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.IsAcute(tt.p0, tt.p1, tt.p2)
			assert.Equal(t, tt.expected, result, "Expected %v", tt.expected)
		})
	}
}

func TestIsObtuse(t *testing.T) {
	tests := []struct {
		name       string
		p0, p1, p2 geom.Coordinate
		expected   bool
	}{
		{
			name:     "Obtuse",
			p0:       geom.NewCoordinate(1, 0),
			p1:       geom.NewCoordinate(0, 0),
			p2:       geom.NewCoordinate(-1, 1),
			expected: true,
		},
		{
			name:     "RightAngle",
			p0:       geom.NewCoordinate(1, 0),
			p1:       geom.NewCoordinate(0, 0),
			p2:       geom.NewCoordinate(0, 1),
			expected: false,
		},
		{
			name:     "Acute",
			p0:       geom.NewCoordinate(1, 0),
			p1:       geom.NewCoordinate(0, 0),
			p2:       geom.NewCoordinate(1, 1),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.IsObtuse(tt.p0, tt.p1, tt.p2)
			assert.Equal(t, tt.expected, result, "Expected %v", tt.expected)
		})
	}
}

func TestIsRight(t *testing.T) {
	tests := []struct {
		name       string
		p0, p1, p2 geom.Coordinate
		expected   bool
	}{
		{
			name:     "RightAngle",
			p0:       geom.NewCoordinate(1, 0),
			p1:       geom.NewCoordinate(0, 0),
			p2:       geom.NewCoordinate(0, 1),
			expected: true,
		},
		{
			name:     "Acute",
			p0:       geom.NewCoordinate(1, 0),
			p1:       geom.NewCoordinate(0, 0),
			p2:       geom.NewCoordinate(1, 1),
			expected: false,
		},
		{
			name:     "Obtuse",
			p0:       geom.NewCoordinate(1, 0),
			p1:       geom.NewCoordinate(0, 0),
			p2:       geom.NewCoordinate(-1, 1),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.IsRight(tt.p0, tt.p1, tt.p2)
			assert.Equal(t, tt.expected, result, "Expected %v", tt.expected)
		})
	}
}

func TestBisector(t *testing.T) {
	tests := []struct {
		name       string
		p0, p1, p2 geom.Coordinate
	}{
		{
			name: "RightAngle",
			p0:   geom.NewCoordinate(1, 0),
			p1:   geom.NewCoordinate(0, 0),
			p2:   geom.NewCoordinate(0, 1),
		},
		{
			name: "AcuteAngle",
			p0:   geom.NewCoordinate(1, 0),
			p1:   geom.NewCoordinate(0, 0),
			p2:   geom.NewCoordinate(1, 1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.Bisector(tt.p0, tt.p1, tt.p2)
			// Bisector should be in range (-Pi, Pi]
			assert.True(t, result > -math.Pi && result <= math.Pi, "Bisector angle %v not in range (-Pi, Pi]", result)
		})
	}
}
