package overlay

import (
	"testing"

	"github.com/go-topology-suite/gts/geom"
	"github.com/go-topology-suite/gts/noding"
)

// Additional tests to improve coverage

func TestEmptyGeometryHandling(t *testing.T) {
	// Create a simple polygon
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)
	polys := []*geom.Polygon{poly}

	// Test intersection with empty
	result := polygonPolygonIntersection(nil, polys)
	if !result.IsEmpty() {
		t.Error("Intersection with empty should be empty")
	}

	result = polygonPolygonIntersection(polys, nil)
	if !result.IsEmpty() {
		t.Error("Intersection with empty should be empty")
	}

	// Test union with empty
	result = polygonPolygonUnion(nil, polys)
	if result.IsEmpty() {
		t.Error("Union with empty should return non-empty polygon")
	}

	result = polygonPolygonUnion(polys, nil)
	if result.IsEmpty() {
		t.Error("Union with empty should return non-empty polygon")
	}

	// Test difference with empty
	result = polygonPolygonDifference(nil, polys)
	if !result.IsEmpty() {
		t.Error("Empty minus polygon should be empty")
	}

	result = polygonPolygonDifference(polys, nil)
	if result.IsEmpty() {
		t.Error("Polygon minus empty should be polygon")
	}

	// Test symmetric difference with empty
	result = polygonPolygonSymDifference(nil, polys)
	if result.IsEmpty() {
		t.Error("SymDifference with empty should return non-empty polygon")
	}

	result = polygonPolygonSymDifference(polys, nil)
	if result.IsEmpty() {
		t.Error("SymDifference with empty should return non-empty polygon")
	}
}

func TestMultiPolygonOverlay(t *testing.T) {
	// Create two polygons
	shell1 := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := geom.NewLinearRingXY(20, 0, 30, 0, 30, 10, 20, 10, 20, 0)
	poly2 := geom.NewPolygon(shell2, nil)

	polys := []*geom.Polygon{poly1, poly2}

	// Test intersection (should fall back to noded overlay)
	result := polygonPolygonIntersection(polys, polys)
	if result.IsEmpty() {
		t.Error("Intersection of polygon set with itself should not be empty")
	}
}

func TestExtremCoordinates(t *testing.T) {
	// Test with very large coordinates
	shell := geom.NewLinearRingXY(1e100, 1e100, 1e100+10, 1e100, 1e100+10, 1e100+10, 1e100, 1e100+10, 1e100, 1e100)
	poly := geom.NewPolygon(shell, nil)
	polys := []*geom.Polygon{poly}

	// Should bound the coordinates and not crash
	result := polygonPolygonIntersection(polys, polys)
	if result.IsEmpty() {
		t.Error("Self-intersection should not be empty")
	}
}

func TestCollectPolygonsEdgeCases(t *testing.T) {
	// Test with empty slice
	result := collectPolygons(nil)
	if !result.IsEmpty() {
		t.Error("Collecting empty polygons should return empty")
	}

	// Test with single polygon
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)
	result = collectPolygons([]*geom.Polygon{poly})
	if _, ok := result.(*geom.Polygon); !ok {
		t.Errorf("Single polygon should return Polygon, got %T", result)
	}

	// Test with multiple polygons
	shell2 := geom.NewLinearRingXY(20, 0, 30, 0, 30, 10, 20, 10, 20, 0)
	poly2 := geom.NewPolygon(shell2, nil)
	result = collectPolygons([]*geom.Polygon{poly, poly2})
	if _, ok := result.(*geom.MultiPolygon); !ok {
		t.Errorf("Multiple polygons should return MultiPolygon, got %T", result)
	}
}

