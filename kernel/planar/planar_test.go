package planar

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel"
)

var k = Kernel{}

func xy(x, y float64) geom.XY { return geom.XY{X: x, Y: y} }

func TestPlanarSatisfiesKernel(t *testing.T) {
	var _ kernel.Kernel = Kernel{}
	assert.Equal(t, "planar", Default.Name(), "Default.Name()")
}

func TestDistance(t *testing.T) {
	got := k.Distance(xy(0, 0), xy(3, 4))
	assert.Equal(t, 5.0, got, "Distance(3-4-5)")
	assert.Equal(t, 25.0, k.DistanceSquared(xy(0, 0), xy(3, 4)), "DistanceSquared")
}

func TestOrient(t *testing.T) {
	a, b := xy(0, 0), xy(1, 0)
	cases := []struct {
		c    geom.XY
		want kernel.Orientation
	}{
		{xy(0, 1), kernel.CounterClockwise},
		{xy(0, -1), kernel.Clockwise},
		{xy(2, 0), kernel.Collinear},
	}
	for _, tc := range cases {
		got := k.Orient(a, b, tc.c)
		assert.Equalf(t, tc.want, got, "Orient(%v,%v,%v)", a, b, tc.c)
	}
}

func TestSegmentIntersection(t *testing.T) {
	cases := []struct {
		name           string
		a1, a2, b1, b2 geom.XY
		want           geom.XY
		ok             bool
	}{
		{"crossing-X", xy(0, 0), xy(2, 2), xy(0, 2), xy(2, 0), xy(1, 1), true},
		{"parallel", xy(0, 0), xy(1, 0), xy(0, 1), xy(1, 1), geom.XY{}, false},
		{"disjoint-collinear", xy(0, 0), xy(1, 0), xy(2, 0), xy(3, 0), geom.XY{}, false},
		{"miss-bound", xy(0, 0), xy(1, 1), xy(2, 0), xy(3, 1), geom.XY{}, false},
	}
	for _, tc := range cases {
		got, ok := k.SegmentIntersection(tc.a1, tc.a2, tc.b1, tc.b2)
		assert.Equalf(t, tc.ok, ok, "%s: ok", tc.name)
		if ok {
			assert.Equalf(t, tc.want, got, "%s", tc.name)
		}
	}
}

func TestPointInRing(t *testing.T) {
	square := []geom.XY{{X: 0, Y: 0}, {X: 0, Y: 10}, {X: 10, Y: 10}, {X: 10, Y: 0}, {X: 0, Y: 0}}
	cases := []struct {
		p    geom.XY
		want kernel.Containment
	}{
		{xy(5, 5), kernel.Inside},
		{xy(15, 5), kernel.Outside},
		{xy(0, 5), kernel.OnBoundary},
		{xy(0, 0), kernel.OnBoundary},
		{xy(10, 10), kernel.OnBoundary},
	}
	for _, tc := range cases {
		got := k.PointInRing(tc.p, square)
		assert.Equalf(t, tc.want, got, "PointInRing(%v)", tc.p)
	}
}

func TestRingAreaSign(t *testing.T) {
	ccw := []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0}}
	assert.Equal(t, 100.0, k.RingArea(ccw), "CCW area")
	cw := []geom.XY{{X: 0, Y: 0}, {X: 0, Y: 10}, {X: 10, Y: 10}, {X: 10, Y: 0}, {X: 0, Y: 0}}
	assert.Equal(t, -100.0, k.RingArea(cw), "CW area")
}

func TestInitialBearingAndDestination(t *testing.T) {
	// Cardinal directions
	cases := []struct {
		a, b geom.XY
		want float64 // degrees clockwise from +Y
	}{
		{xy(0, 0), xy(0, 1), 0},    // north
		{xy(0, 0), xy(1, 0), 90},   // east
		{xy(0, 0), xy(0, -1), 180}, // south
		{xy(0, 0), xy(-1, 0), 270}, // west
	}
	for _, tc := range cases {
		got := k.InitialBearing(tc.a, tc.b)
		assert.InDeltaf(t, tc.want, got, 1e-9, "InitialBearing(%v,%v)", tc.a, tc.b)
	}
	// Round-trip via Destination
	dst := k.Destination(xy(0, 0), 90, 5)
	assert.InDeltaf(t, 5.0, dst.X, 1e-9, "Destination east 5 X = %v", dst)
	assert.InDeltaf(t, 0.0, dst.Y, 1e-9, "Destination east 5 Y = %v", dst)
}

func TestMidpointAndAngle(t *testing.T) {
	got := k.Midpoint(xy(0, 0), xy(2, 4))
	assert.Equal(t, xy(1, 2), got, "Midpoint")
	right := k.AngleBetween(xy(1, 0), xy(0, 0), xy(0, 1))
	assert.InDelta(t, math.Pi/2, right, 1e-9, "right angle")
}

func TestSegmentDistance(t *testing.T) {
	d := k.SegmentDistance(xy(1, 1), xy(0, 0), xy(2, 0))
	assert.InDelta(t, 1.0, d, 1e-9, "SegmentDistance")
	// Past the endpoint: distance to the endpoint, not the line.
	d2 := k.SegmentDistance(xy(5, 0), xy(0, 0), xy(2, 0))
	assert.InDelta(t, 3.0, d2, 1e-9, "SegmentDistance past-end")
	// Degenerate segment: falls back to point distance.
	d3 := k.SegmentDistance(xy(0, 0), xy(3, 4), xy(3, 4))
	assert.Equal(t, 5.0, d3, "Degenerate segment distance")
}
