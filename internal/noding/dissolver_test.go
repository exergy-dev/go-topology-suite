package noding

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/terra-geo/terra/geom"
)

func TestDissolveSegmentStrings_DropsExactDuplicates(t *testing.T) {
	a := &SegmentString{Coords: []geom.XY{{0, 0}, {1, 1}}}
	b := &SegmentString{Coords: []geom.XY{{0, 0}, {1, 1}}}
	out := DissolveSegmentStrings([]*SegmentString{a, b})
	assert.Len(t, out, 1)
}

func TestDissolveSegmentStrings_DropsReverseEquals(t *testing.T) {
	a := &SegmentString{Coords: []geom.XY{{0, 0}, {1, 0}, {1, 1}}}
	b := &SegmentString{Coords: []geom.XY{{1, 1}, {1, 0}, {0, 0}}}
	out := DissolveSegmentStrings([]*SegmentString{a, b})
	assert.Len(t, out, 1)
}

func TestDissolveSegmentStrings_KeepsDistinct(t *testing.T) {
	a := &SegmentString{Coords: []geom.XY{{0, 0}, {1, 1}}}
	b := &SegmentString{Coords: []geom.XY{{0, 0}, {2, 2}}}
	out := DissolveSegmentStrings([]*SegmentString{a, b})
	assert.Len(t, out, 2)
}

func TestDissolveSegmentStrings_PreservesFirstTag(t *testing.T) {
	a := &SegmentString{Coords: []geom.XY{{0, 0}, {1, 1}}, Tag: 7}
	b := &SegmentString{Coords: []geom.XY{{1, 1}, {0, 0}}, Tag: 99}
	out := DissolveSegmentStrings([]*SegmentString{a, b})
	assert.Len(t, out, 1)
	assert.Equal(t, 7, out[0].Tag)
}

func TestDissolveSegmentStrings_EmptyInput(t *testing.T) {
	out := DissolveSegmentStrings(nil)
	assert.Nil(t, out)
}
