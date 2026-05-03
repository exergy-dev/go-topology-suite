package linemerge

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/wkt"
)

func mustParse(t *testing.T, s string) geom.Geometry {
	t.Helper()
	g, err := wkt.Unmarshal(s)
	require.NoError(t, err)
	return g
}

// Two separate, non-touching lines must remain two outputs.
func TestMerge_SeparateLinesNotMerged(t *testing.T) {
	a := mustParse(t, "LINESTRING (0 0, 1 0)")
	b := mustParse(t, "LINESTRING (2 0, 3 0)")
	out := Merge([]geom.Geometry{a, b})
	require.Len(t, out, 2, "two disjoint lines should stay separate")
}

// Y-junction: three lines meeting at a degree-3 node. The junction
// is not merged — three inputs produce three outputs.
func TestMerge_YJunctionStaysSeparate(t *testing.T) {
	a := mustParse(t, "LINESTRING (0 0, 1 0)")
	b := mustParse(t, "LINESTRING (0 0, -1 0)")
	c := mustParse(t, "LINESTRING (0 0, 0 1)")
	out := Merge([]geom.Geometry{a, b, c})
	require.Len(t, out, 3, "Y-junction at degree-3 node must keep three lines")
}

// Simple chain: three lines meeting end-to-end at degree-2 nodes
// merge into a single polyline.
func TestMerge_SimpleChainCollapses(t *testing.T) {
	a := mustParse(t, "LINESTRING (0 0, 1 0)")
	b := mustParse(t, "LINESTRING (1 0, 2 0)")
	c := mustParse(t, "LINESTRING (2 0, 3 0)")
	out := Merge([]geom.Geometry{a, b, c})
	require.Len(t, out, 1, "chain at degree-2 nodes should merge to one")
	got := out[0]
	assert.Equal(t, 4, got.NumPoints(), "merged polyline should have 4 points")
	assert.Equal(t, geom.XY{X: 0, Y: 0}, got.PointAt(0))
	assert.Equal(t, geom.XY{X: 3, Y: 0}, got.PointAt(got.NumPoints()-1))
}

// Reverse-direction chain: input lines are oriented inconsistently,
// merge should still produce one polyline by traversing the chain.
func TestMerge_ChainWithReversedSegment(t *testing.T) {
	a := mustParse(t, "LINESTRING (0 0, 1 0)")
	b := mustParse(t, "LINESTRING (2 0, 1 0)") // reversed
	c := mustParse(t, "LINESTRING (2 0, 3 0)")
	out := Merge([]geom.Geometry{a, b, c})
	require.Len(t, out, 1)
	got := out[0]
	assert.Equal(t, 4, got.NumPoints())
}

// Empty / trivial inputs are dropped.
func TestMerge_DropsTrivialInputs(t *testing.T) {
	a := mustParse(t, "LINESTRING (0 0, 1 0)")
	b := mustParse(t, "LINESTRING EMPTY")
	c := mustParse(t, "LINESTRING (5 5, 5 5)") // no distinct vertices
	out := Merge([]geom.Geometry{a, b, c})
	require.Len(t, out, 1, "only non-trivial line should survive")
}

// Isolated loop: a chain whose endpoints coincide forms a single
// closed polyline. All nodes are degree-2.
func TestMerge_IsolatedLoop(t *testing.T) {
	a := mustParse(t, "LINESTRING (0 0, 1 0)")
	b := mustParse(t, "LINESTRING (1 0, 1 1)")
	c := mustParse(t, "LINESTRING (1 1, 0 0)")
	out := Merge([]geom.Geometry{a, b, c})
	require.Len(t, out, 1, "loop should merge to one closed polyline")
	got := out[0]
	assert.Equal(t, got.PointAt(0), got.PointAt(got.NumPoints()-1),
		"isolated loop result should be closed")
}

// MultiLineString input is decomposed into its members.
func TestMerge_MultiLineStringDecomposed(t *testing.T) {
	mls := mustParse(t, "MULTILINESTRING ((0 0, 1 0), (1 0, 2 0))")
	out := Merge([]geom.Geometry{mls})
	require.Len(t, out, 1, "two-segment MLS should merge to one")
}
