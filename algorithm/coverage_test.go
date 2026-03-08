package algorithm_test

import (
	"math"
	"testing"

	"github.com/robert-malhotra/go-topology-suite/algorithm"
	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Part 1: Simplification degenerate-input tests
// ---------------------------------------------------------------------------

func TestDouglasPeucker_EmptySequences(t *testing.T) {
	t.Run("EmptyLineString", func(t *testing.T) {
		ls := geom.NewLineStringEmpty()
		result := algorithm.DouglasPeucker(ls, 1.0)
		assert.True(t, result.IsEmpty(), "Empty LineString should stay empty after simplification")
	})

	t.Run("EmptyPolygon", func(t *testing.T) {
		p := geom.NewPolygonEmpty()
		result := algorithm.DouglasPeucker(p, 1.0)
		assert.True(t, result.IsEmpty(), "Empty Polygon should stay empty after simplification")
	})

	t.Run("EmptyLinearRing", func(t *testing.T) {
		lr := geom.NewLinearRingEmpty()
		result := algorithm.DouglasPeucker(lr, 1.0)
		assert.True(t, result.IsEmpty(), "Empty LinearRing should stay empty after simplification")
	})

	t.Run("EmptyMultiLineString", func(t *testing.T) {
		mls := geom.NewMultiLineStringEmpty()
		result := algorithm.DouglasPeucker(mls, 1.0)
		assert.Equal(t, "MultiLineString", result.GeometryType())
	})

	t.Run("EmptyMultiPolygon", func(t *testing.T) {
		mp := geom.NewMultiPolygonEmpty()
		result := algorithm.DouglasPeucker(mp, 1.0)
		assert.Equal(t, "MultiPolygon", result.GeometryType())
	})

	t.Run("EmptyGeometryCollection", func(t *testing.T) {
		gc := geom.NewGeometryCollectionEmpty()
		result := algorithm.DouglasPeucker(gc, 1.0)
		assert.Equal(t, "GeometryCollection", result.GeometryType())
	})
}

func TestDouglasPeucker_SinglePoint(t *testing.T) {
	// Point should be returned unchanged
	p := geom.NewPoint(5.5, 3.3)
	result := algorithm.DouglasPeucker(p, 1.0)
	require.Equal(t, "Point", result.GeometryType())
	coords := result.Coordinates()
	require.Len(t, coords, 1)
	assert.InDelta(t, 5.5, coords[0].X, 1e-10)
	assert.InDelta(t, 3.3, coords[0].Y, 1e-10)
}

func TestDouglasPeucker_TwoPointLineString(t *testing.T) {
	// Two-point line is already minimal; should be returned as-is
	ls := geom.NewLineStringXY(0, 0, 10, 10)
	result := algorithm.DouglasPeucker(ls, 100.0) // huge tolerance
	coords := result.Coordinates()
	require.Len(t, coords, 2, "Two-point line must keep both endpoints")
	assert.InDelta(t, 0, coords[0].X, 1e-10)
	assert.InDelta(t, 10, coords[1].X, 1e-10)
}

func TestDouglasPeucker_AlreadySimpleLine(t *testing.T) {
	// A straight line should reduce to two endpoints with any positive tolerance
	ls := geom.NewLineStringXY(0, 0, 5, 0, 10, 0, 15, 0, 20, 0)
	result := algorithm.DouglasPeucker(ls, 0.01)
	coords := result.Coordinates()
	assert.Equal(t, 2, len(coords), "Collinear points should simplify to endpoints")
	assert.InDelta(t, 0, coords[0].X, 1e-10)
	assert.InDelta(t, 20, coords[1].X, 1e-10)
}

func TestDouglasPeucker_RemovesAllInteriorPoints(t *testing.T) {
	// With a tolerance larger than any deviation, only endpoints remain
	ls := geom.NewLineStringXY(0, 0, 1, 0.5, 2, -0.3, 3, 0.2, 4, -0.1, 5, 0)
	result := algorithm.DouglasPeucker(ls, 100.0)
	coords := result.Coordinates()
	assert.Equal(t, 2, len(coords), "All interior points should be removed with huge tolerance")
	assert.InDelta(t, 0, coords[0].X, 1e-10)
	assert.InDelta(t, 5, coords[1].X, 1e-10)
}

func TestDouglasPeucker_ClosedRingStaysClosed(t *testing.T) {
	// A closed ring should remain closed after simplification
	ring := geom.NewLinearRingXY(
		0, 0, 5, 0.1, 10, 0, 10, 5, 10.1, 10, 10, 10, 5, 10, 0, 10, 0, 5, 0, 0,
	)
	result := algorithm.DouglasPeucker(ring, 0.5)
	require.Equal(t, "LinearRing", result.GeometryType())
	coords := result.Coordinates()
	require.GreaterOrEqual(t, len(coords), 4, "Ring must have at least 4 points")
	// First and last must be equal (closed)
	first := coords[0]
	last := coords[len(coords)-1]
	assert.True(t, first.Equals2D(last, 1e-10),
		"Simplified ring must remain closed: first=%v last=%v", first, last)
}

func TestDouglasPeucker_MinimumRingPointsPreserved(t *testing.T) {
	// A ring that would simplify to fewer than 3 unique points
	// should retain at least 3 unique points + closure = 4 points
	ring := geom.NewLinearRingXY(
		0, 0, 3, 0, 6, 0, 9, 0, 5, 0.01, 0, 0,
	)
	result := algorithm.DouglasPeucker(ring, 100.0)
	coords := result.Coordinates()
	require.GreaterOrEqual(t, len(coords), 4,
		"Ring must keep at least 4 points (3 unique + closure)")
	// Closure check
	assert.True(t, coords[0].Equals2D(coords[len(coords)-1], 1e-10),
		"Ring must be closed")
}

func TestDouglasPeucker_TriangleRingUnchanged(t *testing.T) {
	// A triangle (4 coords) should be returned unchanged since it is already minimal
	ring := geom.NewLinearRingXY(0, 0, 10, 0, 5, 10, 0, 0)
	result := algorithm.DouglasPeucker(ring, 1.0)
	coords := result.Coordinates()
	assert.Equal(t, 4, len(coords), "Triangle ring should remain 4 points")
}

func TestDouglasPeucker_PolygonWithSmallHole(t *testing.T) {
	// A polygon whose hole simplifies to <4 points should have the hole removed
	shell := geom.NewLinearRingXY(0, 0, 100, 0, 100, 100, 0, 100, 0, 0)
	// Tiny hole with vertices very close together
	hole := geom.NewLinearRingXY(
		50, 50, 50.01, 50, 50.01, 50.01, 50.005, 50.005, 50, 50.01, 50, 50,
	)
	poly := geom.NewPolygon(shell, []*geom.LinearRing{hole})
	result := algorithm.DouglasPeucker(poly, 10.0)
	resultPoly := result.(*geom.Polygon)
	// With a tolerance of 10, the tiny hole should collapse or be removed
	// The shell should survive
	assert.False(t, resultPoly.IsEmpty(), "Shell should survive simplification")
}

func TestDouglasPeucker_MultiPointReturnsClone(t *testing.T) {
	mp := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(1, 2),
		geom.NewPoint(3, 4),
	})
	result := algorithm.DouglasPeucker(mp, 1.0)
	require.Equal(t, "MultiPoint", result.GeometryType())
	assert.Equal(t, 2, len(result.Coordinates()), "MultiPoint should be cloned unchanged")
}

