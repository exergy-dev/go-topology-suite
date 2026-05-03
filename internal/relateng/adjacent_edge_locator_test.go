package relateng

import (
	"testing"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/wkt"
)

func TestAdjacentEdgeLocator_SharedBoundary_Interior(t *testing.T) {
	// Two adjacent polygons sharing the segment x=10. A point on the
	// shared edge should be reported as INTERIOR of the GC union.
	g, err := wkt.Unmarshal("GEOMETRYCOLLECTION(POLYGON((0 0,10 0,10 10,0 10,0 0)),POLYGON((10 0,20 0,20 10,10 10,10 0)))")
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	loc := NewAdjacentEdgeLocator(g)
	if got := loc.Locate(geom.XY{X: 10, Y: 5}); got != LocInterior {
		t.Errorf("Locate on shared edge = %d, want LocInterior(%d)", got, LocInterior)
	}
}

func TestAdjacentEdgeLocator_OuterBoundary_Boundary(t *testing.T) {
	g, err := wkt.Unmarshal("GEOMETRYCOLLECTION(POLYGON((0 0,10 0,10 10,0 10,0 0)),POLYGON((10 0,20 0,20 10,10 10,10 0)))")
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	loc := NewAdjacentEdgeLocator(g)
	// A point on the outer boundary of the union (top edge of the left polygon).
	if got := loc.Locate(geom.XY{X: 5, Y: 10}); got != LocBoundary {
		t.Errorf("Locate on outer edge = %d, want LocBoundary(%d)", got, LocBoundary)
	}
}
