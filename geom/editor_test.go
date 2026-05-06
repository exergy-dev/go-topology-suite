package geom

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEdit_Point(t *testing.T) {
	p := NewPoint(nil, XY{1, 2})
	out := Edit(p, func(xy XY) XY { return XY{xy.X + 1, xy.Y * 2} })
	got, ok := out.(*Point)
	require.True(t, ok, "expected *Point, got %T", out)
	assert.Equal(t, XY{2, 4}, got.XY())
}

func TestEdit_LineString(t *testing.T) {
	ls := NewLineString(nil, []XY{{0, 0}, {1, 1}, {2, 0}})
	out := Edit(ls, func(xy XY) XY { return XY{xy.X * 10, xy.Y * 10} }).(*LineString)
	require.Equal(t, 3, out.NumPoints())
	assert.Equal(t, XY{10, 10}, out.PointAt(1))
}

func TestEdit_Polygon(t *testing.T) {
	shell := []XY{{0, 0}, {10, 0}, {10, 10}, {0, 10}, {0, 0}}
	hole := []XY{{2, 2}, {4, 2}, {4, 4}, {2, 4}, {2, 2}}
	p := NewPolygon(nil, shell, hole)
	out := Edit(p, func(xy XY) XY { return XY{xy.X + 100, xy.Y + 100} }).(*Polygon)
	require.Equal(t, 2, out.NumRings())
	assert.Equal(t, XY{100, 100}, out.Ring(0)[0])
	assert.Equal(t, XY{102, 102}, out.Ring(1)[0])
}

func TestEdit_GeometryCollection(t *testing.T) {
	gc := NewGeometryCollection(nil,
		NewPoint(nil, XY{1, 1}),
		NewLineString(nil, []XY{{0, 0}, {2, 2}}))
	out := Edit(gc, func(xy XY) XY { return XY{-xy.X, -xy.Y} }).(*GeometryCollection)
	require.Equal(t, 2, out.NumGeometries())
	pt := out.GeometryAt(0).(*Point)
	assert.Equal(t, XY{-1, -1}, pt.XY())
}
