package noding

import (
	"github.com/go-topology-suite/gts/algorithm"
	"github.com/go-topology-suite/gts/geom"
)

// SegmentIntersector is the interface for classes that process intersections
// between segments during noding. Implementations can track statistics,
// record intersection points, or perform other operations when intersections
// are found.
type SegmentIntersector interface {
	// ProcessIntersections is called when two segments may intersect.
	// e0 and e1 are the segment strings, and segIndex0 and segIndex1 are
	// the indices of the segments within their respective segment strings.
	ProcessIntersections(
		e0 *NodedSegmentString, segIndex0 int,
		e1 *NodedSegmentString, segIndex1 int,
	)

	// IsDone returns true if processing is complete and no further
	// intersections need to be computed.
	IsDone() bool
}

// IntersectionAdder is a SegmentIntersector that finds all intersections
// between segments and adds them as nodes to the segment strings.
type IntersectionAdder struct {
	// properIntersectionCount tracks the number of proper intersections found
	properIntersectionCount int

	// hasIntersection is true if any intersection was found
	hasIntersection bool

	// hasProperIntersection is true if any proper intersection was found
	hasProperIntersection bool

	// hasProperInteriorIntersection is true if any proper interior intersection was found
	hasProperInteriorIntersection bool

	// numTests tracks the number of intersection tests performed
	numTests int

	// allowProperInteriorIntersections controls whether proper interior intersections
	// are allowed (true) or treated as errors (false)
	allowProperInteriorIntersections bool
}

// NewIntersectionAdder creates a new IntersectionAdder.
func NewIntersectionAdder() *IntersectionAdder {
	return &IntersectionAdder{
		allowProperInteriorIntersections: true,
	}
}

// ProcessIntersections processes potential intersections between two segments.
func (ia *IntersectionAdder) ProcessIntersections(
	e0 *NodedSegmentString, segIndex0 int,
	e1 *NodedSegmentString, segIndex1 int,
) {
	// Don't test a segment with itself
	if e0 == e1 && segIndex0 == segIndex1 {
		return
	}

	// Don't test adjacent segments in the same segment string
	// They will always share an endpoint, which is not a proper intersection
	if e0 == e1 {
		diff := abs(float64(segIndex0 - segIndex1))
		if diff == 1 {
			return
		}
		// For closed rings, also skip first and last segments
		if e0.IsClosed() && diff == float64(e0.Size()-1) {
			return
		}
	}

	ia.numTests++

	// Get segment coordinates
	p00 := e0.GetCoordinate(segIndex0)
	p01 := e0.GetCoordinate(segIndex0 + 1)
	p10 := e1.GetCoordinate(segIndex1)
	p11 := e1.GetCoordinate(segIndex1 + 1)

	// Compute intersection
	result := algorithm.LineIntersection(p00, p01, p10, p11)

	if !result.HasIntersection {
		return
	}

	ia.hasIntersection = true

	if result.IsProper {
		ia.properIntersectionCount++
		ia.hasProperIntersection = true

		// Check if it's an interior intersection (not at segment endpoints)
		if !isEndpoint(result.Intersection, p00, p01, p10, p11) {
			ia.hasProperInteriorIntersection = true
		}
	}

	// Add the intersection point(s) as nodes
	ia.addIntersection(e0, segIndex0, e1, segIndex1, result.Intersection)

	// If there's a second intersection (collinear overlap), add it too
	if result.IsCollinear && !result.Intersection2.IsNaN() {
		ia.addIntersection(e0, segIndex0, e1, segIndex1, result.Intersection2)
	}
}

// addIntersection adds an intersection point as a node to both segment strings.
func (ia *IntersectionAdder) addIntersection(
	e0 *NodedSegmentString, segIndex0 int,
	e1 *NodedSegmentString, segIndex1 int,
	intPt geom.Coordinate,
) {
	// Add node to first segment string
	param0 := ComputeSegmentIntersectionParameter(
		e0.GetCoordinate(segIndex0),
		e0.GetCoordinate(segIndex0+1),
		intPt,
	)
	node0 := NewSegmentNode(intPt, segIndex0, param0)
	e0.AddNode(node0)

	// Add node to second segment string
	param1 := ComputeSegmentIntersectionParameter(
		e1.GetCoordinate(segIndex1),
		e1.GetCoordinate(segIndex1+1),
		intPt,
	)
	node1 := NewSegmentNode(intPt, segIndex1, param1)
	e1.AddNode(node1)
}

// isEndpoint checks if a coordinate is an endpoint of either segment.
func isEndpoint(coord, p00, p01, p10, p11 geom.Coordinate) bool {
	return coord.Equals2D(p00, geom.DefaultEpsilon) ||
		coord.Equals2D(p01, geom.DefaultEpsilon) ||
		coord.Equals2D(p10, geom.DefaultEpsilon) ||
		coord.Equals2D(p11, geom.DefaultEpsilon)
}