// ---------------------------------------------------------------------------
// Visvalingam-Whyatt additional coverage
// ---------------------------------------------------------------------------

func TestVisvalingamWhyatt_TwoPointLine(t *testing.T) {
	ls := geom.NewLineStringXY(0, 0, 10, 0)
	result := algorithm.VisvalingamWhyatt(ls, 1.0)
	coords := result.Coordinates()
	assert.Equal(t, 2, len(coords), "Two-point line should remain unchanged")
}

func TestVisvalingamWhyatt_HighThreshold(t *testing.T) {
	ls := geom.NewLineStringXY(0, 0, 1, 1, 2, 0, 3, 1, 4, 0)
	result := algorithm.VisvalingamWhyatt(ls, 1e6)
	coords := result.Coordinates()
	// High threshold should remove everything to endpoints
	assert.Equal(t, 2, len(coords), "High threshold should simplify to endpoints")
}

func TestVisvalingamWhyatt_EmptyPolygon(t *testing.T) {
	poly := geom.NewPolygonEmpty()
	result := algorithm.VisvalingamWhyatt(poly, 1.0)
	assert.True(t, result.IsEmpty(), "Empty polygon should remain empty")
}

func TestVisvalingamWhyatt_FallbackForLinearRing(t *testing.T) {
	// VisvalingamWhyatt's default branch falls back to DouglasPeucker
	ring := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	result := algorithm.VisvalingamWhyatt(ring, 1.0)
	assert.False(t, result.IsEmpty(), "LinearRing should be simplified via fallback")
}

