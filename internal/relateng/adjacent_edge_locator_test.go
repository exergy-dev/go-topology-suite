package relateng

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/wkt"
)

func TestAdjacentEdgeLocator_SharedBoundary_Interior(t *testing.T) {
	// Two adjacent polygons sharing the segment x=10. A point on the
	// shared edge should be reported as INTERIOR of the GC union.
	g, err := wkt.Unmarshal("GEOMETRYCOLLECTION(POLYGON((0 0,10 0,10 10,0 10,0 0)),POLYGON((10 0,20 0,20 10,10 10,10 0)))")
	require.NoError(t, err)
	loc := NewAdjacentEdgeLocator(g)
	assert.Equal(t, LocInterior, loc.Locate(geom.XY{X: 10, Y: 5}),
		"Locate on shared edge should be LocInterior(%d)", LocInterior)
}

func TestAdjacentEdgeLocator_OuterBoundary_Boundary(t *testing.T) {
	g, err := wkt.Unmarshal("GEOMETRYCOLLECTION(POLYGON((0 0,10 0,10 10,0 10,0 0)),POLYGON((10 0,20 0,20 10,10 10,10 0)))")
	require.NoError(t, err)
	loc := NewAdjacentEdgeLocator(g)
	// A point on the outer boundary of the union (top edge of the left polygon).
	assert.Equal(t, LocBoundary, loc.Locate(geom.XY{X: 5, Y: 10}),
		"Locate on outer edge should be LocBoundary(%d)", LocBoundary)
}

func TestAdjacentEdgeLocator_SharedBoundary_InexactProjection(t *testing.T) {
	// JTS TestRelateGC.xml #11/#14: (4 2) is on the shared edge x=4 but
	// its projection parameter on (4 6)-(4 1) is the inexact 0.8. A
	// float-distance == 0 on-segment test missed it when the closest
	// point computation was FMA-fused (arm64), reporting BOUNDARY.
	g, err := wkt.Unmarshal("GEOMETRYCOLLECTION (POLYGON ((1 1, 1 6, 4 6, 4 1, 1 1)), POLYGON ((9 1, 4 1, 4 6, 9 6, 9 1)))")
	require.NoError(t, err)
	loc := NewAdjacentEdgeLocator(g)
	for _, p := range []geom.XY{{X: 4, Y: 2}, {X: 4, Y: 3}, {X: 4, Y: 4}, {X: 4, Y: 5}} {
		assert.Equal(t, LocInterior, loc.Locate(p), "Locate(%v)", p)
	}
}
