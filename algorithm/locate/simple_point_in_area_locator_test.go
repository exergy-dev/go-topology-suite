package locate

import (
	"testing"

	"github.com/terra-geo/terra/geom"
)

func square() *geom.Polygon {
	return geom.NewPolygon(nil, []geom.XY{{0, 0}, {10, 0}, {10, 10}, {0, 10}, {0, 0}})
}

// 10x10 square with a 2x2 hole at (4..6, 4..6).
func squareWithHole() *geom.Polygon {
	shell := []geom.XY{{0, 0}, {10, 0}, {10, 10}, {0, 10}, {0, 0}}
	hole := []geom.XY{{4, 4}, {6, 4}, {6, 6}, {4, 6}, {4, 4}}
	return geom.NewPolygon(nil, shell, hole)
}

func TestSimpleLocate_Polygon(t *testing.T) {
	loc := NewSimplePointLocator(square())
	cases := []struct {
		p    geom.XY
		want Location
	}{
		{geom.XY{5, 5}, Interior},
		{geom.XY{0, 0}, Boundary},
		{geom.XY{10, 5}, Boundary},
		{geom.XY{5, 10}, Boundary},
		{geom.XY{-1, 5}, Exterior},
		{geom.XY{20, 20}, Exterior},
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
		{geom.XY{1, 1}, Interior},  // shell interior, outside hole
		{geom.XY{5, 5}, Exterior},  // inside the hole
		{geom.XY{4, 5}, Boundary},  // on hole boundary
		{geom.XY{6, 5}, Boundary},  // on hole boundary
		{geom.XY{0, 5}, Boundary},  // on shell boundary
	}
	for _, c := range cases {
		if got := loc.Locate(c.p); got != c.want {
			t.Errorf("Locate(%v): got %s want %s", c.p, got, c.want)
		}
	}
}

func TestSimpleLocateInGeometry_MultiPolygon(t *testing.T) {
	a := geom.NewPolygon(nil, []geom.XY{{0, 0}, {2, 0}, {2, 2}, {0, 2}, {0, 0}})
	b := geom.NewPolygon(nil, []geom.XY{{10, 10}, {12, 10}, {12, 12}, {10, 12}, {10, 10}})
	mp := geom.NewMultiPolygon(nil, a, b)
	if got := LocateInGeometry(geom.XY{1, 1}, mp); got != Interior {
		t.Errorf("multi-poly inside A: got %s", got)
	}
	if got := LocateInGeometry(geom.XY{11, 11}, mp); got != Interior {
		t.Errorf("multi-poly inside B: got %s", got)
	}
	if got := LocateInGeometry(geom.XY{5, 5}, mp); got != Exterior {
		t.Errorf("multi-poly between: got %s", got)
	}
	if got := LocateInGeometry(geom.XY{2, 1}, mp); got != Boundary {
		t.Errorf("multi-poly on boundary of A: got %s", got)
	}
}

func TestSimpleLocate_Empty(t *testing.T) {
	empty := geom.NewEmptyPolygon(nil, geom.LayoutXY)
	if got := NewSimplePointLocator(empty).Locate(geom.XY{0, 0}); got != Exterior {
		t.Errorf("empty polygon: got %s", got)
	}
	if got := LocateInGeometry(geom.XY{0, 0}, empty); got != Exterior {
		t.Errorf("LocateInGeometry empty: got %s", got)
	}
	if IsContained(geom.XY{0, 0}, empty) {
		t.Errorf("IsContained empty: expected false")
	}
}

func TestLocationString(t *testing.T) {
	if Interior.String() != "INTERIOR" || Boundary.String() != "BOUNDARY" || Exterior.String() != "EXTERIOR" {
		t.Errorf("Location.String drift")
	}
}
