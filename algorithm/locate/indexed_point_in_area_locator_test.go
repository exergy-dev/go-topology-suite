package locate

import (
	"math"
	"math/rand"
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexedLocate_Polygon(t *testing.T) {
	loc := NewIndexedPointLocator(square())
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

func TestIndexedLocate_Hole(t *testing.T) {
	loc := NewIndexedPointLocator(squareWithHole())
	cases := []struct {
		p    geom.XY
		want Location
	}{
		{geom.XY{X: 1, Y: 1}, Interior},
		{geom.XY{X: 5, Y: 5}, Exterior},
		{geom.XY{X: 4, Y: 5}, Boundary},
		{geom.XY{X: 6, Y: 5}, Boundary},
	}
	for _, c := range cases {
		assert.Equalf(t, c.want, loc.Locate(c.p), "Locate(%v)", c.p)
	}
}

// Build a non-trivial polygon (a regular 32-gon with a smaller offset
// hexagonal hole) and verify that the indexed locator agrees with the
// simple locator on many random query points.
func TestIndexedMatchesSimple_Random(t *testing.T) {
	const n = 32
	shell := make([]geom.XY, 0, n+1)
	for i := 0; i < n; i++ {
		ang := 2 * math.Pi * float64(i) / n
		shell = append(shell, geom.XY{X: 50 + 40*math.Cos(ang), Y: 50 + 40*math.Sin(ang)})
	}
	shell = append(shell, shell[0])

	hole := make([]geom.XY, 0, 7)
	// CW hole (the locator works on either orientation since it is a
	// crossing-count algorithm).
	for i := 0; i < 6; i++ {
		ang := -2 * math.Pi * float64(i) / 6
		hole = append(hole, geom.XY{X: 60 + 8*math.Cos(ang), Y: 50 + 8*math.Sin(ang)})
	}
	hole = append(hole, hole[0])

	poly := geom.NewPolygon(nil, shell, hole)

	indexed := NewIndexedPointLocator(poly)
	simple := NewSimplePointLocator(poly)

	r := rand.New(rand.NewSource(42))
	for i := 0; i < 1000; i++ {
		p := geom.XY{X: r.Float64() * 100, Y: r.Float64() * 100}
		want := simple.Locate(p)
		got := indexed.Locate(p)
		require.Equalf(t, want, got, "disagree at %v", p)
	}
}

func TestIndexedLocate_MultiPolygon(t *testing.T) {
	a := geom.NewPolygon(nil, []geom.XY{{X: 0, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 2}, {X: 0, Y: 2}, {X: 0, Y: 0}})
	b := geom.NewPolygon(nil, []geom.XY{{X: 10, Y: 10}, {X: 12, Y: 10}, {X: 12, Y: 12}, {X: 10, Y: 12}, {X: 10, Y: 10}})
	mp := geom.NewMultiPolygon(nil, a, b)
	loc := NewIndexedPointLocator(mp)
	assert.Equalf(t, Interior, loc.Locate(geom.XY{X: 1, Y: 1}), "inside A")
	assert.Equalf(t, Interior, loc.Locate(geom.XY{X: 11, Y: 11}), "inside B")
	assert.Equalf(t, Exterior, loc.Locate(geom.XY{X: 5, Y: 5}), "between")
}

func TestIndexedLocate_Empty(t *testing.T) {
	empty := geom.NewEmptyPolygon(nil, geom.LayoutXY)
	loc := NewIndexedPointLocator(empty)
	assert.Equalf(t, Exterior, loc.Locate(geom.XY{X: 0, Y: 0}), "empty")
}

// Both locators should satisfy the PointOnGeometryLocator interface.
func TestPointOnGeometryLocatorInterface(t *testing.T) {
	var _ PointOnGeometryLocator = NewSimplePointLocator(square())
	var _ PointOnGeometryLocator = NewIndexedPointLocator(square())
}
