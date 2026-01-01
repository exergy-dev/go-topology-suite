// Package polygonize provides functionality for constructing polygons from
// collections of LineStrings that form closed rings.
//
// The polygonization process takes a collection of line segments (edges) and:
// 1. Nodes the edges to find all intersection points
// 2. Builds a planar graph from the noded edges
// 3. Extracts minimal cycles (polygon rings) from the graph
// 4. Classifies rings as shells (exterior) or holes based on orientation
// 5. Assigns holes to their containing shells
// 6. Returns the resulting polygons
//
// This is useful for reconstructing polygons from edge data, such as:
// - Converting topological data structures to polygons
// - Building polygons from boundary representations
// - Processing CAD/GIS data that stores boundaries as separate edges
package polygonize

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/robert-malhotra/go-topology-suite/algorithm"
	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/noding"
)

// Polygonizer constructs polygons from a collection of LineStrings.
type Polygonizer struct {
	// Input line strings (edges)
	lines []*geom.LineString

	// Results
	polygons        []*geom.Polygon
	danglingLines   []*geom.LineString
	cutLines        []*geom.LineString
	invalidRingLines []*geom.LineString

	// Internal state
	graph *EdgeGraph
}

// NewPolygonizer creates a new Polygonizer.
func NewPolygonizer() *Polygonizer {
	return &Polygonizer{
		lines:            make([]*geom.LineString, 0),
		polygons:         make([]*geom.Polygon, 0),
		danglingLines:    make([]*geom.LineString, 0),
		cutLines:         make([]*geom.LineString, 0),
		invalidRingLines: make([]*geom.LineString, 0),
	}
}

// Add adds a LineString to be polygonized.
func (p *Polygonizer) Add(line *geom.LineString) {
	if line != nil && !line.IsEmpty() {
		p.lines = append(p.lines, line)
	}
}

// AddAll adds multiple LineStrings to be polygonized.
func (p *Polygonizer) AddAll(lines []*geom.LineString) {
	for _, line := range lines {
		p.Add(line)
	}
}

// GetPolygons returns the polygons formed by the input line strings.
// This triggers the polygonization if not already performed.
func (p *Polygonizer) GetPolygons() []*geom.Polygon {
	if p.graph == nil {
		p.polygonize()
	}
	return p.polygons
}

// GetDangles returns edges that are connected at only one endpoint.
func (p *Polygonizer) GetDangles() []*geom.LineString {
	if p.graph == nil {
		p.polygonize()
	}
	return p.danglingLines
}

// GetCutEdges returns edges that are connected on both ends but are not part of any minimal cycle.
func (p *Polygonizer) GetCutEdges() []*geom.LineString {
	if p.graph == nil {
		p.polygonize()
	}
	return p.cutLines
}

// GetInvalidRingLines returns edges that form invalid rings (self-intersecting, etc).
func (p *Polygonizer) GetInvalidRingLines() []*geom.LineString {
	if p.graph == nil {
		p.polygonize()
	}
	return p.invalidRingLines
}

// polygonize performs the actual polygonization algorithm.
func (p *Polygonizer) polygonize() {
	if len(p.lines) == 0 {
		return
	}

	// Step 1: Node all the input lines to find intersection points
	nodedLines := p.nodeLines()

	// Step 2: Build the planar graph from noded edges
	p.graph = buildGraph(nodedLines)

	// Step 3: Mark dangling edges before ring finding
	p.markDanglingEdges()

	// Step 4: Find all minimal cycles (rings)
	rings := p.graph.findRings()

	// Step 5: Classify rings and build polygons
	p.buildPolygons(rings)
}

