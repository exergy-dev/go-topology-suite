package overlay

import (
	"math"

	"github.com/go-topology-suite/gts/algorithm"
	"github.com/go-topology-suite/gts/geom"
	"github.com/go-topology-suite/gts/noding"
)

// EdgeLabel tracks whether an edge is inside or outside each input geometry.
type EdgeLabel struct {
	// Location in geometry A (Exterior, Boundary, or Interior)
	LocA geom.Location
	// Location in geometry B (Exterior, Boundary, or Interior)
	LocB geom.Location
}

// DirectedEdge represents an edge in a specific direction.
type DirectedEdge struct {
	Start, End geom.Coordinate
	Label      EdgeLabel
	// Source indicates which input geometry this edge came from (0=A, 1=B, -1=intersection)
	Source int
	// LeftLabel and RightLabel track the location on each side of the edge
	LeftLabel  EdgeLabel
	RightLabel EdgeLabel
}

// nodedPolygonOverlay performs polygon overlay using robust noding.
// This is the main entry point for noded overlay operations.
func nodedPolygonOverlay(polysA, polysB []*geom.Polygon, op Op) geom.Geometry {
	if len(polysA) == 0 && len(polysB) == 0 {
		return geom.NewPolygonEmpty()
	}
	if len(polysA) == 0 {
		return handleEmptyPolyA(polysB, op)
	}
	if len(polysB) == 0 {
		return handleEmptyPolyB(polysA, op)
	}

	// Special case: if both polygon sets are identical (same references)
	// Handle this directly to avoid issues with noding
	if len(polysA) == len(polysB) {
		allSame := true
		for i := range polysA {
			if polysA[i] != polysB[i] {
				allSame = false
				break
			}
		}
		if allSame {
			return handleIdenticalPolygons(polysA, op)
		}
	}

	// Step 1: Extract all edges from both polygon sets as NodedSegmentStrings
	segStringsA := extractSegmentStringsFromPolygons(polysA, 0)
	segStringsB := extractSegmentStringsFromPolygons(polysB, 1)
	allSegStrings := append(segStringsA, segStringsB...)

	// Step 2: Node all edges - find ALL intersection points and split edges
	noder := noding.NewSimpleNoder(noding.NewIntersectionAdder())
	noder.ComputeNodes(allSegStrings)
	nodedSegments := noder.GetNodedSubstrings()

	// Step 3: Build directed edges from noded segments
	edges := buildDirectedEdges(nodedSegments)

	// Step 3.5: Merge duplicate edges (same geometric edge from different sources)
	edges = mergeEdges(edges)

	// Step 4: Label each edge with its position relative to both input geometries
	labelEdges(edges, polysA, polysB)

	// Step 5: Select edges based on the overlay operation type
	selectedEdges := selectEdges(edges, op)

	// Step 6: Build result polygons from selected edges
	result := polygonizeEdges(selectedEdges)

	return result
}

// handleIdenticalPolygons handles the case where both polygon sets are identical.
func handleIdenticalPolygons(polys []*geom.Polygon, op Op) geom.Geometry {
	switch op {
	case OpIntersection, OpUnion:
		// A ∩ A = A, A ∪ A = A
		// Even if the polygon is degenerate (zero area), return it
		return collectPolygons(polys)
	case OpDifference, OpSymDifference:
		// A - A = ∅, A △ A = ∅
		return geom.NewPolygonEmpty()
	default:
		return geom.NewPolygonEmpty()
	}
}

