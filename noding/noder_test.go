package noding

import (
	"testing"

	"github.com/go-topology-suite/gts/geom"
)

// TestSegmentString tests basic SegmentString functionality
func TestSegmentString(t *testing.T) {
	coords := geom.NewCoordinateSequenceXY(0, 0, 10, 10, 20, 0)
	ss := NewSegmentString(coords, "test")

	if ss.Size() != 2 {
		t.Errorf("Expected 2 segments, got %d", ss.Size())
	}

	if ss.Context() != "test" {
		t.Errorf("Expected context 'test', got %v", ss.Context())
	}

	if ss.IsClosed() {
		t.Error("Expected segment string to not be closed")
	}

	// Test closed segment string
	closedCoords := geom.NewCoordinateSequenceXY(0, 0, 10, 10, 20, 0, 0, 0)
	closedSS := NewSegmentString(closedCoords, nil)

	if !closedSS.IsClosed() {
		t.Error("Expected segment string to be closed")
	}
}

// TestNodedSegmentString tests NodedSegmentString functionality
func TestNodedSegmentString(t *testing.T) {
	coords := geom.NewCoordinateSequenceXY(0, 0, 10, 10)
	nss := NewNodedSegmentString(coords, nil)

	// Add a node at the midpoint
	midpoint := geom.NewCoordinate(5, 5)
	node := NewSegmentNode(midpoint, 0, 0.5)
	nss.AddNode(node)

	nodedCoords := nss.NodedCoordinates()

	if len(nodedCoords) != 3 {
		t.Errorf("Expected 3 coordinates after noding, got %d", len(nodedCoords))
	}

	if !nodedCoords[1].Equals2D(midpoint, geom.DefaultEpsilon) {
		t.Errorf("Expected middle coordinate to be %v, got %v", midpoint, nodedCoords[1])
	}
}

// TestNodedSegmentStringMultipleNodes tests adding multiple nodes to a segment
func TestNodedSegmentStringMultipleNodes(t *testing.T) {
	coords := geom.NewCoordinateSequenceXY(0, 0, 10, 0)
	nss := NewNodedSegmentString(coords, nil)

	// Add multiple nodes along the segment
	nss.AddNode(NewSegmentNode(geom.NewCoordinate(2, 0), 0, 0.2))
	nss.AddNode(NewSegmentNode(geom.NewCoordinate(8, 0), 0, 0.8))
	nss.AddNode(NewSegmentNode(geom.NewCoordinate(5, 0), 0, 0.5))

	nodedCoords := nss.NodedCoordinates()

	if len(nodedCoords) != 5 {
		t.Errorf("Expected 5 coordinates, got %d", len(nodedCoords))
	}

	// Verify they're in order along the segment
	expectedX := []float64{0, 2, 5, 8, 10}
	for i, expected := range expectedX {
		if nodedCoords[i].X != expected {
			t.Errorf("Coordinate %d: expected X=%f, got %f", i, expected, nodedCoords[i].X)
		}
	}
}

// TestComputeSegmentIntersectionParameter tests the parameter calculation
func TestComputeSegmentIntersectionParameter(t *testing.T) {
	p0 := geom.NewCoordinate(0, 0)
	p1 := geom.NewCoordinate(10, 0)

	tests := []struct {
		point    geom.Coordinate
		expected float64
	}{
		{geom.NewCoordinate(0, 0), 0.0},
		{geom.NewCoordinate(5, 0), 0.5},
		{geom.NewCoordinate(10, 0), 1.0},
		{geom.NewCoordinate(2, 0), 0.2},
		{geom.NewCoordinate(7.5, 0), 0.75},
	}

	for _, test := range tests {
		param := ComputeSegmentIntersectionParameter(p0, p1, test.point)
		if param != test.expected {
			t.Errorf("Point %v: expected parameter %f, got %f",
				test.point, test.expected, param)
		}
	}
}

