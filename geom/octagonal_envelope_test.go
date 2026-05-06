package geom

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOctagonalEnvelope_NullByDefault(t *testing.T) {
	oe := NewOctagonalEnvelope()
	require.True(t, oe.IsNull(), "expected null envelope")
	require.False(t, oe.IntersectsXY(XY{X: 0, Y: 0}), "null envelope should not intersect any point")
}

func TestOctagonalEnvelope_ExpandSinglePoint(t *testing.T) {
	oe := NewOctagonalEnvelope()
	oe.ExpandToInclude(3, 4)
	require.False(t, oe.IsNull(), "expected non-null after expand")
	require.True(t, oe.MinX() == 3 && oe.MaxX() == 3 && oe.MinY() == 4 && oe.MaxY() == 4,
		"axis bounds wrong: %v..%v / %v..%v", oe.MinX(), oe.MaxX(), oe.MinY(), oe.MaxY())
	require.True(t, oe.MinA() == 7 && oe.MaxA() == 7 && oe.MinB() == -1 && oe.MaxB() == -1,
		"diag bounds wrong: A=%v..%v B=%v..%v", oe.MinA(), oe.MaxA(), oe.MinB(), oe.MaxB())
}

func TestOctagonalEnvelope_FromGeometryAndIntersects(t *testing.T) {
	g := NewPolygon(nil, []XY{{0, 0}, {4, 0}, {4, 4}, {0, 4}, {0, 0}})
	oe := NewOctagonalEnvelopeFromGeometry(g)
	require.False(t, oe.IsNull(), "expected non-null")
	// Inside the square — must be inside the octagon too.
	for _, p := range []XY{{2, 2}, {0, 0}, {4, 4}, {1, 3}} {
		require.True(t, oe.IntersectsXY(p), "expected octagon to contain %+v", p)
	}
	// Far away should not.
	require.False(t, oe.IntersectsXY(XY{X: -10, Y: -10}), "octagon should not contain far point")
}

func TestOctagonalEnvelope_ContainsAndIntersects(t *testing.T) {
	g := NewPolygon(nil, []XY{{0, 0}, {10, 0}, {10, 10}, {0, 10}, {0, 0}})
	oe := NewOctagonalEnvelopeFromGeometry(g)
	inner := NewOctagonalEnvelope()
	inner.ExpandToInclude(2, 2)
	inner.ExpandToInclude(5, 5)
	require.True(t, oe.Contains(inner), "oe should contain inner")
	require.True(t, oe.Intersects(inner), "oe should intersect inner")
	disjoint := NewOctagonalEnvelope()
	disjoint.ExpandToInclude(100, 100)
	require.False(t, oe.Intersects(disjoint), "disjoint envelopes should not intersect")
}

func TestOctagonalEnvelope_ToGeometry_Polygon(t *testing.T) {
	g := NewPolygon(nil, []XY{{0, 0}, {10, 0}, {10, 10}, {0, 10}, {0, 0}})
	oe := NewOctagonalEnvelopeFromGeometry(g)
	got := oe.ToGeometry(nil)
	poly, ok := got.(*Polygon)
	require.True(t, ok, "expected Polygon, got %T", got)
	require.NotZero(t, poly.NumRings(), "expected non-empty polygon")
	// All input vertices should be contained in the octagon.
	for _, p := range []XY{{0, 0}, {10, 0}, {10, 10}, {0, 10}} {
		require.True(t, oe.IntersectsXY(p), "octagon should contain %+v", p)
	}
}

func TestOctagonalEnvelope_ToGeometry_Point(t *testing.T) {
	oe := NewOctagonalEnvelope()
	oe.ExpandToInclude(5, 5)
	g := oe.ToGeometry(nil)
	_, ok := g.(*Point)
	require.True(t, ok, "expected Point, got %T", g)
}

func TestOctagonalEnvelope_ToGeometry_NullEmpty(t *testing.T) {
	oe := NewOctagonalEnvelope()
	g := oe.ToGeometry(nil)
	require.True(t, g.IsEmpty(), "expected empty geometry for null envelope")
}

func TestOctagonalEnvelope_ExpandBy(t *testing.T) {
	oe := NewOctagonalEnvelope()
	oe.ExpandToInclude(0, 0)
	oe.ExpandBy(1)
	require.False(t, oe.IsNull(), "non-null expected after positive expandBy")
	assert.True(t, oe.IntersectsXY(XY{X: 0.5, Y: 0.5}), "expanded envelope should contain (0.5,0.5)")
}
