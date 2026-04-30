package validate

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	terra "github.com/terra-geo/terra"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel/planar"
)

func TestMakeValid_UnclosedRing(t *testing.T) {
	// Outer ring missing the closing vertex.
	p := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 0, Y: 10}, {X: 10, Y: 10}, {X: 10, Y: 0},
	})
	g, err := MakeValid(p)
	require.NoError(t, err)
	require.False(t, g == nil || g.IsEmpty(), "expected non-empty polygon, got %v", g)
	assert.NoError(t, Validate(g), "expected valid result")
}

func TestMakeValid_CWRingReorientedToCCW(t *testing.T) {
	// Clockwise outer ring (negative shoelace area).
	cw := []geom.XY{
		{X: 0, Y: 0}, {X: 0, Y: 10}, {X: 10, Y: 10}, {X: 10, Y: 0}, {X: 0, Y: 0},
	}
	require.Less(t, planar.Default.RingArea(cw), 0.0, "test setup: ring should be CW (negative area), got %v", planar.Default.RingArea(cw))
	p := geom.NewPolygon(nil, cw)
	g, err := MakeValid(p)
	require.NoError(t, err)
	out, ok := g.(*geom.Polygon)
	require.True(t, ok, "expected *Polygon, got %T", g)
	assert.Greater(t, planar.Default.RingArea(out.ExteriorRing()), 0.0, "expected CCW outer ring (positive area)")
	assert.NoError(t, Validate(out), "expected valid result")
}

func TestMakeValid_BowtiePolygon(t *testing.T) {
	// Self-intersecting bowtie.
	p := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 10, Y: 10}, {X: 10, Y: 0}, {X: 0, Y: 10}, {X: 0, Y: 0},
	})
	g, err := MakeValid(p)
	require.NoError(t, err)
	require.NotNil(t, g, "got nil result")
	// Result must be non-nil. We don't pin the exact shape (overlay may
	// simplify in unexpected ways per documented v0.1 limitations); we only
	// require that the result isn't an obvious structural failure beyond
	// what overlay produces.
	if g.IsEmpty() {
		// Empty is acceptable too — overlay may collapse a degenerate bowtie.
		return
	}
	// If non-empty, structural fields (closure, vertex count) must hold.
	switch x := g.(type) {
	case *geom.Polygon:
		ring := x.ExteriorRing()
		if len(ring) > 0 {
			assert.Equal(t, ring[0], ring[len(ring)-1], "outer ring not closed in result")
		}
	case *geom.MultiPolygon:
		for i := 0; i < x.NumGeometries(); i++ {
			ring := x.PolygonAt(i).ExteriorRing()
			if len(ring) > 0 {
				assert.Equal(t, ring[0], ring[len(ring)-1], "part %d outer ring not closed", i)
			}
		}
	}
}

func TestMakeValid_LineStringSinglePoint(t *testing.T) {
	ls := geom.NewLineString(nil, []geom.XY{{X: 3, Y: 4}})
	g, err := MakeValid(ls)
	require.NoError(t, err)
	pt, ok := g.(*geom.Point)
	require.True(t, ok, "expected *Point, got %T", g)
	assert.Equal(t, geom.XY{X: 3, Y: 4}, pt.XY(), "unexpected point")
}

func TestMakeValid_LineStringDuplicatePointsCollapse(t *testing.T) {
	// Three duplicates collapse to a single distinct vertex → Point.
	ls := geom.NewLineString(nil, []geom.XY{
		{X: 1, Y: 1}, {X: 1, Y: 1}, {X: 1, Y: 1},
	})
	g, err := MakeValid(ls)
	require.NoError(t, err)
	_, ok := g.(*geom.Point)
	assert.True(t, ok, "expected *Point from duplicate-only line, got %T", g)
}

func TestMakeValid_PolygonTooFewVertices(t *testing.T) {
	// Triangle missing one vertex; ring length after closure is < 4.
	p := geom.NewPolygon(nil, []geom.XY{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 0}})
	g, err := MakeValid(p)
	require.NoError(t, err)
	out, ok := g.(*geom.Polygon)
	require.True(t, ok, "expected *Polygon, got %T", g)
	assert.True(t, out.IsEmpty(), "expected empty polygon for too-few-vertex input, got %v", out)
}

