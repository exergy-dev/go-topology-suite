package geom

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPointXYMConstructor(t *testing.T) {
	p := NewPointXYM(nil, XYM{X: 1, Y: 2, M: 3.14})
	assert.Equal(t, LayoutXYM, p.Layout(), "layout")
	assert.Equal(t, 3.14, p.M(), "M")
	assert.True(t, math.IsNaN(p.Z()), "Z on XYM should be NaN, got %v", p.Z())
}

func TestPointXYZMConstructor(t *testing.T) {
	p := NewPointXYZM(nil, XYZM{X: 1, Y: 2, Z: 3, M: 4})
	assert.Equal(t, LayoutXYZM, p.Layout(), "layout")
	assert.Equal(t, 3.0, p.Z(), "Z")
	assert.Equal(t, 4.0, p.M(), "M")
}

func TestPointXYZMZAndM(t *testing.T) {
	xy := NewPoint(nil, XY{X: 1, Y: 2})
	assert.True(t, math.IsNaN(xy.Z()), "XY point Z should be NaN")
	assert.True(t, math.IsNaN(xy.M()), "XY point M should be NaN")
}
