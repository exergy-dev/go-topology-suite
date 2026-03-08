package geom_test

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
)

func TestLinearRing_IsValid_SelfIntersecting(t *testing.T) {
	// Figure-8 shape: ring crosses itself
	// Coordinates form an X pattern that self-intersects
	lr := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	})
	assert.False(t, lr.IsValid(), "Self-intersecting ring (figure-8) should be invalid")
}

func TestPolygon_IsValid_ShellSelfIntersecting(t *testing.T) {
	// Bowtie polygon: shell crosses itself (figure-8 shape)
	shell := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	})
	poly := geom.NewPolygon(shell, nil)
	assert.False(t, poly.IsValid(), "Bowtie polygon (self-intersecting shell) should be invalid")
}

func TestPolygon_IsValid_HoleOutsideShell(t *testing.T) {
	// Shell is a 10x10 square at origin
	shell := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	})
	// Hole is completely outside the shell (at 20,20)
	hole := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(20, 20),
		geom.NewCoordinate(20, 30),
		geom.NewCoordinate(30, 30),
		geom.NewCoordinate(30, 20),
		geom.NewCoordinate(20, 20),
	})
	poly := geom.NewPolygon(shell, []*geom.LinearRing{hole})
	assert.False(t, poly.IsValid(), "Polygon with hole outside shell should be invalid")
}

func TestPolygon_IsValid_ShellHoleCrossing(t *testing.T) {
	// Shell is a 10x10 square at origin
	shell := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	})
	// Hole extends outside the shell (crosses the shell boundary)
	// Part inside: (2,2)-(8,2)-(8,8)-(2,8)
	// But it extends beyond the shell at (15,5)
	hole := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(2, 2),
		geom.NewCoordinate(2, 8),
		geom.NewCoordinate(15, 8), // Outside shell
		geom.NewCoordinate(15, 2), // Outside shell
		geom.NewCoordinate(2, 2),
	})
	poly := geom.NewPolygon(shell, []*geom.LinearRing{hole})
	assert.False(t, poly.IsValid(), "Polygon with hole crossing shell should be invalid")
}

func TestPolygon_IsValid_HolesNested(t *testing.T) {
	// Large shell
	shell := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(30, 0),
		geom.NewCoordinate(30, 30),
		geom.NewCoordinate(0, 30),
		geom.NewCoordinate(0, 0),
	})
	// Outer hole
	hole1 := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(5, 5),
		geom.NewCoordinate(5, 25),
		geom.NewCoordinate(25, 25),
		geom.NewCoordinate(25, 5),
		geom.NewCoordinate(5, 5),
	})
	// Inner hole (nested inside hole1)
	hole2 := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(10, 20),
		geom.NewCoordinate(20, 20),
		geom.NewCoordinate(20, 10),
		geom.NewCoordinate(10, 10),
	})
	poly := geom.NewPolygon(shell, []*geom.LinearRing{hole1, hole2})
	assert.False(t, poly.IsValid(), "Polygon with nested holes should be invalid")
}

func TestPolygon_IsValid_HolesCrossing(t *testing.T) {
	// Large shell
	shell := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(30, 0),
		geom.NewCoordinate(30, 30),
		geom.NewCoordinate(0, 30),
		geom.NewCoordinate(0, 0),
	})
	// Hole 1: horizontal rectangle
	hole1 := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(5, 10),
		geom.NewCoordinate(5, 20),
		geom.NewCoordinate(25, 20),
		geom.NewCoordinate(25, 10),
		geom.NewCoordinate(5, 10),
	})
	// Hole 2: vertical rectangle that crosses hole1
	hole2 := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(10, 5),
		geom.NewCoordinate(10, 25),
		geom.NewCoordinate(20, 25),
		geom.NewCoordinate(20, 5),
		geom.NewCoordinate(10, 5),
	})
	poly := geom.NewPolygon(shell, []*geom.LinearRing{hole1, hole2})
	assert.False(t, poly.IsValid(), "Polygon with crossing holes should be invalid")
}

