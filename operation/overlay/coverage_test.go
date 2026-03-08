package overlay

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/noding"
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
	// Test extracting segment strings from polygons
	shell := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)

	segStrings := extractSegmentStringsFromPolygons([]*geom.Polygon{poly}, 0)
	if len(segStrings) != 4 {
		t.Errorf("Should extract 4 edges from square, got %d", len(segStrings))
	}

	// Test with polygon with hole
	hole := mustLinearRingXY(2, 2, 8, 2, 8, 8, 2, 8, 2, 2)
	polyWithHole := geom.NewPolygon(shell, []*geom.LinearRing{hole})

	segStrings = extractSegmentStringsFromPolygons([]*geom.Polygon{polyWithHole}, 0)
	if len(segStrings) != 8 {
		t.Errorf("Should extract 8 edges from polygon with hole, got %d", len(segStrings))
	}

	// Test building directed edges
	coords1 := geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
	}
	ss1 := noding.NewNodedSegmentString(coords1, &EdgeContext{Source: 0, IsHole: false})

	edges := buildDirectedEdges([]*noding.NodedSegmentString{ss1})
	if len(edges) != 1 {
		t.Errorf("Should build 1 edge, got %d", len(edges))
	}

	// Test labeling edges
	shell1 := mustLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)
	shell2 := mustLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
	poly2 := geom.NewPolygon(shell2, nil)

	labelEdges(edges, []*geom.Polygon{poly1}, []*geom.Polygon{poly2})
	if len(edges) != 1 {
		t.Error("Label edges should not modify edge count")
	}

	// Test select edges
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