// IsDone returns false, as we always want to find all intersections.
func (ia *IntersectionAdder) IsDone() bool {
	return false
}

// HasIntersection returns true if any intersection was found.
func (ia *IntersectionAdder) HasIntersection() bool {
	return ia.hasIntersection
}

// HasProperIntersection returns true if any proper intersection was found.
// A proper intersection is one where the segments cross in their interiors.
func (ia *IntersectionAdder) HasProperIntersection() bool {
	return ia.hasProperIntersection
}

// HasProperInteriorIntersection returns true if any proper interior intersection was found.
func (ia *IntersectionAdder) HasProperInteriorIntersection() bool {
	return ia.hasProperInteriorIntersection
}

// ProperIntersectionCount returns the number of proper intersections found.
func (ia *IntersectionAdder) ProperIntersectionCount() int {
	return ia.properIntersectionCount
}

// NumTests returns the number of intersection tests performed.
func (ia *IntersectionAdder) NumTests() int {
	return ia.numTests
}

// IntersectionCounter is a simple SegmentIntersector that just counts intersections.
type IntersectionCounter struct {
	count    int
	numTests int
}

// NewIntersectionCounter creates a new IntersectionCounter.
func NewIntersectionCounter() *IntersectionCounter {
	return &IntersectionCounter{}
}

// ProcessIntersections checks if two segments intersect and increments the counter.
func (ic *IntersectionCounter) ProcessIntersections(
	e0 *NodedSegmentString, segIndex0 int,
	e1 *NodedSegmentString, segIndex1 int,
) {
	// Don't test a segment with itself
	if e0 == e1 && segIndex0 == segIndex1 {
		return
	}

	// Don't test adjacent segments in the same segment string
	if e0 == e1 {
		diff := abs(float64(segIndex0 - segIndex1))
		if diff == 1 {
			return
		}
		// For closed rings, also skip first and last segments
		if e0.IsClosed() && diff == float64(e0.Size()-1) {
			return
		}
	}

	ic.numTests++

	// Get segment coordinates
	p00 := e0.GetCoordinate(segIndex0)
	p01 := e0.GetCoordinate(segIndex0 + 1)
	p10 := e1.GetCoordinate(segIndex1)
	p11 := e1.GetCoordinate(segIndex1 + 1)

	// Compute intersection
	result := algorithm.LineIntersection(p00, p01, p10, p11)

	if result.HasIntersection {
		ic.count++
	}
}

// IsDone returns false, as we want to count all intersections.
func (ic *IntersectionCounter) IsDone() bool {
	return false
}

// Count returns the number of intersections found.
func (ic *IntersectionCounter) Count() int {
	return ic.count
}

// NumTests returns the number of intersection tests performed.
func (ic *IntersectionCounter) NumTests() int {
	return ic.numTests
}

// IntersectionFinderAdder is a SegmentIntersector that finds interior intersections
// and adds them to a collection, while also adding them as nodes.
type IntersectionFinderAdder struct {
	*IntersectionAdder
	intersections []geom.Coordinate
}

// NewIntersectionFinderAdder creates a new IntersectionFinderAdder.
func NewIntersectionFinderAdder() *IntersectionFinderAdder {
	return &IntersectionFinderAdder{
		IntersectionAdder: NewIntersectionAdder(),
		intersections:     make([]geom.Coordinate, 0),
	}
}

// ProcessIntersections processes intersections and records them.
func (ifa *IntersectionFinderAdder) ProcessIntersections(
	e0 *NodedSegmentString, segIndex0 int,
	e1 *NodedSegmentString, segIndex1 int,
) {
	// Let the base class do the work
	ifa.IntersectionAdder.ProcessIntersections(e0, segIndex0, e1, segIndex1)

	// Record the intersection if it's proper and interior
	if ifa.hasProperInteriorIntersection {
		// Get segment coordinates
		p00 := e0.GetCoordinate(segIndex0)
		p01 := e0.GetCoordinate(segIndex0 + 1)
		p10 := e1.GetCoordinate(segIndex1)
		p11 := e1.GetCoordinate(segIndex1 + 1)

		// Compute intersection again to get the point
		result := algorithm.LineIntersection(p00, p01, p10, p11)

		if result.HasIntersection && result.IsProper {
			ifa.intersections = append(ifa.intersections, result.Intersection)
		}
	}
}

// Intersections returns all interior intersection points found.
func (ifa *IntersectionFinderAdder) Intersections() []geom.Coordinate {
	return ifa.intersections
}