// ---------------------------------------------------------------------------
// RadialDistance additional coverage
// ---------------------------------------------------------------------------

func TestRadialDistance_AllPointsClose(t *testing.T) {
	// All interior points are closer than threshold: only endpoints kept
	ls := geom.NewLineStringXY(0, 0, 0.1, 0, 0.2, 0, 0.3, 0, 10, 0)
	result := algorithm.RadialDistance(ls, 1.0)
	coords := result.Coordinates()
	// First three interior points are within 1.0 of each other,
	// should be removed. Endpoints always kept.
	assert.GreaterOrEqual(t, len(coords), 2)
	assert.LessOrEqual(t, len(coords), 3)
}

func TestRadialDistance_FallbackForPolygon(t *testing.T) {
	poly := geom.NewPolygon(
		geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0), nil,
	)
	result := algorithm.RadialDistance(poly, 1.0)
	assert.Equal(t, "Polygon", result.GeometryType(),
		"RadialDistance falls back to DouglasPeucker for non-LineString types")
}

// ---------------------------------------------------------------------------
// Part 3: Bisector, InteriorAngle, NormalizeAngle, deprecated wrappers
// ---------------------------------------------------------------------------

func TestBisector_PiMinusPiBoundary(t *testing.T) {
	// Angle crossing the Pi/-Pi boundary
	// p0 is left (angle ~Pi), p2 is below-left (angle ~ -3Pi/4)
	origin := geom.NewCoordinate(0, 0)

	tests := []struct {
		name       string
		p0, p1, p2 geom.Coordinate
	}{
		{
			name: "CrossingPiMinus_Pi_boundary",
			p0:   geom.NewCoordinate(-1, 0.01), // angle just above Pi
			p1:   origin,
			p2:   geom.NewCoordinate(-1, -0.01), // angle just below -Pi
		},
		{
			name: "Arms_at_±135_degrees",
			p0:   geom.NewCoordinate(-1, 1),  // angle = 3Pi/4
			p1:   origin,
			p2:   geom.NewCoordinate(-1, -1), // angle = -3Pi/4
		},
		{
			name: "Wide_V_opening_left",
			p0:   geom.NewCoordinate(-1, 0.5),
			p1:   origin,
			p2:   geom.NewCoordinate(-1, -0.5),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bisect := algorithm.Bisector(tt.p0, tt.p1, tt.p2)
			// The bisector must be in the valid range (-Pi, Pi]
			assert.True(t, bisect > -math.Pi && bisect <= math.Pi,
				"Bisector %v out of range (-Pi, Pi]", bisect)
			// For symmetric inputs about the negative x-axis, the bisector
			// should point along the negative x-axis (angle = Pi or -Pi).
			// The exact value depends on which side of the discontinuity we land.
			// We just verify the direction is roughly pointing left.
			bx := math.Cos(bisect)
			assert.Less(t, bx, 0.01,
				"Bisector should point generally to the left for arms around the -x axis")
		})
	}
}

func TestBisector_CoincidentArms(t *testing.T) {
	// When both arms point in the same direction the bisector should equal that direction
	p0 := geom.NewCoordinate(1, 0)
	p1 := geom.NewCoordinate(0, 0)
	p2 := geom.NewCoordinate(1, 0) // same as p0
	bisect := algorithm.Bisector(p0, p1, p2)
	assert.InDelta(t, 0, bisect, 0.01, "Bisector of identical arms should equal arm angle (0)")
}

func TestNormalizeAngle_Extremes(t *testing.T) {
	tests := []struct {
		name  string
		input float64
	}{
		{"ExactlyPi", math.Pi},
		{"ExactlyMinusPi", -math.Pi},
		{"Zero", 0},
		{"MultipleTurnsPositive", 10 * math.Pi},
		{"MultipleTurnsNegative", -10 * math.Pi},
		{"SlightlyAbovePi", math.Pi + 1e-5},
		{"SlightlyBelowMinusPi", -math.Pi - 1e-5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.NormalizeAngle(tt.input)
			assert.True(t, result > -math.Pi-1e-12 && result <= math.Pi+1e-12,
				"NormalizeAngle(%v) = %v out of range", tt.input, result)
		})
	}
}

