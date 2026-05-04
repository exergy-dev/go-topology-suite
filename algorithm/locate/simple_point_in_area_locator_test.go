package locate

import (
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
)

func square() *geom.Polygon {
	return geom.NewPolygon(nil, []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0}})
}

// 10x10 square with a 2x2 hole at (4..6, 4..6).
func squareWithHole() *geom.Polygon {
	shell := []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0}}
	hole := []geom.XY{{X: 4, Y: 4}, {X: 6, Y: 4}, {X: 6, Y: 6}, {X: 4, Y: 6}, {X: 4, Y: 4}}
	return geom.NewPolygon(nil, shell, hole)
}

func TestSimpleLocate_Polygon(t *testing.T) {
	loc := NewSimplePointLocator(square())
	cases := []struct {
		p    geom.XY
		want Location
	}{
		{geom.XY{X: 5, Y: 5}, Interior},
		{geom.XY{X: 0, Y: 0}, Boundary},
		{geom.XY{X: 10, Y: 5}, Boundary},
		{geom.XY{X: 5, Y: 10}, Boundary},
		{geom.XY{X: -1, Y: 5}, Exterior},
		{geom.XY{X: 20, Y: 20}, Exterior},
	}
	for _, c := range cases {
		if got := loc.Locate(c.p); got != c.want {
			t.Errorf("Locate(%v): got %s want %s", c.p, got, c.want)
		}
	}
}

func TestSimpleLocate_Hole(t *testing.T) {
	loc := NewSimplePointLocator(squareWithHole())
	cases := []struct {
		p    geom.XY
		want Location
	}{
		{geom.XY{X: 1, Y: 1}, Interior}, // shell interior, outside hole
		{geom.XY{X: 5, Y: 5}, Exterior}, // inside the hole
		{geom.XY{X: 4, Y: 5}, Boundary}, // on hole boundary
		{geom.XY{X: 6, Y: 5}, Boundary}, // on hole boundary
		{geom.XY{X: 0, Y: 5}, Boundary}, // on shell boundary
	}
	for _, c := range cases {
		if got := loc.Locate(c.p); got != c.want {
			t.Errorf("Locate(%v): got %s want %s", c.p, got, c.want)
		}
	}
}

func TestSimpleLocateInGeometry_MultiPolygon(t *testing.T) {
	a := geom.NewPolygon(nil, []geom.XY{{X: 0, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 2}, {X: 0, Y: 2}, {X: 0, Y: 0}})
	b := geom.NewPolygon(nil, []geom.XY{{X: 10, Y: 10}, {X: 12, Y: 10}, {X: 12, Y: 12}, {X: 10, Y: 12}, {X: 10, Y: 10}})
	mp := geom.NewMultiPolygon(nil, a, b)
	if got := LocateInGeometry(geom.XY{X: 1, Y: 1}, mp); got != Interior {
		t.Errorf("multi-poly inside A: got %s", got)
	}
	if got := LocateInGeometry(geom.XY{X: 11, Y: 11}, mp); got != Interior {
		t.Errorf("multi-poly inside B: got %s", got)
	}
	if got := LocateInGeometry(geom.XY{X: 5, Y: 5}, mp); got != Exterior {
		t.Errorf("multi-poly between: got %s", got)
	}
	if got := LocateInGeometry(geom.XY{X: 2, Y: 1}, mp); got != Boundary {
		t.Errorf("multi-poly on boundary of A: got %s", got)
	}
}

func TestSimpleLocate_Empty(t *testing.T) {
	empty := geom.NewEmptyPolygon(nil, geom.LayoutXY)
	if got := NewSimplePointLocator(empty).Locate(geom.XY{X: 0, Y: 0}); got != Exterior {
		t.Errorf("empty polygon: got %s", got)
	}
	if got := LocateInGeometry(geom.XY{X: 0, Y: 0}, empty); got != Exterior {
		t.Errorf("LocateInGeometry empty: got %s", got)
	}
	if IsContained(geom.XY{X: 0, Y: 0}, empty) {
		t.Errorf("IsContained empty: expected false")
	}
}

func TestLocationString(t *testing.T) {
	if Interior.String() != "INTERIOR" || Boundary.String() != "BOUNDARY" || Exterior.String() != "EXTERIOR" {
		t.Errorf("Location.String drift")
	}
}
