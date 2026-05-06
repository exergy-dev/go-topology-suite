package geom

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPointsOf(t *testing.T) {
	gc := NewGeometryCollection(nil,
		NewPoint(nil, XY{1, 1}),
		NewLineString(nil, []XY{{0, 0}, {1, 0}}),
		NewMultiPoint(nil, []XY{{2, 2}, {3, 3}}),
	)
	pts := PointsOf(gc)
	require.Equal(t, 3, len(pts))
}

func TestLineStringsOf(t *testing.T) {
	mls := NewMultiLineString(nil,
		NewLineString(nil, []XY{{0, 0}, {1, 1}}),
		NewLineString(nil, []XY{{2, 2}, {3, 3}}),
	)
	gc := NewGeometryCollection(nil, mls, NewPoint(nil, XY{5, 5}))
	ls := LineStringsOf(gc)
	require.Equal(t, 2, len(ls))
}

func TestPolygonsOf(t *testing.T) {
	p := NewPolygon(nil, []XY{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}})
	mp := NewMultiPolygon(nil, p, p)
	gc := NewGeometryCollection(nil, mp, p)
	got := PolygonsOf(gc)
	require.Equal(t, 3, len(got))
}

func TestExtracters_Empty(t *testing.T) {
	assert.Nil(t, PointsOf(nil), "nil should give nil")
	gc := NewGeometryCollection(nil)
	assert.Equal(t, 0, len(PolygonsOf(gc)), "empty collection should give zero polygons")
}