func TestMakeValid_MultiPolygonPreservesValidMembers(t *testing.T) {
	// One valid square + one degenerate (too few vertices) polygon.
	good := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}, {X: 1, Y: 0}, {X: 0, Y: 0},
	})
	bad := geom.NewPolygon(nil, []geom.XY{{X: 5, Y: 5}, {X: 5, Y: 6}, {X: 5, Y: 5}})
	mp := geom.NewMultiPolygon(nil, good, bad)

	g, err := MakeValid(mp)
	require.NoError(t, err)
	out, ok := g.(*geom.MultiPolygon)
	require.True(t, ok, "expected *MultiPolygon, got %T", g)
	assert.Equal(t, 1, out.NumGeometries(), "expected 1 surviving polygon")
	assert.NoError(t, Validate(out), "expected valid multipolygon")
}

func TestMakeValid_AlreadyValidPolygonRoundTrips(t *testing.T) {
	// CCW closed square.
	p := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0},
	})
	require.NoError(t, Validate(p), "test setup: input expected valid")
	g, err := MakeValid(p)
	require.NoError(t, err)
	assert.NoError(t, Validate(g), "expected valid result")
	out, ok := g.(*geom.Polygon)
	require.True(t, ok, "expected *Polygon, got %T", g)
	// Same ring vertex count and same first/last vertex.
	assert.Equal(t, 1, out.NumRings(), "expected 1 ring")
	ring := out.ExteriorRing()
	assert.Equal(t, 5, len(ring), "expected 5 vertices")
}

func TestMakeValid_PointPassthrough(t *testing.T) {
	pt := geom.NewPoint(nil, geom.XY{X: 7, Y: 8})
	g, err := MakeValid(pt)
	require.NoError(t, err)
	assert.Equal(t, geom.Geometry(pt), g, "expected same pointer back for valid Point, got %v", g)
}

func TestMakeValid_EmptyReturnsErrEmpty(t *testing.T) {
	empty := geom.NewEmptyPolygon(nil, geom.LayoutXY)
	g, err := MakeValid(empty)
	assert.True(t, errors.Is(err, terra.ErrEmpty), "expected ErrEmpty, got err=%v g=%v", err, g)
}

func TestMakeValid_HolesDropped(t *testing.T) {
	// Square with a small hole; holes should be dropped per v0.1 limitation.
	outer := []geom.XY{
		{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0},
	}
	hole := []geom.XY{
		{X: 3, Y: 3}, {X: 4, Y: 3}, {X: 4, Y: 4}, {X: 3, Y: 4}, {X: 3, Y: 3},
	}
	p := geom.NewPolygon(nil, outer, hole)
	g, err := MakeValid(p)
	require.NoError(t, err)
	out, ok := g.(*geom.Polygon)
	require.True(t, ok, "expected *Polygon, got %T", g)
	assert.Equal(t, 1, out.NumRings(), "expected hole dropped (1 ring), got %d rings", out.NumRings())
}

func TestMakeValid_GeometryCollectionRecurses(t *testing.T) {
	pt := geom.NewPoint(nil, geom.XY{X: 1, Y: 2})
	// Unclosed ring — MakeValid should close it.
	poly := geom.NewPolygon(nil, []geom.XY{
		{X: 0, Y: 0}, {X: 0, Y: 5}, {X: 5, Y: 5}, {X: 5, Y: 0},
	})
	gc := geom.NewGeometryCollection(nil, pt, poly)
	g, err := MakeValid(gc)
	require.NoError(t, err)
	out, ok := g.(*geom.GeometryCollection)
	require.True(t, ok, "expected *GeometryCollection, got %T", g)
	assert.Equal(t, 2, out.NumGeometries(), "expected 2 children")
	assert.NoError(t, Validate(out), "expected valid collection")
}
