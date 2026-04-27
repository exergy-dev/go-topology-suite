package topology

import (
	"math"
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

func TestTracePolygonBoundaryFacesLabelsSinglePolygonInterior(t *testing.T) {
	poly := geom.NewPolygon(
		mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0),
		nil,
	)

	faces := TracePolygonBoundaryFaces(BuildPolygonBoundaryGraph([]*geom.Polygon{poly}, nil))

	face, ok := findLabeledFace(faces, geom.LocationInterior, geom.LocationExterior)
	if !ok {
		t.Fatalf("expected A-only interior face, got %#v", faces)
	}
	if area := math.Abs(geom.SignedArea(face.Ring)); area != 100 {
		t.Fatalf("expected square face area 100, got %v", area)
	}
}

func TestTracePolygonBoundaryFacesLabelsOverlappingInteriorFace(t *testing.T) {
	left := geom.NewPolygon(
		mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0),
		nil,
	)
	right := geom.NewPolygon(
		mustLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5),
		nil,
	)

	faces := TracePolygonBoundaryFaces(BuildPolygonBoundaryGraph(
		[]*geom.Polygon{left},
		[]*geom.Polygon{right},
	))

	face, ok := findLabeledFace(faces, geom.LocationInterior, geom.LocationInterior)
	if !ok {
		t.Fatalf("expected A+B interior face, got %#v", faces)
	}
	if area := math.Abs(geom.SignedArea(face.Ring)); area != 25 {
		t.Fatalf("expected overlap face area 25, got %v", area)
	}
}

func TestTracePolygonBoundaryFacesLabelsHoleAsExteriorFace(t *testing.T) {
	shell := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	hole := mustLinearRingXY(2, 2, 8, 2, 8, 8, 2, 8, 2, 2)
	poly := geom.NewPolygon(shell, []*geom.LinearRing{hole})

	faces := TracePolygonBoundaryFaces(BuildPolygonBoundaryGraph([]*geom.Polygon{poly}, nil))

	for _, face := range faces {
		if face.Label.LocA != geom.LocationExterior || face.Label.LocB != geom.LocationExterior {
			continue
		}
		if area := math.Abs(geom.SignedArea(face.Ring)); area == 36 {
			return
		}
	}
	t.Fatalf("expected hole face labeled exterior with area 36, got %#v", faces)
}

func findLabeledFace(faces []LabeledRing, locA, locB geom.Location) (LabeledRing, bool) {
	for _, face := range faces {
		if face.Label.LocA == locA && face.Label.LocB == locB {
			return face, true
		}
	}
	return LabeledRing{}, false
}
