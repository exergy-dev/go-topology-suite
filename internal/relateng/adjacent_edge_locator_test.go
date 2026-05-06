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
