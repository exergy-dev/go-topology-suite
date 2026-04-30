package geom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyEnvelopeIsEmpty(t *testing.T) {
	e := EmptyEnvelope()
	assert.True(t, e.IsEmpty(), "EmptyEnvelope().IsEmpty() should be true")
	assert.Equal(t, 0.0, e.Width(), "empty envelope width")
	assert.Equal(t, 0.0, e.Height(), "empty envelope height")
	assert.Equal(t, 0.0, e.Area(), "empty envelope area")
}

func TestEnvelopeExpandToIncludeXY(t *testing.T) {
	e := EmptyEnvelope()
	e = e.ExpandToIncludeXY(XY{1, 2})
	assert.False(t, e.IsEmpty(), "envelope still empty after first expand")
	assert.Equal(t, 1.0, e.MinX, "MinX after first expand")
	assert.Equal(t, 2.0, e.MinY, "MinY after first expand")
	assert.Equal(t, 1.0, e.MaxX, "MaxX after first expand")
	assert.Equal(t, 2.0, e.MaxY, "MaxY after first expand")
	e = e.ExpandToIncludeXY(XY{-1, 5})
	assert.Equal(t, -1.0, e.MinX, "MinX after second expand")
	assert.Equal(t, 2.0, e.MinY, "MinY after second expand")
	assert.Equal(t, 1.0, e.MaxX, "MaxX after second expand")
	assert.Equal(t, 5.0, e.MaxY, "MaxY after second expand")
}

func TestEnvelopeIntersects(t *testing.T) {
	a := Envelope{0, 0, 10, 10}
	b := Envelope{5, 5, 15, 15}
	c := Envelope{20, 20, 30, 30}

	assert.True(t, a.Intersects(b), "a should intersect b")
	assert.False(t, a.Intersects(c), "a should not intersect c")
	assert.False(t, a.Intersects(EmptyEnvelope()), "nothing should intersect empty")
}

func TestEnvelopeContainsXY(t *testing.T) {
	e := Envelope{0, 0, 10, 10}
	cases := []struct {
		p    XY
		want bool
	}{
		{XY{5, 5}, true},
		{XY{0, 0}, true},   // boundary
		{XY{10, 10}, true}, // boundary
		{XY{-1, 5}, false},
		{XY{5, 11}, false},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, e.ContainsXY(c.p), "Contains(%v)", c.p)
	}
}

func TestEnvelopeOfFlat(t *testing.T) {
	flat := []float64{1, 2, 3, 4, -1, 5}
	e := envelopeOfFlat(flat, 2)
	assert.Equal(t, -1.0, e.MinX, "envelopeOfFlat MinX")
	assert.Equal(t, 2.0, e.MinY, "envelopeOfFlat MinY")
	assert.Equal(t, 3.0, e.MaxX, "envelopeOfFlat MaxX")
	assert.Equal(t, 5.0, e.MaxY, "envelopeOfFlat MaxY")

	flat3 := []float64{1, 2, 99, 3, 4, 99, -1, 5, 99}
	e3 := envelopeOfFlat(flat3, 3)
	assert.Equal(t, -1.0, e3.MinX, "envelopeOfFlat XYZ MinX")
	assert.Equal(t, 2.0, e3.MinY, "envelopeOfFlat XYZ MinY")
	assert.Equal(t, 3.0, e3.MaxX, "envelopeOfFlat XYZ MaxX")
	assert.Equal(t, 5.0, e3.MaxY, "envelopeOfFlat XYZ MaxY")
}
