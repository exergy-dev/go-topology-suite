package simplify

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/wkt"
)

func TestSimplifyStraightLine(t *testing.T) {
	// Three collinear points: the middle should drop with any positive tolerance.
	g, _ := wkt.Unmarshal("LINESTRING (0 0, 1 1, 2 2)")
	out := Simplify(g, 0.01)
	ls := out.(*geom.LineString)
	assert.Equal(t, 2, ls.NumPoints(), "collinear simplification produced %d points, want 2", ls.NumPoints())
}

func TestSimplifyKeepsBumps(t *testing.T) {
	// Sharp bump at (1, 1) should survive a small tolerance.
	g, _ := wkt.Unmarshal("LINESTRING (0 0, 1 1, 2 0)")
	out := Simplify(g, 0.5)
	ls := out.(*geom.LineString)
	assert.Equal(t, 3, ls.NumPoints(), "expected 3 points kept, got %d", ls.NumPoints())
	// At higher tolerance, bump collapses.
	out2 := Simplify(g, 2)
	ls2 := out2.(*geom.LineString)
	assert.Equal(t, 2, ls2.NumPoints(), "aggressive tol: expected 2 points, got %d", ls2.NumPoints())
}

func TestSimplifyZeroToleranceUnchanged(t *testing.T) {
	g, _ := wkt.Unmarshal("LINESTRING (0 0, 0.001 0, 1 0)")
	out := Simplify(g, 0)
	ls := out.(*geom.LineString)
	assert.Equal(t, 3, ls.NumPoints(), "zero tolerance should not change geometry, got %d", ls.NumPoints())
}

func TestSimplifyPolygon(t *testing.T) {
	// Square with extra collinear point on top edge.
	g, _ := wkt.Unmarshal("POLYGON ((0 0, 5 10, 10 10, 10 0, 0 0))")
	out := Simplify(g, 100) // very aggressive
	assert.True(t, out.(*geom.Polygon).IsEmpty(), "JTS-compatible DP collapse should produce POLYGON EMPTY")
}

func TestSimplifyPoint(t *testing.T) {
	g, _ := wkt.Unmarshal("POINT (1 2)")
	out := Simplify(g, 0.5)
	assert.Equal(t, geom.PointType, out.Type(), "simplify of point should be point, got %v", out.Type())
}
