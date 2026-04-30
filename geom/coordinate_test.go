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

func TestXYEqualOrBothNaN(t *testing.T) {
	nan := math.NaN()
	a := XY{nan, 2}
	b := XY{nan, 2}
	assert.True(t, a.EqualOrBothNaN(b), "EqualOrBothNaN should treat matching NaN as equal")
	assert.False(t, a.Equal(b), "plain Equal should not treat NaN==NaN")
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
