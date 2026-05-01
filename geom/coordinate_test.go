package geom

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXYEqual(t *testing.T) {
	a := XY{1, 2}
	b := XY{1, 2}
	c := XY{1, 3}
	assert.True(t, a.Equal(b))
	assert.False(t, a.Equal(c))
}

func TestXYEqual_NaN(t *testing.T) {
	nan := math.NaN()
	a := XY{nan, 2}
	b := XY{nan, 2}
	c := XY{nan, 3}
	// Equal is now NaN-safe: NaN ordinates compare equal to NaN ordinates.
	assert.True(t, a.Equal(b), "Equal should treat matching NaN as equal")
	assert.False(t, a.Equal(c), "Equal should still respect non-NaN differences")
	// EqualOrBothNaN is preserved as a deprecated synonym.
	assert.True(t, a.EqualOrBothNaN(b))
}

func TestXYEqualBitwise(t *testing.T) {
	nan := math.NaN()
	a := XY{nan, 2}
	b := XY{nan, 2}
	// EqualBitwise uses raw IEEE-754 == — NaN never equals NaN.
	assert.False(t, a.EqualBitwise(b), "EqualBitwise should not treat NaN==NaN")
	// Non-NaN bit-equal values still compare equal.
	c := XY{1.5, 2.5}
	d := XY{1.5, 2.5}
	assert.True(t, c.EqualBitwise(d))
}

func TestLayoutStride(t *testing.T) {
	assert.Equal(t, 2, LayoutXY.Stride())
	assert.Equal(t, 3, LayoutXYZ.Stride())
	assert.Equal(t, 3, LayoutXYM.Stride())
	assert.Equal(t, 4, LayoutXYZM.Stride())
	assert.Equal(t, 0, NoLayout.Stride())
}

func TestLayoutHasZHasM(t *testing.T) {
	assert.False(t, LayoutXY.HasZ())
	assert.False(t, LayoutXY.HasM())
	assert.True(t, LayoutXYZ.HasZ())
	assert.False(t, LayoutXYZ.HasM())
	assert.False(t, LayoutXYM.HasZ())
	assert.True(t, LayoutXYM.HasM())
	assert.True(t, LayoutXYZM.HasZ())
	assert.True(t, LayoutXYZM.HasM())
}
