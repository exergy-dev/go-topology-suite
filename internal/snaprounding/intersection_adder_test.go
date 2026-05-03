package snaprounding

import (
	"testing"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/internal/noding"
)

func TestIntersectionAdder_Cross(t *testing.T) {
	a := &noding.SegmentString{Coords: []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 10}}}
	b := &noding.SegmentString{Coords: []geom.XY{{X: 0, Y: 10}, {X: 10, Y: 0}}}
	ad := NewIntersectionAdder(0.01)
	ad.Process([]*noding.SegmentString{a, b})
	pts := ad.Points()
	found := false
	for _, p := range pts {
		if p.EqualBitwise(geom.XY{X: 5, Y: 5}) {
			found = true
		}
	}
	if !found {
		t.Errorf("expected (5,5) in %v", pts)
	}
}

func TestIntersectionAdder_TouchAtEndpoint(t *testing.T) {
	// Segments share an endpoint — JTS records this as a non-interior
	// intersection and skips it.
	a := &noding.SegmentString{Coords: []geom.XY{{X: 0, Y: 0}, {X: 5, Y: 5}}}
	b := &noding.SegmentString{Coords: []geom.XY{{X: 5, Y: 5}, {X: 10, Y: 0}}}
	ad := NewIntersectionAdder(0.01)
	ad.Process([]*noding.SegmentString{a, b})
	if len(ad.Points()) != 0 {
		t.Errorf("expected no interior intersection, got %v", ad.Points())
	}
}

func TestIntersectionAdder_NearVertex(t *testing.T) {
	// Endpoint of one segment lies within tol of the interior of another.
	a := &noding.SegmentString{Coords: []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}}}
	b := &noding.SegmentString{Coords: []geom.XY{{X: 5, Y: 0.001}, {X: 5, Y: 5}}}
	ad := NewIntersectionAdder(0.1)
	ad.Process([]*noding.SegmentString{a, b})
	pts := ad.Points()
	if len(pts) == 0 {
		t.Fatalf("expected near-vertex hit, got none")
	}
	found := false
	for _, p := range pts {
		if p.EqualBitwise(geom.XY{X: 5, Y: 0.001}) {
			found = true
		}
	}
	if !found {
		t.Errorf("expected (5, 0.001) in %v", pts)
	}
}

func TestIntersectionAdder_NoCross(t *testing.T) {
	a := &noding.SegmentString{Coords: []geom.XY{{X: 0, Y: 0}, {X: 1, Y: 1}}}
	b := &noding.SegmentString{Coords: []geom.XY{{X: 10, Y: 10}, {X: 20, Y: 20}}}
	ad := NewIntersectionAdder(0.01)
	ad.Process([]*noding.SegmentString{a, b})
	if got := ad.Points(); len(got) != 0 {
		t.Errorf("expected no intersections, got %v", got)
	}
}
