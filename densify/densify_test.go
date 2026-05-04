package densify

import (
	"math"
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
)

func segLen(a, b geom.XY) float64 {
	return math.Hypot(b.X-a.X, b.Y-a.Y)
}

func TestDensifyLineString(t *testing.T) {
	ls := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}})
	out := Densify(ls, 3.0).(*geom.LineString)
	if out.NumPoints() < 4 {
		t.Fatalf("expected at least 4 points after densifying, got %d", out.NumPoints())
	}
	for i := 0; i < out.NumPoints()-1; i++ {
		if d := segLen(out.PointAt(i), out.PointAt(i+1)); d > 3.0+1e-9 {
			t.Errorf("segment %d has length %g > tol", i, d)
		}
	}
	// endpoints preserved
	if out.PointAt(0) != (geom.XY{X: 0, Y: 0}) || out.PointAt(out.NumPoints()-1) != (geom.XY{X: 10, Y: 0}) {
		t.Errorf("endpoints not preserved")
	}
}

func TestDensifyShortLineUntouched(t *testing.T) {
	ls := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 1, Y: 0}})
	out := Densify(ls, 5.0).(*geom.LineString)
	if out.NumPoints() != 2 {
		t.Errorf("expected 2 points (no densification), got %d", out.NumPoints())
	}
}

func TestDensifyPoint(t *testing.T) {
	p := geom.NewPoint(nil, geom.XY{X: 5, Y: 5})
	out := Densify(p, 1.0)
	if out != p {
		t.Errorf("Point should be returned as-is")
	}
}

func TestDensifyPolygon(t *testing.T) {
	ring := []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0}}
	p := geom.NewPolygon(nil, ring)
	out := Densify(p, 3.0).(*geom.Polygon)
	r := out.Ring(0)
	if len(r) <= len(ring) {
		t.Fatalf("expected polygon ring densified, got %d vertices", len(r))
	}
	for i := 0; i < len(r)-1; i++ {
		if d := segLen(r[i], r[i+1]); d > 3.0+1e-9 {
			t.Errorf("ring segment %d length %g > tol", i, d)
		}
	}
	if r[0] != ring[0] || r[len(r)-1] != ring[len(ring)-1] {
		t.Errorf("ring endpoints not preserved")
	}
}

func TestDensifyNonPositiveTol(t *testing.T) {
	ls := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 100, Y: 0}})
	if Densify(ls, 0).(*geom.LineString).NumPoints() != 2 {
		t.Errorf("zero tol should be no-op")
	}
	if Densify(ls, -1).(*geom.LineString).NumPoints() != 2 {
		t.Errorf("negative tol should be no-op")
	}
}

func TestDensifyMultiLineString(t *testing.T) {
	ls1 := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}})
	ls2 := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 5}, {X: 8, Y: 5}})
	mls := geom.NewMultiLineString(nil, ls1, ls2)
	out := Densify(mls, 3.0).(*geom.MultiLineString)
	if out.NumGeometries() != 2 {
		t.Fatalf("expected 2 lines, got %d", out.NumGeometries())
	}
}
