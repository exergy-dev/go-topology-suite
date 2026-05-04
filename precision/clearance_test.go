package precision

import (
	"math"
	"testing"

	"github.com/terra-geo/terra/geom"
)

func approxEqual(a, b, tol float64) bool {
	if math.IsInf(a, +1) && math.IsInf(b, +1) {
		return true
	}
	return math.Abs(a-b) <= tol
}

func TestMinimumClearance_Square(t *testing.T) {
	// Unit square: closest pair are adjacent vertices distance 1; no
	// vertex is closer to a non-adjacent edge.
	poly := geom.NewPolygon(nil, []geom.XY{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0}})
	d, _ := MinimumClearance(poly)
	if !approxEqual(d, 1.0, 1e-12) {
		t.Errorf("square: got %v want 1.0", d)
	}
}

func TestMinimumClearance_Rectangle(t *testing.T) {
	// 4x10 rectangle: minimum vertex-vertex distance is 4 (the short
	// side); no closer vertex-edge pair exists.
	poly := geom.NewPolygon(nil, []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 4}, {X: 0, Y: 4}, {X: 0, Y: 0}})
	d, _ := MinimumClearance(poly)
	if !approxEqual(d, 4.0, 1e-12) {
		t.Errorf("rectangle: got %v want 4.0", d)
	}
}

// JTS reference test: a near-collinear vertex creates a small clearance
// equal to the perpendicular distance to the opposite edge.
func TestMinimumClearance_NearCollinear(t *testing.T) {
	// Triangle (0,0)-(10,0)-(5, 0.001): the apex sits 0.001 above the
	// base segment; clearance == 0.001.
	poly := geom.NewPolygon(nil, []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 0.001}, {X: 0, Y: 0}})
	d, _ := MinimumClearance(poly)
	if !approxEqual(d, 0.001, 1e-9) {
		t.Errorf("near-collinear: got %v want 0.001", d)
	}
}

func TestMinimumClearance_LineString(t *testing.T) {
	// Two-vertex segment of length 5: only one vertex pair, distance 5.
	ls := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 5, Y: 0}})
	d, _ := MinimumClearance(ls)
	if !approxEqual(d, 5.0, 1e-12) {
		t.Errorf("2-pt linestring: got %v want 5.0", d)
	}
}

func TestMinimumClearance_SinglePoint(t *testing.T) {
	// Single point — no pair, no segment, clearance is +Inf.
	pt := geom.NewPoint(nil, geom.XY{X: 1, Y: 2})
	d, _ := MinimumClearance(pt)
	if !math.IsInf(d, +1) {
		t.Errorf("single point: got %v want +Inf", d)
	}
}

func TestMinimumClearance_DuplicateMultiPoint(t *testing.T) {
	// MultiPoint with two identical members: vertex pair has distance 0,
	// which is filtered out, so no clearance is found.
	mp := geom.NewMultiPoint(nil, []geom.XY{{X: 1, Y: 1}, {X: 1, Y: 1}})
	d, _ := MinimumClearance(mp)
	if !math.IsInf(d, +1) {
		t.Errorf("duplicate multipoint: got %v want +Inf", d)
	}
}

func TestMinimumClearance_Empty(t *testing.T) {
	empty := geom.NewEmptyPolygon(nil, geom.LayoutXY)
	d, _ := MinimumClearance(empty)
	if !math.IsInf(d, +1) {
		t.Errorf("empty: got %v want +Inf", d)
	}
}

func TestMinimumClearance_Witness(t *testing.T) {
	poly := geom.NewPolygon(nil, []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 5, Y: 0.001}, {X: 0, Y: 0}})
	d, seg := MinimumClearance(poly)
	if !approxEqual(d, 0.001, 1e-9) {
		t.Fatalf("got %v", d)
	}
	// One endpoint should be the apex (5, 0.001); the other should lie on
	// the base segment somewhere.
	apex := geom.XY{X: 5, Y: 0.001}
	if seg[0] != apex && seg[1] != apex {
		t.Errorf("witness must include apex %v; got %v", apex, seg)
	}
	// Both witness coordinates lie within the polygon envelope.
	for _, p := range seg {
		if p.X < 0 || p.X > 10 || p.Y < 0 || p.Y > 0.001 {
			t.Errorf("witness point %v outside expected envelope", p)
		}
	}
}

func TestSimpleMinimumClearance_LineApi(t *testing.T) {
	poly := geom.NewPolygon(nil, []geom.XY{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: 0, Y: 0}})
	smc := NewSimpleMinimumClearance(poly)
	if !approxEqual(smc.Distance(), 1.0, 1e-12) {
		t.Errorf("Distance: got %v", smc.Distance())
	}
	line := smc.Line()
	if line.NumPoints() != 2 {
		t.Errorf("Line: expected 2 points, got %d", line.NumPoints())
	}
}

func TestSimpleMinimumClearance_LineApiEmpty(t *testing.T) {
	empty := geom.NewEmptyPolygon(nil, geom.LayoutXY)
	smc := NewSimpleMinimumClearance(empty)
	line := smc.Line()
	if !line.IsEmpty() {
		t.Errorf("empty input: expected empty line, got %d points", line.NumPoints())
	}
}