func TestHandleEmptyPolygons(t *testing.T) {
	// Test handleEmptyPolyA
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
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

func TestBoundCoordinate(t *testing.T) {
	// Test normal coordinate
	c := geom.NewCoordinate(10, 20)
	bounded := boundCoordinate(c)
	if bounded.X != 10 || bounded.Y != 20 {
		t.Error("Normal coordinates should be unchanged")
	}

	// Test NaN
	c = geom.NewCoordinate(0, 0)
	zero := 0.0
	c.X = zero / zero // NaN
	bounded = boundCoordinate(c)
	if bounded.X != 0 {
		t.Error("NaN should be bounded to 0")
	}

	// Test infinity
	c.X = 1e308
	c.Y = -1e308
	bounded = boundCoordinate(c)
	if bounded.X >= 1e200 || bounded.Y <= -1e200 {
		t.Error("Extreme values should be bounded")
	}
}

func TestClipPolygonDifferenceEdgeCases(t *testing.T) {
	// Test with non-intersecting polygons
	shell1 := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := geom.NewLinearRingXY(100, 100, 110, 100, 110, 110, 100, 110, 100, 100)
	poly2 := geom.NewPolygon(shell2, nil)

	result := clipPolygonDifference(poly1, poly2)
	if len(result) != 1 {
		t.Error("Non-intersecting difference should return original polygon")
	}

	// Test with empty polygon
	empty := geom.NewPolygonEmpty()
	result = clipPolygonDifference(empty, poly2)
	if len(result) != 0 {
		t.Error("Empty difference should return nil")
	}

	result = clipPolygonDifference(poly1, empty)
	if len(result) != 1 {
		t.Error("Polygon minus empty should return polygon")
	}
}

func TestParameterOnSegment(t *testing.T) {
	a := geom.NewCoordinate(0, 0)
	b := geom.NewCoordinate(10, 0)
	p := geom.NewCoordinate(5, 0)

	// Test midpoint
	param := parameterOnSegment(a, b, p)
	if param < 0.4 || param > 0.6 {
		t.Errorf("Parameter for midpoint should be ~0.5, got %f", param)
	}

	// Test vertical segment
	a = geom.NewCoordinate(0, 0)
	b = geom.NewCoordinate(0, 10)
	p = geom.NewCoordinate(0, 5)
	param = parameterOnSegment(a, b, p)
	if param < 0.4 || param > 0.6 {
		t.Errorf("Parameter for midpoint on vertical should be ~0.5, got %f", param)
	}

	// Test degenerate segment
	a = geom.NewCoordinate(5, 5)
	b = geom.NewCoordinate(5, 5)
	p = geom.NewCoordinate(5, 5)
	param = parameterOnSegment(a, b, p)
	if param != 0 {
		t.Errorf("Parameter for degenerate segment should be 0, got %f", param)
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

func TestMergePolygonsNonOverlapping(t *testing.T) {
	// Test merging non-overlapping polygons
	shell1 := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)

	shell2 := geom.NewLinearRingXY(100, 100, 110, 100, 110, 110, 100, 110, 100, 100)
	poly2 := geom.NewPolygon(shell2, nil)

	result := mergePolygons(poly1, poly2)
	// Should return one of the polygons or nil if merge fails
	if result != nil && result.IsEmpty() {
		t.Error("Merge result should not be empty if non-nil")
	}

	// Test with empty polygon
	empty := geom.NewPolygonEmpty()
	result = mergePolygons(empty, poly2)
	if result != poly2 {
		t.Error("Merging empty with polygon should return polygon")
	}

	result = mergePolygons(poly1, empty)
	if result != poly1 {
		t.Error("Merging polygon with empty should return polygon")
	}
}

func TestClipPolygonToPolygonEdgeCases(t *testing.T) {
	// Test with empty polygons
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)
	empty := geom.NewPolygonEmpty()

	result := clipPolygonToPolygon(empty, poly)
	if !result.IsEmpty() {
		t.Error("Clipping empty should return empty")
	}

	result = clipPolygonToPolygon(poly, empty)
	if !result.IsEmpty() {
		t.Error("Clipping to empty should return empty")
	}

	// Test with non-intersecting polygons
	shell2 := geom.NewLinearRingXY(100, 100, 110, 100, 110, 110, 100, 110, 100, 100)
	poly2 := geom.NewPolygon(shell2, nil)

	result = clipPolygonToPolygon(poly, poly2)
	if !result.IsEmpty() {
		t.Error("Clipping non-intersecting should return empty")
	}
}

func TestLineSegmentIntersect(t *testing.T) {
	// Test intersecting segments
	a1 := geom.NewCoordinate(0, 0)
	a2 := geom.NewCoordinate(10, 10)
	b1 := geom.NewCoordinate(0, 10)
	b2 := geom.NewCoordinate(10, 0)

	result := lineSegmentIntersect(a1, a2, b1, b2)
	if result == nil {
		t.Error("Intersecting segments should have intersection")
	}

	// Test parallel segments
	a1 = geom.NewCoordinate(0, 0)
	a2 = geom.NewCoordinate(10, 0)
	b1 = geom.NewCoordinate(0, 5)
	b2 = geom.NewCoordinate(10, 5)

	result = lineSegmentIntersect(a1, a2, b1, b2)
	if result != nil {
		t.Error("Parallel segments should not have intersection")
	}
}

func TestSutherlandHodgmanClipEdgeCases(t *testing.T) {
	// Test with less than 3 coordinates
	subject := geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
	}
	clip := geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
	}

	result := sutherlandHodgmanClip(subject, clip)
	if result != nil {
		t.Error("Clipping with too few points should return nil")
	}
}

