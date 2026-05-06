package snaprounding

import (
	"testing"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/internal/noding"
	"github.com/stretchr/testify/assert"
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
	assert.Truef(t, found, "expected (5,5) in %v", pts)
}

func TestIntersectionAdder_TouchAtEndpoint(t *testing.T) {
	// Segments share an endpoint — JTS records this as a non-interior
	// intersection and skips it.
	a := &noding.SegmentString{Coords: []geom.XY{{X: 0, Y: 0}, {X: 5, Y: 5}}}
	b := &noding.SegmentString{Coords: []geom.XY{{X: 5, Y: 5}, {X: 10, Y: 0}}}
	ad := NewIntersectionAdder(0.01)
	ad.Process([]*noding.SegmentString{a, b})
	assert.Equalf(t, 0, len(ad.Points()), "expected no interior intersection, got %v", ad.Points())
}

func TestIntersectionAdder_NearVertex(t *testing.T) {
	// Endpoint of one segment lies within tol of the interior of another.
	a := &noding.SegmentString{Coords: []geom.XY{{X: 0, Y: 0}, {X: 10, Y: 0}}}
	b := &noding.SegmentString{Coords: []geom.XY{{X: 5, Y: 0.001}, {X: 5, Y: 5}}}
	ad := NewIntersectionAdder(0.1)
	ad.Process([]*noding.SegmentString{a, b})
	pts := ad.Points()
	assert.NotEqualf(t, 0, len(pts), "expected near-vertex hit, got none")
	found := false
	for _, p := range pts {
		if p.EqualBitwise(geom.XY{X: 5, Y: 0.001}) {
			found = true
		}
	}
	assert.Truef(t, found, "expected (5, 0.001) in %v", pts)
}

func TestIntersectionAdder_NoCross(t *testing.T) {
	a := &noding.SegmentString{Coords: []geom.XY{{X: 0, Y: 0}, {X: 1, Y: 1}}}
	b := &noding.SegmentString{Coords: []geom.XY{{X: 10, Y: 10}, {X: 20, Y: 20}}}
	ad := NewIntersectionAdder(0.01)
	ad.Process([]*noding.SegmentString{a, b})
	got := ad.Points()
	assert.Equalf(t, 0, len(got), "expected no intersections, got %v", got)
}
