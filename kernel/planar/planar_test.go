package planar

import (
	"math"
	"testing"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel"
)

var k = Kernel{}

func xy(x, y float64) geom.XY { return geom.XY{X: x, Y: y} }

func TestPlanarSatisfiesKernel(t *testing.T) {
	var _ kernel.Kernel = Kernel{}
	if Default.Name() != "planar" {
		t.Errorf("Default.Name() = %q", Default.Name())
	}
}

func TestDistance(t *testing.T) {
	got := k.Distance(xy(0, 0), xy(3, 4))
	if got != 5 {
		t.Errorf("Distance(3-4-5) = %v", got)
	}
	if k.DistanceSquared(xy(0, 0), xy(3, 4)) != 25 {
		t.Errorf("DistanceSquared = %v", k.DistanceSquared(xy(0, 0), xy(3, 4)))
	}
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
		if got != tc.want {
			t.Errorf("Orient(%v,%v,%v) = %v, want %v", a, b, tc.c, got, tc.want)
		}
	}
}

func TestSegmentIntersection(t *testing.T) {
	cases := []struct {
		name             string
		a1, a2, b1, b2   geom.XY
		want             geom.XY
		ok               bool
	}{
		{"crossing-X", xy(0, 0), xy(2, 2), xy(0, 2), xy(2, 0), xy(1, 1), true},
		{"parallel", xy(0, 0), xy(1, 0), xy(0, 1), xy(1, 1), geom.XY{}, false},
		{"disjoint-collinear", xy(0, 0), xy(1, 0), xy(2, 0), xy(3, 0), geom.XY{}, false},
		{"miss-bound", xy(0, 0), xy(1, 1), xy(2, 0), xy(3, 1), geom.XY{}, false},
	}
	for _, tc := range cases {
		got, ok := k.SegmentIntersection(tc.a1, tc.a2, tc.b1, tc.b2)
		if ok != tc.ok {
			t.Errorf("%s: ok = %v, want %v", tc.name, ok, tc.ok)
		}
		if ok && got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
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
		if got != tc.want {
			t.Errorf("PointInRing(%v) = %v, want %v", tc.p, got, tc.want)
		}
	}
}

func TestRingAreaSign(t *testing.T) {
	ccw := []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 0, Y: 10}, {X: 0, Y: 0}}
	if k.RingArea(ccw) != 100 {
		t.Errorf("CCW area = %v, want 100", k.RingArea(ccw))
	}
	cw := []geom.XY{{X: 0, Y: 0}, {X: 0, Y: 10}, {X: 10, Y: 10}, {X: 10, Y: 0}, {X: 0, Y: 0}}
	if k.RingArea(cw) != -100 {
		t.Errorf("CW area = %v, want -100", k.RingArea(cw))
	}
}

func TestInitialBearingAndDestination(t *testing.T) {
	// Cardinal directions
	cases := []struct {
		a, b geom.XY
		want float64 // degrees clockwise from +Y
	}{
		{xy(0, 0), xy(0, 1), 0},   // north
		{xy(0, 0), xy(1, 0), 90},  // east
		{xy(0, 0), xy(0, -1), 180}, // south
		{xy(0, 0), xy(-1, 0), 270}, // west
	}
	for _, tc := range cases {
		got := k.InitialBearing(tc.a, tc.b)
		if math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("InitialBearing(%v,%v) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
	// Round-trip via Destination
	dst := k.Destination(xy(0, 0), 90, 5)
	if math.Abs(dst.X-5) > 1e-9 || math.Abs(dst.Y) > 1e-9 {
		t.Errorf("Destination east 5 = %v", dst)
	}
}

func TestMidpointAndAngle(t *testing.T) {
	got := k.Midpoint(xy(0, 0), xy(2, 4))
	if got != (xy(1, 2)) {
		t.Errorf("Midpoint = %v", got)
	}
	right := k.AngleBetween(xy(1, 0), xy(0, 0), xy(0, 1))
	if math.Abs(right-math.Pi/2) > 1e-9 {
		t.Errorf("right angle = %v rad", right)
	}
}

func TestSegmentDistance(t *testing.T) {
	d := k.SegmentDistance(xy(1, 1), xy(0, 0), xy(2, 0))
	if math.Abs(d-1) > 1e-9 {
		t.Errorf("SegmentDistance = %v, want 1", d)
	}
	// Past the endpoint: distance to the endpoint, not the line.
	d2 := k.SegmentDistance(xy(5, 0), xy(0, 0), xy(2, 0))
	if math.Abs(d2-3) > 1e-9 {
		t.Errorf("SegmentDistance past-end = %v, want 3", d2)
	}
	// Degenerate segment: falls back to point distance.
	d3 := k.SegmentDistance(xy(0, 0), xy(3, 4), xy(3, 4))
	if d3 != 5 {
		t.Errorf("Degenerate segment distance = %v, want 5", d3)
	}
}
