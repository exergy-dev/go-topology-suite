package relateng

import (
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
)

func TestRelateNode_AddLineEdges_CCWOrder(t *testing.T) {
	pt := geom.XY{X: 0, Y: 0}
	east := geom.XY{X: 1, Y: 0}
	north := geom.XY{X: 0, Y: 1}
	west := geom.XY{X: -1, Y: 0}
	south := geom.XY{X: 0, Y: -1}
	n := NewRelateNode(pt)
	// Add in scrambled order; CCW order is east, north, west, south.
	n.addLineEdge(true, west)
	n.addLineEdge(true, east)
	n.addLineEdge(true, south)
	n.addLineEdge(true, north)
	if len(n.edges) != 4 {
		t.Fatalf("got %d edges, want 4", len(n.edges))
	}
	wantOrder := []geom.XY{east, north, west, south}
	for i, e := range n.edges {
		if e.dirPt != wantOrder[i] {
			t.Errorf("edge[%d] dirPt = %v, want %v", i, e.dirPt, wantOrder[i])
		}
	}
}

func TestRelateNode_DegenerateEdge_Skipped(t *testing.T) {
	pt := geom.XY{X: 1, Y: 1}
	n := NewRelateNode(pt)
	if e := n.addLineEdge(true, pt); e != nil {
		t.Errorf("zero-length edge returned non-nil: %v", e)
	}
	if len(n.edges) != 0 {
		t.Errorf("edges should be empty, got %d", len(n.edges))
	}
}

func TestRelateNode_LineCrossing_FinishPropagates(t *testing.T) {
	// Two crossing lines: A goes east-west, B goes north-south, meeting at origin.
	pt := geom.XY{X: 0, Y: 0}
	east := geom.XY{X: 1, Y: 0}
	west := geom.XY{X: -1, Y: 0}
	north := geom.XY{X: 0, Y: 1}
	south := geom.XY{X: 0, Y: -1}
	n := NewRelateNode(pt)
	n.AddEdgesFromSection(NewNodeSection(true, DimL, 0, 0, nil, false, &west, pt, &east))
	n.AddEdgesFromSection(NewNodeSection(false, DimL, 0, 0, nil, false, &south, pt, &north))
	n.Finish(false, false)
	// All edges should have all positions filled (not locUnknown).
	for i, e := range n.edges {
		for _, pos := range []int{posOn, posLeft, posRight} {
			if e.location(true, pos) == locUnknown {
				t.Errorf("edge[%d] A pos %d is unknown", i, pos)
			}
			if e.location(false, pos) == locUnknown {
				t.Errorf("edge[%d] B pos %d is unknown", i, pos)
			}
		}
	}
}

func TestCompareAngle_Quadrants(t *testing.T) {
	o := geom.XY{X: 0, Y: 0}
	cases := []struct {
		p, q geom.XY
		want int
	}{
		{geom.XY{X: 1, Y: 0}, geom.XY{X: 0, Y: 1}, -1}, // east < north
		{geom.XY{X: 0, Y: 1}, geom.XY{X: -1, Y: 0}, -1},
		{geom.XY{X: 1, Y: 0}, geom.XY{X: 1, Y: 0}, 0},
		{geom.XY{X: 1, Y: 1}, geom.XY{X: 2, Y: 2}, 0}, // collinear
	}
	for _, c := range cases {
		if got := compareAngle(o, c.p, c.q); got != c.want {
			t.Errorf("compareAngle(%v,%v) = %d, want %d", c.p, c.q, got, c.want)
		}
	}
}
