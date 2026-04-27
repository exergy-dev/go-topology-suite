package overlay

import (
	"sort"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/internal/topology"
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
	return nodedPolygonOverlayWithPrecision(polysA, polysB, op, nil)
}

func nodedPolygonOverlayWithPrecision(polysA, polysB []*geom.Polygon, op Op, pm geom.PrecisionModel) geom.Geometry {
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

	// Step 1: Build a shared noded and labeled boundary graph.
	graphEdges := topology.BuildPolygonBoundaryGraphWithPrecision(polysA, polysB, pm)

	// Step 2: Select labeled faces and dissolve adjacent selected faces into
	// result shells and holes.
	result := polygonizeLabeledFaces(graphEdges, op)

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

func directedEdgesFromBoundaryGraph(graphEdges []topology.PolygonBoundaryEdge) []*DirectedEdge {
	edges := make([]*DirectedEdge, 0, len(graphEdges))
	for _, graphEdge := range graphEdges {
		left := EdgeLabel{LocA: graphEdge.Left.LocA, LocB: graphEdge.Left.LocB}
		right := EdgeLabel{LocA: graphEdge.Right.LocA, LocB: graphEdge.Right.LocB}
		edges = append(edges, &DirectedEdge{
			Start:      graphEdge.Start,
			End:        graphEdge.End,
			Source:     -1,
			Label:      left,
			LeftLabel:  left,
			RightLabel: right,
		})
	}
	return edges
}

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

		leftInResult := labelInResult(leftAInside, leftBInside, op)
		rightInResult := labelInResult(rightAInside, rightBInside, op)

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

func labelInResult(aInside, bInside bool, op Op) bool {
	switch op {
	case OpIntersection:
		return aInside && bInside
	case OpUnion:
		return aInside || bInside
	case OpDifference:
		return aInside && !bInside
	case OpSymDifference:
		return (aInside && !bInside) || (!aInside && bInside)
	default:
		return false
	}
}

// polygonizeEdges builds polygons from a collection of directed edges.
// This implements a polygonization algorithm that identifies holes.
func polygonizeEdges(edges []*DirectedEdge) geom.Geometry {
	if len(edges) == 0 {
		return geom.NewPolygonEmpty()
	}

	segments := make([]topology.DirectedSegment, 0, len(edges))
	for _, edge := range edges {
		segments = append(segments, topology.DirectedSegment{
			Start: edge.Start,
			End:   edge.End,
		})
	}
	rings := topology.TraceRingsFromDirectedSegments(segments)

	// Convert rings to polygons
	if len(rings) == 0 {
		return geom.NewPolygonEmpty()
	}

	polygons := topology.PolygonsFromRings(rings, false)
	if len(polygons) == 0 {
		return geom.NewPolygonEmpty()
	}
	polygons = normalizeOverlayPolygons(polygons)
	if len(polygons) == 1 {
		return polygons[0]
	}
	return geom.NewMultiPolygon(polygons)
}

func polygonizeLabeledFaces(graphEdges []topology.PolygonBoundaryEdge, op Op) geom.Geometry {
	faces := topology.TracePolygonBoundaryFaces(graphEdges)
	if len(faces) == 0 {
		return geom.NewPolygonEmpty()
	}

	rings := make([]geom.CoordinateSequence, 0, len(faces))
	for _, face := range faces {
		if !polygonLabelInResult(face.Label, op) {
			continue
		}
		rings = append(rings, face.Ring)
	}
	if len(rings) == 0 {
		return geom.NewPolygonEmpty()
	}

	rings = dissolveFaceRings(rings)
	if len(rings) == 0 {
		return geom.NewPolygonEmpty()
	}

	polygons := topology.PolygonsFromRings(rings, true)
	if len(polygons) == 0 {
		return geom.NewPolygonEmpty()
	}
	polygons = normalizeOverlayPolygons(polygons)
	if len(polygons) == 1 {
		return polygons[0]
	}
	return geom.NewMultiPolygon(polygons)
}

func polygonLabelInResult(label topology.PolygonEdgeLabel, op Op) bool {
	return labelInResult(
		label.LocA == geom.LocationInterior,
		label.LocB == geom.LocationInterior,
		op,
	)
}

type dissolvedSegmentKey struct {
	x1, y1, x2, y2 float64
}

func dissolveFaceRings(rings []geom.CoordinateSequence) []geom.CoordinateSequence {
	segmentsByKey := make(map[dissolvedSegmentKey][]topology.DirectedSegment)
	for _, ring := range rings {
		if len(ring) < 2 {
			continue
		}
		for i := 1; i < len(ring); i++ {
			start := ring[i-1]
			end := ring[i]
			if start.Equals2D(end, geom.DefaultEpsilon) {
				continue
			}
			key := makeDissolvedSegmentKey(start, end)
			segmentsByKey[key] = append(segmentsByKey[key], topology.DirectedSegment{
				Start: start,
				End:   end,
			})
		}
	}

	boundarySegments := make([]topology.DirectedSegment, 0, len(segmentsByKey))
	for _, segments := range segmentsByKey {
		if len(segments) == 1 {
			boundarySegments = append(boundarySegments, segments[0])
			continue
		}
		boundarySegments = append(boundarySegments, unpairedDissolvedSegments(segments)...)
	}

	return topology.TraceRingsFromDirectedSegments(boundarySegments)
}

func unpairedDissolvedSegments(segments []topology.DirectedSegment) []topology.DirectedSegment {
	used := make([]bool, len(segments))
	for i := range segments {
		if used[i] {
			continue
		}
		for j := i + 1; j < len(segments); j++ {
			if used[j] {
				continue
			}
			if segments[i].Start.Equals2D(segments[j].End, geom.DefaultEpsilon) &&
				segments[i].End.Equals2D(segments[j].Start, geom.DefaultEpsilon) {
				used[i] = true
				used[j] = true
				break
			}
		}
	}

	unpaired := make([]topology.DirectedSegment, 0, len(segments))
	for i, segment := range segments {
		if !used[i] {
			unpaired = append(unpaired, segment)
		}
	}
	return unpaired
}

func makeDissolvedSegmentKey(a, b geom.Coordinate) dissolvedSegmentKey {
	if b.X < a.X || (b.X == a.X && b.Y < a.Y) {
		a, b = b, a
	}
	return dissolvedSegmentKey{a.X, a.Y, b.X, b.Y}
}

func normalizeOverlayPolygons(polygons []*geom.Polygon) []*geom.Polygon {
	normalized := make([]*geom.Polygon, 0, len(polygons))
	for _, polygon := range polygons {
		normalized = append(normalized, normalizeOverlayPolygon(polygon))
	}
	sort.Slice(normalized, func(i, j int) bool {
		return geom.Compare(normalized[i], normalized[j]) < 0
	})
	return normalized
}

func normalizeOverlayPolygon(polygon *geom.Polygon) *geom.Polygon {
	shell := polygon.ExteriorRing().Normalized().(*geom.LinearRing)
	if shell.IsCW() {
		shell = shell.Reverse().Normalized().(*geom.LinearRing)
	}

	holes := make([]*geom.LinearRing, 0, polygon.NumInteriorRings())
	for i := 0; i < polygon.NumInteriorRings(); i++ {
		hole := polygon.InteriorRingN(i).Normalized().(*geom.LinearRing)
		if hole.IsCCW() {
			hole = hole.Reverse().Normalized().(*geom.LinearRing)
		}
		holes = append(holes, hole)
	}
	sort.Slice(holes, func(i, j int) bool {
		return compareLinearRings(holes[i], holes[j]) < 0
	})

	return geom.NewPolygon(shell, holes)
}

func compareLinearRings(a, b *geom.LinearRing) int {
	coordsA := a.Coordinates()
	coordsB := b.Coordinates()
	for i := 0; i < len(coordsA) && i < len(coordsB); i++ {
		if coordsA[i].X < coordsB[i].X {
			return -1
		}
		if coordsA[i].X > coordsB[i].X {
			return 1
		}
		if coordsA[i].Y < coordsB[i].Y {
			return -1
		}
		if coordsA[i].Y > coordsB[i].Y {
			return 1
		}
	}
	if len(coordsA) < len(coordsB) {
		return -1
	}
	if len(coordsA) > len(coordsB) {
		return 1
	}
	return 0
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