func TestPointInPolygon(t *testing.T) {
	ring := geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	}

	// Test point inside
	p := geom.NewCoordinate(5, 5)
	result := pointInPolygon(p, ring)
	if result <= 0 {
		t.Error("Point inside should return positive")
	}

	// Test point outside
	p = geom.NewCoordinate(20, 20)
	result = pointInPolygon(p, ring)
	if result >= 0 {
		t.Error("Point outside should return negative")
	}

	// Test with degenerate ring
	tinyRing := geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(1, 0),
	}
	result = pointInPolygon(p, tinyRing)
	if result >= 0 {
		t.Error("Degenerate ring should return negative")
	}
}

func TestClipPolygonOutsideEdgeCases(t *testing.T) {
	subject := geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	}

	// Test with degenerate clip polygon
	clip := geom.CoordinateSequence{
		geom.NewCoordinate(5, 5),
		geom.NewCoordinate(6, 5),
	}

	result := clipPolygonOutside(subject, clip)
	if len(result) == 0 {
		t.Error("Clipping with degenerate polygon should return something")
	}
}

func TestLineLineIntersect(t *testing.T) {
	// Test proper intersection
	a1 := geom.NewCoordinate(0, 0)
	a2 := geom.NewCoordinate(10, 10)
	b1 := geom.NewCoordinate(0, 10)
	b2 := geom.NewCoordinate(10, 0)

	result := lineLineIntersect(a1, a2, b1, b2)
	if result == nil {
		t.Error("Should have intersection")
	}

	// Test parallel lines
	a1 = geom.NewCoordinate(0, 0)
	a2 = geom.NewCoordinate(10, 0)
	b1 = geom.NewCoordinate(0, 5)
	b2 = geom.NewCoordinate(10, 5)

	result = lineLineIntersect(a1, a2, b1, b2)
	if result != nil {
		t.Error("Parallel lines should not intersect")
	}

	// Test non-intersecting segments
	a1 = geom.NewCoordinate(0, 0)
	a2 = geom.NewCoordinate(1, 0)
	b1 = geom.NewCoordinate(10, 0)
	b2 = geom.NewCoordinate(11, 0)

	result = lineLineIntersect(a1, a2, b1, b2)
	if result != nil {
		t.Error("Non-overlapping collinear segments should not intersect")
	}
}

