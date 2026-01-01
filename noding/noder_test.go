package noding

import (
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSegmentString tests basic SegmentString functionality
func TestSegmentString(t *testing.T) {
	coords := geom.NewCoordinateSequenceXY(0, 0, 10, 10, 20, 0)
	ss := NewSegmentString(coords, "test")

	assert.Equal(t, 2, ss.Size(), "Expected 2 segments")
	assert.Equal(t, "test", ss.Context(), "Expected context 'test'")
	assert.False(t, ss.IsClosed(), "Expected segment string to not be closed")

	// Test closed segment string
	closedCoords := geom.NewCoordinateSequenceXY(0, 0, 10, 10, 20, 0, 0, 0)
	closedSS := NewSegmentString(closedCoords, nil)

	assert.True(t, closedSS.IsClosed(), "Expected segment string to be closed")
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

	require.Len(t, nodedCoords, 3, "Expected 3 coordinates after noding")
	assert.True(t, nodedCoords[1].Equals2D(midpoint, geom.DefaultEpsilon),
		"Expected middle coordinate to be %v, got %v", midpoint, nodedCoords[1])
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

	require.Len(t, nodedCoords, 5, "Expected 5 coordinates")

	// Verify they're in order along the segment
	expectedX := []float64{0, 2, 5, 8, 10}
	for i, expected := range expectedX {
		assert.Equal(t, expected, nodedCoords[i].X, "Coordinate %d: expected X=%f", i, expected)
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
		assert.Equal(t, test.expected, param, "Point %v: expected parameter %f", test.point, test.expected)
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

	assert.True(t, adder.HasIntersection(), "Expected to find intersection")
	assert.True(t, adder.HasProperIntersection(), "Expected proper intersection")
	assert.Equal(t, 1, adder.ProperIntersectionCount(), "Expected 1 proper intersection")

	// Check that nodes were added
	assert.Len(t, ss1.Nodes(), 1, "Expected 1 node on ss1")
	assert.Len(t, ss2.Nodes(), 1, "Expected 1 node on ss2")

	// Verify the intersection point is at (5,5)
	expectedIntersection := geom.NewCoordinate(5, 5)
	assert.True(t, ss1.Nodes()[0].Coord.Equals2D(expectedIntersection, 0.01),
		"Expected intersection at %v, got %v", expectedIntersection, ss1.Nodes()[0].Coord)
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

	assert.False(t, adder.HasIntersection(), "Expected no intersection for parallel lines")
	assert.Empty(t, ss1.Nodes(), "Expected 0 nodes on ss1")
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

	assert.True(t, adder.HasIntersection(), "Expected to find intersections")

	// The line should intersect two sides of the triangle
	assert.GreaterOrEqual(t, adder.ProperIntersectionCount(), 2, "Expected at least 2 intersections")
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

	require.Len(t, nodedSS, 2, "Expected 2 noded segment strings")

	// Each noded segment string should have 3 coordinates
	// (start, intersection, end)
	for i, nss := range nodedSS {
		coords := nss.Coordinates()
		assert.Len(t, coords, 3, "Noded string %d: expected 3 coordinates", i)
	}

	// Verify the middle coordinate is the intersection point
	expectedIntersection := geom.NewCoordinate(5, 5)
	for i, nss := range nodedSS {
		coords := nss.Coordinates()
		assert.True(t, coords[1].Equals2D(expectedIntersection, 0.01),
			"Noded string %d: expected middle point at %v, got %v", i, expectedIntersection, coords[1])
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
	assert.Equal(t, 3, counter.Count(), "Expected 3 intersections")
	assert.NotZero(t, counter.NumTests(), "Expected some intersection tests to be performed")
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

			assert.Equal(t, test.expectedFound, found, "Expected found=%v", test.expectedFound)

			if found {
				assert.Equal(t, test.expectedIndex, index, "Expected index=%d", test.expectedIndex)
				assert.InDelta(t, test.expectedParam, param, 0.01, "Expected param≈%f", test.expectedParam)
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

	assert.True(t, adder.HasIntersection(), "Expected to find intersection in collinear segments")

	// Should have added nodes for the overlap
	assert.NotEmpty(t, ss1.Nodes(), "Expected nodes to be added to ss1")
	assert.NotEmpty(t, ss2.Nodes(), "Expected nodes to be added to ss2")
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

	assert.True(t, adder.HasIntersection(), "Expected to find intersection at endpoint")

	// Endpoint intersections are not proper
	assert.False(t, adder.HasProperIntersection(), "Endpoint intersection should not be proper")
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

	assert.True(t, adder.HasIntersection(), "Expected to find self-intersection")
	assert.True(t, adder.HasProperIntersection(), "Expected proper self-intersection")
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
	assert.False(t, adder.HasIntersection(), "Expected no intersections with empty/degenerate segments")
}

// TestValidatingNoderSuccess tests ValidatingNoder with valid noding
func TestValidatingNoderSuccess(t *testing.T) {
	// Two non-intersecting parallel lines - should validate successfully
	ss1 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 0, 10, 0),
		nil,
	)
	ss2 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 5, 10, 5),
		nil,
	)

	baseNoder := NewSimpleNoder(NewIntersectionAdder())
	vn := NewValidatingNoder(baseNoder)
	vn.ComputeNodes([]*NodedSegmentString{ss1, ss2})
	_ = vn.GetNodedSubstrings()

	assert.NoError(t, vn.ValidationError(), "Expected no validation error")
}

// TestValidatingNoderWithIntersections tests that ValidationError is set when intersections remain
func TestValidatingNoderWithIntersections(t *testing.T) {
	// Two intersecting lines - the base noder will find intersections
	ss1 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 0, 10, 10),
		nil,
	)
	ss2 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 10, 10, 0),
		nil,
	)

	// Use a counter (not adder) so nodes aren't added - this simulates incomplete noding
	baseNoder := NewSimpleNoder(NewIntersectionCounter())
	vn := NewValidatingNoder(baseNoder)
	vn.ComputeNodes([]*NodedSegmentString{ss1, ss2})
	_ = vn.GetNodedSubstrings()

	require.Error(t, vn.ValidationError(), "Expected validation error when intersections remain")

	// Check error type
	nodingErr, ok := vn.ValidationError().(*NodingError)
	require.True(t, ok, "Expected NodingError type")
	assert.NotZero(t, nodingErr.IntersectionCount, "Expected IntersectionCount > 0 in NodingError")
}

// TestValidatingNoderErrorReset tests that ValidationError is reset between calls
func TestValidatingNoderErrorReset(t *testing.T) {
	// First: invalid noding (counter doesn't add nodes)
	ss1 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 0, 10, 10),
		nil,
	)
	ss2 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 10, 10, 0),
		nil,
	)

	baseNoder := NewSimpleNoder(NewIntersectionCounter())
	vn := NewValidatingNoder(baseNoder)
	vn.ComputeNodes([]*NodedSegmentString{ss1, ss2})
	_ = vn.GetNodedSubstrings()

	require.Error(t, vn.ValidationError(), "Expected validation error after first noding")

	// Second: valid noding (parallel lines)
	ss3 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 0, 10, 0),
		nil,
	)
	ss4 := NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 5, 10, 5),
		nil,
	)

	vn.ComputeNodes([]*NodedSegmentString{ss3, ss4})
	_ = vn.GetNodedSubstrings()

	assert.NoError(t, vn.ValidationError(), "Expected no validation error after second noding")
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