// TestSimpleNoderTwoIntersectingLines tests finding intersection of two lines
func TestSimpleNoderTwoIntersectingLines(t *testing.T) {
	// Create two crossing line segments
	// Line 1: (0,0) to (10,10)
	// Line 2: (0,10) to (10,0)
	// They intersect at (5,5)

	ss1 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 0, 10, 10),
		"line1",
	)
	ss2 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 10, 10, 0),
		"line2",
	)

	adder := NewIntersectionAdder()
	noder := NewSimpleNoder(adder)
	noder.ComputeNodes([]*NodedSegmentString{ss1, ss2})

	if !adder.HasIntersection() {
		t.Error("Expected to find intersection")
	}

	if !adder.HasProperIntersection() {
		t.Error("Expected proper intersection")
	}

	if adder.ProperIntersectionCount() != 1 {
		t.Errorf("Expected 1 proper intersection, got %d", adder.ProperIntersectionCount())
	}

	// Check that nodes were added
	if len(ss1.Nodes()) != 1 {
		t.Errorf("Expected 1 node on ss1, got %d", len(ss1.Nodes()))
	}

	if len(ss2.Nodes()) != 1 {
		t.Errorf("Expected 1 node on ss2, got %d", len(ss2.Nodes()))
	}

	// Verify the intersection point is at (5,5)
	expectedIntersection := geom.NewCoordinate(5, 5)
	if !ss1.Nodes()[0].Coord.Equals2D(expectedIntersection, 0.01) {
		t.Errorf("Expected intersection at %v, got %v",
			expectedIntersection, ss1.Nodes()[0].Coord)
	}
}

// TestSimpleNoderNoIntersection tests segments that don't intersect
func TestSimpleNoderNoIntersection(t *testing.T) {
	// Two parallel lines
	ss1 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 0, 10, 0),
		"line1",
	)
	ss2 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 5, 10, 5),
		"line2",
	)

	adder := NewIntersectionAdder()
	noder := NewSimpleNoder(adder)
	noder.ComputeNodes([]*NodedSegmentString{ss1, ss2})

	if adder.HasIntersection() {
		t.Error("Expected no intersection for parallel lines")
	}

	if len(ss1.Nodes()) != 0 {
		t.Errorf("Expected 0 nodes on ss1, got %d", len(ss1.Nodes()))
	}
}

// TestSimpleNoderMultipleSegments tests a more complex case
func TestSimpleNoderMultipleSegments(t *testing.T) {
	// Create a triangular path that intersects with a line
	// Triangle: (0,0) -> (10,0) -> (5,10) -> (0,0)
	// Line: (0,5) -> (10,5)

	ss1 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 0, 10, 0, 5, 10, 0, 0),
		"triangle",
	)
	ss2 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 5, 10, 5),
		"line",
	)

	adder := NewIntersectionAdder()
	noder := NewSimpleNoder(adder)
	noder.ComputeNodes([]*NodedSegmentString{ss1, ss2})

	if !adder.HasIntersection() {
		t.Error("Expected to find intersections")
	}

	// The line should intersect two sides of the triangle
	if adder.ProperIntersectionCount() < 2 {
		t.Errorf("Expected at least 2 intersections, got %d",
			adder.ProperIntersectionCount())
	}
}

// TestGetNodedSubstrings tests getting noded substrings
func TestGetNodedSubstrings(t *testing.T) {
	// Two crossing lines
	ss1 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 0, 10, 10),
		"line1",
	)
	ss2 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 10, 10, 0),
		"line2",
	)

	noder := NewSimpleNoder(NewIntersectionAdder())
	noder.ComputeNodes([]*NodedSegmentString{ss1, ss2})

	nodedSS := noder.GetNodedSubstrings()

	if len(nodedSS) != 2 {
		t.Errorf("Expected 2 noded segment strings, got %d", len(nodedSS))
	}

	// Each noded segment string should have 3 coordinates
	// (start, intersection, end)
	for i, nss := range nodedSS {
		coords := nss.Coordinates()
		if len(coords) != 3 {
			t.Errorf("Noded string %d: expected 3 coordinates, got %d", i, len(coords))
		}
	}

	// Verify the middle coordinate is the intersection point
	expectedIntersection := geom.NewCoordinate(5, 5)
	for i, nss := range nodedSS {
		coords := nss.Coordinates()
		if !coords[1].Equals2D(expectedIntersection, 0.01) {
			t.Errorf("Noded string %d: expected middle point at %v, got %v",
				i, expectedIntersection, coords[1])
		}
	}
}

