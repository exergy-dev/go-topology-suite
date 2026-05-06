package measure

import (
	"math"
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func approxXY(a, b geom.XY, tol float64) bool {
	return math.Abs(a.X-b.X) <= tol && math.Abs(a.Y-b.Y) <= tol
}

func TestCentroidBuilderEmpty(t *testing.T) {
	b := NewCentroidBuilder()
	_, ok := b.Centroid()
	assert.False(t, ok, "empty builder must report no centroid")
}

func TestCentroidBuilderSinglePolygonAgreesWithOneShot(t *testing.T) {
	pts := []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0}}
	poly := geom.NewPolygon(nil, pts)
	b := NewCentroidBuilder()
	b.Add(poly)
	got, ok := b.Centroid()
	require.True(t, ok, "Centroid() returned !ok")
	want := geom.XY{X: 5, Y: 5}
	assert.True(t, approxXY(got, want, 1e-12), "centroid: got %v, want %v", got, want)
	// One-shot Centroid agrees.
	one := Centroid(poly)
	assert.True(t, approxXY(got, one.XY(), 1e-9), "builder %v vs one-shot %v disagree", got, one.XY())
}

func TestCentroidBuilderPointsOnly(t *testing.T) {
	b := NewCentroidBuilder()
	b.Add(geom.NewPoint(nil, geom.XY{X: 0, Y: 0}))
	b.Add(geom.NewPoint(nil, geom.XY{X: 10, Y: 0}))
	b.Add(geom.NewPoint(nil, geom.XY{X: 5, Y: 6}))
	got, ok := b.Centroid()
	require.True(t, ok, "expected centroid")
	want := geom.XY{X: 5, Y: 2}
	assert.True(t, approxXY(got, want, 1e-12), "3-point average: got %v, want %v", got, want)
}

func TestCentroidBuilderLineString(t *testing.T) {
	ls := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}})
	b := NewCentroidBuilder()
	b.Add(ls)
	got, ok := b.Centroid()
	require.True(t, ok, "expected centroid")
	want := geom.XY{X: 5, Y: 0}
	assert.True(t, approxXY(got, want, 1e-12), "linestring centroid: got %v, want %v", got, want)
}

func TestCentroidBuilderMixedDimensionsAreaWins(t *testing.T) {
	// Adding a polygon and a stray point: the polygon dimension (2)
	// dominates so the result must be the polygon centroid.
	pts := []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0}}
	poly := geom.NewPolygon(nil, pts)
	b := NewCentroidBuilder()
	b.Add(poly)
	b.Add(geom.NewPoint(nil, geom.XY{X: 1000, Y: 1000}))
	b.Add(geom.NewLineString(nil, []geom.XY{{X: -100, Y: -100}, {X: -200, Y: -200}}))
	got, ok := b.Centroid()
	require.True(t, ok, "expected centroid")
	want := geom.XY{X: 5, Y: 5}
	assert.True(t, approxXY(got, want, 1e-9), "mixed-dim: areal must dominate, got %v want %v", got, want)
}

func TestCentroidBuilderTwoPolygonsCombined(t *testing.T) {
	// Two unit squares: one at (0,0)-(10,10), one at (20,0)-(30,10).
	// Each centroid is (5,5) and (25,5); equal areas; combined is (15,5).
	p1 := geom.NewPolygon(nil, []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0}})
	p2 := geom.NewPolygon(nil, []geom.XY{{X: 20, Y: 0}, {X: 30, Y: 0}, {X: 30, Y: 10}, {X: 20, Y: 10}, {X: 20, Y: 0}})
	b := NewCentroidBuilder()
	b.Add(p1)
	b.Add(p2)
	got, ok := b.Centroid()
	require.True(t, ok, "expected centroid")
	want := geom.XY{X: 15, Y: 5}
	assert.True(t, approxXY(got, want, 1e-9), "two-square combined: got %v, want %v", got, want)
}

func TestCentroidBuilderPolygonWithHole(t *testing.T) {
	// Big square (0..10,0..10) with a centred 4x4 hole — symmetric, so
	// centroid is exactly (5,5).
	shell := []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0}}
	hole := []geom.XY{{X: 3, Y: 3}, {X: 7, Y: 3}, {X: 7, Y: 7}, {X: 3, Y: 7}, {X: 3, Y: 3}}
	poly := geom.NewPolygon(nil, shell, hole)
	b := NewCentroidBuilder()
	b.Add(poly)
	got, ok := b.Centroid()
	require.True(t, ok, "expected centroid")
	want := geom.XY{X: 5, Y: 5}
	assert.True(t, approxXY(got, want, 1e-9), "with hole: got %v, want %v", got, want)
}

func TestCentroidBuilderLinealOnlyAgreesWithOneShot(t *testing.T) {
	ls := geom.NewLineString(nil, []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}})
	b := NewCentroidBuilder()
	b.Add(ls)
	got, ok := b.Centroid()
	require.True(t, ok, "expected centroid")
	one := Centroid(ls)
	assert.True(t, approxXY(got, one.XY(), 1e-9), "lineal one-shot vs builder: got %v vs %v", got, one.XY())
}
