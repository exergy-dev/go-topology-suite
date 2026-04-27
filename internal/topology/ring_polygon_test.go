package topology

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

func TestPolygonsFromRingsAssignsHoleToShell(t *testing.T) {
	shell := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0).Coordinates()
	hole := mustLinearRingXY(2, 2, 2, 8, 8, 8, 8, 2, 2, 2).Coordinates()

	polygons := PolygonsFromRings([]geom.CoordinateSequence{shell, hole}, false)

	if len(polygons) != 1 {
		t.Fatalf("expected one polygon, got %d", len(polygons))
	}
	if polygons[0].NumInteriorRings() != 1 {
		t.Fatalf("expected one assigned hole, got %d", polygons[0].NumInteriorRings())
	}
}

func TestDeduplicateRingsIgnoresDirectionAndStartVertex(t *testing.T) {
	ring := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0).Coordinates()
	sameRing := geom.CoordinateSequence{
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(10, 10),
	}

	rings := DeduplicateRings([]geom.CoordinateSequence{ring, sameRing})

	if len(rings) != 1 {
		t.Fatalf("expected duplicate rings to collapse to one, got %d", len(rings))
	}
}
