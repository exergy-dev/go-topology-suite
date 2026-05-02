package predicate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestEqualsRingRotation verifies that two polygons with the same
// vertex set described in different cyclic orders are reported equal.
func TestEqualsRingRotation(t *testing.T) {
	a := mustParse(t, "POLYGON ((0 0, 4 0, 4 3, 0 3, 0 0))")
	b := mustParse(t, "POLYGON ((4 3, 0 3, 0 0, 4 0, 4 3))")
	eq, err := Equals(a, b)
	require.NoError(t, err)
	require.True(t, eq, "rotated ring start should be equal")
}

// TestEqualsRingReverse verifies that a polygon and its reversed-ring
// twin (same vertex set, opposite orientation) are reported equal.
// Mirrors the JTS TestSimplify case#13 scenario where Douglas-Peucker
// emits a CCW outer ring when JTS emits CW (or vice versa).
func TestEqualsRingReverse(t *testing.T) {
	a := mustParse(t, "POLYGON ((0 0, 4 0, 4 3, 0 3, 0 0))")
	b := mustParse(t, "POLYGON ((0 0, 0 3, 4 3, 4 0, 0 0))")
	eq, err := Equals(a, b)
	require.NoError(t, err)
	require.True(t, eq, "ring-reversed polygon should be equal")
}

// TestEqualsRingReverseRotated combines rotation + reversal: same
// vertex set, opposite orientation, different starting vertex.
func TestEqualsRingReverseRotated(t *testing.T) {
	a := mustParse(t, "POLYGON ((0 0, 4 0, 4 3, 0 3, 0 0))")
	b := mustParse(t, "POLYGON ((4 3, 4 0, 0 0, 0 3, 4 3))")
	eq, err := Equals(a, b)
	require.NoError(t, err)
	require.True(t, eq, "reversed-and-rotated polygon should be equal")
}

// TestEqualsRingRotationWithHoles verifies hole rings also accept
// rotation/reversal independently of the outer ring.
func TestEqualsRingRotationWithHoles(t *testing.T) {
	a := mustParse(t,
		"POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0), (2 2, 2 4, 4 4, 4 2, 2 2))")
	b := mustParse(t,
		"POLYGON ((10 10, 0 10, 0 0, 10 0, 10 10), (4 4, 4 2, 2 2, 2 4, 4 4))")
	eq, err := Equals(a, b)
	require.NoError(t, err)
	require.True(t, eq, "polygon with rotated hole should be equal")
}

// TestEqualsSimplifyCase13 reproduces the JTS TestSimplify case#13
// signature: same point set, opposite orientation, plus 1 ULP float
// noise in two computed vertices.
func TestEqualsSimplifyCase13(t *testing.T) {
	// CW orientation, exact JTS expected output.
	expected := mustParse(t,
		"POLYGON ((10 10, 10 80, 45.714285714285715 80, 20 20, 80 20, "+
			"54.285714285714285 80, 90 80, 90 10, 10 10))")
	// CCW orientation, 1 ULP off in 54.285…29 (vs JTS 54.285…85).
	got := mustParse(t,
		"POLYGON ((10 80, 10 10, 90 10, 90 80, 54.28571428571429 80, "+
			"80 20, 20 20, 45.714285714285715 80, 10 80))")
	eq, err := Equals(got, expected)
	require.NoError(t, err)
	require.True(t, eq, "simplify case#13: opposite orientation + 1 ULP noise")
}

// TestEqualsRingDifferentVertex verifies that polygons with genuinely
// different vertex sets (not just rotation/reversal) compare unequal.
func TestEqualsRingDifferentVertex(t *testing.T) {
	a := mustParse(t, "POLYGON ((0 0, 4 0, 4 3, 0 3, 0 0))")
	b := mustParse(t, "POLYGON ((0 0, 5 0, 5 3, 0 3, 0 0))")
	eq, err := Equals(a, b)
	require.NoError(t, err)
	require.False(t, eq, "different vertex set should not be equal")
}

// TestEqualsMultiPolygonReorderedMembers verifies that multi-polygons
// whose member polygons are stored in different orders compare equal.
func TestEqualsMultiPolygonReorderedMembers(t *testing.T) {
	a := mustParse(t,
		"MULTIPOLYGON (((0 0, 1 0, 1 1, 0 1, 0 0)), "+
			"((10 10, 11 10, 11 11, 10 11, 10 10)))")
	b := mustParse(t,
		"MULTIPOLYGON (((10 10, 11 10, 11 11, 10 11, 10 10)), "+
			"((0 0, 1 0, 1 1, 0 1, 0 0)))")
	eq, err := Equals(a, b)
	require.NoError(t, err)
	require.True(t, eq, "multipolygon with reordered members should be equal")
}
