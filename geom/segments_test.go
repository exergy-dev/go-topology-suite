package geom_test

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

func TestSegmentsIntersect_EdgeCases(t *testing.T) {
	t.Run("ProperCrossing", func(t *testing.T) {
		a1 := geom.NewCoordinate(0, 0)
		a2 := geom.NewCoordinate(10, 10)
		b1 := geom.NewCoordinate(0, 10)
		b2 := geom.NewCoordinate(10, 0)
		if !geom.SegmentsIntersect(a1, a2, b1, b2) {
			t.Error("Expected crossing segments to intersect")
		}
	})

	t.Run("EndpointTouch", func(t *testing.T) {
		a1 := geom.NewCoordinate(0, 0)
		a2 := geom.NewCoordinate(10, 0)
		b1 := geom.NewCoordinate(10, 0)
		b2 := geom.NewCoordinate(10, 10)
		if !geom.SegmentsIntersect(a1, a2, b1, b2) {
			t.Error("Expected endpoint-touching segments to intersect")
		}
	})

	t.Run("CollinearOverlap", func(t *testing.T) {
		a1 := geom.NewCoordinate(0, 0)
		a2 := geom.NewCoordinate(10, 0)
		b1 := geom.NewCoordinate(5, 0)
		b2 := geom.NewCoordinate(15, 0)
		if !geom.SegmentsIntersect(a1, a2, b1, b2) {
			t.Error("Expected overlapping collinear segments to intersect")
		}
	})

	t.Run("CollinearDisjoint", func(t *testing.T) {
		a1 := geom.NewCoordinate(0, 0)
		a2 := geom.NewCoordinate(10, 0)
		b1 := geom.NewCoordinate(11, 0)
		b2 := geom.NewCoordinate(20, 0)
		if geom.SegmentsIntersect(a1, a2, b1, b2) {
			t.Error("Expected disjoint collinear segments to not intersect")
		}
	})
}