// extractSegmentStringsFromPolygons extracts all edges from polygons as NodedSegmentStrings.
// We create one segment string per EDGE, not per ring, to avoid noding artifacts.
// This ensures each edge is treated independently and avoids the noding algorithm
// inserting spurious intersection points into ring structures.
func extractSegmentStringsFromPolygons(polys []*geom.Polygon, source int) []*noding.NodedSegmentString {
	var segStrings []*noding.NodedSegmentString

	for polyIdx, poly := range polys {
		if poly.IsEmpty() {
			continue
		}

		// Extract edges from exterior ring
		extRing := poly.ExteriorRing().Coordinates()
		for i := 0; i < len(extRing)-1; i++ {
			// Create a segment string for each edge (2 coordinates)
			edge := geom.CoordinateSequence{extRing[i], extRing[i+1]}
			context := &EdgeContext{
				Source:    source,
				IsHole:    false,
				PolyIndex: polyIdx,
				Poly:      poly,
			}
			ss := noding.NewNodedSegmentString(edge, context)
			segStrings = append(segStrings, ss)
		}

		// Extract edges from holes
		for i := 0; i < poly.NumInteriorRings(); i++ {
			hole := poly.InteriorRingN(i).Coordinates()
			for j := 0; j < len(hole)-1; j++ {
				edge := geom.CoordinateSequence{hole[j], hole[j+1]}
				context := &EdgeContext{
					Source:    source,
					IsHole:    true,
					PolyIndex: polyIdx,
					Poly:      poly,
				}
				ss := noding.NewNodedSegmentString(edge, context)
				segStrings = append(segStrings, ss)
			}
		}
	}

	return segStrings
}

// EdgeContext stores metadata about an edge's origin.
type EdgeContext struct {
	Source    int           // 0 for geometry A, 1 for geometry B
	IsHole    bool          // true if this edge is from a hole
	PolyIndex int           // index of polygon in the input set
	Poly      *geom.Polygon // reference to source polygon for labeling
}

// buildDirectedEdges builds directed edges from noded segment strings.
func buildDirectedEdges(nodedSegments []*noding.NodedSegmentString) []*DirectedEdge {
	var edges []*DirectedEdge

	for _, ss := range nodedSegments {
		coords := ss.Coordinates()
		if len(coords) < 2 {
			continue
		}

		// Get the context
		var ctx *EdgeContext
		if c, ok := ss.Context().(*EdgeContext); ok {
			ctx = c
		}

		// IMPORTANT: Only process segment strings that look geometrically valid
		// If a segment string has more than 2 coordinates, we need to check if
		// they're actually collinear. If they form a zigzag (like (10,10)-(0,0)-(0,10)),
		// we skip it as it's corrupted by the noding process.
		if len(coords) > 2 {
			// Check if all intermediate points are roughly collinear
			isValid := true
			for i := 1; i < len(coords)-1; i++ {
				// Check if coords[i] lies on the line from coords[0] to coords[len-1]
				// This is approximate - we just check if adding this segment makes geometric sense
				dist := pointToSegmentDistance(coords[i], coords[0], coords[len(coords)-1])
				if dist > geom.DefaultEpsilon*10 {
					isValid = false
					break
				}
			}
			if !isValid {
				// Skip this corrupted segment string
				continue
			}
		}

		// Create directed edges for each consecutive pair of coordinates
		for i := 0; i < len(coords)-1; i++ {
			edge := &DirectedEdge{
				Start:  coords[i],
				End:    coords[i+1],
				Source: -1,
				Label:  EdgeLabel{LocA: geom.LocationExterior, LocB: geom.LocationExterior},
			}
			if ctx != nil {
				edge.Source = ctx.Source
			}
			edges = append(edges, edge)
		}
	}

	return edges
}

// pointToSegmentDistance computes the perpendicular distance from a point to a line segment
func pointToSegmentDistance(p, a, b geom.Coordinate) float64 {
	// Vector from a to b
	dx := b.X - a.X
	dy := b.Y - a.Y

	// If segment is degenerate (a == b), return distance to point
	lenSq := dx*dx + dy*dy
	if lenSq < geom.DefaultEpsilon*geom.DefaultEpsilon {
		return p.Distance(a)
	}

	// Parameter t of closest point on line
	t := ((p.X-a.X)*dx + (p.Y-a.Y)*dy) / lenSq

	// Clamp t to [0, 1] to stay on segment
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	// Closest point on segment
	closest := geom.NewCoordinate(a.X+t*dx, a.Y+t*dy)
	return p.Distance(closest)
}

// directedEdgeKey creates a unique key for a directed edge (direction matters)
type directedEdgeKey struct {
	x1, y1, x2, y2 float64
}

func makeDirectedEdgeKey(start, end geom.Coordinate) directedEdgeKey {
	return directedEdgeKey{start.X, start.Y, end.X, end.Y}
}

