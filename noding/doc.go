// Package noding provides types and functions for computing nodes (intersection points)
// in collections of line segments.
//
// # Overview
//
// Noding is a critical component for robust overlay operations in computational geometry.
// It ensures that all intersection points between line segments are found and that
// segments are properly split at those points to maintain topological consistency.
//
// The noding process takes a collection of SegmentStrings and produces a new collection
// where all intersections have been computed and the segments have been split at those
// intersection points.
//
// # Example Usage
//
//	// Create two crossing line segments
//	ss1 := noding.NewNodedSegmentString(
//	    geom.NewCoordinateSequenceXY(0, 0, 10, 10),
//	    "line1",
//	)
//	ss2 := noding.NewNodedSegmentString(
//	    geom.NewCoordinateSequenceXY(0, 10, 10, 0),
//	    "line2",
//	)
//
//	// Compute nodes (intersections)
//	noder := noding.NewSimpleNoder(noding.NewIntersectionAdder())
//	noder.ComputeNodes([]*noding.NodedSegmentString{ss1, ss2})
//
//	// Get the noded result (segments split at intersection points)
//	nodedStrings := noder.GetNodedSubstrings()
//
//	// Each noded segment string now has the intersection point inserted
//	for _, nss := range nodedStrings {
//	    coords := nss.Coordinates()
//	    // coords will be: [start, intersection, end]
//	}
//
// # Key Concepts
//
// SegmentString: A sequence of line segments represented by a coordinate sequence.
// It can be split at intersection points.
//
// NodedSegmentString: A SegmentString that tracks nodes (intersection points)
// that have been added to it.
//
// Noder: An algorithm that finds all intersections between segment strings and
// splits them at those points.
//
// SegmentIntersector: Processes intersections as they are found. Different
// implementations can count intersections, record them, or add them as nodes.
//
// # Noder Implementations
//
// SimpleNoder: A basic O(n²) implementation that uses brute-force comparison.
// Suitable for small datasets.
//
// ScaledNoder: Wraps another noder and applies coordinate scaling for improved
// numerical robustness.
//
// ValidatingNoder: Wraps another noder and validates that the result is
// properly noded (no intersections remain).
//
// IteratedNoder: Runs a noder multiple times until no more intersections
// are found, useful for handling numerical robustness issues.
//
// # SegmentIntersector Implementations
//
// IntersectionAdder: Finds all intersections and adds them as nodes to the
// segment strings.
//
// IntersectionCounter: Simply counts the number of intersections without
// modifying the segment strings.
//
// IntersectionFinderAdder: Finds and records interior intersections while
// also adding them as nodes.
package noding
