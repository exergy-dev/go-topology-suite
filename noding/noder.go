// Package noding provides types and functions for computing nodes (intersection points)
// in collections of line segments. Noding is a critical component for robust overlay
// operations, ensuring that all intersection points are found and segments are split
// at those points to maintain topological consistency.
//
// The noding process takes a collection of SegmentStrings and produces a new collection
// where all intersections have been computed and the segments have been split at those
// intersection points.
package noding

import (
	"github.com/robert-malhotra/go-topology-suite/geom"
)

// SegmentString represents a sequence of line segments that can be noded.
// A SegmentString is a mutable data structure that tracks coordinates and
// can be split at intersection points.
type SegmentString struct {
	// coords is the sequence of coordinates defining the segments
	coords geom.CoordinateSequence

	// context is arbitrary data associated with this segment string
	context interface{}
}

// NewSegmentString creates a new SegmentString from a coordinate sequence.
func NewSegmentString(coords geom.CoordinateSequence, context interface{}) *SegmentString {
	return &SegmentString{
		coords:  coords,
		context: context,
	}
}

// Coordinates returns the coordinate sequence of this segment string.
func (ss *SegmentString) Coordinates() geom.CoordinateSequence {
	return ss.coords
}

// SetCoordinates sets the coordinate sequence of this segment string.
func (ss *SegmentString) SetCoordinates(coords geom.CoordinateSequence) {
	ss.coords = coords
}

// Context returns the context data associated with this segment string.
func (ss *SegmentString) Context() interface{} {
	return ss.context
}

// SetContext sets the context data for this segment string.
func (ss *SegmentString) SetContext(context interface{}) {
	ss.context = context
}

// Size returns the number of segments in this segment string.
// The number of segments is one less than the number of coordinates.
func (ss *SegmentString) Size() int {
	return len(ss.coords) - 1
}

// GetCoordinate returns the coordinate at the given index.
func (ss *SegmentString) GetCoordinate(index int) geom.Coordinate {
	return ss.coords[index]
}

// IsClosed returns true if the segment string is closed (forms a ring).
func (ss *SegmentString) IsClosed() bool {
	if len(ss.coords) < 2 {
		return false
	}
	return ss.coords[0].Equals2D(ss.coords[len(ss.coords)-1], geom.DefaultEpsilon)
}

// NodedSegmentString represents a segment string that has been noded.
// It tracks the intersection points (nodes) that have been added.
type NodedSegmentString struct {
	*SegmentString
	nodes []SegmentNode
}

// NewNodedSegmentString creates a new NodedSegmentString.
func NewNodedSegmentString(coords geom.CoordinateSequence, context interface{}) *NodedSegmentString {
	return &NodedSegmentString{
		SegmentString: NewSegmentString(coords, context),
		nodes:         make([]SegmentNode, 0),
	}
}

// AddNode adds a node (intersection point) to this segment string.
func (nss *NodedSegmentString) AddNode(node SegmentNode) {
	nss.nodes = append(nss.nodes, node)
}

// Nodes returns all nodes that have been added to this segment string.
func (nss *NodedSegmentString) Nodes() []SegmentNode {
	return nss.nodes
}

// NodedCoordinates returns a new coordinate sequence with all nodes inserted.
// This splits the original segments at all intersection points.
func (nss *NodedSegmentString) NodedCoordinates() geom.CoordinateSequence {
	if len(nss.nodes) == 0 {
		return nss.coords
	}

	// Build a map of segment index to nodes on that segment
	segmentNodes := make(map[int][]SegmentNode)
	for _, node := range nss.nodes {
		segmentNodes[node.SegmentIndex] = append(segmentNodes[node.SegmentIndex], node)
	}

	// Build the result coordinate sequence
	result := make(geom.CoordinateSequence, 0, len(nss.coords)+len(nss.nodes))

	for i := 0; i < len(nss.coords); i++ {
		result = append(result, nss.coords[i])

		// If there are nodes on the segment starting at this coordinate,
		// add them in order
		if nodes, ok := segmentNodes[i]; ok {
			// Sort nodes by parameter t (position along segment)
			sortNodesByParameter(nodes)

			for _, node := range nodes {
				// Only add if not duplicate of segment endpoint
				if !node.Coord.Equals2D(nss.coords[i], geom.DefaultEpsilon) &&
					!node.Coord.Equals2D(nss.coords[i+1], geom.DefaultEpsilon) {
					result = append(result, node.Coord)
				}
			}
		}
	}

	return result
}