func TestPolygon_IsValid_ValidPolygonWithHole(t *testing.T) {
	// Counter-clockwise shell (20x20 square)
	shell := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(20, 0),
		geom.NewCoordinate(20, 20),
		geom.NewCoordinate(0, 20),
		geom.NewCoordinate(0, 0),
	})
	// Clockwise hole inside shell (10x10 centered)
	hole := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(5, 5),
		geom.NewCoordinate(5, 15),
		geom.NewCoordinate(15, 15),
		geom.NewCoordinate(15, 5),
		geom.NewCoordinate(5, 5),
	})
	poly := geom.NewPolygon(shell, []*geom.LinearRing{hole})
	assert.True(t, poly.IsValid(), "Valid polygon with proper hole should be valid")
}

func TestPolygon_IsValid_EmptyPolygon(t *testing.T) {
	poly := geom.NewPolygonEmpty()
	assert.True(t, poly.IsValid(), "Empty polygon should be valid")
}

func TestPolygon_IsValid_SimpleValidPolygon(t *testing.T) {
	// Simple counter-clockwise square
	shell := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	})
	poly := geom.NewPolygon(shell, nil)
	assert.True(t, poly.IsValid(), "Simple valid polygon should be valid")
}

func TestPolygon_IsValid_CWShellIsValid(t *testing.T) {
	// Clockwise shell - orientation should not affect validity
	shell := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(0, 0),
	})
	poly := geom.NewPolygon(shell, nil)
	assert.True(t, poly.IsValid(), "Polygon with clockwise shell should be valid (orientation-agnostic)")
}

func TestPolygon_IsValid_CCWHoleIsValid(t *testing.T) {
	// Counter-clockwise shell
	shell := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(20, 0),
		geom.NewCoordinate(20, 20),
		geom.NewCoordinate(0, 20),
		geom.NewCoordinate(0, 0),
	})
	// Counter-clockwise hole - orientation should not affect validity
	hole := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(5, 5),
		geom.NewCoordinate(15, 5),
		geom.NewCoordinate(15, 15),
		geom.NewCoordinate(5, 15),
		geom.NewCoordinate(5, 5),
	})
	poly := geom.NewPolygon(shell, []*geom.LinearRing{hole})
	assert.True(t, poly.IsValid(), "Polygon with counter-clockwise hole should be valid (orientation-agnostic)")
}

func TestPolygon_Normalize_EnforcesOrientation(t *testing.T) {
	// Clockwise shell
	shell := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(0, 0),
	})
	assert.True(t, shell.IsCW(), "Shell should start as CW")

	// CCW hole
	hole := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(2, 2),
		geom.NewCoordinate(8, 2),
		geom.NewCoordinate(8, 8),
		geom.NewCoordinate(2, 8),
		geom.NewCoordinate(2, 2),
	})
	assert.True(t, hole.IsCCW(), "Hole should start as CCW")

	poly := geom.NewPolygon(shell, []*geom.LinearRing{hole})
	normalized := poly.Normalized().(*geom.Polygon)

	// After Normalized, shell should be CCW and hole should be CW
	assert.True(t, normalized.ExteriorRing().IsCCW(), "After Normalized, shell should be CCW")
	assert.True(t, normalized.InteriorRingN(0).IsCW(), "After Normalized, hole should be CW")
}

func TestLinearRing_IsValid_Empty(t *testing.T) {
	lr := geom.NewLinearRingEmpty()
	assert.True(t, lr.IsValid(), "Empty ring should be valid")
}

func TestLinearRing_IsValid_ValidRing(t *testing.T) {
	lr := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	})
	assert.True(t, lr.IsValid(), "Valid ring should be valid")
}

func TestPolygon_IsValid_MultipleValidHoles(t *testing.T) {
	// Counter-clockwise shell (30x30 square)
	shell := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(30, 0),
		geom.NewCoordinate(30, 30),
		geom.NewCoordinate(0, 30),
		geom.NewCoordinate(0, 0),
	})
	// Two separate clockwise holes
	hole1 := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(2, 2),
		geom.NewCoordinate(2, 8),
		geom.NewCoordinate(8, 8),
		geom.NewCoordinate(8, 2),
		geom.NewCoordinate(2, 2),
	})
	hole2 := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(22, 22),
		geom.NewCoordinate(22, 28),
		geom.NewCoordinate(28, 28),
		geom.NewCoordinate(28, 22),
		geom.NewCoordinate(22, 22),
	})
	poly := geom.NewPolygon(shell, []*geom.LinearRing{hole1, hole2})
	assert.True(t, poly.IsValid(), "Polygon with multiple separate holes should be valid")
}
