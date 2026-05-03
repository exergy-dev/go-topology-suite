package geom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapCollection_AppliesFunctionToEachMember(t *testing.T) {
	a := NewPoint(nil, XY{1, 2})
	b := NewPoint(nil, XY{3, 4})
	gc := NewGeometryCollection(nil, a, b)

	shifted := MapCollection(gc, func(g Geometry) Geometry {
		p := g.(*Point)
		xy := p.XY()
		return NewPoint(nil, XY{xy.X + 10, xy.Y + 100})
	})

	require := assert.New(t)
	require.Equal(2, shifted.NumGeometries())
	require.Equal(XY{11, 102}, shifted.GeometryAt(0).(*Point).XY())
	require.Equal(XY{13, 104}, shifted.GeometryAt(1).(*Point).XY())
}

func TestMapCollection_DropsEmptyResults(t *testing.T) {
	a := NewPoint(nil, XY{1, 2})
	b := NewPoint(nil, XY{3, 4})
	gc := NewGeometryCollection(nil, a, b)

	// Drop b by mapping it to an empty Point.
	out := MapCollection(gc, func(g Geometry) Geometry {
		p := g.(*Point)
		if p.XY() == (XY{3, 4}) {
			return NewEmptyPoint(nil, LayoutXY)
		}
		return p
	})

	assert.Equal(t, 1, out.NumGeometries())
}

func TestMapCollection_NilInput(t *testing.T) {
	assert.Nil(t, MapCollection(nil, func(g Geometry) Geometry { return g }))
}

func TestMapCollection_EmptyCollection(t *testing.T) {
	gc := NewGeometryCollection(nil)
	out := MapCollection(gc, func(g Geometry) Geometry { return g })
	assert.Equal(t, 0, out.NumGeometries())
}

func TestMapCollection_DropsNilResults(t *testing.T) {
	a := NewPoint(nil, XY{1, 2})
	b := NewPoint(nil, XY{3, 4})
	gc := NewGeometryCollection(nil, a, b)

	out := MapCollection(gc, func(g Geometry) Geometry { return nil })
	assert.Equal(t, 0, out.NumGeometries())
}
