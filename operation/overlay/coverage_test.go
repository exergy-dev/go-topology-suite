package overlay

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/internal/topology"
)

// Additional tests to improve coverage for noded overlay functions

func TestHandleEmptyPolygons(t *testing.T) {
	// Test handleEmptyPolyA
	shell := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)
	polysB := []*geom.Polygon{poly}

	result := handleEmptyPolyA(polysB, OpIntersection)
	if !result.IsEmpty() {
		t.Error("Empty A intersection should be empty")
	}

	result = handleEmptyPolyA(polysB, OpUnion)
	if result.IsEmpty() {
		t.Error("Empty A union should return B")
	}

	result = handleEmptyPolyA(polysB, OpDifference)
	if !result.IsEmpty() {
		t.Error("Empty A difference should be empty")
	}

	result = handleEmptyPolyA(polysB, OpSymDifference)
	if result.IsEmpty() {
		t.Error("Empty A symdifference should return B")
	}

	// Test handleEmptyPolyB
	polysA := []*geom.Polygon{poly}

	result = handleEmptyPolyB(polysA, OpIntersection)
	if !result.IsEmpty() {
		t.Error("Empty B intersection should be empty")
	}

	result = handleEmptyPolyB(polysA, OpUnion)
	if result.IsEmpty() {
		t.Error("Empty B union should return A")
	}

	result = handleEmptyPolyB(polysA, OpDifference)
	if result.IsEmpty() {
		t.Error("Empty B difference should return A")
	}

	result = handleEmptyPolyB(polysA, OpSymDifference)
	if result.IsEmpty() {
		t.Error("Empty B symdifference should return A")
	}
}

func TestPolygonizeEdgesEmpty(t *testing.T) {
	// Test with no edges
	result := polygonizeEdges(nil)
	if !result.IsEmpty() {
		t.Error("Polygonizing no edges should return empty")
	}

	result = polygonizeEdges([]*DirectedEdge{})
	if !result.IsEmpty() {
		t.Error("Polygonizing empty edge list should return empty")
	}
}

func TestNodingEdgeCases(t *testing.T) {
	shell := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)

	graphEdges := topology.BuildPolygonBoundaryGraph([]*geom.Polygon{poly}, nil)
	if len(graphEdges) != 4 {
		t.Errorf("Should build 4 boundary graph edges from square, got %d", len(graphEdges))
	}

	hole := mustLinearRingXY(2, 2, 8, 2, 8, 8, 2, 8, 2, 2)
	polyWithHole := geom.NewPolygon(shell, []*geom.LinearRing{hole})

	graphEdges = topology.BuildPolygonBoundaryGraph([]*geom.Polygon{polyWithHole}, nil)
	if len(graphEdges) != 8 {
		t.Errorf("Should build 8 boundary graph edges from polygon with hole, got %d", len(graphEdges))
	}

	shell1 := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)
	shell2 := mustLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	poly2 := geom.NewPolygon(shell2, nil)

	edges := directedEdgesFromBoundaryGraph(topology.BuildPolygonBoundaryGraph(
		[]*geom.Polygon{poly1},
		[]*geom.Polygon{poly2},
	))
	if len(edges) == 0 {
		t.Fatal("Should build directed edges from polygon boundary graph")
	}

	selectedIntersection := selectEdges(edges, OpIntersection)
	selectedUnion := selectEdges(edges, OpUnion)
	selectedDifference := selectEdges(edges, OpDifference)
	selectedSymDiff := selectEdges(edges, OpSymDifference)

	// Just verify they don't panic and return slices
	_ = selectedIntersection
	_ = selectedUnion
	_ = selectedDifference
	_ = selectedSymDiff
}

func TestPolygonizeLabeledFacesIntersection(t *testing.T) {
	left := geom.NewPolygon(
		mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0),
		nil,
	)
	right := geom.NewPolygon(
		mustLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5),
		nil,
	)

	result := polygonizeLabeledFaces(
		topology.BuildPolygonBoundaryGraph([]*geom.Polygon{left}, []*geom.Polygon{right}),
		OpIntersection,
	)
	poly, ok := result.(*geom.Polygon)
	if !ok {
		t.Fatalf("expected polygon intersection from labeled face helper, got %T", result)
	}
	if area := poly.Area(); area != 25 {
		t.Fatalf("expected labeled face intersection area 25, got %v", area)
	}
}