// TestIntersectionCounter tests the IntersectionCounter
func TestIntersectionCounter(t *testing.T) {
	ss1 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 0, 10, 10),
		nil,
	)
	ss2 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 10, 10, 0),
		nil,
	)
	ss3 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(5, 0, 5, 10),
		nil,
	)

	counter := NewIntersectionCounter()
	noder := NewSimpleNoder(counter)
	noder.ComputeNodes([]*NodedSegmentString{ss1, ss2, ss3})

	// ss1 and ss2 intersect, ss1 and ss3 intersect, ss2 and ss3 intersect
	if counter.Count() != 3 {
		t.Errorf("Expected 3 intersections, got %d", counter.Count())
	}

	if counter.NumTests() == 0 {
		t.Error("Expected some intersection tests to be performed")
	}
}

// TestFindSegmentForCoordinate tests finding which segment contains a coordinate
func TestFindSegmentForCoordinate(t *testing.T) {
	coords := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 10)
	ss := NewSegmentString(coords, nil)

	tests := []struct {
		name          string
		coord         geom.Coordinate
		expectedIndex int
		expectedParam float64
		expectedFound bool
	}{
		{
			name:          "Start point",
			coord:         geom.NewCoordinate(0, 0),
			expectedIndex: 0,
			expectedParam: 0.0,
			expectedFound: true,
		},
		{
			name:          "Midpoint of first segment",
			coord:         geom.NewCoordinate(5, 0),
			expectedIndex: 0,
			expectedParam: 0.5,
			expectedFound: true,
		},
		{
			name:          "End point of first segment",
			coord:         geom.NewCoordinate(10, 0),
			expectedIndex: 0,
			expectedParam: 1.0,
			expectedFound: true,
		},
		{
			name:          "Midpoint of second segment",
			coord:         geom.NewCoordinate(10, 5),
			expectedIndex: 1,
			expectedParam: 0.5,
			expectedFound: true,
		},
		{
			name:          "Point not on any segment",
			coord:         geom.NewCoordinate(5, 5),
			expectedIndex: -1,
			expectedFound: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			index, param, found := FindSegmentForCoordinate(ss, test.coord, geom.DefaultEpsilon)

			if found != test.expectedFound {
				t.Errorf("Expected found=%v, got %v", test.expectedFound, found)
			}

			if found {
				if index != test.expectedIndex {
					t.Errorf("Expected index=%d, got %d", test.expectedIndex, index)
				}

				if param < test.expectedParam-0.01 || param > test.expectedParam+0.01 {
					t.Errorf("Expected param≈%f, got %f", test.expectedParam, param)
				}
			}
		})
	}
}

// TestCollinearIntersection tests handling of collinear overlapping segments
func TestCollinearIntersection(t *testing.T) {
	// Two segments on the same line that overlap
	// Segment 1: (0,0) to (10,0)
	// Segment 2: (5,0) to (15,0)
	// They overlap from (5,0) to (10,0)

	ss1 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 0, 10, 0),
		"seg1",
	)
	ss2 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(5, 0, 15, 0),
		"seg2",
	)

	adder := NewIntersectionAdder()
	noder := NewSimpleNoder(adder)
	noder.ComputeNodes([]*NodedSegmentString{ss1, ss2})

	if !adder.HasIntersection() {
		t.Error("Expected to find intersection in collinear segments")
	}

	// Should have added nodes for the overlap
	if len(ss1.Nodes()) == 0 {
		t.Error("Expected nodes to be added to ss1")
	}

	if len(ss2.Nodes()) == 0 {
		t.Error("Expected nodes to be added to ss2")
	}
}