// markDanglingEdges identifies and marks dangling edges before ring finding.
// This prevents dangling edges from interfering with the ring-finding algorithm.
func (p *Polygonizer) markDanglingEdges() {
	if p.graph == nil {
		return
	}

	// Build node degree map by counting unique edges at each node
	// We count the number of edges emanating from each node
	nodeDegree := make(map[string]int)
	seenEdges := make(map[*DirectedEdge]bool)

	for _, edgeList := range p.graph.edges {
		for _, edge := range edgeList {
			// Count each undirected edge only once
			if seenEdges[edge] || seenEdges[edge.Sym] {
				continue
			}
			seenEdges[edge] = true
			if edge.Sym != nil {
				seenEdges[edge.Sym] = true
			}

			startKey := coordToKey(edge.Start)
			endKey := coordToKey(edge.End)
			nodeDegree[startKey]++
			nodeDegree[endKey]++
		}
	}

	// Track which edges have been recorded as dangling to avoid duplicates
	recorded := make(map[*DirectedEdge]bool)

	// Iteratively mark dangling edges (edges with degree-1 endpoints)
	// We iterate because removing a dangling edge may create new dangles
	changed := true
	for changed {
		changed = false
		for _, edgeList := range p.graph.edges {
			for _, edge := range edgeList {
				if edge.Used {
					continue
				}

				startKey := coordToKey(edge.Start)
				endKey := coordToKey(edge.End)

				// Check if either endpoint has degree 1
				if nodeDegree[startKey] == 1 || nodeDegree[endKey] == 1 {
					// Record as dangling before marking used
					if !recorded[edge] && !recorded[edge.Sym] {
						coords := geom.CoordinateSequence{edge.Start, edge.End}
						p.danglingLines = append(p.danglingLines, geom.NewLineString(coords))
						recorded[edge] = true
						if edge.Sym != nil {
							recorded[edge.Sym] = true
						}
					}

					// Mark edge and its symmetric as used (dangling)
					edge.Used = true
					if edge.Sym != nil {
						edge.Sym.Used = true
					}

					// Decrement degree at both endpoints
					nodeDegree[startKey]--
					nodeDegree[endKey]--

					changed = true
				}
			}
		}
	}
}

// nodeLines nodes all input lines to split them at intersection points.
func (p *Polygonizer) nodeLines() []*geom.LineString {
	// Convert LineStrings to NodedSegmentStrings
	segStrings := make([]*noding.NodedSegmentString, 0, len(p.lines))
	for i, line := range p.lines {
		coords := line.Coordinates()
		// Create one segment string per edge to avoid artifacts
		for j := 0; j < len(coords)-1; j++ {
			edge := geom.CoordinateSequence{coords[j], coords[j+1]}
			ss := noding.NewNodedSegmentString(edge, i)
			segStrings = append(segStrings, ss)
		}
	}

	// Node the segment strings
	noder := noding.NewSimpleNoder(noding.NewIntersectionAdder())
	noder.ComputeNodes(segStrings)
	nodedSegments := noder.GetNodedSubstrings()

	// Convert back to LineStrings
	result := make([]*geom.LineString, 0, len(nodedSegments))
	for _, ss := range nodedSegments {
		coords := ss.Coordinates()
		if len(coords) >= 2 {
			result = append(result, geom.NewLineString(coords))
		}
	}

	return result
}