func TestSortPointsByAngle(t *testing.T) {
	// Test with less than 3 points
	coords := geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(1, 0),
	}
	result := sortPointsByAngle(coords)
	if len(result) != 2 {
		t.Error("Should return same coordinates for <3 points")
	}

	// Test with 3+ points
	coords = geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
	}
	result = sortPointsByAngle(coords)
	if len(result) != 4 {
		t.Error("Should return same number of coordinates")
	}
}

func TestIsLeft(t *testing.T) {
	a := geom.NewCoordinate(0, 0)
	b := geom.NewCoordinate(10, 0)

	// Point to the left
	p := geom.NewCoordinate(5, 5)
	result := isLeft(a, b, p)
	if result <= 0 {
		t.Error("Point above line should be left")
	}

	// Point to the right
	p = geom.NewCoordinate(5, -5)
	result = isLeft(a, b, p)
	if result >= 0 {
		t.Error("Point below line should be right")
	}

	// Point on line
	p = geom.NewCoordinate(5, 0)
	result = isLeft(a, b, p)
	if result != 0 {
		t.Error("Point on line should be 0")
	}
}

func TestSegmentSegmentIntersect(t *testing.T) {
	// Test basic intersection
	a1 := geom.NewCoordinate(0, 0)
	a2 := geom.NewCoordinate(10, 10)
	b1 := geom.NewCoordinate(0, 10)
	b2 := geom.NewCoordinate(10, 0)

	result := segmentSegmentIntersect(a1, a2, b1, b2)
	if result == nil {
		t.Error("Intersecting segments should have result")
	}
}

func TestBoundPolygonWithHoles(t *testing.T) {
	// Test polygon with holes
	ext := geom.NewLinearRingXY(0, 0, 1e200, 0, 1e200, 1e200, 0, 1e200, 0, 0)
	hole := geom.NewLinearRingXY(10, 10, 20, 10, 20, 20, 10, 20, 10, 10)
	poly := geom.NewPolygon(ext, []*geom.LinearRing{hole})

	result := boundPolygon(poly)
	if result.IsEmpty() {
		t.Error("Bounding should not make polygon empty")
	}

	// Check that extreme coordinates were bounded
	coords := result.ExteriorRing().Coordinates()
	for _, c := range coords {
		if c.X > 1e150 || c.Y > 1e150 {
			t.Error("Coordinates should be bounded")
		}
	}
}

func TestTraceUnionBoundary(t *testing.T) {
	// Test with overlapping polygons
	shellA := geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	}
	shellB := geom.CoordinateSequence{
		geom.NewCoordinate(5, 5),
		geom.NewCoordinate(15, 5),
		geom.NewCoordinate(15, 15),
		geom.NewCoordinate(5, 15),
		geom.NewCoordinate(5, 5),
	}

	result := traceUnionBoundary(shellA, shellB)
	if len(result) == 0 {
		t.Error("Union boundary should not be empty")
	}

	// Test with contained polygon
	shellC := geom.CoordinateSequence{
		geom.NewCoordinate(2, 2),
		geom.NewCoordinate(8, 2),
		geom.NewCoordinate(8, 8),
		geom.NewCoordinate(2, 8),
		geom.NewCoordinate(2, 2),
	}

	result = traceUnionBoundary(shellA, shellC)
	if len(result) == 0 {
		t.Error("Union boundary with contained should not be empty")
	}
}

func TestNodingEdgeCases(t *testing.T) {
	// Test extracting segment strings from polygons
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly := geom.NewPolygon(shell, nil)

	segStrings := extractSegmentStringsFromPolygons([]*geom.Polygon{poly}, 0)
	if len(segStrings) != 4 {
		t.Errorf("Should extract 4 edges from square, got %d", len(segStrings))
	}

	// Test with polygon with hole
	hole := geom.NewLinearRingXY(2, 2, 8, 2, 8, 8, 2, 8, 2, 2)
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
	shell1 := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	poly1 := geom.NewPolygon(shell1, nil)
	shell2 := geom.NewLinearRingXY(5, 5, 15, 5, 15, 15, 5, 15, 5, 5)
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