// sortNodesByParameter sorts nodes by their parameter value along the segment.
func sortNodesByParameter(nodes []SegmentNode) {
	// Simple bubble sort - fine for small number of nodes per segment
	n := len(nodes)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if nodes[j].Parameter > nodes[j+1].Parameter {
				nodes[j], nodes[j+1] = nodes[j+1], nodes[j]
			}
		}
	}
}

// SegmentNode represents a node (intersection point) on a segment.
type SegmentNode struct {
	// Coord is the coordinate of the intersection point
	Coord geom.Coordinate

	// SegmentIndex is the index of the segment in the SegmentString (0-based)
	SegmentIndex int

	// Parameter is the position along the segment [0,1] where the intersection occurs
	Parameter float64
}

// NewSegmentNode creates a new SegmentNode.
func NewSegmentNode(coord geom.Coordinate, segmentIndex int, parameter float64) SegmentNode {
	return SegmentNode{
		Coord:        coord,
		SegmentIndex: segmentIndex,
		Parameter:    parameter,
	}
}

// Noder is the interface for algorithms that compute nodes (intersection points)
// in collections of segment strings.
//
// The noding process finds all intersection points between segments and can
// optionally split the segments at those points.
type Noder interface {
	// ComputeNodes computes all nodes (intersections) for the given segment strings.
	// This modifies the segment strings in place by adding nodes.
	ComputeNodes(segmentStrings []*NodedSegmentString)

	// GetNodedSubstrings returns the noded segment strings after computing nodes.
	// Each returned segment string has been split at all intersection points.
	GetNodedSubstrings() []*NodedSegmentString
}

// SegmentNodeIndex represents the index position of a node on a specific segment.
type SegmentNodeIndex struct {
	SegmentString *NodedSegmentString
	SegmentIndex  int
	Parameter     float64
}

// ComputeSegmentIntersectionParameter computes the parameter (0 to 1) for where
// a coordinate lies on a segment defined by two endpoints.
// Returns 0 if at the start, 1 if at the end, or a value in between.
func ComputeSegmentIntersectionParameter(p0, p1, intersection geom.Coordinate) float64 {
	dx := p1.X - p0.X
	dy := p1.Y - p0.Y

	lenSq := dx*dx + dy*dy
	if lenSq < geom.DefaultEpsilon {
		return 0.0
	}

	// Compute parameter using the axis with larger extent for better numerical stability
	if abs(dx) > abs(dy) {
		return (intersection.X - p0.X) / dx
	}
	return (intersection.Y - p0.Y) / dy
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// FindSegmentForCoordinate finds which segment in a SegmentString contains
// or is closest to a given coordinate. Returns the segment index and parameter.
func FindSegmentForCoordinate(ss *SegmentString, coord geom.Coordinate, tolerance float64) (int, float64, bool) {
	for i := 0; i < ss.Size(); i++ {
		p0 := ss.GetCoordinate(i)
		p1 := ss.GetCoordinate(i + 1)

		// Check if coordinate is at segment endpoints
		if coord.Equals2D(p0, tolerance) {
			return i, 0.0, true
		}
		if coord.Equals2D(p1, tolerance) {
			return i, 1.0, true
		}

		// Check if coordinate lies on the segment
		param := ComputeSegmentIntersectionParameter(p0, p1, coord)
		if param >= -tolerance && param <= 1+tolerance {
			// Verify the point actually lies on the segment
			projX := p0.X + param*(p1.X-p0.X)
			projY := p0.Y + param*(p1.Y-p0.Y)
			proj := geom.NewCoordinate(projX, projY)

			if proj.Distance(coord) < tolerance {
				return i, param, true
			}
		}
	}

	return -1, 0.0, false
}
