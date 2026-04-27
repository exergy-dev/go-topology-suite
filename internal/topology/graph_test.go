package topology

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

func TestBuildPolygonBoundaryGraphLabelsSharedEdgeSides(t *testing.T) {
	left := geom.NewPolygon(
		mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0),
		nil,
	)
	right := geom.NewPolygon(
		mustLinearRingXY(10, 0, 20, 0, 20, 10, 10, 10, 10, 0),
		nil,
	)

	edges := BuildPolygonBoundaryGraph([]*geom.Polygon{left}, []*geom.Polygon{right})

	for _, edge := range edges {
		if !edge.Sources.InA() || !edge.Sources.InB() {
			continue
		}
		if !edge.Start.Equals2D(geom.NewCoordinate(10, 0), geom.DefaultEpsilon) ||
			!edge.End.Equals2D(geom.NewCoordinate(10, 10), geom.DefaultEpsilon) {
			t.Fatalf("unexpected shared edge: %+v", edge)
		}
		if edge.Left.LocA == edge.Right.LocA || edge.Left.LocB == edge.Right.LocB {
			t.Fatalf("shared boundary should separate interior/exterior labels: %+v", edge)
		}
		return
	}
	t.Fatal("expected shared edge in boundary graph")
}

func TestBuildPolygonBoundaryGraphSplitsCrossingBoundaries(t *testing.T) {
	square := geom.NewPolygon(
		mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0),
		nil,
	)
	cutter := geom.NewPolygon(
		mustLinearRingXY(5, -5, 15, 5, 5, 15, -5, 5, 5, -5),
		nil,
	)

	edges := BuildPolygonBoundaryGraph([]*geom.Polygon{square}, []*geom.Polygon{cutter})

	for _, edge := range edges {
		if edge.Start.Equals2D(geom.NewCoordinate(0, 0), geom.DefaultEpsilon) ||
			edge.End.Equals2D(geom.NewCoordinate(0, 0), geom.DefaultEpsilon) {
			return
		}
	}
	t.Fatal("expected noded boundary graph to include crossing endpoint")
}

func TestBuildPolygonBoundaryGraphWithPrecisionSnapsEdgesAndLabelsWithoutMutatingInputs(t *testing.T) {
	a := geom.NewPolygon(
		mustLinearRingXY(0.49, 0.49, 2.49, 0.49, 2.49, 2.49, 0.49, 2.49, 0.49, 0.49),
		nil,
	)

	edges := BuildPolygonBoundaryGraphWithPrecision(
		[]*geom.Polygon{a},
		nil,
		geom.NewFixedPrecision(1),
	)

	for _, edge := range edges {
		if !edge.Sources.InA() || edge.Sources.InB() {
			t.Fatalf("single-input graph edge has unexpected sources: %+v", edge)
		}
		if !edge.Start.Equals2D(geom.NewCoordinate(2, 0), geom.DefaultEpsilon) ||
			!edge.End.Equals2D(geom.NewCoordinate(2, 2), geom.DefaultEpsilon) {
			continue
		}
		if edge.Left.LocA != geom.LocationInterior || edge.Right.LocA != geom.LocationExterior {
			t.Fatalf("expected labels from snapped polygon boundary, got %+v", edge)
		}
		if got := a.ExteriorRing().Coordinates()[0]; got.X != 0.49 || got.Y != 0.49 {
			t.Fatalf("precision graph mutated input polygon coordinate: %v", got)
		}
		return
	}
	t.Fatalf("expected snapped vertical edge from (2,0) to (2,2), got %#v", edges)
}
