package noding_test

import (
	"fmt"

	"github.com/go-topology-suite/gts/geom"
	"github.com/go-topology-suite/gts/noding"
)

// ExampleSimpleNoder demonstrates basic usage of the noding package
// to find and split segments at their intersection points.
func ExampleSimpleNoder() {
	// Create two crossing line segments
	// Line 1: diagonal from bottom-left to top-right
	// Line 2: diagonal from top-left to bottom-right
	line1 := noding.NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 0, 10, 10),
		"line1",
	)
	line2 := noding.NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 10, 10, 0),
		"line2",
	)

	// Create a noder with an intersection adder
	intersectionAdder := noding.NewIntersectionAdder()
	simpleNoder := noding.NewSimpleNoder(intersectionAdder)

	// Compute nodes (find all intersections)
	simpleNoder.ComputeNodes([]*noding.NodedSegmentString{line1, line2})

	// Get the noded result
	nodedStrings := simpleNoder.GetNodedSubstrings()

	fmt.Printf("Found %d intersections\n", intersectionAdder.ProperIntersectionCount())
	fmt.Printf("Generated %d noded segment strings\n", len(nodedStrings))

	// Each segment string has been split at the intersection point
	for i, nss := range nodedStrings {
		coords := nss.Coordinates()
		fmt.Printf("Noded string %d has %d coordinates\n", i+1, len(coords))
	}

	// Output:
	// Found 1 intersections
	// Generated 2 noded segment strings
	// Noded string 1 has 3 coordinates
	// Noded string 2 has 3 coordinates
}

// ExampleIntersectionCounter demonstrates counting intersections
// without modifying the segment strings.
func ExampleIntersectionCounter() {
	// Create a grid of horizontal and vertical lines
	var segments []*noding.NodedSegmentString

	// Add 3 horizontal lines
	for y := 0; y < 3; y++ {
		seg := noding.NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(0, float64(y), 2, float64(y)),
			nil,
		)
		segments = append(segments, seg)
	}

	// Add 3 vertical lines
	for x := 0; x < 3; x++ {
		seg := noding.NewNodedSegmentString(
			geom.NewCoordinateSequenceXY(float64(x), 0, float64(x), 2),
			nil,
		)
		segments = append(segments, seg)
	}

	// Count intersections without modifying the segments
	counter := noding.NewIntersectionCounter()
	simpleNoder := noding.NewSimpleNoder(counter)
	simpleNoder.ComputeNodes(segments)

	fmt.Printf("Total intersections: %d\n", counter.Count())
	fmt.Printf("Total tests performed: %d\n", counter.NumTests())

	// Output:
	// Total intersections: 9
	// Total tests performed: 15
}

// ExampleNodedSegmentString demonstrates how nodes are added to segments.
func ExampleNodedSegmentString() {
	// Create a segment from (0,0) to (10,0)
	coords := geom.NewCoordinateSequenceXY(0, 0, 10, 0)
	nss := noding.NewNodedSegmentString(coords, nil)

	fmt.Printf("Original coordinates: %d\n", len(nss.Coordinates()))

	// Manually add nodes (in practice, these would be found by intersection)
	nss.AddNode(noding.NewSegmentNode(geom.NewCoordinate(3, 0), 0, 0.3))
	nss.AddNode(noding.NewSegmentNode(geom.NewCoordinate(7, 0), 0, 0.7))

	// Get the noded coordinates (with nodes inserted)
	nodedCoords := nss.NodedCoordinates()

	fmt.Printf("Noded coordinates: %d\n", len(nodedCoords))
	fmt.Printf("Nodes added: %d\n", len(nss.Nodes()))

	// Print all coordinates
	for i, coord := range nodedCoords {
		fmt.Printf("  [%d] (%.0f, %.0f)\n", i, coord.X, coord.Y)
	}

	// Output:
	// Original coordinates: 2
	// Noded coordinates: 4
	// Nodes added: 2
	//   [0] (0, 0)
	//   [1] (3, 0)
	//   [2] (7, 0)
	//   [3] (10, 0)
}

// ExampleSegmentString_IsClosed demonstrates checking if a segment string is closed.
func ExampleSegmentString_IsClosed() {
	// Open segment string
	open := noding.NewSegmentString(
		geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10),
		nil,
	)
	fmt.Printf("Open segment is closed: %v\n", open.IsClosed())

	// Closed segment string (ring)
	closed := noding.NewSegmentString(
		geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 0),
		nil,
	)
	fmt.Printf("Closed segment is closed: %v\n", closed.IsClosed())

	// Output:
	// Open segment is closed: false
	// Closed segment is closed: true
}

// ExampleFindSegmentForCoordinate demonstrates finding which segment contains a point.
func ExampleFindSegmentForCoordinate() {
	// Create a path with 3 segments
	coords := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 10)
	ss := noding.NewSegmentString(coords, nil)

	// Find which segment contains the midpoint of the second segment
	point := geom.NewCoordinate(10, 5)
	segIndex, param, found := noding.FindSegmentForCoordinate(ss, point, geom.DefaultEpsilon)

	if found {
		fmt.Printf("Point found on segment %d at parameter %.2f\n", segIndex, param)
	} else {
		fmt.Println("Point not found on any segment")
	}

	// Output:
	// Point found on segment 1 at parameter 0.50
}

// ExampleIntersectionAdder demonstrates finding proper intersections.
func ExampleIntersectionAdder() {
	// Create a triangle
	triangle := noding.NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 0, 10, 0, 5, 8, 0, 0),
		"triangle",
	)

	// Create a line that crosses the triangle
	line := noding.NewNodedSegmentString(
		geom.NewCoordinateSequenceXY(0, 4, 10, 4),
		"line",
	)

	// Find intersections
	adder := noding.NewIntersectionAdder()
	noder := noding.NewSimpleNoder(adder)
	noder.ComputeNodes([]*noding.NodedSegmentString{triangle, line})

	fmt.Printf("Has intersections: %v\n", adder.HasIntersection())
	fmt.Printf("Has proper intersections: %v\n", adder.HasProperIntersection())
	fmt.Printf("Proper intersection count: %d\n", adder.ProperIntersectionCount())
	fmt.Printf("Triangle nodes added: %d\n", len(triangle.Nodes()))
	fmt.Printf("Line nodes added: %d\n", len(line.Nodes()))

	// Output:
	// Has intersections: true
	// Has proper intersections: true
	// Proper intersection count: 2
	// Triangle nodes added: 2
	// Line nodes added: 2
}