// TestEndpointIntersection tests segments that touch at endpoints
func TestEndpointIntersection(t *testing.T) {
	// Two segments that share an endpoint
	// Segment 1: (0,0) to (5,5)
	// Segment 2: (5,5) to (10,0)

	ss1 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 0, 5, 5),
		"seg1",
	)
	ss2 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(5, 5, 10, 0),
		"seg2",
	)

	adder := NewIntersectionAdder()
	noder := NewSimpleNoder(adder)
	noder.ComputeNodes([]*NodedSegmentString{ss1, ss2})

	if !adder.HasIntersection() {
		t.Error("Expected to find intersection at endpoint")
	}

	// Endpoint intersections are not proper
	if adder.HasProperIntersection() {
		t.Error("Endpoint intersection should not be proper")
	}
}

// TestSelfIntersection tests a segment string that intersects itself
func TestSelfIntersection(t *testing.T) {
	// A figure-8 shape
	// (0,0) -> (10,10) -> (10,0) -> (0,10)
	// The crossing lines intersect at (5,5)

	ss := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 0, 10, 10, 10, 0, 0, 10),
		"figure8",
	)

	adder := NewIntersectionAdder()
	noder := NewSimpleNoder(adder)
	noder.ComputeNodes([]*NodedSegmentString{ss})

	if !adder.HasIntersection() {
		t.Error("Expected to find self-intersection")
	}

	if !adder.HasProperIntersection() {
		t.Error("Expected proper self-intersection")
	}
}

// TestEmptySegmentString tests handling of empty or degenerate segment strings
func TestEmptySegmentString(t *testing.T) {
	// Empty segment string
	ss1 := NewNodedSegmentString(geom.CoordinateSequence{}, nil)

	// Single point (no segments)
	ss2 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(5, 5),
		nil,
	)

	// Normal segment
	ss3 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 0, 10, 0),
		nil,
	)

	adder := NewIntersectionAdder()
	noder := NewSimpleNoder(adder)
	noder.ComputeNodes([]*NodedSegmentString{ss1, ss2, ss3})

	// Should not crash and should find no intersections
	if adder.HasIntersection() {
		t.Error("Expected no intersections with empty/degenerate segments")
	}
}

// BenchmarkSimpleNoder benchmarks the simple noder
func BenchmarkSimpleNoder(b *testing.B) {
	// Create a grid of segments
	segments := make([]*NodedSegmentString, 0)

	// Horizontal lines
	for y := 0; y < 10; y++ {
		ss := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, float64(y), 10, float64(y)),
			nil,
		)
		segments = append(segments, ss)
	}

	// Vertical lines
	for x := 0; x < 10; x++ {
		ss := NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(float64(x), 0, float64(x), 10),
			nil,
		)
		segments = append(segments, ss)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset nodes
		for _, ss := range segments {
			ss.nodes = nil
		}

		adder := NewIntersectionAdder()
		noder := NewSimpleNoder(adder)
		noder.ComputeNodes(segments)
	}
}

// BenchmarkGetNodedSubstrings benchmarks creating noded substrings
func BenchmarkGetNodedSubstrings(b *testing.B) {
	// Create intersecting segments
	ss1 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 0, 10, 10, 20, 0),
		nil,
	)
	ss2 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 10, 10, 0, 20, 10),
		nil,
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset nodes
		ss1.nodes = nil
		ss2.nodes = nil

		noder := NewSimpleNoder(NewIntersectionAdder())
		noder.ComputeNodes([]*NodedSegmentString{ss1, ss2})
		_ = noder.GetNodedSubstrings()
	}
}
