// Package linemerge provides operations for merging connected LineStrings.
//
// The line merge operation takes a collection of LineStrings and merges them
// into longer LineStrings where they share endpoints. This is useful for:
//   - Simplifying road networks
//   - Combining fragmented line data
//   - Converting polygonized edges back to simplified paths
//
// The algorithm builds a planar graph from the input LineStrings, where nodes
// represent endpoints and edges represent the line segments. It then traverses
// the graph to find maximal sequences of connected lines that can be merged.
//
// Lines are merged when:
//   - They share an endpoint
//   - The shared endpoint has degree 2 (exactly 2 lines meet)
//   - They can be connected end-to-end
//
// Lines are NOT merged when:
//   - They don't share endpoints
//   - A branching point exists (degree > 2)
//   - They would create invalid geometry
//
// Example usage:
//
//	line1 := geom.NewLineStringXY(0, 0, 1, 1)
//	line2 := geom.NewLineStringXY(1, 1, 2, 2)
//
//	merger := NewLineMerger()
//	merger.Add(line1)
//	merger.Add(line2)
//	result := merger.GetMergedLineStrings()
//	// result contains one LineString: (0,0) -> (1,1) -> (2,2)
package linemerge

import (
	"fmt"

	"github.com/go-topology-suite/gts/geom"
)

// LineMerger merges LineStrings that share endpoints.
// It builds a graph of connected line segments and merges sequences
// of LineStrings that can be joined end-to-end.
type LineMerger struct {
	// Input linestrings
	lines []*geom.LineString
	// Graph structure for tracking connections
	graph *lineGraph
	// Result after merging
	merged []*geom.LineString
}

// NewLineMerger creates a new LineMerger.
func NewLineMerger() *LineMerger {
	return &LineMerger{
		graph: newLineGraph(),
	}
}

// Add adds a LineString to be merged.
func (lm *LineMerger) Add(line *geom.LineString) {
	if line == nil || line.IsEmpty() {
		return
	}
	lm.lines = append(lm.lines, line)
}

// AddMultiLineString adds all LineStrings from a MultiLineString.
func (lm *LineMerger) AddMultiLineString(mls *geom.MultiLineString) {
	if mls == nil || mls.IsEmpty() {
		return
	}
	for i := 0; i < mls.NumGeometries(); i++ {
		if line, ok := mls.GeometryN(i).(*geom.LineString); ok {
			lm.Add(line)
		}
	}
}

// AddLineStrings adds multiple LineStrings.
func (lm *LineMerger) AddLineStrings(lines []*geom.LineString) {
	for _, line := range lines {
		lm.Add(line)
	}
}

// GetMergedLineStrings performs the merge and returns the result.
// LineStrings that share endpoints are merged into longer LineStrings.
// LineStrings that form closed loops are returned as closed LineStrings.
// Branching points (where more than 2 lines meet) prevent merging.
func (lm *LineMerger) GetMergedLineStrings() []*geom.LineString {
	if lm.merged != nil {
		return lm.merged
	}

	// Build the graph from input lines
	lm.buildGraph()

	// Merge connected sequences
	lm.merged = lm.mergeSequences()

	return lm.merged
}

// buildGraph constructs the line graph from input LineStrings.
func (lm *LineMerger) buildGraph() {
	for _, line := range lm.lines {
		lm.graph.addEdge(line)
	}
}

// mergeSequences traverses the graph and merges connected LineStrings.
func (lm *LineMerger) mergeSequences() []*geom.LineString {
	var result []*geom.LineString

	// Track which edges have been used
	used := make(map[*edge]bool)

	// Process all edges
	for _, e := range lm.graph.edges {
		if used[e] {
			continue
		}

		// Try to build a sequence from this edge
		sequence := lm.buildSequence(e, used)
		if len(sequence) > 0 {
			merged := lm.mergeEdges(sequence)
			if merged != nil {
				result = append(result, merged)
			}
		}
	}

	return result
}

// buildSequence builds a sequence of connected edges starting from the given edge.
func (lm *LineMerger) buildSequence(start *edge, used map[*edge]bool) []*edge {
	var sequence []*edge
	sequence = append(sequence, start)
	used[start] = true

	current := start

	// Try to extend forward
	for {
		endCoord := current.endCoord
		node := lm.graph.getNode(endCoord)

		if node == nil {
			break
		}

		// Find an unvisited edge connected to this endpoint
		var next *edge
		var originalEdge *edge
		for _, e := range node.edges {
			if used[e] {
				continue
			}

			// Check if this edge connects to current edge
			if e.startCoord.Equals2D(endCoord, geom.DefaultEpsilon) {
				next = e
				originalEdge = e
				break
			} else if e.endCoord.Equals2D(endCoord, geom.DefaultEpsilon) {
				// Need to reverse this edge
				next = e.reversed()
				originalEdge = e
				break
			}
		}

		if next == nil {
			break
		}

		// Check for branching (degree > 2)
		if len(node.edges) > 2 {
			break
		}

		sequence = append(sequence, next)
		used[originalEdge] = true
		current = next
	}

	// Try to extend backward from start
	current = start
	for {
		startCoord := current.startCoord
		node := lm.graph.getNode(startCoord)

		if node == nil {
			break
		}

		// Find an unvisited edge connected to this start point
		var prev *edge
		var originalEdge *edge
		for _, e := range node.edges {
			if used[e] {
				continue
			}

			// Check if this edge connects to current edge
			if e.endCoord.Equals2D(startCoord, geom.DefaultEpsilon) {
				prev = e
				originalEdge = e
				break
			} else if e.startCoord.Equals2D(startCoord, geom.DefaultEpsilon) {
				// Need to reverse this edge
				prev = e.reversed()
				originalEdge = e
				break
			}
		}

		if prev == nil {
			break
		}

		// Check for branching (degree > 2)
		if len(node.edges) > 2 {
			break
		}

		// Prepend to sequence
		sequence = append([]*edge{prev}, sequence...)
		used[originalEdge] = true
		current = prev
	}

	return sequence
}