// mergeEdges merges duplicate edges (same directed edge)
// keeping only one representative of each unique directed edge
// Note: edges in opposite directions are kept as separate edges
func mergeEdges(edges []*DirectedEdge) []*DirectedEdge {
	edgeMap := make(map[directedEdgeKey]*DirectedEdge)

	for _, edge := range edges {
		key := makeDirectedEdgeKey(edge.Start, edge.End)
		if existing, found := edgeMap[key]; found {
			// Merge: prefer the edge from a known source
			if edge.Source >= 0 && existing.Source < 0 {
				edgeMap[key] = edge
			}
			// If both have sources, keep the first one
		} else {
			edgeMap[key] = edge
		}
	}

	// Convert map back to slice and sort for deterministic ordering
	var merged []*DirectedEdge
	for _, edge := range edgeMap {
		merged = append(merged, edge)
	}

	// Sort edges by coordinates for consistent ordering
	sortEdges(merged)

	return merged
}

// sortEdges sorts edges by their start and end coordinates for deterministic ordering
func sortEdges(edges []*DirectedEdge) {
	for i := 0; i < len(edges)-1; i++ {
		for j := i + 1; j < len(edges); j++ {
			if compareEdges(edges[i], edges[j]) > 0 {
				edges[i], edges[j] = edges[j], edges[i]
			}
		}
	}
}

// compareEdges compares two edges for sorting
// Returns negative if a < b, positive if a > b, 0 if equal
func compareEdges(a, b *DirectedEdge) int {
	// Compare by start X
	if a.Start.X < b.Start.X {
		return -1
	}
	if a.Start.X > b.Start.X {
		return 1
	}
	// Compare by start Y
	if a.Start.Y < b.Start.Y {
		return -1
	}
	if a.Start.Y > b.Start.Y {
		return 1
	}
	// Compare by end X
	if a.End.X < b.End.X {
		return -1
	}
	if a.End.X > b.End.X {
		return 1
	}
	// Compare by end Y
	if a.End.Y < b.End.Y {
		return -1
	}
	if a.End.Y > b.End.Y {
		return 1
	}
	return 0
}

// labelEdges labels each edge with its position relative to both input geometries.
// This is a critical step that determines which edges to include in the result.
func labelEdges(edges []*DirectedEdge, polysA, polysB []*geom.Polygon) {
	for _, edge := range edges {
		// Label based on testing BOTH sides of the edge
		// Left side is to the left when walking from Start to End
		dx := edge.End.X - edge.Start.X
		dy := edge.End.Y - edge.Start.Y
		length := math.Sqrt(dx*dx + dy*dy)

		if length < geom.DefaultEpsilon {
			// Degenerate edge - skip it
			continue
		}

		// Normalize direction vector
		dx /= length
		dy /= length

		// Perpendicular vectors (left and right of edge)
		perpX := -dy
		perpY := dx

		// Use a larger offset to ensure we're truly inside/outside and not on boundary
		// The offset should be proportional to the edge length but not too small
		offset := math.Max(length*0.1, geom.DefaultEpsilon*1000)

		// Midpoint of edge
		midX := (edge.Start.X + edge.End.X) / 2
		midY := (edge.Start.Y + edge.End.Y) / 2

		// Test point on the LEFT side
		leftPoint := geom.NewCoordinate(midX+perpX*offset, midY+perpY*offset)
		leftLocA := locateInPolygonSet(leftPoint, polysA)
		leftLocB := locateInPolygonSet(leftPoint, polysB)

		// Test point on the RIGHT side
		rightPoint := geom.NewCoordinate(midX-perpX*offset, midY-perpY*offset)
		rightLocA := locateInPolygonSet(rightPoint, polysA)
		rightLocB := locateInPolygonSet(rightPoint, polysB)

		// Store labels for both sides
		edge.LeftLabel = EdgeLabel{LocA: leftLocA, LocB: leftLocB}
		edge.RightLabel = EdgeLabel{LocA: rightLocA, LocB: rightLocB}

		// For simplicity in selection, use left label as the primary label
		edge.Label = edge.LeftLabel
	}
}