// buildPolygons classifies rings as shells/holes and builds polygons.
func (p *Polygonizer) buildPolygons(rings []geom.CoordinateSequence) {
	if len(rings) == 0 {
		return
	}

	// Deduplicate rings - remove duplicate rings that are the same geometry
	// (the ring-finding algorithm may find the same ring in both directions)
	uniqueRings := deduplicateRings(rings)

	// Separate shells from holes based on orientation
	var shells []geom.CoordinateSequence
	var holes []geom.CoordinateSequence

	for _, ring := range uniqueRings {
		// Ensure ring is closed
		if !ring.IsClosed(geom.DefaultEpsilon) {
			ring = append(ring, ring[0].Clone())
		}

		// Need at least 4 points for a valid ring
		if len(ring) < 4 {
			continue
		}

		// Classify by signed area (CCW = shell, CW = hole)
		area := geom.SignedArea(ring)
		if math.Abs(area) < geom.DefaultEpsilon {
			// Degenerate ring - skip
			continue
		}

		if area > 0 {
			// Counter-clockwise = exterior ring (shell)
			shells = append(shells, ring)
		} else {
			// Clockwise = interior ring (hole)
			holes = append(holes, ring)
		}
	}

	// If no shells but we have holes, treat the largest hole as a shell
	if len(shells) == 0 && len(holes) > 0 {
		largestIdx := 0
		largestArea := -geom.SignedArea(holes[0])
		for i := 1; i < len(holes); i++ {
			area := -geom.SignedArea(holes[i])
			if area > largestArea {
				largestArea = area
				largestIdx = i
			}
		}
		// Reverse to make it CCW
		shells = append(shells, holes[largestIdx].Reverse())
		holes = append(holes[:largestIdx], holes[largestIdx+1:]...)
	}

	// Assign holes to shells
	for _, shellCoords := range shells {
		shellRing := geom.NewLinearRing(shellCoords)
		shellPoly := geom.NewPolygon(shellRing, nil)

		var assignedHoles []*geom.LinearRing

		for _, holeCoords := range holes {
			// Use an interior point for robust containment test
			holePoint := computeInteriorPoint(holeCoords)
			loc := algorithm.PointLocationInPolygon(holePoint, shellPoly)
			if loc == geom.LocationInterior {
				assignedHoles = append(assignedHoles, geom.NewLinearRing(holeCoords))
			}
		}

		poly := geom.NewPolygon(shellRing, assignedHoles)
		if !poly.IsEmpty() && poly.Area() > geom.DefaultEpsilon {
			p.polygons = append(p.polygons, poly)
		}
	}
}

// coordToKey returns a string key for a coordinate.
func coordToKey(c geom.Coordinate) string {
	return fmt.Sprintf("%.10f,%.10f", c.X, c.Y)
}

// deduplicateRings removes duplicate rings.
// Two rings are duplicates if they have the same coordinates (in same or reverse order).
func deduplicateRings(rings []geom.CoordinateSequence) []geom.CoordinateSequence {
	if len(rings) <= 1 {
		return rings
	}

	var unique []geom.CoordinateSequence
	seen := make(map[string]bool)

	for _, ring := range rings {
		// Create a canonical representation of the ring
		key := ringKey(ring)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, ring)
		}
	}

	return unique
}

// ringKey creates a unique key for a ring that's the same for the ring
// traversed in either direction and starting from any point.
func ringKey(ring geom.CoordinateSequence) string {
	if len(ring) == 0 {
		return ""
	}

	// Normalize the ring: find min coordinate and ensure we traverse in CCW direction
	minIdx := 0
	for i := 1; i < len(ring); i++ {
		if ring[i].X < ring[minIdx].X ||
			(ring[i].X == ring[minIdx].X && ring[i].Y < ring[minIdx].Y) {
			minIdx = i
		}
	}

	// Check if ring is CCW or CW
	area := geom.SignedArea(ring)

	// Build key starting from min coordinate
	var coords []string
	n := len(ring)
	if ring.IsClosed(geom.DefaultEpsilon) {
		n-- // Don't include closing point in key
	}

	if area >= 0 {
		// CCW - traverse forward from min
		for i := 0; i < n; i++ {
			idx := (minIdx + i) % n
			coords = append(coords, fmt.Sprintf("%.10f,%.10f", ring[idx].X, ring[idx].Y))
		}
	} else {
		// CW - traverse backward from min to make it CCW
		for i := 0; i < n; i++ {
			idx := (minIdx - i + n) % n
			coords = append(coords, fmt.Sprintf("%.10f,%.10f", ring[idx].X, ring[idx].Y))
		}
	}

	return strings.Join(coords, ";")
}

