package dissolve

import (
	"testing"

	"github.com/terra-geo/terra/geom"
)

func ls(pts ...geom.XY) *geom.LineString {
	return geom.NewLineString(nil, pts)
}

func TestDissolve_NoSharedSegments(t *testing.T) {
	a := ls(geom.XY{X: 0, Y: 0}, geom.XY{X: 1, Y: 0})
	b := ls(geom.XY{X: 2, Y: 0}, geom.XY{X: 3, Y: 0})
	out := LineDissolver([]geom.Geometry{a, b})
	if len(out) != 2 {
		t.Fatalf("got %d, want 2 (disjoint inputs)", len(out))
	}
}

func TestDissolve_SharedSegmentDeduped(t *testing.T) {
	// Two identical lines — should collapse to one.
	a := ls(geom.XY{X: 0, Y: 0}, geom.XY{X: 1, Y: 0})
	b := ls(geom.XY{X: 0, Y: 0}, geom.XY{X: 1, Y: 0})
	out := LineDissolver([]geom.Geometry{a, b})
	if len(out) != 1 {
		t.Fatalf("got %d, want 1", len(out))
	}
	if out[0].NumPoints() != 2 {
		t.Errorf("npoints %d", out[0].NumPoints())
	}
}

func TestDissolve_ChainMergedThroughDegree2(t *testing.T) {
	// (0,0)→(1,0)→(2,0) — interior vertex degree 2, should merge.
	a := ls(geom.XY{X: 0, Y: 0}, geom.XY{X: 1, Y: 0})
	b := ls(geom.XY{X: 1, Y: 0}, geom.XY{X: 2, Y: 0})
	out := LineDissolver([]geom.Geometry{a, b})
	if len(out) != 1 {
		t.Fatalf("got %d, want 1 (merged chain)", len(out))
	}
	if out[0].NumPoints() != 3 {
		t.Errorf("expected 3 points after merge, got %d", out[0].NumPoints())
	}
}

func TestDissolve_BranchingNotMerged(t *testing.T) {
	// Three lines meeting at (1,0) — degree-3 node, should produce 3 chains.
	a := ls(geom.XY{X: 0, Y: 0}, geom.XY{X: 1, Y: 0})
	b := ls(geom.XY{X: 1, Y: 0}, geom.XY{X: 2, Y: 0})
	c := ls(geom.XY{X: 1, Y: 0}, geom.XY{X: 1, Y: 1})
	out := LineDissolver([]geom.Geometry{a, b, c})
	if len(out) != 3 {
		t.Fatalf("got %d, want 3 (T-junction)", len(out))
	}
}

func TestDissolve_IsolatedRing(t *testing.T) {
	// Closed square — every node degree 2, emit single ring.
	a := ls(geom.XY{X: 0, Y: 0}, geom.XY{X: 1, Y: 0}, geom.XY{X: 1, Y: 1}, geom.XY{X: 0, Y: 1}, geom.XY{X: 0, Y: 0})
	out := LineDissolver([]geom.Geometry{a})
	if len(out) != 1 {
		t.Fatalf("got %d, want 1 ring", len(out))
	}
	if out[0].NumPoints() != 5 {
		t.Errorf("expected closed ring of 5 points, got %d", out[0].NumPoints())
	}
	if !out[0].IsClosed() {
		t.Errorf("ring should be closed")
	}
}

func TestDissolve_PolygonBoundary(t *testing.T) {
	// Two adjacent unit squares share an edge from (1,0) to (1,1).
	// Each occurrence of the shared edge collapses to a single graph
	// edge; the two endpoints (1,0) and (1,1) are degree-3 nodes
	// (each connects to the other side of both polygons plus the
	// shared edge). Result: three chains — left U-perimeter, shared
	// edge, right U-perimeter. (LineDissolver does NOT remove
	// shared interior edges; a separate operation is needed for that.)
	left := geom.NewPolygon(nil, []geom.XY{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0}})
	right := geom.NewPolygon(nil, []geom.XY{{X: 1, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 1}, {X: 1, Y: 1}, {X: 1, Y: 0}})
	out := LineDissolver([]geom.Geometry{left, right})
	if len(out) != 3 {
		t.Fatalf("got %d chains, want 3", len(out))
	}
}

func TestDissolve_EmptyInput(t *testing.T) {
	if got := LineDissolver(nil); got != nil {
		t.Errorf("nil input should give nil, got %v", got)
	}
}
