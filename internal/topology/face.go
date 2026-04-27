package topology

import (
	"math"
	"sort"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

// LabeledDirectedSegment is an oriented boundary edge with the face label on
// its left side.
type LabeledDirectedSegment struct {
	Start, End geom.Coordinate
	Left       PolygonEdgeLabel
}

// LabeledRing is a traced face boundary together with the face label.
type LabeledRing struct {
	Ring  geom.CoordinateSequence
	Label PolygonEdgeLabel
}

// TracePolygonBoundaryFaces traces labeled faces from a polygon boundary graph.
// Each boundary edge contributes two directed half-edges: the stored direction
// has the edge's Left label, and the reverse direction has the edge's Right
// label.
func TracePolygonBoundaryFaces(edges []PolygonBoundaryEdge) []LabeledRing {
	segments := make([]LabeledDirectedSegment, 0, len(edges)*2)
	for _, edge := range edges {
		segments = append(segments,
			LabeledDirectedSegment{
				Start: edge.Start,
				End:   edge.End,
				Left:  edge.Left,
			},
			LabeledDirectedSegment{
				Start: edge.End,
				End:   edge.Start,
				Left:  edge.Right,
			},
		)
	}
	return TraceLabeledRingsFromDirectedSegments(segments)
}

// TraceLabeledRingsFromDirectedSegments traces closed rings for directed
// half-edges whose left-side labels identify the face being followed.
func TraceLabeledRingsFromDirectedSegments(segments []LabeledDirectedSegment) []LabeledRing {
	if len(segments) == 0 {
		return nil
	}

	edgeMap := make(map[geom.CoordinateXY][]int)
	for i, segment := range segments {
		if segment.Start.Distance(segment.End) <= geom.DefaultEpsilon {
			continue
		}
		edgeMap[segment.Start.XY()] = append(edgeMap[segment.Start.XY()], i)
	}
	sortLabeledAdjacency(edgeMap, segments)

	used := make([]bool, len(segments))
	rings := make([]LabeledRing, 0)
	for i := range segments {
		if used[i] {
			continue
		}
		ring, path := traceLabeledRing(i, segments, edgeMap, used)
		if len(ring) < 4 {
			continue
		}
		for _, idx := range path {
			used[idx] = true
		}
		rings = append(rings, LabeledRing{
			Ring:  ring,
			Label: segments[i].Left,
		})
	}
	return rings
}

func traceLabeledRing(startIdx int, segments []LabeledDirectedSegment, edgeMap map[geom.CoordinateXY][]int, used []bool) (geom.CoordinateSequence, []int) {
	start := segments[startIdx]
	ring := geom.CoordinateSequence{start.Start}
	path := []int{startIdx}
	seen := map[int]struct{}{startIdx: {}}
	currentIdx := startIdx

	maxSteps := len(segments) + 1
	for steps := 0; steps < maxSteps; steps++ {
		current := segments[currentIdx]
		ring = append(ring, current.End)
		if current.End.Equals2D(start.Start, geom.DefaultEpsilon) {
			return ring, path
		}

		nextIdx := findNextLabeledSegment(currentIdx, start.Left, segments, edgeMap, used, seen)
		if nextIdx < 0 {
			return nil, nil
		}
		path = append(path, nextIdx)
		seen[nextIdx] = struct{}{}
		currentIdx = nextIdx
	}
	return nil, nil
}

func findNextLabeledSegment(
	currentIdx int,
	label PolygonEdgeLabel,
	segments []LabeledDirectedSegment,
	edgeMap map[geom.CoordinateXY][]int,
	used []bool,
	seen map[int]struct{},
) int {
	current := segments[currentIdx]
	candidates := edgeMap[current.End.XY()]
	if len(candidates) == 0 {
		return -1
	}

	reverseIdx := -1
	available := make([]int, 0, len(candidates))
	for _, candidateIdx := range candidates {
		if used[candidateIdx] || !samePolygonEdgeLabel(segments[candidateIdx].Left, label) {
			continue
		}
		if _, ok := seen[candidateIdx]; ok {
			continue
		}
		if segments[candidateIdx].End.Equals2D(current.Start, geom.DefaultEpsilon) {
			reverseIdx = candidateIdx
			continue
		}
		available = append(available, candidateIdx)
	}
	if len(available) == 0 {
		return reverseIdx
	}

	reverseAngle := math.Atan2(
		current.Start.Y-current.End.Y,
		current.Start.X-current.End.X,
	)

	bestIdx := -1
	bestAngleDiff := math.MaxFloat64
	for _, candidateIdx := range available {
		candidate := segments[candidateIdx]
		outgoingAngle := math.Atan2(
			candidate.End.Y-candidate.Start.Y,
			candidate.End.X-candidate.Start.X,
		)
		angleDiff := clockwiseAngle(reverseAngle, outgoingAngle)
		if angleDiff < bestAngleDiff {
			bestAngleDiff = angleDiff
			bestIdx = candidateIdx
		}
	}
	return bestIdx
}

func samePolygonEdgeLabel(a, b PolygonEdgeLabel) bool {
	return a.LocA == b.LocA && a.LocB == b.LocB
}

func clockwiseAngle(from, to float64) float64 {
	diff := from - to
	for diff <= 0 {
		diff += 2 * math.Pi
	}
	for diff > 2*math.Pi {
		diff -= 2 * math.Pi
	}
	return diff
}

func sortLabeledAdjacency(edgeMap map[geom.CoordinateXY][]int, segments []LabeledDirectedSegment) {
	for key := range edgeMap {
		sort.Slice(edgeMap[key], func(i, j int) bool {
			a := segments[edgeMap[key][i]]
			b := segments[edgeMap[key][j]]
			angleA := math.Atan2(a.End.Y-a.Start.Y, a.End.X-a.Start.X)
			angleB := math.Atan2(b.End.Y-b.Start.Y, b.End.X-b.Start.X)
			if angleA != angleB {
				return angleA < angleB
			}
			if a.End.X != b.End.X {
				return a.End.X < b.End.X
			}
			return a.End.Y < b.End.Y
		})
	}
}