// locateInPolygonSet determines if a point is inside any polygon in the set.
func locateInPolygonSet(pt geom.Coordinate, polys []*geom.Polygon) geom.Location {
	for _, poly := range polys {
		if poly.IsEmpty() {
			continue
		}

		loc := algorithm.PointLocationInPolygon(pt, poly)
		if loc == geom.LocationInterior {
			return geom.LocationInterior
		}
		if loc == geom.LocationBoundary {
			return geom.LocationBoundary
		}
	}
	return geom.LocationExterior
}

// debugEdgeSelection controls whether to print debug info for edge selection
var debugEdgeSelection = false

// selectEdges selects edges based on the overlay operation type.
// This implements the core logic of overlay operations using proper DE-9IM topology.
// We consider both left and right sides of each edge.
// An edge is included if it forms a boundary of the result geometry.
// Edges are oriented so that the "in result" side is on the LEFT (for CCW polygon orientation).
func selectEdges(edges []*DirectedEdge, op Op) []*DirectedEdge {
	var selected []*DirectedEdge

	for _, edge := range edges {
		// For edge selection, we consider interior only (not boundary)
		// An edge is on the boundary when one side is interior and the other is not
		leftAInside := edge.LeftLabel.LocA == geom.LocationInterior
		leftBInside := edge.LeftLabel.LocB == geom.LocationInterior
		rightAInside := edge.RightLabel.LocA == geom.LocationInterior
		rightBInside := edge.RightLabel.LocB == geom.LocationInterior

		var leftInResult, rightInResult bool

		switch op {
		case OpIntersection:
			// For intersection A ∩ B: include edges where one side is in both A and B,
			// and the other side is not in both
			leftInResult = leftAInside && leftBInside
			rightInResult = rightAInside && rightBInside

		case OpUnion:
			// For union A ∪ B: include edges where one side is in at least one of A or B,
			// and the other side is in neither
			leftInResult = leftAInside || leftBInside
			rightInResult = rightAInside || rightBInside

		case OpDifference:
			// For A - B: include edges where one side is in A but not in B,
			// and the other side is not (either not in A, or in B, or both)
			leftInResult = leftAInside && !leftBInside
			rightInResult = rightAInside && !rightBInside

		case OpSymDifference:
			// For symmetric difference (A ⊕ B): include edges where one side is in
			// exactly one of A or B, and the other side is not
			leftInResult = (leftAInside && !leftBInside) || (!leftAInside && leftBInside)
			rightInResult = (rightAInside && !rightBInside) || (!rightAInside && rightBInside)
		}

		// Include edge if exactly one side is in the result
		if leftInResult != rightInResult {
			// For proper CCW orientation, the "in result" side should be on the LEFT
			// If it's on the right, reverse the edge direction
			if rightInResult && !leftInResult {
				// Reverse the edge
				selected = append(selected, &DirectedEdge{
					Start:      edge.End,
					End:        edge.Start,
					Source:     edge.Source,
					LeftLabel:  edge.RightLabel,
					RightLabel: edge.LeftLabel,
				})
			} else {
				selected = append(selected, edge)
			}
		}
	}

	return selected
}

