package locate

import (
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
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
		assert.Equalf(t, c.want, loc.Locate(c.p), "Locate(%v)", c.p)
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
		assert.Equalf(t, c.want, loc.Locate(c.p), "Locate(%v)", c.p)
	}
}

func TestSimpleLocateInGeometry_MultiPolygon(t *testing.T) {
	a := geom.NewPolygon(nil, []geom.XY{{X: 0, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 2}, {X: 0, Y: 2}, {X: 0, Y: 0}})
	b := geom.NewPolygon(nil, []geom.XY{{X: 10, Y: 10}, {X: 12, Y: 10}, {X: 12, Y: 12}, {X: 10, Y: 12}, {X: 10, Y: 10}})
	mp := geom.NewMultiPolygon(nil, a, b)
	assert.Equalf(t, Interior, LocateInGeometry(geom.XY{X: 1, Y: 1}, mp), "multi-poly inside A")
	assert.Equalf(t, Interior, LocateInGeometry(geom.XY{X: 11, Y: 11}, mp), "multi-poly inside B")
	assert.Equalf(t, Exterior, LocateInGeometry(geom.XY{X: 5, Y: 5}, mp), "multi-poly between")
	assert.Equalf(t, Boundary, LocateInGeometry(geom.XY{X: 2, Y: 1}, mp), "multi-poly on boundary of A")
}

func TestSimpleLocate_Empty(t *testing.T) {
	empty := geom.NewEmptyPolygon(nil, geom.LayoutXY)
	assert.Equalf(t, Exterior, NewSimplePointLocator(empty).Locate(geom.XY{X: 0, Y: 0}), "empty polygon")
	assert.Equalf(t, Exterior, LocateInGeometry(geom.XY{X: 0, Y: 0}, empty), "LocateInGeometry empty")
	assert.Falsef(t, IsContained(geom.XY{X: 0, Y: 0}, empty), "IsContained empty")
}

func TestLocationString(t *testing.T) {
	assert.Equal(t, "INTERIOR", Interior.String())
	assert.Equal(t, "BOUNDARY", Boundary.String())
	assert.Equal(t, "EXTERIOR", Exterior.String())
}
