package simplify

import (
	"testing"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/wkt"
)

func TestSimplifyStraightLine(t *testing.T) {
	// Three collinear points: the middle should drop with any positive tolerance.
	g, _ := wkt.Unmarshal("LINESTRING (0 0, 1 1, 2 2)")
	out := Simplify(g, 0.01)
	if ls := out.(*geom.LineString); ls.NumPoints() != 2 {
		t.Errorf("collinear simplification produced %d points, want 2", ls.NumPoints())
	}
}

func TestSimplifyKeepsBumps(t *testing.T) {
	// Sharp bump at (1, 1) should survive a small tolerance.
	g, _ := wkt.Unmarshal("LINESTRING (0 0, 1 1, 2 0)")
	out := Simplify(g, 0.5)
	if ls := out.(*geom.LineString); ls.NumPoints() != 3 {
		t.Errorf("expected 3 points kept, got %d", ls.NumPoints())
	}
	// At higher tolerance, bump collapses.
	out2 := Simplify(g, 2)
	if ls := out2.(*geom.LineString); ls.NumPoints() != 2 {
		t.Errorf("aggressive tol: expected 2 points, got %d", ls.NumPoints())
	}
}

func TestSimplifyZeroToleranceUnchanged(t *testing.T) {
	g, _ := wkt.Unmarshal("LINESTRING (0 0, 0.001 0, 1 0)")
	out := Simplify(g, 0)
	if ls := out.(*geom.LineString); ls.NumPoints() != 3 {
		t.Errorf("zero tolerance should not change geometry, got %d", ls.NumPoints())
	}
}

func TestSimplifyPolygon(t *testing.T) {
	// Square with extra collinear point on top edge.
	g, _ := wkt.Unmarshal("POLYGON ((0 0, 5 10, 10 10, 10 0, 0 0))")
	out := Simplify(g, 100) // very aggressive
	// Should fall back to original (refuses to over-simplify).
	if out.(*geom.Polygon).NumRings() == 0 {
		t.Errorf("over-simplification produced empty polygon")
	}
}

func TestSimplifyPoint(t *testing.T) {
	g, _ := wkt.Unmarshal("POINT (1 2)")
	out := Simplify(g, 0.5)
	if out.Type() != geom.PointType {
		t.Errorf("simplify of point should be point, got %v", out.Type())
	}
}