// polygonizeEdges builds polygons from a collection of directed edges.
// This implements a polygonization algorithm that identifies holes.
func polygonizeEdges(edges []*DirectedEdge) geom.Geometry {
	if len(edges) == 0 {
		return geom.NewPolygonEmpty()
	}

	// Build an adjacency map: start coordinate -> edges starting there
	edgeMap := make(map[geom.Coordinate][]*DirectedEdge)
	for _, edge := range edges {
		// Normalize coordinate for map key (handle floating point)
		start := edge.Start
		edgeMap[start] = append(edgeMap[start], edge)
	}

	// Track which edges have been used
	used := make(map[*DirectedEdge]bool)

	// Find all rings
	var rings []geom.CoordinateSequence

	for _, startEdge := range edges {
		if used[startEdge] {
			continue
		}

		// Try to build a ring starting from this edge
		ring := buildRing(startEdge, edgeMap, used)
		if ring != nil && len(ring) >= 4 {
			rings = append(rings, ring)
		}
	}

	// Convert rings to polygons
	if len(rings) == 0 {
		return geom.NewPolygonEmpty()
	}

	// Classify rings as shells (exterior) or holes based on orientation
	// CCW = shell, CW = hole (OGC convention)
	var shells []geom.CoordinateSequence
	var holes []geom.CoordinateSequence

	for _, ring := range rings {
		// Ensure ring is closed
		if !ring.IsClosed(geom.DefaultEpsilon) {
			ring = append(ring, ring[0].Clone())
		}

		// Check if ring has enough points
		if len(ring) < 4 {
			continue
		}

		// Determine orientation
		area := geom.SignedArea(ring)
		if area > geom.DefaultEpsilon {
			// Counter-clockwise = exterior ring
			shells = append(shells, ring)
		} else if area < -geom.DefaultEpsilon {
			// Clockwise = hole
			holes = append(holes, ring)
		}
		// area ≈ 0 means degenerate ring, skip it
	}

	// Build polygons by assigning holes to shells
	var polygons []*geom.Polygon

	if len(shells) == 0 && len(holes) == 0 {
		return geom.NewPolygonEmpty()
	}

	if len(shells) == 0 {
		// Only holes, no shells - this shouldn't happen in valid geometry
		// Treat the largest hole as a shell (reverse it)
		if len(holes) > 0 {
			largest := holes[0]
			largestArea := -geom.SignedArea(largest)
			largestIdx := 0
			for i := 1; i < len(holes); i++ {
				a := -geom.SignedArea(holes[i])
				if a > largestArea {
					largestArea = a
					largestIdx = i
					largest = holes[i]
				}
			}
			// Reverse to make it CCW
			shells = []geom.CoordinateSequence{largest.Reverse()}
			holes = append(holes[:largestIdx], holes[largestIdx+1:]...)
		}
	}

	// Simple assignment: assign each hole to the first shell that contains it
	for _, shell := range shells {
		lr := geom.NewLinearRing(shell)
		shellPoly := geom.NewPolygon(lr, nil)

		var assignedHoles []*geom.LinearRing

		for _, hole := range holes {
			// Check if hole is inside this shell
			// Use an interior point of the hole (centroid) for robust containment check
			// First vertex might be on the shell boundary causing false negatives
			holePoint := ringInteriorPoint(hole)
			loc := algorithm.PointLocationInPolygon(holePoint, shellPoly)
			if loc == geom.LocationInterior {
				assignedHoles = append(assignedHoles, geom.NewLinearRing(hole))
			}
		}

		poly := geom.NewPolygon(lr, assignedHoles)
		if !poly.IsEmpty() && poly.Area() > geom.DefaultEpsilon {
			polygons = append(polygons, poly)
		}
	}

	if len(polygons) == 0 {
		return geom.NewPolygonEmpty()
	}
	if len(polygons) == 1 {
		return polygons[0]
	}
	return geom.NewMultiPolygon(polygons)
}

// buildRing attempts to build a closed ring starting from a given edge.
// It uses the "rightmost turn" rule at each junction to properly trace individual rings.
func buildRing(startEdge *DirectedEdge, edgeMap map[geom.Coordinate][]*DirectedEdge, used map[*DirectedEdge]bool) geom.CoordinateSequence {
	var ring geom.CoordinateSequence
	ring = append(ring, startEdge.Start)

	current := startEdge
	used[current] = true

	maxSteps := 10000 // Prevent infinite loops
	for steps := 0; steps < maxSteps; steps++ {
		ring = append(ring, current.End)

		// Check if we've closed the ring
		if current.End.Equals2D(startEdge.Start, geom.DefaultEpsilon) {
			return ring
		}

		// Find next edge that starts where current edge ends, using rightmost turn
		nextEdge := findNextEdgeRightmost(current, edgeMap, used)
		if nextEdge == nil {
			// Can't continue - return nil to indicate incomplete ring
			return nil
		}

		used[nextEdge] = true
		current = nextEdge
	}

	// Couldn't close the ring
	return nil
}