// computeInteriorPoint computes a point guaranteed to be inside the ring.
func computeInteriorPoint(ring geom.CoordinateSequence) geom.Coordinate {
	if len(ring) < 3 {
		return ring[0]
	}

	// Use centroid
	var cx, cy float64
	n := len(ring)
	if ring.IsClosed(geom.DefaultEpsilon) {
		n-- // Exclude closing point
	}

	for i := 0; i < n; i++ {
		cx += ring[i].X
		cy += ring[i].Y
	}

	if n > 0 {
		cx /= float64(n)
		cy /= float64(n)
	}

	return geom.NewCoordinate(cx, cy)
}

// EdgeGraph represents a planar graph of edges for polygonization.
type EdgeGraph struct {
	// Map from coordinate to list of edges starting at that coordinate
	edges map[geom.Coordinate][]*DirectedEdge
}

// DirectedEdge represents an edge in a specific direction.
type DirectedEdge struct {
	Start geom.Coordinate
	End   geom.Coordinate
	Used  bool
	Sym   *DirectedEdge // Symmetric edge (reverse direction)
}

// newEdgeGraph creates a new EdgeGraph.
func newEdgeGraph() *EdgeGraph {
	return &EdgeGraph{
		edges: make(map[geom.Coordinate][]*DirectedEdge),
	}
}

// addEdge adds a directed edge to the graph.
func (g *EdgeGraph) addEdge(start, end geom.Coordinate) {
	// Create forward edge
	fwd := &DirectedEdge{Start: start, End: end}

	// Create reverse edge
	rev := &DirectedEdge{Start: end, End: start}

	// Link them as symmetric
	fwd.Sym = rev
	rev.Sym = fwd

	// Add to adjacency map (using fuzzy lookup)
	g.addEdgeToMap(start, fwd)
	g.addEdgeToMap(end, rev)
}

// addEdgeToMap adds an edge to the adjacency map with fuzzy coordinate matching.
func (g *EdgeGraph) addEdgeToMap(coord geom.Coordinate, edge *DirectedEdge) {
	// Try to find existing coordinate that matches
	for existingCoord := range g.edges {
		if coord.Equals2D(existingCoord, geom.DefaultEpsilon) {
			g.edges[existingCoord] = append(g.edges[existingCoord], edge)
			return
		}
	}
	// No match found, add new coordinate
	g.edges[coord] = []*DirectedEdge{edge}
}

// getEdges returns all edges starting at a coordinate (with fuzzy matching).
func (g *EdgeGraph) getEdges(coord geom.Coordinate) []*DirectedEdge {
	// Try exact match first
	if edges, ok := g.edges[coord]; ok {
		return edges
	}

	// Try fuzzy match
	for existingCoord, edges := range g.edges {
		if coord.Equals2D(existingCoord, geom.DefaultEpsilon) {
			return edges
		}
	}

	return nil
}

// findRings finds all minimal cycles in the graph using the "rightmost turn" algorithm.
func (g *EdgeGraph) findRings() []geom.CoordinateSequence {
	var rings []geom.CoordinateSequence

	// Collect all edges into a sorted slice for deterministic iteration
	var allEdges []*DirectedEdge
	for _, edgeList := range g.edges {
		allEdges = append(allEdges, edgeList...)
	}

	// Sort edges by start coordinate then end coordinate for deterministic order
	sortEdges(allEdges)

	// Try to build a ring from each unused edge
	for _, startEdge := range allEdges {
		if startEdge.Used {
			continue
		}

		ring := g.buildRing(startEdge)
		if ring != nil && len(ring) >= 4 {
			rings = append(rings, ring)
		}
	}

	return rings
}

