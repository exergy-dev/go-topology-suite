package topology

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

func TestTraceRingsFromDirectedSegmentsBuildsSquare(t *testing.T) {
	segments := []DirectedSegment{
		{Start: geom.NewCoordinate(0, 0), End: geom.NewCoordinate(10, 0)},
		{Start: geom.NewCoordinate(10, 0), End: geom.NewCoordinate(10, 10)},
		{Start: geom.NewCoordinate(10, 10), End: geom.NewCoordinate(0, 10)},
		{Start: geom.NewCoordinate(0, 10), End: geom.NewCoordinate(0, 0)},
	}

	rings := TraceRingsFromDirectedSegments(segments)

	if len(rings) != 1 {
		t.Fatalf("expected one ring, got %d", len(rings))
	}
	if !rings[0].IsClosed(geom.DefaultEpsilon) {
		t.Fatalf("expected closed ring: %v", rings[0])
	}
	if area := geom.SignedArea(rings[0]); area <= 0 {
		t.Fatalf("expected counter-clockwise ring, got area %v", area)
	}
}

func TestTraceRingsFromDirectedSegmentsIgnoresSeparateDanglingEdge(t *testing.T) {
	segments := []DirectedSegment{
		{Start: geom.NewCoordinate(0, 0), End: geom.NewCoordinate(10, 0)},
		{Start: geom.NewCoordinate(10, 0), End: geom.NewCoordinate(10, 10)},
		{Start: geom.NewCoordinate(10, 10), End: geom.NewCoordinate(0, 10)},
		{Start: geom.NewCoordinate(0, 10), End: geom.NewCoordinate(0, 0)},
		{Start: geom.NewCoordinate(20, 0), End: geom.NewCoordinate(30, 0)},
	}

	rings := TraceRingsFromDirectedSegments(segments)

	if len(rings) != 1 {
		t.Fatalf("expected dangling branch to be ignored, got %d rings", len(rings))
	}
	if len(rings[0]) != 5 {
		t.Fatalf("expected square ring with 5 coordinates, got %v", rings[0])
	}
}
