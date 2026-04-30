package geom

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terra-geo/terra/crs"
)

func TestPointConstruction(t *testing.T) {
	p := NewPoint(crs.WGS84, XY{-75.16, 39.95})
	require.False(t, p.IsEmpty(), "point should not be empty")
	assert.Equal(t, PointType, p.Type(), "Type")
	assert.Equal(t, LayoutXY, p.Layout(), "Layout")
	assert.Equal(t, 1, p.NumGeometries(), "NumGeometries")
	got := p.XY()
	assert.Equal(t, -75.16, got.X, "XY().X")
	assert.Equal(t, 39.95, got.Y, "XY().Y")
	env := p.Envelope()
	assert.Equal(t, -75.16, env.MinX, "envelope MinX")
	assert.Equal(t, -75.16, env.MaxX, "envelope MaxX")
}

func TestEmptyPoint(t *testing.T) {
	p := NewEmptyPoint(crs.WGS84, LayoutXY)
	assert.True(t, p.IsEmpty(), "empty point should be empty")
	assert.True(t, p.Envelope().IsEmpty(), "empty point envelope should be empty")
}

func TestLineStringConstruction(t *testing.T) {
	ls := NewLineString(crs.WGS84, []XY{{0, 0}, {1, 1}, {2, 2}})
	require.Equal(t, 3, ls.NumPoints(), "NumPoints")
	got := ls.PointAt(1)
	assert.Equal(t, 1.0, got.X, "PointAt(1).X")
	assert.Equal(t, 1.0, got.Y, "PointAt(1).Y")
	count := 0
	for p := range ls.CoordsXY() {
		_ = p
		count++
	}
	assert.Equal(t, 3, count, "CoordsXY iterator yield count")
}

func TestPolygonConstruction(t *testing.T) {
	outer := []XY{{0, 0}, {0, 10}, {10, 10}, {10, 0}, {0, 0}}
	hole := []XY{{2, 2}, {2, 4}, {4, 4}, {4, 2}, {2, 2}}
	p := NewPolygon(crs.WGS84, outer, hole)
	require.Equal(t, 2, p.NumRings(), "NumRings")
	got := p.ExteriorRing()
	assert.Len(t, got, 5, "exterior len")
	holes := p.InteriorRings()
	require.Len(t, holes, 1, "holes count")
	assert.Len(t, holes[0], 5, "hole[0] len")
	env := p.Envelope()
	assert.Equal(t, 0.0, env.MinX, "polygon envelope MinX")
	assert.Equal(t, 10.0, env.MaxX, "polygon envelope MaxX")
	assert.Equal(t, 0.0, env.MinY, "polygon envelope MinY")
	assert.Equal(t, 10.0, env.MaxY, "polygon envelope MaxY")
}

func TestMultiLineStringEnvelopeUnion(t *testing.T) {
	a := NewLineString(crs.WGS84, []XY{{0, 0}, {1, 1}})
	b := NewLineString(crs.WGS84, []XY{{5, 5}, {6, 6}})
	m := NewMultiLineString(crs.WGS84, a, b)
	env := m.Envelope()
	assert.Equal(t, 0.0, env.MinX, "union envelope MinX")
	assert.Equal(t, 6.0, env.MaxX, "union envelope MaxX")
	assert.Equal(t, 0.0, env.MinY, "union envelope MinY")
	assert.Equal(t, 6.0, env.MaxY, "union envelope MaxY")
	assert.Equal(t, 2, m.NumGeometries(), "NumGeometries")
}

func TestGeometryInterfaceSatisfaction(t *testing.T) {
	// Compile-time checks that every concrete type implements Geometry.
	var _ Geometry = (*Point)(nil)
	var _ Geometry = (*LineString)(nil)
	var _ Geometry = (*Polygon)(nil)
	var _ Geometry = (*MultiPoint)(nil)
	var _ Geometry = (*MultiLineString)(nil)
	var _ Geometry = (*MultiPolygon)(nil)
	var _ Geometry = (*GeometryCollection)(nil)
}
