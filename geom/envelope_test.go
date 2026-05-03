package geom

import (
	"math"
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

func TestEnvelopeExpandBy(t *testing.T) {
	e := Envelope{1, 2, 3, 4}
	r := e.ExpandBy(0.5)
	assert.Equal(t, 0.5, r.MinX, "ExpandBy MinX")
	assert.Equal(t, 1.5, r.MinY, "ExpandBy MinY")
	assert.Equal(t, 3.5, r.MaxX, "ExpandBy MaxX")
	assert.Equal(t, 4.5, r.MaxY, "ExpandBy MaxY")

	// Empty envelope stays empty.
	empty := EmptyEnvelope().ExpandBy(1)
	assert.True(t, empty.IsEmpty(), "ExpandBy empty stays empty")

	// Negative shrink that collapses to empty.
	collapsed := Envelope{0, 0, 1, 1}.ExpandBy(-2)
	assert.True(t, collapsed.IsEmpty(), "ExpandBy collapsing to empty")
}

func TestEnvelopeDistance(t *testing.T) {
	a := Envelope{0, 0, 10, 10}
	b := Envelope{20, 20, 30, 30}
	assert.InDelta(t, 10*1.4142135623730951, a.Distance(b), 1e-9, "diagonal distance")

	// Intersecting -> 0.
	c := Envelope{5, 5, 15, 15}
	assert.Equal(t, 0.0, a.Distance(c), "intersecting distance is 0")

	// Disjoint along X only.
	d := Envelope{20, 0, 30, 10}
	assert.InDelta(t, 10.0, a.Distance(d), 1e-9, "x-only gap")

	// Disjoint along Y only.
	e2 := Envelope{0, 20, 10, 30}
	assert.InDelta(t, 10.0, a.Distance(e2), 1e-9, "y-only gap")

	// Empty envelope -> 0.
	assert.Equal(t, 0.0, a.Distance(EmptyEnvelope()), "distance to empty")
}

func TestEnvelopeDisjoint(t *testing.T) {
	a := Envelope{0, 0, 10, 10}
	b := Envelope{5, 5, 15, 15}
	c := Envelope{20, 20, 30, 30}
	assert.False(t, a.Disjoint(b), "overlap not disjoint")
	assert.True(t, a.Disjoint(c), "no overlap is disjoint")
	assert.True(t, a.Disjoint(EmptyEnvelope()), "empty is always disjoint")
}

func TestEnvelopeOverlaps(t *testing.T) {
	a := Envelope{0, 0, 10, 10}
	b := Envelope{5, 5, 15, 15}
	c := Envelope{20, 20, 30, 30}
	assert.True(t, a.Overlaps(b), "overlapping envelopes")
	assert.False(t, a.Overlaps(c), "non-overlapping envelopes")
}

func TestEnvelopeContainsProperly(t *testing.T) {
	outer := Envelope{0, 0, 10, 10}
	inner := Envelope{2, 2, 8, 8}
	touching := Envelope{0, 0, 5, 5}

	assert.True(t, outer.ContainsProperly(inner), "strictly inside")
	assert.False(t, outer.ContainsProperly(touching), "touches boundary")
	assert.False(t, outer.ContainsProperly(outer), "self does not contain properly")
	assert.False(t, outer.ContainsProperly(EmptyEnvelope()), "empty inner not contained properly")
	assert.False(t, EmptyEnvelope().ContainsProperly(inner), "empty outer contains nothing")
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

func TestEnvelopeOfFlatSkipsNaN(t *testing.T) {
	nan := math.NaN()
	// Middle vertex has NaN X — must not poison the envelope.
	flat := []float64{0, 0, nan, 5, 10, 10}
	e := envelopeOfFlat(flat, 2)
	assert.False(t, math.IsNaN(e.MinX), "MinX must not be NaN")
	assert.False(t, math.IsNaN(e.MaxX), "MaxX must not be NaN")
	assert.Equal(t, 0.0, e.MinX, "MinX from non-NaN vertices")
	assert.Equal(t, 0.0, e.MinY, "MinY from non-NaN vertices")
	assert.Equal(t, 10.0, e.MaxX, "MaxX from non-NaN vertices")
	assert.Equal(t, 10.0, e.MaxY, "MaxY from non-NaN vertices")

	// First vertex NaN: the seed should advance to the first finite vertex.
	flat2 := []float64{nan, nan, 1, 2, 3, 4}
	e2 := envelopeOfFlat(flat2, 2)
	assert.Equal(t, 1.0, e2.MinX, "MinX skipping NaN seed")
	assert.Equal(t, 2.0, e2.MinY, "MinY skipping NaN seed")
	assert.Equal(t, 3.0, e2.MaxX, "MaxX skipping NaN seed")
	assert.Equal(t, 4.0, e2.MaxY, "MaxY skipping NaN seed")

	// All-NaN coordinates collapse to the canonical empty envelope.
	allNaN := []float64{nan, nan, nan, nan}
	eAll := envelopeOfFlat(allNaN, 2)
	assert.True(t, eAll.IsEmpty(), "all-NaN flat -> empty envelope")
}