// sortEdges sorts edges by their start and end coordinates for deterministic processing.
func sortEdges(edges []*DirectedEdge) {
	sort.Slice(edges, func(i, j int) bool {
		// Compare by start coordinate first
		if edges[i].Start.X != edges[j].Start.X {
			return edges[i].Start.X < edges[j].Start.X
		}
		if edges[i].Start.Y != edges[j].Start.Y {
			return edges[i].Start.Y < edges[j].Start.Y
		}
		// Then by end coordinate
		if edges[i].End.X != edges[j].End.X {
			return edges[i].End.X < edges[j].End.X
		}
		return edges[i].End.Y < edges[j].End.Y
	})
}

// buildRing builds a ring starting from the given edge using rightmost turn.
func (g *EdgeGraph) buildRing(startEdge *DirectedEdge) geom.CoordinateSequence {
	if startEdge.Used {
		return nil
	}

	var ring geom.CoordinateSequence
	ring = append(ring, startEdge.Start)

	current := startEdge
	visited := make(map[*DirectedEdge]bool)
	maxSteps := 10000 // Prevent infinite loops

	for steps := 0; steps < maxSteps; steps++ {
		// Mark edge as used for this ring attempt
		visited[current] = true

		// Add endpoint
		ring = append(ring, current.End)

		// Check if we've closed the ring
		if current.End.Equals2D(startEdge.Start, geom.DefaultEpsilon) && len(ring) >= 4 {
			// Successfully closed ring - mark all edges as permanently used
			for edge := range visited {
				edge.Used = true
			}
			return ring
		}

		// Find next edge using rightmost turn
		nextEdge := g.findNextEdgeRightmost(current)
		if nextEdge == nil || visited[nextEdge] {
			// Can't continue - this didn't form a valid ring
			return nil
		}

		current = nextEdge
	}

	// Couldn't close the ring
	return nil
}

// findNextEdgeRightmost finds the next edge using the rightmost turn rule.
func (g *EdgeGraph) findNextEdgeRightmost(incoming *DirectedEdge) *DirectedEdge {
	// Get all edges starting at the endpoint
	candidates := g.getEdges(incoming.End)
	if len(candidates) == 0 {
		return nil
	}

	// Filter out used edges and the symmetric edge (backtrack)
	var available []*DirectedEdge
	for _, edge := range candidates {
		if !edge.Used && edge != incoming.Sym {
			available = append(available, edge)
		}
	}

	if len(available) == 0 {
		return nil
	}

	if len(available) == 1 {
		return available[0]
	}

	// Find edge with smallest clockwise angle (rightmost turn)
	incomingAngle := math.Atan2(
		incoming.End.Y-incoming.Start.Y,
		incoming.End.X-incoming.Start.X,
	)

	var bestEdge *DirectedEdge
	bestAngleDiff := math.MaxFloat64

	for _, edge := range available {
		outgoingAngle := math.Atan2(
			edge.End.Y-edge.Start.Y,
			edge.End.X-edge.Start.X,
		)

		// Compute signed angle difference (CCW positive, CW negative)
		angleDiff := outgoingAngle - incomingAngle

		// Normalize to (-π, π]
		for angleDiff <= -math.Pi {
			angleDiff += 2 * math.Pi
		}
		for angleDiff > math.Pi {
			angleDiff -= 2 * math.Pi
		}

		// For rightmost turn, we want most negative angle (most clockwise)
		if angleDiff < bestAngleDiff {
			bestAngleDiff = angleDiff
			bestEdge = edge
		}
	}

	return bestEdge
}

// buildGraph builds an EdgeGraph from a collection of noded LineStrings.
func buildGraph(lines []*geom.LineString) *EdgeGraph {
	graph := newEdgeGraph()

	for _, line := range lines {
		coords := line.Coordinates()
		for i := 0; i < len(coords)-1; i++ {
			graph.addEdge(coords[i], coords[i+1])
		}
	}

	return graph
}

// Polygonize is a convenience function that polygonizes a collection of LineStrings
// and returns the resulting polygons.
func Polygonize(lines []*geom.LineString) []*geom.Polygon {
	p := NewPolygonizer()
	p.AddAll(lines)
	return p.GetPolygons()
}
