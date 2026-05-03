package noding

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/terra-geo/terra/geom"
)

func TestHasAnyIntersection_TwoCrossingLines(t *testing.T) {
	a := &SegmentString{Coords: []geom.XY{{0, 0}, {10, 10}}}
	b := &SegmentString{Coords: []geom.XY{{0, 10}, {10, 0}}}
	assert.True(t, HasAnyIntersection([]*SegmentString{a, b}))
}

func TestHasAnyIntersection_ParallelDisjoint(t *testing.T) {
	a := &SegmentString{Coords: []geom.XY{{0, 0}, {10, 0}}}
	b := &SegmentString{Coords: []geom.XY{{0, 1}, {10, 1}}}
	assert.False(t, HasAnyIntersection([]*SegmentString{a, b}))
}

func TestHasAnyIntersection_TouchingEndpointCountsByDefault(t *testing.T) {
	a := &SegmentString{Coords: []geom.XY{{0, 0}, {5, 5}}}
	b := &SegmentString{Coords: []geom.XY{{5, 5}, {10, 0}}}
	assert.True(t, HasAnyIntersection([]*SegmentString{a, b}))
}

func TestSegmentIntersectionDetector_FindProperIgnoresEndpointTouch(t *testing.T) {
	a := &SegmentString{Coords: []geom.XY{{0, 0}, {5, 5}}}
	b := &SegmentString{Coords: []geom.XY{{5, 5}, {10, 0}}}
	d := NewSegmentIntersectionDetector()
	d.FindProper = true
	d.Detect([]*SegmentString{a, b})
	assert.False(t, d.HasIntersection())
}

func TestSegmentIntersectionDetector_FindProperFlagsCrossing(t *testing.T) {
	a := &SegmentString{Coords: []geom.XY{{0, 0}, {10, 10}}}
	b := &SegmentString{Coords: []geom.XY{{0, 10}, {10, 0}}}
	d := NewSegmentIntersectionDetector()
	d.FindProper = true
	d.Detect([]*SegmentString{a, b})
	assert.True(t, d.HasIntersection())
	assert.True(t, d.HasProperIntersection())
	assert.Equal(t, geom.XY{X: 5, Y: 5}, d.IntersectionPoint())
}

func TestSegmentIntersectionDetector_SingleStringSelfCross(t *testing.T) {
	// Figure-eight: vertex 1->2 crosses vertex 3->0 (closing edge).
	ss := &SegmentString{Coords: []geom.XY{{0, 0}, {10, 10}, {0, 10}, {10, 0}, {0, 0}}}
	d := NewSegmentIntersectionDetector()
	d.Detect([]*SegmentString{ss})
	assert.True(t, d.HasIntersection())
}

func TestSegmentIntersectionDetector_EmptyInput(t *testing.T) {
	d := NewSegmentIntersectionDetector()
	d.Detect(nil)
	assert.False(t, d.HasIntersection())
}