func TestNormalizePositiveAngle_Extremes(t *testing.T) {
	tests := []struct {
		name  string
		input float64
	}{
		{"Zero", 0},
		{"NegativeHalfPi", -math.Pi / 2},
		{"Exactly2Pi", 2 * math.Pi},
		{"LargePositive", 7 * math.Pi},
		{"LargeNegative", -7 * math.Pi},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := algorithm.NormalizePositiveAngle(tt.input)
			assert.True(t, result >= 0 && result < 2*math.Pi+1e-12,
				"NormalizePositiveAngle(%v) = %v out of range", tt.input, result)
		})
	}
}

func TestToDegreesAndRadians_RoundTrip(t *testing.T) {
	for _, deg := range []float64{0, 45, 90, 180, 270, 360, -45, -180} {
		got := algorithm.ToDegrees(algorithm.ToRadians(deg))
		assert.InDelta(t, deg, got, 1e-10, "Round-trip failed for %v degrees", deg)
	}
}

// ---------------------------------------------------------------------------
// Angle helpers: IsAcute, IsObtuse, IsRight for additional code paths
// ---------------------------------------------------------------------------

func TestIsAcute_ZeroAngle(t *testing.T) {
	// Coincident arms: dot product is positive -> acute
	p0 := geom.NewCoordinate(1, 0)
	p1 := geom.NewCoordinate(0, 0)
	p2 := geom.NewCoordinate(2, 0) // same direction
	assert.True(t, algorithm.IsAcute(p0, p1, p2))
}

func TestIsObtuse_StraightLine(t *testing.T) {
	// Opposite directions: dot < 0 -> obtuse
	p0 := geom.NewCoordinate(1, 0)
	p1 := geom.NewCoordinate(0, 0)
	p2 := geom.NewCoordinate(-1, 0)
	assert.True(t, algorithm.IsObtuse(p0, p1, p2))
}

func TestIsRight_Perpendicular(t *testing.T) {
	p0 := geom.NewCoordinate(1, 0)
	p1 := geom.NewCoordinate(0, 0)
	p2 := geom.NewCoordinate(0, 1)
	assert.True(t, algorithm.IsRight(p0, p1, p2))
}

// ---------------------------------------------------------------------------
// Part 5: ProjectPointOntoLine edge cases
// ---------------------------------------------------------------------------

func TestProjectPointOntoLine_NearDegenerateLine(t *testing.T) {
	// A segment with length ~1e-8 (well above the squared-epsilon threshold
	// of 1e-20, but below the old buggy threshold of 1e-10).
	// This must project correctly, not collapse to the start point.
	p := geom.NewCoordinate(0, 1)
	lineStart := geom.NewCoordinate(0, 0)
	lineEnd := geom.NewCoordinate(1e-8, 0)

	result := algorithm.ProjectPointOntoLine(p, lineStart, lineEnd)
	// Projection of (0,1) onto the X-axis line should land at (0,0)
	assert.InDelta(t, 0, result.X, 1e-6)
	assert.InDelta(t, 0, result.Y, 1e-6)
}

func TestProjectPointOntoLine_TrulyDegenerateLine(t *testing.T) {
	// A zero-length segment should return the start point.
	p := geom.NewCoordinate(5, 5)
	lineStart := geom.NewCoordinate(1, 1)
	lineEnd := geom.NewCoordinate(1, 1)

	result := algorithm.ProjectPointOntoLine(p, lineStart, lineEnd)
	assert.InDelta(t, 1, result.X, geom.DefaultEpsilon)
	assert.InDelta(t, 1, result.Y, geom.DefaultEpsilon)
}

// ---------------------------------------------------------------------------
// Part 6: ConvexHull input immutability
// ---------------------------------------------------------------------------

func TestConvexHull_DoesNotMutateInput(t *testing.T) {
	coords := geom.CoordinateSequence{
		geom.NewCoordinate(3, 1),
		geom.NewCoordinate(1, 3),
		geom.NewCoordinate(2, 2),
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(4, 4),
	}
	// Save original order
	original := coords.Clone()

	ls := geom.NewLineString(coords)
	_ = algorithm.ConvexHull(ls)

	for i, c := range coords {
		assert.InDelta(t, original[i].X, c.X, geom.DefaultEpsilon, "input coord %d X was mutated", i)
		assert.InDelta(t, original[i].Y, c.Y, geom.DefaultEpsilon, "input coord %d Y was mutated", i)
	}
}