// mergeEdges combines a sequence of edges into a single LineString.
func (lm *LineMerger) mergeEdges(edges []*edge) *geom.LineString {
	if len(edges) == 0 {
		return nil
	}

	if len(edges) == 1 {
		return edges[0].line
	}

	// Combine coordinates from all edges
	var coords geom.CoordinateSequence

	for i, e := range edges {
		lineCoords := e.line.Coordinates()
		if e.isReversed {
			lineCoords = lineCoords.Reverse()
		}

		if i == 0 {
			// Add all coordinates from first edge
			coords = append(coords, lineCoords...)
		} else {
			// Skip first coordinate of subsequent edges (it duplicates the last coord of previous)
			if len(lineCoords) > 1 {
				coords = append(coords, lineCoords[1:]...)
			}
		}
	}

	return geom.NewLineString(coords)
}

// MergeLineStrings is a convenience function that merges a slice of LineStrings.
func MergeLineStrings(lines []*geom.LineString) []*geom.LineString {
	merger := NewLineMerger()
	merger.AddLineStrings(lines)
	return merger.GetMergedLineStrings()
}

// MergeMultiLineString is a convenience function that merges a MultiLineString.
func MergeMultiLineString(mls *geom.MultiLineString) *geom.MultiLineString {
	merger := NewLineMerger()
	merger.AddMultiLineString(mls)
	merged := merger.GetMergedLineStrings()
	return geom.NewMultiLineString(merged)
}

// lineGraph represents the graph structure of connected LineStrings.
type lineGraph struct {
	// Map from coordinate to node
	nodes map[string]*node
	// All edges in the graph
	edges []*edge
}

func newLineGraph() *lineGraph {
	return &lineGraph{
		nodes: make(map[string]*node),
	}
}

// addEdge adds a LineString as an edge in the graph.
func (g *lineGraph) addEdge(line *geom.LineString) {
	if line.IsEmpty() {
		return
	}

	coords := line.Coordinates()
	startCoord := coords.First()
	endCoord := coords.Last()

	e := &edge{
		line:       line,
		startCoord: startCoord,
		endCoord:   endCoord,
	}

	g.edges = append(g.edges, e)

	// Add to node adjacency lists
	startNode := g.getOrCreateNode(startCoord)
	startNode.addEdge(e)

	// Only add to end node if not a closed ring or if end is different from start
	if !startCoord.Equals2D(endCoord, geom.DefaultEpsilon) {
		endNode := g.getOrCreateNode(endCoord)
		endNode.addEdge(e)
	}
}

// getOrCreateNode gets or creates a node for the given coordinate.
func (g *lineGraph) getOrCreateNode(coord geom.Coordinate) *node {
	key := coordKey(coord)
	n, exists := g.nodes[key]
	if !exists {
		n = &node{
			coord: coord,
		}
		g.nodes[key] = n
	}
	return n
}

// getNode gets a node for the given coordinate, or nil if not found.
func (g *lineGraph) getNode(coord geom.Coordinate) *node {
	key := coordKey(coord)
	return g.nodes[key]
}

// node represents a point where LineStrings connect.
type node struct {
	coord geom.Coordinate
	edges []*edge
}

// addEdge adds an edge to this node.
func (n *node) addEdge(e *edge) {
	n.edges = append(n.edges, e)
}

// degree returns the number of edges connected to this node.
func (n *node) degree() int {
	return len(n.edges)
}

// edge represents a LineString in the graph.
type edge struct {
	line       *geom.LineString
	startCoord geom.Coordinate
	endCoord   geom.Coordinate
	isReversed bool
}

// reversed returns a new edge with reversed direction.
func (e *edge) reversed() *edge {
	return &edge{
		line:       e.line,
		startCoord: e.endCoord,
		endCoord:   e.startCoord,
		isReversed: !e.isReversed,
	}
}

// coordKey creates a string key for a coordinate for use in maps.
func coordKey(c geom.Coordinate) string {
	return fmt.Sprintf("%.10f,%.10f", c.X, c.Y)
}