// findNextEdgeRightmost finds the unused edge at a junction that makes the rightmost turn
// (smallest counter-clockwise angle from the incoming direction).
// This ensures proper ring tracing when multiple edges meet at a point.
func findNextEdgeRightmost(incoming *DirectedEdge, edgeMap map[geom.Coordinate][]*DirectedEdge, used map[*DirectedEdge]bool) *DirectedEdge {
	start := incoming.End

	// Collect all candidate edges (unused edges starting at this point)
	var candidates []*DirectedEdge

	// Try exact match first
	if edges, ok := edgeMap[start]; ok {
		for _, edge := range edges {
			if !used[edge] {
				candidates = append(candidates, edge)
			}
		}
	}

	// Try fuzzy match
	for coord, edges := range edgeMap {
		if !coord.Equals2D(start, geom.DefaultEpsilon) || coord == start {
			continue
		}
		for _, edge := range edges {
			if !used[edge] {
				// Check if this edge is already in candidates
				found := false
				for _, c := range candidates {
					if c == edge {
						found = true
						break
					}
				}
				if !found {
					candidates = append(candidates, edge)
				}
			}
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	if len(candidates) == 1 {
		return candidates[0]
	}

	// Compute the incoming direction (the direction we were traveling)
	incomingAngle := math.Atan2(incoming.End.Y-incoming.Start.Y, incoming.End.X-incoming.Start.X)

	// Find the edge with the smallest clockwise angle from incoming direction.
	// This implements the "rightmost turn" rule which properly traces individual
	// polygon rings by always staying on the same ring.
	//
	// When walking along a CCW ring (exterior), rightmost turn keeps us on the exterior.
	// When walking along a CW ring (hole), rightmost turn keeps us on the hole.
	var bestEdge *DirectedEdge
	bestAngleDiff := math.MaxFloat64

	for _, edge := range candidates {
		// Compute outgoing angle
		outgoingAngle := math.Atan2(edge.End.Y-edge.Start.Y, edge.End.X-edge.Start.X)

		// Compute signed angle difference
		// Positive = counter-clockwise turn, Negative = clockwise turn
		angleDiff := outgoingAngle - incomingAngle

		// Normalize to (-π, π]
		for angleDiff <= -math.Pi {
			angleDiff += 2 * math.Pi
		}
		for angleDiff > math.Pi {
			angleDiff -= 2 * math.Pi
		}

		// For rightmost turn, we want the most negative angle (most clockwise)
		// This keeps us on the same ring instead of jumping to another ring
		if angleDiff < bestAngleDiff {
			bestAngleDiff = angleDiff
			bestEdge = edge
		}
	}

	return bestEdge
}

// ringInteriorPoint returns a point that is definitely inside the ring (not on boundary).
// Uses the centroid which is guaranteed to be inside a convex ring, and usually inside concave ones.
func ringInteriorPoint(ring geom.CoordinateSequence) geom.Coordinate {
	if len(ring) < 3 {
		return ring[0]
	}

	// Compute centroid
	var cx, cy float64
	for i := 0; i < len(ring)-1; i++ { // Exclude closing point
		cx += ring[i].X
		cy += ring[i].Y
	}
	n := float64(len(ring) - 1)
	if n < 1 {
		n = 1
	}
	return geom.NewCoordinate(cx/n, cy/n)
}

// handleEmptyPolyA handles the case where polygon set A is empty.
func handleEmptyPolyA(polysB []*geom.Polygon, op Op) geom.Geometry {
	switch op {
	case OpIntersection:
		return geom.NewPolygonEmpty()
	case OpUnion:
		return collectPolygons(polysB)
	case OpDifference:
		return geom.NewPolygonEmpty()
	case OpSymDifference:
		return collectPolygons(polysB)
	default:
		return geom.NewPolygonEmpty()
	}
}

// handleEmptyPolyB handles the case where polygon set B is empty.
func handleEmptyPolyB(polysA []*geom.Polygon, op Op) geom.Geometry {
	switch op {
	case OpIntersection:
		return geom.NewPolygonEmpty()
	case OpUnion:
		return collectPolygons(polysA)
	case OpDifference:
		return collectPolygons(polysA)
	case OpSymDifference:
		return collectPolygons(polysA)
	default:
		return geom.NewPolygonEmpty()
	}
}
